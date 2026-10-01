// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package backchannel

import (
	"fmt"
	"time"

	"github.com/thunder-id/thunderid/internal/flow/session"
	oauthconfig "github.com/thunder-id/thunderid/internal/oauth/config"
	"github.com/thunder-id/thunderid/internal/oauth/oauth2/tokenservice"
	syshttp "github.com/thunder-id/thunderid/internal/system/http"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// Initialize builds the dispatcher and registers it on the session termination hook. It returns nil
// when back-channel logout is disabled, or when there is no hook because no session store exists.
// The caller starts the dispatcher once the server is about to serve, and stops it during graceful
// shutdown.
func Initialize(
	tokens tokenservice.TokenBuilderInterface,
	clients providers.ActorProvider,
	events providers.ObservabilityProvider,
	hook session.TerminationHook,
	cfg oauthconfig.Config,
) (DispatcherInterface, error) {
	bcl := cfg.OAuth.Logout.Backchannel
	if !bcl.IsEnabled() || hook == nil {
		return nil, nil
	}
	httpClient := syshttp.NewHTTPClientWithoutRedirects(
		time.Duration(bcl.RequestTimeout)*time.Second, bcl.RejectsPrivateAddresses())
	d := newDispatcher(bcl, tokens, clients, httpClient, events)
	if err := hook.Add(d); err != nil {
		return nil, fmt.Errorf("failed to register the back-channel logout dispatcher: %w", err)
	}
	return d, nil
}
