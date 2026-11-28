// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// Package resolve provides functionality for resolving design configurations.
package resolve

import (
	"context"
	"errors"
	"strings"

	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"

	"github.com/thunder-id/thunderid/internal/application"
	"github.com/thunder-id/thunderid/internal/design/common"
	layoutmgt "github.com/thunder-id/thunderid/internal/design/layout/mgt"
	thememgt "github.com/thunder-id/thunderid/internal/design/theme/mgt"
	"github.com/thunder-id/thunderid/internal/inboundclient"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/internal/system/security"
)

const serviceLogger = "DesignResolveService"

// DesignResolveServiceInterface defines the interface for the design resolve service.
type DesignResolveServiceInterface interface {
	ResolveDesign(
		ctx context.Context, resolveType providers.DesignResolveType, id string,
	) (*providers.DesignResponse, *tidcommon.ServiceError)
	ResolveDesignContent(
		ctx context.Context, themeID, colorScheme, content string,
	) (string, *tidcommon.ServiceError)
}

// designResolveService is the default implementation of the DesignResolveServiceInterface.
type designResolveService struct {
	themeMgtService    thememgt.ThemeMgtServiceInterface
	layoutMgtService   layoutmgt.LayoutMgtServiceInterface
	applicationService application.ApplicationServiceInterface
	inboundClient      inboundclient.InboundClientServiceInterface
	logger             *log.Logger
}

// newDesignResolveService creates a new instance of DesignResolveService with injected dependencies.
func newDesignResolveService(
	themeMgtService thememgt.ThemeMgtServiceInterface,
	layoutMgtService layoutmgt.LayoutMgtServiceInterface,
	applicationService application.ApplicationServiceInterface,
	inboundClient inboundclient.InboundClientServiceInterface,
) DesignResolveServiceInterface {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, serviceLogger))
	return &designResolveService{
		themeMgtService:    themeMgtService,
		layoutMgtService:   layoutMgtService,
		applicationService: applicationService,
		inboundClient:      inboundClient,
		logger:             logger,
	}
}

// ResolveDesign resolves a design configuration by type and ID.
// TODO: Add support for OU type and fallback logic.
func (drs *designResolveService) ResolveDesign(
	ctx context.Context, resolveType providers.DesignResolveType, id string,
) (*providers.DesignResponse, *tidcommon.ServiceError) {
	if resolveType == "" {
		return nil, &common.ErrorInvalidResolveType
	}

	if id == "" {
		return nil, &common.ErrorMissingResolveID
	}

	// Currently only APP type is supported
	if resolveType != providers.DesignResolveTypeAPP {
		return nil, &common.ErrorUnsupportedResolveType
	}

	// Get the application by ID
	if drs.applicationService == nil {
		drs.logger.Error(ctx, "Application service is not available")
		return nil, &tidcommon.InternalServerError
	}

	themeID, layoutID, svcErr := drs.resolveClientDesign(ctx, id)
	if svcErr != nil {
		return nil, svcErr
	}

	// Check if the application has theme or layout configured
	if themeID == "" && layoutID == "" {
		return nil, &common.ErrorApplicationHasNoDesign
	}

	designResponse := &providers.DesignResponse{}

	// Get theme configuration if available
	if themeID != "" {
		if drs.themeMgtService == nil {
			drs.logger.Error(ctx, "Theme management service is not available")
			return nil, &tidcommon.InternalServerError
		}

		themeConfig, svcErr := drs.themeMgtService.GetTheme(ctx, themeID)
		if svcErr != nil {
			if svcErr.Code == thememgt.ErrorThemeNotFound.Code {
				// The referenced theme has been deleted; fall back to the system default by leaving
				// the theme unset in the response.
				drs.logger.Warn(ctx, "Application references a deleted theme; falling back to default",
					log.String("applicationID", id),
					log.String("themeID", themeID))
			} else {
				return nil, svcErr
			}
		} else {
			designResponse.Theme = themeConfig.Theme
		}
	}

	// Get layout configuration if available
	if layoutID != "" {
		if drs.layoutMgtService == nil {
			drs.logger.Error(ctx, "Layout management service is not available")
			return nil, &tidcommon.InternalServerError
		}

		layoutConfig, svcErr := drs.layoutMgtService.GetLayout(ctx, layoutID)
		if svcErr != nil {
			if svcErr.Code == layoutmgt.ErrorLayoutNotFound.Code {
				// The referenced layout has been deleted; fall back to the system default by leaving
				// the layout unset in the response.
				drs.logger.Warn(ctx, "Application references a deleted layout; falling back to default",
					log.String("applicationID", id),
					log.String("layoutID", layoutID))
			} else {
				return nil, svcErr
			}
		} else {
			designResponse.Layout = layoutConfig.Layout
		}
	}

	drs.logger.Debug(ctx, "Successfully resolved design configuration",
		log.String("type", string(resolveType)),
		log.String("id", id),
		log.String("themeID", themeID),
		log.String("layoutID", layoutID))

	return designResponse, nil
}

