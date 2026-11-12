// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/thunder-id/thunderid/internal/system/config"
	syshttp "github.com/thunder-id/thunderid/internal/system/http"
	"github.com/thunder-id/thunderid/internal/system/middleware"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// Initialize builds the AAuth person server and registers its routes. It returns nil when the
// person server is disabled.
func Initialize(ctx context.Context, mux *http.ServeMux, cfg config.AAuthConfig, issuer string,
	crypto providers.RuntimeCryptoProvider, runtimeStore providers.RuntimeStoreProvider,
	codeRedeemer authorizationCodeRedeemer, actorProvider providers.ActorProvider) error {
	if !cfg.IsEnabled() {
		return nil
	}
	if !isServerIdentifier(issuer) {
		return fmt.Errorf("aauth: the server public URL %q must be an https origin without a port or path", issuer)
	}
	if cfg.SigningKeyID == "" || cfg.InteractionClientID == "" {
		return errors.New("aauth: signing_key_id and interaction_client_id are required")
	}
	client, err := newFetchClient(cfg)
	if err != nil {
		return err
	}
	service, err := newAAuthService(ctx, issuer, cfg.SigningKeyID, cfg.InteractionClientID, crypto,
		newAAuthStore(runtimeStore), newIssuerKeyFetcher(client), codeRedeemer, actorProvider)
	if err != nil {
		return err
	}
	registerRoutes(mux, newAAuthHandler(service))
	return nil
}

// newFetchClient builds the client that fetches agent and resource metadata. It follows no
// redirects, and refuses private addresses unless allowed for local development.
func newFetchClient(cfg config.AAuthConfig) (syshttp.HTTPClientInterface, error) {
	rejectPrivate := !cfg.IsPrivateNetworkAllowed()
	if cfg.TrustedCAFile == "" {
		return syshttp.NewHTTPClientWithoutRedirects(httpFetchTimeout, rejectPrivate), nil
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	pem, err := os.ReadFile(cfg.TrustedCAFile)
	if err != nil {
		return nil, fmt.Errorf("aauth: reading trusted_ca_file: %w", err)
	}
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("aauth: trusted_ca_file has no certificates")
	}
	return syshttp.NewHTTPClientWithRootCAs(httpFetchTimeout, pool, rejectPrivate), nil
}

// registerRoutes registers the person server's routes.
func registerRoutes(mux *http.ServeMux, h *aauthHandler) {
	public := middleware.CORSOptions{
		AllowedMethods: []string{"GET"},
		AllowedHeaders: middleware.DefaultAllowedHeaders,
		MaxAge:         600,
	}
	mux.HandleFunc(middleware.WithCORS("GET "+pathPersonMetadata, h.HandleMetadata, public))
	mux.HandleFunc(middleware.WithCORS("GET "+pathJWKS, h.HandleJWKS, public))
	mux.HandleFunc("POST "+pathPersonToken, h.HandlePersonToken)
	mux.HandleFunc("POST "+pathAuthToken, h.HandleAuthToken)
	mux.HandleFunc("POST "+pathMission, h.HandleMissionProposal)
	mux.HandleFunc("POST "+pathMission+"/{s256}", h.HandleMission)
	mux.HandleFunc("GET "+pathPending+"/{id}", h.HandlePending)
	mux.HandleFunc("GET "+pathInteraction, h.HandleInteraction)
	mux.HandleFunc("GET "+pathInteraction+"/callback", h.HandleInteractionCallback)
	mux.HandleFunc("POST "+pathInteraction+"/decision", h.HandleInteractionDecision)
}
