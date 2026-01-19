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

package application

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/mcp/tools/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
)

type ApplicationToolsTestSuite struct {
	suite.Suite
	mockAppService *applicationmock.ApplicationServiceInterfaceMock
	tools          *applicationTools
}

func TestApplicationToolsTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationToolsTestSuite))
}

func (suite *ApplicationToolsTestSuite) SetupTest() {
	suite.mockAppService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.tools = NewApplicationTools(suite.mockAppService)
}

func (suite *ApplicationToolsTestSuite) TestListApplications() {
	mockResponse := &model.ApplicationListResponse{
		TotalResults: 1,
		Applications: []model.BasicApplicationResponse{
			{ID: "app-1", Name: "Test App"},
		},
	}
	suite.mockAppService.EXPECT().GetApplicationList().Return(mockResponse, nil)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, output, err := suite.tools.listApplications(ctx, req, nil)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), 1, output.TotalCount)
	assert.Equal(suite.T(), "app-1", output.Applications[0].ID)
}

func (suite *ApplicationToolsTestSuite) TestListApplications_Error() {
	suite.mockAppService.EXPECT().GetApplicationList().Return(nil,
		&serviceerror.ServiceError{ErrorDescription: "some error"})

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, output, err := suite.tools.listApplications(ctx, req, nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "some error")
	assert.Nil(suite.T(), result)
	assert.Empty(suite.T(), output)
}

func (suite *ApplicationToolsTestSuite) TestGetApplicationByID() {
	mockApp := &model.Application{
		ID:   "app-1",
		Name: "Test App",
	}
	suite.mockAppService.EXPECT().GetApplication("app-1").Return(mockApp, nil)

	result, output, err := suite.tools.getApplicationByID(ctx(), nil, common.IDInput{ID: "app-1"})

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "app-1", output.ID)
}

func (suite *ApplicationToolsTestSuite) TestGetApplicationByID_Error() {
	suite.mockAppService.EXPECT().GetApplication("app-1").Return(nil,
		&serviceerror.ServiceError{ErrorDescription: "not found"})

	result, output, err := suite.tools.getApplicationByID(ctx(), nil, common.IDInput{ID: "app-1"})

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not found")
	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), output)
}

func (suite *ApplicationToolsTestSuite) TestGetApplicationByClientID() {
	mockOAuthApp := &model.OAuthAppConfigProcessedDTO{
		AppID: "app-1",
	}
	mockApp := &model.Application{
		ID:   "app-1",
		Name: "Test App",
	}

	suite.mockAppService.EXPECT().GetOAuthApplication("client-1").Return(mockOAuthApp, nil)
	suite.mockAppService.EXPECT().GetApplication("app-1").Return(mockApp, nil)

	result, output, err := suite.tools.getApplicationByClientID(ctx(), nil, clientIDInput{ClientID: "client-1"})

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "app-1", output.ID)
}

func (suite *ApplicationToolsTestSuite) TestGetApplicationByClientID_OAuthError() {
	suite.mockAppService.EXPECT().GetOAuthApplication("client-1").Return(nil,
		&serviceerror.ServiceError{ErrorDescription: "oauth error"})

	result, output, err := suite.tools.getApplicationByClientID(ctx(), nil, clientIDInput{ClientID: "client-1"})

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "oauth error")
	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), output)
}

func (suite *ApplicationToolsTestSuite) TestGetApplicationByClientID_AppError() {
	mockOAuthApp := &model.OAuthAppConfigProcessedDTO{
		AppID: "app-1",
	}

	suite.mockAppService.EXPECT().GetOAuthApplication("client-1").Return(mockOAuthApp, nil)
	suite.mockAppService.EXPECT().GetApplication("app-1").Return(nil,
		&serviceerror.ServiceError{ErrorDescription: "app error"})

	result, output, err := suite.tools.getApplicationByClientID(ctx(), nil, clientIDInput{ClientID: "client-1"})

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "app error")
	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), output)
}

func (suite *ApplicationToolsTestSuite) TestCreateApplication() {
	input := model.ApplicationDTO{
		Name: "New App",
	}
	createdApp := &model.ApplicationDTO{
		ID:   "app-1",
		Name: "New App",
	}

	// Expect ApplyDefaults to have been called, so input might change slightly (e.g. user_attributes).
	// But in test mock call matching, exact struct might fail if ApplyDefaults modifies it.
	// We use mock.MatchedBy or just mock.AnythingOfType if needed, but simple matching helps.
	// Since ApplyDefaults logic is in common/utils.go and tested there, we assume it modifies input.
	// The service call receives the modified input.

	suite.mockAppService.EXPECT().CreateApplication(mock.MatchedBy(func(arg *model.ApplicationDTO) bool {
		return arg.Name == "New App"
	})).Return(createdApp, nil)

	result, output, err := suite.tools.createApplication(ctx(), nil, input)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "app-1", output.ID)
}

func (suite *ApplicationToolsTestSuite) TestCreateApplication_Error() {
	input := model.ApplicationDTO{Name: "New App"}
	suite.mockAppService.EXPECT().CreateApplication(mock.Anything).Return(nil,
		&serviceerror.ServiceError{ErrorDescription: "create error"})

	result, output, err := suite.tools.createApplication(ctx(), nil, input)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "create error")
	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), output)
}

func (suite *ApplicationToolsTestSuite) TestUpdateApplication() {
	input := model.ApplicationDTO{
		ID:   "app-1",
		Name: "Updated App",
	}
	updatedApp := &model.ApplicationDTO{
		ID:   "app-1",
		Name: "Updated App",
	}

	suite.mockAppService.EXPECT().UpdateApplication("app-1", &input).Return(updatedApp, nil)

	result, output, err := suite.tools.updateApplication(ctx(), nil, input)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), "Updated App", output.Name)
}

func (suite *ApplicationToolsTestSuite) TestUpdateApplication_Error() {
	input := model.ApplicationDTO{
		ID:   "app-1",
		Name: "Updated App",
	}
	suite.mockAppService.EXPECT().UpdateApplication("app-1", &input).Return(nil,
		&serviceerror.ServiceError{ErrorDescription: "update error"})

	result, output, err := suite.tools.updateApplication(ctx(), nil, input)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "update error")
	assert.Nil(suite.T(), result)
	assert.Nil(suite.T(), output)
}

func (suite *ApplicationToolsTestSuite) TestGetApplicationTemplates() {
	result, templates, err := suite.tools.getApplicationTemplates(ctx(), nil, nil)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.NotNil(suite.T(), templates)
	assert.Contains(suite.T(), templates, "spa")
	assert.Contains(suite.T(), templates, "mobile")
	assert.Contains(suite.T(), templates, "server")
	assert.Contains(suite.T(), templates, "m2m")
}

func (suite *ApplicationToolsTestSuite) TestRegisterTools() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}, nil)

	// Just verifying it runs without panic and registers
	suite.tools.RegisterTools(server)
}

// Helpers
func ctx() context.Context {
	return context.Background()
}
