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

package branding

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	serverconst "github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

const (
	linkRelNext  = "next"
	linkRelPrev  = "prev"
	linkRelFirst = "first"
	linkRelLast  = "last"
)

type BrandingServiceTestSuite struct {
	suite.Suite
	mockStore *brandingStoreInterfaceMock
	service   BrandingServiceInterface
}

func TestBrandingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(BrandingServiceTestSuite))
}

func (suite *BrandingServiceTestSuite) SetupTest() {
	suite.mockStore = newBrandingStoreInterfaceMock(suite.T())
	suite.service = newBrandingService(suite.mockStore)
}

// GetBrandingList Tests
func (suite *BrandingServiceTestSuite) TestGetBrandingList_Success() {
	expectedBrandings := []Branding{
		{ID: "brand1", DisplayName: "Application 1 Branding"},
		{ID: "brand2", DisplayName: "Application 2 Branding"},
	}

	suite.mockStore.On("GetBrandingListCount").Return(2, nil)
	suite.mockStore.On("GetBrandingList", 10, 0).Return(expectedBrandings, nil)

	result, err := suite.service.GetBrandingList(10, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(2, result.TotalResults)
	suite.Equal(2, result.Count)
	suite.Equal(1, result.StartIndex)
	suite.Len(result.Brandings, 2)
	suite.Equal("brand1", result.Brandings[0].ID)
	suite.Equal("Application 1 Branding", result.Brandings[0].DisplayName)
	suite.Equal("brand2", result.Brandings[1].ID)
	suite.Equal("Application 2 Branding", result.Brandings[1].DisplayName)
}

func (suite *BrandingServiceTestSuite) TestGetBrandingList_WithPaginationLinks() {
	expectedBrandings := []Branding{
		{ID: "brand1", DisplayName: "Application 1 Branding"},
	}
	suite.mockStore.On("GetBrandingListCount").Return(15, nil)
	suite.mockStore.On("GetBrandingList", 10, 0).Return(expectedBrandings, nil)

	result, err := suite.service.GetBrandingList(10, 0)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(15, result.TotalResults)
	// Should have "next" link, and possibly "last" link
	suite.GreaterOrEqual(len(result.Links), 1)
	hasNext := false
	for _, link := range result.Links {
		if link.Rel == linkRelNext {
			hasNext = true
		}
	}
	suite.True(hasNext)
}

func (suite *BrandingServiceTestSuite) TestGetBrandingList_WithPrevAndNextLinks() {
	expectedBrandings := []Branding{
		{ID: "brand1", DisplayName: "Application 1 Branding"},
	}
	suite.mockStore.On("GetBrandingListCount").Return(25, nil)
	suite.mockStore.On("GetBrandingList", 10, 10).Return(expectedBrandings, nil)

	result, err := suite.service.GetBrandingList(10, 10)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(25, result.TotalResults)
	suite.GreaterOrEqual(len(result.Links), 2) // Should have "prev" and "next" links
}

func (suite *BrandingServiceTestSuite) TestGetBrandingList_InvalidLimit_Zero() {
	result, err := suite.service.GetBrandingList(0, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidLimit.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestGetBrandingList_InvalidLimit_TooLarge() {
	result, err := suite.service.GetBrandingList(serverconst.MaxPageSize+1, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidLimit.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestGetBrandingList_InvalidOffset_Negative() {
	result, err := suite.service.GetBrandingList(10, -1)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidOffset.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestGetBrandingList_CountError() {
	suite.mockStore.On("GetBrandingListCount").Return(0, errors.New("database error"))

	result, err := suite.service.GetBrandingList(10, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestGetBrandingList_ListError() {
	suite.mockStore.On("GetBrandingListCount").Return(10, nil)
	suite.mockStore.On("GetBrandingList", 10, 0).Return(nil, errors.New("database error"))

	result, err := suite.service.GetBrandingList(10, 0)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// CreateBranding Tests
func (suite *BrandingServiceTestSuite) TestCreateBranding_Success() {
	preferencesJSON := json.RawMessage(`{"theme":{"activeColorScheme":"dark"}}`)
	request := CreateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockStore.On("CreateBranding", mock.AnythingOfType("string"), request).Return(nil)

	result, err := suite.service.CreateBranding(request)

	suite.Nil(err)
	suite.NotNil(result)
	suite.NotEmpty(result.ID)
	suite.Equal("Application 1 Branding", result.DisplayName)
	suite.Equal(preferencesJSON, result.Preferences)
}

func (suite *BrandingServiceTestSuite) TestCreateBranding_MissingDisplayName() {
	request := CreateBrandingRequest{
		DisplayName: "",
		Preferences: json.RawMessage(`{}`),
	}

	result, err := suite.service.CreateBranding(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingDisplayName.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestCreateBranding_MissingPreferences() {
	request := CreateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: json.RawMessage(""),
	}

	result, err := suite.service.CreateBranding(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingPreferences.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestCreateBranding_InvalidJSON() {
	request := CreateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: json.RawMessage("invalid json"),
	}

	result, err := suite.service.CreateBranding(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidPreferences.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestCreateBranding_ArrayInsteadOfObject() {
	request := CreateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: json.RawMessage(`["item1","item2"]`),
	}

	result, err := suite.service.CreateBranding(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidPreferences.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestCreateBranding_PrimitiveInsteadOfObject() {
	request := CreateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: json.RawMessage(`"string"`),
	}

	result, err := suite.service.CreateBranding(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidPreferences.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestCreateBranding_StoreError() {
	preferencesJSON := json.RawMessage(`{"theme":"dark"}`)
	request := CreateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockStore.On("CreateBranding", mock.AnythingOfType("string"), request).Return(errors.New("database error"))

	result, err := suite.service.CreateBranding(request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// GetBranding Tests
func (suite *BrandingServiceTestSuite) TestGetBranding_Success() {
	expectedBranding := Branding{
		ID:          "brand1",
		DisplayName: "Application 1 Branding",
		Preferences: json.RawMessage(`{"theme":"dark"}`),
	}

	suite.mockStore.On("GetBranding", "brand1").Return(expectedBranding, nil)

	result, err := suite.service.GetBranding("brand1")

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("brand1", result.ID)
	suite.Equal("Application 1 Branding", result.DisplayName)
}

func (suite *BrandingServiceTestSuite) TestGetBranding_MissingID() {
	result, err := suite.service.GetBranding("")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingBrandingID.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestGetBranding_NotFound() {
	suite.mockStore.On("GetBranding", "brand1").Return(Branding{}, ErrBrandingNotFound)

	result, err := suite.service.GetBranding("brand1")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorBrandingNotFound.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestGetBranding_StoreError() {
	suite.mockStore.On("GetBranding", "brand1").Return(Branding{}, errors.New("database error"))

	result, err := suite.service.GetBranding("brand1")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// UpdateBranding Tests
func (suite *BrandingServiceTestSuite) TestUpdateBranding_Success() {
	preferencesJSON := json.RawMessage(`{"theme":{"activeColorScheme":"light"}}`)
	request := UpdateBrandingRequest{
		DisplayName: "Application 2 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockStore.On("UpdateBranding", "brand1", request).Return(nil)

	result, err := suite.service.UpdateBranding("brand1", request)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal("brand1", result.ID)
	suite.Equal("Application 2 Branding", result.DisplayName)
	suite.Equal(preferencesJSON, result.Preferences)
}

func (suite *BrandingServiceTestSuite) TestUpdateBranding_MissingID() {
	request := UpdateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: json.RawMessage(`{}`),
	}

	result, err := suite.service.UpdateBranding("", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingBrandingID.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestUpdateBranding_MissingDisplayName() {
	request := UpdateBrandingRequest{
		DisplayName: "",
		Preferences: json.RawMessage(`{}`),
	}

	result, err := suite.service.UpdateBranding("brand1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingDisplayName.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestUpdateBranding_MissingPreferences() {
	request := UpdateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: json.RawMessage(""),
	}

	result, err := suite.service.UpdateBranding("brand1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorMissingPreferences.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestUpdateBranding_InvalidJSON() {
	request := UpdateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: json.RawMessage("invalid"),
	}

	result, err := suite.service.UpdateBranding("brand1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorInvalidPreferences.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestUpdateBranding_NotFound() {
	preferencesJSON := json.RawMessage(`{}`)
	request := UpdateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockStore.On("UpdateBranding", "brand1", request).Return(ErrBrandingNotFound)

	result, err := suite.service.UpdateBranding("brand1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(ErrorBrandingNotFound.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestUpdateBranding_StoreError() {
	preferencesJSON := json.RawMessage(`{}`)
	request := UpdateBrandingRequest{
		DisplayName: "Application 1 Branding",
		Preferences: preferencesJSON,
	}

	suite.mockStore.On("UpdateBranding", "brand1", request).Return(errors.New("database error"))

	result, err := suite.service.UpdateBranding("brand1", request)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// DeleteBranding Tests
func (suite *BrandingServiceTestSuite) TestDeleteBranding_Success() {
	suite.mockStore.On("IsBrandingExist", "brand1").Return(true, nil)
	suite.mockStore.On("GetApplicationsCountByBrandingID", "brand1").Return(0, nil)
	suite.mockStore.On("DeleteBranding", "brand1").Return(nil)

	err := suite.service.DeleteBranding("brand1")

	suite.Nil(err)
}

func (suite *BrandingServiceTestSuite) TestDeleteBranding_MissingID() {
	err := suite.service.DeleteBranding("")

	suite.NotNil(err)
	suite.Equal(ErrorMissingBrandingID.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestDeleteBranding_NotExists() {
	suite.mockStore.On("IsBrandingExist", "brand1").Return(false, nil)

	err := suite.service.DeleteBranding("brand1")

	suite.Nil(err) // Returns nil when branding doesn't exist
}

func (suite *BrandingServiceTestSuite) TestDeleteBranding_CheckExistenceError() {
	suite.mockStore.On("IsBrandingExist", "brand1").Return(false, errors.New("database error"))

	err := suite.service.DeleteBranding("brand1")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestDeleteBranding_HasApplications() {
	suite.mockStore.On("IsBrandingExist", "brand1").Return(true, nil)
	suite.mockStore.On("GetApplicationsCountByBrandingID", "brand1").Return(2, nil)

	err := suite.service.DeleteBranding("brand1")

	suite.NotNil(err)
	suite.Equal(ErrorCannotDeleteBranding.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestDeleteBranding_GetApplicationsCountError() {
	suite.mockStore.On("IsBrandingExist", "brand1").Return(true, nil)
	suite.mockStore.On("GetApplicationsCountByBrandingID", "brand1").Return(0, errors.New("database error"))

	err := suite.service.DeleteBranding("brand1")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

func (suite *BrandingServiceTestSuite) TestDeleteBranding_DeleteError() {
	suite.mockStore.On("IsBrandingExist", "brand1").Return(true, nil)
	suite.mockStore.On("GetApplicationsCountByBrandingID", "brand1").Return(0, nil)
	suite.mockStore.On("DeleteBranding", "brand1").Return(errors.New("database error"))

	err := suite.service.DeleteBranding("brand1")

	suite.NotNil(err)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// IsBrandingExist Tests
func (suite *BrandingServiceTestSuite) TestIsBrandingExist_Success_True() {
	suite.mockStore.On("IsBrandingExist", "brand1").Return(true, nil)

	exists, err := suite.service.IsBrandingExist("brand1")

	suite.Nil(err)
	suite.True(exists)
}

func (suite *BrandingServiceTestSuite) TestIsBrandingExist_Success_False() {
	suite.mockStore.On("IsBrandingExist", "brand1").Return(false, nil)

	exists, err := suite.service.IsBrandingExist("brand1")

	suite.Nil(err)
	suite.False(exists)
}

func (suite *BrandingServiceTestSuite) TestIsBrandingExist_EmptyID() {
	exists, err := suite.service.IsBrandingExist("")

	suite.Nil(err)
	suite.False(exists)
}

func (suite *BrandingServiceTestSuite) TestIsBrandingExist_StoreError() {
	suite.mockStore.On("IsBrandingExist", "brand1").Return(false, errors.New("database error"))

	exists, err := suite.service.IsBrandingExist("brand1")

	suite.NotNil(err)
	suite.False(exists)
	suite.Equal(serviceerror.InternalServerError.Code, err.Code)
}

// buildPaginationLinks Tests
func (suite *BrandingServiceTestSuite) TestBuildPaginationLinks_FirstPage() {
	links := buildPaginationLinks(10, 0, 25)
	// First page with more results: should have "next" and "last" links
	suite.GreaterOrEqual(len(links), 1)
	hasNext := false
	for _, link := range links {
		if link.Rel == linkRelNext {
			hasNext = true
		}
	}
	suite.True(hasNext)
}

func (suite *BrandingServiceTestSuite) TestBuildPaginationLinks_MiddlePage() {
	links := buildPaginationLinks(10, 10, 25)
	suite.GreaterOrEqual(len(links), 2)
	hasPrev := false
	hasNext := false
	for _, link := range links {
		if link.Rel == linkRelPrev {
			hasPrev = true
		}
		if link.Rel == linkRelNext {
			hasNext = true
		}
	}
	suite.True(hasPrev)
	suite.True(hasNext)
}

func (suite *BrandingServiceTestSuite) TestBuildPaginationLinks_LastPage() {
	links := buildPaginationLinks(10, 20, 25)
	suite.GreaterOrEqual(len(links), 1)
	hasPrev := false
	for _, link := range links {
		if link.Rel == linkRelPrev {
			hasPrev = true
		}
		suite.NotEqual(linkRelNext, link.Rel)
	}
	suite.True(hasPrev)
}

func (suite *BrandingServiceTestSuite) TestBuildPaginationLinks_WithFirstAndLast() {
	links := buildPaginationLinks(10, 10, 50)
	hasFirst := false
	hasLast := false
	for _, link := range links {
		if link.Rel == linkRelFirst {
			hasFirst = true
		}
		if link.Rel == linkRelLast {
			hasLast = true
		}
	}
	suite.True(hasFirst)
	suite.True(hasLast)
}

func (suite *BrandingServiceTestSuite) TestBuildPaginationLinks_ExactFit() {
	links := buildPaginationLinks(10, 0, 10)
	suite.Len(links, 0) // No links needed when results fit exactly
}

func (suite *BrandingServiceTestSuite) TestBuildPaginationLinks_LastPageOffsetCalculation() {
	// Test case where lastPageOffset calculation is different from current offset
	links := buildPaginationLinks(10, 5, 25)
	// Should have prev, next, and last links
	hasLast := false
	for _, link := range links {
		if link.Rel == linkRelLast {
			hasLast = true
		}
	}
	suite.True(hasLast)
}
