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

// Package dao provides the dao layer for managing flow graphs.
package dao

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	authnmodel "github.com/asgardeo/thunder/internal/authn/model"
	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/jsonmodel"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/flow/utils"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/database/client"
	dbmodel "github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/database/provider"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

var (
	instance *FlowDAO
	once     sync.Once
)

// FlowDAOInterface defines the interface for the flow data access object.
type FlowDAOInterface interface {
	Init() error
	RegisterGraph(graphID string, g model.GraphInterface)
	GetGraph(graphID string) (model.GraphInterface, bool)
	IsValidGraphID(graphID string) bool
	GetContextFromStore(flowID string) (model.EngineContext, bool)
	StoreContextInStore(flowID string, context model.EngineContext) error
	RemoveContextFromStore(flowID string) error
}

// FlowDAO is the implementation of FlowDAOInterface.
type FlowDAO struct {
	graphs   map[string]model.GraphInterface
	ctxStore map[string]model.EngineContext
	mu       sync.Mutex
	dbClient client.DBClientInterface // runtime DB client
}

// GetFlowDAO returns a singleton instance of FlowDAOInterface.
func GetFlowDAO() FlowDAOInterface {
	once.Do(func() {
		var dbClient client.DBClientInterface
		var err error
		dbClient, err = provider.NewDBProvider().GetDBClient("runtime")
		logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowDAO"))
		if err != nil {
			logger.Error("Failed to connect to runtime DB", log.Error(err))
		} else {
			logger.Info("Successfully connected to runtime DB")
		}
		instance = &FlowDAO{
			graphs:   make(map[string]model.GraphInterface),
			ctxStore: make(map[string]model.EngineContext),
			mu:       sync.Mutex{},
			dbClient: dbClient,
		}
	})
	return instance
}

