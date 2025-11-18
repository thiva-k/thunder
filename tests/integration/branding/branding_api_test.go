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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	testServerURL    = "https://localhost:8095"
	brandingBasePath = "/branding"
)

var (
	testBrandingPreferences = json.RawMessage(`{
		"theme": {
			"activeColorScheme": "dark",
			"colorSchemes": {
				"dark": {
					"colors": {
						"primary": {
							"main": "#1976d2",
							"dark": "#0d47a1",
							"contrastText": "#ffffff"
						},
						"secondary": {
							"main": "#9c27b0",
							"dark": "#6a0080",
							"contrastText": "#ffffff"
						}
					}
				}
			}
		}
	}`)

	testBrandingPreferences2 = json.RawMessage(`{
		"theme": {
			"activeColorScheme": "light",
			"colorSchemes": {
				"light": {
					"colors": {
						"primary": {
							"main": "#2196f3",
							"dark": "#1976d2",
							"contrastText": "#ffffff"
						}
					}
				}
			}
		}
	}`)

	testBrandingPreferencesUpdate = json.RawMessage(`{
		"theme": {
			"activeColorScheme": "light",
			"colorSchemes": {
				"light": {
					"colors": {
						"primary": {
							"main": "#42a5f5",
							"dark": "#1976d2",
							"contrastText": "#ffffff"
						}
					}
				},
				"dark": {
					"colors": {
						"primary": {
							"main": "#1976d2",
							"dark": "#0d47a1",
							"contrastText": "#ffffff"
						}
					}
				}
			}
		}
	}`)
)

var (
	sharedBrandingID string // Shared branding created in SetupSuite
)

type BrandingAPITestSuite struct {
	suite.Suite
	client *http.Client
}

func TestBrandingAPITestSuite(t *testing.T) {
	suite.Run(t, new(BrandingAPITestSuite))
}

func (suite *BrandingAPITestSuite) SetupSuite() {
	// Create HTTP client that skips TLS verification for testing
	suite.client = testutils.GetHTTPClient()

	// Create a shared branding that can be used by multiple tests
	sharedBranding := CreateBrandingRequest{
		DisplayName: "Shared Test Branding",
		Preferences: testBrandingPreferences,
	}
	branding, err := suite.createBranding(sharedBranding)
	suite.Require().NoError(err, "Failed to create shared branding")
	sharedBrandingID = branding.ID
}

func (suite *BrandingAPITestSuite) TearDownSuite() {
	// Cleanup
	if sharedBrandingID != "" {
		_ = suite.deleteBranding(sharedBrandingID)
	}
}

// Helper function to create a branding configuration
func (suite *BrandingAPITestSuite) createBranding(request CreateBrandingRequest) (*BrandingResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal branding request: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+brandingBasePath, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return nil, fmt.Errorf("expected status 201, got %d. Code: %s, Message: %s", resp.StatusCode, errResp.Code, errResp.Message)
		}
		return nil, fmt.Errorf("expected status 201, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var branding BrandingResponse
	if err := json.Unmarshal(bodyBytes, &branding); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w. Response: %s", err, string(bodyBytes))
	}

	return &branding, nil
}

// Helper function to get a branding configuration by ID
func (suite *BrandingAPITestSuite) getBranding(id string) (*BrandingResponse, error) {
	req, err := http.NewRequest("GET", testServerURL+brandingBasePath+"/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return nil, fmt.Errorf("expected status 200, got %d. Code: %s, Message: %s", resp.StatusCode, errResp.Code, errResp.Message)
		}
		return nil, fmt.Errorf("expected status 200, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var branding BrandingResponse
	if err := json.Unmarshal(bodyBytes, &branding); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w. Response: %s", err, string(bodyBytes))
	}

	return &branding, nil
}

// Helper function to list branding configurations
func (suite *BrandingAPITestSuite) listBrandings(limit, offset int) (*BrandingListResponse, error) {
	params := url.Values{}
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Add("offset", fmt.Sprintf("%d", offset))
	}

	url := testServerURL + brandingBasePath
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return nil, fmt.Errorf("expected status 200, got %d. Code: %s, Message: %s", resp.StatusCode, errResp.Code, errResp.Message)
		}
		return nil, fmt.Errorf("expected status 200, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var listResponse BrandingListResponse
	if err := json.Unmarshal(bodyBytes, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w. Response: %s", err, string(bodyBytes))
	}

	return &listResponse, nil
}

