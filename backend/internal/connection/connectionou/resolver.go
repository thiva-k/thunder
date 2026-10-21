// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package connectionou

import (
	"context"

	"github.com/thunder-id/thunderid/internal/ou"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/internal/system/security"
	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
)

// ResolverInterface resolves and checks the organization unit a connection is owned by.
type ResolverInterface interface {
	Resolve(ctx context.Context, ouID, ouHandle string) (string, *tidcommon.ServiceError)
	ResolveForUpdate(
		ctx context.Context, requestedID, requestedHandle, currentID string,
	) (string, *tidcommon.ServiceError)
}

// resolver is the default implementation of ResolverInterface.
type resolver struct {
	ouService ou.OrganizationUnitServiceInterface
	errs      Errors
	logger    *log.Logger
}

// Resolve returns the id of the organization unit a connection is owned by. An id is used as given, and a handle is
// resolved to the id of the unit it names; when both are given the id wins. The unit has to exist. Only a lookup that
// finds no such unit is reported as not found, so a failed or refused lookup is not taken for one.
//
// A handle is resolved with the runtime context: it comes from an import or a declarative file, which name the owner
// by handle on behalf of the deployment and not of whoever runs the import.
func (r *resolver) Resolve(ctx context.Context, ouID, ouHandle string) (string, *tidcommon.ServiceError) {
	if ouID != "" && ouHandle != "" {
		r.logger.Warn(ctx, "Both ouId and ouHandle provided for a connection; ouHandle ignored")
	}

	if ouID == "" {
		if ouHandle == "" {
			return "", &r.errs.Required
		}
		unit, svcErr := r.ouService.GetOrganizationUnitByPath(security.WithRuntimeContext(ctx), ouHandle)
		if svcErr != nil {
			if svcErr.Code == ou.ErrorOrganizationUnitNotFound.Code {
				return "", &r.errs.NotFound
			}
			r.logger.Error(ctx, "Failed to look up the organization unit of a connection by handle",
				log.String("errorCode", svcErr.Code))
			return "", &tidcommon.InternalServerError
		}
		return unit.ID, nil
	}

	exists, svcErr := r.ouService.IsOrganizationUnitExists(ctx, ouID)
	if svcErr != nil {
		r.logger.Error(ctx, "Failed to look up the organization unit of a connection",
			log.String("errorCode", svcErr.Code))
		return "", &tidcommon.InternalServerError
	}
	if !exists {
		return "", &r.errs.NotFound
	}
	return ouID, nil
}

// ResolveForUpdate returns the owner an updated connection has. An owner cannot change once the connection is
// created, so a request that leaves the owner out keeps the current one, and one that names an owner has to name the
// one the connection has, by id or by handle.
func (r *resolver) ResolveForUpdate(
	ctx context.Context, requestedID, requestedHandle, currentID string,
) (string, *tidcommon.ServiceError) {
	if requestedID == "" && requestedHandle == "" {
		return currentID, nil
	}
	resolved := requestedID
	if resolved == "" {
		var svcErr *tidcommon.ServiceError
		if resolved, svcErr = r.Resolve(ctx, "", requestedHandle); svcErr != nil {
			return "", svcErr
		}
	}
	if resolved != currentID {
		return "", &r.errs.Changed
	}
	return currentID, nil
}
