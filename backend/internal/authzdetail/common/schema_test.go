// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SchemaTestSuite struct {
	suite.Suite
}

func TestSchemaTestSuite(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}

func (s *SchemaTestSuite) TestCompileSchema_Rejects() {
	cases := map[string]string{
		"empty":                ``,
		"invalid JSON":         `{"type":`,
		"not an object schema": `{"type":"string"}`,
		"multi-type schema":    `{"type":["object","null"]}`,
		"no type or props":     `{}`,
		"unsupported draft":    `{"$schema":"http://json-schema.org/draft-04/schema#","type":"object"}`,
		"invalid pattern":      `{"type":"object","properties":{"a":{"type":"string","pattern":"("}}}`,
		"draft 07":             `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object"}`,
		"remote reference":     `{"type":"object","properties":{"a":{"$ref":"https://example.com/s.json"}}}`,
	}
	for name, schema := range cases {
		_, err := CompileSchema(json.RawMessage(schema))
		s.Error(err, name)
	}
}

func (s *SchemaTestSuite) TestCompileSchema_RefusesUnknownTypeNames() {
	for name, schema := range map[string]string{
		"property":   `{"type":"object","properties":{"amount":{"type":"nope"}}}`,
		"nested":     `{"type":"object","properties":{"a":{"type":"object","properties":{"b":{"type":"nmber"}}}}}`,
		"list items": `{"type":"object","properties":{"tags":{"type":"array","items":{"type":"text"}}}}`,
	} {
		_, err := CompileSchema(json.RawMessage(schema))
		s.Require().Error(err, name)
		s.Contains(err.Error(), "unknown type", name)
	}
}

func (s *SchemaTestSuite) TestCompileSchema_RefusesIntegerAndNullTypes() {
	for schema, message := range map[string]string{
		`{"type":"object","properties":{"n":{"type":"integer"}}}`: "n: type integer is not supported; use number",
		`{"type":"object","properties":{"ids":{"type":"array","items":{"type":"integer"}}}}`: "ids: type integer is " +
			"not supported; use number",
		`{"type":"object","properties":{"gone":{"type":"null"}}}`: "gone: type null is not supported",
	} {
		_, err := CompileSchema(json.RawMessage(schema))
		s.Require().Error(err, schema)
		s.Equal(message, err.Error(), schema)
	}
}

func (s *SchemaTestSuite) TestCompileSchema_AcceptsWholeNumbersAsNumberWithMultipleOfOne() {
	_, err := CompileSchema(json.RawMessage(`{"type":"object","properties":{"count":{"type":"number",` +
		`"multipleOf":1,"minimum":1,"policy":{"modification":"decrease","modifiableBy":["user"]}}}}`))
	s.NoError(err)
}

