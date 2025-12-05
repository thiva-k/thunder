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

package brandingmgt

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/asgardeo/thunder/internal/branding/common"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

const handlerLoggerComponentName = "BrandingMgtHandler"

// brandingMgtHandler is the handler for branding management operations.
type brandingMgtHandler struct {
	brandingMgtService BrandingMgtServiceInterface
	logger             *log.Logger
}

// newBrandingMgtHandler creates a new instance of brandingMgtHandler
func newBrandingMgtHandler(brandingMgtService BrandingMgtServiceInterface) *brandingMgtHandler {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLoggerComponentName))
	return &brandingMgtHandler{
		brandingMgtService: brandingMgtService,
		logger:             logger,
	}
}

// HandleBrandingListRequest handles the list branding configurations request.
func (bh *brandingMgtHandler) HandleBrandingListRequest(w http.ResponseWriter, r *http.Request) {
	limit, offset, svcErr := parsePaginationParams(r.URL.Query())
	if svcErr != nil {
		handleError(w, bh.logger, svcErr)
		return
	}

	brandingList, svcErr := bh.brandingMgtService.GetBrandingList(limit, offset)
	if svcErr != nil {
		handleError(w, bh.logger, svcErr)
		return
	}

	brandings := make([]BrandingListItem, 0, len(brandingList.Brandings))
	for _, branding := range brandingList.Brandings {
		brandings = append(brandings, BrandingListItem{
			ID:          branding.ID,
			DisplayName: branding.DisplayName,
		})
	}

	brandingListResponse := &BrandingListResponse{
		TotalResults: brandingList.TotalResults,
		StartIndex:   brandingList.StartIndex,
		Count:        brandingList.Count,
		Brandings:    brandings,
		Links:        toHTTPLinks(brandingList.Links),
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	isErr := writeToResponse(w, brandingListResponse, bh.logger)
	if isErr {
		return
	}

	bh.logger.Debug("Successfully listed branding configurations with pagination",
		log.Int("limit", limit), log.Int("offset", offset),
		log.Int("totalResults", brandingListResponse.TotalResults),
		log.Int("count", brandingListResponse.Count))
}

// HandleBrandingPostRequest handles the create branding configuration request.
func (bh *brandingMgtHandler) HandleBrandingPostRequest(w http.ResponseWriter, r *http.Request) {
	createRequest, err := sysutils.DecodeJSONBody[CreateBrandingRequest](r)
	if err != nil {
		handleError(w, bh.logger, &common.ErrorInvalidRequestFormat)
		return
	}

	createdBranding, svcErr := bh.brandingMgtService.CreateBranding(*createRequest)
	if svcErr != nil {
		handleError(w, bh.logger, svcErr)
		return
	}

	brandingResponse := common.BrandingResponse{
		ID:          createdBranding.ID,
		DisplayName: createdBranding.DisplayName,
		Preferences: createdBranding.Preferences,
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)

	isErr := writeToResponse(w, brandingResponse, bh.logger)
	if isErr {
		return
	}

	bh.logger.Debug("Successfully created branding configuration", log.String("id", createdBranding.ID))
}

// HandleBrandingGetRequest handles the get branding configuration request.
func (bh *brandingMgtHandler) HandleBrandingGetRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	branding, svcErr := bh.brandingMgtService.GetBranding(id)
	if svcErr != nil {
		handleError(w, bh.logger, svcErr)
		return
	}

	brandingResponse := common.BrandingResponse{
		ID:          branding.ID,
		DisplayName: branding.DisplayName,
		Preferences: branding.Preferences,
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	isErr := writeToResponse(w, brandingResponse, bh.logger)
	if isErr {
		return
	}

	bh.logger.Debug("Successfully retrieved branding configuration", log.String("id", id))
}

// HandleBrandingPutRequest handles the update branding configuration request.
func (bh *brandingMgtHandler) HandleBrandingPutRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	updateRequest, err := sysutils.DecodeJSONBody[UpdateBrandingRequest](r)
	if err != nil {
		handleError(w, bh.logger, &common.ErrorInvalidRequestFormat)
		return
	}

	updatedBranding, svcErr := bh.brandingMgtService.UpdateBranding(id, *updateRequest)
	if svcErr != nil {
		handleError(w, bh.logger, svcErr)
		return
	}

	brandingResponse := common.BrandingResponse{
		ID:          updatedBranding.ID,
		DisplayName: updatedBranding.DisplayName,
		Preferences: updatedBranding.Preferences,
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	isErr := writeToResponse(w, brandingResponse, bh.logger)
	if isErr {
		return
	}

	bh.logger.Debug("Successfully updated branding configuration", log.String("id", id))
}

// HandleBrandingDeleteRequest handles the delete branding configuration request.
func (bh *brandingMgtHandler) HandleBrandingDeleteRequest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	svcErr := bh.brandingMgtService.DeleteBranding(id)
	if svcErr != nil {
		handleError(w, bh.logger, svcErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	bh.logger.Debug("Successfully deleted branding configuration", log.String("id", id))
}

// parsePaginationParams parses limit and offset query parameters from the request.
func parsePaginationParams(query url.Values) (int, int, *serviceerror.ServiceError) {
	limit := 0
	offset := 0

	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
			return 0, 0, &common.ErrorInvalidLimit
		} else {
			limit = parsedLimit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err != nil {
			return 0, 0, &common.ErrorInvalidOffset
		} else {
			offset = parsedOffset
		}
	}

	if limit == 0 {
		limit = serverconst.DefaultPageSize
	}

	return limit, offset, nil
}

// toHTTPLinks converts service layer Links to HTTP LinkResponses.
func toHTTPLinks(links []Link) []LinkResponse {
	httpLinks := make([]LinkResponse, len(links))
	for i, link := range links {
		httpLinks[i] = LinkResponse(link)
	}
	return httpLinks
}

// writeToResponse encodes the response as JSON and writes it to the ResponseWriter.
func writeToResponse(w http.ResponseWriter, response any, logger *log.Logger) bool {
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Error encoding response", log.Error(err))
		handleEncodingError(w)
		return true
	}
	return false
}

// handleError handles service errors and returns appropriate HTTP responses.
func handleError(w http.ResponseWriter, logger *log.Logger,
	svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case common.ErrorBrandingNotFound.Code:
			statusCode = http.StatusNotFound
		case common.ErrorCannotDeleteBranding.Code:
			statusCode = http.StatusConflict
		case common.ErrorInvalidRequestFormat.Code, common.ErrorMissingBrandingID.Code,
			common.ErrorInvalidLimit.Code, common.ErrorInvalidOffset.Code,
			common.ErrorMissingPreferences.Code, common.ErrorInvalidPreferences.Code,
			common.ErrorMissingDisplayName.Code:
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusBadRequest
		}
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(statusCode)

	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Error encoding error response", log.Error(err))
		handleEncodingError(w)
		return
	}
}

// handleEncodingError handles errors that occur during response encoding.
func handleEncodingError(w http.ResponseWriter) {
	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusInternalServerError)
}
