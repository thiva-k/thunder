/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package openid4vp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/thunder-id/thunderid/internal/system/error/apierror"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
	"github.com/thunder-id/thunderid/internal/system/i18n/core"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/internal/system/middleware"
	sysutils "github.com/thunder-id/thunderid/internal/system/utils"
)

const requestObjectContentType = "application/oauth-authz-req+jwt"

// handler serves the wallet-facing OpenID4VP endpoints.
type handler struct {
	service *Service
}

func newHandler(service *Service) *handler {
	return &handler{service: service}
}

func (h *handler) handleRequestObject(w http.ResponseWriter, r *http.Request) {
	state := sysutils.SanitizeString(r.URL.Query().Get("state"))
	if state == "" {
		writeServiceErrorResponse(w, &ErrorInvalidRequest)
		return
	}

	jar, err := h.service.requestObject(r.Context(), state)
	if err != nil {
		writeServiceErrorResponse(w, toServiceError(err))
		return
	}

	w.Header().Set("Content-Type", requestObjectContentType)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if _, werr := w.Write([]byte(jar)); werr != nil {
		log.GetLogger().Error("Failed to write request object response", log.Error(werr))
	}
}

func (h *handler) handleResponse(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeServiceErrorResponse(w, &ErrorInvalidRequest)
		return
	}

	state := sysutils.SanitizeString(r.FormValue("state"))
	response := r.FormValue("response")
	if state == "" || response == "" {
		writeServiceErrorResponse(w, &ErrorInvalidRequest)
		return
	}

	if _, err := h.service.submitResponse(r.Context(), state, []byte(response)); err != nil {
		writeServiceErrorResponse(w, toServiceError(err))
		return
	}

	body := map[string]string{}
	if redirect := h.service.resultRedirectURI(state); redirect != "" {
		body["redirect_uri"] = redirect
	}
	sysutils.WriteSuccessResponse(w, http.StatusOK, body)
}

const (
	apiInitiatePath = "/openid4vp/initiate"
	apiStatusPath   = "/openid4vp/status/{txn_id}"
	apiStatusPrefix = "/openid4vp/status/"
)

const defaultResultTokenValidity = 300 * time.Second

var (
	// ErrorUnknownDefinition indicates the requested presentation_definition_id is not registered.
	ErrorUnknownDefinition = serviceerror.ServiceError{
		Type: serviceerror.ClientErrorType,
		Code: "EUDI-1004",
		Error: core.I18nMessage{
			Key:          "error.eudi.unknown_definition",
			DefaultValue: "Unknown presentation definition",
		},
		ErrorDescription: core.I18nMessage{
			Key:          "error.eudi.unknown_definition_description",
			DefaultValue: "No presentation definition is registered for the supplied id",
		},
	}
)

// rpHandler serves the RP-facing OpenID4VP REST endpoints. It mirrors the
// patterns of other Thunder admin handlers: a small dependency struct, JSON
// request/response shapes, and the standard apierror envelope on failure.
type rpHandler struct {
	service              *Service
	issuer               resultTokenIssuer
	rpStatusBase         string
	resultTokenValidity  time.Duration
	requestStateValidity time.Duration
}

// newRPHandler builds the RP-facing API handler. A zero validity falls back to defaultResultTokenValidity.
func newRPHandler(svc *Service, issuer resultTokenIssuer, baseURL string, validity time.Duration) *rpHandler {
	if validity <= 0 {
		validity = defaultResultTokenValidity
	}
	return &rpHandler{
		service:              svc,
		issuer:               issuer,
		rpStatusBase:         strings.TrimRight(baseURL, "/") + apiStatusPrefix,
		resultTokenValidity:  validity,
		requestStateValidity: svc.cfg.TTL,
	}
}

func (h *rpHandler) handleInitiate(w http.ResponseWriter, r *http.Request) {
	var req initiateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceErrorResponse(w, &ErrorInvalidRequest)
		return
	}
	if strings.TrimSpace(req.DefinitionID) == "" || strings.TrimSpace(req.RPID) == "" {
		writeServiceErrorResponse(w, &ErrorInvalidRequest)
		return
	}
	if !h.service.registry.has(req.DefinitionID) {
		writeServiceErrorResponse(w, &ErrorUnknownDefinition)
		return
	}

	init, err := h.service.initiateForRP(r.Context(), req.DefinitionID, req.RPID)
	if err != nil {
		log.GetLogger().Error("Failed to initiate OpenID4VP transaction", log.Error(err))
		writeServiceErrorResponse(w, toServiceError(err))
		return
	}

	rs, lookupErr := h.service.lookupState(r.Context(), init.State)
	expiresAt := time.Now().Add(h.requestStateValidity)
	if lookupErr == nil && rs != nil {
		expiresAt = rs.ExpiresAt
	}

	resp := initiateResponse{
		TxnID:     init.State,
		WalletURL: WalletAuthorizationURI(init.ClientID, init.RequestURI),
		StatusURL: h.rpStatusBase + init.State,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
	}
	sysutils.WriteSuccessResponse(w, http.StatusOK, resp)
}