func (s *SchemaTestSuite) TestCompileSchema_RefusesUnsupportedKeywords() {
	cases := map[string]string{
		"a property with no type":   `{"type":"object","properties":{"type":{"type":"string"},"payee":{}}}`,
		"list items with no type":   `{"type":"object","properties":{"tags":{"type":"array","items":{}}}}`,
		"a misspelled keyword":      `{"type":"object","properties":{"a":{"type":"string","maxLenght":3}}}`,
		"a keyword of another type": `{"type":"object","properties":{"a":{"type":"string","minimum":1}}}`,
		"an unasserted format":      `{"type":"object","properties":{"e":{"type":"string","format":"email"}}}`,
		"anyOf": `{"type":"object","properties":{"a":{"type":"string"}},` +
			`"anyOf":[{"required":["a"]}]}`,
		"if and then": `{"type":"object","properties":{"m":{"type":"string"},"b":{"type":"string"}},` +
			`"if":{"properties":{"m":{"const":"swift"}}},"then":{"required":["b"]}}`,
		"a nested oneOf": `{"type":"object","properties":{"a":{"type":"string","oneOf":[{"const":"x"}]}}}`,
		"a reference": `{"type":"object","$defs":{"m":{"type":"string"}},` +
			`"properties":{"a":{"$ref":"#/$defs/m"}}}`,
		"an open object":     `{"type":"object","additionalProperties":true,"properties":{"a":{"type":"string"}}}`,
		"pattern properties": `{"type":"object","patternProperties":{"^x-":{"type":"string"}}}`,
		"a default":          `{"type":"object","properties":{"a":{"type":"string","default":"x"}}}`,
		"a list of types":    `{"type":"object","properties":{"a":{"type":["number","null"]}}}`,
		"a policy at the root": `{"type":"object","policy":{"modification":"none"},` +
			`"properties":{"a":{"type":"string"}}}`,
		"a policy on list items": `{"type":"object","properties":{"a":{"type":"array",` +
			`"items":{"type":"number","policy":{"modification":"none"}}}}}`,
		"a policy under list items": `{"type":"object","properties":{"lines":{"type":"array","items":{` +
			`"type":"object","properties":{"amount":{"type":"number","policy":{"suppliedBy":"processor"}}}}}}}`,
		"a list without items":        `{"type":"object","properties":{"actions":{"type":"array"}}}`,
		"a non-string type member":    `{"type":"object","properties":{"type":{"type":"number"}}}`,
		"a non-string type value":     `{"type":"object","properties":{"type":{"enum":["a",1]}}}`,
		"a type member with a format": `{"type":"object","properties":{"type":{"type":"string","format":"date"}}}`,
		"an empty enum":               `{"type":"object","properties":{"a":{"enum":[]}}}`,
		"a repeated enum value": `{"type":"object","properties":{"a":{"type":"array",` +
			`"items":{"type":"number","enum":[1,1.0]}}}}`,
		"a number out of range":       `{"type":"object","properties":{"a":{"type":"number","maximum":1e400}}}`,
		"an undeclared required name": `{"type":"object","required":["nope"],"properties":{"a":{"type":"string"}}}`,
		"an undeclared nested required name": `{"type":"object","properties":{"o":{"type":"object",` +
			`"required":["zzz"],"properties":{"a":{"type":"string"}}}}}`,
		"an empty title": `{"type":"object","properties":{"a":{"type":"string","title":""}}}`,
		"an empty modifiableBy": `{"type":"object","properties":{"a":{"type":"number",` +
			`"policy":{"modifiableBy":[]}}}}`,
		"a name with surrounding space": `{"type":"object","properties":{" a":{"type":"string"}}}`,
		"an enum value with surrounding space": `{"type":"object","properties":{"a":{"type":"array",` +
			`"items":{"type":"string","enum":[" x"]}}}}`,
		"a date and time the user changes": `{"type":"object","properties":{"at":{"type":"string",` +
			`"format":"date-time","policy":{"modification":"decrease","modifiableBy":["user"]}}}}`,
	}
	for name, raw := range cases {
		_, err := CompileSchema(json.RawMessage(raw))
		s.Error(err, name)
	}
}

func (s *SchemaTestSuite) TestCompileSchema_AcceptsSupportedKeywords() {
	_, err := CompileSchema(json.RawMessage(`{"$schema":"https://json-schema.org/draft/2020-12/schema",` +
		`"type":"object","title":"Payment","description":"A payment","required":["amount"],"properties":{` +
		`"amount":{"type":"number","minimum":0,"exclusiveMaximum":10000,"multipleOf":0.01,` +
		`"policy":{"modification":"decrease","modifiableBy":["user"]}},` +
		`"count":{"type":"number","multipleOf":1,"maximum":5},` +
		`"currency":{"enum":["EUR","USD"]},` +
		`"note":{"type":"string","maxLength":140,"pattern":"^[A-Za-z ]*$"},` +
		`"until":{"type":"string","format":"date"},"at":{"type":"string","format":"date-time"},` +
		`"recurring":{"type":"boolean"},"kind":{"const":"single"},` +
		`"actions":{"type":"array","items":{"type":"string","enum":["pay","status"]},"minItems":1,"maxItems":2,` +
		`"uniqueItems":true},"payee":{"type":"object","required":["iban"],"properties":{"iban":{"type":"string"}}}}}`))
	s.NoError(err)
}

func (s *SchemaTestSuite) TestCompileSchema_AcceptsAListOfObjectsWithoutPolicies() {
	_, err := CompileSchema(json.RawMessage(`{"type":"object","properties":{"lines":{"type":"array",` +
		`"items":{"type":"object","required":["amount"],"properties":{"amount":{"type":"number"},` +
		`"tags":{"type":"array","items":{"type":"string"}}}}}}}`))
	s.NoError(err)
}

