// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package serverconfig

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/tests/integration/testutils"
)

const cimdBlockedClientID = "https://claude.ai/oauth/blocked.json"

var cimdConfigURL = serverConfigURL + "/cimd"

// cimdSectionValue mirrors the cimd section value.
type cimdSectionValue struct {
	Enabled          bool     `json:"enabled"`
	BlockedClientIDs []string `json:"blockedClientIds"`
}

// cimdLayers mirrors the layered GET/PUT response of the cimd section.
type cimdLayers struct {
	ReadOnly cimdSectionValue `json:"readOnly"`
	Writable cimdSectionValue `json:"writable"`
	Merged   cimdSectionValue `json:"merged"`
}

type CIMDConfigAPITestSuite struct {
	suite.Suite
	adminClient *http.Client
}

func TestCIMDConfigAPITestSuite(t *testing.T) {
	suite.Run(t, new(CIMDConfigAPITestSuite))
}

func (suite *CIMDConfigAPITestSuite) SetupSuite() {
	suite.adminClient = testutils.GetHTTPClient()
}

func (suite *CIMDConfigAPITestSuite) SetupTest()     { suite.clear() }
func (suite *CIMDConfigAPITestSuite) TearDownSuite() { suite.clear() }

func (suite *CIMDConfigAPITestSuite) TestListIncludesSection() {
	status, body := suite.get(serverConfigURL)
	suite.Require().Equal(http.StatusOK, status)

	var names []string
	suite.Require().NoError(json.Unmarshal(body, &names))
	suite.Contains(names, "cimd")
}

func (suite *CIMDConfigAPITestSuite) TestEmptySectionReadsAsEmptyLists() {
	layers := suite.getLayers()

	suite.False(layers.Merged.Enabled)
	suite.NotNil(layers.Merged.BlockedClientIDs)
	suite.Empty(layers.Merged.BlockedClientIDs)
}

func (suite *CIMDConfigAPITestSuite) TestPutPersistsAndReadsBack() {
	status, body := suite.put(`{"enabled":true,"blockedClientIds":["` + cimdBlockedClientID + `"]}`)
	suite.Require().Equal(http.StatusOK, status, string(body))

	layers := suite.getLayers()
	suite.True(layers.Merged.Enabled)
	suite.Equal([]string{cimdBlockedClientID}, layers.Merged.BlockedClientIDs)
}

func (suite *CIMDConfigAPITestSuite) TestPutInvalidValueReturns400AndDoesNotPersist() {
	cases := map[string]string{
		"unknown field":               `{"enabled":true,"unknown":true}`,
		"policies are a resource":     `{"enabled":true,"policies":[]}`,
		"blocked client ID not a URL": `{"blockedClientIds":["claude.ai"]}`,
		"blocked client ID twice": `{"blockedClientIds":["` + cimdBlockedClientID + `","` +
			cimdBlockedClientID + `"]}`,
	}
	for name, body := range cases {
		suite.Run(name, func() {
			status, respBody := suite.put(body)
			suite.Equal(http.StatusBadRequest, status)
			suite.Equal("SCF-1003", suite.errorCode(respBody))
			suite.Empty(suite.getLayers().Writable.BlockedClientIDs)
		})
	}
}

func (suite *CIMDConfigAPITestSuite) clear() {
	status, body := suite.put(`{}`)
	suite.Require().Equal(http.StatusOK, status, string(body))
}

// put writes the section the way the plane under test takes server configuration.
func (suite *CIMDConfigAPITestSuite) put(body string) (int, []byte) {
	status, respBody, err := testutils.PutWritableServerConfigWithStatus("cimd", []byte(body))
	suite.Require().NoError(err)
	return status, respBody
}

func (suite *CIMDConfigAPITestSuite) get(url string) (int, []byte) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	suite.Require().NoError(err)

	resp, err := suite.adminClient.Do(req)
	suite.Require().NoError(err)
	defer closeBodyQuietly(suite.T(), resp.Body)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	return resp.StatusCode, body
}

func (suite *CIMDConfigAPITestSuite) getLayers() cimdLayers {
	status, body := suite.get(cimdConfigURL)
	suite.Require().Equal(http.StatusOK, status)
	var layers cimdLayers
	suite.Require().NoError(json.Unmarshal(body, &layers))
	return layers
}

func (suite *CIMDConfigAPITestSuite) errorCode(body []byte) string {
	var errResp apiErrorResponse
	suite.Require().NoError(json.Unmarshal(body, &errResp))
	return errResp.Code
}
