// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// Package common provides the RFC 9396 authorization details logic shared by management and runtime.
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/thunder-id/thunderid/internal/system/valuecompare"
)

// TemplatePlaceholder matches a {{field.path}} placeholder in a consent template.
var TemplatePlaceholder = regexp.MustCompile(`\{\{\s*([^{}\s]+)\s*\}\}`)

const (
	// FieldType is the member every authorization detail carries, naming its type.
	FieldType = "type"
	// schemaTypeObject is the JSON Schema type of an authorization details object.
	schemaTypeObject = "object"
	// schemaDraft202012 is the only JSON Schema dialect a type may declare.
	schemaDraft202012 = "https://json-schema.org/draft/2020-12/schema"
)

// jsonSchemaTypes are the JSON Schema type names a type's schema may use.
var jsonSchemaTypes = []string{"array", "boolean", "number", "object", "string"}

// unsupportedTypes are the JSON Schema type names a type's schema may not use, with the reason each is refused.
var unsupportedTypes = map[string]string{
	"integer": "type integer is not supported; use number",
	"null":    "type null is not supported",
}

// Keywords a type's schema may use; others are refused so that no keyword is silently ignored.
var (
	commonKeywords = []string{"type", "title", "description", "enum", "const"}
	typeKeywords   = map[string][]string{
		"object":  {"properties", "required"},
		"string":  {"minLength", "maxLength", "pattern", "format"},
		"number":  {"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf"},
		"boolean": {},
		"array":   {"items", "minItems", "maxItems", "uniqueItems"},
	}
	// assertedFormats are the string formats a schema may declare.
	assertedFormats = []string{"date", "date-time"}
)

// CompileSchema checks a registered schema and compiles it for validating details of its type.
// Every object is closed to undeclared fields, as RFC 9396 section 5 requires.
func CompileSchema(raw json.RawMessage) (*jsonschema.Resolved, error) {
	if len(raw) == 0 {
		return nil, errors.New("schema is required")
	}
	if err := checkKeywords(raw); err != nil {
		return nil, err
	}
	root, err := parseSchema(raw)
	if err != nil {
		return nil, err
	}
	if root.Schema != "" && root.Schema != schemaDraft202012 {
		return nil, fmt.Errorf("schema must use JSON Schema draft 2020-12")
	}
	if !describesObject(root) {
		return nil, errors.New("schema must describe an object")
	}
	if _, err := policiesOf(root); err != nil {
		return nil, err
	}

	if root.Properties == nil {
		root.Properties = map[string]*jsonschema.Schema{}
	}
	if _, ok := root.Properties[FieldType]; !ok {
		root.Properties[FieldType] = &jsonschema.Schema{Type: "string"}
	}
	if !slices.Contains(root.Required, FieldType) {
		root.Required = append(root.Required, FieldType)
	}
	closeObjects(root)

	resolved, err := root.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("schema is not valid JSON Schema 2020-12: %w", err)
	}
	return resolved, nil
}

// describesObject reports whether a schema describes an object.
func describesObject(s *jsonschema.Schema) bool {
	return s.Type == schemaTypeObject
}

// closeObjects forbids undeclared properties on every object schema in the tree.
func closeObjects(s *jsonschema.Schema) {
	if s == nil {
		return
	}
	if s.Type == schemaTypeObject {
		s.AdditionalProperties = &jsonschema.Schema{Not: &jsonschema.Schema{}}
	}
	for _, child := range s.Properties {
		closeObjects(child)
	}
	closeObjects(s.Items)
}

// parseSchema reads a registered schema.
func parseSchema(raw json.RawMessage) (*jsonschema.Schema, error) {
	var root jsonschema.Schema
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("schema is not valid JSON Schema: %w", err)
	}
	return &root, nil
}

