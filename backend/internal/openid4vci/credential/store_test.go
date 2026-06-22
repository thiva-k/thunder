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

package credential

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/tests/mocks/database/providermock"
)

const testDeploymentID = "test-deployment-id"

type ConfigurationStoreTestSuite struct {
	suite.Suite
	mockDBProvider *providermock.DBProviderInterfaceMock
	mockDBClient   *providermock.DBClientInterfaceMock
	store          *configurationStore
}

func TestConfigurationStoreTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationStoreTestSuite))
}

func (suite *ConfigurationStoreTestSuite) SetupTest() {
	suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
	suite.store = &configurationStore{
		dbProvider:   suite.mockDBProvider,
		deploymentID: testDeploymentID,
	}
}

func (suite *ConfigurationStoreTestSuite) refreshMocks() {
	suite.mockDBProvider = providermock.NewDBProviderInterfaceMock(suite.T())
	suite.mockDBClient = providermock.NewDBClientInterfaceMock(suite.T())
	suite.store = &configurationStore{
		dbProvider:   suite.mockDBProvider,
		deploymentID: testDeploymentID,
	}
}

// rowFromConfig simulates a persisted row by marshaling a DTO the way the store
// does, returning the column map buildConfigurationDTOFromRow consumes.
func rowFromConfig(t *testing.T, dto CredentialConfigurationDTO) map[string]interface{} {
	t.Helper()
	claims, display, err := marshalConfiguration(dto)
	require.NoError(t, err)
	row := map[string]interface{}{
		"id":               dto.ID,
		"handle":           dto.Handle,
		"ou_id":            dto.OUID,
		"format":           dto.Format,
		"vct":              dto.VCT,
		"claims":           claims,
		"display":          display,
		"validity_seconds": nil,
	}
	if dto.ValiditySeconds != nil {
		row["validity_seconds"] = int64(*dto.ValiditySeconds)
	}
	return row
}

func (suite *ConfigurationStoreTestSuite) TestRoundTripClaimsAndDisplay() {
	validity := 3600
	logoURI := "https://example.com/logo.png"
	original := CredentialConfigurationDTO{
		ID:     "cfg-1",
		Handle: "eudi-pid",
		Format: DefaultCredentialFormat,
		VCT:    "urn:eudi:pid:de:1",
		Claims: []ClaimMapping{
			{Name: "given_name", DisplayName: "Given Name"},
			{Name: "family_name", DisplayName: "Family Name"},
		},
		Display: &CredentialDisplay{
			Name:    "EUDI PID",
			Locale:  "en-US",
			LogoURI: logoURI,
		},
		ValiditySeconds: &validity,
	}

	got, err := buildConfigurationDTOFromRow(rowFromConfig(suite.T(), original))
	suite.Require().NoError(err)
	suite.Equal(original.ID, got.ID)
	suite.Equal(original.Claims, got.Claims)
	suite.Require().NotNil(got.Display)
	suite.Equal("EUDI PID", got.Display.Name)
	suite.Equal(logoURI, got.Display.LogoURI)
	suite.Require().NotNil(got.ValiditySeconds)
	suite.Equal(validity, *got.ValiditySeconds)
}

func (suite *ConfigurationStoreTestSuite) TestRoundTripNullFields() {
	got, err := buildConfigurationDTOFromRow(rowFromConfig(suite.T(), CredentialConfigurationDTO{
		ID: "cfg-2", Handle: "h", VCT: "v", Format: DefaultCredentialFormat,
	}))
	suite.Require().NoError(err)
	suite.Empty(got.Claims)
	suite.Nil(got.Display)
	suite.Nil(got.ValiditySeconds)
}

func (suite *ConfigurationStoreTestSuite) TestColumnNullableInt() {
	i := columnNullableInt(int64(3600))
	suite.Require().NotNil(i)
	suite.Equal(3600, *i)

	suite.Nil(columnNullableInt(nil))
}