// Init initializes the FlowDAO by loading graph configurations into runtime.
func (c *FlowDAO) Init() error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowDAO"))
	logger.Debug("Initializing the flow DAO layer")

	configDir := config.GetThunderRuntime().Config.Flow.GraphDirectory
	if configDir == "" {
		logger.Info("Graph directory is not set. No graphs will be loaded.")
		return nil
	}

	configDir = filepath.Join(config.GetThunderRuntime().ThunderHome, configDir)
	configDir = filepath.Clean(configDir)

	logger.Debug("Loading graphs from config directory", log.String("configDir", configDir))

	files, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Config directory does not exist. No graphs will be loaded.",
				log.String("configDir", configDir))
			return nil
		}
		return fmt.Errorf("failed to read config directory %s: %w", configDir, err)
	}

	if len(files) == 0 {
		logger.Info("No graph configuration files found in the configured directory. No graphs will be loaded.")
		return nil
	}
	logger.Debug("Found graph definition files in the graph directory", log.Int("fileCount", len(files)))

	// Process each JSON file in the directory
	flowGraphs := make(map[string]model.GraphInterface)
	for _, file := range files {
		// Skip directories and non-JSON files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			logger.Debug("Skipping non-JSON file or directory",
				log.String("fileName", file.Name()), log.Bool("isDir", file.IsDir()))
			continue
		}
		filePath := filepath.Join(configDir, file.Name())
		filePath = filepath.Clean(filePath)

		// Read the file content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("Failed to read graph file", log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Parse the JSON into the flow model
		var jsonGraph jsonmodel.GraphDefinition
		if err := json.Unmarshal(fileContent, &jsonGraph); err != nil {
			logger.Warn("Failed to parse JSON in file", log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Convert the JSON graph definition to the graph model
		graphModel, err := utils.BuildGraphFromDefinition(&jsonGraph)
		if err != nil {
			logger.Warn("Failed to convert graph definition to graph model",
				log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Log the graph model as JSON for debugging
		if logger.IsDebugEnabled() {
			jsonString, err := graphModel.ToJSON()
			if err != nil {
				logger.Warn("Failed to convert graph model to JSON", log.String("filePath", filePath), log.Error(err))
			} else {
				logger.Debug("Graph model loaded successfully", log.String("graphID", graphModel.GetID()),
					log.String("json", jsonString))
			}
		}

		// Append graph to the flowGraphs map
		flowGraphs[graphModel.GetID()] = graphModel
	}

	// Register all loaded graphs
	inferredGraphCount := 0
	for graphID, graph := range flowGraphs {
		// Create and register the equivalent registration graph if not found already.
		registrationGraphID := c.getRegistrationGraphID(graphID)
		_, exists := c.graphs[registrationGraphID]
		if !exists && graph.GetType() == constants.FlowTypeAuthentication {
			if err := c.createAndRegisterRegistrationGraph(registrationGraphID, graph, logger); err != nil {
				logger.Error("Failed creating registration graph", log.String("graphID", graphID), log.Error(err))
				continue
			}
			inferredGraphCount++
		}

		logger.Debug("Registering graph", log.String("graphType", string(graph.GetType())),
			log.String("graphID", graphID))
		c.RegisterGraph(graphID, graph)

		// Persist the graph to the runtime DB
		if err := c.persistGraphToDB(graph, logger); err != nil {
			logger.Warn("Failed to persist graph to DB", log.String("graphID", graphID), log.Error(err))
		}
	}

	logger.Debug("Flow DAO initialized successfully", log.Int("configuredGraphCount", len(flowGraphs)),
		log.Int("inferredGraphCount", inferredGraphCount))

	return nil
}

// RegisterGraph registers a graph with the FlowDAO by its ID.
func (c *FlowDAO) RegisterGraph(graphID string, g model.GraphInterface) {
	c.graphs[graphID] = g
}

// GetGraph retrieves a graph by its ID
func (c *FlowDAO) GetGraph(graphID string) (model.GraphInterface, bool) {
	g, ok := c.graphs[graphID]
	return g, ok
}

// IsValidGraphID checks if the provided graph ID is valid and exists in the DAO.
func (c *FlowDAO) IsValidGraphID(graphID string) bool {
	if graphID == "" {
		return false
	}
	_, exists := c.graphs[graphID]
	return exists
}

// GetContextFromStore retrieves the flow context from the store based on the flow ID.
func (c *FlowDAO) GetContextFromStore(flowID string) (model.EngineContext, bool) {

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowDAO"), log.String("flowID", flowID))
	logger.Info("Executing GetContextFromStore")
	// Query flow_context
	query := dbmodel.DBQuery{ID: "select_flow_context", Query: "SELECT flow_id, flow_type, app_id, current_node_id, current_action_id, graph_id, is_authenticated, authenticated_user_id, user_input_data, runtime_data, authenticated_user_attributes FROM flow_context WHERE flow_id = ?"}
	rows, err := c.dbClient.Query(query, flowID)
	if err != nil || len(rows) == 0 {
		logger.Info("No flow context found in DB", log.Error(err))
		return model.EngineContext{}, false
	}
	row := rows[0]
	logger.Info("Fetched flow context from DB")
	// Extract and deserialize
	flowIDVal := row["flow_id"].(string)
	flowType := row["flow_type"].(string)
	appID := row["app_id"].(string)
	currentNodeID := row["current_node_id"].(string)
	currentActionID := ""
	if row["current_action_id"] != nil {
		currentActionID = row["current_action_id"].(string)
	}
	graphID := row["graph_id"].(string)
	isAuthenticated := false
	if row["is_authenticated"] != nil {
		switch v := row["is_authenticated"].(type) {
		case bool:
			isAuthenticated = v
		case int64:
			isAuthenticated = v != 0
		}
	}
	authenticatedUserID := ""
	if row["authenticated_user_id"] != nil {
		authenticatedUserID = row["authenticated_user_id"].(string)
	}
	userInputData := map[string]string{}
	if row["user_input_data"] != nil {
		_ = json.Unmarshal([]byte(row["user_input_data"].(string)), &userInputData)
	}
	runtimeData := map[string]string{}
	if row["runtime_data"] != nil {
		_ = json.Unmarshal([]byte(row["runtime_data"].(string)), &runtimeData)
	}
	authUserAttrs := map[string]string{}
	if row["authenticated_user_attributes"] != nil {
		_ = json.Unmarshal([]byte(row["authenticated_user_attributes"].(string)), &authUserAttrs)
	}
	// Query flow_current_node_response
	query2 := dbmodel.DBQuery{ID: "select_node_response", Query: "SELECT status, type, failure_reason, redirect_url, next_node_id, assertion, authenticated_user_id, is_authenticated, required_data, additional_data, actions, runtime_data, authenticated_user_attributes FROM flow_current_node_response WHERE flow_id = ?"}
	rows2, err := c.dbClient.Query(query2, flowID)
	var nodeResp *model.NodeResponse
	if err == nil && len(rows2) > 0 {
		row2 := rows2[0]
		status := row2["status"].(string)
		typ := row2["type"].(string)
		failureReason := ""
		if row2["failure_reason"] != nil {
			failureReason = row2["failure_reason"].(string)
		}
		redirectURL := ""
		if row2["redirect_url"] != nil {
			redirectURL = row2["redirect_url"].(string)
		}
		nextNodeID := ""
		if row2["next_node_id"] != nil {
			nextNodeID = row2["next_node_id"].(string)
		}
		assertion := ""
		if row2["assertion"] != nil {
			assertion = row2["assertion"].(string)
		}
		nodeRespAuthUserID := ""
		if row2["authenticated_user_id"] != nil {
			nodeRespAuthUserID = row2["authenticated_user_id"].(string)
		}
		nodeRespIsAuth := false
		if row2["is_authenticated"] != nil {
			switch v := row2["is_authenticated"].(type) {
			case bool:
				nodeRespIsAuth = v
			case int64:
				nodeRespIsAuth = v != 0
			}
		}
		requiredData := []model.InputData{}
		if row2["required_data"] != nil {
			_ = json.Unmarshal([]byte(row2["required_data"].(string)), &requiredData)
		}
		additionalData := map[string]string{}
		if row2["additional_data"] != nil {
			_ = json.Unmarshal([]byte(row2["additional_data"].(string)), &additionalData)
		}
		actions := []model.Action{}
		if row2["actions"] != nil {
			_ = json.Unmarshal([]byte(row2["actions"].(string)), &actions)
		}
		nodeRespRuntimeData := map[string]string{}
		if row2["runtime_data"] != nil {
			_ = json.Unmarshal([]byte(row2["runtime_data"].(string)), &nodeRespRuntimeData)
		}
		nodeRespAuthUserAttrs := map[string]string{}
		if row2["authenticated_user_attributes"] != nil {
			_ = json.Unmarshal([]byte(row2["authenticated_user_attributes"].(string)), &nodeRespAuthUserAttrs)
		}
		nodeRespAuthUser := authnmodel.AuthenticatedUser{
			IsAuthenticated: nodeRespIsAuth,
			UserID:          nodeRespAuthUserID,
			Attributes:      nodeRespAuthUserAttrs,
		}
		nodeResp = &model.NodeResponse{
			Status:            constants.NodeStatus(status),
			Type:              constants.NodeResponseType(typ),
			FailureReason:     failureReason,
			RequiredData:      requiredData,
			AdditionalData:    additionalData,
			RedirectURL:       redirectURL,
			Actions:           actions,
			NextNodeID:        nextNodeID,
			RuntimeData:       nodeRespRuntimeData,
			AuthenticatedUser: nodeRespAuthUser,
			Assertion:         assertion,
		}
		logger.Info("Fetched node response from DB")
	}
	// Reconstruct AuthenticatedUser
	authUser := authnmodel.AuthenticatedUser{
		IsAuthenticated: isAuthenticated,
		UserID:          authenticatedUserID,
		Attributes:      authUserAttrs,
	}
	// Reconstruct Graph and CurrentNode
	graph, node := c.reconstructGraphAndNode(graphID, currentNodeID, logger)
	logger.Info("Reconstructed graph and current node from DB")

	engineCtx := model.EngineContext{
		FlowID:              flowIDVal,
		FlowType:            constants.FlowType(flowType),
		AppID:               appID,
		UserInputData:       userInputData,
		RuntimeData:         runtimeData,
		CurrentNode:         node,
		CurrentNodeResponse: nodeResp,
		CurrentActionID:     currentActionID,
		Graph:               graph,
		AuthenticatedUser:   authUser,
	}
	logger.Info("Returning EngineContext from GetContextFromStore", log.Any("engineContext", engineCtx))
	return engineCtx, true
}

func (c *FlowDAO) StoreContextInStore(flowID string, context model.EngineContext) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowDAO"), log.String("flowID", flowID))
	logger.Info("Executing StoreContextInStore")
	if flowID == "" {
		return fmt.Errorf("flow ID cannot be empty")
	}
	logger.Info("About to upsert EngineContext in flow_context", log.Any("engineContext", context))
	// Serialize maps
	userInputDataJSON, _ := json.Marshal(context.UserInputData)
	runtimeDataJSON, _ := json.Marshal(context.RuntimeData)
	authUserAttrsJSON, _ := json.Marshal(context.AuthenticatedUser.Attributes)
	isAuthenticated := context.AuthenticatedUser.IsAuthenticated
	authenticatedUserID := context.AuthenticatedUser.UserID
	// Upsert into flow_context
	upsertQuery := dbmodel.DBQuery{ID: "upsert_flow_context", Query: `INSERT INTO flow_context (flow_id, flow_type, app_id, current_node_id, current_action_id, graph_id, is_authenticated, authenticated_user_id, user_input_data, runtime_data, authenticated_user_attributes, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT(flow_id) DO UPDATE SET flow_type=excluded.flow_type, app_id=excluded.app_id, current_node_id=excluded.current_node_id, current_action_id=excluded.current_action_id, graph_id=excluded.graph_id, is_authenticated=excluded.is_authenticated, authenticated_user_id=excluded.authenticated_user_id, user_input_data=excluded.user_input_data, runtime_data=excluded.runtime_data, authenticated_user_attributes=excluded.authenticated_user_attributes, updated_at=CURRENT_TIMESTAMP`}
	_, err := c.dbClient.Execute(upsertQuery, flowID, string(context.FlowType), context.AppID, c.getNodeID(context.CurrentNode), context.CurrentActionID, c.getGraphID(context.Graph), isAuthenticated, authenticatedUserID, string(userInputDataJSON), string(runtimeDataJSON), string(authUserAttrsJSON))
	if err != nil {
		logger.Info("Failed to upsert flow context in DB", log.Error(err))
		return err
	}
	logger.Info("Upserted flow context in DB")
	// Upsert node response if present
	if context.CurrentNodeResponse != nil {
		logger.Info("About to upsert NodeResponse in flow_current_node_response", log.Any("nodeResponse", context.CurrentNodeResponse))
		requiredDataJSON, _ := json.Marshal(context.CurrentNodeResponse.RequiredData)
		additionalDataJSON, _ := json.Marshal(context.CurrentNodeResponse.AdditionalData)
		actionsJSON, _ := json.Marshal(context.CurrentNodeResponse.Actions)
		nodeRespRuntimeDataJSON, _ := json.Marshal(context.CurrentNodeResponse.RuntimeData)
		nodeRespAuthUserAttrsJSON, _ := json.Marshal(context.CurrentNodeResponse.AuthenticatedUser.Attributes)
		nodeRespIsAuth := context.CurrentNodeResponse.AuthenticatedUser.IsAuthenticated
		upsertNodeRespQuery := dbmodel.DBQuery{ID: "upsert_node_response", Query: `INSERT INTO flow_current_node_response (flow_id, status, type, failure_reason, redirect_url, next_node_id, assertion, authenticated_user_id, is_authenticated, required_data, additional_data, actions, runtime_data, authenticated_user_attributes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(flow_id) DO UPDATE SET status=excluded.status, type=excluded.type, failure_reason=excluded.failure_reason, redirect_url=excluded.redirect_url, next_node_id=excluded.next_node_id, assertion=excluded.assertion, authenticated_user_id=excluded.authenticated_user_id, is_authenticated=excluded.is_authenticated, required_data=excluded.required_data, additional_data=excluded.additional_data, actions=excluded.actions, runtime_data=excluded.runtime_data, authenticated_user_attributes=excluded.authenticated_user_attributes`}
		_, err := c.dbClient.Execute(upsertNodeRespQuery, flowID, string(context.CurrentNodeResponse.Status), string(context.CurrentNodeResponse.Type), context.CurrentNodeResponse.FailureReason, context.CurrentNodeResponse.RedirectURL, context.CurrentNodeResponse.NextNodeID, context.CurrentNodeResponse.Assertion, context.CurrentNodeResponse.AuthenticatedUser.UserID, nodeRespIsAuth, string(requiredDataJSON), string(additionalDataJSON), string(actionsJSON), string(nodeRespRuntimeDataJSON), string(nodeRespAuthUserAttrsJSON))
		if err != nil {
			logger.Info("Failed to upsert node response in DB", log.Error(err))
			return err
		}
		logger.Info("Upserted node response in DB")
	}
	return nil
}

func (c *FlowDAO) RemoveContextFromStore(flowID string) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowDAO"), log.String("flowID", flowID))
	logger.Info("Executing DeleteContextFromStore")
	if flowID == "" {
		return fmt.Errorf("flow ID cannot be empty")
	}
	deleteNodeRespQuery := dbmodel.DBQuery{ID: "delete_node_response", Query: "DELETE FROM flow_current_node_response WHERE flow_id = ?"}
	_, err := c.dbClient.Execute(deleteNodeRespQuery, flowID)
	if err != nil {
		logger.Info("Failed to delete node response from DB", log.Error(err))
		return err
	}
	logger.Info("Deleted node response from DB")
	deleteFlowContextQuery := dbmodel.DBQuery{ID: "delete_flow_context", Query: "DELETE FROM flow_context WHERE flow_id = ?"}
	_, err = c.dbClient.Execute(deleteFlowContextQuery, flowID)
	if err != nil {
		logger.Info("Failed to delete flow context from DB", log.Error(err))
		return err
	}
	logger.Info("Deleted flow context from DB")
	return nil
}

