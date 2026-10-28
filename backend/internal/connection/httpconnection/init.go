// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// Package httpconnection manages HTTP endpoint connections: a named outbound HTTP endpoint with its
// URL, timeout, plain headers and authentication, whose secrets are stored encrypted.
package httpconnection

import (
	"github.com/thunder-id/thunderid/internal/connection/connectionou"
	"github.com/thunder-id/thunderid/internal/ou"
)

// Initialize creates the HTTP connection service over the configuration database.
func Initialize(ouService ou.OrganizationUnitServiceInterface) HTTPConnectionServiceInterface {
	ouResolver := connectionou.Initialize(ouService, connectionou.Errors{
		Required: connectionou.ErrorOURequired,
		NotFound: connectionou.ErrorInvalidOU,
		Changed:  connectionou.ErrorOUUpdateNotAllowed,
	})
	return newHTTPConnectionService(newHTTPConnectionStore(), ouResolver)
}
