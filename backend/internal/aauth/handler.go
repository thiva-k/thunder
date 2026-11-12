// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/thunder-id/thunderid/internal/system/log"
)

// decisionCookie carries the decision token so that only the browser that signed in can decide.
const decisionCookie = "aauth_decision"

// aauthHandler serves the person server's HTTP endpoints.
type aauthHandler struct {
	service *aauthService
	logger  *log.Logger
}

func newAAuthHandler(service *aauthService) *aauthHandler {
	return &aauthHandler{
		service: service,
		logger:  log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AAuthHandler")),
	}
}

// HandleMetadata serves /.well-known/aauth-person.json. Agents may sign the request; the
// signature is not needed to read public metadata.
func (h *aauthHandler) HandleMetadata(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.metadata())
}

// HandleJWKS serves the person server's signing keys.
func (h *aauthHandler) HandleJWKS(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.jwks())
}

// HandlePersonToken serves the person token endpoint.
func (h *aauthHandler) HandlePersonToken(w http.ResponseWriter, r *http.Request) {
	agent, body, ok := h.authenticate(w, r)
	if !ok {
		return
	}
	var req personTokenRequest
	if perr := decodeJSONBody(r, body, &req); perr != nil {
		writeProblem(w, perr)
		return
	}
	outcome, perr := h.service.requestPersonToken(r.Context(), agent, &req)
	h.writeOutcome(w, outcome, perr)
}

// HandleAuthToken serves the auth token endpoint.
func (h *aauthHandler) HandleAuthToken(w http.ResponseWriter, r *http.Request) {
	agent, body, ok := h.authenticate(w, r)
	if !ok {
		return
	}
	var req authTokenRequest
	if perr := decodeJSONBody(r, body, &req); perr != nil {
		writeProblem(w, perr)
		return
	}
	outcome, perr := h.service.requestAuthToken(r.Context(), agent, &req)
	h.writeOutcome(w, outcome, perr)
}

// HandleMissionProposal serves POST {mission_endpoint}.
func (h *aauthHandler) HandleMissionProposal(w http.ResponseWriter, r *http.Request) {
	agent, body, ok := h.authenticate(w, r)
	if !ok {
		return
	}
	var proposal missionProposal
	if perr := decodeJSONBody(r, body, &proposal); perr != nil {
		writeProblem(w, perr)
		return
	}
	outcome, perr := h.service.proposeMission(r.Context(), agent, &proposal)
	h.writeOutcome(w, outcome, perr)
}

// HandleMission serves POST {mission_endpoint}/{mission_s256}.
func (h *aauthHandler) HandleMission(w http.ResponseWriter, r *http.Request) {
	agent, body, ok := h.authenticate(w, r)
	if !ok {
		return
	}
	var action missionAction
	if perr := decodeJSONBody(r, body, &action); perr != nil {
		writeProblem(w, perr)
		return
	}
	outcome, perr := h.service.missionRequest(r.Context(), agent, r.PathValue("s256"), &action)
	h.writeOutcome(w, outcome, perr)
}

// HandlePending serves GET on a pending URL.
func (h *aauthHandler) HandlePending(w http.ResponseWriter, r *http.Request) {
	agent, _, ok := h.authenticate(w, r)
	if !ok {
		return
	}
	outcome, perr := h.service.poll(r.Context(), agent, r.PathValue("id"))
	h.writeOutcome(w, outcome, perr)
}

// HandleInteraction serves the page the person arrives at with an interaction code.
func (h *aauthHandler) HandleInteraction(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		renderPage(w, http.StatusOK, &interactionView{})
		return
	}
	redirect, view := h.service.startInteraction(r.Context(), code)
	if view != nil {
		renderPage(w, http.StatusGone, view)
		return
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

// HandleInteractionCallback receives the person's sign-in and shows the consent page.
func (h *aauthHandler) HandleInteractionCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	view := h.service.completeSignIn(r.Context(), q.Get("state"), q.Get("code"), q.Get("error"))
	if view.DecisionToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name: decisionCookie, Value: view.DecisionToken, Path: pathInteraction,
			HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode, MaxAge: int(pendingLifetime.Seconds()),
		})
	}
	status := http.StatusOK
	if view.Error != "" {
		status = http.StatusBadRequest
	}
	renderPage(w, status, view)
}

