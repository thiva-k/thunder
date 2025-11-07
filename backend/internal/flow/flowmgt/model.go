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

package flowmgt

// graphDefinition represents the direct graph structure from JSON
type graphDefinition struct {
	ID    string           `json:"id"`
	Type  string           `json:"type"`
	Nodes []nodeDefinition `json:"nodes"`
}

// nodeDefinition represents a node in the graph definition
type nodeDefinition struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Properties map[string]string  `json:"properties,omitempty"`
	InputData  []inputDefinition  `json:"inputData"`
	Executor   executorDefinition `json:"executor"`
	Next       []string           `json:"next,omitempty"`
}

// inputDefinition represents an input parameter for a node
type inputDefinition struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// executorDefinition represents the executor configuration for a node
type executorDefinition struct {
	Name string `json:"name"`
}
