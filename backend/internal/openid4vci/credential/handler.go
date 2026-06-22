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

package credential

import (
	"context"
	"net/http"
	"strings"

	"github.com/thunder-id/thunderid/internal/system/error/apierror"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
	"github.com/thunder-id/thunderid/internal/system/middleware"
	sysutils "github.com/thunder-id/thunderid/internal/system/utils"
)

const configurationsPath = "/openid4vci/credential-configurations"

// configurationHandler serves the management API for credential configurations.
type configurationHandler struct {
	service CredentialConfigurationServiceInterface
}

// newConfigurationHandler builds the credential-configuration management handler.
func newConfigurationHandler(service CredentialConfigurationServiceInterface) *configurationHandler {
	return &configurationHandler{service: service}
}

// HandleCreate creates a credential configuration.
func (h *configurationHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	req, err := sysutils.DecodeJSONBody[credentialConfigurationRequest](r)
	if err != nil {
		writeConfigurationError(r.Context(), w, &ErrorConfigurationInvalidRequest)
		return
	}
	created, svcErr := h.service.Create(r.Context(), requestToDTO(req))
	if svcErr != nil {
		writeConfigurationError(r.Context(), w, svcErr)
		return
	}
	sysutils.WriteSuccessResponse(r.Context(), w, http.StatusCreated, toResponse(*created))
}

// HandleList returns a minimal summary of all credential configurations.
func (h *configurationHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	summaries, svcErr := h.service.ListSummaries(r.Context())
	if svcErr != nil {
		writeConfigurationError(r.Context(), w, svcErr)
		return
	}
	sysutils.WriteSuccessResponse(r.Context(), w, http.StatusOK, summaries)
}

// HandleGet returns a single credential configuration.
func (h *configurationHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeConfigurationError(r.Context(), w, &ErrorConfigurationInvalidRequest)
		return
	}
	dto, svcErr := h.service.Get(r.Context(), id)
	if svcErr != nil {
		writeConfigurationError(r.Context(), w, svcErr)
		return
	}
	sysutils.WriteSuccessResponse(r.Context(), w, http.StatusOK, toResponse(*dto))
}

// HandleUpdate updates a credential configuration.
func (h *configurationHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeConfigurationError(r.Context(), w, &ErrorConfigurationInvalidRequest)
		return
	}
	req, err := sysutils.DecodeJSONBody[credentialConfigurationRequest](r)
	if err != nil {
		writeConfigurationError(r.Context(), w, &ErrorConfigurationInvalidRequest)
		return
	}
	updated, svcErr := h.service.Update(r.Context(), id, requestToDTO(req))
	if svcErr != nil {
		writeConfigurationError(r.Context(), w, svcErr)
		return
	}
	sysutils.WriteSuccessResponse(r.Context(), w, http.StatusOK, toResponse(*updated))
}

// HandleDelete deletes a credential configuration.
func (h *configurationHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeConfigurationError(r.Context(), w, &ErrorConfigurationInvalidRequest)
		return
	}
	if svcErr := h.service.Delete(r.Context(), id); svcErr != nil {
		writeConfigurationError(r.Context(), w, svcErr)
		return
	}
	sysutils.WriteSuccessResponse(r.Context(), w, http.StatusNoContent, nil)
}

// requestToDTO maps and sanitizes an API request to a managed DTO.
func requestToDTO(req *credentialConfigurationRequest) *CredentialConfigurationDTO {
	return &CredentialConfigurationDTO{
		Handle:          sysutils.SanitizeString(req.Handle),
		OUID:            sysutils.SanitizeString(req.OUID),
		OUHandle:        sysutils.SanitizeString(req.OUHandle),
		Format:          sysutils.SanitizeString(req.Format),
		VCT:             sysutils.SanitizeString(req.VCT),
		Claims:          sanitizeClaims(req.Claims),
		Display:         sanitizeDisplay(req.Display),
		ValiditySeconds: req.ValiditySeconds,
	}
}

// sanitizeClaims trims claim names and display names, dropping entries with no name.
func sanitizeClaims(in []ClaimMapping) []ClaimMapping {
	out := make([]ClaimMapping, 0, len(in))
	for _, c := range in {
		name := sysutils.SanitizeString(c.Name)
		if name == "" {
			continue
		}
		out = append(out, ClaimMapping{Name: name, DisplayName: sysutils.SanitizeString(c.DisplayName)})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// sanitizeDisplay trims the display fields, returning nil when all are empty.
func sanitizeDisplay(in *CredentialDisplay) *CredentialDisplay {
	if in == nil {
		return nil
	}
	d := CredentialDisplay{
		Name:    sysutils.SanitizeString(in.Name),
		Locale:  sysutils.SanitizeString(in.Locale),
		LogoURI: sysutils.SanitizeString(in.LogoURI),
	}
	if d.Name == "" && d.Locale == "" && d.LogoURI == "" {
		return nil
	}
	return &d
}

func writeConfigurationError(ctx context.Context, w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	status := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		status = configurationClientErrorStatus(svcErr.Code)
	}
	sysutils.WriteErrorResponse(ctx, w, status, apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	})
}

// registerRoutes registers the admin-facing management endpoints. They are
// intentionally NOT in the public-paths allowlist, so the platform auth
// middleware protects them.
func registerRoutes(mux *http.ServeMux, h *configurationHandler) {
	collectionOpts := middleware.CORSOptions{
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   middleware.DefaultAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           600,
	}
	resourceOpts := middleware.CORSOptions{
		AllowedMethods:   []string{"GET", "PUT", "DELETE"},
		AllowedHeaders:   middleware.DefaultAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           600,
	}

	mux.HandleFunc(middleware.WithCORS("POST "+configurationsPath,
		middleware.CorrelationIDMiddleware(http.HandlerFunc(h.HandleCreate)).ServeHTTP, collectionOpts))
	mux.HandleFunc(middleware.WithCORS("GET "+configurationsPath,
		middleware.CorrelationIDMiddleware(http.HandlerFunc(h.HandleList)).ServeHTTP, collectionOpts))
	mux.HandleFunc(middleware.WithCORS("GET "+configurationsPath+"/{id}",
		middleware.CorrelationIDMiddleware(http.HandlerFunc(h.HandleGet)).ServeHTTP, resourceOpts))
	mux.HandleFunc(middleware.WithCORS("PUT "+configurationsPath+"/{id}",
		middleware.CorrelationIDMiddleware(http.HandlerFunc(h.HandleUpdate)).ServeHTTP, resourceOpts))
	mux.HandleFunc(middleware.WithCORS("DELETE "+configurationsPath+"/{id}",
		middleware.CorrelationIDMiddleware(http.HandlerFunc(h.HandleDelete)).ServeHTTP, resourceOpts))

	mux.HandleFunc(middleware.WithCORS("OPTIONS "+configurationsPath,
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }, collectionOpts))
	mux.HandleFunc(middleware.WithCORS("OPTIONS "+configurationsPath+"/{id}",
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }, resourceOpts))
}
