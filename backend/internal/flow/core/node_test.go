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

type NodeTestSuite struct {
	suite.Suite
}

func TestNodeTestSuite(t *testing.T) {
	suite.Run(t, new(NodeTestSuite))
}

func (s *NodeTestSuite) TestExecuteBaseNodeReturnsError() {
	node := newTaskExecutionNode("node-1", nil, false, false)

	resp, err := node.Execute(&NodeContext{FlowID: "f1"})

	s.NotNil(err)
	s.Nil(resp)
}

func (s *NodeTestSuite) TestStartAndFinalFlags() {
	node := newDecisionNode("d1", nil, false, false)

	s.False(node.IsStartNode())
	s.False(node.IsFinalNode())

	node.SetAsStartNode()
	s.True(node.IsStartNode())

	node.SetAsFinalNode()
	s.True(node.IsFinalNode())
}

func (s *NodeTestSuite) TestNextAndPreviousNodeListBehavior() {
	// next node list behavior
	n := newPromptOnlyNode("p1", nil, false, false)

	s.Empty(n.GetNextNodeList())

	n.SetNextNodeList(nil)
	s.Empty(n.GetNextNodeList())

	n.AddNextNodeID("")
	s.Empty(n.GetNextNodeList())

	n.AddNextNodeID("n1")
	n.AddNextNodeID("n1")
	n.AddNextNodeID("n2")
	s.Len(n.GetNextNodeList(), 2)
	s.Contains(n.GetNextNodeList(), "n1")
	s.Contains(n.GetNextNodeList(), "n2")

	n.RemoveNextNodeID("n1")
	s.Len(n.GetNextNodeList(), 1)
	s.NotContains(n.GetNextNodeList(), "n1")

	n.RemoveNextNodeID("")
	n.RemoveNextNodeID("nope")
	s.Len(n.GetNextNodeList(), 1)

	// previous node list behavior
	p := newPromptOnlyNode("p2", nil, false, false)

	s.Empty(p.GetPreviousNodeList())

	p.SetPreviousNodeList(nil)
	s.Empty(p.GetPreviousNodeList())

	p.AddPreviousNodeID("")
	s.Empty(p.GetPreviousNodeList())

	p.AddPreviousNodeID("p1")
	p.AddPreviousNodeID("p2")
	p.AddPreviousNodeID("p2")
	s.Len(p.GetPreviousNodeList(), 2)
	s.Contains(p.GetPreviousNodeList(), "p1")
	s.Contains(p.GetPreviousNodeList(), "p2")

	p.RemovePreviousNodeID("p1")
	s.Len(p.GetPreviousNodeList(), 1)
	s.NotContains(p.GetPreviousNodeList(), "p1")

	p.RemovePreviousNodeID("")
	p.RemovePreviousNodeID("nope")
	s.Len(p.GetPreviousNodeList(), 1)
}

func (s *NodeTestSuite) TestInputDataAndProperties() {
	props := map[string]string{"k": "v"}
	node := newTaskExecutionNode("t1", props, false, false)

	s.Equal(props, node.GetProperties())

	inputs := []common.InputData{{Name: "i1", Required: true}}
	node.SetInputData(inputs)
	s.Equal(inputs, node.GetInputData())
}
