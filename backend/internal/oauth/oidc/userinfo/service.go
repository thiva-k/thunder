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
	"strings"

	"github.com/asgardeo/thunder/internal/application"
	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oidc/claims"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/jwt"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

// UserInfoServiceInterface defines the interface for the OIDC UserInfo service.
type UserInfoServiceInterface interface {
	GetUserInfo(accessToken string) (UserInfoResponse, *serviceerror.ServiceError)
}

// userInfoService implements the UserInfoServiceInterface.
type userInfoService struct {
	jwtService         jwt.JWTServiceInterface
	userService        user.UserServiceInterface
	applicationService application.ApplicationServiceInterface
}

// newUserInfoService creates a new userInfoService instance.
func newUserInfoService(
	jwtService jwt.JWTServiceInterface,
	userService user.UserServiceInterface,
	applicationService application.ApplicationServiceInterface,
) UserInfoServiceInterface {
	return &userInfoService{
		jwtService:         jwtService,
		userService:        userService,
		applicationService: applicationService,
	}
}

// GetUserInfo validates the access token and returns OIDC-compliant user claims.
func (s *userInfoService) GetUserInfo(accessToken string) (UserInfoResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "UserInfoService"))

	if accessToken == "" {
		return nil, &ErrorMissingToken
	}

	if err := s.jwtService.VerifyJWT(accessToken, "", ""); err != nil {
		logger.Debug("Failed to verify access token", log.Error(err))
		return nil, &ErrorInvalidTokenSignature
	}

	_, payload, err := jwt.DecodeJWT(accessToken)
	if err != nil {
		logger.Debug("Failed to decode access token", log.Error(err))
		return nil, &ErrorInvalidTokenFormat
	}

	sub, ok := payload["sub"].(string)
	if !ok || sub == "" {
		logger.Debug("Missing or invalid 'sub' claim in access token")
		return nil, &ErrorMissingSubject
	}

	scopeString, _ := payload["scope"].(string)
	var scopes []string
	if scopeString != "" {
		scopes = strings.Split(scopeString, " ")
	}

	hasOpenIDScope := false
	for _, scope := range scopes {
		if scope == "openid" {
			hasOpenIDScope = true
			break
		}
	}
	if !hasOpenIDScope {
		logger.Debug("Access token does not contain 'openid' scope")
		return nil, &ErrorInsufficientScope
	}

	clientID, _ := payload["aud"].(string)

	user, svcErr := s.userService.GetUser(sub)
	if svcErr != nil {
		logger.Error("Failed to fetch user", log.String("sub", sub), log.Any("error", svcErr))
		return nil, &ErrorUserNotFound
	}

	var userAttributes map[string]interface{}
	if err := json.Unmarshal(user.Attributes, &userAttributes); err != nil {
		logger.Error("Failed to unmarshal user attributes", log.String("sub", sub), log.Error(err))
		return nil, &ErrorUserAttributesProcessing
	}

	var oauthApp *appmodel.OAuthAppConfigProcessedDTO
	if clientID != "" {
		app, svcErr := s.applicationService.GetOAuthApplication(clientID)
		if svcErr != nil {
			logger.Debug("Could not fetch application config, using standard scope claims",
				log.String("clientID", clientID), log.Any("error", svcErr))
		} else {
			oauthApp = app
		}
	}

	userClaims := claims.BuildUserClaims(scopes, userAttributes, oauthApp, false)
	userClaims["sub"] = sub

	logger.Debug("UserInfo response prepared", log.String("sub", sub), log.Int("claimCount", len(userClaims)))

	return userClaims, nil
}