func (suite *ConfigurationStoreTestSuite) TestCreate() {
	testCases := []struct {
		name       string
		dto        CredentialConfigurationDTO
		setupMocks func()
		shouldErr  bool
	}{
		{
			name: "Success",
			dto: CredentialConfigurationDTO{
				ID: "cfg-1", Handle: "eudi-pid", OUID: "ou-1",
				Format: DefaultCredentialFormat, VCT: "urn:eudi:pid:de:1",
				Claims: []ClaimMapping{{Name: "given_name", DisplayName: "Given Name"}},
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateConfiguration,
					"cfg-1", "eudi-pid", "ou-1", DefaultCredentialFormat, "urn:eudi:pid:de:1",
					mock.Anything, nil, nil, testDeploymentID,
				).Return(int64(1), nil)
			},
		},
		{
			name: "ExecuteError",
			dto:  CredentialConfigurationDTO{ID: "cfg-1", Handle: "h", VCT: "v"},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryCreateConfiguration,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(int64(0), errors.New("insert failed"))
			},
			shouldErr: true,
		},
		{
			name: "DBClientError",
			dto:  CredentialConfigurationDTO{ID: "cfg-1", Handle: "h", VCT: "v"},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, errors.New("db error"))
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.refreshMocks()
			tc.setupMocks()
			err := suite.store.Create(context.Background(), tc.dto)
			if tc.shouldErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *ConfigurationStoreTestSuite) TestGetByID() {
	testCases := []struct {
		name        string
		id          string
		setupMocks  func()
		checkResult func(*CredentialConfigurationDTO)
		shouldErr   bool
		errIs       error
	}{
		{
			name: "Success",
			id:   "cfg-1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				validity := 3600
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetConfigurationByID,
					"cfg-1", testDeploymentID,
				).Return([]map[string]interface{}{
					rowFromConfig(suite.T(), CredentialConfigurationDTO{
						ID: "cfg-1", Handle: "eudi-pid", OUID: "ou-1",
						Format: DefaultCredentialFormat, VCT: "urn:eudi:pid:de:1",
						Claims:          []ClaimMapping{{Name: "given_name", DisplayName: "Given Name"}},
						ValiditySeconds: &validity,
					}),
				}, nil)
			},
			checkResult: func(got *CredentialConfigurationDTO) {
				suite.Equal("cfg-1", got.ID)
				suite.Equal("eudi-pid", got.Handle)
				suite.Require().Len(got.Claims, 1)
				suite.Equal("given_name", got.Claims[0].Name)
				suite.Require().NotNil(got.ValiditySeconds)
				suite.Equal(3600, *got.ValiditySeconds)
			},
		},
		{
			name: "NotFound",
			id:   "missing",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetConfigurationByID,
					"missing", testDeploymentID,
				).Return([]map[string]interface{}{}, nil)
			},
			shouldErr: true,
			errIs:     ErrNotFound,
		},
		{
			name: "QueryError",
			id:   "cfg-1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetConfigurationByID,
					"cfg-1", testDeploymentID,
				).Return(nil, errors.New("query failed"))
			},
			shouldErr: true,
		},
		{
			name: "DBClientError",
			id:   "cfg-1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, errors.New("db error"))
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.refreshMocks()
			tc.setupMocks()
			got, err := suite.store.GetByID(context.Background(), tc.id)
			if tc.shouldErr {
				suite.Error(err)
				suite.Nil(got)
				if tc.errIs != nil {
					suite.ErrorIs(err, tc.errIs)
				}
			} else {
				suite.NoError(err)
				suite.Require().NotNil(got)
				if tc.checkResult != nil {
					tc.checkResult(got)
				}
			}
		})
	}
}

func (suite *ConfigurationStoreTestSuite) TestGetByHandle() {
	testCases := []struct {
		name       string
		handle     string
		setupMocks func()
		shouldErr  bool
		errIs      error
	}{
		{
			name:   "Success",
			handle: "eudi-pid",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetConfigurationByHandle,
					"eudi-pid", testDeploymentID,
				).Return([]map[string]interface{}{
					rowFromConfig(suite.T(), CredentialConfigurationDTO{
						ID: "cfg-1", Handle: "eudi-pid", Format: DefaultCredentialFormat, VCT: "urn:eudi:pid:de:1",
					}),
				}, nil)
			},
		},
		{
			name:   "NotFound",
			handle: "missing",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryGetConfigurationByHandle,
					"missing", testDeploymentID,
				).Return([]map[string]interface{}{}, nil)
			},
			shouldErr: true,
			errIs:     ErrNotFound,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.refreshMocks()
			tc.setupMocks()
			got, err := suite.store.GetByHandle(context.Background(), tc.handle)
			if tc.shouldErr {
				suite.Error(err)
				suite.Nil(got)
				if tc.errIs != nil {
					suite.ErrorIs(err, tc.errIs)
				}
			} else {
				suite.NoError(err)
				suite.NotNil(got)
			}
		})
	}
}