// handleStatus issues a result token on COMPLETED; FAILED/EXPIRED carry a diagnostic but no token.
func (h *rpHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	txnID := strings.TrimSpace(extractTxnID(r))
	if txnID == "" {
		writeServiceErrorResponse(w, &ErrorInvalidRequest)
		return
	}

	rs, err := h.service.lookupState(r.Context(), txnID)
	switch {
	case errors.Is(err, ErrUnknownState):
		writeServiceErrorResponse(w, &ErrorUnknownState)
		return
	case errors.Is(err, ErrExpiredState):
		sysutils.WriteSuccessResponse(w, http.StatusOK, statusResponse{Status: "EXPIRED"})
		return
	case err != nil:
		writeServiceErrorResponse(w, toServiceError(err))
		return
	}

	switch rs.Status {
	case StatusPending:
		sysutils.WriteSuccessResponse(w, http.StatusOK, statusResponse{Status: "PENDING"})
	case StatusFailed:
		sysutils.WriteSuccessResponse(w, http.StatusOK, statusResponse{
			Status: "FAILED",
			Error:  rs.FailureReason,
		})
	case StatusCompleted:
		rpID := rs.RPID
		if rpID == "" {
			rpID = rs.ClientID
		}
		token, tokenErr := h.issuer.issueResultToken(
			r.Context(), rpID, rs, int64(h.resultTokenValidity.Seconds()))
		if tokenErr != nil {
			log.GetLogger().Error("Failed to issue result token", log.Error(tokenErr))
			writeServiceErrorResponse(w, &serviceerror.InternalServerError)
			return
		}
		sysutils.WriteSuccessResponse(w, http.StatusOK, statusResponse{
			Status:      "COMPLETED",
			ResultToken: token,
		})
	default:
		writeServiceErrorResponse(w, &serviceerror.InternalServerError)
	}
}

// extractTxnID resolves txn_id from a Go-1.22 path value or the trailing path segment.
func extractTxnID(r *http.Request) string {
	if v := r.PathValue("txn_id"); v != "" {
		return v
	}
	return strings.TrimPrefix(r.URL.Path, apiStatusPrefix)
}

func writeServiceErrorResponse(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		statusCode = clientErrorStatusCode(svcErr.Code)
	}
	sysutils.WriteErrorResponse(w, statusCode, apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	})
}

func clientErrorStatusCode(code string) int {
	if code == ErrorUnknownState.Code {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func registerRoutes(mux *http.ServeMux, h *handler) {
	opts := middleware.CORSOptions{
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   middleware.DefaultAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           600,
	}

	mux.HandleFunc(middleware.WithCORS("GET "+requestURIPath,
		middleware.CorrelationIDMiddleware(http.HandlerFunc(h.handleRequestObject)).ServeHTTP, opts))
	mux.HandleFunc(middleware.WithCORS("POST "+responseURIPath,
		middleware.CorrelationIDMiddleware(http.HandlerFunc(h.handleResponse)).ServeHTTP, opts))

	for _, path := range []string{requestURIPath, responseURIPath} {
		mux.HandleFunc(middleware.WithCORS("OPTIONS "+path,
			func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }, opts))
	}
}

func registerRPRoutes(mux *http.ServeMux, h *rpHandler) {
	opts := middleware.CORSOptions{
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   middleware.DefaultAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           600,
	}

	mux.HandleFunc(middleware.WithCORS("POST "+apiInitiatePath,
		middleware.CorrelationIDMiddleware(http.HandlerFunc(h.handleInitiate)).ServeHTTP, opts))
	mux.HandleFunc(middleware.WithCORS("GET "+apiStatusPath,
		middleware.CorrelationIDMiddleware(http.HandlerFunc(h.handleStatus)).ServeHTTP, opts))

	for _, path := range []string{apiInitiatePath, apiStatusPath} {
		mux.HandleFunc(middleware.WithCORS("OPTIONS "+path,
			func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }, opts))
	}
}
