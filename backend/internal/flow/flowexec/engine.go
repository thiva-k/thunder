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

package flowexec

import (
	"errors"
	"fmt"
	"time"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// flowEngineInterface defines the interface for the flow engine.
type flowEngineInterface interface {
	Execute(ctx *EngineContext) (FlowStep, *serviceerror.ServiceError)
}

// FlowEngine is the main engine implementation for orchestrating flow executions.
type flowEngine struct {
	executorRegistry executor.ExecutorRegistryInterface
}

// GetFlowEngine returns a singleton instance of FlowEngine.
func newFlowEngine(executorRegistry executor.ExecutorRegistryInterface) flowEngineInterface {
	return &flowEngine{
		executorRegistry: executorRegistry,
	}
}

// Execute executes a step in the flow
func (fe *flowEngine) Execute(ctx *EngineContext) (FlowStep, *serviceerror.ServiceError) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowEngine"))

	flowStep := FlowStep{
		FlowID: ctx.FlowID,
	}

	currentNode, err := fe.setCurrentExecutionNode(ctx, logger)
	if err != nil {
		return flowStep, err
	}

	// Execute the graph nodes until a terminal condition is met or currentNode is nil
	for currentNode != nil {
		logger.Debug("Executing node", log.String("nodeID", currentNode.GetID()),
			log.String("nodeType", string(currentNode.GetType())))

		svcErr := fe.setNodeExecutor(currentNode, logger)
		if svcErr != nil {
			return flowStep, svcErr
		}

		nodeCtx := &core.NodeContext{
			FlowID:            ctx.FlowID,
			FlowType:          ctx.FlowType,
			AppID:             ctx.AppID,
			CurrentActionID:   ctx.CurrentActionID,
			NodeInputData:     ctx.CurrentNode.GetInputData(),
			UserInputData:     ctx.UserInputData,
			RuntimeData:       ctx.RuntimeData,
			Application:       ctx.Application,
			AuthenticatedUser: ctx.AuthenticatedUser,
			ExecutionHistory:  ctx.ExecutionHistory,
		}
		if nodeCtx.NodeInputData == nil {
			nodeCtx.NodeInputData = make([]common.InputData, 0)
		}
		if nodeCtx.UserInputData == nil {
			nodeCtx.UserInputData = make(map[string]string)
		}
		if nodeCtx.RuntimeData == nil {
			nodeCtx.RuntimeData = make(map[string]string)
		}

		executionStartTime := time.Now().Unix()
		nodeResp, nodeErr := currentNode.Execute(nodeCtx)
		executionEndTime := time.Now().Unix()

		recordNodeExecution(ctx, currentNode, nodeResp, nodeErr, executionStartTime, executionEndTime)

		if nodeErr != nil {
			return flowStep, nodeErr
		}

		fe.updateContextWithNodeResponse(ctx, nodeResp)

		nextNode, continueExecution, svcErr := fe.processNodeResponse(ctx, currentNode, nodeResp,
			&flowStep, logger)
		if svcErr != nil {
			return flowStep, svcErr
		}
		if !continueExecution {
			return flowStep, nil
		}
		currentNode = nextNode
	}

	// If we reach here, it means the flow has been executed successfully.
	flowStep.Status = common.FlowStatusComplete
	if ctx.CurrentNodeResponse != nil && ctx.CurrentNodeResponse.Assertion != "" {
		flowStep.Assertion = ctx.CurrentNodeResponse.Assertion
	}

	return flowStep, nil
}

