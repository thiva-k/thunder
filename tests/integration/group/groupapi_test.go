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

package group

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
)

var (
	testOU = "456e8400-e29b-41d4-a716-446655440001"

	groupToCreate = CreateGroupRequest{
		Name:               "Test Group",
		OrganizationUnitId: testOU,
		Members: []Member{
			{
				Id:   "550e8400-e29b-41d4-a716-446655440000",
				Type: MemberTypeUser,
			},
		},
	}
)

var createdGroupID string

type GroupAPITestSuite struct {
	suite.Suite
}

func (suite *GroupAPITestSuite) SetupSuite() {
	id, err := createGroup(suite)
	if err != nil {
		suite.T().Fatalf("Failed to create group during setup: %v", err)
	} else {
		createdGroupID = id
	}
}

func (suite *GroupAPITestSuite) TearDownSuite() {
	if createdGroupID != "" {
		err := deleteGroup(createdGroupID)
		if err != nil {
			suite.T().Fatalf("Failed to delete group during teardown: %v", err)
		}
	}
}

func (suite *GroupAPITestSuite) TestGetGroup() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for retrieval")
	}

	// Get the created group
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups/"+createdGroupID, nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			suite.T().Logf("Failed to close response body: %v", err)
		}
	}()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.T().Fatalf("Failed to read response body: %v", err)
	}

	var retrievedGroup Group
	err = json.Unmarshal(body, &retrievedGroup)
	suite.Require().NoError(err)

	// Verify the retrieved group
	createdGroup := buildCreatedGroup()
	suite.Equal(createdGroup.Id, retrievedGroup.Id)
	suite.Equal(createdGroup.Name, retrievedGroup.Name)
	suite.Equal(createdGroup.OrganizationUnitId, retrievedGroup.OrganizationUnitId)
}

func (suite *GroupAPITestSuite) TestListGroups() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available, group creation failed in setup")
	}

	// List groups
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.T().Fatalf("Failed to read response body: %v", err)
	}

	var groupListResponse GroupListResponse
	err = json.Unmarshal(body, &groupListResponse)
	suite.Require().NoError(err)

	// Verify response structure
	suite.GreaterOrEqual(groupListResponse.TotalResults, 1, "Should have at least one group")
	suite.Equal(1, groupListResponse.StartIndex, "StartIndex should be 1 for non-paginated request")
	suite.Equal(groupListResponse.TotalResults, groupListResponse.Count, "Count should equal TotalResults for non-paginated request")
	suite.Equal(len(groupListResponse.Groups), groupListResponse.Count, "Groups array length should match Count")
	suite.Equal(0, len(groupListResponse.Links), "Links should be empty for non-paginated request")

	// Verify the list contains our created group
	found := false
	createdGroup := buildCreatedGroup()
	for _, group := range groupListResponse.Groups {
		if group.Id == createdGroup.Id {
			found = true
			suite.Equal(createdGroup.Name, group.Name)
			suite.Equal(createdGroup.OrganizationUnitId, group.OrganizationUnitId)
			break
		}
	}
	suite.True(found, "Created group should be in the list")
}

func (suite *GroupAPITestSuite) TestListGroupsWithPagination() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available, group creation failed in setup")
	}

	// Test pagination with limit=1, offset=0
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups?limit=1&offset=0", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var groupListResponse GroupListResponse
	err = json.Unmarshal(body, &groupListResponse)
	suite.Require().NoError(err)

	// Verify pagination structure
	suite.GreaterOrEqual(groupListResponse.TotalResults, 1, "Should have at least one group")
	suite.Equal(1, groupListResponse.StartIndex, "StartIndex should be 1 for offset=0")
	suite.LessOrEqual(groupListResponse.Count, 1, "Count should be at most 1 due to limit=1")
	suite.LessOrEqual(len(groupListResponse.Groups), 1, "Should return at most 1 group")

	// Verify links structure when there might be more pages
	if groupListResponse.TotalResults > 1 {
		suite.NotEmpty(groupListResponse.Links, "Should have pagination links when there are more results")

		// Check for next link
		hasNext := false
		for _, link := range groupListResponse.Links {
			if link.Rel == "next" {
				hasNext = true
				suite.Contains(link.Href, "offset=1", "Next link should have offset=1")
				suite.Contains(link.Href, "limit=1", "Next link should have limit=1")
			}
		}
		suite.True(hasNext, "Should have next link when there are more results")
	}
}

