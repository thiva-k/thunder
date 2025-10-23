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

package application

import (
	"errors"
	"slices"
	"strings"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/flow/flowmgt"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	oauthutils "github.com/asgardeo/thunder/internal/oauth/oauth2/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/hash"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// ApplicationServiceInterface defines the interface for the application service.
type ApplicationServiceInterface interface {
	CreateApplication(app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	GetApplicationList() (*model.ApplicationListResponse, *serviceerror.ServiceError)
	GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO, *serviceerror.ServiceError)
	GetApplication(appID string) (*model.ApplicationProcessedDTO, *serviceerror.ServiceError)
	UpdateApplication(appID string, app *model.ApplicationDTO) (*model.ApplicationDTO, *serviceerror.ServiceError)
	DeleteApplication(appID string) *serviceerror.ServiceError
}

// ApplicationService is the default implementation of the ApplicationServiceInterface.
type applicationService struct {
	appStore    applicationStoreInterface
	certService cert.CertificateServiceInterface
}

// TODO: this needs to be removed once all the dependencies are refactored to use DI.

// GetApplicationService creates a new instance of ApplicationService.
func GetApplicationService() ApplicationServiceInterface {
	return &applicationService{
		appStore:    newCachedBackedApplicationStore(),
		certService: cert.NewCertificateService(),
	}
}

func newApplicationService(appStore applicationStoreInterface,
	certService cert.CertificateServiceInterface) ApplicationServiceInterface {
	return &applicationService{
		appStore:    appStore,
		certService: certService,
	}
}

// CreateApplication creates the application.
func (as *applicationService) CreateApplication(app *model.ApplicationDTO) (*model.ApplicationDTO,
	*serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if app == nil {
		return nil, &ErrorApplicationNil
	}
	if app.Name == "" {
		return nil, &ErrorInvalidApplicationName
	}

	// Check if an application with the same name already exists
	existingApp, appCheckErr := as.appStore.GetApplicationByName(app.Name)
	if appCheckErr != nil && !errors.Is(appCheckErr, model.ApplicationNotFoundError) {
		logger.Error("Failed to check existing application by name", log.Error(appCheckErr),
			log.String("appName", app.Name))
		return nil, &ErrorInternalServerError
	}
	if existingApp != nil {
		return nil, &ErrorApplicationAlreadyExistsWithName
	}

	inboundAuthConfig, svcErr := validateAndProcessInboundAuthConfig(as.appStore, app, nil, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	if svcErr := validateAuthFlowGraphID(app); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := validateRegistrationFlowGraphID(app); svcErr != nil {
		return nil, svcErr
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, &ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, &ErrorInvalidLogoURL
	}

	appID := sysutils.GenerateUUID()

	// Process token configuration
	rootToken, finalOAuthAccessToken, finalOAuthIDToken, finalOAuthTokenIssuer := processTokenConfiguration(app)

	// Validate and prepare the certificate if provided.
	appCert, svcErr := as.getValidatedCertificateForCreate(appID, app)
	if svcErr != nil {
		return nil, svcErr
	}

	// Validate and prepare the OAuth app certificate if provided.
	var oauthCert *cert.Certificate
	if inboundAuthConfig != nil {
		oauthCert, svcErr = as.getValidatedOAuthAppCertificateForCreate(
			inboundAuthConfig.OAuthAppConfig.ClientID, inboundAuthConfig.OAuthAppConfig)
		if svcErr != nil {
			return nil, svcErr
		}
	}

	processedDTO := &model.ApplicationProcessedDTO{
		ID:                        appID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Token:                     rootToken,
		TosURI:                    app.TosURI,
		PolicyURI:                 app.PolicyURI,
		Contacts:                  app.Contacts,
	}
	if inboundAuthConfig != nil {
		// Wrap the finalOAuthAccessToken and finalOAuthIDToken in OAuthTokenConfig structure
		oAuthTokenConfig := &model.OAuthTokenConfig{
			Issuer:      finalOAuthTokenIssuer,
			AccessToken: finalOAuthAccessToken,
			IDToken:     finalOAuthIDToken,
		}

		processedInboundAuthConfig := model.InboundAuthConfigProcessedDTO{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				HashedClientSecret:      getProcessedClientSecret(inboundAuthConfig.OAuthAppConfig),
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				PKCERequired:            inboundAuthConfig.OAuthAppConfig.PKCERequired,
				PublicClient:            inboundAuthConfig.OAuthAppConfig.PublicClient,
				Token:                   oAuthTokenConfig,
				Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
			},
		}
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
	}

	// Create the application certificate if provided.
	returnCert, svcErr := as.createApplicationCertificate(appCert)
	if svcErr != nil {
		return nil, svcErr
	}

	// Create the OAuth app certificate if provided.
	var returnOAuthCert *model.OAuthAppCertificate
	if oauthCert != nil {
		returnOAuthCert, svcErr = as.createOAuthAppCertificate(oauthCert)
		if svcErr != nil {
			if appCert != nil {
				deleteErr := as.rollbackAppCertificateCreation(appID)
				if deleteErr != nil {
					return nil, deleteErr
				}
			}
			return nil, svcErr
		}
	}

	// Create the application.
	storeErr := as.appStore.CreateApplication(*processedDTO)
	if storeErr != nil {
		logger.Error("Failed to create application", log.Error(storeErr), log.String("appID", appID))

		// Rollback the certificate creation if it was successful.
		if appCert != nil {
			deleteErr := as.rollbackAppCertificateCreation(appID)
			if deleteErr != nil {
				return nil, deleteErr
			}
		}

		// Rollback the OAuth app certificate creation if it was successful.
		if oauthCert != nil {
			deleteErr := as.rollbackOAuthAppCertificateCreation(inboundAuthConfig.OAuthAppConfig.ClientID)
			if deleteErr != nil {
				return nil, deleteErr
			}
		}

		return nil, &ErrorInternalServerError
	}

	returnApp := &model.ApplicationDTO{
		ID:                        appID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Token:                     rootToken,
		Certificate:               returnCert,
		TosURI:                    app.TosURI,
		PolicyURI:                 app.PolicyURI,
		Contacts:                  app.Contacts,
	}
	if inboundAuthConfig != nil {
		// Construct the return DTO with processed token configuration
		returnTokenConfig := &model.OAuthTokenConfig{
			Issuer:      finalOAuthTokenIssuer,
			AccessToken: finalOAuthAccessToken,
			IDToken:     finalOAuthIDToken,
		}

		returnInboundAuthConfig := model.InboundAuthConfigDTO{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				ClientSecret:            inboundAuthConfig.OAuthAppConfig.ClientSecret,
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				PKCERequired:            inboundAuthConfig.OAuthAppConfig.PKCERequired,
				PublicClient:            inboundAuthConfig.OAuthAppConfig.PublicClient,
				Token:                   returnTokenConfig,
				Certificate:             returnOAuthCert,
				Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
			},
		}
		returnApp.InboundAuthConfig = []model.InboundAuthConfigDTO{returnInboundAuthConfig}
	}

	return returnApp, nil
}

