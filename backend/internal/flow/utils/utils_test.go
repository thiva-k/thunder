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

package utils

import (
	"testing"

	"github.com/asgardeo/thunder/internal/flow/jsonmodel"
	"github.com/asgardeo/thunder/internal/flow/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

// BuildGraphFromDefinition tests

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_NilDefinition() {
	graph, err := BuildGraphFromDefinition(nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), graph)
	assert.Contains(suite.T(), err.Error(), "graph definition is nil or has no nodes")
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_EmptyNodes() {
	definition := &jsonmodel.GraphDefinition{
		ID:    "test-graph",
		Type:  "authentication",
		Nodes: []jsonmodel.NodeDefinition{},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), graph)
	assert.Contains(suite.T(), err.Error(), "graph definition is nil or has no nodes")
}

func (suite *UtilsTestSuite) TestBuildGraphFromDefinition_InvalidGraphType() {
	definition := &jsonmodel.GraphDefinition{
		ID:   "test-graph",
		Type: "invalid-type",
		Nodes: []jsonmodel.NodeDefinition{
			{
				ID:   "test-node",
				Type: "prompt-only",
			},
		},
	}

	graph, err := BuildGraphFromDefinition(definition)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), graph)
	assert.Contains(suite.T(), err.Error(), "error while retrieving graph type")
}

// GetExecutorByName tests

func (suite *UtilsTestSuite) TestGetExecutorByName_NilConfig() {
	executor, err := GetExecutorByName(nil)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), executor)
	assert.Contains(suite.T(), err.Error(), "executor configuration cannot be nil")
}

func (suite *UtilsTestSuite) TestGetExecutorByName_EmptyName() {
	config := &model.ExecutorConfig{
		Name: "",
	}

	executor, err := GetExecutorByName(config)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), executor)
	assert.Contains(suite.T(), err.Error(), "executor name cannot be empty")
}

func (suite *UtilsTestSuite) TestGetExecutorByName_UnknownExecutor() {
	config := &model.ExecutorConfig{
		Name: "UnknownExecutor",
	}

	executor, err := GetExecutorByName(config)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), executor)
	assert.Contains(suite.T(), err.Error(), "executor with name UnknownExecutor not found")
}

func (suite *UtilsTestSuite) TestGetExecutorByName_AttributeCollector() {
	config := &model.ExecutorConfig{
		Name: "AttributeCollector",
		Properties: map[string]string{
			"attribute": "email",
		},
	}

	executor, err := GetExecutorByName(config)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), executor)
}

func (suite *UtilsTestSuite) TestGetExecutorByName_ProvisioningExecutor() {
	config := &model.ExecutorConfig{
		Name: "ProvisioningExecutor",
		Properties: map[string]string{
			"userstore": "primary",
		},
	}

	executor, err := GetExecutorByName(config)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), executor)
}

func (suite *UtilsTestSuite) TestGetExecutorByName_AuthAssertExecutor() {
	config := &model.ExecutorConfig{
		Name: "AuthAssertExecutor",
		Properties: map[string]string{
			"assertion_type": "saml",
		},
	}

	executor, err := GetExecutorByName(config)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), executor)
}
