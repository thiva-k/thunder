// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package actorprovider

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"strings"

	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"

	"github.com/thunder-id/thunderid/internal/cimd"
	"github.com/thunder-id/thunderid/internal/entityprovider"
	"github.com/thunder-id/thunderid/internal/entitytype"
	"github.com/thunder-id/thunderid/internal/inboundclient"
	"github.com/thunder-id/thunderid/internal/role"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/internal/system/security"
	"github.com/thunder-id/thunderid/internal/system/utils"
)

// ClientRulesResolverInterface resolves the runtime rules of the clients whose entities belong to one entity
// category. The domain that owns the category implements it; the actor provider does not know why a
// client has its rules.
type ClientRulesResolverInterface interface {
	// GetClientRules returns the rules of the client of the given entity, or nil when none apply.
	GetClientRules(ctx context.Context, entityID string) (*providers.ClientRules, *tidcommon.ServiceError)
}

// actorProvider delegates actor resolution to inbound-client and entity-provider services, and
// actor authentication to the authentication provider. The CIMD service supplies the name and logo of a client
// that the Client ID domains of an application or agent admit, when the login page names it by its URL.
type actorProvider struct {
	inboundClient  inboundclient.InboundClientServiceInterface
	cimdService    cimd.CIMDServiceInterface
	entityProvider entityprovider.EntityProviderInterface
	authnProvider  providers.AuthnProviderManager
	roleService    role.RoleServiceInterface
	entityTypeSvc  entitytype.EntityTypeServiceInterface
	rulesResolvers map[providers.EntityCategory]ClientRulesResolverInterface
	logger         *log.Logger
}

// newActorProvider creates a new actorProvider backed by the given inbound-client, entity-provider,
// authentication provider, and role service, with the client rules resolver of each entity category.
func newActorProvider(
	inboundClient inboundclient.InboundClientServiceInterface,
	entityProvider entityprovider.EntityProviderInterface,
	authnProvider providers.AuthnProviderManager,
	roleService role.RoleServiceInterface,
	entityTypeSvc entitytype.EntityTypeServiceInterface,
	rulesResolvers map[providers.EntityCategory]ClientRulesResolverInterface,
) *actorProvider {
	return &actorProvider{
		inboundClient:  inboundClient,
		entityProvider: entityProvider,
		authnProvider:  authnProvider,
		roleService:    roleService,
		entityTypeSvc:  entityTypeSvc,
		rulesResolvers: rulesResolvers,
		logger:         log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ActorProvider")),
	}
}

// GetOAuthClientByClientID returns the OAuth client for the given client ID, including a client the Client ID
// domains of an application or agent admit, or nil when there is none.
func (p *actorProvider) GetOAuthClientByClientID(
	ctx context.Context, clientID string,
) (*providers.OAuthClient, *tidcommon.ServiceError) {
	client, err := p.inboundClient.GetOAuthClientByClientID(ctx, clientID)
	if err != nil {
		p.logger.Error(ctx, "Failed to fetch OAuth client", log.String("clientID", clientID), log.Error(err))
		return nil, &tidcommon.InternalServerError
	}
	return p.withClientRules(ctx, toProviderOAuthClient(client))
}

// IsOAuthClientAccessibleFromOU reports whether an OAuth2 client may act on behalf of an
// organization unit, which the inbound client service answers from the sharing policies.
func (p *actorProvider) IsOAuthClientAccessibleFromOU(
	ctx context.Context, client *providers.OAuthClient, ouID string,
) (bool, *tidcommon.ServiceError) {
	return p.inboundClient.IsClientAccessibleFromOU(ctx, client, ouID)
}

// GetOAuthClientByID returns the runtime OAuth client for the given entity UUID.
func (p *actorProvider) GetOAuthClientByID(
	ctx context.Context, id string,
) (*providers.OAuthClient, *tidcommon.ServiceError) {
	client, err := p.inboundClient.GetOAuthClientByEntityID(ctx, id)
	if err != nil {
		p.logger.Error(ctx, "Failed to fetch OAuth client by entity ID", log.String("id", id), log.Error(err))
		return nil, &tidcommon.InternalServerError
	}
	return p.withClientRules(ctx, toProviderOAuthClient(client))
}