// resolveClientDesign returns the theme and layout of the application with the given ID, or, for the Client
// Identifier URL of a client the Client ID domains of an application or agent admit, of that application or
// agent. An unknown ID reports no design, the same as an application with nothing configured, since this
// endpoint is unauthenticated and must not confirm which IDs exist.
func (drs *designResolveService) resolveClientDesign(ctx context.Context, id string) (
	themeID, layoutID string, svcErr *tidcommon.ServiceError) {
	// The design of any organization unit's application is public.
	ctx = security.WithRuntimeContext(ctx)
	app, svcErr := drs.applicationService.GetApplication(ctx, id)
	if svcErr == nil {
		return app.ThemeID, app.LayoutID, nil
	}
	if svcErr.Code != application.ErrorApplicationNotFound.Code &&
		svcErr.Code != application.ErrorInvalidApplicationID.Code {
		return "", "", svcErr
	}
	if inbound, err := drs.domainClientInbound(ctx, id); err != nil {
		drs.logger.Error(ctx, "Failed to resolve the client a Client ID domain admits", log.Error(err))
		return "", "", &tidcommon.InternalServerError
	} else if inbound != nil {
		return inbound.ThemeID, inbound.LayoutID, nil
	}
	if svcErr.Code == application.ErrorInvalidApplicationID.Code {
		return "", "", &common.ErrorMissingResolveID
	}
	drs.logger.Debug(ctx, "No design resolved; application does not exist", log.String("applicationID", id))
	return "", "", &common.ErrorApplicationHasNoDesign
}

// domainClientInbound returns the inbound client of the application or agent whose Client ID domains admit the
// client with the given Client Identifier URL, or nil when id names no such client.
func (drs *designResolveService) domainClientInbound(ctx context.Context, id string) (
	*providers.InboundClient, error) {
	if drs.inboundClient == nil || !strings.HasPrefix(id, clientIdentifierURLPrefix) {
		return nil, nil
	}
	client, err := drs.inboundClient.GetOAuthClientByClientID(ctx, id)
	if err != nil || client == nil || !client.ResolvedByClientIDDomain {
		return nil, err
	}
	inbound, err := drs.inboundClient.GetInboundClientByEntityID(ctx, client.ID)
	if errors.Is(err, inboundclient.ErrInboundClientNotFound) {
		return nil, nil
	}
	return inbound, err
}

// ResolveDesignContent replaces every "{{design(<dot.path>)}}" placeholder in content with its value
// from the theme's color scheme. Unknown placeholders are left as-is; content with no placeholder is
// returned unchanged without loading the theme.
//
// Each resolved value is HTML-escaped (see substituteDesignTokens) so it is safe inside the HTML the
// email body places it in.
func (drs *designResolveService) ResolveDesignContent(
	ctx context.Context, themeID, colorScheme, content string,
) (string, *tidcommon.ServiceError) {
	if !designPlaceholderRegex.MatchString(content) {
		return content, nil
	}

	tokens, svcErr := drs.resolveDesignTokens(ctx, themeID, colorScheme)
	if svcErr != nil {
		return "", svcErr
	}

	return substituteDesignTokens(content, tokens), nil
}

// resolveDesignTokens resolves a theme into a dot-path token map (e.g. "palette.primary.main" ->
// "#fa7b3f"). It loads the theme, resolves the applicable theme (the root with the color scheme that
// applies promoted onto it, the requested one or the theme's defaultColorScheme when that is empty or
// absent), and flattens it so both scheme-level values (palette.*) and root-level values (typography.*,
// shape.*) are addressable. It returns InternalServerError (logged with the theme ID) when the theme is
// empty, malformed, or defines no color scheme that applies. Resolution is generic: any token present in
// the theme resolves, and which tokens a consumer references is the consumer's concern. The values are
// raw, theme-authored strings: a consumer must encode them for its output context before use to avoid
// injection.
func (drs *designResolveService) resolveDesignTokens(
	ctx context.Context, themeID, colorScheme string,
) (map[string]string, *tidcommon.ServiceError) {
	if drs.themeMgtService == nil {
		drs.logger.Error(ctx, "Theme management service is not available")
		return nil, &tidcommon.InternalServerError
	}

	theme, svcErr := drs.themeMgtService.GetTheme(ctx, themeID)
	if svcErr != nil {
		return nil, svcErr
	}

	applicableTheme, err := resolveApplicableTheme(theme.Theme, colorScheme)
	if err != nil {
		drs.logger.Error(ctx, "Failed to resolve applicable theme for design tokens",
			log.String("themeID", themeID), log.Error(err))
		return nil, &tidcommon.InternalServerError
	}

	// rootPrefix is the starting dot-path prefix: top-level theme keys have no parent path.
	const rootPrefix = ""
	return flattenTokens(rootPrefix, applicableTheme), nil
}
