// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ModifyTestSuite struct {
	suite.Suite
}

func TestModifyTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyTestSuite))
}

func (s *ModifyTestSuite) TestCompareOrderedUnsupported() {
	_, err := compareOrdered(json.RawMessage(`"1"`), json.RawMessage(`"2"`), "weight")
	s.Error(err)
	_, err = compareOrdered(json.RawMessage(`true`), json.RawMessage(`"2"`), valueTypeNumber)
	s.Error(err)
	_, err = compareOrdered(json.RawMessage(`"soon"`), json.RawMessage(`"2026-01-31"`), valueTypeDate)
	s.Error(err)
	_, err = compareOrdered(json.RawMessage(`"x"`), json.RawMessage(`"y"`), "integer")
	s.Error(err, "whole numbers are ordered as numbers, not as a type of their own")
	_, err = compareOrdered(json.RawMessage(`"1"`), json.RawMessage(`2`), valueTypeNumber)
	s.Error(err, "a number written as text is not a number")
}

func (s *ModifyTestSuite) TestSchemaEquivalent() {
	base := `{"type":"object","required":["type","b","a"],"properties":{"a":{"type":"string"},` +
		`"b":{"type":"number","policy":{"suppliedBy":"client","modification":"decrease",` +
		`"modifiableBy":["user","client"]}},"c":{"type":"object","properties":{}}}}`
	same := `{"type":"object","required":["a","b"],"properties":{"c":{"type":"object"},"b":{"type":"number",` +
		`"policy":{"modification":"decrease","modifiableBy":["client","user","user"]}},"a":{"type":"string"}}}`
	s.True(SchemaEquivalent(json.RawMessage(base), json.RawMessage(same)))

	for name, other := range map[string]string{
		"a field required no more": `{"type":"object","required":["a"],"properties":{"a":{"type":"string"},` +
			`"b":{"type":"number","policy":{"modification":"decrease","modifiableBy":["user","client"]}},` +
			`"c":{"type":"object"}}}`,
		"a party removed": `{"type":"object","required":["a","b"],"properties":{"a":{"type":"string"},` +
			`"b":{"type":"number","policy":{"modification":"decrease","modifiableBy":["user"]}},` +
			`"c":{"type":"object"}}}`,
		"a limit added": `{"type":"object","required":["a","b"],"properties":{"a":{"type":"string","maxLength":5},` +
			`"b":{"type":"number","policy":{"modification":"decrease","modifiableBy":["user","client"]}},` +
			`"c":{"type":"object"}}}`,
	} {
		s.False(SchemaEquivalent(json.RawMessage(base), json.RawMessage(other)), name)
	}
	s.False(SchemaEquivalent(json.RawMessage(`{`), json.RawMessage(`{}`)))
}
