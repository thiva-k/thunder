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

// Package tools provides utility functions for MCP schema generation and manipulation.
package tools

import (
	"encoding/json"
	"slices"

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

// WithDefaults applies default values to the generated schema properties.
func WithDefaults(defaults map[string]any) func(*jsonschema.Schema) {
	return func(root *jsonschema.Schema) {
		walkSchema(root, func(s *jsonschema.Schema) {
			for key, value := range defaults {
				if prop, ok := s.Properties[key]; ok {
					if raw, err := json.Marshal(value); err == nil {
						prop.Default = raw
					}
				}
			}
		})
	}
}

// WithEnum applies enum constraints to a property in the schema.
func WithEnum(parent, child string, values []string) func(*jsonschema.Schema) {
	return func(root *jsonschema.Schema) {
		walkSchema(root, func(s *jsonschema.Schema) {
			// Targeted parent of property provided
			if parent != "" {
				if parentProp, ok := s.Properties[parent]; ok {
					var targetProps map[string]*jsonschema.Schema

					// Handle array of objects (lists) or nested objects
					if parentProp.Items != nil && parentProp.Items.Properties != nil {
						targetProps = parentProp.Items.Properties
					} else if parentProp.Properties != nil {
						targetProps = parentProp.Properties
					}

					if targetProps != nil {
						if childProp, ok := targetProps[child]; ok {
							applyEnum(childProp, values)
							targetProps[child] = childProp
						}
					}
				}
			} else {
				// No parent provided
				if prop, ok := s.Properties[child]; ok {
					applyEnum(prop, values)
					s.Properties[child] = prop
				}
			}
		})
	}
}

// WithRequired marks the specified fields as required in the schema.
func WithRequired(fields ...string) func(*jsonschema.Schema) {
	return func(root *jsonschema.Schema) {
		walkSchema(root, func(s *jsonschema.Schema) {
			for _, f := range fields {
				if _, ok := s.Properties[f]; ok && !slices.Contains(s.Required, f) {
					s.Required = append(s.Required, f)
				}
			}
		})
	}
}

// WithRemove removes specified fields from the schema properties.
func WithRemove(fields ...string) func(*jsonschema.Schema) {
	return func(root *jsonschema.Schema) {
		for _, f := range fields {
			delete(root.Properties, f)
			root.Required = slices.DeleteFunc(root.Required, func(r string) bool {
				return r == f
			})
		}
	}
}

// WalkSchema recursively visits the schema and its children, invoking the callback for each.
func walkSchema(root *jsonschema.Schema, visit func(*jsonschema.Schema)) {
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

func stringSliceToAny(strings []string) []any {
	anys := make([]any, len(strings))
	for i, s := range strings {
		anys[i] = s
	}
	return anys
}

// Helper to apply enum values to a schema property
func applyEnum(prop *jsonschema.Schema, values []string) {
	if prop.Items != nil {
		prop.Items.Enum = stringSliceToAny(values)
	} else {
		prop.Enum = stringSliceToAny(values)
	}
}
