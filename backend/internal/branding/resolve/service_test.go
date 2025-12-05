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

package brandingresolve

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/branding/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/applicationmock"
	"github.com/asgardeo/thunder/tests/mocks/brandingmock"
)

const (
	testAppIDForService      = "app-123"
	testBrandingIDForService = "brand-456"
)

type BrandingResolveServiceTestSuite struct {
	suite.Suite
	mockBrandingService    *brandingmock.BrandingMgtServiceInterfaceMock
	mockApplicationService *applicationmock.ApplicationServiceInterfaceMock
	service                BrandingResolveServiceInterface
}

func TestBrandingResolveServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BrandingResolveServiceTestSuite))
}

func (suite *BrandingResolveServiceTestSuite) SetupTest() {
	suite.mockBrandingService = brandingmock.NewBrandingMgtServiceInterfaceMock(suite.T())
	suite.mockApplicationService = applicationmock.NewApplicationServiceInterfaceMock(suite.T())
	suite.service = newBrandingResolveService(suite.mockBrandingService, suite.mockApplicationService)
}

// ResolveBranding Tests
func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_Success() {
	appID := testAppIDForService
	brandingID := testBrandingIDForService
	resolveType := common.BrandingResolveTypeAPP

	expectedApp := &model.Application{
		ID:         appID,
		Name:       "Test Application",
		BrandingID: brandingID,
	}

	expectedBranding := &common.Branding{
		ID:          brandingID,
		DisplayName: "Test Branding",
		Preferences: json.RawMessage(`{"theme":{"activeColorScheme":"dark"}}`),
	}

	suite.mockApplicationService.On("GetApplication", appID).Return(expectedApp, nil)
	suite.mockBrandingService.On("GetBranding", brandingID).Return(expectedBranding, nil)

	result, err := suite.service.ResolveBranding(resolveType, appID)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(brandingID, result.ID)
	suite.Equal("Test Branding", result.DisplayName)
	suite.Equal(expectedBranding.Preferences, result.Preferences)
	suite.mockApplicationService.AssertExpectations(suite.T())
	suite.mockBrandingService.AssertExpectations(suite.T())
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_LowerCaseType_ReturnsUnsupportedType() {
	appID := testAppIDForService
	// Lowercase type is not normalized by service (handler does this),
	// so "app" != "APP" results in unsupported type error
	resolveType := common.BrandingResolveType("app")

	result, err := suite.service.ResolveBranding(resolveType, appID)

	suite.NotNil(err)
	suite.Nil(result)
	suite.Equal(common.ErrorUnsupportedResolveType.Code, err.Code)
	suite.mockApplicationService.AssertNotCalled(suite.T(), "GetApplication", mock.Anything)
	suite.mockBrandingService.AssertNotCalled(suite.T(), "GetBranding", mock.Anything)
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_EmptyType() {
	appID := testAppIDForService
	resolveType := common.BrandingResolveType("")

	result, err := suite.service.ResolveBranding(resolveType, appID)

	suite.NotNil(err)
	suite.Nil(result)
	suite.Equal(common.ErrorInvalidResolveType.Code, err.Code)
	suite.Equal(common.ErrorInvalidResolveType.Error, err.Error)
	suite.mockApplicationService.AssertNotCalled(suite.T(), "GetApplication", mock.Anything)
	suite.mockBrandingService.AssertNotCalled(suite.T(), "GetBranding", mock.Anything)
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_EmptyID() {
	resolveType := common.BrandingResolveTypeAPP
	id := ""

	result, err := suite.service.ResolveBranding(resolveType, id)

	suite.NotNil(err)
	suite.Nil(result)
	suite.Equal(common.ErrorMissingResolveID.Code, err.Code)
	suite.Equal(common.ErrorMissingResolveID.Error, err.Error)
	suite.mockApplicationService.AssertNotCalled(suite.T(), "GetApplication", mock.Anything)
	suite.mockBrandingService.AssertNotCalled(suite.T(), "GetBranding", mock.Anything)
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_UnsupportedType() {
	appID := testAppIDForService
	resolveType := common.BrandingResolveTypeOU

	result, err := suite.service.ResolveBranding(resolveType, appID)

	suite.NotNil(err)
	suite.Nil(result)
	suite.Equal(common.ErrorUnsupportedResolveType.Code, err.Code)
	suite.Equal(common.ErrorUnsupportedResolveType.Error, err.Error)
	suite.mockApplicationService.AssertNotCalled(suite.T(), "GetApplication", mock.Anything)
	suite.mockBrandingService.AssertNotCalled(suite.T(), "GetBranding", mock.Anything)
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_ApplicationNotFound() {
	appID := testAppIDForService
	resolveType := common.BrandingResolveTypeAPP
	appNotFoundError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "APP-1001",
		Error:            "Application not found",
		ErrorDescription: "The application with the specified id does not exist",
	}

	suite.mockApplicationService.On("GetApplication", appID).Return(nil, appNotFoundError)

	result, err := suite.service.ResolveBranding(resolveType, appID)

	suite.NotNil(err)
	suite.Nil(result)
	// Service should convert APP-1001 to BRD-1014
	suite.Equal(common.ErrorApplicationNotFound.Code, err.Code)
	suite.Equal(common.ErrorApplicationNotFound.Error, err.Error)
	suite.mockApplicationService.AssertExpectations(suite.T())
	suite.mockBrandingService.AssertNotCalled(suite.T(), "GetBranding", mock.Anything)
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_ApplicationHasNoBranding() {
	appID := testAppIDForService
	resolveType := common.BrandingResolveTypeAPP

	expectedApp := &model.Application{
		ID:         appID,
		Name:       "Test Application",
		BrandingID: "", // No branding ID
	}

	suite.mockApplicationService.On("GetApplication", appID).Return(expectedApp, nil)

	result, err := suite.service.ResolveBranding(resolveType, appID)

	suite.NotNil(err)
	suite.Nil(result)
	suite.Equal(common.ErrorApplicationHasNoBranding.Code, err.Code)
	suite.Equal(common.ErrorApplicationHasNoBranding.Error, err.Error)
	suite.mockApplicationService.AssertExpectations(suite.T())
	suite.mockBrandingService.AssertNotCalled(suite.T(), "GetBranding", mock.Anything)
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_BrandingNotFound_ReturnsInternalServerError() {
	appID := testAppIDForService
	brandingID := testBrandingIDForService
	resolveType := common.BrandingResolveTypeAPP

	expectedApp := &model.Application{
		ID:         appID,
		Name:       "Test Application",
		BrandingID: brandingID,
	}

	brandingNotFoundError := &serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1003",
		Error:            "Branding configuration not found",
		ErrorDescription: "The branding configuration with the specified id does not exist",
	}

	suite.mockApplicationService.On("GetApplication", appID).Return(expectedApp, nil)
	suite.mockBrandingService.On("GetBranding", brandingID).Return(nil, brandingNotFoundError)

	result, err := suite.service.ResolveBranding(resolveType, appID)

	suite.NotNil(err)
	suite.Nil(result)
	// Branding not found when app has branding ID is a data integrity issue, should return internal server error
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
	suite.mockApplicationService.AssertExpectations(suite.T())
	suite.mockBrandingService.AssertExpectations(suite.T())
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_ApplicationServiceNil() {
	appID := testAppIDForService
	resolveType := common.BrandingResolveTypeAPP

	// Create service with nil application service
	service := newBrandingResolveService(suite.mockBrandingService, nil)

	result, err := service.ResolveBranding(resolveType, appID)

	suite.NotNil(err)
	suite.Nil(result)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
	suite.mockApplicationService.AssertNotCalled(suite.T(), "GetApplication", mock.Anything)
	suite.mockBrandingService.AssertNotCalled(suite.T(), "GetBranding", mock.Anything)
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_InternalServerErrorFromApplicationService() {
	appID := testAppIDForService
	resolveType := common.BrandingResolveTypeAPP

	suite.mockApplicationService.On("GetApplication", appID).Return(nil, &serviceerror.InternalServerError)

	result, err := suite.service.ResolveBranding(resolveType, appID)

	suite.NotNil(err)
	suite.Nil(result)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
	suite.mockApplicationService.AssertExpectations(suite.T())
	suite.mockBrandingService.AssertNotCalled(suite.T(), "GetBranding", mock.Anything)
}

func (suite *BrandingResolveServiceTestSuite) TestResolveBranding_InternalServerErrorFromBrandingService() {
	appID := testAppIDForService
	brandingID := testBrandingIDForService
	resolveType := common.BrandingResolveTypeAPP

	expectedApp := &model.Application{
		ID:         appID,
		Name:       "Test Application",
		BrandingID: brandingID,
	}

	suite.mockApplicationService.On("GetApplication", appID).Return(expectedApp, nil)
	suite.mockBrandingService.On("GetBranding", brandingID).Return(nil, &serviceerror.InternalServerError)

	result, err := suite.service.ResolveBranding(resolveType, appID)

	suite.NotNil(err)
	suite.Nil(result)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
	suite.mockApplicationService.AssertExpectations(suite.T())
	suite.mockBrandingService.AssertExpectations(suite.T())
}