// GetApplicationList list the applications.
func (as *applicationService) GetApplicationList() (*model.ApplicationListResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	totalCount, err := as.appStore.GetTotalApplicationCount()
	if err != nil {
		logger.Error("Failed to retrieve total application count", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	applications, err := as.appStore.GetApplicationList()
	if err != nil {
		logger.Error("Failed to retrieve application list", log.Error(err))
		return nil, &ErrorInternalServerError
	}

	applicationList := make([]model.BasicApplicationResponse, 0, len(applications))
	for _, app := range applications {
		applicationList = append(applicationList, buildBasicApplicationResponse(app))
	}

	response := &model.ApplicationListResponse{
		TotalResults: totalCount,
		Count:        len(applications),
		Applications: applicationList,
	}

	return response, nil
}

// buildBasicApplicationResponse builds a basic application response from the processed application DTO.
func buildBasicApplicationResponse(app model.BasicApplicationDTO) model.BasicApplicationResponse {
	return model.BasicApplicationResponse{
		ID:                        app.ID,
		Name:                      app.Name,
		Description:               app.Description,
		ClientID:                  app.ClientID,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
	}
}

// GetOAuthApplication retrieves the OAuth application based on the client id.
func (as *applicationService) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessedDTO,
	*serviceerror.ServiceError) {
	if clientID == "" {
		return nil, &ErrorInvalidClientID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	oauthApp, err := as.appStore.GetOAuthApplication(clientID)
	if err != nil {
		if errors.Is(err, model.ApplicationNotFoundError) {
			return nil, &ErrorApplicationNotFound
		}

		logger.Error("Failed to retrieve OAuth application", log.Error(err),
			log.String("clientID", log.MaskString(clientID)))
		return nil, &ErrorInternalServerError
	}
	if oauthApp == nil {
		return nil, &ErrorApplicationNotFound
	}

	cert, certErr := as.getOAuthAppCertificate(clientID)
	if certErr != nil {
		return nil, certErr
	}
	oauthApp.Certificate = cert

	return oauthApp, nil
}

// GetApplication get the application for given app id.
func (as *applicationService) GetApplication(appID string) (*model.ApplicationProcessedDTO,
	*serviceerror.ServiceError) {
	if appID == "" {
		return nil, &ErrorInvalidApplicationID
	}

	application, err := as.appStore.GetApplicationByID(appID)
	if err != nil {
		return nil, as.handleApplicationRetrievalError(err)
	}

	return as.enrichApplicationWithCertificate(application)
}

// handleApplicationRetrievalError handles common error scenarios when retrieving applications from the
// application store. It maps specific errors, such as ApplicationNotFoundError, to corresponding service errors.
func (as *applicationService) handleApplicationRetrievalError(err error) *serviceerror.ServiceError {
	if errors.Is(err, model.ApplicationNotFoundError) {
		return &ErrorApplicationNotFound
	}
	return &ErrorInternalServerError
}

// enrichApplicationWithCertificate retrieves and adds the certificate to the application.
func (as *applicationService) enrichApplicationWithCertificate(application *model.ApplicationProcessedDTO) (
	*model.ApplicationProcessedDTO, *serviceerror.ServiceError) {
	cert, certErr := as.getApplicationCertificate(application.ID)
	if certErr != nil {
		return nil, certErr
	}
	application.Certificate = cert

	if len(application.InboundAuthConfig) > 0 {
		for i := range application.InboundAuthConfig {
			if application.InboundAuthConfig[i].OAuthAppConfig != nil {
				oauthCert, oauthCertErr := as.getOAuthAppCertificate(application.InboundAuthConfig[i].OAuthAppConfig.ClientID)
				if oauthCertErr != nil {
					return nil, oauthCertErr
				}
				application.InboundAuthConfig[i].OAuthAppConfig.Certificate = oauthCert
			}
		}
	}

	return application, nil
}

// UpdateApplication update the application for given app id.
func (as *applicationService) UpdateApplication(appID string, app *model.ApplicationDTO) (
	*model.ApplicationDTO, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if appID == "" {
		return nil, &ErrorInvalidApplicationID
	}
	if app == nil {
		return nil, &ErrorApplicationNil
	}
	if app.Name == "" {
		return nil, &ErrorInvalidApplicationName
	}

	existingApp, appCheckErr := as.appStore.GetApplicationByID(appID)
	if appCheckErr != nil {
		if errors.Is(appCheckErr, model.ApplicationNotFoundError) {
			return nil, &ErrorApplicationNotFound
		}
		logger.Error("Failed to get existing application", log.Error(appCheckErr), log.String("appID", appID))
		return nil, &ErrorInternalServerError
	}
	if existingApp == nil {
		logger.Debug("Application not found for update", log.String("appID", appID))
		return nil, &ErrorApplicationNotFound
	}

	// If the application name is changed, check if an application with the new name already exists.
	if existingApp.Name != app.Name {
		existingAppWithName, appCheckErr := as.appStore.GetApplicationByName(app.Name)
		if appCheckErr != nil && !errors.Is(appCheckErr, model.ApplicationNotFoundError) {
			logger.Error("Failed to check existing application by name", log.Error(appCheckErr),
				log.String("appName", app.Name))
			return nil, &ErrorInternalServerError
		}
		if existingAppWithName != nil {
			return nil, &ErrorApplicationAlreadyExistsWithName
		}
	}

	inboundAuthConfig, svcErr := validateAndProcessInboundAuthConfig(as.appStore, app, existingApp, logger)
	if svcErr != nil {
		return nil, svcErr
	}

	if svcErr := validateAuthFlowGraphID(app); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := validateRegistrationFlowGraphID(app); svcErr != nil {
		return nil, svcErr
	}

	if app.URL != "" && !sysutils.IsValidURI(app.URL) {
		return nil, &ErrorInvalidApplicationURL
	}
	if app.LogoURL != "" && !sysutils.IsValidURI(app.LogoURL) {
		return nil, &ErrorInvalidLogoURL
	}

	existingCert, updatedCert, returnCert, svcErr := as.updateApplicationCertificate(app)
	if svcErr != nil {
		return nil, svcErr
	}

	var existingOAuthCert, updatedOAuthCert *cert.Certificate
	var returnOAuthCert *model.OAuthAppCertificate
	if inboundAuthConfig != nil {
		existingOAuthCert, updatedOAuthCert, returnOAuthCert, svcErr = as.updateOAuthAppCertificate(
			inboundAuthConfig.OAuthAppConfig)
		if svcErr != nil {
			rollbackErr := as.rollbackApplicationCertificateUpdate(appID, existingCert, updatedCert)
			if rollbackErr != nil {
				return nil, rollbackErr
			}
			return nil, svcErr
		}
	}

	// Process token configuration
	rootToken, finalOAuthAccessToken, finalOAuthIDToken, finalOAuthTokenIssuer := processTokenConfiguration(app)

	processedDTO := &model.ApplicationProcessedDTO{
		ID:                        appID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Token:                     rootToken,
		TosURI:                    app.TosURI,
		PolicyURI:                 app.PolicyURI,
		Contacts:                  app.Contacts,
	}
	if inboundAuthConfig != nil {
		// Wrap the finalOAuthAccessToken and finalOAuthIDToken in OAuthTokenConfig structure
		oAuthTokenConfig := &model.OAuthTokenConfig{
			Issuer:      finalOAuthTokenIssuer,
			AccessToken: finalOAuthAccessToken,
			IDToken:     finalOAuthIDToken,
		}

		processedInboundAuthConfig := model.InboundAuthConfigProcessedDTO{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigProcessedDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				HashedClientSecret:      getProcessedClientSecret(inboundAuthConfig.OAuthAppConfig),
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				PKCERequired:            inboundAuthConfig.OAuthAppConfig.PKCERequired,
				PublicClient:            inboundAuthConfig.OAuthAppConfig.PublicClient,
				Token:                   oAuthTokenConfig,
				Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
			},
		}
		processedDTO.InboundAuthConfig = []model.InboundAuthConfigProcessedDTO{processedInboundAuthConfig}
	}

	storeErr := as.appStore.UpdateApplication(existingApp, processedDTO)
	if storeErr != nil {
		logger.Error("Failed to update application", log.Error(storeErr), log.String("appID", appID))

		rollbackErr := as.rollbackApplicationCertificateUpdate(appID, existingCert, updatedCert)
		if rollbackErr != nil {
			return nil, rollbackErr
		}

		if inboundAuthConfig != nil {
			rollbackErr := as.rollbackOAuthAppCertificateUpdate(
				inboundAuthConfig.OAuthAppConfig.ClientID, existingOAuthCert, updatedOAuthCert)
			if rollbackErr != nil {
				return nil, rollbackErr
			}
		}

		return nil, &ErrorInternalServerError
	}

	returnApp := &model.ApplicationDTO{
		ID:                        appID,
		Name:                      app.Name,
		Description:               app.Description,
		AuthFlowGraphID:           app.AuthFlowGraphID,
		RegistrationFlowGraphID:   app.RegistrationFlowGraphID,
		IsRegistrationFlowEnabled: app.IsRegistrationFlowEnabled,
		URL:                       app.URL,
		LogoURL:                   app.LogoURL,
		Token:                     rootToken,
		Certificate:               returnCert,
		TosURI:                    app.TosURI,
		PolicyURI:                 app.PolicyURI,
		Contacts:                  app.Contacts,
	}
	if inboundAuthConfig != nil {
		// Construct the return DTO with processed token configuration
		returnTokenConfig := &model.OAuthTokenConfig{
			Issuer:      finalOAuthTokenIssuer,
			AccessToken: finalOAuthAccessToken,
			IDToken:     finalOAuthIDToken,
		}

		returnInboundAuthConfig := model.InboundAuthConfigDTO{
			Type: model.OAuthInboundAuthType,
			OAuthAppConfig: &model.OAuthAppConfigDTO{
				AppID:                   appID,
				ClientID:                inboundAuthConfig.OAuthAppConfig.ClientID,
				ClientSecret:            inboundAuthConfig.OAuthAppConfig.ClientSecret,
				RedirectURIs:            inboundAuthConfig.OAuthAppConfig.RedirectURIs,
				GrantTypes:              inboundAuthConfig.OAuthAppConfig.GrantTypes,
				ResponseTypes:           inboundAuthConfig.OAuthAppConfig.ResponseTypes,
				TokenEndpointAuthMethod: inboundAuthConfig.OAuthAppConfig.TokenEndpointAuthMethod,
				PKCERequired:            inboundAuthConfig.OAuthAppConfig.PKCERequired,
				PublicClient:            inboundAuthConfig.OAuthAppConfig.PublicClient,
				Token:                   returnTokenConfig,
				Certificate:             returnOAuthCert,
				Scopes:                  inboundAuthConfig.OAuthAppConfig.Scopes,
			},
		}
		returnApp.InboundAuthConfig = []model.InboundAuthConfigDTO{returnInboundAuthConfig}
	}

	return returnApp, nil
}

// DeleteApplication delete the application for given app id.
func (as *applicationService) DeleteApplication(appID string) *serviceerror.ServiceError {
	if appID == "" {
		return &ErrorInvalidApplicationID
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	// Get the application to retrieve OAuth client ID before deletion
	app, appErr := as.appStore.GetApplicationByID(appID)
	if appErr != nil {
		if errors.Is(appErr, model.ApplicationNotFoundError) {
			logger.Debug("Application not found for the deletion", log.String("appID", appID))
			return nil
		}
		logger.Error("Error while getting the application for deletion", log.Error(appErr), log.String("appID", appID))
		return &ErrorInternalServerError
	}

	// Delete the application from the store
	appErr = as.appStore.DeleteApplication(appID)
	if appErr != nil {
		if errors.Is(appErr, model.ApplicationNotFoundError) {
			logger.Debug("Application not found for the deletion", log.String("appID", appID))
			return nil
		}
		logger.Error("Error while deleting the application", log.Error(appErr), log.String("appID", appID))
		return &ErrorInternalServerError
	}

	// Delete the application certificate
	svcErr := as.deleteApplicationCertificate(appID)
	if svcErr != nil {
		return svcErr
	}

	// Delete the OAuth app certificate if OAuth app exists
	if app != nil && len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil {
		clientID := app.InboundAuthConfig[0].OAuthAppConfig.ClientID
		svcErr := as.deleteOAuthAppCertificate(clientID)
		if svcErr != nil {
			return svcErr
		}
	}

	return nil
}

// validateAuthFlowGraphID validates the auth flow graph ID for the application.
// If the graph ID is not provided, it sets the default authentication flow graph ID.
func validateAuthFlowGraphID(app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.AuthFlowGraphID != "" {
		isValidFlowGraphID := flowmgt.GetFlowMgtService().IsValidGraphID(app.AuthFlowGraphID)
		if !isValidFlowGraphID {
			return &ErrorInvalidAuthFlowGraphID
		}
	} else {
		app.AuthFlowGraphID = getDefaultAuthFlowGraphID()
	}

	return nil
}

// validateRegistrationFlowGraphID validates the registration flow graph ID for the application.
// If the graph ID is not provided, it attempts to infer it from the auth flow graph ID.
func validateRegistrationFlowGraphID(app *model.ApplicationDTO) *serviceerror.ServiceError {
	if app.RegistrationFlowGraphID != "" {
		isValidFlowGraphID := flowmgt.GetFlowMgtService().IsValidGraphID(app.RegistrationFlowGraphID)
		if !isValidFlowGraphID {
			return &ErrorInvalidRegistrationFlowGraphID
		}
	} else {
		if strings.HasPrefix(app.AuthFlowGraphID, model.AuthFlowGraphPrefix) {
			suffix := strings.TrimPrefix(app.AuthFlowGraphID, model.AuthFlowGraphPrefix)
			app.RegistrationFlowGraphID = model.RegistrationFlowGraphPrefix + suffix
		} else {
			return &ErrorInvalidRegistrationFlowGraphID
		}
	}

	return nil
}

// validateOAuthParamsForCreateAndUpdate validates the OAuth parameters for creating or updating an application.
func validateOAuthParamsForCreateAndUpdate(app *model.ApplicationDTO) (*model.InboundAuthConfigDTO,
	*serviceerror.ServiceError) {
	if len(app.InboundAuthConfig) == 0 {
		return nil, nil
	}

	// TODO: Need to refactor when supporting other/multiple inbound auth types.
	if app.InboundAuthConfig[0].Type != model.OAuthInboundAuthType {
		return nil, &ErrorInvalidInboundAuthConfig
	}
	inboundAuthConfig := app.InboundAuthConfig[0]
	if inboundAuthConfig.OAuthAppConfig == nil {
		return nil, &ErrorInvalidInboundAuthConfig
	}

	oauthAppConfig := inboundAuthConfig.OAuthAppConfig

	// Apply defaults for OAuth configuration if not specified.
	if len(oauthAppConfig.GrantTypes) == 0 {
		oauthAppConfig.GrantTypes = []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode}
	}
	if len(oauthAppConfig.ResponseTypes) == 0 {
		if slices.Contains(oauthAppConfig.GrantTypes, oauth2const.GrantTypeAuthorizationCode) {
			oauthAppConfig.ResponseTypes = []oauth2const.ResponseType{oauth2const.ResponseTypeCode}
		}
	}
	if oauthAppConfig.TokenEndpointAuthMethod == "" {
		oauthAppConfig.TokenEndpointAuthMethod = oauth2const.TokenEndpointAuthMethodClientSecretBasic
	}

	// Validate redirect URIs
	if err := validateRedirectURIs(oauthAppConfig); err != nil {
		return nil, err
	}

	// Validate grant types and response types
	if err := validateGrantTypesAndResponseTypes(oauthAppConfig); err != nil {
		return nil, err
	}

	// Validate token endpoint authentication method
	if err := validateTokenEndpointAuthMethod(oauthAppConfig); err != nil {
		return nil, err
	}

	// Validate JWKS configuration
	if err := validateJWKSConfiguration(oauthAppConfig); err != nil {
		return nil, err
	}

	// Validate public client configurations
	if oauthAppConfig.PublicClient {
		if err := validatePublicClientConfiguration(oauthAppConfig); err != nil {
			return nil, err
		}
	}

	return &inboundAuthConfig, nil
}

// validateAndProcessInboundAuthConfig validates and processes inbound auth configuration for
// creating or updating an application.
func validateAndProcessInboundAuthConfig(appStore applicationStoreInterface, app *model.ApplicationDTO,
	existingApp *model.ApplicationProcessedDTO, logger *log.Logger) (
	*model.InboundAuthConfigDTO, *serviceerror.ServiceError) {
	inboundAuthConfig, err := validateOAuthParamsForCreateAndUpdate(app)
	if err != nil {
		return nil, err
	}

	if inboundAuthConfig == nil {
		return nil, nil
	}

	clientID := inboundAuthConfig.OAuthAppConfig.ClientID

	// For update operation
	if existingApp != nil && len(existingApp.InboundAuthConfig) > 0 {
		existingClientID := existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID

		if clientID == "" {
			// Generate OAuth 2.0 compliant client ID with proper entropy and URL-safe format
			generatedClientID, err := oauthutils.GenerateOAuth2ClientID()
			if err != nil {
				logger.Error("Failed to generate OAuth client ID", log.Error(err))
				return nil, &ErrorInternalServerError
			}
			inboundAuthConfig.OAuthAppConfig.ClientID = generatedClientID
		} else if clientID != existingClientID {
			existingAppWithClientID, clientCheckErr := appStore.GetOAuthApplication(clientID)
			if clientCheckErr != nil && !errors.Is(clientCheckErr, model.ApplicationNotFoundError) {
				logger.Error("Failed to check existing application by client ID", log.Error(clientCheckErr),
					log.String("clientID", clientID))
				return nil, &ErrorInternalServerError
			}
			if existingAppWithClientID != nil {
				return nil, &ErrorApplicationAlreadyExistsWithClientID
			}
		}
	} else { // For create operation
		if clientID == "" {
			// Generate OAuth 2.0 compliant client ID with proper entropy and URL-safe format
			generatedClientID, err := oauthutils.GenerateOAuth2ClientID()
			if err != nil {
				logger.Error("Failed to generate OAuth client ID", log.Error(err))
				return nil, &ErrorInternalServerError
			}
			inboundAuthConfig.OAuthAppConfig.ClientID = generatedClientID
		} else {
			existingAppWithClientID, clientCheckErr := appStore.GetOAuthApplication(clientID)
			if clientCheckErr != nil && !errors.Is(clientCheckErr, model.ApplicationNotFoundError) {
				logger.Error("Failed to check existing application by client ID", log.Error(clientCheckErr),
					log.String("clientID", clientID))
				return nil, &ErrorInternalServerError
			}
			if existingAppWithClientID != nil {
				return nil, &ErrorApplicationAlreadyExistsWithClientID
			}
		}
	}

	// Generate OAuth 2.0 compliant client secret with high entropy for security
	// Only generate client secret for confidential clients
	if inboundAuthConfig.OAuthAppConfig.ClientSecret == "" && !inboundAuthConfig.OAuthAppConfig.PublicClient {
		generatedClientSecret, err := oauthutils.GenerateOAuth2ClientSecret()
		if err != nil {
			logger.Error("Failed to generate OAuth client secret", log.Error(err))
			return nil, &ErrorInternalServerError
		}
		inboundAuthConfig.OAuthAppConfig.ClientSecret = generatedClientSecret
	}

	return inboundAuthConfig, nil
}

// getDefaultAuthFlowGraphID returns the configured default authentication flow graph ID.
func getDefaultAuthFlowGraphID() string {
	authFlowConfig := config.GetThunderRuntime().Config.Flow.Authn
	return authFlowConfig.DefaultFlow
}

// getValidatedCertificateForCreate validates and returns the certificate for the application during creation.
func (as *applicationService) getValidatedCertificateForCreate(appID string, app *model.ApplicationDTO) (
	*cert.Certificate, *serviceerror.ServiceError) {
	if app.Certificate == nil || app.Certificate.Type == "" || app.Certificate.Type == cert.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(appID, "", app)
}

// getValidatedCertificateForUpdate validates and returns the certificate for the application during update.
func (as *applicationService) getValidatedCertificateForUpdate(certID string, app *model.ApplicationDTO) (
	*cert.Certificate, *serviceerror.ServiceError) {
	if app.Certificate == nil || app.Certificate.Type == "" || app.Certificate.Type == cert.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedCertificateInput(app.ID, certID, app)
}

// validateAndBuildCertificate validates certificate input and builds a cert.Certificate object.
func validateAndBuildCertificate(
	refType cert.CertificateReferenceType,
	refID, certID string,
	certType cert.CertificateType,
	certValue string,
) (*cert.Certificate, *serviceerror.ServiceError) {
	switch certType {
	case cert.CertificateTypeJWKS:
		if certValue == "" {
			return nil, &ErrorInvalidCertificateValue
		}
		return &cert.Certificate{
			ID:      certID,
			RefType: refType,
			RefID:   refID,
			Type:    cert.CertificateTypeJWKS,
			Value:   certValue,
		}, nil
	case cert.CertificateTypeJWKSURI:
		if !sysutils.IsValidURI(certValue) {
			return nil, &ErrorInvalidJWKSURI
		}
		return &cert.Certificate{
			ID:      certID,
			RefType: refType,
			RefID:   refID,
			Type:    cert.CertificateTypeJWKSURI,
			Value:   certValue,
		}, nil
	default:
		return nil, &ErrorInvalidCertificateType
	}
}

// getValidatedCertificateInput is a helper method that validates and returns the certificate.
func getValidatedCertificateInput(appID, certID string, app *model.ApplicationDTO) (*cert.Certificate,
	*serviceerror.ServiceError) {
	return validateAndBuildCertificate(
		cert.CertificateReferenceTypeApplication,
		appID,
		certID,
		app.Certificate.Type,
		app.Certificate.Value,
	)
}

// createCertificateInternal is a generic helper for creating certificates.
func (as *applicationService) createCertificateInternal(
	certificate *cert.Certificate,
) *serviceerror.ServiceError {
	if certificate == nil {
		return nil
	}

	_, svcErr := as.certService.CreateCertificate(certificate)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to create " + string(certificate.RefType) + " certificate: " + svcErr.ErrorDescription
			return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))
		logger.Error("Failed to create "+string(certificate.RefType)+" certificate", log.Any("serviceError", svcErr))
		return &ErrorCertificateServerError
	}
	return nil
}

