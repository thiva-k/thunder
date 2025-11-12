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
	"github.com/asgardeo/thunder/internal/system/log"
)

const decisionNodeLoggerComponentName = "DecisionNode"

// decisionNode represents a node that makes decisions based on input data.
type decisionNode struct {
	*node
}

// newDecisionNode creates a new DecisionNode with the given details.
func newDecisionNode(id string, properties map[string]interface{}, isStartNode bool, isFinalNode bool) NodeInterface {
	return &decisionNode{
		node: &node{
			id:               id,
			_type:            common.NodeTypeDecision,
			properties:       properties,
			isStartNode:      isStartNode,
			isFinalNode:      isFinalNode,
			nextNodeList:     []string{},
			previousNodeList: []string{},
			inputData:        []common.InputData{},
		},
	}
}

// Execute executes the decision node logic based on the current context.
func (n *decisionNode) Execute(ctx *NodeContext) (*common.NodeResponse, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, decisionNodeLoggerComponentName),
		log.String(log.LoggerKeyNodeID, n.GetID()),
		log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing decision node")

	triggeredActionID := ctx.CurrentActionID
	if triggeredActionID != "" {
		return n.triggerAction(triggeredActionID)
	}

	return n.prepareActionInput(logger)
}

// triggerAction processes the action triggered by the user and determines the next node to transition to.
func (n *decisionNode) triggerAction(actionID string) (*common.NodeResponse,
	*serviceerror.ServiceError) {
	nextNodeIDs := n.GetNextNodeList()
	if len(nextNodeIDs) == 0 {
		return &common.NodeResponse{
			Status:        common.NodeStatusFailure,
			Type:          "",
			FailureReason: "No next nodes defined for the decision node.",
		}, nil
	}

	var nextNodeID string
	for _, nextNodeIDCandidate := range nextNodeIDs {
		if nextNodeIDCandidate == actionID {
			nextNodeID = nextNodeIDCandidate
			break
		}
	}
	if nextNodeID == "" {
		return &common.NodeResponse{
			Status:        common.NodeStatusFailure,
			Type:          "",
			FailureReason: "No matching next node found for the triggered action ID.",
		}, nil
	}

	return &common.NodeResponse{
		Status:     common.NodeStatusComplete,
		Type:       "",
		NextNodeID: nextNodeID,
	}, nil
}

// prepareActionInput prepares the input for the action to be triggered by the user.
func (n *decisionNode) prepareActionInput(logger *log.Logger) (
	*common.NodeResponse, *serviceerror.ServiceError) {
	actions := n.getActionsList()
	if len(actions) == 0 {
		logger.Error("No outgoing edges defined for the decision node")
		return nil, &serviceerror.InternalServerError
	}

	return &common.NodeResponse{
		Status:         common.NodeStatusIncomplete,
		Type:           common.NodeResponseTypeView,
		Actions:        actions,
		FailureReason:  "",
		RequiredData:   make([]common.InputData, 0),
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}, nil
}

// getActionsList retrieves the list of actions available for the decision node.
func (n *decisionNode) getActionsList() []common.Action {
	actions := []common.Action{}
	for _, nextNodeID := range n.GetNextNodeList() {
		action := common.Action{
			Type: common.ActionTypeView,
			ID:   nextNodeID,
		}
		actions = append(actions, action)
	}
	return actions
}
