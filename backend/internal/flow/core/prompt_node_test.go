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

type PromptOnlyNodeTestSuite struct {
	suite.Suite
}

func TestPromptOnlyNodeTestSuite(t *testing.T) {
	suite.Run(t, new(PromptOnlyNodeTestSuite))
}

func (s *PromptOnlyNodeTestSuite) TestNewPromptOnlyNode() {
	node := newPromptOnlyNode("prompt-1", map[string]interface{}{"key": "value"}, true, false)

	s.NotNil(node)
	s.Equal("prompt-1", node.GetID())
	s.Equal(common.NodeTypePromptOnly, node.GetType())
	s.True(node.IsStartNode())
	s.False(node.IsFinalNode())
}

func (s *PromptOnlyNodeTestSuite) TestExecuteNoInputData() {
	node := newPromptOnlyNode("prompt-1", map[string]interface{}{}, false, false)
	ctx := &NodeContext{FlowID: "test-flow", UserInputData: map[string]string{}}

	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status)
	s.Equal(common.NodeResponseType(""), resp.Type)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithRequiredData() {
	tests := []struct {
		name           string
		userInputData  map[string]string
		expectComplete bool
		requiredCount  int
	}{
		{"No user input provided", map[string]string{}, false, 2},
		{
			"All required data provided",
			map[string]string{"username": "testuser", "email": "test@example.com"},
			true,
			0,
		},
		{"Partial data provided", map[string]string{"username": "testuser"}, false, 1},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			node := newPromptOnlyNode("prompt-1", map[string]interface{}{}, false, false)
			node.SetInputData([]common.InputData{
				{Name: "username", Required: true},
				{Name: "email", Required: true},
			})

			ctx := &NodeContext{FlowID: "test-flow", UserInputData: tt.userInputData}
			resp, err := node.Execute(ctx)

			s.Nil(err)
			s.NotNil(resp)

			if tt.expectComplete {
				s.Equal(common.NodeStatusComplete, resp.Status)
				s.Equal(common.NodeResponseType(""), resp.Type)
			} else {
				s.Equal(common.NodeStatusIncomplete, resp.Status)
				s.Equal(common.NodeResponseTypeView, resp.Type)
				s.Len(resp.RequiredData, tt.requiredCount)
			}
		})
	}
}

func (s *PromptOnlyNodeTestSuite) TestExecuteWithOptionalData() {
	node := newPromptOnlyNode("prompt-1", map[string]interface{}{}, false, false)
	node.SetInputData([]common.InputData{
		{Name: "username", Required: true},
		{Name: "nickname", Required: false},
	})

	ctx := &NodeContext{FlowID: "test-flow", UserInputData: map[string]string{"username": "testuser"}}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusComplete, resp.Status)
	s.Equal(common.NodeResponseType(""), resp.Type)
}

func (s *PromptOnlyNodeTestSuite) TestExecuteMissingRequiredOnly() {
	node := newPromptOnlyNode("prompt-1", map[string]interface{}{}, false, false)
	node.SetInputData([]common.InputData{
		{Name: "username", Required: true},
		{Name: "nickname", Required: false},
	})

	ctx := &NodeContext{FlowID: "test-flow", UserInputData: map[string]string{"nickname": "testnick"}}
	resp, err := node.Execute(ctx)

	s.Nil(err)
	s.NotNil(resp)
	s.Equal(common.NodeStatusIncomplete, resp.Status)
	s.Equal(common.NodeResponseTypeView, resp.Type)
	s.Len(resp.RequiredData, 1)

	foundRequired := false
	for _, data := range resp.RequiredData {
		if data.Name == "username" && data.Required {
			foundRequired = true
		}
	}
	s.True(foundRequired)
}