// createApplicationCertificate creates a certificate for the application.
func (as *applicationService) createApplicationCertificate(certificate *cert.Certificate) (
	*model.ApplicationCertificate, *serviceerror.ServiceError) {
	if err := as.createCertificateInternal(certificate); err != nil {
		return nil, err
	}
	return buildApplicationCertificate(certificate), nil
}

// rollbackCertificateCreation is a generic helper for rolling back certificate creation.
func (as *applicationService) rollbackCertificateCreation(
	refType cert.CertificateReferenceType,
	refID string,
) *serviceerror.ServiceError {
	deleteErr := as.certService.DeleteCertificateByReference(refType, refID)
	if deleteErr != nil {
		if deleteErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to rollback " + string(refType) + " certificate creation: " + deleteErr.ErrorDescription
			return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		return &ErrorCertificateServerError
	}
	return nil
}

// rollbackAppCertificateCreation rolls back the application certificate creation in case of an error during
// application creation.
func (as *applicationService) rollbackAppCertificateCreation(appID string) *serviceerror.ServiceError {
	return as.rollbackCertificateCreation(cert.CertificateReferenceTypeApplication, appID)
}

// deleteApplicationCertificate deletes the certificate associated with the application.
func (as *applicationService) deleteApplicationCertificate(appID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if certErr := as.certService.DeleteCertificateByReference(
		cert.CertificateReferenceTypeApplication, appID); certErr != nil {
		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to delete application certificate: " +
				certErr.ErrorDescription
			return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		logger.Error("Failed to delete application certificate", log.String("appID", appID),
			log.Any("serviceError", certErr))
		return &ErrorCertificateServerError
	}

	return nil
}