// getRegistrationGraphID constructs the registration graph ID from the auth graph ID.
func (c *FlowDAO) getRegistrationGraphID(authGraphID string) string {
	return constants.RegistrationFlowGraphPrefix + strings.TrimPrefix(authGraphID, constants.AuthFlowGraphPrefix)
}

// createAndRegisterRegistrationGraph creates a registration graph from an authentication graph and registers it.
func (c *FlowDAO) createAndRegisterRegistrationGraph(registrationGraphID string, authGraph model.GraphInterface,
	logger *log.Logger) error {
	registrationGraph, err := c.createRegistrationGraph(registrationGraphID, authGraph)
	if err != nil {
		return fmt.Errorf("failed to infer registration graph: %w", err)
	}

	if logger.IsDebugEnabled() {
		registrationGraphJSON, err := registrationGraph.ToJSON()
		if err != nil {
			logger.Warn("Failed to convert graph model to JSON", log.String("graphID", registrationGraphID),
				log.Error(err))
		} else {
			logger.Debug("Graph model loaded successfully", log.String("graphID", registrationGraph.GetID()),
				log.String("json", registrationGraphJSON))
		}
	}

	logger.Debug("Registering inferred registration graph", log.String("graphID", registrationGraph.GetID()))
	c.RegisterGraph(registrationGraph.GetID(), registrationGraph)
	return nil
}

