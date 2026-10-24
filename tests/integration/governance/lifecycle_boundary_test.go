// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package governance

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/thunder-id/thunderid/tests/integration/flow/common"
	"github.com/thunder-id/thunderid/tests/integration/testutils"
)

const (
	boundaryConsoleURI = "https://localhost:8095/console"

	// The bootstrap organization unit and system resource server: the root permission lives there.
	bootstrapOUID          = "01900000-0000-7000-8000-000000000001"
	systemResourceServerID = "01900000-0000-7000-8000-000000000020"
	boundaryScopedRSHandle = "https://boundary-test.example.com/system"

	// adminAuthenticationRequiredCode is the administration gate's answer to a caller without root.
	// /flow/execute is a public path, so the middleware drops a token that fails its root check and
	// the gate sees an anonymous caller.
	adminAuthenticationRequiredCode = "FES-1017"
)

// rootRequired is the answer the administration gate gives a caller without the root permission.
var rootRequired = flowAnswer{status: http.StatusUnauthorized, code: adminAuthenticationRequiredCode}

// LifecycleBoundaryTestSuite checks who may run unlock, and on whom. Each refusal checks the answer
// that refused it. The cross-unit and undecidable-authorization branches are unit-tested in
// identitygovernance: the administration gate admits only root, which is outside every boundary, so no
// caller a test can mint reaches them. It runs under the shipped lockout policy.
type LifecycleBoundaryTestSuite struct {
	suite.Suite
	unitA, unitB *userFixture
	rootToken    string
	rootID       string
	scopedToken  string
}

func TestLifecycleBoundaryTestSuite(t *testing.T) {
	suite.Run(t, new(LifecycleBoundaryTestSuite))
}

func (ts *LifecycleBoundaryTestSuite) SetupSuite() {
	t := ts.T()
	schema := testutils.UserType{Schema: credentialsSchema(nil)}
	ts.unitA = newUserFixture(t, "boundary_a", schema)
	ts.unitB = newUserFixture(t, "boundary_b", schema)

	// A second root operator, so a self-targeting test acts on a throwaway account rather than the
	// administrator every other suite runs as.
	rootID, rootUsername, rootPassword := ts.unitA.newUser(t, "root", nil)
	ts.rootID = rootID
	roleID, err := testutils.CreateRole(testutils.Role{
		Name: "Boundary Root " + ts.unitA.suffix,
		OUID: bootstrapOUID,
		Permissions: []testutils.ResourcePermissions{
			{ResourceServerID: systemResourceServerID, Permissions: []string{"system"}},
		},
		Assignments: []testutils.Assignment{{ID: rootID, Type: "user"}},
	})
	ts.Require().NoError(err, "failed to grant the root permission")
	cleanupRole(t, roleID)
	root, err := testutils.ObtainAccessTokenWithPassword("CONSOLE", boundaryConsoleURI, "system",
		rootUsername, rootPassword, true)
	ts.Require().NoError(err, "the root operator must sign in to the console")
	ts.rootToken = root.AccessToken

	// An operator holding user management in unit A only, the closest thing to a delegated
	// administrator the product can express.
	scopedID, scopedUsername, scopedPassword := ts.unitA.newUser(t, "scoped", nil)
	rsID, err := testutils.CreateSystemScopedResourceServer(ts.unitA.ouID,
		"Governance Boundary RS "+ts.unitA.suffix, boundaryScopedRSHandle, "user")
	ts.Require().NoError(err, "failed to create the scoped resource server")
	t.Cleanup(func() {
		if err := testutils.DeleteResourceServerWithChildren(rsID); err != nil {
			t.Logf("teardown: failed to delete resource server: %v", err)
		}
	})
	roleID, err = testutils.CreateRole(testutils.Role{
		Name: "Boundary User Manager " + ts.unitA.suffix,
		OUID: ts.unitA.ouID,
		Permissions: []testutils.ResourcePermissions{
			{ResourceServerID: rsID, Permissions: []string{"system:user", "system:user:view"}},
		},
		Assignments: []testutils.Assignment{{ID: scopedID, Type: "user"}},
	})
	ts.Require().NoError(err, "failed to grant user management")
	cleanupRole(t, roleID)
	scoped, err := testutils.ObtainAccessTokenWithPassword("CONSOLE", boundaryConsoleURI,
		"system:user system:user:view", scopedUsername, scopedPassword, true, "", boundaryScopedRSHandle)
	ts.Require().NoError(err, "the scoped operator must sign in to the console")
	ts.scopedToken = scoped.AccessToken
}