// getCertificateByReference is a generic helper to retrieve certificates.
func (as *applicationService) getCertificateByReference(
	refType cert.CertificateReferenceType, refID string) (*cert.Certificate, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	certificate, certErr := as.certService.GetCertificateByReference(refType, refID)

	if certErr != nil {
		if certErr.Code == cert.ErrorCertificateNotFound.Code {
			return nil, nil // Return nil certificate (not an error)
		}

		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to retrieve " + string(refType) + " certificate: " + certErr.ErrorDescription
			return nil, serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		logger.Error("Failed to retrieve "+string(refType)+" certificate", log.Any("serviceError", certErr),
			log.String("referenceType", string(refType)), log.String("referenceID", refID))
		return nil, &ErrorCertificateServerError
	}

	return certificate, nil
}

// handleCertificateError handles certificate operation errors with appropriate logging and error wrapping.
func (as *applicationService) handleCertificateError(
	svcErr *serviceerror.ServiceError, operation, refType, refID string) *serviceerror.ServiceError {
	if svcErr.Type == serviceerror.ClientErrorType {
		errorDescription := operation + " " + refType + " certificate: " + svcErr.ErrorDescription
		return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))
	logger.Error(operation+" "+refType+" certificate", log.Any("serviceError", svcErr),
		log.String("referenceType", refType), log.String("referenceID", refID))
	return &ErrorCertificateServerError
}

