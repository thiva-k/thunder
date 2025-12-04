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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/branding/common"
	"github.com/asgardeo/thunder/internal/system/error/apierror"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

const (
	testAppID      = "app-123"
	testBrandingID = "brand-456"
)

type BrandingResolveHandlerTestSuite struct {
	suite.Suite
	mockService *BrandingResolveServiceInterfaceMock
	handler     *brandingResolveHandler
}

func TestBrandingResolveHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(BrandingResolveHandlerTestSuite))
}

func (suite *BrandingResolveHandlerTestSuite) SetupTest() {
	suite.mockService = NewBrandingResolveServiceInterfaceMock(suite.T())
	suite.handler = newBrandingResolveHandler(suite.mockService)
}

// assertErrorResponse is a helper function to assert error responses in tests.
func (suite *BrandingResolveHandlerTestSuite) assertErrorResponse(
	w *httptest.ResponseRecorder,
	expectedCode string,
	expectedMessage string,
	expectedDescription string,
) {
	suite.Equal(http.StatusBadRequest, w.Code)
	suite.Equal("application/json", w.Header().Get("Content-Type"))

	var errorResp apierror.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	suite.NoError(err)
	suite.Equal(expectedCode, errorResp.Code)
	suite.Equal(expectedMessage, errorResp.Message)
	if expectedDescription != "" {
		suite.Equal(expectedDescription, errorResp.Description)
	}
	suite.mockService.AssertExpectations(suite.T())
}

// assertServiceErrorResponse is a helper function to assert service error responses in tests.
func (suite *BrandingResolveHandlerTestSuite) assertServiceErrorResponse(
	w *httptest.ResponseRecorder,
	expectedError *serviceerror.ServiceError,
	expectedStatusCode int,
) {
	suite.Equal(expectedStatusCode, w.Code)
	suite.Equal("application/json", w.Header().Get("Content-Type"))

	var errorResp apierror.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	suite.NoError(err)
	suite.Equal(expectedError.Code, errorResp.Code)
	suite.Equal(expectedError.Error, errorResp.Message)
	suite.mockService.AssertExpectations(suite.T())
}

// assertNotFoundErrorResponse is a helper function to assert not found error responses in tests.
func (suite *BrandingResolveHandlerTestSuite) assertNotFoundErrorResponse(
	w *httptest.ResponseRecorder,
	expectedError *serviceerror.ServiceError,
) {
	suite.Equal(http.StatusNotFound, w.Code)
	suite.Equal("application/json", w.Header().Get("Content-Type"))

	var errorResp apierror.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	suite.NoError(err)
	suite.Equal(expectedError.Code, errorResp.Code)
	suite.Equal(expectedError.Error, errorResp.Message)
	suite.Equal(expectedError.ErrorDescription, errorResp.Description)
	suite.mockService.AssertExpectations(suite.T())
}

