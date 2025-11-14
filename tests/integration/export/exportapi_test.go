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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
)

// ExportAPITestSuite is a test suite for export API tests.
type ExportAPITestSuite struct {
	suite.Suite
}

// TestExportAPITestSuite runs the export API test suite.
func TestExportAPITestSuite(t *testing.T) {
	suite.Run(t, new(ExportAPITestSuite))
}

// SetupSuite sets up the test suite.
func (ts *ExportAPITestSuite) SetupSuite() {
	// Initialize any setup if needed
}

// TearDownSuite tears down the test suite.
func (ts *ExportAPITestSuite) TearDownSuite() {
	// Clean up any resources if needed
}

// TestApplicationExportYAML tests the application export functionality returning YAML.
func (ts *ExportAPITestSuite) TestApplicationExportYAML() {
	// Create a test application first
	app := Application{
		Name:                      "Export Test App",
		Description:               "Test application for export functionality",
		IsRegistrationFlowEnabled: true,
		URL:                       "https://exporttest.example.com",
		LogoURL:                   "https://exporttest.example.com/logo.png",
		AuthFlowGraphID:           "auth_flow_config_basic",
		RegistrationFlowGraphID:   "registration_flow_config_basic",
		Certificate: &ApplicationCert{
			Type:  "NONE",
			Value: "",
		},
		InboundAuthConfig: []InboundAuthConfig{
			{
				Type: "oauth2",
				OAuthAppConfig: &OAuthAppConfig{
					ClientID:                "export_test_client",
					ClientSecret:            "export_test_secret",
					RedirectURIs:            []string{"https://exporttest.example.com/callback"},
					GrantTypes:              []string{"authorization_code", "refresh_token"},
					ResponseTypes:           []string{"code"},
					TokenEndpointAuthMethod: "client_secret_basic",
					PKCERequired:            false,
					PublicClient:            false,
				},
			},
		},
	}

	appID, err := ts.createApplication(app)
	ts.Require().NoError(err)
	defer ts.deleteApplication(appID)

	// Test YAML export functionality
	exportRequest := ExportRequest{
		Applications: []string{appID},
	}

	yamlContent, err := ts.exportResourcesYAML(exportRequest)
	ts.Require().NoError(err)
	ts.Require().NotEmpty(yamlContent)

	// Verify the exported YAML content
	ts.Assert().Contains(yamlContent, "name: Export Test App")
	ts.Assert().Contains(yamlContent, "description: Test application for export functionality")
	ts.Assert().Contains(yamlContent, "client_id: {{.EXPORT_TEST_APP_CLIENT_ID}}")
	ts.Assert().NotContains(yamlContent, "export_test_secret") // Client secret should not be exported
	ts.Assert().Contains(yamlContent, "# File: Export_Test_App.yaml")

	// Test JSON export functionality for backward compatibility
	exportResponse, err := ts.exportResourcesJSON(exportRequest)
	ts.Require().NoError(err)
	ts.Require().NotNil(exportResponse)
	ts.Assert().Len(exportResponse.Files, 1)

	// Verify the exported file
	exportedFile := exportResponse.Files[0]
	ts.Assert().Equal("Export_Test_App.yaml", exportedFile.FileName)
	ts.Assert().Contains(exportedFile.Content, "name: Export Test App")
}

// TestExportWithInvalidApplicationID tests export with invalid application ID.
func (ts *ExportAPITestSuite) TestExportWithInvalidApplicationID() {
	// Test export with invalid application ID
	invalidExportRequest := ExportRequest{
		Applications: []string{"invalid-uuid"},
	}

	_, err := ts.exportResourcesYAML(invalidExportRequest)
	ts.Require().Error(err)
}

// TestExportWithEmptyRequest tests export with empty request.
func (ts *ExportAPITestSuite) TestExportWithEmptyRequest() {
	// Test export with empty request
	emptyExportRequest := ExportRequest{
		Applications: []string{},
	}

	_, err := ts.exportResourcesYAML(emptyExportRequest)
	ts.Require().Error(err)
}

// Helper functions

func (ts *ExportAPITestSuite) createApplication(app Application) (string, error) {
	appJSON, err := json.Marshal(app)
	if err != nil {
		return "", fmt.Errorf("failed to marshal application: %w", err)
	}

	reqBody := bytes.NewReader(appJSON)
	req, err := http.NewRequest("POST", testServerURL+"/applications", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	var createdApp Application
	err = json.NewDecoder(resp.Body).Decode(&createdApp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	id := createdApp.ID
	if id == "" {
		return "", fmt.Errorf("response does not contain id")
	}
	return id, nil
}

func (ts *ExportAPITestSuite) deleteApplication(appID string) error {
	req, err := http.NewRequest("DELETE", testServerURL+"/applications/"+appID, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expected status 204, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}
	return nil
}

func (ts *ExportAPITestSuite) exportResourcesYAML(exportRequest ExportRequest) (string, error) {
	reqJSON, err := json.Marshal(exportRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal export request: %w", err)
	}

	reqBody := bytes.NewReader(reqJSON)
	req, err := http.NewRequest("POST", testServerURL+"/export", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create export request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send export request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 200, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	// Verify Content-Type is application/yaml
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/yaml") {
		return "", fmt.Errorf("expected Content-Type to contain 'application/yaml', got '%s'", contentType)
	}

	yamlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read YAML response: %w", err)
	}

	return string(yamlContent), nil
}

func (ts *ExportAPITestSuite) exportResourcesJSON(exportRequest ExportRequest) (*ExportResponse, error) {
	reqJSON, err := json.Marshal(exportRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal export request: %w", err)
	}

	reqBody := bytes.NewReader(reqJSON)
	req, err := http.NewRequest("POST", testServerURL+"/export/json", reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create export request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send export request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("expected status 200, got %d. Response: %s", resp.StatusCode, string(responseBody))
	}

	var exportResponse ExportResponse
	err = json.NewDecoder(resp.Body).Decode(&exportResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse export response: %w", err)
	}

	return &exportResponse, nil
}
