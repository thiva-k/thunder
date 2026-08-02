// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package ciba

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/thunder-id/thunderid/tests/integration/testutils"
)

const (
	// cibaGrantType is the OpenID Connect CIBA grant type identifier (providers.GrantTypeCIBA).
	cibaGrantType = "urn:openid:params:grant-type:ciba"
	// cibaBackchannelEndpoint is the backchannel authentication endpoint
	// (oauth2const.OAuth2BackchannelAuthEndpoint).
	cibaBackchannelEndpoint = "/oauth2/bc-authorize"
	// cibaCallbackEndpoint is the shared flow-callback endpoint.
	cibaCallbackEndpoint = "/oauth2/auth/callback"
	// cibaTokenEndpoint is the token endpoint used for polling.
	cibaTokenEndpoint = "/oauth2/token" // #nosec G101
	// cibaPollIntervalSeconds mirrors oauth2const.CIBADefaultIntervalSeconds (the minimum interval
	// between token polls while a request is pending).
	cibaPollIntervalSeconds = 5
	// cibaMockNotificationServerPort is the port for this suite's mock notification server. It must
	// not collide with other integration suites (the SMS auth suite uses 8098).
	cibaMockNotificationServerPort = 8099

	cibaClientID     = "ciba_test_client_123"
	cibaClientSecret = "ciba_test_secret_123"
	cibaTestUsername = "ciba_test_user"
	cibaTestPassword = "cibapass123"
)

type CIBATestSuite struct {
	suite.Suite
	ouID   string
	client *http.Client
}

func TestCIBATestSuite(t *testing.T) {
	suite.Run(t, new(CIBATestSuite))
}

func (ts *CIBATestSuite) SetupSuite() {
	ts.client = testutils.GetHTTPClient()

	ouID, err := testutils.CreateOrganizationUnit(testutils.OrganizationUnit{
		Handle:      "ciba-test-ou",
		Name:        "CIBA Test OU",
		Description: "Organization unit for CIBA integration tests",
		Parent:      nil,
	})
	ts.Require().NoError(err, "Failed to create test organization unit")
	ts.ouID = ouID
}

func (ts *CIBATestSuite) TearDownSuite() {
	if ts.ouID != "" {
		if err := testutils.DeleteOrganizationUnit(ts.ouID); err != nil {
			ts.T().Logf("Failed to delete test organization unit: %v", err)
		}
	}
}

