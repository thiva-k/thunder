/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

// Package authnprovider provides authentication provider implementations.
package authnprovider

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"

	"github.com/asgardeo/thunder/internal/entity"
	"github.com/asgardeo/thunder/internal/system/crypto/hash"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/user"
)

type defaultAuthnProvider struct {
	userSvc       user.UserServiceInterface
	entityService entity.EntityServiceInterface
}

// newDefaultAuthnProvider creates a new internal authn provider.
func newDefaultAuthnProvider(
	userSvc user.UserServiceInterface,
	entityService entity.EntityServiceInterface,
) AuthnProviderInterface {
	return &defaultAuthnProvider{
		userSvc:       userSvc,
		entityService: entityService,
	}
}

// Authenticate authenticates an entity (user or application) based on the provided identifiers.
// If identifiers contain "clientId", it authenticates as an application via the entity service.
// Otherwise, it delegates to the user service for user authentication.
func (p *defaultAuthnProvider) Authenticate(
	ctx context.Context,
	identifiers, credentials map[string]interface{},
	metadata *AuthnMetadata,
) (*AuthnResult, *AuthnProviderError) {
	// Route to app authentication if clientId is present.
	if _, hasClientID := identifiers["clientId"]; hasClientID {
		return p.authenticateApp(ctx, identifiers, credentials)
	}

	// Default: user authentication via user service.
	return p.authenticateUser(ctx, identifiers, credentials)
}

// authenticateUser authenticates a user using the internal user service.
func (p *defaultAuthnProvider) authenticateUser(
	ctx context.Context,
	identifiers, credentials map[string]interface{},
) (*AuthnResult, *AuthnProviderError) {
	authResponse, authErr := p.userSvc.AuthenticateUser(ctx, identifiers, credentials)
	if authErr != nil {
		if authErr.Type == serviceerror.ClientErrorType {
			if authErr.Code == user.ErrorUserNotFound.Code {
				return nil, NewError(ErrorCodeUserNotFound, authErr.Error, authErr.ErrorDescription)
			}
			return nil, NewError(ErrorCodeAuthenticationFailed, authErr.Error, authErr.ErrorDescription)
		}
		return nil, NewError(ErrorCodeSystemError, authErr.Error, authErr.ErrorDescription)
	}

	userResult, getUserErr := p.userSvc.GetUser(ctx, authResponse.ID, false)
	if getUserErr != nil {
		if getUserErr.Code == user.ErrorUserNotFound.Code {
			return nil, NewError(ErrorCodeUserNotFound, getUserErr.Error, getUserErr.ErrorDescription)
		}
		return nil, NewError(ErrorCodeSystemError, getUserErr.Error, getUserErr.ErrorDescription)
	}

	var attributes map[string]interface{}
	if len(userResult.Attributes) > 0 {
		if err := json.Unmarshal(userResult.Attributes, &attributes); err != nil {
			return nil, NewError(ErrorCodeSystemError, "Failed to get allowed attributes", err.Error())
		}
	}

	availableAttributes := &AvailableAttributes{
		Attributes:    make(map[string]*AttributeMetadataResponse),
		Verifications: make(map[string]*VerificationResponse),
	}
	for k := range attributes {
		availableAttributes.Attributes[k] = &AttributeMetadataResponse{
			AssuranceMetadataResponse: &AssuranceMetadataResponse{
				IsVerified:     false,
				VerificationID: "",
			},
		}
	}

	return &AuthnResult{
		UserID:              authResponse.ID,
		UserType:            userResult.Type,
		OUID:                userResult.OUID,
		Token:               authResponse.ID,
		AvailableAttributes: availableAttributes,
	}, nil
}

