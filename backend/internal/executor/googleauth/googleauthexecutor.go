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

// Package googleauth provides the Google OIDC authentication executor.
package googleauth

import (
	"fmt"
	"time"

	"github.com/asgardeo/thunder/internal/executor/oauth/model"
	"github.com/asgardeo/thunder/internal/executor/oidcauth"
	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	jwtutils "github.com/asgardeo/thunder/internal/system/crypto/jwt/utils"
	"github.com/asgardeo/thunder/internal/system/log"
)

const loggerComponentName = "GoogleOIDCAuthExecutor"

// GoogleOIDCAuthExecutor implements the OIDC authentication executor for Google.
type GoogleOIDCAuthExecutor struct {
	*oidcauth.OIDCAuthExecutor
}

var _ flowmodel.ExecutorInterface = (*GoogleOIDCAuthExecutor)(nil)

// NewGoogleOIDCAuthExecutorFromProps creates a new instance of GoogleOIDCAuthExecutor with the provided properties.
func NewGoogleOIDCAuthExecutorFromProps(execProps flowmodel.ExecutorProperties,
	oAuthProps *model.BasicOAuthExecProperties) oidcauth.OIDCAuthExecutorInterface {
	// Prepare the complete OAuth properties for Google
	compOAuthProps := &model.OAuthExecProperties{
		AuthorizationEndpoint: googleAuthorizeEndpoint,
		TokenEndpoint:         googleTokenEndpoint,
		UserInfoEndpoint:      googleUserInfoEndpoint,
		JwksEndpoint:          googleJwksEndpoint,
		ClientID:              oAuthProps.ClientID,
		ClientSecret:          oAuthProps.ClientSecret,
		RedirectURI:           oAuthProps.RedirectURI,
		Scopes:                oAuthProps.Scopes,
		AdditionalParams:      oAuthProps.AdditionalParams,
	}

	defaultInputs := []flowmodel.InputData{
		{
			Name:     "code",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "nonce",
			Type:     "string",
			Required: false,
		},
	}

	base := oidcauth.NewOIDCAuthExecutor("google_oidc_auth_executor", execProps.Name,
		defaultInputs, execProps.Properties, compOAuthProps)
	exec, ok := base.(*oidcauth.OIDCAuthExecutor)
	if !ok {
		panic("failed to cast GoogleOIDCAuthExecutor to OIDCAuthExecutor")
	}
	return &GoogleOIDCAuthExecutor{
		OIDCAuthExecutor: exec,
	}
}

// NewGoogleOIDCAuthExecutor creates a new instance of GoogleOIDCAuthExecutor with the provided details.
func NewGoogleOIDCAuthExecutor(id, name string, properties map[string]string,
	clientID, clientSecret, redirectURI string, scopes []string,
	additionalParams map[string]string) oidcauth.OIDCAuthExecutorInterface {
	// Prepare the OAuth properties for Google
	oAuthProps := &model.OAuthExecProperties{
		AuthorizationEndpoint: googleAuthorizeEndpoint,
		TokenEndpoint:         googleTokenEndpoint,
		UserInfoEndpoint:      googleUserInfoEndpoint,
		JwksEndpoint:          googleJwksEndpoint,
		ClientID:              clientID,
		ClientSecret:          clientSecret,
		RedirectURI:           redirectURI,
		Scopes:                scopes,
		AdditionalParams:      additionalParams,
	}

	defaultInputs := []flowmodel.InputData{
		{
			Name:     "code",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "nonce",
			Type:     "string",
			Required: false,
		},
	}

	base := oidcauth.NewOIDCAuthExecutor(id, name, defaultInputs, properties, oAuthProps)
	exec, ok := base.(*oidcauth.OIDCAuthExecutor)
	if !ok {
		panic("failed to cast GoogleOIDCAuthExecutor to OIDCAuthExecutor")
	}
	return &GoogleOIDCAuthExecutor{
		OIDCAuthExecutor: exec,
	}
}

// Execute executes the Google OIDC authentication flow.
func (g *GoogleOIDCAuthExecutor) Execute(ctx *flowmodel.NodeContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Executing Google OIDC auth executor",
		log.String("executorID", g.GetID()), log.String("flowID", ctx.FlowID))

	execResp := &flowmodel.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if g.CheckInputData(ctx, execResp) {
		logger.Debug("Required input data for Google OIDC auth executor is not provided")
		err := g.BuildAuthorizeFlow(ctx, execResp)
		if err != nil {
			return nil, err
		}

		logger.Debug("Google OIDC auth executor execution completed",
			log.String("status", string(execResp.Status)))
	} else {
		err := g.ProcessAuthFlowResponse(ctx, execResp)
		if err != nil {
			return nil, err
		}

		logger.Debug("Google OIDC auth executor execution completed",
			log.String("status", string(execResp.Status)),
			log.Bool("isAuthenticated", execResp.AuthenticatedUser.IsAuthenticated))
	}

	return execResp, nil
}

// ValidateIDToken validates the ID token received from Google.
func (g *GoogleOIDCAuthExecutor) ValidateIDToken(execResp *flowmodel.ExecutorResponse, idToken string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Validating ID token")

	if g.GetJWKSEndpoint() == "" {
		return fmt.Errorf("JWKS endpoint is not configured for Google OIDC executor")
	}

	// Verify the id token signature.
	signErr := jwtutils.VerifyJWTSignatureWithJWKS(idToken, g.GetJWKSEndpoint())
	if signErr != nil {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "ID token signature verification failed: " + signErr.Error()
		return nil
	}

	// Parse the JWT claims from the ID token.
	claims, err := jwtutils.ParseJWTClaims(idToken)
	if err != nil {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Failed to parse ID token claims: " + err.Error()
		return nil
	}

	// Validate issuer
	iss, ok := claims["iss"].(string)
	if !ok || (iss != "accounts.google.com" && iss != "https://accounts.google.com") {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = fmt.Sprintf("Invalid issuer: %s in the id token", iss)
		return nil
	}

	// Validate audience
	aud, ok := claims["aud"].(string)
	if !ok || aud != g.GetOAuthProperties().ClientID {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = fmt.Sprintf("Invalid audience: %s in the id token", aud)
		return nil
	}

	// Validate expiration time
	exp, ok := claims["exp"].(float64)
	if !ok {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Missing expiration claim in the id token"
		return nil
	}
	if time.Now().Unix() >= int64(exp) {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "ID token has expired"
		return nil
	}

	// Check if token was issued in the future (to prevent clock skew issues)
	iat, ok := claims["iat"].(float64)
	if !ok {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "Missing issued at claim in the id token"
		return nil
	}
	if time.Now().Unix() < int64(iat) {
		execResp.Status = flowconst.ExecFailure
		execResp.FailureReason = "ID token was issued in the future"
		return nil
	}

	// Check for specific domain if configured in additional params
	if hd, found := claims["hd"]; found {
		if domain, exists := g.GetOAuthProperties().AdditionalParams["hd"]; exists && domain != "" {
			if hdStr, ok := hd.(string); !ok || hdStr != domain {
				execResp.Status = flowconst.ExecFailure
				execResp.FailureReason = fmt.Sprintf("ID token is not from the expected hosted domain: %s", domain)
				return nil
			}
		}
	}

	return nil
}