// TestCIBAGrantFlow exercises the full Client-Initiated Backchannel Authentication (CIBA) grant
// end to end: it initiates a backchannel request, recovers the server-initiated flow's executionId
// from an out-of-band notification, completes the authentication flow, drives the state machine
// (PENDING -> AUTHENTICATED -> CONSUMED) through the callback and token endpoints, and asserts the
// one-time-use enforcement backed by the runtime store's CompareFieldAndSwap primitive.
func (ts *CIBATestSuite) TestCIBAGrantFlow() {
	// Mock notification server captures the CIBA notification (which carries the invite link with
	// the executionId). It is a plain HTTP server; the sender below is pointed at it via the API.
	mockServer := testutils.NewMockNotificationServer(cibaMockNotificationServerPort)
	ts.Require().NoError(mockServer.Start(), "Failed to start mock notification server")
	defer func() { _ = mockServer.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// A custom notification sender that POSTs rendered messages to the mock server. This is a DB
	// resource created via the API, so no server restart is required.
	senderID, err := testutils.CreateNotificationSender(testutils.NotificationSender{
		Name:        "CIBA Test Sender",
		Description: "Sender for CIBA integration test",
		Provider:    "custom",
		Properties: []testutils.SenderProperty{
			{Name: "url", Value: mockServer.GetSendSMSURL()},
			{Name: "http_method", Value: "POST"},
			{Name: "content_type", Value: "JSON"},
		},
	})
	ts.Require().NoError(err, "Failed to create notification sender")
	defer func() { _ = testutils.DeleteNotificationSender(senderID) }()

	// User type + user. mobile_number is the recipient the SMS executor resolves from the
	// identified user; username/password back the credential confirmation step.
	userTypeID, err := testutils.CreateUserType(testutils.UserType{
		Name: "ciba-test-person",
		OUID: ts.ouID,
		Schema: map[string]interface{}{
			"username":      map[string]interface{}{"type": "string"},
			"password":      map[string]interface{}{"type": "string", "credential": true},
			"email":         map[string]interface{}{"type": "string"},
			"mobile_number": map[string]interface{}{"type": "string"},
		},
	})
	ts.Require().NoError(err, "Failed to create CIBA test user type")
	defer func() { _ = testutils.DeleteUserType(userTypeID) }()

	userID, err := testutils.CreateUser(testutils.User{
		OUID: ts.ouID,
		Type: "ciba-test-person",
		Attributes: json.RawMessage(`{
			"username": "` + cibaTestUsername + `",
			"password": "` + cibaTestPassword + `",
			"email": "ciba_test_user@example.com",
			"mobile_number": "+1987654321"
		}`),
	})
	ts.Require().NoError(err, "Failed to create CIBA test user")
	defer func() { _ = testutils.DeleteUser(userID) }()

	// CIBA authentication flow. bc-authorize runs this server-side with login_hint. The
	// IdentifyingExecutor resolves the user (login_hint -> username), the InviteExecutor mints a
	// link carrying executionId + auth_req_id, the SMSExecutor delivers it via the mock server, and
	// the flow then pauses at the notification-sent prompt. The resumed flow re-enters via the
	// invite-verify node, which skips challenge validation so it can be resumed cold using only the
	// executionId + inviteToken recovered from the notification.
	flowID, err := testutils.CreateFlow(testutils.Flow{
		Name:     "CIBA Test Auth Flow",
		FlowType: "AUTHENTICATION",
		Handle:   "auth_flow_ciba_test",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "identify_user",
			},
			{
				"id":   "identify_user",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "IdentifyingExecutor",
					"mode": "identify",
				},
				"properties": map[string]interface{}{
					"loginHintAttribute": "username",
				},
				"onSuccess": "generate_invite",
			},
			{
				"id":   "generate_invite",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "InviteExecutor",
					"mode": "generate",
				},
				"onSuccess": "send_ciba_notification",
			},
			{
				"id":   "send_ciba_notification",
				"type": "TASK_EXECUTION",
				"properties": map[string]interface{}{
					"senderId":    senderID,
					"smsTemplate": "CIBA_NOTIFICATION",
				},
				"executor": map[string]interface{}{
					"name": "SMSExecutor",
				},
				"onSuccess": "notification_sent",
			},
			{
				// The server-initiated segment pauses here after the notification is sent; the
				// resumed flow re-enters via the invite-verify node below, which skips challenge
				// validation so it can be resumed cold using only the executionId + inviteToken.
				"id":   "notification_sent",
				"type": "PROMPT",
				"next": "verify_invite",
			},
			{
				"id":   "verify_invite",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "InviteExecutor",
					"mode": "verify",
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_invite_token",
							"identifier": "inviteToken",
							"type":       "HIDDEN",
							"required":   true,
						},
					},
				},
				"onSuccess": "prompt_credentials",
			},
			{
				"id":   "prompt_credentials",
				"type": "PROMPT",
				"prompts": []map[string]interface{}{
					{
						"inputs": []map[string]interface{}{
							{
								"ref":        "input_001",
								"identifier": "username",
								"type":       "TEXT_INPUT",
								"required":   true,
							},
							{
								"ref":        "input_002",
								"identifier": "password",
								"type":       "PASSWORD_INPUT",
								"required":   true,
							},
						},
						"action": map[string]interface{}{
							"ref":      "action_001",
							"nextNode": "credentials_auth",
						},
					},
				},
			},
			{
				"id":   "credentials_auth",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "CredentialsAuthExecutor",
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_001",
							"identifier": "username",
							"type":       "TEXT_INPUT",
							"required":   true,
						},
						{
							"ref":        "input_002",
							"identifier": "password",
							"type":       "PASSWORD_INPUT",
							"required":   true,
						},
					},
				},
				"onSuccess": "auth_assert",
			},
			{
				"id":   "auth_assert",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "AuthAssertExecutor",
				},
				"onSuccess": "end",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	})
	ts.Require().NoError(err, "Failed to create CIBA auth flow")
	defer func() { _ = testutils.DeleteFlow(flowID) }()

	appID := ts.createCIBATestApplication(flowID)
	defer func() { _ = testutils.DeleteApplication(appID) }()

	// Step 1: Backchannel authorization request.
	status, bcResp := ts.cibaBackchannelAuthorize(cibaTestUsername, "openid")
	ts.Require().Equal(http.StatusOK, status, "bc-authorize should succeed")
	ts.Require().NotEmpty(bcResp.AuthReqID, "bc-authorize response should carry auth_req_id")
	ts.Require().Equal(int64(cibaPollIntervalSeconds), bcResp.Interval, "interval should be the default")

	// Step 2: While pending, the token endpoint enforces the polling interval — the first poll is
	// authorization_pending, an immediate re-poll is slow_down. Neither consumes the request.
	pending := ts.cibaPollToken(bcResp.AuthReqID)
	ts.Require().Equal(http.StatusBadRequest, pending.statusCode)
	ts.Require().Equal("authorization_pending", pending.errorCode)

	slowDown := ts.cibaPollToken(bcResp.AuthReqID)
	ts.Require().Equal(http.StatusBadRequest, slowDown.statusCode)
	ts.Require().Equal("slow_down", slowDown.errorCode)

	// Step 3: Recover the executionId, auth_req_id, and inviteToken from the captured notification.
	var executionID, notifiedAuthReqID, inviteToken string
	ts.Require().Eventually(func() bool {
		msg := mockServer.GetLastMessage()
		if msg == nil {
			return false
		}
		executionID = extractCIBALinkParam(msg.Message, "executionId")
		notifiedAuthReqID = extractCIBALinkParam(msg.Message, "auth_req_id")
		inviteToken = extractCIBALinkParam(msg.Message, "inviteToken")
		return executionID != "" && inviteToken != ""
	}, 5*time.Second, 100*time.Millisecond, "Expected CIBA notification carrying the executionId")
	ts.Require().Equal(bcResp.AuthReqID, notifiedAuthReqID,
		"notification auth_req_id should match the bc-authorize response")

	// Step 4: Resume the paused flow cold via the invite-verify node (submitting the inviteToken,
	// which skips challenge validation), then approve by submitting the user's credentials.
	resumeStep, err := testutils.ExecuteAuthenticationFlow(executionID,
		map[string]string{"inviteToken": inviteToken}, "")
	ts.Require().NoError(err, "should resume the flow with the invite token")
	flowStep, err := testutils.ExecuteAuthenticationFlow(executionID, map[string]string{
		"username": cibaTestUsername,
		"password": cibaTestPassword,
	}, "action_001", resumeStep.ChallengeToken)
	ts.Require().NoError(err, "should complete the authentication flow")
	ts.Require().Equal("COMPLETE", flowStep.FlowStatus)
	ts.Require().NotEmpty(flowStep.Assertion, "flow completion should yield an assertion")

	// Step 5: Post the assertion to the callback to drive MarkAuthenticated (PENDING -> AUTHENTICATED).
	ts.Require().Equal(http.StatusOK, ts.cibaPostCallback(bcResp.AuthReqID, flowStep.Assertion),
		"CIBA callback should accept the assertion")

	// Step 6: Poll the token endpoint to drive MarkConsumed and issue tokens. The pending polls
	// recently stamped LastPolledAt; once AUTHENTICATED the handler skips the interval check, but
	// retry once on slow_down for robustness against timing.
	tokenRes := ts.cibaPollToken(bcResp.AuthReqID)
	if tokenRes.statusCode == http.StatusBadRequest && tokenRes.errorCode == "slow_down" {
		time.Sleep(cibaPollIntervalSeconds * time.Second)
		tokenRes = ts.cibaPollToken(bcResp.AuthReqID)
	}
	ts.Require().Equal(http.StatusOK, tokenRes.statusCode, "AUTHENTICATED request should issue tokens")
	ts.Require().NotEmpty(tokenRes.accessToken, "response should carry an access_token")

	claims, err := testutils.DecodeJWT(tokenRes.accessToken)
	ts.Require().NoError(err, "issued access token should be a decodable JWT")
	ts.Require().Equal(userID, claims.Sub, "token subject should be the CIBA user")

	// Step 7: A second poll is rejected — the request is CONSUMED (one-time use).
	reuse := ts.cibaPollToken(bcResp.AuthReqID)
	ts.Require().Equal(http.StatusBadRequest, reuse.statusCode)
	ts.Require().Equal("invalid_grant", reuse.errorCode, "a consumed request must not issue tokens again")
}