// updateCertificateByReference is a generic helper to update or create certificates.
func (as *applicationService) updateCertificateByReference(
	refType cert.CertificateReferenceType, refID string,
	existingCert, updatedCert *cert.Certificate) *serviceerror.ServiceError {
	if updatedCert != nil {
		if existingCert != nil {
			if _, svcErr := as.certService.UpdateCertificateByID(existingCert.ID, updatedCert); svcErr != nil {
				return as.handleCertificateError(svcErr, "Failed to update", string(refType), refID)
			}
		} else {
			if _, svcErr := as.certService.CreateCertificate(updatedCert); svcErr != nil {
				return as.handleCertificateError(svcErr, "Failed to create", string(refType), refID)
			}
		}
	} else if existingCert != nil {
		if deleteErr := as.certService.DeleteCertificateByReference(refType, refID); deleteErr != nil {
			return as.handleCertificateError(deleteErr, "Failed to delete", string(refType), refID)
		}
	}

	return nil
}

// rollbackCertificateUpdate is a generic helper to rollback certificate updates.
func (as *applicationService) rollbackCertificateUpdate(
	refType cert.CertificateReferenceType, refID string,
	existingCert, updatedCert *cert.Certificate) *serviceerror.ServiceError {
	if updatedCert != nil {
		if existingCert != nil {
			if _, svcErr := as.certService.UpdateCertificateByID(existingCert.ID, existingCert); svcErr != nil {
				return as.handleCertificateError(svcErr, "Failed to revert", string(refType), refID)
			}
		} else {
			if deleteErr := as.certService.DeleteCertificateByReference(refType, refID); deleteErr != nil {
				operation := "Failed to delete " + string(refType) + " certificate after update failure"
				return as.handleCertificateError(deleteErr, operation, "", refID)
			}
		}
	} else if existingCert != nil {
		if _, svcErr := as.certService.CreateCertificate(existingCert); svcErr != nil {
			operation := "Failed to revert " + string(refType) + " certificate creation"
			return as.handleCertificateError(svcErr, operation, "", refID)
		}
	}

	return nil
}

