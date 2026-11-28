// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// Package actorprovider resolves inbound actors (applications, agents, and similar entities)
// for runtime layers that need a unified view of inbound-client and backing records.
package actorprovider

import (
	"github.com/thunder-id/thunderid/internal/cimd"
	"github.com/thunder-id/thunderid/internal/entityprovider"
	"github.com/thunder-id/thunderid/internal/entitytype"
	"github.com/thunder-id/thunderid/internal/inboundclient"
	"github.com/thunder-id/thunderid/internal/role"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// Option adjusts the actor provider for the services the server runs.
type Option func(*actorProvider)

// ResolvingCIMDDocumentClients lets the actor provider name a client the Client ID domains of an application or
// agent admit as its Client ID Metadata Document names it, when the login page addresses it by its URL.
func ResolvingCIMDDocumentClients(cimdService cimd.CIMDServiceInterface) Option {
	return func(p *actorProvider) { p.cimdService = cimdService }
}

// Initialize creates the default ActorProvider backed by inbound-client, entity-provider,
// authentication provider, and role services. rulesResolvers holds the client rules resolver of each
// entity category that has rules.
func Initialize(
	inboundClient inboundclient.InboundClientServiceInterface,
	entityProvider entityprovider.EntityProviderInterface,
	authnProvider providers.AuthnProviderManager,
	roleService role.RoleServiceInterface,
	entityTypeService entitytype.EntityTypeServiceInterface,
	rulesResolvers map[providers.EntityCategory]ClientRulesResolverInterface,
	options ...Option,
) providers.ActorProvider {
	provider := newActorProvider(inboundClient, entityProvider, authnProvider, roleService, entityTypeService,
		rulesResolvers)
	for _, option := range options {
		option(provider)
	}
	return provider
}
