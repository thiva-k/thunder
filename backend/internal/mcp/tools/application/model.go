/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

// Package application provides application management tool models.
//
//nolint:lll
package application

import (
	"github.com/asgardeo/thunder/internal/application/model"
)

// applicationListOutput represents the output for list_applications tool.
type applicationListOutput struct {
	TotalCount   int                              `json:"total_count" jsonschema:"Total number of applications available."`
	Applications []model.BasicApplicationResponse `json:"applications" jsonschema:"List of applications."`
}

// clientIDInput represents input for client ID-based lookups.
type clientIDInput struct {
	ClientID string `json:"client_id" jsonschema:"OAuth client ID to search for"`
}
