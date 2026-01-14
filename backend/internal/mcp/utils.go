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

	"github.com/google/jsonschema-go/jsonschema"
)

// EnumDefinition defines the metadata for an enum type.
type EnumDefinition struct {
	Type        reflect.Type
	Values      []string
	Description string
}

// GenerateTypeSchemas creates a map of custom schemas for the given enums.
func GenerateTypeSchemas(enums ...EnumDefinition) map[reflect.Type]*jsonschema.Schema {
	schemas := make(map[reflect.Type]*jsonschema.Schema)
	for _, enum := range enums {
		schemas[enum.Type] = &jsonschema.Schema{
			Type:        "string",
			Enum:        stringSliceToAny(enum.Values),
			Description: enum.Description,
		}
	}
	return schemas
}

// stringSliceToAny converts a string slice to an any slice.
func stringSliceToAny(s []string) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}