// createCIBATestApplication creates an OAuth application that allows the CIBA grant and is bound to
// the given authentication flow, returning its application ID.
func (ts *CIBATestSuite) createCIBATestApplication(authFlowID string) string {
	app := map[string]interface{}{
		"name":                      "CIBATestApp",
		"description":               "Application for CIBA integration test",
		"ouId":                      ts.ouID,
		"type":                      "fullstack",
		"authFlowId":                authFlowID,
		"isRegistrationFlowEnabled": false,
		"allowedUserTypes":          []string{"ciba-test-person"},
		"inboundAuthConfig": []map[string]interface{}{
			{
				"type": "oauth2",
				"config": map[string]interface{}{
					"clientId":                cibaClientID,
					"clientSecret":            cibaClientSecret,
					"redirectUris":            []string{"https://localhost:3000"},
					"grantTypes":              []string{cibaGrantType, "refresh_token"},
					"tokenEndpointAuthMethod": "client_secret_basic",
				},
			},
		},
	}

	jsonData, err := json.Marshal(app)
	ts.Require().NoError(err)

	req, err := http.NewRequest("POST", testutils.TestServerURL+"/applications", bytes.NewBuffer(jsonData))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		ts.T().Fatalf("Failed to create CIBA application. Status: %d, Response: %s",
			resp.StatusCode, string(bodyBytes))
	}

	var respData map[string]interface{}
	ts.Require().NoError(json.NewDecoder(resp.Body).Decode(&respData))
	return respData["id"].(string)
}

