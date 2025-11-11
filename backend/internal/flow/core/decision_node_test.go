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

package core

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/flow/common"
)

type DecisionNodeTestSuite struct {
	suite.Suite
}

func TestDecisionNodeTestSuite(t *testing.T) {
	suite.Run(t, new(DecisionNodeTestSuite))
}

func (s *DecisionNodeTestSuite) TestNewDecisionNode() {
	node := newDecisionNode("decision-1", map[string]string{"key": "value"}, true, false)

	s.NotNil(node)
	s.Equal("decision-1", node.GetID())
	s.Equal(common.NodeTypeDecision, node.GetType())
	s.True(node.IsStartNode())
	s.False(node.IsFinalNode())
}

func (s *DecisionNodeTestSuite) TestExecuteWithValidActionID() {
	tests := []struct {
		name            string
		currentActionID string
		expectedNextID  string
	}{
		{"Valid action ID matching first next node", "next-1", "next-1"},
		{"Valid action ID matching second next node", "next-2", "next-2"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			node := newDecisionNode("decision-1", map[string]string{}, false, false)
			node.AddNextNodeID("next-1")
			node.AddNextNodeID("next-2")

			ctx := &NodeContext{FlowID: "test-flow", CurrentActionID: tt.currentActionID}
			resp, err := node.Execute(ctx)

			s.Nil(err)
			s.NotNil(resp)
			s.Equal(common.NodeStatusComplete, resp.Status)
			s.Equal(tt.expectedNextID, resp.NextNodeID)
		})
	}
}

func (s *DecisionNodeTestSuite) TestExecuteWithInvalidActionID() {
	node := newDecisionNode("decision-1", map[string]string{}, false, false)
	node.AddNextNodeID("next-1")
	node.AddNextNodeID("next-2")

	ctx := &NodeContext{FlowID: "test-flow", CurrentActionID: "invalid-node"}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusFailure, resp.Status)
	s.Empty(resp.NextNodeID)
	s.NotEmpty(resp.FailureReason)
}

func (s *DecisionNodeTestSuite) TestExecuteWithoutActionID() {
	node := newDecisionNode("decision-1", map[string]string{}, false, false)
	node.AddNextNodeID("next-1")
	node.AddNextNodeID("next-2")

	ctx := &NodeContext{FlowID: "test-flow", CurrentActionID: ""}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal(common.NodeResponseTypeView, resp.Type)
	s.NotNil(resp.Actions)
	s.Len(resp.Actions, 2)

	actionIDs := make([]string, len(resp.Actions))
	for i, action := range resp.Actions {
		actionIDs[i] = action.ID
		s.Equal(common.ActionTypeView, action.Type)
	}
	s.Contains(actionIDs, "next-1")
	s.Contains(actionIDs, "next-2")
}

func (s *DecisionNodeTestSuite) TestExecuteNoNextNodesWithActionIDFailure() {
	node := newDecisionNode("decision-1", map[string]string{}, false, false)
	ctx := &NodeContext{FlowID: "test-flow", CurrentActionID: "some-action"}

	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusFailure, resp.Status)
	s.NotEmpty(resp.FailureReason)
}

func (s *DecisionNodeTestSuite) TestExecuteNoNextNodesWithoutActionIDError() {
	node := newDecisionNode("decision-1", map[string]string{}, false, false)
	ctx := &NodeContext{FlowID: "test-flow", CurrentActionID: ""}

	resp, err := node.Execute(ctx)

	s.NotNil(err)
	s.Nil(resp)
}
