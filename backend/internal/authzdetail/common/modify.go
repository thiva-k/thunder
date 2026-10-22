// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/thunder-id/thunderid/internal/system/valuecompare"
)

// decodeJSON decodes a value keeping numbers as their literal text, so that no precision is lost.
func decodeJSON(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var v any
	if err := decoder.Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

// SchemaEquivalent reports whether two registered schemas define the same type, ignoring order and defaults.
func SchemaEquivalent(a, b json.RawMessage) bool {
	av, errA := decodeJSON(a)
	bv, errB := decodeJSON(b)
	if errA != nil || errB != nil {
		return false
	}
	return valuesEqual(canonicalSchema(av, true), canonicalSchema(bv, true))
}

// canonicalSchema returns a decoded schema in the form SchemaEquivalent compares.
func canonicalSchema(value any, root bool) any {
	schema, ok := value.(map[string]any)
	if !ok {
		return value
	}
	out := make(map[string]any, len(schema))
	for key, member := range schema {
		switch key {
		case "properties":
			properties, isObject := member.(map[string]any)
			if !isObject {
				out[key] = member
				continue
			}
			if len(properties) == 0 {
				continue
			}
			canonical := make(map[string]any, len(properties))
			for name, property := range properties {
				canonical[name] = canonicalSchema(property, false)
			}
			out[key] = canonical
		case "items":
			out[key] = canonicalSchema(member, false)
		case "required":
			names := sortedNames(member)
			if root {
				names = slices.DeleteFunc(names, func(name any) bool { return name == FieldType })
			}
			if len(names) > 0 {
				out[key] = names
			}
		case policyKeyword:
			if policy := canonicalPolicy(member); len(policy) > 0 {
				out[key] = policy
			}
		default:
			out[key] = member
		}
	}
	return out
}

// canonicalPolicy returns a decoded policy without default settings and with its parties sorted.
func canonicalPolicy(value any) map[string]any {
	policy, ok := value.(map[string]any)
	if !ok {
		return map[string]any{"": value}
	}
	out := make(map[string]any, len(policy))
	for key, setting := range policy {
		switch {
		case key == "suppliedBy" && setting == PartyClient, key == "modification" && setting == modificationNone:
		case key == "modifiableBy":
			out[key] = sortedNames(setting)
		default:
			out[key] = setting
		}
	}
	return out
}

// sortedNames returns a decoded list of strings sorted and without repeats; any other value is kept.
func sortedNames(value any) []any {
	items, ok := value.([]any)
	if !ok {
		return []any{value}
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		name, isString := item.(string)
		if !isString {
			return items
		}
		names = append(names, name)
	}
	slices.Sort(names)
	names = slices.Compact(names)
	out := make([]any, len(names))
	for i, name := range names {
		out[i] = name
	}
	return out
}

// valuesEqual compares two decoded JSON values, treating numbers by value.
func valuesEqual(a, b any) bool {
	switch av := a.(type) {
	case json.Number:
		bv, ok := b.(json.Number)
		if !ok {
			return false
		}
		if equal, err := valuecompare.Equal(av.String(), bv.String(), valuecompare.TypeNumber); err == nil {
			return equal
		}
		return av == bv
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !valuesEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, v := range av {
			if w, ok := bv[k]; !ok || !valuesEqual(v, w) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}

// compareOrdered compares two scalar values of an ordered value type, returning -1, 0 or 1.
func compareOrdered(a, b json.RawMessage, as string) (int, error) {
	t := valuecompare.Type(as)
	if !t.IsOrdered() {
		return 0, fmt.Errorf("unsupported ordering %q", as)
	}
	aText, err := orderedText(a, t)
	if err != nil {
		return 0, err
	}
	bText, err := orderedText(b, t)
	if err != nil {
		return 0, err
	}
	return valuecompare.Compare(aText, bText, t)
}

// orderedText returns the text of an ordered value, refusing a value of the wrong JSON type.
func orderedText(raw json.RawMessage, t valuecompare.Type) (string, error) {
	v, err := decodeJSON(raw)
	if err != nil {
		return "", err
	}
	switch value := v.(type) {
	case json.Number:
		if t == valuecompare.TypeNumber {
			return value.String(), nil
		}
	case string:
		if t != valuecompare.TypeNumber {
			return value, nil
		}
	}
	return "", fmt.Errorf("value is not a %s", t)
}
