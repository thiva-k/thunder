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

package mcp

import (
	"reflect"

	"github.com/asgardeo/thunder/internal/application/model"
	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/google/jsonschema-go/jsonschema"
)

// getCustomSchemas returns custom JSON Schema definitions for OAuth types.
// This enables enum validation for grant_types, response_types, and token_endpoint_auth_method.
func getCustomSchemas() map[reflect.Type]*jsonschema.Schema {
	return map[reflect.Type]*jsonschema.Schema{
		// GrantType enum - allowed OAuth grant types
		reflect.TypeFor[oauth2const.GrantType](): {
			Type: "string",
			Enum: []any{
				"authorization_code",
				"client_credentials",
				"refresh_token",
				"urn:ietf:params:oauth:grant-type:token-exchange",
			},
			Description: "OAuth grant type",
		},
		// ResponseType enum - allowed OAuth response types
		reflect.TypeFor[oauth2const.ResponseType](): {
			Type:        "string",
			Enum:        []any{"code"},
			Description: "OAuth response type",
		},
		// TokenEndpointAuthMethod enum - allowed token endpoint auth methods
		reflect.TypeFor[oauth2const.TokenEndpointAuthMethod](): {
			Type:        "string",
			Enum:        []any{"client_secret_basic", "client_secret_post", "none"},
			Description: "Token endpoint authentication method",
		},
	}
}

// generateApplicationDTOSchema generates a JSON Schema for ApplicationDTO with enum support.
// Used for create_application tool.
func generateApplicationDTOSchema() *jsonschema.Schema {
	opts := &jsonschema.ForOptions{TypeSchemas: getCustomSchemas()}
	schema, err := jsonschema.For[model.ApplicationDTO](opts)
	if err != nil {
		// Fall back to auto-inference if custom schema fails
		schema, _ = jsonschema.For[model.ApplicationDTO](nil)
	}
	return schema
}

// generateUpdateApplicationDTOSchema generates a JSON Schema for ApplicationDTO with 'id' as required.
// Used for update_application tool.
func generateUpdateApplicationDTOSchema() *jsonschema.Schema {
	schema := generateApplicationDTOSchema()
	// Ensure 'id' is in the required list for update operations
	hasID := false
	for _, r := range schema.Required {
		if r == "id" {
			hasID = true
			break
		}
	}
	if !hasID {
		schema.Required = append([]string{"id"}, schema.Required...)
	}
	return schema
}

// generateNotificationSenderRequestSchema generates a JSON Schema for NotificationSenderRequest.
// This ensures the 'properties' field is correctly typed as an array of objects.
func generateNotificationSenderRequestSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Request to create a notification sender",
		Properties: map[string]*jsonschema.Schema{
			"name": {
				Type:        "string",
				Description: "Name of the notification sender",
			},
			"description": {
				Type:        "string",
				Description: "Description of the notification sender",
			},
			"provider": {
				Type:        "string",
				Description: "Provider type: twilio, vonage, or custom",
				Enum:        []any{"twilio", "vonage", "custom"},
			},
			"properties": {
				Type:        "array",
				Description: "Provider-specific configuration properties",
				Items: &jsonschema.Schema{
					Type:        "object",
					Description: "A configuration property",
					Properties: map[string]*jsonschema.Schema{
						"name": {
							Type:        "string",
							Description: "Property name (e.g., url, http_method, content_type)",
						},
						"value": {
							Type:        "string",
							Description: "Property value",
						},
						"is_secret": {
							Type:        "boolean",
							Description: "Whether this property contains sensitive data",
						},
					},
					Required: []string{"name", "value"},
				},
			},
		},
		Required: []string{"name", "provider"},
	}
}

// generateNotificationSenderUpdateSchema generates a JSON Schema for updating a notification sender.
// Used for update_notification_sender tool - requires 'id' field.
func generateNotificationSenderUpdateSchema() *jsonschema.Schema {
	schema := generateNotificationSenderRequestSchema()
	// Add 'id' field for update
	schema.Properties["id"] = &jsonschema.Schema{
		Type:        "string",
		Description: "The unique identifier of the notification sender to update",
	}
	schema.Required = append([]string{"id"}, schema.Required...)
	schema.Description = "Request to update a notification sender"
	return schema
}
