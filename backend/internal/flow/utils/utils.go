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

// Package utils provides utility functions for flow processing.
package utils

import (
	"fmt"

	"github.com/asgardeo/thunder/internal/executor/attributecollect"
	"github.com/asgardeo/thunder/internal/executor/authassert"
	"github.com/asgardeo/thunder/internal/executor/basicauth"
	"github.com/asgardeo/thunder/internal/executor/githubauth"
	"github.com/asgardeo/thunder/internal/executor/googleauth"
	"github.com/asgardeo/thunder/internal/executor/provision"
	"github.com/asgardeo/thunder/internal/executor/smsauth"
	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/jsonmodel"
	"github.com/asgardeo/thunder/internal/flow/model"
	idpmodel "github.com/asgardeo/thunder/internal/idp/model"
	idpservice "github.com/asgardeo/thunder/internal/idp/service"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// BuildGraphFromDefinition builds a graph from a graph definition json.
func BuildGraphFromDefinition(definition *jsonmodel.GraphDefinition) (model.GraphInterface, error) {
	if definition == nil || len(definition.Nodes) == 0 {
		return nil, fmt.Errorf("graph definition is nil or has no nodes")
	}

	// Create a graph
	_type, err := getGraphType(definition.Type)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving graph type: %w", err)
	}
	g := model.NewGraph(definition.ID, _type)

	// Add all nodes to the graph
	edges := make(map[string][]string)
	for _, nodeDef := range definition.Nodes {
		isFinalNode := len(nodeDef.Next) == 0

		// Construct a new node. Here we set isStartNode to false by default.
		node, err := model.NewNode(nodeDef.ID, nodeDef.Type, false, isFinalNode)
		if err != nil {
			return nil, fmt.Errorf("failed to create node %s: %w", nodeDef.ID, err)
		}

		// Set next nodes if defined
		if len(nodeDef.Next) > 0 {
			node.SetNextNodeList(nodeDef.Next)

			// Store edges based on the node definition
			_, exists := edges[nodeDef.ID]
			if !exists {
				edges[nodeDef.ID] = []string{}
			}
			edges[nodeDef.ID] = append(edges[nodeDef.ID], nodeDef.Next...)
		}

		// Convert and set input data from definition
		inputData := make([]model.InputData, len(nodeDef.InputData))
		for i, input := range nodeDef.InputData {
			inputData[i] = model.InputData{
				Name:     input.Name,
				Type:     input.Type,
				Required: input.Required,
			}
		}
		node.SetInputData(inputData)

		// Set the executor config if defined
		if nodeDef.Executor.Name != "" {
			executor, err := getExecutorConfigByName(nodeDef.Executor)
			if err != nil {
				return nil, fmt.Errorf("error while getting executor %s: %w", nodeDef.Executor, err)
			}
			node.SetExecutorConfig(executor)
		} else if nodeDef.Type == string(constants.NodeTypeAuthSuccess) {
			executor, err := getExecutorConfigByName(jsonmodel.ExecutorDefinition{
				Name: "AuthAssertExecutor",
			})
			if err != nil {
				return nil, fmt.Errorf("error while getting default AuthAssertExecutor: %w", err)
			}
			node.SetExecutorConfig(executor)
		}

		err = g.AddNode(node)
		if err != nil {
			return nil, fmt.Errorf("failed to add node %s to the graph: %w", nodeDef.ID, err)
		}
	}

	// Set edges in the graph
	for sourceID, targetIDs := range edges {
		for _, targetID := range targetIDs {
			err := g.AddEdge(sourceID, targetID)
			if err != nil {
				return nil, fmt.Errorf("failed to add edge from %s to %s: %w", sourceID, targetID, err)
			}
		}
	}

	// Determine the start node and set it in the graph
	startNodeID := ""
	for _, node := range g.GetNodes() {
		if len(node.GetPreviousNodeList()) == 0 {
			startNodeID = node.GetID()
			break
		}
	}
	if startNodeID == "" {
		return nil, fmt.Errorf("no start node found in the graph definition")
	}

	err = g.SetStartNode(startNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to set start node ID: %w", err)
	}

	return g, nil
}

