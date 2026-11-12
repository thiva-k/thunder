// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	handler *aauthHandler
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (s *HandlerTestSuite) SetupTest() {
	s.handler = newAAuthHandler(&aauthService{issuer: testIssuer})
}

func (s *HandlerTestSuite) TestPendingResponse() {
	w := httptest.NewRecorder()
	s.handler.writeOutcome(w, &tokenOutcome{pending: &pendingRequest{ID: "p1", Code: "ABCDEFGH", State: statePending,
		ExpiresAt: time.Now().Add(time.Minute)}}, nil)
	s.Equal(http.StatusAccepted, w.Code)
	s.Equal(testIssuer+"/aauth/pending/p1", w.Header().Get("Location"))
	s.Equal("1", w.Header().Get("Retry-After"))
	s.Equal("no-store", w.Header().Get("Cache-Control"))
	s.Equal(`requirement=interaction; url="https://ps.example/aauth/interaction"; code="ABCD-EFGH"`,
		w.Header().Get("AAuth-Requirement"))
	s.JSONEq(`{"status":"pending"}`, w.Body.String())

	w = httptest.NewRecorder()
	s.handler.writeOutcome(w, &tokenOutcome{pending: &pendingRequest{ID: "p1", State: stateInteracting}}, nil)
	s.Empty(w.Header().Get("AAuth-Requirement"))
	s.JSONEq(`{"status":"interacting"}`, w.Body.String())
}

func (s *HandlerTestSuite) TestResultAndProblem() {
	w := httptest.NewRecorder()
	s.handler.writeOutcome(w, &tokenOutcome{result: json.RawMessage(`{"person_token":"t","expires_in":5}`)}, nil)
	s.Equal(http.StatusOK, w.Code)
	s.JSONEq(`{"person_token":"t","expires_in":5}`, w.Body.String())

	w = httptest.NewRecorder()
	s.handler.writeOutcome(w, nil, &protocolError{status: http.StatusForbidden, code: errMissionTerminated,
		extra: map[string]any{"termination_reason": "completed"}})
	s.Equal(http.StatusForbidden, w.Code)
	s.Equal(contentTypeProblemJSON, w.Header().Get("Content-Type"))
	s.JSONEq(`{"error":"mission_terminated","termination_reason":"completed"}`, w.Body.String())
}

func (s *HandlerTestSuite) TestSignatureError() {
	w := httptest.NewRecorder()
	writeSignatureError(w, &sigError{code: sigErrInvalidInput, requiredInput: []string{"@method", "content-digest"}})
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Equal(`error=invalid_input, required_input=("@method" "content-digest")`, w.Header().Get("Signature-Error"))

	w = httptest.NewRecorder()
	writeSignatureError(w, &sigError{code: sigErrUnsupportedScheme})
	s.Equal("jwt", w.Header().Get("Accept-Signature-Scheme"))
	s.Contains(w.Body.String(), "urn:ietf:params:sig-error:unsupported_scheme")
}

func (s *HandlerTestSuite) TestDecodeJSONBody() {
	r := httptest.NewRequest(http.MethodPost, "/aauth/person", nil)
	r.Header.Set("Content-Type", "text/plain")
	var out personTokenRequest
	s.Equal(errInvalidRequest, decodeJSONBody(r, []byte(`{}`), &out).code)
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	s.Equal(errInvalidRequest, decodeJSONBody(r, []byte(`{`), &out).code)
	s.Nil(decodeJSONBody(r, []byte(`{"resource":"https://r.example"}`), &out))
	s.Equal("https://r.example", out.Resource)
}

func (s *HandlerTestSuite) TestConsentPageEscapesAgentContent() {
	w := httptest.NewRecorder()
	renderPage(w, http.StatusOK, &interactionView{Kind: kindPersonToken, DecisionToken: "d", PendingID: "p",
		AgentID: "aauth:a@agent.example", Justification: "<script>alert(1)</script>", Resource: testResource})
	body := w.Body.String()
	s.NotContains(body, "<script>alert(1)</script>")
	s.Contains(body, "Stated by the agent, not verified")
	s.True(strings.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'none'"))
}

func (s *HandlerTestSuite) TestDecisionRequiresTheSigninCookie() {
	r := httptest.NewRequest(http.MethodPost, "/aauth/interaction/decision",
		strings.NewReader("pending_id=p&decision_token=t&decision=approve"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.handler.HandleInteractionDecision(w, r)
	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(w.Body.String(), "signed-in browser")
}
