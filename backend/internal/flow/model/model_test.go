/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package model

import (
	"testing"

	"github.com/asgardeo/thunder/internal/flow/constants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ModelTestSuite struct {
	suite.Suite
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, new(ModelTestSuite))
}

// Basic Graph functionality tests - these should work with the existing Graph implementation

func (suite *ModelTestSuite) TestNewGraph_WithIDAndType() {
	graphID := "test-graph-123"
	graphType := constants.FlowTypeAuthentication

	graph := NewGraph(graphID, graphType)

	assert.NotNil(suite.T(), graph)
	assert.Equal(suite.T(), graphID, graph.GetID())
	assert.Equal(suite.T(), graphType, graph.GetType())
	assert.NotNil(suite.T(), graph.GetNodes())
	assert.Empty(suite.T(), graph.GetNodes())
	assert.NotNil(suite.T(), graph.GetEdges())
	assert.Empty(suite.T(), graph.GetEdges())
}

func (suite *ModelTestSuite) TestNewGraph_WithEmptyID() {
	graphType := constants.FlowTypeAuthentication

	graph := NewGraph("", graphType)

	assert.NotNil(suite.T(), graph)
	assert.NotEmpty(suite.T(), graph.GetID(), "Should generate UUID when ID is empty")
	assert.Equal(suite.T(), graphType, graph.GetType())
}

func (suite *ModelTestSuite) TestNewGraph_WithEmptyType() {
	graphID := "test-graph-123"

	graph := NewGraph(graphID, "")

	assert.NotNil(suite.T(), graph)
	assert.Equal(suite.T(), graphID, graph.GetID())
	assert.Equal(suite.T(), constants.FlowTypeAuthentication, graph.GetType(), "Should default to authentication type")
}

func (suite *ModelTestSuite) TestGraph_AddNode_Success() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node := NewPromptOnlyNode("prompt-node", false, false)

	err := graph.AddNode(node)

	assert.NoError(suite.T(), err)
	nodes := graph.GetNodes()
	assert.Len(suite.T(), nodes, 1)
	assert.Contains(suite.T(), nodes, "prompt-node")
	assert.Equal(suite.T(), node, nodes["prompt-node"])
}

func (suite *ModelTestSuite) TestGraph_AddNode_NilNode() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	err := graph.AddNode(nil)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "node cannot be nil")
	assert.Empty(suite.T(), graph.GetNodes())
}

func (suite *ModelTestSuite) TestGraph_GetNode_Exists() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)
	node := NewPromptOnlyNode("prompt-node", false, false)
	err := graph.AddNode(node)
	assert.NoError(suite.T(), err)

	retrievedNode, exists := graph.GetNode("prompt-node")

	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), node, retrievedNode)
}

func (suite *ModelTestSuite) TestGraph_GetNode_NotExists() {
	graph := NewGraph("test-graph", constants.FlowTypeAuthentication)

	retrievedNode, exists := graph.GetNode("non-existent")

	assert.False(suite.T(), exists)
	assert.Nil(suite.T(), retrievedNode)
}