// getGraphType retrieves the graph type from a string representation.
func getGraphType(graphType string) (constants.FlowType, error) {
	switch graphType {
	case string(constants.FlowTypeAuthentication):
		return constants.FlowTypeAuthentication, nil
	case string(constants.FlowTypeRegistration):
		return constants.FlowTypeRegistration, nil
	default:
		return "", fmt.Errorf("unsupported graph type: %s", graphType)
	}
}

// getExecutorConfigByName constructs an executor configuration by its definition if it exists.
func getExecutorConfigByName(execDef jsonmodel.ExecutorDefinition) (*model.ExecutorConfig, error) {
	if execDef.Name == "" {
		return nil, fmt.Errorf("executor name cannot be empty")
	}

	// At this point, we assume executors and attached IDPs are already registered in the system.
	// Hence validations will not be done at this point.
	var executor model.ExecutorConfig
	switch execDef.Name {
	case "BasicAuthExecutor":
		executor = model.ExecutorConfig{
			Name:    "BasicAuthExecutor",
			IdpName: "Local",
		}
	case "SMSOTPAuthExecutor":
		executor = model.ExecutorConfig{
			Name:       "SMSOTPAuthExecutor",
			IdpName:    "Local",
			Properties: execDef.Properties,
		}
	case "GithubOAuthExecutor":
		executor = model.ExecutorConfig{
			Name:       "GithubOAuthExecutor",
			IdpName:    execDef.IdpName,
			Properties: execDef.Properties,
		}
	case "GoogleOIDCAuthExecutor":
		executor = model.ExecutorConfig{
			Name:       "GoogleOIDCAuthExecutor",
			IdpName:    execDef.IdpName,
			Properties: execDef.Properties,
		}
	case "AttributeCollector":
		executor = model.ExecutorConfig{
			Name:       "AttributeCollector",
			Properties: execDef.Properties,
		}
	case "ProvisioningExecutor":
		executor = model.ExecutorConfig{
			Name:       "ProvisioningExecutor",
			Properties: execDef.Properties,
		}
	case "AuthAssertExecutor":
		executor = model.ExecutorConfig{
			Name: "AuthAssertExecutor",
		}
	default:
		return nil, fmt.Errorf("executor with name %s not found", execDef.Name)
	}

	if executor.Name == "" {
		return nil, fmt.Errorf("executor with name %s could not be created", execDef.Name)
	}

	return &executor, nil
}