// createRegistrationGraph creates a registration graph from an authentication graph.
func (c *FlowDAO) createRegistrationGraph(registrationGraphID string,
	authGraph model.GraphInterface) (model.GraphInterface, error) {
	// Create a new graph from the authentication graph
	registrationGraph := model.NewGraph(registrationGraphID, constants.FlowTypeRegistration)

	nodesCopy, err := sysutils.DeepCopyMapOfClonables(authGraph.GetNodes())
	if err != nil {
		return nil, fmt.Errorf("failed to deep copy nodes from auth graph: %w", err)
	}
	registrationGraph.SetNodes(nodesCopy)
	registrationGraph.SetEdges(sysutils.DeepCopyMapOfStringSlices(authGraph.GetEdges()))

	err = registrationGraph.SetStartNode(authGraph.GetStartNodeID())
	if err != nil {
		return nil, fmt.Errorf("failed to set start node for registration graph: %w", err)
	}

	// Find authentication success nodes to insert provisioning before them
	authSuccessNodeID := ""
	nodes := registrationGraph.GetNodes()
	for nodeID, node := range nodes {
		if node.IsFinalNode() {
			authSuccessNodeID = nodeID
			break
		}
	}
	if authSuccessNodeID == "" {
		return nil, fmt.Errorf("no authentication success node found in the authentication graph")
	}

	// Create and add provisioning node
	provisioningNode, err := c.createProvisioningNode()
	if err != nil {
		return nil, fmt.Errorf("failed to create provisioning node: %w", err)
	}
	err = registrationGraph.AddNode(provisioningNode)
	if err != nil {
		return nil, fmt.Errorf("failed to add provisioning node to registration graph: %w", err)
	}

	// Modify the edges that lead to the auth success node to point to the provisioning node
	for fromNodeID, toNodeIDs := range registrationGraph.GetEdges() {
		for _, toNodeID := range toNodeIDs {
			if toNodeID == authSuccessNodeID {
				err := registrationGraph.RemoveEdge(fromNodeID, toNodeID)
				if err != nil {
					return nil, fmt.Errorf("failed to remove edge from %s to %s: %w", fromNodeID, toNodeID, err)
				}

				err = registrationGraph.AddEdge(fromNodeID, provisioningNode.GetID())
				if err != nil {
					return nil, fmt.Errorf("failed to add edge from %s to provisioning node: %w", fromNodeID, err)
				}
			}
		}
	}

	// Add an edge from the provisioning node to the auth success node
	err = registrationGraph.AddEdge(provisioningNode.GetID(), authSuccessNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to add edge from provisioning node to auth success node: %w", err)
	}

	return registrationGraph, nil
}

