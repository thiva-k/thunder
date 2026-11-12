// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import "net/http"

// AAuth error codes carried in application/problem+json bodies.
const (
	errInvalidRequest        = "invalid_request"
	errInvalidResourceToken  = "invalid_resource_token"
	errExpiredResourceToken  = "expired_resource_token"
	errInvalidPresentedToken = "invalid_presented_token"
	errExpiredPresentedToken = "expired_presented_token"
	errInvalidUpstreamToken  = "invalid_upstream_token"
	errInvalidSubagentToken  = "invalid_subagent_token"
	errExpiredSubagentToken  = "expired_subagent_token"
	errUserUnreachable       = "user_unreachable"
	errServerError           = "server_error"
	errDenied                = "denied"
	errAbandoned             = "abandoned"
	errExpired               = "expired"
	errNotFound              = "not_found"
	errMissionNotFound       = "mission_not_found"
	errMissionTerminated     = "mission_terminated"
)

func badRequest(code, detail string) *protocolError {
	return &protocolError{status: http.StatusBadRequest, code: code, detail: detail}
}

func serverError() *protocolError {
	return &protocolError{status: http.StatusInternalServerError, code: errServerError}
}