func (suite *ConfigurationStoreTestSuite) TestList() {
	testCases := []struct {
		name          string
		setupMocks    func()
		expectedCount int
		shouldErr     bool
	}{
		{
			name: "ReturnsAllRows",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryListConfigurations, testDeploymentID).
					Return([]map[string]interface{}{
						rowFromConfig(suite.T(), CredentialConfigurationDTO{
							ID: "cfg-1", Handle: "h1", Format: DefaultCredentialFormat, VCT: "v1",
						}),
						rowFromConfig(suite.T(), CredentialConfigurationDTO{
							ID: "cfg-2", Handle: "h2", Format: DefaultCredentialFormat, VCT: "v2",
						}),
					}, nil)
			},
			expectedCount: 2,
		},
		{
			name: "Empty",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryListConfigurations, testDeploymentID).
					Return([]map[string]interface{}{}, nil)
			},
			expectedCount: 0,
		},
		{
			name: "QueryError",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("QueryContext", mock.Anything, queryListConfigurations, testDeploymentID).
					Return(nil, errors.New("query failed"))
			},
			shouldErr: true,
		},
		{
			name: "DBClientError",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, errors.New("db error"))
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.refreshMocks()
			tc.setupMocks()
			cfgs, err := suite.store.List(context.Background())
			if tc.shouldErr {
				suite.Error(err)
				suite.Nil(cfgs)
			} else {
				suite.NoError(err)
				suite.Len(cfgs, tc.expectedCount)
			}
		})
	}
}

func (suite *ConfigurationStoreTestSuite) TestUpdate() {
	testCases := []struct {
		name       string
		dto        CredentialConfigurationDTO
		setupMocks func()
		shouldErr  bool
	}{
		{
			name: "Success",
			dto: CredentialConfigurationDTO{
				ID: "cfg-1", Handle: "new-handle", OUID: "ou-1",
				Format: DefaultCredentialFormat, VCT: "urn:eudi:pid:de:1",
			},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryUpdateConfiguration,
					"cfg-1", "new-handle", "ou-1", DefaultCredentialFormat, "urn:eudi:pid:de:1",
					nil, nil, nil, testDeploymentID,
				).Return(int64(1), nil)
			},
		},
		{
			name: "ExecuteError",
			dto:  CredentialConfigurationDTO{ID: "cfg-1", Handle: "h", VCT: "v"},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryUpdateConfiguration,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(int64(0), errors.New("update failed"))
			},
			shouldErr: true,
		},
		{
			name: "DBClientError",
			dto:  CredentialConfigurationDTO{ID: "cfg-1", Handle: "h", VCT: "v"},
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, errors.New("db error"))
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.refreshMocks()
			tc.setupMocks()
			err := suite.store.Update(context.Background(), tc.dto)
			if tc.shouldErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *ConfigurationStoreTestSuite) TestDelete() {
	testCases := []struct {
		name       string
		id         string
		setupMocks func()
		shouldErr  bool
	}{
		{
			name: "Success",
			id:   "cfg-1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteConfiguration,
					"cfg-1", testDeploymentID,
				).Return(int64(1), nil)
			},
		},
		{
			name: "ExecuteError",
			id:   "cfg-1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(suite.mockDBClient, nil)
				suite.mockDBClient.On("ExecuteContext", mock.Anything, queryDeleteConfiguration,
					"cfg-1", testDeploymentID,
				).Return(int64(0), errors.New("delete failed"))
			},
			shouldErr: true,
		},
		{
			name: "DBClientError",
			id:   "cfg-1",
			setupMocks: func() {
				suite.mockDBProvider.On("GetConfigDBClient").Return(nil, errors.New("db error"))
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.refreshMocks()
			tc.setupMocks()
			err := suite.store.Delete(context.Background(), tc.id)
			if tc.shouldErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *ConfigurationStoreTestSuite) TestIsDeclarative() {
	isDeclarative, err := suite.store.IsDeclarative(context.Background(), "any-id")
	suite.NoError(err)
	suite.False(isDeclarative)
}