// createProvisioningNode creates a provisioning node that leads to the specified auth success node
func (c *FlowDAO) createProvisioningNode() (model.NodeInterface, error) {
	provisioningNode, err := model.NewNode(
		"provisioning",
		string(constants.NodeTypeTaskExecution),
		false,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create provisioning node: %w", err)
	}

	execConfig := &model.ExecutorConfig{
		Name:       "ProvisioningExecutor",
		Properties: make(map[string]string),
	}
	provisioningNode.SetExecutorConfig(execConfig)

	return provisioningNode, nil
}

// persistGraphToDB persists a graph to the runtime DB (graphs table) using upsert logic.
func (c *FlowDAO) persistGraphToDB(graph model.GraphInterface, logger *log.Logger) error {
	nodes := graph.GetNodes()
	nodesMap := make(map[string]interface{})
	for id, node := range nodes {
		if node == nil {
			continue
		}
		nodeJSON, err := json.Marshal(node)
		if err != nil {
			logger.Warn("Failed to marshal node for DB persistence", log.String("nodeID", id), log.Error(err))
			continue
		}
		nodesMap[id] = json.RawMessage(nodeJSON)
	}
	nodesJSON, _ := json.Marshal(nodesMap)
	edgesJSON, _ := json.Marshal(graph.GetEdges())
	upsertQuery := dbmodel.DBQuery{ID: "upsert_graph", Query: `INSERT INTO graphs (graph_id, nodes, edges, start_node_id, type) VALUES (?, ?, ?, ?, ?) ON CONFLICT(graph_id) DO UPDATE SET nodes=excluded.nodes, edges=excluded.edges, start_node_id=excluded.start_node_id, type=excluded.type`}
	_, err := c.dbClient.Execute(upsertQuery, graph.GetID(), string(nodesJSON), string(edgesJSON), graph.GetStartNodeID(), string(graph.GetType()))
	if err != nil {
		logger.Warn("Failed to upsert graph in DB", log.String("graphID", graph.GetID()), log.Error(err))
		return err
	}
	logger.Info("Upserted graph in DB", log.String("graphID", graph.GetID()))
	return nil
}