// getApplicationCertificate retrieves the certificate associated with the application.
func (as *applicationService) getApplicationCertificate(appID string) (*model.ApplicationCertificate,
	*serviceerror.ServiceError) {
	certificate, certErr := as.getCertificateByReference(cert.CertificateReferenceTypeApplication, appID)
	if certErr != nil {
		return nil, certErr
	}

	if certificate == nil {
		return &model.ApplicationCertificate{
			Type:  cert.CertificateTypeNone,
			Value: "",
		}, nil
	}

	return &model.ApplicationCertificate{
		Type:  certificate.Type,
		Value: certificate.Value,
	}, nil
}

// updateApplicationCertificate updates the certificate for the application.
func (as *applicationService) updateApplicationCertificate(app *model.ApplicationDTO) (
	*cert.Certificate, *cert.Certificate, *model.ApplicationCertificate, *serviceerror.ServiceError) {
	existingCert, updatedCert, err := as.performCertificateUpdate(
		cert.CertificateReferenceTypeApplication,
		app.ID,
		func(certID string) (*cert.Certificate, *serviceerror.ServiceError) {
			return as.getValidatedCertificateForUpdate(certID, app)
		},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	returnCert := buildApplicationCertificate(updatedCert)
	return existingCert, updatedCert, returnCert, nil
}

// buildCertificateResponse constructs a certificate response with Type and Value fields.
func buildCertificateResponse(updatedCert *cert.Certificate) (cert.CertificateType, string) {
	if updatedCert != nil {
		return updatedCert.Type, updatedCert.Value
	}
	return cert.CertificateTypeNone, ""
}

// buildApplicationCertificate constructs an ApplicationCertificate from a cert.Certificate.
func buildApplicationCertificate(updatedCert *cert.Certificate) *model.ApplicationCertificate {
	certType, certValue := buildCertificateResponse(updatedCert)
	return &model.ApplicationCertificate{
		Type:  certType,
		Value: certValue,
	}
}

// performCertificateUpdate is a generic helper for certificate update operations.
func (as *applicationService) performCertificateUpdate(
	refType cert.CertificateReferenceType,
	refID string,
	validateFn func(certID string) (*cert.Certificate, *serviceerror.ServiceError),
) (*cert.Certificate, *cert.Certificate, *serviceerror.ServiceError) {
	// Retrieve existing certificate
	existingCert, certErr := as.getCertificateByReference(refType, refID)
	if certErr != nil {
		return nil, nil, certErr
	}

	// Get validated updated certificate
	certID := ""
	if existingCert != nil {
		certID = existingCert.ID
	}
	updatedCert, err := validateFn(certID)
	if err != nil {
		return nil, nil, err
	}

	// Perform the update/create/delete operation
	if svcErr := as.updateCertificateByReference(refType, refID, existingCert, updatedCert); svcErr != nil {
		return nil, nil, svcErr
	}

	return existingCert, updatedCert, nil
}

// rollbackApplicationCertificateUpdate rolls back the certificate update for the application in case of an error.
func (as *applicationService) rollbackApplicationCertificateUpdate(appID string,
	existingCert, updatedCert *cert.Certificate) *serviceerror.ServiceError {
	return as.rollbackCertificateUpdate(cert.CertificateReferenceTypeApplication, appID, existingCert, updatedCert)
}

// getValidatedOAuthAppCertificateForCreate validates and returns the OAuth app certificate during creation.
func (as *applicationService) getValidatedOAuthAppCertificateForCreate(clientID string,
	oauthApp *model.OAuthAppConfigDTO) (*cert.Certificate, *serviceerror.ServiceError) {
	if oauthApp.Certificate == nil || oauthApp.Certificate.Type == "" ||
		oauthApp.Certificate.Type == cert.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedOAuthAppCertificateInput(clientID, "", oauthApp)
}

// getValidatedOAuthAppCertificateForUpdate validates and returns the OAuth app certificate during update.
func (as *applicationService) getValidatedOAuthAppCertificateForUpdate(certID string,
	oauthApp *model.OAuthAppConfigDTO) (*cert.Certificate, *serviceerror.ServiceError) {
	if oauthApp.Certificate == nil || oauthApp.Certificate.Type == "" ||
		oauthApp.Certificate.Type == cert.CertificateTypeNone {
		return nil, nil
	}
	return getValidatedOAuthAppCertificateInput(oauthApp.ClientID, certID, oauthApp)
}

// getValidatedOAuthAppCertificateInput is a helper method that validates and returns the OAuth app certificate.
func getValidatedOAuthAppCertificateInput(clientID, certID string,
	oauthApp *model.OAuthAppConfigDTO) (*cert.Certificate, *serviceerror.ServiceError) {
	return validateAndBuildCertificate(
		cert.CertificateReferenceTypeOAuthApp,
		clientID,
		certID,
		oauthApp.Certificate.Type,
		oauthApp.Certificate.Value,
	)
}

// createOAuthAppCertificate creates a certificate for the OAuth application.
func (as *applicationService) createOAuthAppCertificate(certificate *cert.Certificate) (
	*model.OAuthAppCertificate, *serviceerror.ServiceError) {
	if err := as.createCertificateInternal(certificate); err != nil {
		return nil, err
	}
	return buildOAuthAppCertificate(certificate), nil
}

// rollbackOAuthAppCertificateCreation rolls back the OAuth app certificate creation in case of an error.
func (as *applicationService) rollbackOAuthAppCertificateCreation(clientID string) *serviceerror.ServiceError {
	return as.rollbackCertificateCreation(cert.CertificateReferenceTypeOAuthApp, clientID)
}

// deleteOAuthAppCertificate deletes the certificate associated with the OAuth application.
func (as *applicationService) deleteOAuthAppCertificate(clientID string) *serviceerror.ServiceError {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "ApplicationService"))

	if certErr := as.certService.DeleteCertificateByReference(
		cert.CertificateReferenceTypeOAuthApp, clientID); certErr != nil {
		if certErr.Type == serviceerror.ClientErrorType {
			errorDescription := "Failed to delete OAuth app certificate: " +
				certErr.ErrorDescription
			return serviceerror.CustomServiceError(ErrorCertificateClientError, errorDescription)
		}
		logger.Error("Failed to delete OAuth app certificate", log.String("clientID", clientID),
			log.Any("serviceError", certErr))
		return &ErrorCertificateServerError
	}

	return nil
}