func (suite *GroupAPITestSuite) TestListGroupsWithInvalidPagination() {
	// Test with invalid limit parameter
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Test invalid limit (negative)
	req, err := http.NewRequest("GET", testServerURL+"/groups?limit=-1", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp map[string]interface{}
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("GRP-1011", errorResp["code"])
	suite.Equal("Invalid limit parameter", errorResp["message"])

	// Test invalid offset (negative)
	req, err = http.NewRequest("GET", testServerURL+"/groups?offset=-1", nil)
	suite.Require().NoError(err)

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("GRP-1012", errorResp["code"])
	suite.Equal("Invalid offset parameter", errorResp["message"])

	// Test invalid limit (too large)
	req, err = http.NewRequest("GET", testServerURL+"/groups?limit=101", nil)
	suite.Require().NoError(err)

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("GRP-1011", errorResp["code"])
	suite.Equal("Invalid limit parameter", errorResp["message"])
}

func (suite *GroupAPITestSuite) TestListGroupsWithOnlyOffset() {
	// Test with only offset parameter provided (should use default limit=30)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups?offset=0", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var groupListResponse GroupListResponse
	err = json.Unmarshal(body, &groupListResponse)
	suite.Require().NoError(err)

	// Verify that pagination structure is present (should use default limit=30)
	suite.GreaterOrEqual(groupListResponse.TotalResults, 1, "Should have at least one group")
	suite.Equal(1, groupListResponse.StartIndex, "StartIndex should be 1 for offset=0")
	suite.LessOrEqual(groupListResponse.Count, 30, "Count should be at most 30 due to default limit")
}

func (suite *GroupAPITestSuite) TestUpdateGroup() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for update")
	}

	// Update the group
	updateRequest := UpdateGroupRequest{
		Name:               "Updated Test Group",
		OrganizationUnitId: testOU,
		Members:            []Member{}, // Empty members list
	}

	jsonData, err := json.Marshal(updateRequest)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("PUT", testServerURL+"/groups/"+createdGroupID, bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var updatedGroup Group
	err = json.Unmarshal(body, &updatedGroup)
	suite.Require().NoError(err)

	// Verify the update
	suite.Equal(createdGroupID, updatedGroup.Id)
	suite.Equal("Updated Test Group", updatedGroup.Name)
}

func (suite *GroupAPITestSuite) TestDeleteGroup() {
	// Create a temporary group for this test since we don't want to delete the main test group
	tempGroupToCreate := CreateGroupRequest{
		Name:               "Temp Test Group",
		OrganizationUnitId: testOU,
		Members:            []Member{},
	}

	jsonData, err := json.Marshal(tempGroupToCreate)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Create temporary group
	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	var tempGroup Group
	err = json.NewDecoder(resp.Body).Decode(&tempGroup)
	suite.Require().NoError(err)

	// Delete the temporary group
	req, err = http.NewRequest("DELETE", testServerURL+"/groups/"+tempGroup.Id, nil)
	suite.Require().NoError(err)

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify the group is deleted by trying to get it
	getReq, err := http.NewRequest("GET", testServerURL+"/groups/"+tempGroup.Id, nil)
	suite.Require().NoError(err)

	getResp, err := client.Do(getReq)
	if err != nil {
		suite.T().Fatalf("Failed to execute GET request: %v", err)
	}
	defer getResp.Body.Close()

	suite.Equal(http.StatusNotFound, getResp.StatusCode)
}

