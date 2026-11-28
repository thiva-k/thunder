// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/tests/integration/testutils"
)

const (
	// cimdDomainTestClientID is on the application's domain, which no host serves, so its document is never
	// retrieved.
	cimdDomainTestClientID    = "https://cimd-domain.invalid/oauth/client.json"
	cimdDomainTestRedirectURI = "https://cimd-domain.invalid/callback"
	cimdDomainOtherClientID   = "https://cimd-other.invalid/oauth/client.json"
)

// CIMDDomainTestSuite verifies that an unregistered client publishing a Client ID Metadata Document is refused
// as an unknown client unless the Client ID domains of an application admit it while unregistered clients are
// on, and its document can be retrieved. The suite's one application admits the clients on cimd-domain.invalid.
type CIMDDomainTestSuite struct {
	suite.Suite
	appID string
}

func TestCIMDDomainTestSuite(t *testing.T) {
	suite.Run(t, new(CIMDDomainTestSuite))
}

func (ts *CIMDDomainTestSuite) SetupSuite() {
	body, err := json.Marshal(map[string]any{"name": "CIMD domain test", "type": "mcp",
		"inboundAuthConfig": []map[string]any{{"type": "oauth2", "config": map[string]any{
			"clientIdMetadataDocument": true,
			"clientIdDomains":          []string{"cimd-domain.invalid"},
			"grantTypes":               []string{"authorization_code", "refresh_token"},
			"responseTypes":            []string{"code"},
		}}}})
	ts.Require().NoError(err)
	req, err := http.NewRequest(http.MethodPost, testutils.TestServerURL+"/applications", bytes.NewReader(body))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := testutils.GetHTTPClient().Do(req)
	ts.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusCreated, resp.StatusCode, string(respBody))
	var created struct {
		ID string `json:"id"`
	}
	ts.Require().NoError(json.Unmarshal(respBody, &created))
	ts.appID = created.ID
}

func (ts *CIMDDomainTestSuite) TearDownSuite() {
	req, err := http.NewRequest(http.MethodDelete, testutils.TestServerURL+"/applications/"+ts.appID, nil)
	ts.Require().NoError(err)
	resp, err := testutils.GetHTTPClient().Do(req)
	ts.Require().NoError(err)
	_ = resp.Body.Close()
}

func (ts *CIMDDomainTestSuite) TearDownTest() {
	ts.putSection(`{}`)
}

func (ts *CIMDDomainTestSuite) TestRefusedWhenNoDomainMatches() {
	ts.putSection(ts.section(true, `[]`))

	ts.assertUnknownClient(cimdDomainOtherClientID)
}

func (ts *CIMDDomainTestSuite) TestRefusedWhileUnregisteredClientsAreOff() {
	ts.putSection(ts.section(false, `[]`))

	ts.assertUnknownClient(cimdDomainTestClientID)
}

func (ts *CIMDDomainTestSuite) TestRefusedWhenBlocked() {
	ts.putSection(ts.section(true, `["`+cimdDomainTestClientID+`"]`))

	ts.assertUnknownClient(cimdDomainTestClientID)
}

func (ts *CIMDDomainTestSuite) TestRefusedWhenTheDocumentIsUnavailable() {
	ts.putSection(ts.section(true, `[]`))

	ts.assertUnknownClient(cimdDomainTestClientID)
}

// section returns a cimd section that turns unregistered clients on or off and blocks the given client IDs.
func (ts *CIMDDomainTestSuite) section(enabled bool, blocked string) string {
	enabledValue := "false"
	if enabled {
		enabledValue = "true"
	}
	return `{"enabled":` + enabledValue + `,"blockedClientIds":` + blocked + `}`
}

func (ts *CIMDDomainTestSuite) putSection(body string) {
	ts.Require().NoError(testutils.PutWritableServerConfig("cimd", []byte(body)))
}

func (ts *CIMDDomainTestSuite) assertUnknownClient(clientID string) {
	resp, err := testutils.InitiateAuthorizationFlow(clientID, cimdDomainTestRedirectURI, "code", "openid", "s")
	ts.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	ts.Equal(http.StatusFound, resp.StatusCode)
	ts.NoError(testutils.ValidateOAuth2ErrorRedirect(resp.Header.Get("Location"), "invalid_request",
		"Invalid client_id"))
}