// getOAuthAppCertificate retrieves the certificate associated with the OAuth application.
func (as *applicationService) getOAuthAppCertificate(clientID string) (*model.OAuthAppCertificate,
	*serviceerror.ServiceError) {
	certificate, certErr := as.getCertificateByReference(cert.CertificateReferenceTypeOAuthApp, clientID)
	if certErr != nil {
		return nil, certErr
	}

	if certificate == nil {
		return &model.OAuthAppCertificate{
			Type:  cert.CertificateTypeNone,
			Value: "",
		}, nil
	}

	return &model.OAuthAppCertificate{
		Type:  certificate.Type,
		Value: certificate.Value,
	}, nil
}

// updateOAuthAppCertificate updates the certificate for the OAuth application.
func (as *applicationService) updateOAuthAppCertificate(oauthApp *model.OAuthAppConfigDTO) (
	*cert.Certificate, *cert.Certificate, *model.OAuthAppCertificate, *serviceerror.ServiceError) {
	existingCert, updatedCert, err := as.performCertificateUpdate(
		cert.CertificateReferenceTypeOAuthApp,
		oauthApp.ClientID,
		func(certID string) (*cert.Certificate, *serviceerror.ServiceError) {
			return as.getValidatedOAuthAppCertificateForUpdate(certID, oauthApp)
		},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	returnCert := buildOAuthAppCertificate(updatedCert)
	return existingCert, updatedCert, returnCert, nil
}

// buildOAuthAppCertificate constructs an OAuthAppCertificate from a cert.Certificate.
func buildOAuthAppCertificate(updatedCert *cert.Certificate) *model.OAuthAppCertificate {
	certType, certValue := buildCertificateResponse(updatedCert)
	return &model.OAuthAppCertificate{
		Type:  certType,
		Value: certValue,
	}
}

// rollbackOAuthAppCertificateUpdate rolls back the certificate update for the OAuth app in case of an error.
func (as *applicationService) rollbackOAuthAppCertificateUpdate(clientID string,
	existingCert, updatedCert *cert.Certificate) *serviceerror.ServiceError {
	return as.rollbackCertificateUpdate(cert.CertificateReferenceTypeOAuthApp, clientID, existingCert, updatedCert)
}

// getDefaultTokenConfigFromDeployment creates a default token configuration from deployment settings.
func getDefaultTokenConfigFromDeployment() *model.TokenConfig {
	jwtConfig := config.GetThunderRuntime().Config.JWT
	tokenConfig := &model.TokenConfig{
		Issuer:         jwtConfig.Issuer,
		ValidityPeriod: jwtConfig.ValidityPeriod,
	}

	return tokenConfig
}

// processTokenConfiguration processes token configuration for an application, applying defaults where necessary.
func processTokenConfiguration(app *model.ApplicationDTO) (
	*model.TokenConfig, *model.TokenConfig, *model.IDTokenConfig, string) {
	// Resolve root token config
	var rootToken *model.TokenConfig
	if app.Token != nil {
		rootToken = &model.TokenConfig{
			Issuer:         app.Token.Issuer,
			ValidityPeriod: app.Token.ValidityPeriod,
			UserAttributes: app.Token.UserAttributes,
		}

		deploymentDefaults := getDefaultTokenConfigFromDeployment()
		if rootToken.Issuer == "" {
			rootToken.Issuer = deploymentDefaults.Issuer
		}
		if rootToken.ValidityPeriod == 0 {
			rootToken.ValidityPeriod = deploymentDefaults.ValidityPeriod
		}
	} else {
		rootToken = getDefaultTokenConfigFromDeployment()
	}
	if rootToken.UserAttributes == nil {
		rootToken.UserAttributes = make([]string, 0)
	}

	// Resolve OAuth access token config
	var oauthAccessToken *model.TokenConfig
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken != nil {
		oauthAccessToken = app.InboundAuthConfig[0].OAuthAppConfig.Token.AccessToken
	}

	if oauthAccessToken != nil {
		if oauthAccessToken.ValidityPeriod == 0 {
			oauthAccessToken.ValidityPeriod = rootToken.ValidityPeriod
		}
		if oauthAccessToken.UserAttributes == nil {
			oauthAccessToken.UserAttributes = make([]string, 0)
		}
	} else {
		oauthAccessToken = &model.TokenConfig{
			ValidityPeriod: rootToken.ValidityPeriod,
			UserAttributes: rootToken.UserAttributes,
		}
	}

	// Resolve OAuth ID token config
	var oauthIDToken *model.IDTokenConfig
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken != nil {
		oauthIDToken = app.InboundAuthConfig[0].OAuthAppConfig.Token.IDToken
	}

	if oauthIDToken != nil {
		if oauthIDToken.ValidityPeriod == 0 {
			oauthIDToken.ValidityPeriod = rootToken.ValidityPeriod
		}
		if oauthIDToken.UserAttributes == nil {
			oauthIDToken.UserAttributes = make([]string, 0)
		}
		if oauthIDToken.ScopeClaims == nil {
			oauthIDToken.ScopeClaims = make(map[string][]string)
		}
	} else {
		oauthIDToken = &model.IDTokenConfig{
			ValidityPeriod: rootToken.ValidityPeriod,
			UserAttributes: rootToken.UserAttributes,
			ScopeClaims:    make(map[string][]string),
		}
	}

	var tokenIssuer string
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token != nil &&
		app.InboundAuthConfig[0].OAuthAppConfig.Token.Issuer != "" {
		tokenIssuer = app.InboundAuthConfig[0].OAuthAppConfig.Token.Issuer
	} else {
		tokenIssuer = rootToken.Issuer
	}

	return rootToken, oauthAccessToken, oauthIDToken, tokenIssuer
}

// validateRedirectURIs validates redirect URIs format and requirements.
func validateRedirectURIs(oauthConfig *model.OAuthAppConfigDTO) *serviceerror.ServiceError {
	for _, redirectURI := range oauthConfig.RedirectURIs {
		parsedURI, err := sysutils.ParseURL(redirectURI)
		if err != nil {
			return &ErrorInvalidRedirectURI
		}

		if parsedURI.Scheme == "" || parsedURI.Host == "" {
			return &ErrorInvalidRedirectURI
		}

		if parsedURI.Fragment != "" {
			return serviceerror.CustomServiceError(
				ErrorInvalidRedirectURI,
				"Redirect URIs must not contain a fragment component",
			)
		}
	}

	if slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeAuthorizationCode) &&
		len(oauthConfig.RedirectURIs) == 0 {
		return serviceerror.CustomServiceError(
			ErrorInvalidOAuthConfiguration,
			"authorization_code grant type requires redirect URIs",
		)
	}

	return nil
}