// resolveDomainClient resolves an ID that is the Client Identifier URL of a client the Client ID domains of an
// application or agent admit, as the login page names such a client. It returns the ID of that application or
// agent and the client's document, or an empty ID when id names no such client.
func (p *actorProvider) resolveDomainClient(ctx context.Context, id string) (
	string, *cimd.DocumentClient, *tidcommon.ServiceError) {
	if p.cimdService == nil || !strings.HasPrefix(id, clientIdentifierURLPrefix) {
		return "", nil, nil
	}
	client, err := p.inboundClient.GetOAuthClientByClientID(ctx, id)
	if err != nil {
		p.logger.Error(ctx, "Failed to fetch OAuth client", log.String("clientID", id), log.Error(err))
		return "", nil, &tidcommon.InternalServerError
	}
	if client == nil || !client.ResolvedByClientIDDomain {
		return "", nil, nil
	}
	document, svcErr := p.cimdService.ResolveDocumentClient(ctx, id)
	if svcErr != nil || document == nil {
		return "", nil, svcErr
	}
	return client.ID, document, nil
}

// withClientRules adds to a client the runtime rules its entity category's resolver returns. A client of
// a category without a resolver has no rules. A failed lookup is a server error, so no client is served
// without rules that may apply to it.
func (p *actorProvider) withClientRules(
	ctx context.Context, client *providers.OAuthClient,
) (*providers.OAuthClient, *tidcommon.ServiceError) {
	if client == nil {
		return nil, nil
	}
	resolver, ok := p.rulesResolvers[client.EntityCategory]
	if !ok {
		return client, nil
	}
	rules, svcErr := resolver.GetClientRules(ctx, client.ID)
	if svcErr != nil {
		p.logger.Error(ctx, "Failed to resolve the client rules",
			log.String("id", client.ID), log.String("errorCode", svcErr.Code))
		return nil, &tidcommon.InternalServerError
	}
	if rules != nil {
		client.Rules = *rules
	}
	return client, nil
}

// GetOAuthProfileByID returns the stored OAuth profile for the given entity UUID.
func (p *actorProvider) GetOAuthProfileByID(
	ctx context.Context, id string,
) (*providers.OAuthProfile, *tidcommon.ServiceError) {
	profile, err := p.inboundClient.GetOAuthProfileByEntityID(ctx, id)
	if err != nil {
		if errors.Is(err, inboundclient.ErrInboundClientNotFound) {
			return nil, &ErrorActorNotFound
		}
		p.logger.Error(ctx, "Failed to fetch OAuth profile by entity ID",
			log.String("id", id), log.Error(err))
		return nil, &tidcommon.InternalServerError
	}
	return profile, nil
}

// GetInboundClientByID returns the inbound-client row for the given ID. An ID that is the Client Identifier URL
// of a client the Client ID domains of an application or agent admit returns that application's or agent's
// inbound client with the client's own logo and links.
func (p *actorProvider) GetInboundClientByID(
	ctx context.Context, id string,
) (*providers.InboundClient, *tidcommon.ServiceError) {
	client, err := p.inboundClient.GetInboundClientByEntityID(ctx, id)
	if err != nil {
		if errors.Is(err, inboundclient.ErrInboundClientNotFound) {
			return p.getDomainInboundClient(ctx, id)
		}
		p.logger.Error(ctx, "Failed to fetch inbound client", log.String("id", id), log.Error(err))
		return nil, &tidcommon.InternalServerError
	}
	return client, nil
}