func (suite *GroupAPITestSuite) TestGetNonExistentGroup() {
	// Try to get a non-existent group
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups/non-existent-id", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	if err != nil {
		suite.T().Fatalf("Failed to execute GET request: %v", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *GroupAPITestSuite) TestCreateGroupWithInvalidData() {
	// Try to create a group with invalid data (missing name)
	invalidGroup := map[string]interface{}{
		"organizationUnitId": testOU,
	}

	jsonData, err := json.Marshal(invalidGroup)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		suite.T().Fatalf("Failed to execute POST request: %v", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (suite *GroupAPITestSuite) TestCreateGroupWithInvalidUserID() {
	// Try to create a group with an invalid user ID
	invalidGroup := CreateGroupRequest{
		Name:               "Group with Invalid User",
		OrganizationUnitId: testOU,
		Members: []Member{
			{
				Id:   "invalid-user-id-12345",
				Type: MemberTypeUser,
			},
		},
	}

	jsonData, err := json.Marshal(invalidGroup)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	// Verify the error response
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp map[string]interface{}
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("GRP-1007", errorResp["code"])
	suite.Equal("Invalid user member ID", errorResp["message"])
	suite.Contains(errorResp["description"], "One or more user member IDs in the request do not exist")
}

func (suite *GroupAPITestSuite) TestCreateGroupWithMixedValidInvalidUserIDs() {
	// Try to create a group with a mix of valid and invalid user IDs
	invalidGroup := CreateGroupRequest{
		Name:               "Group with Mixed User IDs",
		OrganizationUnitId: testOU,
		Members: []Member{
			{
				Id:   "550e8400-e29b-41d4-a716-446655440000", // This might be valid from setup
				Type: MemberTypeUser,
			},
			{
				Id:   "invalid-user-id-12345", // This is invalid
				Type: MemberTypeUser,
			},
			{
				Id:   "another-invalid-user-67890", // This is also invalid
				Type: MemberTypeUser,
			},
		},
	}

	jsonData, err := json.Marshal(invalidGroup)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	// Verify the error response
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp map[string]interface{}
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("GRP-1007", errorResp["code"])
	suite.Equal("Invalid user member ID", errorResp["message"])
}

func (suite *GroupAPITestSuite) TestCreateGroupWithEmptyUserList() {
	// Create a group with empty user list (should succeed)
	validGroup := CreateGroupRequest{
		Name:               "Group with Empty Users",
		OrganizationUnitId: testOU,
		Members:            []Member{}, // Empty members list
	}

	jsonData, err := json.Marshal(validGroup)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	// Clean up: get the created group ID and delete it
	var createdGroup Group
	err = json.NewDecoder(resp.Body).Decode(&createdGroup)
	suite.Require().NoError(err)

	// Delete the temporary group
	deleteErr := deleteGroup(createdGroup.Id)
	suite.Require().NoError(deleteErr)
}

func (suite *GroupAPITestSuite) TestUpdateGroupWithInvalidUserID() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for update test")
	}

	// Try to update the group with an invalid user ID
	updateRequest := UpdateGroupRequest{
		Name:               "Updated Group with Invalid User",
		OrganizationUnitId: testOU,
		Members: []Member{
			{
				Id:   "invalid-user-id-update",
				Type: MemberTypeUser,
			},
		},
	}

	jsonData, err := json.Marshal(updateRequest)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("PUT", testServerURL+"/groups/"+createdGroupID, bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	// Verify the error response
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp map[string]interface{}
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("GRP-1007", errorResp["code"])
	suite.Equal("Invalid user member ID", errorResp["message"])
	suite.Contains(errorResp["description"], "One or more user member IDs in the request do not exist")
}

func (suite *GroupAPITestSuite) TestUpdateGroupWithValidEmptyUserList() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for update test")
	}

	// Update the group with empty user list (should succeed)
	updateRequest := UpdateGroupRequest{
		Name:               "Updated Group with Empty Users",
		OrganizationUnitId: testOU,
		Members:            []Member{}, // Empty members list
	}

	jsonData, err := json.Marshal(updateRequest)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("PUT", testServerURL+"/groups/"+createdGroupID, bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	// Verify the update
	var updatedGroup Group
	err = json.NewDecoder(resp.Body).Decode(&updatedGroup)
	suite.Require().NoError(err)

	suite.Equal(createdGroupID, updatedGroup.Id)
	suite.Equal("Updated Group with Empty Users", updatedGroup.Name)
}

func (suite *GroupAPITestSuite) TestUpdateGroupWithMultipleInvalidUserIDs() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for update test")
	}

	// Try to update the group with multiple invalid user IDs
	updateRequest := UpdateGroupRequest{
		Name:               "Updated Group with Multiple Invalid Users",
		OrganizationUnitId: testOU,
		Members: []Member{
			{
				Id:   "invalid-user-1",
				Type: MemberTypeUser,
			},
			{
				Id:   "invalid-user-2",
				Type: MemberTypeUser,
			},
			{
				Id:   "invalid-user-3",
				Type: MemberTypeUser,
			},
		},
	}

	jsonData, err := json.Marshal(updateRequest)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("PUT", testServerURL+"/groups/"+createdGroupID, bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	// Verify the error response
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp map[string]interface{}
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	suite.Equal("GRP-1007", errorResp["code"])
	suite.Equal("Invalid user member ID", errorResp["message"])
}