// Helper to reconstruct graph and current node from DB
func (c *FlowDAO) reconstructGraphAndNode(graphID, currentNodeID string, logger *log.Logger) (model.GraphInterface, model.NodeInterface) {
	const defaultGraphReconstructErr = "Failed to reconstruct graph or node from DB"
	query := dbmodel.DBQuery{ID: "select_graph", Query: "SELECT nodes, edges, start_node_id, type FROM graphs WHERE graph_id = ?"}
	rows, err := c.dbClient.Query(query, graphID)
	if err != nil || len(rows) == 0 {
		logger.Info(defaultGraphReconstructErr, log.Error(err))
		return nil, nil
	}
	row := rows[0]
	nodesJSON := row["nodes"].(string)
	edgesJSON := row["edges"].(string)
	startNodeID := row["start_node_id"].(string)
	graphType := row["type"].(string)
	nodesMap := map[string]json.RawMessage{}
	if len(nodesJSON) > 0 {
		_ = json.Unmarshal([]byte(nodesJSON), &nodesMap)
	}
	edgesMap := map[string][]string{}
	if len(edgesJSON) > 0 {
		_ = json.Unmarshal([]byte(edgesJSON), &edgesMap)
	}
	graph := model.NewGraph(graphID, constants.FlowType(graphType))
	for _, nodeRaw := range nodesMap {
		node, err := unmarshalNode(nodeRaw)
		if err == nil {
			graph.AddNode(node)
		} else {
			logger.Warn("Failed to unmarshal node from DB", log.Error(err))
		}
	}
	graph.SetEdges(edgesMap)
	err = graph.SetStartNode(startNodeID)
	if err != nil {
		logger.Error("Failed to set start node in reconstructed graph", log.String("startNodeID", startNodeID), log.Error(err))
	}
	node, _ := graph.GetNode(currentNodeID)
	return graph, node
}

