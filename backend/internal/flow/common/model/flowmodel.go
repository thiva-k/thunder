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

// Package model defines the data structures and interfaces for flow execution and graph representation.
package model

import (
	"github.com/asgardeo/thunder/internal/application/model"
	authncm "github.com/asgardeo/thunder/internal/authn/common"
	"github.com/asgardeo/thunder/internal/flow/common"
)

// NodeExecutionRecord represents a record of a node execution in the flow.
type NodeExecutionRecord struct {
	NodeID       string              `json:"node_id"`
	NodeType     string              `json:"node_type"`
	ExecutorName string              `json:"executor_name,omitempty"`
	ExecutorType common.ExecutorType `json:"executor_type,omitempty"`
	Step         int                 `json:"step"`
	Status       common.FlowStatus   `json:"status"`
	Executions   []ExecutionAttempt  `json:"executions"`
	StartTime    int64               `json:"start_time,omitempty"`
	EndTime      int64               `json:"end_time,omitempty"`
}

// GetDuration calculates the duration of the execution in milliseconds.
func (n *NodeExecutionRecord) GetDuration() int64 {
	return getDuration(n.StartTime, n.EndTime)
}

// ExecutionAttempt represents a single execution attempt of a node.
type ExecutionAttempt struct {
	Attempt   int               `json:"attempt"`
	Timestamp int64             `json:"timestamp"`
	Status    common.FlowStatus `json:"status"`
	StartTime int64             `json:"start_time"`
	EndTime   int64             `json:"end_time"`
}

// GetDuration calculates the duration of the execution attempt in milliseconds.
func (e *ExecutionAttempt) GetDuration() int64 {
	return getDuration(e.StartTime, e.EndTime)
}

// getDuration calculates the duration between startTime and endTime in milliseconds.
func getDuration(startTime int64, endTime int64) int64 {
	if startTime == 0 || endTime == 0 {
		return 0
	}
	return (endTime - startTime) * 1000
}

// EngineContext holds the overall context used by the flow engine during execution.
type EngineContext struct {
	FlowID        string
	FlowType      common.FlowType
	AppID         string
	UserInputData map[string]string
	RuntimeData   map[string]string

	CurrentNode         NodeInterface
	CurrentNodeResponse *NodeResponse
	CurrentActionID     string

	Graph       GraphInterface
	Application model.ApplicationProcessedDTO

	AuthenticatedUser authncm.AuthenticatedUser
	ExecutionHistory  map[string]*NodeExecutionRecord
}

// NodeContext holds the context for a specific node in the flow execution.
type NodeContext struct {
	FlowID          string
	FlowType        common.FlowType
	AppID           string
	CurrentActionID string

	NodeProperties map[string]string
	NodeInputData  []InputData
	UserInputData  map[string]string
	RuntimeData    map[string]string

	Application       model.ApplicationProcessedDTO
	AuthenticatedUser authncm.AuthenticatedUser
	ExecutionHistory  map[string]*NodeExecutionRecord
}

// FlowStep represents the outcome of a individual flow step
type FlowStep struct {
	FlowID        string
	StepID        string
	Type          common.FlowStepType
	Status        common.FlowStatus
	Data          FlowData
	Assertion     string
	FailureReason string
}

// FlowData holds the data returned by a flow execution step
type FlowData struct {
	Inputs         []InputData       `json:"inputs,omitempty"`
	RedirectURL    string            `json:"redirectURL,omitempty"`
	Actions        []Action          `json:"actions,omitempty"`
	AdditionalData map[string]string `json:"additionalData,omitempty"`
}

// InputData represents the input data required for a flow step
type InputData struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// Action represents an action to be executed in a flow step
type Action struct {
	Type common.ActionType `json:"type"`
	ID   string            `json:"id"`
	// Executor *ExecutorModel `json:"executor,omitempty"`
}

// ExecutorModel represents an executor configuration within an action
type ExecutorModel struct {
	Name string `json:"name"`
}

// FlowRequest represents the flow execution API request body
type FlowRequest struct {
	ApplicationID string            `json:"applicationId"`
	FlowType      string            `json:"flowType"`
	FlowID        string            `json:"flowId"`
	ActionID      string            `json:"actionId"`
	Inputs        map[string]string `json:"inputs"`
}

// FlowResponse represents the flow execution API response body
type FlowResponse struct {
	FlowID        string   `json:"flowId"`
	StepID        string   `json:"stepId,omitempty"`
	FlowStatus    string   `json:"flowStatus"`
	Type          string   `json:"type,omitempty"`
	Data          FlowData `json:"data,omitempty"`
	Assertion     string   `json:"assertion,omitempty"`
	FailureReason string   `json:"failureReason,omitempty"`
}

// FlowInitContext represents the context for initiating a new flow with runtime data
type FlowInitContext struct {
	ApplicationID string
	FlowType      string
	RuntimeData   map[string]string
}
