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

package brandingresolve

import (
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/branding/common"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

const handlerLogger = "BrandingResolveHandler"

// brandingResolveHandler is the handler for branding resolve operations.
type brandingResolveHandler struct {
	resolveService BrandingResolveServiceInterface
	logger         *log.Logger
}

// newBrandingResolveHandler creates a new instance of brandingResolveHandler.
func newBrandingResolveHandler(resolveService BrandingResolveServiceInterface) *brandingResolveHandler {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, handlerLogger))
	return &brandingResolveHandler{
		resolveService: resolveService,
		logger:         logger,
	}
}

// HandleResolveRequest handles the resolve branding configuration request.
func (rh *brandingResolveHandler) HandleResolveRequest(w http.ResponseWriter, r *http.Request) {
	resolveType := common.BrandingResolveType(strings.ToUpper(r.URL.Query().Get("type")))
	id := r.URL.Query().Get("id")

	brandingResponse, svcErr := rh.resolveService.ResolveBranding(resolveType, id)
	if svcErr != nil {
		rh.handleError(w, svcErr)
		return
	}

	utils.WriteSuccessResponse(w, http.StatusOK, brandingResponse)

	rh.logger.Debug("Successfully resolved branding configuration",
		log.String("type", string(resolveType)),
		log.String("id", id),
		log.String("brandingId", brandingResponse.ID))
}

// handleError handles service errors and returns appropriate HTTP responses.
func (rh *brandingResolveHandler) handleError(w http.ResponseWriter, svcErr *serviceerror.ServiceError) {
	statusCode := http.StatusInternalServerError
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case common.ErrorInvalidResolveType.Code,
			common.ErrorMissingResolveID.Code,
			common.ErrorUnsupportedResolveType.Code:
			statusCode = http.StatusBadRequest
		case common.ErrorApplicationHasNoBranding.Code,
			common.ErrorApplicationNotFound.Code:
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusBadRequest
		}
	}

	errResp := apierror.ErrorResponse{
		Code:        svcErr.Code,
		Message:     svcErr.Error,
		Description: svcErr.ErrorDescription,
	}

	utils.WriteErrorResponse(w, statusCode, errResp)
}