// getDomainInboundClient returns the inbound client of the application or agent whose Client ID domains admit
// the client with the given Client Identifier URL, with the client's own logo and links in place of its own.
func (p *actorProvider) getDomainInboundClient(
	ctx context.Context, id string,
) (*providers.InboundClient, *tidcommon.ServiceError) {
	entityID, document, svcErr := p.resolveDomainClient(ctx, id)
	if svcErr != nil {
		return nil, svcErr
	}
	if entityID == "" {
		return nil, &ErrorActorNotFound
	}
	client, err := p.inboundClient.GetInboundClientByEntityID(ctx, entityID)
	if err != nil {
		if errors.Is(err, inboundclient.ErrInboundClientNotFound) {
			return nil, &ErrorActorNotFound
		}
		p.logger.Error(ctx, "Failed to fetch inbound client", log.String("id", entityID), log.Error(err))
		return nil, &tidcommon.InternalServerError
	}
	overlaid := *client
	overlaid.Properties = maps.Clone(client.Properties)
	if overlaid.Properties == nil {
		overlaid.Properties = map[string]interface{}{}
	}
	for _, key := range documentPropertyKeys {
		delete(overlaid.Properties, key)
	}
	maps.Copy(overlaid.Properties, document.Properties)
	return &overlaid, nil
}

// AuthenticateActor verifies the supplied credentials against the actor resolved from the given
// identifiers. It returns nil when authentication succeeds and a service error otherwise. The
// default implementation delegates to the authentication provider, which performs the credential
// lookup and constant-time verification generically through the entity layer. A custom actor
// provider may implement its own scheme.
func (p *actorProvider) AuthenticateActor(
	ctx context.Context, identifiers, credentials map[string]interface{},
) *tidcommon.ServiceError {
	_, _, svcErr := p.authnProvider.AuthenticateUser(ctx, identifiers, credentials, nil, nil, providers.AuthUser{})
	return svcErr
}

// GetActor returns the backing entity record for the given actor ID. An ID that is the Client Identifier URL of
// a client the Client ID domains of an application or agent admit returns that application's or agent's entity,
// named as the client's document names it.
func (p *actorProvider) GetActor(ctx context.Context, actorID string) (*providers.Entity, *tidcommon.ServiceError) {
	entity, epErr := p.entityProvider.GetEntityProfile(ctx, actorID)
	if epErr != nil {
		if epErr.Code == entityprovider.ErrorCodeEntityNotFound {
			return p.getDomainActor(ctx, actorID, epErr)
		}
		return nil, mapEntityProviderError(epErr)
	}
	if entity.Category == providers.EntityCategoryAgent {
		ctx := security.WithRuntimeContext(ctx)
		displayPaths := entitytype.ResolveDisplayAttributePaths(
			ctx, entitytype.TypeCategoryAgent, []string{entity.Type}, p.entityTypeSvc, p.logger)
		entity.Display = utils.ResolveDisplay(entity.ID, entity.Type, entity.Attributes, displayPaths)
	}
	return entity, nil
}

// getDomainActor returns the entity of the application or agent whose Client ID domains admit the client with
// the given Client Identifier URL, named as the client's document names it. notFound is returned, mapped, when
// the URL names no such client.
func (p *actorProvider) getDomainActor(
	ctx context.Context, id string, notFound *entityprovider.EntityProviderError,
) (*providers.Entity, *tidcommon.ServiceError) {
	entityID, document, svcErr := p.resolveDomainClient(ctx, id)
	if svcErr != nil {
		return nil, svcErr
	}
	if entityID == "" {
		return nil, mapEntityProviderError(notFound)
	}
	entity, epErr := p.entityProvider.GetEntityProfile(ctx, entityID)
	if epErr != nil {
		return nil, mapEntityProviderError(epErr)
	}
	attributes := readEntitySystemAttributes(entity)
	attributes["name"] = document.Name
	systemAttributes, err := json.Marshal(attributes)
	if err != nil {
		p.logger.Error(ctx, "Failed to marshal the system attributes of a client", log.Error(err))
		return nil, &tidcommon.InternalServerError
	}
	named := *entity
	named.SystemAttributes = systemAttributes
	named.Display = document.Name
	return &named, nil
}