// Helper function to update a branding configuration
func (suite *BrandingAPITestSuite) updateBranding(id string, request UpdateBrandingRequest) (*BrandingResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal branding request: %w", err)
	}

	req, err := http.NewRequest("PUT", testServerURL+brandingBasePath+"/"+id, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return nil, fmt.Errorf("expected status 200, got %d. Code: %s, Message: %s", resp.StatusCode, errResp.Code, errResp.Message)
		}
		return nil, fmt.Errorf("expected status 200, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	var branding BrandingResponse
	if err := json.Unmarshal(bodyBytes, &branding); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w. Response: %s", err, string(bodyBytes))
	}

	return &branding, nil
}

// Helper function to delete a branding configuration
func (suite *BrandingAPITestSuite) deleteBranding(id string) error {
	req, err := http.NewRequest("DELETE", testServerURL+brandingBasePath+"/"+id, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := suite.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		bodyBytes, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return fmt.Errorf("expected status 204 or 404, got %d. Code: %s, Message: %s", resp.StatusCode, errResp.Code, errResp.Message)
		}
		return fmt.Errorf("expected status 204 or 404, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Create Branding - Success
func (suite *BrandingAPITestSuite) TestCreateBranding_Success() {
	request := CreateBrandingRequest{
		DisplayName: "Test Branding Success",
		Preferences: testBrandingPreferences2,
	}

	branding, err := suite.createBranding(request)
	suite.Require().NoError(err)
	suite.Require().NotNil(branding)

	suite.NotEmpty(branding.ID)
	suite.Equal("Test Branding Success", branding.DisplayName)
	suite.NotEmpty(branding.Preferences)

	// Cleanup
	_ = suite.deleteBranding(branding.ID)
}

// Create Branding - Validation Errors
func (suite *BrandingAPITestSuite) TestCreateBranding_ValidationErrors() {
	testCases := []struct {
		name        string
		requestBody string
		expectedErr string
	}{
		{
			name:        "Missing DisplayName",
			requestBody: `{"preferences": {}}`,
			expectedErr: "BRD-1005",
		},
		{
			name:        "Missing Preferences",
			requestBody: `{"displayName": "Test"}`,
			expectedErr: "BRD-1006",
		},
		{
			name:        "Invalid JSON Preferences",
			requestBody: `{"displayName": "Test", "preferences": invalid json}`,
			expectedErr: "BRD-1001",
		},
		{
			name:        "Array Instead of Object",
			requestBody: `{"displayName": "Test", "preferences": ["item1", "item2"]}`,
			expectedErr: "BRD-1007",
		},
		{
			name:        "Primitive Instead of Object",
			requestBody: `{"displayName": "Test", "preferences": "string"}`,
			expectedErr: "BRD-1007",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, err := http.NewRequest("POST", testServerURL+brandingBasePath, bytes.NewReader([]byte(tc.requestBody)))
			suite.Require().NoError(err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := suite.client.Do(req)
			suite.Require().NoError(err)
			defer resp.Body.Close()

			suite.Equal(http.StatusBadRequest, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			suite.Require().NoError(err)

			var errResp ErrorResponse
			err = json.Unmarshal(bodyBytes, &errResp)
			suite.Require().NoError(err)
			suite.Equal(tc.expectedErr, errResp.Code)
		})
	}
}

// Get Branding - Success
func (suite *BrandingAPITestSuite) TestGetBranding_Success() {
	suite.Require().NotEmpty(sharedBrandingID, "Shared branding must be created in SetupSuite")

	branding, err := suite.getBranding(sharedBrandingID)
	suite.Require().NoError(err)
	suite.Require().NotNil(branding)

	suite.Equal(sharedBrandingID, branding.ID)
	suite.NotEmpty(branding.Preferences)
}

// Get Branding - Not Found
func (suite *BrandingAPITestSuite) TestGetBranding_NotFound() {
	branding, err := suite.getBranding("00000000-0000-0000-0000-000000000000")
	suite.Error(err)
	suite.Nil(branding)
	suite.Contains(err.Error(), "BRD-1003")
}

// List Brandings - Success
func (suite *BrandingAPITestSuite) TestListBrandings_Success() {
	suite.Require().NotEmpty(sharedBrandingID, "Shared branding must be created in SetupSuite")

	response, err := suite.listBrandings(0, 0)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	suite.GreaterOrEqual(response.TotalResults, 1)
	suite.GreaterOrEqual(response.Count, 1)
	suite.NotEmpty(response.Brandings)

	// Verify our shared branding is in the list
	found := false
	for _, branding := range response.Brandings {
		if branding.ID == sharedBrandingID {
			found = true
			suite.NotEmpty(branding.DisplayName)
			break
		}
	}
	suite.True(found, "Shared branding should be in the list")
}

// List Brandings - Pagination
func (suite *BrandingAPITestSuite) TestListBrandings_Pagination() {
	// Create additional brandings for pagination testing
	branding1, err := suite.createBranding(CreateBrandingRequest{
		DisplayName: "Pagination Branding 1",
		Preferences: testBrandingPreferences2,
	})
	suite.Require().NoError(err)
	defer suite.deleteBranding(branding1.ID)

	branding2, err := suite.createBranding(CreateBrandingRequest{
		DisplayName: "Pagination Branding 2",
		Preferences: testBrandingPreferences2,
	})
	suite.Require().NoError(err)
	defer suite.deleteBranding(branding2.ID)

	// Test with limit
	response, err := suite.listBrandings(2, 0)
	suite.Require().NoError(err)
	suite.Require().NotNil(response)

	suite.GreaterOrEqual(response.TotalResults, 3)
	suite.LessOrEqual(response.Count, 2)
	suite.LessOrEqual(len(response.Brandings), 2)

	// Test pagination links
	if response.TotalResults > response.Count {
		suite.NotEmpty(response.Links)
		hasNext := false
		for _, link := range response.Links {
			if link.Rel == "next" {
				hasNext = true
				break
			}
		}
		suite.True(hasNext, "Should have next link when there are more results")
	}
}

// List Brandings - Invalid Pagination Parameters
func (suite *BrandingAPITestSuite) TestListBrandings_InvalidPagination() {
	testCases := []struct {
		name        string
		limit       int
		offset      int
		expectedErr string
	}{
		{
			name:        "Invalid Limit - Zero",
			limit:       0,
			offset:      0,
			expectedErr: "", // When limit is 0, default is applied, so no error
		},
		{
			name:        "Invalid Limit - Negative",
			limit:       -1,
			offset:      0,
			expectedErr: "BRD-1008",
		},
		{
			name:        "Invalid Offset - Negative",
			limit:       10,
			offset:      -1,
			expectedErr: "BRD-1009",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := url.Values{}
			if tc.limit != 0 {
				params.Add("limit", fmt.Sprintf("%d", tc.limit))
			}
			if tc.offset != 0 {
				params.Add("offset", fmt.Sprintf("%d", tc.offset))
			}

			url := testServerURL + brandingBasePath
			if len(params) > 0 {
				url += "?" + params.Encode()
			}

			req, err := http.NewRequest("GET", url, nil)
			suite.Require().NoError(err)

			resp, err := suite.client.Do(req)
			suite.Require().NoError(err)
			defer resp.Body.Close()

			if tc.expectedErr == "" {
				// When limit is 0, default is applied, so expect success
				suite.Equal(http.StatusOK, resp.StatusCode)
			} else {
				suite.Equal(http.StatusBadRequest, resp.StatusCode)

				bodyBytes, err := io.ReadAll(resp.Body)
				suite.Require().NoError(err)

				var errResp ErrorResponse
				err = json.Unmarshal(bodyBytes, &errResp)
				suite.Require().NoError(err)
				suite.Equal(tc.expectedErr, errResp.Code)
			}
		})
	}
}

// Update Branding - Success
func (suite *BrandingAPITestSuite) TestUpdateBranding_Success() {
	// Create a branding for update testing
	branding, err := suite.createBranding(CreateBrandingRequest{
		DisplayName: "Test Branding Update",
		Preferences: testBrandingPreferences,
	})
	suite.Require().NoError(err)
	defer suite.deleteBranding(branding.ID)

	updateRequest := UpdateBrandingRequest{
		DisplayName: "Updated Test Branding",
		Preferences: testBrandingPreferencesUpdate,
	}

	updatedBranding, err := suite.updateBranding(branding.ID, updateRequest)
	suite.Require().NoError(err)
	suite.Require().NotNil(updatedBranding)

	suite.Equal(branding.ID, updatedBranding.ID)
	suite.Equal("Updated Test Branding", updatedBranding.DisplayName)
	suite.NotEmpty(updatedBranding.Preferences)

	// Verify the update by getting the branding again
	retrievedBranding, err := suite.getBranding(branding.ID)
	suite.Require().NoError(err)
	suite.Equal(branding.ID, retrievedBranding.ID)
}

// Update Branding - Not Found
func (suite *BrandingAPITestSuite) TestUpdateBranding_NotFound() {
	updateRequest := UpdateBrandingRequest{
		DisplayName: "Test Branding",
		Preferences: testBrandingPreferencesUpdate,
	}

	branding, err := suite.updateBranding("00000000-0000-0000-0000-000000000000", updateRequest)
	suite.Error(err)
	suite.Nil(branding)
	suite.Contains(err.Error(), "BRD-1003")
}

// Update Branding - Validation Errors
func (suite *BrandingAPITestSuite) TestUpdateBranding_ValidationErrors() {
	// Create a branding for update testing
	branding, err := suite.createBranding(CreateBrandingRequest{
		DisplayName: "Test Branding Validation",
		Preferences: testBrandingPreferences,
	})
	suite.Require().NoError(err)
	defer suite.deleteBranding(branding.ID)

	testCases := []struct {
		name        string
		requestBody string
		expectedErr string
	}{
		{
			name:        "Missing DisplayName",
			requestBody: `{"preferences": {}}`,
			expectedErr: "BRD-1005",
		},
		{
			name:        "Missing Preferences",
			requestBody: `{"displayName": "Test"}`,
			expectedErr: "BRD-1006",
		},
		{
			name:        "Invalid JSON Preferences",
			requestBody: `{"displayName": "Test", "preferences": invalid json}`,
			expectedErr: "BRD-1001",
		},
		{
			name:        "Array Instead of Object",
			requestBody: `{"displayName": "Test", "preferences": ["item1", "item2"]}`,
			expectedErr: "BRD-1007",
		},
		{
			name:        "Primitive Instead of Object",
			requestBody: `{"displayName": "Test", "preferences": "string"}`,
			expectedErr: "BRD-1007",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, err := http.NewRequest("PUT", testServerURL+brandingBasePath+"/"+branding.ID, bytes.NewReader([]byte(tc.requestBody)))
			suite.Require().NoError(err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := suite.client.Do(req)
			suite.Require().NoError(err)
			defer resp.Body.Close()

			suite.Equal(http.StatusBadRequest, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			suite.Require().NoError(err)

			var errResp ErrorResponse
			err = json.Unmarshal(bodyBytes, &errResp)
			suite.Require().NoError(err)
			suite.Equal(tc.expectedErr, errResp.Code)
		})
	}
}

// Delete Branding - Success
func (suite *BrandingAPITestSuite) TestDeleteBranding_Success() {
	// Create a branding for delete testing
	branding, err := suite.createBranding(CreateBrandingRequest{
		DisplayName: "Test Branding Delete",
		Preferences: testBrandingPreferences,
	})
	suite.Require().NoError(err)

	err = suite.deleteBranding(branding.ID)
	suite.NoError(err)

	// Verify deletion by trying to get the branding
	_, err = suite.getBranding(branding.ID)
	suite.Error(err)
	suite.Contains(err.Error(), "BRD-1003")
}

// Delete Branding - Not Found
func (suite *BrandingAPITestSuite) TestDeleteBranding_NotFound() {
	err := suite.deleteBranding("00000000-0000-0000-0000-000000000000")
	// Delete should not error for non-existent branding (returns 204 or 404)
	suite.NoError(err)
}
