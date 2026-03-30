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

package granthandlers

import (
	"context"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/tokenservice"
	"github.com/asgardeo/thunder/internal/role"
	"github.com/asgardeo/thunder/internal/system/log"
)

// clientCredentialsGrantHandler handles the client credentials grant type.
type clientCredentialsGrantHandler struct {
	tokenBuilder tokenservice.TokenBuilderInterface
	roleService  role.RoleServiceInterface
}

// newClientCredentialsGrantHandler creates a new instance of ClientCredentialsGrantHandler.
func newClientCredentialsGrantHandler(
	tokenBuilder tokenservice.TokenBuilderInterface,
	roleService role.RoleServiceInterface,
) GrantHandlerInterface {
	return &clientCredentialsGrantHandler{
		tokenBuilder: tokenBuilder,
		roleService:  roleService,
	}
}

// ValidateGrant validates the client credentials grant type.
func (h *clientCredentialsGrantHandler) ValidateGrant(ctx context.Context, tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) *model.ErrorResponse {
	if constants.GrantType(tokenRequest.GrantType) != constants.GrantTypeClientCredentials {
		return &model.ErrorResponse{
			Error:            constants.ErrorUnsupportedGrantType,
			ErrorDescription: "Unsupported grant type",
		}
	}

	return nil
}

// HandleGrant handles the client credentials grant type.
// Scopes are filtered based on the application's role-based permissions.
func (h *clientCredentialsGrantHandler) HandleGrant(ctx context.Context, tokenRequest *model.TokenRequest,
	oauthApp *appmodel.OAuthAppConfigProcessedDTO) (
	*model.TokenResponseDTO, *model.ErrorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ClientCredentialsGrant"))

	scopes := tokenservice.ParseScopes(tokenRequest.Scope)

	// Filter requested scopes based on the app's authorized permissions via RBAC.
	// The app entity ID is the same as the app ID.
	if h.roleService != nil && len(scopes) > 0 {
		authorizedPerms, svcErr := h.roleService.GetAuthorizedPermissions(
			ctx, oauthApp.AppID, nil, scopes)
		if svcErr != nil {
			logger.Warn("Failed to resolve authorized permissions for app, proceeding with requested scopes",
				log.String("appID", oauthApp.AppID), log.String("error", svcErr.Error))
		} else {
			scopes = authorizedPerms
		}
	}

	finalAudience := tokenservice.DetermineAudience("", tokenRequest.Resource, "", tokenRequest.ClientID)

	accessToken, err := h.tokenBuilder.BuildAccessToken(&tokenservice.AccessTokenBuildContext{
		Subject:        tokenRequest.ClientID,
		Audience:       finalAudience,
		ClientID:       tokenRequest.ClientID,
		Scopes:         scopes,
		UserAttributes: make(map[string]interface{}),
		GrantType:      string(constants.GrantTypeClientCredentials),
		OAuthApp:       oauthApp,
	})
	if err != nil {
		return nil, &model.ErrorResponse{
			Error:            constants.ErrorServerError,
			ErrorDescription: "Failed to generate token",
		}
	}

	return &model.TokenResponseDTO{
		AccessToken: *accessToken,
	}, nil
}
