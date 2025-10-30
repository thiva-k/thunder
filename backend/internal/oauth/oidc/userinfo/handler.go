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

package userinfo

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

// userInfoHandler handles OIDC UserInfo requests.
type userInfoHandler struct {
	service UserInfoServiceInterface
}

// newUserInfoHandler creates a new userInfo handler.
func newUserInfoHandler(userInfoService UserInfoServiceInterface) *userInfoHandler {
	return &userInfoHandler{
		service: userInfoService,
	}
}

// HandleUserInfo handles the UserInfo endpoint request.
func (h *userInfoHandler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserInfoHandler"))

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logger.Debug("Missing Authorization header")
		h.handleError(w, logger, &ErrorMissingToken)
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		logger.Debug("Invalid Authorization header format")
		h.handleError(w, logger, &ErrorMissingToken)
		return
	}

	accessToken := parts[1]
	if accessToken == "" {
		logger.Debug("Empty access token")
		h.handleError(w, logger, &ErrorMissingToken)
		return
	}

	userInfo, svcErr := h.service.GetUserInfo(accessToken)
	if svcErr != nil {
		h.handleError(w, logger, svcErr)
		return
	}

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userInfo); err != nil {
		logger.Error("Error encoding UserInfo response", log.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	logger.Debug("UserInfo response sent successfully")
}

// handleError handles service errors and writes appropriate HTTP responses.
func (h *userInfoHandler) handleError(w http.ResponseWriter, logger *log.Logger,
	svcErr *serviceerror.ServiceError) {

	w.Header().Set(serverconst.ContentTypeHeaderName, serverconst.ContentTypeJSON)

	var statusCode int
	if svcErr.Type == serviceerror.ClientErrorType {
		switch svcErr.Code {
		case ErrorInsufficientScope.Code:
			statusCode = http.StatusForbidden
		default:
			statusCode = http.StatusUnauthorized
		}
	} else {
		statusCode = http.StatusInternalServerError
	}

	w.WriteHeader(statusCode)

	errResp := model.ErrorResponse{
		Error:            svcErr.Code,
		ErrorDescription: svcErr.ErrorDescription,
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		logger.Error("Error encoding error response", log.Error(err))
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}