// HandleResolveRequest Tests
func (suite *BrandingResolveHandlerTestSuite) TestHandleResolveRequest_Success() {
	expectedResponse := &common.BrandingResponse{
		ID:          testBrandingID,
		DisplayName: "Test Branding",
		Preferences: json.RawMessage(`{"theme":{"activeColorScheme":"dark"}}`),
	}

	suite.mockService.On("ResolveBranding", common.BrandingResolveTypeAPP, testAppID).Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/branding/resolve?type=APP&id=app-123", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleResolveRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	suite.Equal("application/json", w.Header().Get("Content-Type"))

	var response common.BrandingResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal(expectedResponse.ID, response.ID)
	suite.Equal(expectedResponse.DisplayName, response.DisplayName)
	suite.Equal(string(expectedResponse.Preferences), string(response.Preferences))
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *BrandingResolveHandlerTestSuite) TestHandleResolveRequest_Success_WithLowerCaseType() {
	expectedResponse := &common.BrandingResponse{
		ID:          testBrandingID,
		DisplayName: "Test Branding",
		Preferences: json.RawMessage(`{"theme":{"activeColorScheme":"dark"}}`),
	}

	// Handler converts lowercase "app" to uppercase "APP" before calling service
	suite.mockService.On("ResolveBranding", common.BrandingResolveTypeAPP, testAppID).Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/branding/resolve?type=app&id=app-123", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleResolveRequest(w, req)

	suite.Equal(http.StatusOK, w.Code)
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *BrandingResolveHandlerTestSuite) TestHandleResolveRequest_MissingType() {
	suite.mockService.On("ResolveBranding", common.BrandingResolveType(""), testAppID).
		Return(nil, &common.ErrorInvalidResolveType)

	req := httptest.NewRequest(http.MethodGet, "/branding/resolve?id=app-123", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleResolveRequest(w, req)

	suite.assertErrorResponse(
		w,
		common.ErrorInvalidResolveType.Code,
		common.ErrorInvalidResolveType.Error,
		common.ErrorInvalidResolveType.ErrorDescription,
	)
}

func (suite *BrandingResolveHandlerTestSuite) TestHandleResolveRequest_MissingID() {
	suite.mockService.On("ResolveBranding", common.BrandingResolveTypeAPP, "").Return(nil, &common.ErrorMissingResolveID)

	req := httptest.NewRequest(http.MethodGet, "/branding/resolve?type=APP", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleResolveRequest(w, req)

	suite.assertErrorResponse(
		w,
		common.ErrorMissingResolveID.Code,
		common.ErrorMissingResolveID.Error,
		common.ErrorMissingResolveID.ErrorDescription,
	)
}

func (suite *BrandingResolveHandlerTestSuite) TestHandleResolveRequest_UnsupportedType() {
	suite.mockService.On("ResolveBranding", common.BrandingResolveTypeOU, testAppID).
		Return(nil, &common.ErrorUnsupportedResolveType)

	req := httptest.NewRequest(http.MethodGet, "/branding/resolve?type=OU&id=app-123", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleResolveRequest(w, req)

	suite.assertErrorResponse(
		w,
		common.ErrorUnsupportedResolveType.Code,
		common.ErrorUnsupportedResolveType.Error,
		common.ErrorUnsupportedResolveType.ErrorDescription,
	)
}

func (suite *BrandingResolveHandlerTestSuite) TestHandleResolveRequest_ApplicationHasNoBranding() {
	suite.mockService.On("ResolveBranding", common.BrandingResolveTypeAPP, testAppID).
		Return(nil, &common.ErrorApplicationHasNoBranding)

	req := httptest.NewRequest(http.MethodGet, "/branding/resolve?type=APP&id=app-123", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleResolveRequest(w, req)

	suite.assertNotFoundErrorResponse(w, &common.ErrorApplicationHasNoBranding)
}

func (suite *BrandingResolveHandlerTestSuite) TestHandleResolveRequest_InternalServerError() {
	suite.mockService.On("ResolveBranding", common.BrandingResolveTypeAPP, testAppID).
		Return(nil, &serviceerror.InternalServerError)

	req := httptest.NewRequest(http.MethodGet, "/branding/resolve?type=APP&id=app-123", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleResolveRequest(w, req)

	suite.assertServiceErrorResponse(w, &serviceerror.InternalServerError, http.StatusInternalServerError)
}

func (suite *BrandingResolveHandlerTestSuite) TestHandleResolveRequest_ApplicationNotFound() {
	suite.mockService.On("ResolveBranding", common.BrandingResolveTypeAPP, testAppID).
		Return(nil, &common.ErrorApplicationNotFound)

	req := httptest.NewRequest(http.MethodGet, "/branding/resolve?type=APP&id=app-123", nil)
	w := httptest.NewRecorder()

	suite.handler.HandleResolveRequest(w, req)

	suite.assertNotFoundErrorResponse(w, &common.ErrorApplicationNotFound)
}