// HandleInteractionDecision records the person's approve or deny decision.
func (h *aauthHandler) HandleInteractionDecision(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderPage(w, http.StatusBadRequest, &interactionView{Error: "The form could not be read."})
		return
	}
	token := r.PostFormValue("decision_token")
	cookie, err := r.Cookie(decisionCookie)
	if err != nil || cookie.Value != token {
		renderPage(w, http.StatusBadRequest,
			&interactionView{Error: "This decision was not made in the signed-in browser."})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: decisionCookie, Value: "", Path: pathInteraction, MaxAge: -1,
		HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
	view := h.service.decide(r.Context(), r.PostFormValue("pending_id"), token,
		r.PostFormValue("decision") == "approve")
	status := http.StatusOK
	if view.Error != "" {
		status = http.StatusBadRequest
	}
	renderPage(w, status, view)
}

// authenticate verifies the request signature and writes the 401 when it fails.
func (h *aauthHandler) authenticate(w http.ResponseWriter, r *http.Request) (*agentIdentity, []byte, bool) {
	agent, body, sigErr := h.service.authenticate(r)
	if sigErr != nil {
		h.logger.Debug(r.Context(), "AAuth request signature rejected",
			log.String("error", sigErr.code), log.String("detail", sigErr.detail))
		writeSignatureError(w, sigErr)
		return nil, nil, false
	}
	return agent, body, true
}

// writeOutcome writes a 200 result, a 202 deferred response, or a problem.
func (h *aauthHandler) writeOutcome(w http.ResponseWriter, outcome *tokenOutcome, perr *protocolError) {
	if perr != nil {
		writeProblem(w, perr)
		return
	}
	if outcome.pending == nil {
		if raw, ok := outcome.result.(json.RawMessage); ok {
			w.Header().Set("Content-Type", contentTypeJSON)
			w.Header().Set(headerCacheControl, "no-store")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(raw) // #nosec G705 -- a JSON body this server built, served as application/json
			return
		}
		w.Header().Set(headerCacheControl, "no-store")
		writeJSON(w, http.StatusOK, outcome.result)
		return
	}
	p := outcome.pending
	w.Header().Set(headerLocation, h.service.pendingURL(p.ID))
	w.Header().Set(headerRetryAfter, strconv.Itoa(pollInterval))
	w.Header().Set(headerCacheControl, "no-store")
	if p.State == statePending {
		w.Header().Set(headerAAuthRequirement, "requirement=interaction; url="+
			sfString(h.service.interactionURL())+"; code="+sfString(displayCode(p.Code)))
	}
	status := "pending"
	if p.State != statePending {
		status = "interacting"
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": status})
}

func decodeJSONBody(r *http.Request, body []byte, out any) *protocolError {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != contentTypeJSON {
		return badRequest(errInvalidRequest, "the body must be application/json")
	}
	if err := json.Unmarshal(body, out); err != nil {
		return badRequest(errInvalidRequest, "the body is not valid JSON")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeProblem(w http.ResponseWriter, perr *protocolError) {
	body := map[string]any{"error": perr.code}
	if perr.detail != "" {
		body["detail"] = perr.detail
	}
	for k, v := range perr.extra {
		if v != "" {
			body[k] = v
		}
	}
	w.Header().Set("Content-Type", contentTypeProblemJSON)
	w.Header().Set(headerCacheControl, "no-store")
	w.WriteHeader(perr.status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeSignatureError(w http.ResponseWriter, sigErr *sigError) {
	value := "error=" + sigErr.code
	if len(sigErr.requiredInput) > 0 {
		quoted := make([]string, 0, len(sigErr.requiredInput))
		for _, c := range sigErr.requiredInput {
			quoted = append(quoted, sfString(c))
		}
		value += ", required_input=(" + strings.Join(quoted, " ") + ")"
	}
	w.Header().Set(headerSignatureError, value)
	switch sigErr.code {
	case sigErrUnsupportedScheme:
		w.Header().Set(headerAcceptSignatureScheme, signatureKeySchemeJWT)
	case sigErrUnsupportedAlgorithm:
		w.Header().Set(headerAcceptSignatureAlg, acceptedSignatureAlgorithms)
	}
	w.Header().Set("Content-Type", contentTypeProblemJSON)
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   "urn:ietf:params:sig-error:" + sigErr.code,
		"status": http.StatusUnauthorized,
		"error":  sigErr.code,
		"detail": sigErr.detail,
	})
}

func renderPage(w http.ResponseWriter, status int, view *interactionView) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set(headerCacheControl, "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'")
	w.WriteHeader(status)
	_ = interactionPage.Execute(w, view)
}

// interactionPage renders the interaction and consent pages. Agent-asserted values are shown in a
// separate, labeled section from what the person server and the resource assert.
var interactionPage = template.Must(template.New("interaction").Parse(interactionHTML))

//go:embed interaction.html
var interactionHTML string
