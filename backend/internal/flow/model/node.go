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
	"errors"
	"fmt"

	authnmodel "github.com/asgardeo/thunder/internal/authn/model"
	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// NodeResponse represents the response from a node execution
type NodeResponse struct {
	Status            constants.NodeStatus         `json:"status"`
	Type              constants.NodeResponseType   `json:"type"`
	FailureReason     string                       `json:"failure_reason,omitempty"`
	RequiredData      []InputData                  `json:"required_data,omitempty"`
	AdditionalData    map[string]string            `json:"additional_data,omitempty"`
	RedirectURL       string                       `json:"redirect_url,omitempty"`
	Actions           []Action                     `json:"actions,omitempty"`
	NextNodeID        string                       `json:"next_node_id,omitempty"`
	RuntimeData       map[string]string            `json:"runtime_data,omitempty"`
	AuthenticatedUser authnmodel.AuthenticatedUser `json:"authenticated_user,omitempty"`
	Assertion         string                       `json:"assertion,omitempty"`
}

// NodeInterface defines the interface for nodes in the graph
type NodeInterface interface {
	sysutils.ClonableInterface
	Execute(ctx *NodeContext) (*NodeResponse, *serviceerror.ServiceError)
	GetID() string
	GetType() constants.NodeType
	IsStartNode() bool
	SetAsStartNode()
	IsFinalNode() bool
	SetAsFinalNode()
	GetNextNodeList() []string
	SetNextNodeList(nextNodeIDList []string)
	AddNextNodeID(nextNodeID string)
	RemoveNextNodeID(nextNodeID string)
	GetPreviousNodeList() []string
	SetPreviousNodeList(previousNodeIDList []string)
	AddPreviousNodeID(previousNodeID string)
	RemovePreviousNodeID(previousNodeID string)
	GetInputData() []InputData
	SetInputData(inputData []InputData)
	GetExecutorConfig() *ExecutorConfig
	SetExecutorConfig(executorConfig *ExecutorConfig)
	GetExecutor() ExecutorInterface
	SetExecutor(executor ExecutorInterface)
}

// Node implements the NodeInterface
type Node struct {
	ID               string             `json:"id"`
	Type             constants.NodeType `json:"type"`
	IsStartNodeField bool               `json:"isStartNode"`
	IsFinalNodeField bool               `json:"isFinalNode"`
	NextNodeList     []string           `json:"nextNodeList"`
	PreviousNodeList []string           `json:"previousNodeList"`
	InputData        []InputData        `json:"inputData"`
	ExecutorConfig   *ExecutorConfig    `json:"executorConfig"`
}

var _ NodeInterface = (*Node)(nil)

// NewNode creates a new Node with the given type and properties.
func NewNode(id string, _type string, isStartNode bool, isFinalNode bool) (NodeInterface, error) {
	var nodeType constants.NodeType
	if _type == "" {
		return nil, errors.New("node type cannot be empty")
	} else {
		nodeType = constants.NodeType(_type)
	}

	switch nodeType {
	case constants.NodeTypeTaskExecution:
		return NewTaskExecutionNode(id, isStartNode, isFinalNode), nil
	case constants.NodeTypeDecision:
		return NewDecisionNode(id, isStartNode, isFinalNode), nil
	case constants.NodeTypePromptOnly:
		return NewPromptOnlyNode(id, isStartNode, isFinalNode), nil
	case constants.NodeTypeAuthSuccess:
		return NewTaskExecutionNode(id, isStartNode, isFinalNode), nil
	default:
		return nil, errors.New("unsupported node type: " + _type)
	}
}

// Execute executes the node
func (n *Node) Execute(ctx *NodeContext) (*NodeResponse, *serviceerror.ServiceError) {
	return nil, nil
}

// GetID returns the node's ID
func (n *Node) GetID() string {
	return n.ID
}

// GetType returns the node's type
func (n *Node) GetType() constants.NodeType {
	return n.Type
}

// IsStartNode checks if the node is a start node
func (n *Node) IsStartNode() bool {
	return n.IsStartNodeField
}

// SetAsStartNode sets the node as a start node
func (n *Node) SetAsStartNode() {
	n.IsStartNodeField = true
}

// IsFinalNode checks if the node is a final node
func (n *Node) IsFinalNode() bool {
	return n.IsFinalNodeField
}

// SetAsFinalNode sets the node as a final node
func (n *Node) SetAsFinalNode() {
	n.IsFinalNodeField = true
}

// GetNextNodeList returns the list of next node IDs
func (n *Node) GetNextNodeList() []string {
	if n.NextNodeList == nil {
		return []string{}
	}
	return n.NextNodeList
}

// SetNextNodeList sets the list of next node IDs
func (n *Node) SetNextNodeList(nextNodeIDList []string) {
	if nextNodeIDList == nil {
		n.NextNodeList = []string{}
	} else {
		n.NextNodeList = nextNodeIDList
	}
}

