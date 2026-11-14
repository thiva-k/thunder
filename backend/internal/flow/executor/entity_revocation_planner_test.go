// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/internal/flow/core"
	"github.com/thunder-id/thunderid/internal/revocation"
	"github.com/thunder-id/thunderid/internal/system/cache"
	"github.com/thunder-id/thunderid/internal/system/config"
	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
	"github.com/thunder-id/thunderid/tests/mocks/actorprovidermock"
)

// testPlannedEntityID is the entity each planner test plans for.
const testPlannedEntityID = "entity-1"

type EntityRevocationPlannerTestSuite struct {
	suite.Suite
	factory core.FlowFactoryInterface
	actors  *actorprovidermock.ActorProviderMock
	planner *entityRevocationPlanner
}

func TestEntityRevocationPlannerTestSuite(t *testing.T) {
	suite.Run(t, new(EntityRevocationPlannerTestSuite))
}

func (s *EntityRevocationPlannerTestSuite) SetupTest() {
	s.Require().NoError(config.InitializeServerRuntime(s.T().TempDir(), &config.Config{}))
	s.T().Cleanup(config.ResetServerRuntime)
	s.factory, _ = core.Initialize(cache.Initialize(config.GetServerRuntime().Config.Cache, "test-deployment"))
	s.actors = actorprovidermock.NewActorProviderMock(s.T())
	s.planner = newEntityRevocationPlanner(s.factory, func(*providers.OAuthClient) time.Duration { return time.Hour })
	s.planner.setActorProvider(s.actors)
}

func (s *EntityRevocationPlannerTestSuite) execute(mode string) (*providers.ExecutorResponse, error) {
	return s.planner.Execute(&providers.NodeContext{
		Context:      context.Background(),
		ExecutorMode: mode,
		UserInputs:   map[string]string{revocationInputEntity: testPlannedEntityID},
	})
}

func (s *EntityRevocationPlannerTestSuite) planOf(resp *providers.ExecutorResponse) revocationPlan {
	s.Require().Equal(providers.ExecComplete, resp.Status)
	plan, err := decodeRevocationPlan(resp.SharedRuntimeData)
	s.Require().NoError(err)
	return plan
}

func (s *EntityRevocationPlannerTestSuite) TestDeclaresItsModesForSystemFlowsOnly() {
	meta := s.planner.GetMeta()
	s.Require().NotNil(meta)
	s.Equal([]string{ExecutorModeEntityRemoval, ExecutorModeClientRemoval, ExecutorModeClientSecretRotation,
		ExecutorModeCredentialUpdate}, meta.SupportedModes)
	s.Equal([]providers.FlowType{providers.FlowTypeSystem}, meta.SupportedFlowTypes)
}

func (s *EntityRevocationPlannerTestSuite) TestSubjectChangesRevokeWhatTheEntityWasIssued() {
	for mode, reason := range map[string]revocation.Reason{
		ExecutorModeEntityRemoval:    revocation.ReasonEntityRemoved,
		ExecutorModeCredentialUpdate: revocation.ReasonCredentialUpdated,
	} {
		s.Run(mode, func() {
			before := time.Now().UTC()
			resp, err := s.execute(mode)
			s.Require().NoError(err)

			plan := s.planOf(resp)
			s.Equal([]revocation.Criterion{{Type: revocation.CriterionTypeSubject, Value: testPlannedEntityID}},
				plan.Criteria)
			s.Equal(revocation.ModeBeforeAction, plan.Mode)
			s.Equal(reason, plan.Reason)
			s.Equal(testPlannedEntityID, plan.TargetID)
			s.False(plan.Cutoff.Before(before))
			s.Zero(plan.TTLSeconds)
		})
	}
}

func (s *EntityRevocationPlannerTestSuite) TestClientChangesRevokeWhatTheClientIssued() {
	for mode, reason := range map[string]revocation.Reason{
		ExecutorModeClientRemoval:        revocation.ReasonClientRemoved,
		ExecutorModeClientSecretRotation: revocation.ReasonClientSecretRotated,
	} {
		s.Run(mode, func() {
			s.actors.EXPECT().GetOAuthClientByID(mock.Anything, testPlannedEntityID).
				Return(&providers.OAuthClient{ClientID: "client-1"}, nil).Once()

			resp, err := s.execute(mode)
			s.Require().NoError(err)

			plan := s.planOf(resp)
			s.Equal([]revocation.Criterion{{Type: revocation.CriterionTypeApplicationKey, Value: "client-1"}},
				plan.Criteria)
			s.Equal(revocation.ModeBeforeAction, plan.Mode)
			s.Equal(reason, plan.Reason)
			s.Equal(testPlannedEntityID, plan.TargetID)
			s.Equal(int64(time.Hour/time.Second), plan.TTLSeconds)
			s.False(plan.NothingToRevoke)
		})
	}
}

func (s *EntityRevocationPlannerTestSuite) TestAnEntityWithoutAClientHasNothingToRevoke() {
	s.actors.EXPECT().GetOAuthClientByID(mock.Anything, testPlannedEntityID).Return(nil, nil).Once()

	resp, err := s.execute(ExecutorModeClientRemoval)
	s.Require().NoError(err)

	plan := s.planOf(resp)
	s.Empty(plan.Criteria)
	s.True(plan.NothingToRevoke)
	s.Equal(testPlannedEntityID, plan.TargetID)
}

func (s *EntityRevocationPlannerTestSuite) TestFailsWhenTheClientCannotBeResolved() {
	s.actors.EXPECT().GetOAuthClientByID(mock.Anything, testPlannedEntityID).
		Return(nil, &tidcommon.InternalServerError).Once()

	resp, err := s.execute(ExecutorModeClientSecretRotation)
	s.Error(err)
	s.Nil(resp)
}

func (s *EntityRevocationPlannerTestSuite) TestFailsWithoutItsCollaborators() {
	planner := newEntityRevocationPlanner(s.factory, nil)
	planner.setActorProvider(s.actors)
	_, err := planner.Execute(&providers.NodeContext{Context: context.Background(),
		ExecutorMode: ExecutorModeClientRemoval, UserInputs: map[string]string{revocationInputEntity: "e"}})
	s.ErrorContains(err, "artifact lifetime resolver")

	planner = newEntityRevocationPlanner(s.factory, func(*providers.OAuthClient) time.Duration { return 0 })
	_, err = planner.Execute(&providers.NodeContext{Context: context.Background(),
		ExecutorMode: ExecutorModeClientRemoval, UserInputs: map[string]string{revocationInputEntity: "e"}})
	s.ErrorContains(err, "actor provider")
}

func (s *EntityRevocationPlannerTestSuite) TestRefusesAnUnsupportedMode() {
	resp, err := s.execute("revoke_all")
	s.Require().NoError(err)
	s.Equal(providers.ExecFailure, resp.Status)
	s.Equal(ErrInvalidRevocationMode.Code, resp.Error.Code)
}

func (s *EntityRevocationPlannerTestSuite) TestAsksForTheEntityWhenItIsMissing() {
	resp, err := s.planner.Execute(&providers.NodeContext{Context: context.Background(),
		ExecutorMode: ExecutorModeEntityRemoval, UserInputs: map[string]string{}})
	s.Require().NoError(err)
	s.Equal(providers.ExecUserInputRequired, resp.Status)
	s.Require().Len(resp.Inputs, 1)
	s.Equal(revocationInputEntity, resp.Inputs[0].Identifier)
	s.Empty(resp.SharedRuntimeData)
}
