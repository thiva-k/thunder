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

package export

import (
	"testing"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testAppID = "test-app-id"
)

// ExportServiceTestSuite defines the test suite for the export service.
type ExportServiceTestSuite struct {
	suite.Suite
	appServiceMock *applicationmock.ApplicationServiceInterfaceMock
	exportService  ExportServiceInterface
}

// SetupTest sets up the test environment before each test.
func (suite *ExportServiceTestSuite) SetupTest() {
	suite.appServiceMock = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.exportService = newExportService(suite.appServiceMock)
}

// TestExportServiceTestSuite runs the test suite.
func TestExportServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ExportServiceTestSuite))
}

// TestExportResources_NilRequest tests ExportResources with nil request.
func (suite *ExportServiceTestSuite) TestExportResources_NilRequest() {
	result, err := suite.exportService.ExportResources(nil)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorInvalidRequest.Code, err.Code)
	assert.Equal(suite.T(), "Invalid export request", err.Error)
}

// TestExportResources_EmptyRequest tests ExportResources with empty request.
func (suite *ExportServiceTestSuite) TestExportResources_EmptyRequest() {
	request := &ExportRequest{}

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
	assert.Equal(suite.T(), "No resources found", err.Error)
}

// TestExportResources_DefaultOptions tests ExportResources with default options.
func (suite *ExportServiceTestSuite) TestExportResources_DefaultOptions() {
	appID := testAppID
	request := &ExportRequest{
		Applications: []string{appID},
	}

	mockApp := &appmodel.Application{
		ID:          appID,
		Name:        "Test App",
		Description: "Test Description",
	}

	suite.appServiceMock.EXPECT().GetApplication(appID).Return(mockApp, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)
	assert.Equal(suite.T(), 1, result.Summary.TotalFiles)
	assert.Contains(suite.T(), result.Summary.ResourceTypes, "applications")
}

