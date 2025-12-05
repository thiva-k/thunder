/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

// Package brandingresolve provides functionality for resolving branding configurations.
package brandingresolve

import (
	"github.com/asgardeo/thunder/internal/application"
	"github.com/asgardeo/thunder/internal/branding/common"
	brandingmgt "github.com/asgardeo/thunder/internal/branding/mgt"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const serviceLogger = "BrandingResolveService"

// BrandingResolveServiceInterface defines the interface for the branding resolve service.
type BrandingResolveServiceInterface interface {
	ResolveBranding(
		resolveType common.BrandingResolveType, id string,
	) (*common.BrandingResponse, *serviceerror.ServiceError)
}

// brandingResolveService is the default implementation of the BrandingResolveServiceInterface.
type brandingResolveService struct {
	brandingMgtService brandingmgt.BrandingMgtServiceInterface
	applicationService application.ApplicationServiceInterface
	logger             *log.Logger
}

// newBrandingResolveService creates a new instance of BrandingResolveService with injected dependencies.
func newBrandingResolveService(
	brandingMgtService brandingmgt.BrandingMgtServiceInterface,
	applicationService application.ApplicationServiceInterface,
) BrandingResolveServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, serviceLogger))
	return &brandingResolveService{
		brandingMgtService: brandingMgtService,
		applicationService: applicationService,
		logger:             logger,
	}
}

// ResolveBranding resolves a branding configuration by type and ID.
// TODO: Add support for OU type and fallback logic.
func (brs *brandingResolveService) ResolveBranding(
	resolveType common.BrandingResolveType, id string,
) (*common.BrandingResponse, *serviceerror.ServiceError) {
	if resolveType == "" {
		return nil, &common.ErrorInvalidResolveType
	}

	if id == "" {
		return nil, &common.ErrorMissingResolveID
	}

	// Currently only APP type is supported
	if resolveType != common.BrandingResolveTypeAPP {
		return nil, &common.ErrorUnsupportedResolveType
	}

	// Get the application by ID
	if brs.applicationService == nil {
		brs.logger.Error("Application service is not available")
		return nil, &serviceerror.InternalServerError
	}

	app, svcErr := brs.applicationService.GetApplication(id)
	if svcErr != nil {
		// Convert application service errors to branding resolve errors
		if svcErr.Code == application.ErrorApplicationNotFound.Code ||
			svcErr.Code == application.ErrorInvalidApplicationID.Code {
			return nil, &common.ErrorApplicationNotFound
		}
		return nil, svcErr
	}

	// Check if the application has a branding ID
	if app.BrandingID == "" {
		return nil, &common.ErrorApplicationHasNoBranding
	}

	// Get the branding configuration
	brandingConfig, svcErr := brs.brandingMgtService.GetBranding(app.BrandingID)
	if svcErr != nil {
		if svcErr.Code == common.ErrorBrandingNotFound.Code {
			brs.logger.Error("Data integrity issue: application references non-existent branding",
				log.String("applicationId", id),
				log.String("brandingId", app.BrandingID))
			return nil, &serviceerror.InternalServerError
		}
		return nil, svcErr
	}

	brandingResponse := &common.BrandingResponse{
		ID:          brandingConfig.ID,
		DisplayName: brandingConfig.DisplayName,
		Preferences: brandingConfig.Preferences,
	}

	brs.logger.Debug("Successfully resolved branding configuration",
		log.String("type", string(resolveType)),
		log.String("id", id),
		log.String("brandingId", brandingConfig.ID))

	return brandingResponse, nil
}