func (suite *GroupAPITestSuite) TestCreateGroupWithMultipleMembers() {
	// Create a temporary user for testing
	testUserID, err := createTestUser()
	if err != nil {
		suite.T().Fatalf("Failed to create test user: %v", err)
	}
	defer func() {
		if deleteErr := deleteTestUser(testUserID); deleteErr != nil {
			suite.T().Logf("Failed to clean up test user: %v", deleteErr)
		}
	}()

	// Create a group with multiple members (user + other members)
	groupWithMembers := CreateGroupRequest{
		Name:               "Group with Multiple Members",
		OrganizationUnitId: testOU,
		Members: []Member{
			{
				Id:   testUserID,
				Type: MemberTypeUser,
			},
			{
				Id:   "550e8400-e29b-41d4-a716-446655440000", // Another user ID
				Type: MemberTypeUser,
			},
		},
	}

	jsonData, err := json.Marshal(groupWithMembers)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	var createdGroup Group
	err = json.NewDecoder(resp.Body).Decode(&createdGroup)
	suite.Require().NoError(err)

	// Verify the created group has the correct members
	suite.Equal(2, len(createdGroup.Members))

	// Verify member types and IDs
	memberIDs := make(map[string]MemberType)
	for _, member := range createdGroup.Members {
		memberIDs[member.Id] = member.Type
	}

	suite.Equal(MemberTypeUser, memberIDs[testUserID])
	suite.Equal(MemberTypeUser, memberIDs["550e8400-e29b-41d4-a716-446655440000"])

	// Clean up: delete the created group
	deleteErr := deleteGroup(createdGroup.Id)
	suite.Require().NoError(deleteErr)
}

func (suite *GroupAPITestSuite) TestUpdateGroupMembers() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for member update test")
	}

	// Create a temporary user for testing
	testUserID, err := createTestUser()
	if err != nil {
		suite.T().Fatalf("Failed to create test user: %v", err)
	}
	defer func() {
		if deleteErr := deleteTestUser(testUserID); deleteErr != nil {
			suite.T().Logf("Failed to clean up test user: %v", deleteErr)
		}
	}()

	// Update the group to add new members
	updateRequest := UpdateGroupRequest{
		Name:               "Updated Group with New Members",
		OrganizationUnitId: testOU,
		Members: []Member{
			{
				Id:   testUserID,
				Type: MemberTypeUser,
			},
		},
	}

	jsonData, err := json.Marshal(updateRequest)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("PUT", testServerURL+"/groups/"+createdGroupID, bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var updatedGroup Group
	err = json.NewDecoder(resp.Body).Decode(&updatedGroup)
	suite.Require().NoError(err)

	// Verify the update
	suite.Equal(createdGroupID, updatedGroup.Id)
	suite.Equal("Updated Group with New Members", updatedGroup.Name)
	suite.Equal(1, len(updatedGroup.Members))
	suite.Equal(testUserID, updatedGroup.Members[0].Id)
	suite.Equal(MemberTypeUser, updatedGroup.Members[0].Type)
}

func (suite *GroupAPITestSuite) TestCreateGroupWithGroupMember() {
	// First create a temporary group that will be used as a member
	tempGroup := CreateGroupRequest{
		Name:               "Temp Member Group",
		OrganizationUnitId: testOU,
		Members:            []Member{},
	}

	jsonData, err := json.Marshal(tempGroup)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Create the member group
	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	var memberGroup Group
	err = json.NewDecoder(resp.Body).Decode(&memberGroup)
	suite.Require().NoError(err)

	// Clean up member group later
	defer func() {
		if deleteErr := deleteGroup(memberGroup.Id); deleteErr != nil {
			suite.T().Logf("Failed to clean up member group: %v", deleteErr)
		}
	}()

	// Now create a parent group that includes the first group as a member
	parentGroup := CreateGroupRequest{
		Name:               "Parent Group with Group Member",
		OrganizationUnitId: testOU,
		Members: []Member{
			{
				Id:   memberGroup.Id,
				Type: MemberTypeGroup,
			},
		},
	}

	jsonData, err = json.Marshal(parentGroup)
	suite.Require().NoError(err)

	req, err = http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	var createdParentGroup Group
	err = json.NewDecoder(resp.Body).Decode(&createdParentGroup)
	suite.Require().NoError(err)

	// Verify the parent group has the correct member
	suite.Equal(1, len(createdParentGroup.Members))
	suite.Equal(memberGroup.Id, createdParentGroup.Members[0].Id)
	suite.Equal(MemberTypeGroup, createdParentGroup.Members[0].Type)

	// Clean up: delete the parent group
	deleteErr := deleteGroup(createdParentGroup.Id)
	suite.Require().NoError(deleteErr)
}

func (suite *GroupAPITestSuite) TestCreateGroupWithInvalidGroupMember() {
	// Try to create a group with an invalid group member ID
	invalidGroup := CreateGroupRequest{
		Name:               "Group with Invalid Group Member",
		OrganizationUnitId: testOU,
		Members: []Member{
			{
				Id:   "invalid-group-id-12345",
				Type: MemberTypeGroup,
			},
		},
	}

	jsonData, err := json.Marshal(invalidGroup)
	suite.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	// Verify the error response
	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var errorResp map[string]interface{}
	err = json.Unmarshal(body, &errorResp)
	suite.Require().NoError(err)

	// The error code might be different for invalid group IDs
	suite.Equal("GRP-1008", errorResp["code"])
	suite.Equal("Invalid group member ID", errorResp["message"])
}