// GetExecutorByName constructs an executor by its definition.
func GetExecutorByName(execConfig *model.ExecutorConfig) (model.ExecutorInterface, error) {
	if execConfig == nil {
		return nil, fmt.Errorf("executor configuration cannot be nil")
	}
	if execConfig.Name == "" {
		return nil, fmt.Errorf("executor name cannot be empty")
	}

	var executor model.ExecutorInterface
	switch execConfig.Name {
	case "BasicAuthExecutor":
		idp, err := getIDP("Local")
		if err != nil {
			return nil, fmt.Errorf("error while getting IDP for BasicAuthExecutor: %w", err)
		}
		executor = basicauth.NewBasicAuthExecutor(idp.ID, idp.Name, execConfig.Properties)
	case "SMSOTPAuthExecutor":
		idp, err := getIDP("Local")
		if err != nil {
			return nil, fmt.Errorf("error while getting IDP for SMSOTPAuthExecutor: %w", err)
		}

		if len(execConfig.Properties) == 0 {
			return nil, fmt.Errorf("properties for SMSOTPAuthExecutor cannot be empty")
		}
		senderName, exists := execConfig.Properties["senderName"]
		if !exists || senderName == "" {
			return nil, fmt.Errorf("senderName property is required for SMSOTPAuthExecutor")
		}
		executor = smsauth.NewSMSOTPAuthExecutor(idp.ID, idp.Name, execConfig.Properties)
	case "GithubOAuthExecutor":
		idp, err := getIDP(execConfig.IdpName)
		if err != nil {
			return nil, fmt.Errorf("error while getting IDP for GithubOAuthExecutor: %w", err)
		}

		clientID, clientSecret, redirectURI, scopes, additionalParams, err := getIDPConfigs(
			idp.Properties, execConfig)
		if err != nil {
			return nil, err
		}

		executor = githubauth.NewGithubOAuthExecutor(idp.ID, idp.Name, execConfig.Properties,
			clientID, clientSecret, redirectURI, scopes, additionalParams)
	case "GoogleOIDCAuthExecutor":
		idp, err := getIDP(execConfig.IdpName)
		if err != nil {
			return nil, fmt.Errorf("error while getting IDP for GoogleOIDCAuthExecutor: %w", err)
		}

		clientID, clientSecret, redirectURI, scopes, additionalParams, err := getIDPConfigs(
			idp.Properties, execConfig)
		if err != nil {
			return nil, err
		}

		executor = googleauth.NewGoogleOIDCAuthExecutor(idp.ID, idp.Name, execConfig.Properties,
			clientID, clientSecret, redirectURI, scopes, additionalParams)
	case "AttributeCollector":
		executor = attributecollect.NewAttributeCollector("attribute-collector", "AttributeCollector",
			execConfig.Properties)
	case "ProvisioningExecutor":
		executor = provision.NewProvisioningExecutor("provisioning-executor", "ProvisioningExecutor",
			execConfig.Properties)
	case "AuthAssertExecutor":
		executor = authassert.NewAuthAssertExecutor("auth-assert-executor", "AuthAssertExecutor",
			execConfig.Properties)
	default:
		return nil, fmt.Errorf("executor with name %s not found", execConfig.Name)
	}

	if executor == nil {
		return nil, fmt.Errorf("executor with name %s could not be created", execConfig.Name)
	}
	return executor, nil
}

// getIDP retrieves the IDP by its name. Returns an error if the IDP does not exist or if the name is empty.
func getIDP(idpName string) (*idpmodel.IDP, error) {
	if idpName == "" {
		return nil, fmt.Errorf("IDP name cannot be empty")
	}

	idpSvc := idpservice.GetIDPService()
	idp, err := idpSvc.GetIdentityProviderByName(idpName)
	if err != nil {
		return nil, fmt.Errorf("error while getting IDP with the name %s: %w", idpName, err)
	}
	if idp == nil {
		return nil, fmt.Errorf("IDP with name %s does not exist", idpName)
	}

	return idp, nil
}

// getIDPConfigs retrieves the IDP configurations for a given executor configuration.
func getIDPConfigs(idpProperties []idpmodel.IDPProperty, execConfig *model.ExecutorConfig) (string,
	string, string, []string, map[string]string, error) {
	if len(idpProperties) == 0 {
		return "", "", "", nil, nil, fmt.Errorf("IDP properties not found for executor with IDP name %s",
			execConfig.IdpName)
	}
	var clientID, clientSecret, redirectURI, scopesStr string
	additionalParams := map[string]string{}
	for _, prop := range idpProperties {
		switch prop.Name {
		case "client_id":
			clientID = prop.Value
		case "client_secret":
			clientSecret = prop.Value
		case "redirect_uri":
			redirectURI = prop.Value
		case "scopes":
			scopesStr = prop.Value
		default:
			additionalParams[prop.Name] = prop.Value
		}
	}
	if clientID == "" || clientSecret == "" || redirectURI == "" || scopesStr == "" {
		return "", "", "", nil, nil, fmt.Errorf("missing required properties for executor with IDP name %s",
			execConfig.IdpName)
	}
	scopes := sysutils.ParseStringArray(scopesStr, ",")

	return clientID, clientSecret, redirectURI, scopes, additionalParams, nil
}