// authenticateApp authenticates an application by verifying the client secret
// against the hashed secret stored in the entity's system credentials.
func (p *defaultAuthnProvider) authenticateApp(
	ctx context.Context,
	identifiers, credentials map[string]interface{},
) (*AuthnResult, *AuthnProviderError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthnProvider"))

	// Identify the entity by clientId.
	entityID, err := p.entityService.IdentifyEntity(ctx, identifiers)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, NewError(ErrorCodeAuthenticationFailed, "Authentication failed", "Client not found")
		}
		return nil, NewError(ErrorCodeSystemError, "Authentication failed", err.Error())
	}

	// Fetch entity with credentials.
	e, _, systemCredsJSON, err := p.entityService.GetEntityWithCredentials(ctx, *entityID)
	if err != nil {
		return nil, NewError(ErrorCodeSystemError, "Authentication failed", err.Error())
	}

	// Verify client secret if provided. If no credentials are given, this is an identify-only call.
	clientSecret, _ := credentials["clientSecret"].(string)
	if clientSecret != "" {
		if !verifyClientSecret(clientSecret, systemCredsJSON) {
			logger.Debug("Client secret verification failed", log.String("entityID", *entityID))
			return nil, NewError(ErrorCodeAuthenticationFailed, "Authentication failed", "Invalid client credentials")
		}
	}

	return &AuthnResult{
		UserID:   e.EntityID,
		UserType: e.EntityType,
		OUID:     e.OrganizationUnitID,
		Token:    e.EntityID,
	}, nil
}

// verifyClientSecret verifies the provided client secret against the stored hashed secret.
func verifyClientSecret(providedSecret string, systemCredsJSON json.RawMessage) bool {
	if len(systemCredsJSON) == 0 {
		return false
	}

	var creds map[string]interface{}
	if err := json.Unmarshal(systemCredsJSON, &creds); err != nil {
		return false
	}

	storedHash, ok := creds["clientSecret"].(string)
	if !ok || storedHash == "" {
		return false
	}

	// Hash the provided secret and compare using constant-time comparison.
	hashedProvided := hash.GenerateThumbprintFromString(providedSecret)
	return subtle.ConstantTimeCompare([]byte(hashedProvided), []byte(storedHash)) == 1
}

// GetAttributes retrieves the user attributes using the internal user service.
func (p *defaultAuthnProvider) GetAttributes(
	ctx context.Context,
	token string,
	requestedAttributes *RequestedAttributes,
	metadata *GetAttributesMetadata,
) (*GetAttributesResult, *AuthnProviderError) {
	userID := token

	userResult, authErr := p.userSvc.GetUser(ctx, userID, false)
	if authErr != nil {
		if authErr.Type == serviceerror.ClientErrorType {
			return nil, NewError(ErrorCodeInvalidToken, authErr.Error, authErr.ErrorDescription)
		}
		return nil, NewError(ErrorCodeSystemError, authErr.Error, authErr.ErrorDescription)
	}

	var allAttributes map[string]interface{}
	if len(userResult.Attributes) > 0 {
		if err := json.Unmarshal(userResult.Attributes, &allAttributes); err != nil {
			return nil, NewError(ErrorCodeSystemError, "System Error", "Failed to unmarshal user attributes")
		}
	}

	attributesResponse := &AttributesResponse{
		Attributes:    make(map[string]*AttributeResponse),
		Verifications: make(map[string]*VerificationResponse),
	}

	if requestedAttributes != nil && len(requestedAttributes.Attributes) > 0 {
		for attrName := range requestedAttributes.Attributes {
			if val, ok := allAttributes[attrName]; ok {
				attributesResponse.Attributes[attrName] = &AttributeResponse{
					Value: val,
					AssuranceMetadataResponse: &AssuranceMetadataResponse{
						IsVerified:     false,
						VerificationID: "",
					},
				}
			}
		}
	} else {
		for attrName, val := range allAttributes {
			attributesResponse.Attributes[attrName] = &AttributeResponse{
				Value: val,
				AssuranceMetadataResponse: &AssuranceMetadataResponse{
					IsVerified:     false,
					VerificationID: "",
				},
			}
		}
	}

	return &GetAttributesResult{
		UserID:             userResult.ID,
		UserType:           userResult.Type,
		OUID:               userResult.OUID,
		AttributesResponse: attributesResponse,
	}, nil
}
