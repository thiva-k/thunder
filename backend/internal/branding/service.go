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

package branding

import (
	"encoding/json"
	"errors"
	"fmt"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "BrandingService"

// BrandingServiceInterface defines the interface for the branding service.
type BrandingServiceInterface interface {
	GetBrandingList(limit, offset int) (*BrandingList, *serviceerror.ServiceError)
	CreateBranding(branding CreateBrandingRequest) (*Branding, *serviceerror.ServiceError)
	GetBranding(id string) (*Branding, *serviceerror.ServiceError)
	UpdateBranding(id string, branding UpdateBrandingRequest) (*Branding, *serviceerror.ServiceError)
	DeleteBranding(id string) *serviceerror.ServiceError
	IsBrandingExist(id string) (bool, *serviceerror.ServiceError)
}

// brandingService is the default implementation of the BrandingServiceInterface.
type brandingService struct {
	brandingStore brandingStoreInterface
	logger        *log.Logger
}

// newBrandingService creates a new instance of BrandingService with injected dependencies.
func newBrandingService(brandingStore brandingStoreInterface) BrandingServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	return &brandingService{
		brandingStore: brandingStore,
		logger:        logger,
	}
}

// GetBrandingList retrieves a list of branding configurations.
func (bs *brandingService) GetBrandingList(limit, offset int) (*BrandingList, *serviceerror.ServiceError) {
	if err := validatePaginationParams(limit, offset); err != nil {
		return nil, err
	}

	totalCount, err := bs.brandingStore.GetBrandingListCount()
	if err != nil {
		bs.logger.Error("Failed to get branding count", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	brandings, err := bs.brandingStore.GetBrandingList(limit, offset)
	if err != nil {
		bs.logger.Error("Failed to list brandings", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	response := &BrandingList{
		TotalResults: totalCount,
		Brandings:    brandings,
		StartIndex:   offset + 1,
		Count:        len(brandings),
		Links:        buildPaginationLinks(limit, offset, totalCount),
	}

	return response, nil
}

// CreateBranding creates a new branding configuration.
func (bs *brandingService) CreateBranding(
	branding CreateBrandingRequest,
) (*Branding, *serviceerror.ServiceError) {
	bs.logger.Debug("Creating branding configuration")

	if branding.DisplayName == "" {
		return nil, &ErrorMissingDisplayName
	}

	if err := bs.validateBrandingPreferences(branding.Preferences); err != nil {
		return nil, err
	}

	id := utils.GenerateUUID()
	if err := bs.brandingStore.CreateBranding(id, branding); err != nil {
		bs.logger.Error("Failed to create branding", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	createdBranding := &Branding{
		ID:          id,
		DisplayName: branding.DisplayName,
		Preferences: branding.Preferences,
	}

	bs.logger.Debug("Successfully created branding", log.String("id", id))
	return createdBranding, nil
}

// GetBranding retrieves a specific branding configuration by its id.
func (bs *brandingService) GetBranding(id string) (*Branding, *serviceerror.ServiceError) {
	bs.logger.Debug("Retrieving branding", log.String("id", id))

	if id == "" {
		return nil, &ErrorMissingBrandingID
	}

	branding, err := bs.brandingStore.GetBranding(id)
	if err != nil {
		if errors.Is(err, ErrBrandingNotFound) {
			bs.logger.Debug("Branding not found", log.String("id", id))
			return nil, &ErrorBrandingNotFound
		}
		bs.logger.Error("Failed to retrieve branding", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	bs.logger.Debug("Successfully retrieved branding", log.String("id", branding.ID))
	return &branding, nil
}

// UpdateBranding updates an existing branding configuration.
func (bs *brandingService) UpdateBranding(
	id string, branding UpdateBrandingRequest) (*Branding, *serviceerror.ServiceError) {
	bs.logger.Debug("Updating branding", log.String("id", id))

	if id == "" {
		return nil, &ErrorMissingBrandingID
	}

	if branding.DisplayName == "" {
		return nil, &ErrorMissingDisplayName
	}

	if err := bs.validateBrandingPreferences(branding.Preferences); err != nil {
		return nil, err
	}

	if err := bs.brandingStore.UpdateBranding(id, branding); err != nil {
		if errors.Is(err, ErrBrandingNotFound) {
			bs.logger.Debug("Branding not found for update", log.String("id", id))
			return nil, &ErrorBrandingNotFound
		}
		bs.logger.Error("Failed to update branding", log.String("id", id), log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	updatedBranding := &Branding{
		ID:          id,
		DisplayName: branding.DisplayName,
		Preferences: branding.Preferences,
	}

	bs.logger.Debug("Successfully updated branding", log.String("id", id))
	return updatedBranding, nil
}

// DeleteBranding deletes the specified branding configuration by its id.
func (bs *brandingService) DeleteBranding(id string) *serviceerror.ServiceError {
	bs.logger.Debug("Deleting branding", log.String("id", id))

	if id == "" {
		return &ErrorMissingBrandingID
	}

	exists, err := bs.brandingStore.IsBrandingExist(id)
	if err != nil {
		bs.logger.Error("Failed to check branding existence", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}
	if !exists {
		bs.logger.Debug("Branding not found", log.String("id", id))
		return nil
	}

	// Check if branding has any applications before deleting
	appCount, err := bs.brandingStore.GetApplicationsCountByBrandingID(id)
	if err != nil {
		bs.logger.Error("Failed to get applications count", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}
	if appCount > 0 {
		bs.logger.Debug("Cannot delete branding with active applications",
			log.String("id", id), log.Int("appCount", appCount))
		return &ErrorCannotDeleteBranding
	}

	if err := bs.brandingStore.DeleteBranding(id); err != nil {
		bs.logger.Error("Failed to delete branding", log.String("id", id), log.Error(err))
		return &serviceerror.InternalServerError
	}

	bs.logger.Debug("Successfully deleted branding", log.String("id", id))
	return nil
}

// validateBrandingPreferences validates the branding preferences JSON.
func (bs *brandingService) validateBrandingPreferences(preferences json.RawMessage) *serviceerror.ServiceError {
	if len(preferences) == 0 {
		return &ErrorMissingPreferences
	}

	// Validate that preferences is valid JSON
	var test interface{}
	if err := json.Unmarshal(preferences, &test); err != nil {
		return &ErrorInvalidPreferences
	}

	// Ensure it's an object, not an array or primitive
	if _, ok := test.(map[string]interface{}); !ok {
		return &ErrorInvalidPreferences
	}

	// TODO: Add additional validations against the JSON Schema
	return nil
}

// validatePaginationParams validates pagination parameters.
func validatePaginationParams(limit, offset int) *serviceerror.ServiceError {
	if limit < 1 || limit > serverconst.MaxPageSize {
		return &ErrorInvalidLimit
	}
	if offset < 0 {
		return &ErrorInvalidOffset
	}
	return nil
}

// buildPaginationLinks builds pagination links for the response.
func buildPaginationLinks(limit, offset, totalCount int) []Link {
	const brandingBasePath = "/branding"
	links := make([]Link, 0)

	if offset > 0 {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=0&limit=%d", brandingBasePath, limit),
			Rel:  "first",
		})

		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", brandingBasePath, prevOffset, limit),
			Rel:  "prev",
		})
	}

	if offset+limit < totalCount {
		nextOffset := offset + limit
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", brandingBasePath, nextOffset, limit),
			Rel:  "next",
		})
	}

	lastPageOffset := ((totalCount - 1) / limit) * limit
	if offset < lastPageOffset {
		links = append(links, Link{
			Href: fmt.Sprintf("%s?offset=%d&limit=%d", brandingBasePath, lastPageOffset, limit),
			Rel:  "last",
		})
	}

	return links
}

// IsBrandingExist checks if a branding configuration exists by its ID.
func (bs *brandingService) IsBrandingExist(id string) (bool, *serviceerror.ServiceError) {
	if id == "" {
		return false, nil
	}

	exists, err := bs.brandingStore.IsBrandingExist(id)
	if err != nil {
		bs.logger.Error("Failed to check branding existence", log.String("id", id), log.Error(err))
		return false, &serviceerror.InternalServerError
	}

	return exists, nil
}