// cleanupRole deletes the role when t ends.
func cleanupRole(t *testing.T, roleID string) {
	t.Helper()
	t.Cleanup(func() {
		if err := testutils.DeleteRole(roleID); err != nil {
			t.Logf("teardown: failed to delete role %s: %v", roleID, err)
		}
	})
}

// An operator cannot unlock themselves, even with root. The refusal is the authority's, after the
// permission check passed.
func (ts *LifecycleBoundaryTestSuite) TestAnOperatorCannotGovernThemselves() {
	ts.refused(ts.rootToken, unlockFlowHandle, ts.rootID, governanceRefusal, "a self-targeted unlock must be refused")
}

// Administration is root-only, so an operator holding user management in one unit is refused at the
// administration gate for a target in another unit and in their own. The target is locked first, so a
// refused unlock is seen to release nothing.
func (ts *LifecycleBoundaryTestSuite) TestAnOperatorWithoutRootIsRefusedInAndOutsideTheirUnit() {
	for _, target := range []struct {
		name string
		unit *userFixture
	}{{"a", ts.unitA}, {"b", ts.unitB}} {
		targetID, username, _ := target.unit.newUser(ts.T(), "scoped_target", nil)
		failAuthentication(ts.T(), "username", username, shippedUserLockThreshold)
		ts.Require().Equal("LOCKED", readStatus(ts.T(), targetID).Value,
			"unit %s: the target must be locked first", target.name)

		ts.refused(ts.scopedToken, unlockFlowHandle, targetID, rootRequired,
			"unit %s: unlock must be refused without root", target.name)

		ts.Equal("LOCKED", readStatus(ts.T(), targetID).Value, "unit %s: the lock must stay", target.name)
	}
}

// An unlock started without a subject asks for one and releases nothing.
func (ts *LifecycleBoundaryTestSuite) TestAnUnlockWithoutASubjectAsksForOne() {
	lockedID, username, _ := ts.unitA.newUser(ts.T(), "no_subject", nil)
	failAuthentication(ts.T(), "username", username, shippedUserLockThreshold)
	ts.Require().Equal("LOCKED", readStatus(ts.T(), lockedID).Value, "the account must be locked first")

	status, raw := executeAdministrationFlowAs(ts.T(), ts.rootToken, unlockFlowHandle, map[string]string{})
	step := flowStepOf(ts.T(), status, raw)

	ts.Equal("INCOMPLETE", step.FlowStatus, "an unlock without a subject must ask for one: %s", raw)
	ts.True(common.HasInput(step.Data.Inputs, subjectInput), "the flow must ask for the subject: %s", raw)
	ts.Equal("LOCKED", readStatus(ts.T(), lockedID).Value, "nothing may be released without a subject")
}

// Owning the account does not authorize unlocking it. A locked holder signed in to the Console with no
// permission cannot unlock themselves: the administration gate refuses, and the lock stays.
func (ts *LifecycleBoundaryTestSuite) TestOwningTheAccountDoesNotAuthorizeItsLifecycle() {
	ownerID, username, password := ts.unitA.newUser(ts.T(), "owner", nil)
	console, err := testutils.ObtainAccessTokenWithPassword("CONSOLE", boundaryConsoleURI, "openid",
		username, password, true)
	ts.Require().NoError(err, "the holder must be able to sign in to the console")

	failAuthentication(ts.T(), "username", username, shippedUserLockThreshold)
	ts.Require().Equal("LOCKED", readStatus(ts.T(), ownerID).Value, "the holder must be locked first")

	ts.refused(console.AccessToken, unlockFlowHandle, ownerID, rootRequired, "the holder's own unlock")

	ts.Equal("LOCKED", readStatus(ts.T(), ownerID).Value, "the holder must not have released their lock")
}

// refused runs a shipped flow as the given operator and requires the given refusal.
func (ts *LifecycleBoundaryTestSuite) refused(token, handle, subject string, want flowAnswer, msg string,
	args ...interface{}) {
	ts.T().Helper()
	got := attemptAdministrationFlowAs(ts.T(), token, handle, map[string]string{subjectInput: subject})
	ts.Equal(want, got, append([]interface{}{msg}, args...)...)
}
