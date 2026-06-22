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

package openid4vci

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/thunder-id/thunderid/internal/oauth/oauth2/dpop"
	"github.com/thunder-id/thunderid/internal/openid4vci/credential"
	"github.com/thunder-id/thunderid/internal/system/config"
	declarativeresource "github.com/thunder-id/thunderid/internal/system/declarative_resource"
	"github.com/thunder-id/thunderid/internal/system/jose/jwt"
	kmprovider "github.com/thunder-id/thunderid/internal/system/kmprovider/common"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/internal/user"
)

// Initialize wires the OpenID4VCI issuer engine, the credential-configuration
// management API, and the wallet-facing endpoints. When no signing key is
// configured, issuance is disabled and Initialize returns nil without error.
func Initialize(
	mux *http.ServeMux, cryptoProvider kmprovider.RuntimeCryptoProvider,
	jwtService jwt.JWTServiceInterface, userService user.UserServiceInterface,
	dpopVerifier dpop.VerifierInterface, ouResolver credential.OUResolver,
) (
	OpenID4VCIServiceInterface, credential.CredentialConfigurationServiceInterface,
	declarativeresource.ResourceExporter, error,
) {
	runtime := config.GetServerRuntime()
	cfg := runtime.Config.OpenID4VCI
	if cfg.SigningKeyID == "" {
		log.GetLogger().Debug(context.Background(), "OpenID4VCI issuer not configured; credential issuance disabled")
		return nil, nil, nil, nil
	}

	// Engine URLs default to the server's public URL; explicit config overrides.
	serverCfg := runtime.Config.Server
	serverURL := strings.TrimRight(config.GetServerURL(&serverCfg), "/")
	credentialIssuer := cfg.CredentialIssuer
	if credentialIssuer == "" {
		credentialIssuer = serverURL
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = serverURL
	}
	authServers := cfg.AuthorizationServers
	if len(authServers) == 0 {
		authServers = []string{serverURL}
	}

	signer, err := newIssuerSigner(context.Background(), cryptoProvider, cfg.SigningKeyID)
	if err != nil {
		return nil, nil, nil, err
	}

	credSvc, credExporter, err := credential.Initialize(mux, ouResolver)
	if err != nil {
		return nil, nil, nil, err
	}

	svc, err := newOpenID4VCIService(serviceConfig{
		CredentialIssuer:     credentialIssuer,
		BaseURL:              baseURL,
		AuthorizationServers: authServers,
		NonceTTL:             time.Duration(cfg.NonceTTLSeconds) * time.Second,
		ProofMaxAge:          time.Duration(cfg.ProofMaxAgeSeconds) * time.Second,
		CredentialValidity:   time.Duration(cfg.CredentialValiditySeconds) * time.Second,
		BatchSize:            cfg.BatchSize,
		EnforceScope:         cfg.EnforceScope,
	}, signer, newDBNonceStore(), newDBOfferStore(),
		jwtService, userService, credSvc)
	if err != nil {
		return nil, nil, nil, err
	}

	registerRoutes(mux, newOpenID4VCIHandler(svc, dpopVerifier, baseURL+credentialPath))
	return svc, credSvc, credExporter, nil
}
