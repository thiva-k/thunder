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
	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// NodeInterface defines the interface for nodes in the graph
type NodeInterface interface {
	Execute(ctx *NodeContext) (*common.NodeResponse, *serviceerror.ServiceError)
	GetID() string
	GetType() common.NodeType
	GetProperties() map[string]string
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
	GetInputData() []common.InputData
	SetInputData(inputData []common.InputData)
}

// ExecutorBackedNodeInterface extends NodeInterface for nodes backed by executors.
// Only task execution nodes implement this interface to delegate their execution logic to executors.
type ExecutorBackedNodeInterface interface {
	NodeInterface
	GetExecutorName() string
	SetExecutorName(name string)
	GetExecutor() ExecutorInterface
	SetExecutor(executor ExecutorInterface)
}

// node implements the NodeInterface
type node struct {
	id               string
	_type            common.NodeType
	properties       map[string]string
	isStartNode      bool
	isFinalNode      bool
	nextNodeList     []string
	previousNodeList []string
	inputData        []common.InputData
}

var _ NodeInterface = (*node)(nil)

// Execute executes the node
func (n *node) Execute(ctx *NodeContext) (*common.NodeResponse, *serviceerror.ServiceError) {
	return nil, nil
}

// GetID returns the node's ID
func (n *node) GetID() string {
	return n.id
}

// GetType returns the node's type
func (n *node) GetType() common.NodeType {
	return n._type
}

// GetProperties returns the node's properties
func (n *node) GetProperties() map[string]string {
	return n.properties
}

// IsStartNode checks if the node is a start node
func (n *node) IsStartNode() bool {
	return n.isStartNode
}

// SetAsStartNode sets the node as a start node
func (n *node) SetAsStartNode() {
	n.isStartNode = true
}

// IsFinalNode checks if the node is a final node
func (n *node) IsFinalNode() bool {
	return n.isFinalNode
}

// SetAsFinalNode sets the node as a final node
func (n *node) SetAsFinalNode() {
	n.isFinalNode = true
}

// GetNextNodeList returns the list of next node IDs
func (n *node) GetNextNodeList() []string {
	if n.nextNodeList == nil {
		return []string{}
	}
	return n.nextNodeList
}

// SetNextNodeList sets the list of next node IDs
func (n *node) SetNextNodeList(nextNodeIDList []string) {
	if nextNodeIDList == nil {
		n.nextNodeList = []string{}
	} else {
		n.nextNodeList = nextNodeIDList
	}
}

// AddNextNodeID adds a next node ID to the list
func (n *node) AddNextNodeID(nextNodeID string) {
	if nextNodeID == "" {
		return
	}
	if n.nextNodeList == nil {
		n.nextNodeList = []string{}
	}
	// Check for duplicates before adding
	for _, id := range n.nextNodeList {
		if id == nextNodeID {
			return
		}
	}
	n.nextNodeList = append(n.nextNodeList, nextNodeID)
}

// RemoveNextNodeID removes a next node ID from the list
func (n *node) RemoveNextNodeID(nextNodeID string) {
	if nextNodeID == "" || n.nextNodeList == nil {
		return
	}

	for i, id := range n.nextNodeList {
		if id == nextNodeID {
			n.nextNodeList = append(n.nextNodeList[:i], n.nextNodeList[i+1:]...)
			return
		}
	}
}

// GetPreviousNodeList returns the list of previous node IDs
func (n *node) GetPreviousNodeList() []string {
	if n.previousNodeList == nil {
		return []string{}
	}
	return n.previousNodeList
}

// SetPreviousNodeList sets the list of previous node IDs
func (n *node) SetPreviousNodeList(previousNodeIDList []string) {
	if previousNodeIDList == nil {
		n.previousNodeList = []string{}
	} else {
		n.previousNodeList = previousNodeIDList
	}
}

// AddPreviousNodeID adds a previous node ID to the list
func (n *node) AddPreviousNodeID(previousNodeID string) {
	if previousNodeID == "" {
		return
	}
	if n.previousNodeList == nil {
		n.previousNodeList = []string{}
	}
	// Check for duplicates before adding
	for _, id := range n.previousNodeList {
		if id == previousNodeID {
			return
		}
	}
	n.previousNodeList = append(n.previousNodeList, previousNodeID)
}

// RemovePreviousNodeID removes a previous node ID from the list
func (n *node) RemovePreviousNodeID(previousNodeID string) {
	if previousNodeID == "" || n.previousNodeList == nil {
		return
	}

	for i, id := range n.previousNodeList {
		if id == previousNodeID {
			n.previousNodeList = append(n.previousNodeList[:i], n.previousNodeList[i+1:]...)
			return
		}
	}
}

// GetInputData returns the input data for the node
func (n *node) GetInputData() []common.InputData {
	return n.inputData
}

// SetInputData sets the input data for the node
func (n *node) SetInputData(inputData []common.InputData) {
	n.inputData = inputData
}