// cibaBackchannelResponse is the JSON body of a successful bc-authorize response.
type cibaBackchannelResponse struct {
	AuthReqID string `json:"auth_req_id"`
	ExpiresIn int64  `json:"expires_in"`
	Interval  int64  `json:"interval"`
}

// cibaBackchannelAuthorize submits a POST /oauth2/bc-authorize request with client_secret_basic
// authentication and returns the HTTP status and parsed response.
func (ts *CIBATestSuite) cibaBackchannelAuthorize(loginHint, scope string) (int, cibaBackchannelResponse) {
	form := url.Values{}
	form.Set("login_hint", loginHint)
	form.Set("scope", scope)

	req, err := http.NewRequest("POST", testutils.TestServerURL+cibaBackchannelEndpoint,
		strings.NewReader(form.Encode()))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(cibaClientID, cibaClientSecret)

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	var body cibaBackchannelResponse
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body
}

// cibaPostCallback posts the completed flow assertion to the CIBA callback and returns the status.
func (ts *CIBATestSuite) cibaPostCallback(authID, assertion string) int {
	payload := map[string]string{
		"authId":    authID,
		"assertion": assertion,
		"type":      cibaGrantType,
	}
	jsonData, err := json.Marshal(payload)
	ts.Require().NoError(err)

	req, err := http.NewRequest("POST", testutils.TestServerURL+cibaCallbackEndpoint, bytes.NewBuffer(jsonData))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()
	return resp.StatusCode
}

// cibaTokenResult captures the outcome of a CIBA token poll.
type cibaTokenResult struct {
	statusCode  int
	accessToken string
	errorCode   string
}

// cibaPollToken polls POST /oauth2/token with the CIBA grant and returns the parsed outcome.
func (ts *CIBATestSuite) cibaPollToken(authReqID string) cibaTokenResult {
	form := url.Values{}
	form.Set("grant_type", cibaGrantType)
	form.Set("auth_req_id", authReqID)

	req, err := http.NewRequest("POST", testutils.TestServerURL+cibaTokenEndpoint,
		strings.NewReader(form.Encode()))
	ts.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(cibaClientID, cibaClientSecret)

	resp, err := ts.client.Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	var raw map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&raw)

	res := cibaTokenResult{statusCode: resp.StatusCode}
	if v, ok := raw["access_token"].(string); ok {
		res.accessToken = v
	}
	if v, ok := raw["error"].(string); ok {
		res.errorCode = v
	}
	return res
}

// extractCIBALinkParam pulls a query parameter value out of the invite link embedded in a captured
// notification body. The body is free text, so the parameter is matched directly.
func extractCIBALinkParam(text, param string) string {
	re := regexp.MustCompile(regexp.QuoteMeta(param) + `=([^&"\s\\]+)`)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
