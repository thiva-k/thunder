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

// Package authz implements the OAuth2 authorization functionality.
package authz

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/log"
)

// AuthorizeServiceInterface defines the interface for authorization services.
type AuthorizeServiceInterface interface {
	GetAuthorizationCodeDetails(clientID string, code string) (*AuthorizationCode, error)
}

// authorizeService implements the AuthorizeService for managing OAuth2 authorization flows.
type authorizeService struct {
	authzStore AuthorizationCodeStoreInterface
	logger     log.Logger
}

// newAuthorizeService creates a new instance of authorizeService with injected dependencies.
func newAuthorizeService(authzStore AuthorizationCodeStoreInterface) AuthorizeServiceInterface {
	return &authorizeService{
		authzStore: authzStore,
		logger:     *log.GetLogger().With(log.String(log.LoggerKeyComponentName, "AuthorizeService")),
	}
}

// GetAuthorizationCodeDetails retrieves and invalidates the authorization code.
func (as *authorizeService) GetAuthorizationCodeDetails(clientID string, code string) (*AuthorizationCode, error) {
	authCode, err := as.authzStore.GetAuthorizationCode(clientID, code)
	if err != nil {
		if errors.Is(err, ErrAuthorizationCodeNotFound) {
			return nil, errors.New("invalid authorization code")
		}
		as.logger.Error("error retrieving authorization code", log.Error(err))
		return nil, errors.New("failed to retrieve authorization code")
	}
	if authCode == nil || authCode.Code == "" {
		return nil, errors.New("invalid authorization code")
	}

	// Invalidate the authorization code after use.
	err = as.authzStore.DeactivateAuthorizationCode(*authCode)
	if err != nil {
		as.logger.Error("error invalidating authorization code", log.Error(err))
		return nil, errors.New("failed to invalidate authorization code")
	}
	return authCode, nil
}
