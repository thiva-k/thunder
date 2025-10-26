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

package common

import (
	"github.com/asgardeo/thunder/internal/idp"
)

// Authenticator name constants.
const (
	AuthenticatorCredentials = "CredentialsAuthenticator"
	AuthenticatorSMSOTP      = "SMSOTPAuthenticator"
	AuthenticatorGoogle      = "GoogleOIDCAuthenticator"
	AuthenticatorGithub      = "GithubOAuthAuthenticator"
	AuthenticatorOAuth       = "OAuthAuthenticator"
	AuthenticatorOIDC        = "OIDCAuthenticator"
)

// AuthenticationFactor represents the type of authentication factor.
type AuthenticationFactor string

const (
	// FactorKnowledge represents "something you know" (e.g., password, PIN).
	FactorKnowledge AuthenticationFactor = "KNOWLEDGE"
	// FactorPossession represents "something you have" (e.g., OTP device, SMS).
	FactorPossession AuthenticationFactor = "POSSESSION"
	// FactorInherence represents "something you are" (e.g., biometrics).
	FactorInherence AuthenticationFactor = "INHERENCE"
)

// AuthenticatorMeta represents an authenticator's metadata including authentication factors.
type AuthenticatorMeta struct {
	// Name is the unique identifier for the authenticator (used in individual authentication APIs)
	Name string
	// Factors represents the authentication factors this authenticator validates
	Factors []AuthenticationFactor
}

// AuthenticatorReference represents an engaged authenticator in the authentication flow.
type AuthenticatorReference struct {
	// Authenticator is the name of the authenticator
	Authenticator string `json:"authenticator"`
	// Step is the step number in the flow where this authenticator was engaged
	Step int `json:"step"`
	// Timestamp is the authenticator engaged time (Unix epoch time in seconds)
	Timestamp int64 `json:"timestamp"`
}

// authenticatorRegistry holds the metadata for all authenticators in the system.
var authenticatorRegistry = map[string]AuthenticatorMeta{
	AuthenticatorCredentials: {
		Name:    AuthenticatorCredentials,
		Factors: []AuthenticationFactor{FactorKnowledge},
	},
	AuthenticatorSMSOTP: {
		Name:    AuthenticatorSMSOTP,
		Factors: []AuthenticationFactor{FactorPossession},
	},
	AuthenticatorGoogle: {
		Name:    AuthenticatorGoogle,
		Factors: []AuthenticationFactor{FactorKnowledge},
	},
	AuthenticatorGithub: {
		Name:    AuthenticatorGithub,
		Factors: []AuthenticationFactor{FactorKnowledge},
	},
	AuthenticatorOAuth: {
		Name:    AuthenticatorOAuth,
		Factors: []AuthenticationFactor{FactorKnowledge},
	},
	AuthenticatorOIDC: {
		Name:    AuthenticatorOIDC,
		Factors: []AuthenticationFactor{FactorKnowledge},
	},
}

// getAuthenticatorMetaData returns the authenticator metadata for the given authenticator.
func getAuthenticatorMetaData(name string) *AuthenticatorMeta {
	if auth, ok := authenticatorRegistry[name]; ok {
		return &auth
	}
	return nil
}

// GetAuthenticatorFactors returns the authentication factors for the given authenticator.
func GetAuthenticatorFactors(name string) []AuthenticationFactor {
	if auth := getAuthenticatorMetaData(name); auth != nil {
		return auth.Factors
	}
	return []AuthenticationFactor{}
}

// GetAuthenticatorNameForIDPType returns the authenticator name for a given IDP type.
func GetAuthenticatorNameForIDPType(idpType idp.IDPType) string {
	switch idpType {
	case idp.IDPTypeGoogle:
		return AuthenticatorGoogle
	case idp.IDPTypeGitHub:
		return AuthenticatorGithub
	case idp.IDPTypeOAuth:
		return AuthenticatorOAuth
	case idp.IDPTypeOIDC:
		return AuthenticatorOIDC
	default:
		return AuthenticatorOAuth
	}
}
