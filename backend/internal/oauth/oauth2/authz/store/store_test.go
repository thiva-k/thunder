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

package store

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/authz/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/tests/mocks/database/clientmock"
	"github.com/asgardeo/thunder/tests/mocks/database/modelmock"
	"github.com/asgardeo/thunder/tests/mocks/database/providermock"
)

type AuthorizationCodeStoreTestSuite struct {
	suite.Suite
	mockDBProvider *providermock.DBProviderInterfaceMock
	mockDBClient   *clientmock.DBClientInterfaceMock
	store          *AuthorizationCodeStore
	testAuthzCode  model.AuthorizationCode
}

func TestAuthorizationCodeStoreTestSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationCodeStoreTestSuite))
}

func (suite *AuthorizationCodeStoreTestSuite) SetupTest() {
	testConfig := &config.Config{
		Database: config.DatabaseConfig{
			Identity: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
			Runtime: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)

	suite.mockDBProvider = &providermock.DBProviderInterfaceMock{}
	suite.mockDBClient = &clientmock.DBClientInterfaceMock{}

	suite.store = &AuthorizationCodeStore{
		DBProvider: suite.mockDBProvider,
	}

	suite.testAuthzCode = model.AuthorizationCode{
		CodeID:              "test-code-id",
		Code:                "test-code",
		ClientID:            "test-client-id",
		RedirectURI:         "https://client.example.com/callback",
		AuthorizedUserID:    "test-user-id",
		TimeCreated:         time.Now(),
		ExpiryTime:          time.Now().Add(10 * time.Minute),
		Scopes:              "read write",
		State:               constants.AuthCodeStateActive,
		CodeChallenge:       "",
		CodeChallengeMethod: "",
	}
}

func (suite *AuthorizationCodeStoreTestSuite) TestNewAuthorizationCodeStore() {
	store := NewAuthorizationCodeStore()
	assert.NotNil(suite.T(), store)
	assert.Implements(suite.T(), (*AuthorizationCodeStoreInterface)(nil), store)
}

func (suite *AuthorizationCodeStoreTestSuite) TestInsertAuthorizationCode_Success() {
	mockTx := &modelmock.TxInterfaceMock{}

	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(mockTx, nil)

	mockTx.On("Exec", constants.QueryInsertAuthorizationCode.Query,
		suite.testAuthzCode.CodeID, suite.testAuthzCode.Code, suite.testAuthzCode.ClientID,
		suite.testAuthzCode.RedirectURI, suite.testAuthzCode.AuthorizedUserID,
		suite.testAuthzCode.TimeCreated, suite.testAuthzCode.ExpiryTime, suite.testAuthzCode.State,
		suite.testAuthzCode.CodeChallenge, suite.testAuthzCode.CodeChallengeMethod).
		Return(nil, nil)

	mockTx.On("Exec", constants.QueryInsertAuthorizationCodeScopes.Query,
		suite.testAuthzCode.CodeID, suite.testAuthzCode.Scopes).
		Return(nil, nil)

	mockTx.On("Commit").Return(nil)

	err := suite.store.InsertAuthorizationCode(suite.testAuthzCode)
	assert.NoError(suite.T(), err)

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
	mockTx.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestInsertAuthorizationCode_DBClientError() {
	suite.mockDBProvider.On("GetDBClient", "runtime").Return(nil, errors.New("db client error"))

	err := suite.store.InsertAuthorizationCode(suite.testAuthzCode)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "db client error")

	suite.mockDBProvider.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestInsertAuthorizationCode_BeginTxError() {
	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(nil, errors.New("tx error"))

	err := suite.store.InsertAuthorizationCode(suite.testAuthzCode)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to begin transaction")

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestInsertAuthorizationCode_ExecError() {
	mockTx := &modelmock.TxInterfaceMock{}

	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(mockTx, nil)

	mockTx.On("Exec", constants.QueryInsertAuthorizationCode.Query,
		suite.testAuthzCode.CodeID, suite.testAuthzCode.Code, suite.testAuthzCode.ClientID,
		suite.testAuthzCode.RedirectURI, suite.testAuthzCode.AuthorizedUserID,
		suite.testAuthzCode.TimeCreated, suite.testAuthzCode.ExpiryTime, suite.testAuthzCode.State,
		suite.testAuthzCode.CodeChallenge, suite.testAuthzCode.CodeChallengeMethod).
		Return(nil, errors.New("exec error"))

	mockTx.On("Rollback").Return(nil)

	err := suite.store.InsertAuthorizationCode(suite.testAuthzCode)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to insert authorization code")

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
	mockTx.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestInsertAuthorizationCode_ScopeExecError() {
	mockTx := &modelmock.TxInterfaceMock{}

	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(mockTx, nil)

	mockTx.On("Exec", constants.QueryInsertAuthorizationCode.Query,
		suite.testAuthzCode.CodeID, suite.testAuthzCode.Code, suite.testAuthzCode.ClientID,
		suite.testAuthzCode.RedirectURI, suite.testAuthzCode.AuthorizedUserID,
		suite.testAuthzCode.TimeCreated, suite.testAuthzCode.ExpiryTime, suite.testAuthzCode.State,
		suite.testAuthzCode.CodeChallenge, suite.testAuthzCode.CodeChallengeMethod).
		Return(nil, nil)

	mockTx.On("Exec", constants.QueryInsertAuthorizationCodeScopes.Query,
		suite.testAuthzCode.CodeID, suite.testAuthzCode.Scopes).
		Return(nil, errors.New("scope exec error"))

	mockTx.On("Rollback").Return(nil)

	err := suite.store.InsertAuthorizationCode(suite.testAuthzCode)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to insert authorization code scopes")

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
	mockTx.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestInsertAuthorizationCode_CommitError() {
	mockTx := &modelmock.TxInterfaceMock{}

	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("BeginTx").Return(mockTx, nil)

	mockTx.On("Exec", constants.QueryInsertAuthorizationCode.Query,
		suite.testAuthzCode.CodeID, suite.testAuthzCode.Code, suite.testAuthzCode.ClientID,
		suite.testAuthzCode.RedirectURI, suite.testAuthzCode.AuthorizedUserID,
		suite.testAuthzCode.TimeCreated, suite.testAuthzCode.ExpiryTime, suite.testAuthzCode.State,
		suite.testAuthzCode.CodeChallenge, suite.testAuthzCode.CodeChallengeMethod).
		Return(nil, nil)

	mockTx.On("Exec", constants.QueryInsertAuthorizationCodeScopes.Query,
		suite.testAuthzCode.CodeID, suite.testAuthzCode.Scopes).
		Return(nil, nil)

	mockTx.On("Commit").Return(errors.New("commit error"))

	err := suite.store.InsertAuthorizationCode(suite.testAuthzCode)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to commit transaction")

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
	mockTx.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestGetAuthorizationCode_Success() {
	testTimeStr := "2023-12-01 10:30:45.123456"
	testTime, _ := time.Parse("2006-01-02 15:04:05.999999999", testTimeStr)

	queryResults := []map[string]interface{}{
		{
			"code_id":            "test-code-id",
			"authorization_code": "test-code",
			"callback_url":       "https://client.example.com/callback",
			"authz_user":         "test-user-id",
			"time_created":       testTimeStr,
			"expiry_time":        testTimeStr,
			"state":              constants.AuthCodeStateActive,
		},
	}

	scopeResults := []map[string]interface{}{
		{"scope": "read write"},
	}

	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", constants.QueryGetAuthorizationCode, "test-client-id", "test-code").
		Return(queryResults, nil)
	suite.mockDBClient.On("Query", constants.QueryGetAuthorizationCodeScopes, "test-code-id").
		Return(scopeResults, nil)

	result, err := suite.store.GetAuthorizationCode("test-client-id", "test-code")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-code-id", result.CodeID)
	assert.Equal(suite.T(), "test-code", result.Code)
	assert.Equal(suite.T(), "test-client-id", result.ClientID)
	assert.Equal(suite.T(), "https://client.example.com/callback", result.RedirectURI)
	assert.Equal(suite.T(), "test-user-id", result.AuthorizedUserID)
	assert.Equal(suite.T(), testTime, result.TimeCreated)
	assert.Equal(suite.T(), testTime, result.ExpiryTime)
	assert.Equal(suite.T(), "read write", result.Scopes)
	assert.Equal(suite.T(), constants.AuthCodeStateActive, result.State)

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestGetAuthorizationCode_DBClientError() {
	suite.mockDBProvider.On("GetDBClient", "runtime").Return(nil, errors.New("db client error"))

	result, err := suite.store.GetAuthorizationCode("test-client-id", "test-code")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), model.AuthorizationCode{}, result)

	suite.mockDBProvider.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestGetAuthorizationCode_QueryError() {
	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", constants.QueryGetAuthorizationCode, "test-client-id", "test-code").
		Return(nil, errors.New("query error"))

	result, err := suite.store.GetAuthorizationCode("test-client-id", "test-code")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "error while retrieving authorization code")
	assert.Equal(suite.T(), model.AuthorizationCode{}, result)

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestGetAuthorizationCode_NoResults() {
	queryResults := []map[string]interface{}{}

	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", constants.QueryGetAuthorizationCode, "test-client-id", "test-code").
		Return(queryResults, nil)

	result, err := suite.store.GetAuthorizationCode("test-client-id", "test-code")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrAuthorizationCodeNotFound, err)
	assert.Equal(suite.T(), model.AuthorizationCode{}, result)

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestGetAuthorizationCode_EmptyCodeID() {
	queryResults := []map[string]interface{}{
		{
			"code_id": "",
		},
	}

	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Query", constants.QueryGetAuthorizationCode, "test-client-id", "test-code").
		Return(queryResults, nil)

	result, err := suite.store.GetAuthorizationCode("test-client-id", "test-code")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), constants.ErrAuthorizationCodeNotFound, err)
	assert.Equal(suite.T(), model.AuthorizationCode{}, result)

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestDeactivateAuthorizationCode_Success() {
	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", constants.QueryUpdateAuthorizationCodeState,
		constants.AuthCodeStateInactive, suite.testAuthzCode.CodeID).Return(int64(1), nil)

	err := suite.store.DeactivateAuthorizationCode(suite.testAuthzCode)
	assert.NoError(suite.T(), err)

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestRevokeAuthorizationCode_Success() {
	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", constants.QueryUpdateAuthorizationCodeState,
		constants.AuthCodeStateRevoked, suite.testAuthzCode.CodeID).Return(int64(1), nil)

	err := suite.store.RevokeAuthorizationCode(suite.testAuthzCode)
	assert.NoError(suite.T(), err)

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestExpireAuthorizationCode_Success() {
	suite.mockDBProvider.On("GetDBClient", "runtime").Return(suite.mockDBClient, nil)
	suite.mockDBClient.On("Execute", constants.QueryUpdateAuthorizationCodeState,
		constants.AuthCodeStateExpired, suite.testAuthzCode.CodeID).Return(int64(1), nil)

	err := suite.store.ExpireAuthorizationCode(suite.testAuthzCode)
	assert.NoError(suite.T(), err)

	suite.mockDBProvider.AssertExpectations(suite.T())
	suite.mockDBClient.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestUpdateAuthorizationCodeState_Error() {
	suite.mockDBProvider.On("GetDBClient", "runtime").Return(nil, errors.New("db client error"))

	err := suite.store.DeactivateAuthorizationCode(suite.testAuthzCode)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "db client error")

	suite.mockDBProvider.AssertExpectations(suite.T())
}

func (suite *AuthorizationCodeStoreTestSuite) TestParseTimeField_StringInput() {
	testTime := "2023-12-01 10:30:45.123456789 extra content"
	expectedTime, _ := time.Parse("2006-01-02 15:04:05.999999999", "2023-12-01 10:30:45.123456789")

	result, err := parseTimeField(testTime, "test_field", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTime, result)
}

func (suite *AuthorizationCodeStoreTestSuite) TestParseTimeField_TimeInput() {
	testTime := time.Now()

	result, err := parseTimeField(testTime, "test_field", nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testTime, result)
}

func (suite *AuthorizationCodeStoreTestSuite) TestTrimTimeString() {
	input := "2023-12-01 10:30:45.123456789 extra content here"
	expected := "2023-12-01 10:30:45.123456789"

	result := trimTimeString(input)
	assert.Equal(suite.T(), expected, result)
}

func (suite *AuthorizationCodeStoreTestSuite) TestTrimTimeString_ShortInput() {
	input := "2023-12-01"

	result := trimTimeString(input)
	assert.Equal(suite.T(), input, result)
}