// walkProperties calls visit for every nested property, parents first, stopping at the first error.
func walkProperties(schema *jsonschema.Schema, prefix string,
	visit func(path, name string, property *jsonschema.Schema) error) error {
	for name, property := range schema.Properties {
		if property == nil {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		if err := visit(path, name, property); err != nil {
			return err
		}
		if err := walkProperties(property, path, visit); err != nil {
			return err
		}
	}
	return nil
}

// SchemaFieldPaths returns the dot-separated path of every property of a registered schema.
func SchemaFieldPaths(raw json.RawMessage) map[string]bool {
	paths := map[string]bool{FieldType: true}
	root, err := parseSchema(raw)
	if err != nil {
		return paths
	}
	_ = walkProperties(root, "", func(path, _ string, _ *jsonschema.Schema) error {
		paths[path] = true
		return nil
	})
	return paths
}

// joinPath appends a property name to a dot-separated path.
func joinPath(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

// describePath names a dot-separated path in an error, the root being the schema itself.
func describePath(path string) string {
	if path == "" {
		return "the schema"
	}
	return path
}

// checkKeywords checks that a registered schema uses only the supported keywords.
func checkKeywords(raw json.RawMessage) error {
	decoded, err := decodeJSON(raw)
	if err != nil {
		return fmt.Errorf("schema is not valid JSON Schema: %w", err)
	}
	root, ok := decoded.(map[string]any)
	if !ok {
		return errors.New("schema must describe an object")
	}
	if err := checkSchemaKeywords(root, "", true, false, false); err != nil {
		return err
	}
	return checkTypeMember(root)
}

// checkNumbers refuses a number no double can hold, naming the field it is in.
func checkNumbers(value any) error {
	switch v := value.(type) {
	case json.Number:
		if err := valuecompare.Validate(v.String(), valuecompare.TypeNumber); err != nil {
			return fmt.Errorf("the number %s is out of range", v.String())
		}
	case map[string]any:
		for _, member := range v {
			if err := checkNumbers(member); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range v {
			if err := checkNumbers(item); err != nil {
				return err
			}
		}
	}
	return nil
}

// limitPairs are the keywords that bound a value from below and above.
var limitPairs = [][2]string{
	{"minimum", "maximum"}, {"exclusiveMinimum", "exclusiveMaximum"},
	{"minLength", "maxLength"}, {"minItems", "maxItems"},
}

// countKeywords are the keywords that count characters or items, so cannot be negative.
var countKeywords = []string{"minLength", "maxLength", "minItems", "maxItems"}

// checkLimits refuses limits no value can satisfy: a lower bound above its upper bound, a negative count, and
// a multipleOf that is not positive.
func checkLimits(schema map[string]any) error {
	for _, pair := range limitPairs {
		low, hasLow := schema[pair[0]].(json.Number)
		high, hasHigh := schema[pair[1]].(json.Number)
		if !hasLow || !hasHigh {
			continue
		}
		if order, ok := numberOrder(low, high); ok && order > 0 {
			return fmt.Errorf("%s cannot be greater than %s", pair[0], pair[1])
		}
	}
	for _, keyword := range countKeywords {
		if count, ok := schema[keyword].(json.Number); ok {
			if order, ok := numberOrder(count, "0"); ok && order < 0 {
				return fmt.Errorf("%s cannot be negative", keyword)
			}
		}
	}
	if divisor, ok := schema["multipleOf"].(json.Number); ok {
		if order, ok := numberOrder(divisor, "0"); ok && order <= 0 {
			return errors.New("multipleOf must be greater than 0")
		}
	}
	return nil
}

// numberOrder compares two JSON numbers exactly, reporting false when either cannot be compared.
func numberOrder(a, b json.Number) (int, bool) {
	order, err := valuecompare.Compare(a.String(), b.String(), valuecompare.TypeNumber)
	return order, err == nil
}

// CheckTypeValue checks that a schema allows the type value it is registered under: a const or enum on its
// type member must include it, or no detail of the type could be valid.
func CheckTypeValue(raw json.RawMessage, typ string) error {
	decoded, err := decodeJSON(raw)
	if err != nil {
		return err
	}
	root, _ := decoded.(map[string]any)
	properties, _ := root["properties"].(map[string]any)
	member, ok := properties[FieldType].(map[string]any)
	if !ok {
		return nil
	}
	values, hasEnum := member["enum"].([]any)
	c, hasConst := member["const"]
	if hasConst {
		values = append(values, c)
	}
	if (hasEnum || hasConst) && !slices.Contains(values, any(typ)) {
		return fieldError(FieldType, "type allows no detail of the type %q; its const or enum must include it", typ)
	}
	return nil
}

// checkTypeMember checks that a declared type member is a plain string.
func checkTypeMember(root map[string]any) error {
	properties, _ := root["properties"].(map[string]any)
	member, ok := properties[FieldType].(map[string]any)
	if !ok {
		return nil
	}
	if t, declared := member["type"]; declared && t != "string" {
		return fieldError(FieldType, "type must be a string")
	}
	if _, hasFormat := member["format"]; hasFormat {
		return fieldError(FieldType, "type must be a plain string, without a format")
	}
	values, hasEnum := member["enum"].([]any)
	if c, hasConst := member["const"]; hasConst {
		values = append(values, c)
	}
	if !hasEnum && member["type"] == nil && member["const"] == nil {
		return nil
	}
	for _, value := range values {
		if _, isString := value.(string); !isString {
			return fieldError(FieldType, "type must be a string")
		}
	}
	return nil
}

// checkSchemaKeywords checks a schema and, recursively, its properties and items. inItems marks a schema
// under a list's items, where no property can carry a policy.
func checkSchemaKeywords(schema map[string]any, path string, root, property, inItems bool) error {
	types, err := schemaTypes(schema["type"])
	if err != nil {
		return fieldError(path, "%s: %w", describePath(path), err)
	}
	// Every list item must constrain its value, so a list without items is refused.
	if _, hasItems := schema["items"]; slices.Contains(types, "array") && !hasItems {
		return fieldError(path, "%s is a list without items; give it an items schema", describePath(path))
	}
	if err := checkAllowedKeywords(schema, path, types, root, property); err != nil {
		return err
	}
	if err := checkValues(schema, path, len(types) > 0 || root); err != nil {
		return err
	}
	if err := checkRequired(schema, path); err != nil {
		return err
	}
	return checkChildren(schema, path, inItems)
}

// checkAllowedKeywords checks that a schema uses only the keywords supported for its type and place.
func checkAllowedKeywords(schema map[string]any, path string, types []string, root, property bool) error {
	allowed := slices.Clone(commonKeywords)
	if root {
		allowed = append(allowed, "$schema")
	}
	// A policy is read only from a property, so anywhere else it would be ignored.
	if property {
		allowed = append(allowed, policyKeyword)
	}
	for _, t := range types {
		allowed = append(allowed, typeKeywords[t]...)
	}
	for _, keyword := range slices.Sorted(maps.Keys(schema)) {
		if !slices.Contains(allowed, keyword) {
			return fieldError(path, "%s uses %q, which a type's schema does not support", describePath(path), keyword)
		}
	}
	return nil
}

// checkValues checks the values of a schema's own keywords.
func checkValues(schema map[string]any, path string, typed bool) error {
	for keyword, value := range schema {
		if keyword == "properties" || keyword == "items" {
			continue
		}
		if err := checkNumbers(value); err != nil {
			return fieldError(path, "%s: %w", describePath(path), err)
		}
	}
	if err := checkLimits(schema); err != nil {
		return fieldError(path, "%s: %w", describePath(path), err)
	}
	if title, ok := schema["title"]; ok && title == "" {
		return fieldError(path, "%s has an empty title; remove it or give it text", describePath(path))
	}
	_, hasConst := schema["const"]
	enum, hasEnum := schema["enum"]
	if !typed && !hasConst && !hasEnum {
		return fieldError(path, "%s does not constrain its value; give it a type, const or enum", describePath(path))
	}
	if hasEnum {
		if err := checkEnum(enum); err != nil {
			return fieldError(path, "%s: %w", describePath(path), err)
		}
	}
	if format, ok := schema["format"]; ok {
		if name, isString := format.(string); !isString || !slices.Contains(assertedFormats, name) {
			return fieldError(path, "%s has an unsupported format %v; use date or date-time",
				describePath(path), format)
		}
	}
	return nil
}

// checkChildren checks the schemas of an object's properties and of a list's items.
func checkChildren(schema map[string]any, path string, inItems bool) error {
	if properties, ok := schema["properties"]; ok {
		members, isObject := properties.(map[string]any)
		if !isObject {
			return fieldError(path, "%s: properties must be an object", describePath(path))
		}
		for _, name := range slices.Sorted(maps.Keys(members)) {
			child := joinPath(path, name)
			if name != strings.TrimSpace(name) {
				return fieldError(child, "the property name %q starts or ends with a space", name)
			}
			member, isSchema := members[name].(map[string]any)
			if !isSchema {
				return fieldError(child, "%s must be a schema object", describePath(child))
			}
			// Policies are read only through nested properties, so none is allowed under a list's items.
			if err := checkSchemaKeywords(member, child, false, !inItems, inItems); err != nil {
				return err
			}
		}
	}
	if items, ok := schema["items"]; ok {
		member, isSchema := items.(map[string]any)
		if !isSchema {
			return fieldError(path, "%s: items must be a schema object", describePath(path))
		}
		return checkSchemaKeywords(member, path, false, false, true)
	}
	return nil
}

// checkRequired checks that each name an object's required list holds is a property it declares.
func checkRequired(schema map[string]any, path string) error {
	required, ok := schema["required"]
	if !ok {
		return nil
	}
	names, isList := required.([]any)
	if !isList {
		return fieldError(path, "%s: required must be a list of property names", describePath(path))
	}
	properties, _ := schema["properties"].(map[string]any)
	for _, name := range names {
		text, isString := name.(string)
		if !isString {
			return fieldError(path, "%s: required must be a list of property names", describePath(path))
		}
		if _, declared := properties[text]; !declared && (path != "" || text != FieldType) {
			return fieldError(path, "%s requires %q, which it does not declare", describePath(path), text)
		}
	}
	return nil
}

// checkEnum checks that an enum is non-empty, has no duplicates and no values padded with spaces.
func checkEnum(value any) error {
	values, ok := value.([]any)
	if !ok || len(values) == 0 {
		return errors.New("enum must list at least one value")
	}
	for _, v := range values {
		if text, isString := v.(string); isString && text != strings.TrimSpace(text) {
			return fmt.Errorf("the enum value %q starts or ends with a space", text)
		}
	}
	for i := range values {
		for j := i + 1; j < len(values); j++ {
			if valuesEqual(values[i], values[j]) {
				return errors.New("enum must not list a value twice")
			}
		}
	}
	return nil
}

// schemaTypes reads a schema's type keyword, which must name a single type.
func schemaTypes(value any) ([]string, error) {
	switch t := value.(type) {
	case nil:
		return nil, nil
	case string:
		if reason, unsupported := unsupportedTypes[t]; unsupported {
			return nil, errors.New(reason)
		}
		if !slices.Contains(jsonSchemaTypes, t) {
			return nil, fmt.Errorf("unknown type %q", t)
		}
		return []string{t}, nil
	default:
		return nil, errors.New("type must name one JSON Schema type")
	}
}
