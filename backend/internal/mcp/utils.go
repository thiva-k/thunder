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
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
)

// GenerateSchema creates a JSON schema for a given type T and applies the provided modifiers.
func GenerateSchema[T any](modifiers ...func(*jsonschema.Schema)) *jsonschema.Schema {
	schema, _ := jsonschema.For[T](&jsonschema.ForOptions{})

	for _, mod := range modifiers {
		mod(schema)
	}

	return schema
}

// WithDefaults applies default values to the generated schema properties, recursively.
func WithDefaults(defaults map[string]any) func(*jsonschema.Schema) {
	return func(root *jsonschema.Schema) {
		WalkSchema(root, func(s *jsonschema.Schema) {
			for key, value := range defaults {
				if prop, ok := s.Properties[key]; ok {
					raw, err := json.Marshal(value)
					if err == nil {
						prop.Default = raw
						s.Properties[key] = prop
					}
				}
			}
		})
	}
}

// WalkSchema recursively visits the schema and its children, invoking the callback for each.
func WalkSchema(root *jsonschema.Schema, visit func(*jsonschema.Schema)) {
	visited := make(map[*jsonschema.Schema]bool)

	var walker func(*jsonschema.Schema)
	walker = func(s *jsonschema.Schema) {
		if s == nil || visited[s] {
			return
		}
		visited[s] = true

		visit(s)

		// Traverse children
		for _, p := range s.Properties {
			walker(p)
		}
		if s.Items != nil {
			walker(s.Items)
		}
		for _, d := range s.Definitions {
			walker(d)
		}
		// Traverse logical composition
		for _, sub := range s.AllOf {
			walker(sub)
		}
		for _, sub := range s.AnyOf {
			walker(sub)
		}
		for _, sub := range s.OneOf {
			walker(sub)
		}
	}

	walker(root)
}

// WithEnum applies enum constraints to a specific property in the schema, recursively.
func WithEnum(property string, values []string) func(*jsonschema.Schema) {
	return func(root *jsonschema.Schema) {
		WalkSchema(root, func(s *jsonschema.Schema) {
			if prop, ok := s.Properties[property]; ok {
				if prop.Items != nil {
					prop.Items.Enum = stringSliceToAny(values)
				} else {
					prop.Enum = stringSliceToAny(values)
				}
				s.Properties[property] = prop
			}
		})
	}
}

// WithRequired marks the specified fields as required in the schema, recursively.
// It will try to find the field in properties and mark it as required in the parent structure.
func WithRequired(fields ...string) func(*jsonschema.Schema) {
	return func(root *jsonschema.Schema) {
		WalkSchema(root, func(s *jsonschema.Schema) {
			requiredMap := make(map[string]struct{})
			for _, f := range s.Required {
				requiredMap[f] = struct{}{}
			}
			
			// Only add if the field actually exists in this schema's properties
			changed := false
			for _, f := range fields {
				if _, ok := s.Properties[f]; ok {
					if _, exists := requiredMap[f]; !exists {
						requiredMap[f] = struct{}{}
						changed = true
					}
				}
			}

			if changed {
				newRequired := make([]string, 0, len(requiredMap))
				for f := range requiredMap {
					newRequired = append(newRequired, f)
				}
				s.Required = newRequired
			}
		})
	}
}

func stringSliceToAny(strings []string) []any {
	anys := make([]any, len(strings))
	for i, s := range strings {
		anys[i] = s
	}
	return anys
}
