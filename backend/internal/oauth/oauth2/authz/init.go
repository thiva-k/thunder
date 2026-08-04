// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"net/http"

	"github.com/thunder-id/thunderid/internal/flow/flowexec"
	oauthconfig "github.com/thunder-id/thunderid/internal/oauth/config"
	"github.com/thunder-id/thunderid/internal/oauth/oauth2/par"
	"github.com/thunder-id/thunderid/internal/oauth/oauth2/revocation"
	"github.com/thunder-id/thunderid/internal/system/constants"
	"github.com/thunder-id/thunderid/internal/system/jose/jwt"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// Initialize initializes the authorization handler and registers its routes.
func Initialize(
	mux *http.ServeMux,
	actorProvider providers.ActorProvider,
	resourceService providers.ResourceServerProvider,
	jwtService jwt.JWTServiceInterface,
	flowExecService flowexec.FlowExecServiceInterface,
	parService par.PARServiceInterface,
	criteriaRevoker revocation.CriteriaRevokerInterface,
	cfg oauthconfig.Config,
	storeProvider providers.RuntimeStoreProvider,
	transactioner providers.Transactioner,
) (AuthorizeServiceInterface, error) {
	authzCodeStore := newAuthorizationCodeStore(storeProvider)
	authzReqStore := newAuthorizationRequestStore(storeProvider)

	authzService := newAuthorizeService(
		actorProvider, resourceService, jwtService, flowExecService,
		authzCodeStore, authzReqStore, parService, transactioner, criteriaRevoker, cfg,
	)
	authzHandler := newAuthorizeHandler(authzService, cfg)
	registerRoutes(mux, authzHandler)
	return authzService, nil
}

// registerRoutes registers the GET /oauth2/authorize route. The POST /oauth2/auth/callback
// route is registered by the callback package which dispatches by grant type.
func registerRoutes(mux *http.ServeMux, authzHandler AuthorizeHandlerInterface) {
	// CORS MUST NOT be enabled on the authorization endpoint.
	// The client redirects the user agent to it; it is not accessed directly via XHR/fetch.
	mux.HandleFunc("GET /oauth2/authorize",
		withFrameProtection(authzHandler.HandleAuthorizeGetRequest))
}

// withFrameProtection wraps an HTTP handler to prevent the page from being embedded in frames.
func withFrameProtection(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(constants.XFrameOptionsHeaderName, constants.XFrameOptionsDeny)
		w.Header().Set(constants.ContentSecurityPolicyHeaderName, constants.ContentSecurityPolicyFrameAncestorsNone)
		handler(w, r)
	}
}