func (s *SchemaTestSuite) TestCompileSchema_NamesTheUnsupportedKeywordFirst() {
	_, err := CompileSchema(json.RawMessage(`{"type":"object","properties":{"a":{"anyOf":[{"type":"string"}]}}}`))
	s.Require().Error(err)
	s.Contains(err.Error(), `uses "anyOf"`)

	_, err = CompileSchema(json.RawMessage(`{"type":"object","properties":{"at":{"type":"string",` +
		`"format":"date-time","policy":{"minimum":"2026-01-31"}}}}`))
	s.Require().Error(err)
	s.Contains(err.Error(), "RFC 3339")
}

func (s *SchemaTestSuite) TestCompileSchema_NamesTheFieldOfAnOutOfRangeNumber() {
	_, err := CompileSchema(json.RawMessage(`{"type":"object","properties":{"a":{"type":"number","maximum":1e400}}}`))
	s.Require().Error(err)
	var fieldErr *FieldError
	s.Require().ErrorAs(err, &fieldErr)
	s.Equal("a", fieldErr.Path)
	s.Contains(err.Error(), "out of range")
	s.NotContains(err.Error(), "Go struct")
}

func (s *SchemaTestSuite) TestCompileSchema_AcceptsTheTypeMemberInTheRootRequiredList() {
	_, err := CompileSchema(json.RawMessage(`{"type":"object","required":["type","a"],` +
		`"properties":{"a":{"type":"string","title":"A"}}}`))
	s.NoError(err)
}

// Limits no value can satisfy are refused.
func (s *SchemaTestSuite) TestCompileSchema_RefusesUnsatisfiableLimits() {
	for name, property := range map[string]string{
		"minimum above maximum":     `{"type":"number","minimum":10,"maximum":1}`,
		"exclusive bounds reversed": `{"type":"number","exclusiveMinimum":5,"exclusiveMaximum":2}`,
		"minLength above maxLength": `{"type":"string","minLength":5,"maxLength":2}`,
		"minItems above maxItems":   `{"type":"array","items":{"type":"string"},"minItems":3,"maxItems":1}`,
		"negative length":           `{"type":"string","minLength":-1}`,
		"zero multipleOf":           `{"type":"number","multipleOf":0}`,
		"negative multipleOf":       `{"type":"number","multipleOf":-2}`,
	} {
		_, err := CompileSchema(json.RawMessage(`{"type":"object","properties":{"a":` + property + `}}`))
		s.Require().Error(err, name)
		var fieldErr *FieldError
		s.Require().ErrorAs(err, &fieldErr, name)
		s.Equal("a", fieldErr.Path, name)
	}
	_, err := CompileSchema(json.RawMessage(`{"type":"object","properties":{"a":{"type":"number","minimum":1,` +
		`"maximum":1,"multipleOf":0.5}}}`))
	s.NoError(err, "equal bounds and a positive multipleOf are satisfiable")
}

// A const or enum on the type member must include the value the type is registered under.
func (s *SchemaTestSuite) TestCheckTypeValue() {
	s.NoError(CheckTypeValue(json.RawMessage(`{"type":"object"}`), "payment"))
	s.NoError(CheckTypeValue(json.RawMessage(`{"type":"object","properties":{"type":{"const":"payment"}}}`),
		"payment"))
	s.NoError(CheckTypeValue(json.RawMessage(`{"type":"object","properties":{"type":{"enum":["a","payment"]}}}`),
		"payment"))
	err := CheckTypeValue(json.RawMessage(`{"type":"object","properties":{"type":{"const":"other"}}}`), "payment")
	var fieldErr *FieldError
	s.Require().ErrorAs(err, &fieldErr)
	s.Equal(FieldType, fieldErr.Path)
	s.Error(CheckTypeValue(json.RawMessage(`{"type":"object","properties":{"type":{"enum":["a"]}}}`), "payment"))
	s.Error(CheckTypeValue(json.RawMessage(`{`), "payment"))
}