// AddNextNodeID adds a next node ID to the list
func (n *Node) AddNextNodeID(nextNodeID string) {
	if nextNodeID == "" {
		return
	}
	if n.NextNodeList == nil {
		n.NextNodeList = []string{}
	}
	// Check for duplicates before adding
	for _, id := range n.NextNodeList {
		if id == nextNodeID {
			return
		}
	}
	n.NextNodeList = append(n.NextNodeList, nextNodeID)
}

// RemoveNextNodeID removes a next node ID from the list
func (n *Node) RemoveNextNodeID(nextNodeID string) {
	if nextNodeID == "" || n.NextNodeList == nil {
		return
	}

	for i, id := range n.NextNodeList {
		if id == nextNodeID {
			n.NextNodeList = append(n.NextNodeList[:i], n.NextNodeList[i+1:]...)
			return
		}
	}
}

// GetPreviousNodeList returns the list of previous node IDs
func (n *Node) GetPreviousNodeList() []string {
	if n.PreviousNodeList == nil {
		return []string{}
	}
	return n.PreviousNodeList
}

// SetPreviousNodeList sets the list of previous node IDs
func (n *Node) SetPreviousNodeList(previousNodeIDList []string) {
	if previousNodeIDList == nil {
		n.PreviousNodeList = []string{}
	} else {
		n.PreviousNodeList = previousNodeIDList
	}
}

// AddPreviousNodeID adds a previous node ID to the list
func (n *Node) AddPreviousNodeID(previousNodeID string) {
	if previousNodeID == "" {
		return
	}
	if n.PreviousNodeList == nil {
		n.PreviousNodeList = []string{}
	}
	// Check for duplicates before adding
	for _, id := range n.PreviousNodeList {
		if id == previousNodeID {
			return
		}
	}
	n.PreviousNodeList = append(n.PreviousNodeList, previousNodeID)
}

// RemovePreviousNodeID removes a previous node ID from the list
func (n *Node) RemovePreviousNodeID(previousNodeID string) {
	if previousNodeID == "" || n.PreviousNodeList == nil {
		return
	}

	for i, id := range n.PreviousNodeList {
		if id == previousNodeID {
			n.PreviousNodeList = append(n.PreviousNodeList[:i], n.PreviousNodeList[i+1:]...)
			return
		}
	}
}

// GetInputData returns the input data for the node
func (n *Node) GetInputData() []InputData {
	return n.InputData
}

// SetInputData sets the input data for the node
func (n *Node) SetInputData(inputData []InputData) {
	n.InputData = inputData
}

// GetExecutorConfig returns the executor configuration for the node
func (n *Node) GetExecutorConfig() *ExecutorConfig {
	return n.ExecutorConfig
}

// SetExecutorConfig sets the executor configuration for the node
func (n *Node) SetExecutorConfig(executorConfig *ExecutorConfig) {
	n.ExecutorConfig = executorConfig
}

// GetExecutor returns the executor associated with the node
func (n *Node) GetExecutor() ExecutorInterface {
	if n.ExecutorConfig == nil {
		return nil
	}
	return n.ExecutorConfig.Executor
}

// SetExecutor sets the executor for the node
func (n *Node) SetExecutor(executor ExecutorInterface) {
	if n.ExecutorConfig == nil {
		n.ExecutorConfig = &ExecutorConfig{}
		n.ExecutorConfig.Name = executor.GetName()
	}
	n.ExecutorConfig.Executor = executor
}

// Clone creates a deep copy of the Node
func (n *Node) Clone() (sysutils.ClonableInterface, error) {
	nextCopy := append([]string{}, n.NextNodeList...)
	prevCopy := append([]string{}, n.PreviousNodeList...)
	inputCopy := append([]InputData{}, n.InputData...)

	var execConfigCopy *ExecutorConfig
	if n.ExecutorConfig != nil {
		execConfigCopy = &ExecutorConfig{
			Name:       n.ExecutorConfig.Name,
			IdpName:    n.ExecutorConfig.IdpName,
			Properties: sysutils.DeepCopyMapOfStrings(n.ExecutorConfig.Properties),
			Executor:   n.ExecutorConfig.Executor,
		}
	}

	nodeCopy, err := NewNode(n.ID, string(n.Type), n.IsStartNodeField, n.IsFinalNodeField)
	if err != nil {
		return nil, fmt.Errorf("failed to clone node: %w", err)
	}

	nodeCopy.SetNextNodeList(nextCopy)
	nodeCopy.SetPreviousNodeList(prevCopy)
	nodeCopy.SetInputData(inputCopy)
	nodeCopy.SetExecutorConfig(execConfigCopy)

	return nodeCopy, nil
}