func createGroup(ts *GroupAPITestSuite) (string, error) {
	jsonData, err := json.Marshal(groupToCreate)
	if err != nil {
		return "", fmt.Errorf("failed to marshal groupToCreate: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/groups", bytes.NewBuffer(jsonData))
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdGroup Group
	err = json.NewDecoder(resp.Body).Decode(&createdGroup)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	return createdGroup.Id, nil
}

func deleteGroup(groupID string) error {
	req, err := http.NewRequest("DELETE", testServerURL+"/groups/"+groupID, nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("expected status 204, got %d", resp.StatusCode)
	}
	return nil
}

// createTestUser creates a test user and returns the user ID
func createTestUser() (string, error) {
	testUser := map[string]interface{}{
		"organizationUnit": "456e8400-e29b-41d4-a716-446655440001",
		"type":             "user",
		"attributes": map[string]interface{}{
			"email":     "testuser@example.com",
			"firstName": "Test",
			"lastName":  "User",
			"password":  "TestPassword123!",
		},
	}

	jsonData, err := json.Marshal(testUser)
	if err != nil {
		return "", fmt.Errorf("failed to marshal test user: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("POST", testServerURL+"/users", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected status 201, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var createdUser map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdUser)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	userID, ok := createdUser["id"].(string)
	if !ok {
		return "", fmt.Errorf("failed to extract user ID from response")
	}

	return userID, nil
}

// deleteTestUser deletes a test user
func deleteTestUser(userID string) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("DELETE", testServerURL+"/users/"+userID, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("expected status 204, got %d", resp.StatusCode)
	}
	return nil
}

func buildCreatedGroup() Group {
	return Group{
		GroupBasic: GroupBasic{
			Id:                 createdGroupID,
			Name:               groupToCreate.Name,
			OrganizationUnitId: groupToCreate.OrganizationUnitId,
		},
		Members: groupToCreate.Members,
	}
}

func TestGroupAPITestSuite(t *testing.T) {
	suite.Run(t, new(GroupAPITestSuite))
}

func (suite *GroupAPITestSuite) TestGetGroupMembers() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for member retrieval")
	}

	// Get the group members
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups/"+createdGroupID+"/members", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var memberListResponse MemberListResponse
	err = json.Unmarshal(body, &memberListResponse)
	suite.Require().NoError(err)

	// Verify the response structure
	suite.GreaterOrEqual(memberListResponse.TotalResults, 1, "Should have at least one member")
	suite.Equal(1, memberListResponse.StartIndex, "StartIndex should be 1 for non-paginated request")
	suite.Equal(memberListResponse.TotalResults, memberListResponse.Count, "Count should equal TotalResults for non-paginated request")
	suite.Equal(len(memberListResponse.Members), memberListResponse.Count, "Members array length should match Count")

	// Verify we have the expected member
	found := false
	for _, member := range memberListResponse.Members {
		if member.Id == "550e8400-e29b-41d4-a716-446655440000" && member.Type == MemberTypeUser {
			found = true
			break
		}
	}
	suite.True(found, "Expected member should be in the list")
}

func (suite *GroupAPITestSuite) TestGetGroupMembersWithPagination() {
	if createdGroupID == "" {
		suite.T().Fatal("Group ID is not available for member retrieval")
	}

	// Test pagination with limit=1, offset=0
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups/"+createdGroupID+"/members?limit=1&offset=0", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var memberListResponse MemberListResponse
	err = json.Unmarshal(body, &memberListResponse)
	suite.Require().NoError(err)

	// Verify pagination structure
	suite.GreaterOrEqual(memberListResponse.TotalResults, 1, "Should have at least one member")
	suite.Equal(1, memberListResponse.StartIndex, "StartIndex should be 1 for offset=0")
	suite.LessOrEqual(memberListResponse.Count, 1, "Count should be at most 1 due to limit=1")
	suite.LessOrEqual(len(memberListResponse.Members), 1, "Should return at most 1 member")
}

func (suite *GroupAPITestSuite) TestGetGroupMembersNotFound() {
	// Try to get members of a non-existent group
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", testServerURL+"/groups/non-existent-id/members", nil)
	suite.Require().NoError(err)

	resp, err := client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusNotFound, resp.StatusCode)
}
