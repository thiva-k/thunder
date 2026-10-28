// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// Package connectionou resolves and checks the organization unit a connection is owned by. Every kind of connection
// (identity provider, notification sender, AuthZEN PDP and HTTP) is owned by exactly one, and each kind's service asks
// this package the same question: given the id and the handle a request or document carries, which unit is it.
package connectionou

import (
	"github.com/thunder-id/thunderid/internal/ou"
	"github.com/thunder-id/thunderid/internal/system/log"
)

// Initialize creates the resolver of a kind of connection. The resolver answers with the errors the kind gives, so
// each kind keeps its own error codes.
func Initialize(ouService ou.OrganizationUnitServiceInterface, errs Errors) ResolverInterface {
	return &resolver{
		ouService: ouService,
		errs:      errs,
		logger:    log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ConnectionOUResolver")),
	}
}