// TestExportResources_ApplicationNotFound tests ExportResources when application is not found.
func (suite *ExportServiceTestSuite) TestExportResources_ApplicationNotFound() {
	appID := "non-existent-app"
	request := &ExportRequest{
		Applications: []string{appID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	appError := &serviceerror.ServiceError{
		Code:  "APP_NOT_FOUND",
		Error: "Application not found",
	}

	suite.appServiceMock.EXPECT().GetApplication(appID).Return(nil, appError)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
}

// TestExportResources_CompleteOAuthApplication tests exporting an application with OAuth config.
func (suite *ExportServiceTestSuite) TestExportResources_CompleteOAuthApplication() {
	appID := "oauth-app-id"
	request := &ExportRequest{
		Applications: []string{appID},
		Options: &ExportOptions{
			Format: "yaml",
			FolderStructure: &FolderStructureOptions{
				GroupByType:       true,
				FileNamingPattern: "${name}_${id}",
			},
		},
	}

	mockOAuthConfig := &appmodel.OAuthAppConfigComplete{
		ClientID:                "client123",
		RedirectURIs:            []string{"http://localhost:3000/callback"},
		GrantTypes:              []oauth2const.GrantType{oauth2const.GrantTypeAuthorizationCode},
		ResponseTypes:           []oauth2const.ResponseType{oauth2const.ResponseTypeCode},
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
		PKCERequired:            true,
		PublicClient:            false,
		Scopes:                  []string{"openid", "profile"},
		Token: &appmodel.OAuthTokenConfig{
			Issuer: "https://localhost:8090",
			AccessToken: &appmodel.AccessTokenConfig{
				ValidityPeriod: 3600,
				UserAttributes: []string{"email", "username"},
			},
			IDToken: &appmodel.IDTokenConfig{
				ValidityPeriod: 1800,
				UserAttributes: []string{"email"},
				ScopeClaims: map[string][]string{
					"profile": {"name", "picture"},
				},
			},
		},
	}

	mockApp := &appmodel.Application{
		ID:          appID,
		Name:        "OAuth Test App",
		Description: "OAuth Test Description",
		URL:         "https://example.com",
		InboundAuthConfig: []appmodel.InboundAuthConfigComplete{
			{
				Type:           appmodel.OAuthInboundAuthType,
				OAuthAppConfig: mockOAuthConfig,
			},
		},
		Token: &appmodel.TokenConfig{
			UserAttributes: []string{"email", "username"},
		},
	}

	suite.appServiceMock.EXPECT().GetApplication(appID).Return(mockApp, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1)

	file := result.Files[0]
	assert.Equal(suite.T(), "OAuth_Test_App_oauth-app-id.yaml", file.FileName)
	assert.Equal(suite.T(), "applications", file.FolderPath)
	assert.Contains(suite.T(), file.Content, "name: OAuth Test App")
	assert.Contains(suite.T(), file.Content, "client_id: {{.O_AUTH_TEST_APP_CLIENT_ID}}")
	assert.Contains(suite.T(), file.Content, "client_secret: {{.O_AUTH_TEST_APP_CLIENT_SECRET}}")
	assert.Contains(suite.T(), file.Content, "redirect_uris:")
	assert.Contains(suite.T(), file.Content, "{{- range .O_AUTH_TEST_APP_REDIRECT_URIS}}")

	assert.Equal(suite.T(), 1, result.Summary.ResourceTypes["applications"])
	assert.Equal(suite.T(), int64(len(file.Content)), file.Size)
}

// TestExportResources_MultipleApplications tests exporting multiple applications.
func (suite *ExportServiceTestSuite) TestExportResources_MultipleApplications() {
	app1ID := "app1"
	app2ID := "app2"
	request := &ExportRequest{
		Applications: []string{app1ID, app2ID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockApp1 := &appmodel.Application{
		ID:          app1ID,
		Name:        "App One",
		Description: "First App",
	}

	mockApp2 := &appmodel.Application{
		ID:          app2ID,
		Name:        "App Two",
		Description: "Second App",
	}

	suite.appServiceMock.EXPECT().GetApplication(app1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(app2ID).Return(mockApp2, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2)
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)
	assert.Equal(suite.T(), 2, result.Summary.ResourceTypes["applications"])
}

// TestExportResources_PartialFailure tests exporting when some applications fail.
func (suite *ExportServiceTestSuite) TestExportResources_PartialFailure() {
	app1ID := "valid-app"
	app2ID := "invalid-app"
	request := &ExportRequest{
		Applications: []string{app1ID, app2ID},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	mockApp1 := &appmodel.Application{
		ID:   app1ID,
		Name: "Valid App",
	}

	appError := &serviceerror.ServiceError{
		Code:  "APP_NOT_FOUND",
		Error: "Application not found",
	}

	suite.appServiceMock.EXPECT().GetApplication(app1ID).Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication(app2ID).Return(nil, appError)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 1) // Only one successful export
	assert.Equal(suite.T(), 1, result.Summary.TotalFiles)
	assert.Len(suite.T(), result.Summary.Errors, 1) // One error recorded
	assert.Equal(suite.T(), "application", result.Summary.Errors[0].ResourceType)
	assert.Equal(suite.T(), app2ID, result.Summary.Errors[0].ResourceID)
}

// TestExportResources_WildcardApplications tests exporting all applications using wildcard.
func (suite *ExportServiceTestSuite) TestExportResources_WildcardApplications() {
	request := &ExportRequest{
		Applications: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// Mock GetApplicationList to return 3 applications
	mockAppList := &appmodel.ApplicationListResponse{
		TotalResults: 3,
		Count:        3,
		Applications: []appmodel.BasicApplicationResponse{
			{ID: "app1", Name: "Application One"},
			{ID: "app2", Name: "Application Two"},
			{ID: "app3", Name: "Application Three"},
		},
	}

	mockApp1 := &appmodel.Application{
		ID:          "app1",
		Name:        "Application One",
		Description: "First App",
	}

	mockApp2 := &appmodel.Application{
		ID:          "app2",
		Name:        "Application Two",
		Description: "Second App",
	}

	mockApp3 := &appmodel.Application{
		ID:          "app3",
		Name:        "Application Three",
		Description: "Third App",
	}

	suite.appServiceMock.EXPECT().GetApplicationList().Return(mockAppList, nil)
	suite.appServiceMock.EXPECT().GetApplication("app1").Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication("app2").Return(mockApp2, nil)
	suite.appServiceMock.EXPECT().GetApplication("app3").Return(mockApp3, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 3) // All 3 applications exported
	assert.Equal(suite.T(), 3, result.Summary.TotalFiles)
	assert.Equal(suite.T(), 3, result.Summary.ResourceTypes["applications"])
	assert.Len(suite.T(), result.Summary.Errors, 0) // No errors
}

// TestExportResources_WildcardApplications_ListFailure tests wildcard export when GetApplicationList fails.
func (suite *ExportServiceTestSuite) TestExportResources_WildcardApplications_ListFailure() {
	request := &ExportRequest{
		Applications: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	listError := &serviceerror.ServiceError{
		Code:  "LIST_FAILED",
		Error: "Failed to list applications",
	}

	suite.appServiceMock.EXPECT().GetApplicationList().Return(nil, listError)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
}

// TestExportResources_WildcardApplications_EmptyList tests wildcard export with empty application list.
func (suite *ExportServiceTestSuite) TestExportResources_WildcardApplications_EmptyList() {
	request := &ExportRequest{
		Applications: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// Mock GetApplicationList to return empty list
	mockAppList := &appmodel.ApplicationListResponse{
		TotalResults: 0,
		Count:        0,
		Applications: []appmodel.BasicApplicationResponse{},
	}

	suite.appServiceMock.EXPECT().GetApplicationList().Return(mockAppList, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), ErrorNoResourcesFound.Code, err.Code)
}

// TestExportResources_WildcardApplications_PartialFailure tests wildcard export with partial failures.
func (suite *ExportServiceTestSuite) TestExportResources_WildcardApplications_PartialFailure() {
	request := &ExportRequest{
		Applications: []string{"*"},
		Options: &ExportOptions{
			Format: "yaml",
		},
	}

	// Mock GetApplicationList to return 3 applications
	mockAppList := &appmodel.ApplicationListResponse{
		TotalResults: 3,
		Count:        3,
		Applications: []appmodel.BasicApplicationResponse{
			{ID: "app1", Name: "Application One"},
			{ID: "app2", Name: "Application Two"},
			{ID: "app3", Name: "Application Three"},
		},
	}

	mockApp1 := &appmodel.Application{
		ID:          "app1",
		Name:        "Application One",
		Description: "First App",
	}

	mockApp3 := &appmodel.Application{
		ID:          "app3",
		Name:        "Application Three",
		Description: "Third App",
	}

	appError := &serviceerror.ServiceError{
		Code:  "APP_NOT_FOUND",
		Error: "Application not found",
	}

	suite.appServiceMock.EXPECT().GetApplicationList().Return(mockAppList, nil)
	suite.appServiceMock.EXPECT().GetApplication("app1").Return(mockApp1, nil)
	suite.appServiceMock.EXPECT().GetApplication("app2").Return(nil, appError)
	suite.appServiceMock.EXPECT().GetApplication("app3").Return(mockApp3, nil)

	result, err := suite.exportService.ExportResources(request)

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Files, 2) // 2 successful exports
	assert.Equal(suite.T(), 2, result.Summary.TotalFiles)
	assert.Equal(suite.T(), 2, result.Summary.ResourceTypes["applications"])
	assert.Len(suite.T(), result.Summary.Errors, 1) // One error recorded
	assert.Equal(suite.T(), "application", result.Summary.Errors[0].ResourceType)
	assert.Equal(suite.T(), "app2", result.Summary.Errors[0].ResourceID)
}