// Local helper to unmarshal a node from JSON
func unmarshalNode(data json.RawMessage) (model.NodeInterface, error) {
	// First, unmarshal just the type
	var typeProbe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeProbe); err != nil {
		return nil, err
	}
	switch typeProbe.Type {
	case string(constants.NodeTypeTaskExecution), string(constants.NodeTypeAuthSuccess):
		var n model.TaskExecutionNode
		if err := json.Unmarshal(data, &n); err != nil {
			return nil, err
		}
		return &n, nil
	case string(constants.NodeTypeDecision):
		var n model.DecisionNode
		if err := json.Unmarshal(data, &n); err != nil {
			return nil, err
		}
		return &n, nil
	case string(constants.NodeTypePromptOnly):
		var n model.PromptOnlyNode
		if err := json.Unmarshal(data, &n); err != nil {
			return nil, err
		}
		return &n, nil
	default:
		// fallback to base node
		var n model.Node
		if err := json.Unmarshal(data, &n); err != nil {
			return nil, err
		}
		return &n, nil
	}
}

// Helper to get node ID from NodeInterface
func (c *FlowDAO) getNodeID(node model.NodeInterface) string {
	if node == nil {
		return ""
	}
	return node.GetID()
}

// Helper to get graph ID from GraphInterface
func (c *FlowDAO) getGraphID(graph model.GraphInterface) string {
	if graph == nil {
		return ""
	}
	return graph.GetID()
}