// GetActorGroups returns transitive group memberships for the given actor ID.
func (p *actorProvider) GetActorGroups(
	ctx context.Context, actorID string,
) ([]providers.EntityGroup, *tidcommon.ServiceError) {
	groups, epErr := p.entityProvider.GetTransitiveEntityGroups(ctx, actorID)
	if epErr != nil {
		if epErr.Code == entityprovider.ErrorCodeNotImplemented {
			return nil, nil
		}
		return nil, mapEntityProviderError(epErr)
	}
	return groups, nil
}

// GetTransitiveGroupAncestors returns the ancestor chain of a single group.
func (p *actorProvider) GetTransitiveGroupAncestors(
	ctx context.Context, groupID string,
) ([]string, *tidcommon.ServiceError) {
	ancestors, epErr := p.entityProvider.GetTransitiveGroupAncestors(ctx, groupID)
	if epErr != nil {
		if epErr.Code == entityprovider.ErrorCodeNotImplemented {
			return nil, nil
		}
		return nil, mapEntityProviderError(epErr)
	}
	return ancestors, nil
}

// GetActorRoles returns the roles assigned to the actor, directly and through the given groups.
func (p *actorProvider) GetActorRoles(
	ctx context.Context, actorID string, groupIDs []string,
) ([]string, *tidcommon.ServiceError) {
	return p.GetActorRolesInOU(ctx, actorID, groupIDs, "")
}

// GetActorRolesInOU returns the names of the roles assigned to the actor directly or through
// groupIDs, counting only the assignments made in ouID, or those made anywhere when it is empty.
func (p *actorProvider) GetActorRolesInOU(
	ctx context.Context, actorID string, groupIDs []string, ouID string,
) ([]string, *tidcommon.ServiceError) {
	if p.roleService == nil {
		return nil, nil
	}
	return p.roleService.GetUserRoles(ctx, actorID, groupIDs, ouID)
}

func mapEntityProviderError(epErr *entityprovider.EntityProviderError) *tidcommon.ServiceError {
	if epErr == nil {
		return nil
	}
	switch epErr.Code {
	case entityprovider.ErrorCodeEntityNotFound:
		return &ErrorEntityNotFound
	case entityprovider.ErrorCodeNotImplemented:
		return &ErrorActorLookupNotSupported
	default:
		return &tidcommon.InternalServerError
	}
}

func toProviderOAuthClient(c *providers.OAuthClient) *providers.OAuthClient {
	if c == nil {
		return nil
	}
	client := &providers.OAuthClient{
		ID:                                 c.ID,
		OUID:                               c.OUID,
		ClientID:                           c.ClientID,
		RedirectURIs:                       c.RedirectURIs,
		PostLogoutRedirectURIs:             c.PostLogoutRedirectURIs,
		BackchannelLogoutURI:               c.BackchannelLogoutURI,
		TokenEndpointAuthMethod:            c.TokenEndpointAuthMethod,
		PKCERequired:                       c.PKCERequired,
		PublicClient:                       c.PublicClient,
		RequirePushedAuthorizationRequests: c.RequirePushedAuthorizationRequests,
		DPoPBoundAccessTokens:              c.DPoPBoundAccessTokens,
		IncludeActClaim:                    c.IncludeActClaim,
		ClientIDMetadataDocument:           c.ClientIDMetadataDocument,
		ClientIDDomains:                    c.ClientIDDomains,
		ResolvedByClientIDDomain:           c.ResolvedByClientIDDomain,
		EntityCategory:                     c.EntityCategory,
		Token:                              c.Token,
		Scopes:                             c.Scopes,
		UserInfo:                           c.UserInfo,
		ScopeClaims:                        c.ScopeClaims,
		Certificate:                        c.Certificate,
		AcrValues:                          c.AcrValues,
	}
	client.GrantTypes = append(client.GrantTypes, c.GrantTypes...)
	client.ResponseTypes = append(client.ResponseTypes, c.ResponseTypes...)
	return client
}