// validateGrantTypesAndResponseTypes validates grant types, response types, and their compatibility.
func validateGrantTypesAndResponseTypes(oauthConfig *model.OAuthAppConfigDTO) *serviceerror.ServiceError {
	for _, grantType := range oauthConfig.GrantTypes {
		if !grantType.IsValid() {
			return &ErrorInvalidGrantType
		}
	}

	for _, responseType := range oauthConfig.ResponseTypes {
		if !responseType.IsValid() {
			return &ErrorInvalidResponseType
		}
	}

	if len(oauthConfig.GrantTypes) == 1 &&
		slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeClientCredentials) &&
		len(oauthConfig.ResponseTypes) > 0 {
		return serviceerror.CustomServiceError(
			ErrorInvalidOAuthConfiguration,
			"client_credentials grant type cannot be used with response types",
		)
	}

	if slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeAuthorizationCode) {
		if len(oauthConfig.ResponseTypes) == 0 ||
			!slices.Contains(oauthConfig.ResponseTypes, oauth2const.ResponseTypeCode) {
			return serviceerror.CustomServiceError(
				ErrorInvalidOAuthConfiguration,
				"authorization_code grant type requires 'code' response type",
			)
		}
	}

	return nil
}

// validateTokenEndpointAuthMethod validates the token endpoint authentication method
// and its compatibility with grant types.
func validateTokenEndpointAuthMethod(oauthConfig *model.OAuthAppConfigDTO) *serviceerror.ServiceError {
	if !oauthConfig.TokenEndpointAuthMethod.IsValid() {
		return &ErrorInvalidTokenEndpointAuthMethod
	}

	if slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeClientCredentials) &&
		oauthConfig.TokenEndpointAuthMethod == oauth2const.TokenEndpointAuthMethodNone {
		return serviceerror.CustomServiceError(
			ErrorInvalidOAuthConfiguration,
			"client_credentials grant type cannot use 'none' authentication method",
		)
	}

	return nil
}

// validateJWKSConfiguration validates JWKS certificate configuration according to RFC 7591.
func validateJWKSConfiguration(oauthConfig *model.OAuthAppConfigDTO) *serviceerror.ServiceError {
	if oauthConfig.Certificate == nil {
		return nil
	}

	// Validate certificate type and value
	if oauthConfig.Certificate.Type == cert.CertificateTypeJWKSURI {
		if oauthConfig.Certificate.Value != "" {
			if !strings.HasPrefix(oauthConfig.Certificate.Value, "https://") {
				return &ErrorJWKSUriNotHTTPS
			}
			// Validate that the URI is well-formed
			if _, err := sysutils.ParseURL(oauthConfig.Certificate.Value); err != nil {
				return &ErrorInvalidJWKSURI
			}
		}
	}

	return nil
}

// validatePublicClientConfiguration validates that public client configurations are correct.
func validatePublicClientConfiguration(oauthConfig *model.OAuthAppConfigDTO) *serviceerror.ServiceError {
	if oauthConfig.TokenEndpointAuthMethod != oauth2const.TokenEndpointAuthMethodNone {
		return serviceerror.CustomServiceError(
			ErrorInvalidPublicClientConfiguration,
			"Public clients must use 'none' as token endpoint authentication method",
		)
	}

	if slices.Contains(oauthConfig.GrantTypes, oauth2const.GrantTypeClientCredentials) {
		return serviceerror.CustomServiceError(
			ErrorInvalidPublicClientConfiguration,
			"Public clients cannot use the client_credentials grant type",
		)
	}

	if oauthConfig.ClientSecret != "" {
		return serviceerror.CustomServiceError(
			ErrorInvalidPublicClientConfiguration,
			"Public clients cannot have client secrets",
		)
	}

	return nil
}

// getProcessedClientSecret returns the hashed client secret for confidential clients, empty string for public clients.
func getProcessedClientSecret(oauthConfig *model.OAuthAppConfigDTO) string {
	if oauthConfig.PublicClient {
		return ""
	}
	return hash.GenerateThumbprintFromString(oauthConfig.ClientSecret)
}