// setCurrentExecutionNode sets the current execution node in the context and returns it.
func (fe *flowEngine) setCurrentExecutionNode(ctx *EngineContext, logger *log.Logger) (
	core.NodeInterface, *serviceerror.ServiceError) {
	graph := ctx.Graph
	if graph == nil {
		logger.Error("Flow graph is not initialized in the context")
		return nil, &serviceerror.InternalServerError
	}

	currentNode := ctx.CurrentNode
	if currentNode == nil {
		logger.Debug("Current node is nil. Setting start node as the current node.")
		var err error
		currentNode, err = graph.GetStartNode()
		if err != nil {
			logger.Error("Start node not found in the flow graph", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		ctx.CurrentNode = currentNode
	}

	// Initialize execution history map if needed
	if ctx.ExecutionHistory == nil {
		ctx.ExecutionHistory = make(map[string]*common.NodeExecutionRecord)
	}

	return currentNode, nil
}

// setNodeExecutor sets the executor for the given node if it is not already set.
func (fe *flowEngine) setNodeExecutor(node core.NodeInterface, logger *log.Logger) *serviceerror.ServiceError {
	if node.GetType() != common.NodeTypeTaskExecution {
		return nil
	}
	executableNode, ok := node.(core.ExecutorBackedNodeInterface)
	if !ok {
		logger.Error("Task execution node does not implement ExecutorBackedNodeInterface",
			log.String("nodeID", node.GetID()))
		return &serviceerror.InternalServerError
	}

	// Return if executor is already set
	if executableNode.GetExecutor() != nil {
		return nil
	}

	logger.Debug("Executor not set for the node. Constructing executor.", log.String("nodeID", node.GetID()))

	executorName := executableNode.GetExecutorName()
	if executorName == "" {
		logger.Error("Executor name not configured for executable node", log.String("nodeID", node.GetID()))
		return &serviceerror.InternalServerError
	}

	executor, err := fe.getExecutorByName(executorName)
	if err != nil {
		logger.Error("Error constructing executor for node", log.String("nodeID", node.GetID()),
			log.String("executorName", executorName), log.Error(err))
		return &serviceerror.InternalServerError
	}

	executableNode.SetExecutor(executor)
	return nil
}

// getExecutorByName retrieves executor instance from the executor registry.
func (fe *flowEngine) getExecutorByName(executorName string) (core.ExecutorInterface, error) {
	exec, err := fe.executorRegistry.GetExecutor(executorName)
	if err != nil {
		return nil, fmt.Errorf("failed to get executor '%s': %w", executorName, err)
	}

	return exec, nil
}

// updateContextWithNodeResponse updates the engine context with the node response and authenticated user.
func (fe *flowEngine) updateContextWithNodeResponse(engineCtx *EngineContext, nodeResp *common.NodeResponse) {
	engineCtx.CurrentNodeResponse = nodeResp
	engineCtx.CurrentActionID = ""

	// Handle runtime data from the node response
	if len(nodeResp.RuntimeData) > 0 {
		if engineCtx.RuntimeData == nil {
			engineCtx.RuntimeData = make(map[string]string)
		}
		engineCtx.RuntimeData = sysutils.MergeStringMaps(engineCtx.RuntimeData, nodeResp.RuntimeData)
	}

	// Handle authenticated user from the node response
	if nodeResp.AuthenticatedUser.IsAuthenticated || engineCtx.FlowType == common.FlowTypeRegistration {
		prevAuthnUserAttrs := engineCtx.AuthenticatedUser.Attributes
		engineCtx.AuthenticatedUser = nodeResp.AuthenticatedUser

		// If engine context already had authenticated user attributes, merge them with the new ones.
		// Here if the same attribute exists in both, the one from the node response will take precedence.
		if len(prevAuthnUserAttrs) > 0 {
			if engineCtx.AuthenticatedUser.Attributes == nil {
				engineCtx.AuthenticatedUser.Attributes = prevAuthnUserAttrs
			} else {
				engineCtx.AuthenticatedUser.Attributes = sysutils.MergeInterfaceMaps(
					prevAuthnUserAttrs, engineCtx.AuthenticatedUser.Attributes)
			}
		}

		// Append user ID as a runtime data if not already set
		if engineCtx.AuthenticatedUser.UserID != "" {
			userID := engineCtx.RuntimeData["userID"]
			if userID == "" {
				if engineCtx.RuntimeData == nil {
					engineCtx.RuntimeData = make(map[string]string)
				}
				engineCtx.RuntimeData["userID"] = engineCtx.AuthenticatedUser.UserID
			}
		}
	}
}

// processNodeResponse processes the node response and determines the next action.
// Returns:
// - The next node to execute.
// - Whether to continue execution.
// - Any service error.
func (fe *flowEngine) processNodeResponse(ctx *EngineContext, currentNode core.NodeInterface,
	nodeResp *common.NodeResponse, flowStep *FlowStep, logger *log.Logger) (
	core.NodeInterface, bool, *serviceerror.ServiceError) {
	if nodeResp.Status == "" {
		logger.Error("Node response status not found in the flow graph")
		return nil, false, &serviceerror.InternalServerError
	}

	switch nodeResp.Status {
	case common.NodeStatusComplete:
		nextNode, svcErr := fe.handleCompletedResponse(ctx, currentNode, nodeResp, logger)
		if svcErr != nil {
			return nil, false, svcErr
		}
		return nextNode, true, nil
	case common.NodeStatusIncomplete:
		svcErr := fe.handleIncompleteResponse(nodeResp, flowStep, logger)
		if svcErr != nil {
			return nil, false, svcErr
		}
		return nil, false, nil
	case common.NodeStatusFailure:
		flowStep.Status = common.FlowStatusError
		flowStep.FailureReason = nodeResp.FailureReason
		return nil, false, nil
	default:
		logger.Error("Unsupported response status returned from the node",
			log.String("status", string(nodeResp.Status)))
		return nil, false, &serviceerror.InternalServerError
	}
}

// handleCompletedResponse handles the completed node and returns the next node to execute.
func (fe *flowEngine) handleCompletedResponse(ctx *EngineContext,
	currentNode core.NodeInterface, nodeResp *common.NodeResponse, logger *log.Logger) (
	core.NodeInterface, *serviceerror.ServiceError) {
	nextNode, err := fe.resolveToNextNode(ctx.Graph, currentNode, nodeResp)
	if err != nil {
		logger.Error("Error moving to the next node", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	ctx.CurrentNode = nextNode
	return nextNode, nil
}

// handleIncompleteResponse handles the node response when the status is incomplete.
// It resolves the flow step details based on the type of node response. The same node will be executed again
// in the next request with the required data.
func (fe *flowEngine) handleIncompleteResponse(nodeResp *common.NodeResponse, flowStep *FlowStep,
	logger *log.Logger) *serviceerror.ServiceError {
	switch nodeResp.Type {
	case common.NodeResponseTypeRedirection:
		err := fe.resolveStepForRedirection(nodeResp, flowStep)
		if err != nil {
			logger.Error("Error while resolving step for redirection", log.Error(err))
			return &serviceerror.InternalServerError
		}
		return nil
	case common.NodeResponseTypeView:
		err := fe.resolveStepDetailsForPrompt(nodeResp, flowStep)
		if err != nil {
			logger.Error("Error while resolving step details for prompt", log.Error(err))
			return &serviceerror.InternalServerError
		}
		return nil
	default:
		logger.Error("Unsupported response type returned from the node",
			log.String("responseType", string(nodeResp.Type)))
		return &serviceerror.InternalServerError
	}
	// TODO: Handle retry scenarios with nodeResp.Type == common.NodeResponseTypeRetry
}

// resolveToNextNode resolves the next node to execute based on the current node.
func (fe *flowEngine) resolveToNextNode(graph core.GraphInterface, currentNode core.NodeInterface,
	nodeResp *common.NodeResponse) (core.NodeInterface, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowEngine"))

	nextNodeID := ""
	if currentNode.GetType() == common.NodeTypeDecision {
		logger.Debug("Current node is a decision node. Trying to resolve next node based on decision.")
		if nodeResp == nil || nodeResp.NextNodeID == "" {
			logger.Debug("No next node ID found in the node response. Returning nil.")
			return nil, nil
		}
		nextNodeID = nodeResp.NextNodeID
	} else {
		// Set the first element of the next node list assuming only decision nodes can have multiple next nodes.
		if len(currentNode.GetNextNodeList()) == 0 {
			logger.Debug("No next node found in the current node. Returning nil.")
			return nil, nil
		}
		nextNodeID = currentNode.GetNextNodeList()[0]
	}
	if nextNodeID == "" {
		logger.Debug("No next node found. Returning nil.")
		return nil, nil
	}

	nextNode, ok := graph.GetNode(nextNodeID)
	if !ok {
		return nil, errors.New("next node not found in the graph")
	}

	logger.Debug("Moving to next node", log.String("nextNodeID", nextNode.GetID()))
	return nextNode, nil
}

// resolveStepForRedirection resolves the flow step details for a redirection response.
func (fe *flowEngine) resolveStepForRedirection(nodeResp *common.NodeResponse, flowStep *FlowStep) error {
	if nodeResp == nil {
		return errors.New("node response is nil")
	}
	if nodeResp.RedirectURL == "" {
		return errors.New("redirect URL not found in the node response")
	}

	if flowStep.Data.AdditionalData == nil {
		flowStep.Data.AdditionalData = make(map[string]string)
		flowStep.Data.AdditionalData = nodeResp.AdditionalData
	} else {
		// Append to the existing additional info
		for key, value := range nodeResp.AdditionalData {
			flowStep.Data.AdditionalData[key] = value
		}
	}

	flowStep.Data.RedirectURL = nodeResp.RedirectURL

	if flowStep.Data.Inputs == nil {
		flowStep.Data.Inputs = make([]common.InputData, 0)
		flowStep.Data.Inputs = nodeResp.RequiredData
	} else {
		// Append to the existing input data
		flowStep.Data.Inputs = append(flowStep.Data.Inputs, nodeResp.RequiredData...)
	}

	flowStep.Status = common.FlowStatusIncomplete
	flowStep.Type = common.StepTypeRedirection
	return nil
}

// resolveStepDetailsForPrompt resolves the step details for a user prompt response.
func (fe *flowEngine) resolveStepDetailsForPrompt(nodeResp *common.NodeResponse, flowStep *FlowStep) error {
	if nodeResp == nil {
		return errors.New("node response is nil")
	}
	if len(nodeResp.RequiredData) == 0 && len(nodeResp.Actions) == 0 {
		return errors.New("no required data or actions found in the node response")
	}

	if len(nodeResp.RequiredData) > 0 {
		if flowStep.Data.Inputs == nil {
			flowStep.Data.Inputs = make([]common.InputData, 0)
			flowStep.Data.Inputs = nodeResp.RequiredData
		} else {
			// Append to the existing input data
			flowStep.Data.Inputs = append(flowStep.Data.Inputs, nodeResp.RequiredData...)
		}
	}

	if len(nodeResp.Actions) > 0 {
		if flowStep.Data.Actions == nil {
			flowStep.Data.Actions = make([]common.Action, 0)
		}
		flowStep.Data.Actions = nodeResp.Actions
	}

	flowStep.Status = common.FlowStatusIncomplete
	flowStep.Type = common.StepTypeView
	return nil
}

// recordNodeExecution adds or updates execution record for the node.
func recordNodeExecution(ctx *EngineContext, node core.NodeInterface, nodeResp *common.NodeResponse,
	nodeErr *serviceerror.ServiceError, executionStartTime int64, executionEndTime int64) {
	nodeID := node.GetID()
	record := ctx.ExecutionHistory[nodeID]

	// Create new record if it does not exist
	if record == nil {
		nextStep := len(ctx.ExecutionHistory) + 1
		newRecord := createExecutionRecord(node, nextStep)
		ctx.ExecutionHistory[nodeID] = &newRecord
		record = &newRecord
	}

	attempt := createExecutionAttempt(record, nodeResp, nodeErr, executionStartTime, executionEndTime)
	record.Executions = append(record.Executions, attempt)

	record.Status = attempt.Status
	record.EndTime = attempt.EndTime
}

// createExecutionRecord creates a new node execution record.
func createExecutionRecord(node core.NodeInterface, step int) common.NodeExecutionRecord {
	record := common.NodeExecutionRecord{
		NodeID:     node.GetID(),
		NodeType:   string(node.GetType()),
		Step:       step,
		Status:     common.FlowStatusIncomplete,
		Executions: make([]common.ExecutionAttempt, 0),
		StartTime:  time.Now().Unix(),
	}

	// Set executor details if applicable (only for executor-backed nodes)
	if node.GetType() == common.NodeTypeTaskExecution {
		if executableNode, ok := node.(core.ExecutorBackedNodeInterface); ok {
			executor := executableNode.GetExecutor()
			if executor != nil {
				record.ExecutorName = executor.GetName()
				record.ExecutorType = executor.GetType()
			}
		}
	}

	return record
}

// createExecutionAttempt creates a new execution attempt.
func createExecutionAttempt(nodeRecord *common.NodeExecutionRecord, nodeResp *common.NodeResponse,
	nodeErr *serviceerror.ServiceError, executionStartTime int64, executionEndTime int64) common.ExecutionAttempt {
	attempt := common.ExecutionAttempt{
		Attempt:   len(nodeRecord.Executions) + 1,
		Timestamp: executionEndTime,
		StartTime: executionStartTime,
		EndTime:   executionEndTime,
	}

	// Determine status
	if nodeErr != nil {
		attempt.Status = common.FlowStatusError
	} else if nodeResp != nil {
		switch nodeResp.Status {
		case common.NodeStatusComplete:
			attempt.Status = common.FlowStatusComplete
		case common.NodeStatusIncomplete:
			attempt.Status = common.FlowStatusIncomplete
		case common.NodeStatusFailure:
			attempt.Status = common.FlowStatusError
		default:
			attempt.Status = common.FlowStatusIncomplete
		}
	}

	return attempt
}
