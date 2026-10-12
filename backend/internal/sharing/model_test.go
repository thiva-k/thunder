// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package sharing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

// ModelTestSuite covers the wire encodings of the request and export types.
type ModelTestSuite struct {
	suite.Suite
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, new(ModelTestSuite))
}

// members returns a pointer to a list, for the fields where absent and empty differ.
func members(v ...string) *[]string {
	out := append([]string{}, v...)
	return &out
}

// fullRequest exercises every tagged field, so a missing or misspelled tag shows up as a name the
// expectations below do not contain. It is not a request the service would accept: every section is
// stated, with the fields of both modes, because the wire names are all this is here to pin.
func fullRequest() PolicyRequest {
	rules := map[string]OverlayRule{
		"permissions": {
			Value:          members("billing"),
			AllowedValues:  members("billing", "bookings"),
			ExcludedValues: members("billing:refunds"),
		},
	}
	return PolicyRequest{
		ID:             "policy-1",
		InitiatingOUID: "ou-1",
		AllUnits:       &AllUnitsRequest{ExcludedOUIDs: []string{"ou-3"}, OverlayRules: rules},
		ChildUnits: &ChildUnitsRequest{
			Mode:          ChildUnitsSelected,
			ExcludedOUIDs: []string{"ou-4"},
			OverlayRules:  rules,
			Units: []ChildUnitRequest{{
				OUID: "ou-2", Reach: ReachSubtree,
				OverlayRules: map[string]OverlayRule{"assignments": {Editable: true}},
			}},
		},
		OtherTreeRoots: &OtherTreeRootsRequest{
			Mode:          OtherTreeRootsSelected,
			ExcludedOUIDs: []string{"root-2"},
			OverlayRules:  rules,
			Units:         []RootUnitRequest{{OUID: "root-1", OverlayRules: rules}},
		},
		Version: 3,
	}
}

// The declarative and REST surfaces have to agree, so the two encodings must use the same names.
func (s *ModelTestSuite) TestPolicyRequestUsesTheSameCamelCaseNamesInBothEncodings() {
	req := fullRequest()

	asJSON, err := json.Marshal(req)
	s.Require().NoError(err)
	asYAML, err := yaml.Marshal(req)
	s.Require().NoError(err)

	var fromJSON map[string]interface{}
	s.Require().NoError(json.Unmarshal(asJSON, &fromJSON))

	// Compared whole rather than key by key, so a missing tag on a nested struct cannot slip past:
	// yaml lowercases field names of its own when a tag is absent.
	s.Equal(fromJSON, s.normalize(asYAML),
		"the two encodings disagree, so some field is missing a matching tag")

	s.ElementsMatch(
		[]string{"id", "initiatingOuId", "allUnits", "childUnits", "otherTreeRoots", "version"},
		keysOf(fromJSON))

	allUnits, ok := fromJSON["allUnits"].(map[string]interface{})
	s.Require().True(ok, "allUnits is absent, so the field is untagged")
	s.ElementsMatch([]string{"excludedOuIds", "overlayRules"}, keysOf(allUnits))

	childUnits, ok := fromJSON["childUnits"].(map[string]interface{})
	s.Require().True(ok, "childUnits is absent, so the field is untagged")
	s.ElementsMatch([]string{"mode", "excludedOuIds", "overlayRules", "units"}, keysOf(childUnits))
	s.Equal("selected", childUnits["mode"])

	unit, ok := childUnits["units"].([]interface{})[0].(map[string]interface{})
	s.Require().True(ok)
	s.ElementsMatch([]string{"ouId", "reach", "overlayRules"}, keysOf(unit))
	s.Equal("subtree", unit["reach"])

	// The roots of other trees are stated the way child units are: the same modes, and a list
	// called units.
	roots, ok := fromJSON["otherTreeRoots"].(map[string]interface{})
	s.Require().True(ok, "otherTreeRoots is absent, so the field is untagged")
	s.ElementsMatch([]string{"mode", "excludedOuIds", "overlayRules", "units"}, keysOf(roots))
	s.Equal("selected", roots["mode"])

	root, ok := roots["units"].([]interface{})[0].(map[string]interface{})
	s.Require().True(ok)
	s.ElementsMatch([]string{"ouId", "overlayRules"}, keysOf(root))

	rule, ok := allUnits["overlayRules"].(map[string]interface{})["permissions"].(map[string]interface{})
	s.Require().True(ok)
	s.ElementsMatch([]string{"editable", "value", "allowedValues", "excludedValues"}, keysOf(rule))
}

// A section the policy does not state is left off the wire altogether, which is what tells a reader
// "not decided" apart from "decided, with nothing to add".
func (s *ModelTestSuite) TestPolicyRequestLeavesAnUnstatedSectionOffTheWire() {
	req := PolicyRequest{ChildUnits: &ChildUnitsRequest{Mode: ChildUnitsAll}}

	asJSON, err := json.Marshal(req)
	s.Require().NoError(err)
	s.JSONEq(`{"childUnits":{"mode":"all"}}`, string(asJSON))

	asYAML, err := yaml.Marshal(req)
	s.Require().NoError(err)
	s.YAMLEq("childUnits:\n  mode: all\n", string(asYAML))
}

// A resource file states a section the way it reads, and each one lands in its own field.
func (s *ModelTestSuite) TestPolicyRequestReadsTheSectionsOfAResourceFile() {
	document := []byte(`
id: storefront-policy
childUnits:
  mode: selected
  units:
    - ouId: acme-retail
      reach: subtree
      overlayRules:
        tokenValidity:
          editable: false
          value: ["900"]
    - ouId: acme-wholesale
      reach: unit
otherTreeRoots:
  mode: all
  excludedOuIds: [contoso]
`)

	var req PolicyRequest
	s.Require().NoError(yaml.Unmarshal(document, &req))

	s.Equal("storefront-policy", req.ID)
	s.Nil(req.AllUnits)
	s.Require().NotNil(req.ChildUnits)
	s.Equal(ChildUnitsSelected, req.ChildUnits.Mode)
	s.Require().Len(req.ChildUnits.Units, 2)
	s.Equal(ChildUnitRequest{
		OUID: "acme-retail", Reach: ReachSubtree,
		OverlayRules: map[string]OverlayRule{"tokenValidity": {Value: members("900")}},
	}, req.ChildUnits.Units[0])
	s.Equal(ChildUnitRequest{OUID: "acme-wholesale", Reach: ReachUnit}, req.ChildUnits.Units[1])
	s.Require().NotNil(req.OtherTreeRoots)
	s.Equal(OtherTreeRootsAll, req.OtherTreeRoots.Mode)
	s.Equal([]string{"contoso"}, req.OtherTreeRoots.ExcludedOUIDs)
}

// A bare "allUnits: {}" states the section with nothing to exclude, which is the whole of the
// decision for a resource shared with everyone.
func (s *ModelTestSuite) TestPolicyRequestReadsAnEmptyAllUnitsSectionAsStated() {
	var fromYAML PolicyRequest
	s.Require().NoError(yaml.Unmarshal([]byte("allUnits: {}\n"), &fromYAML))
	s.NotNil(fromYAML.AllUnits)

	var fromJSON PolicyRequest
	s.Require().NoError(json.Unmarshal([]byte(`{"allUnits":{}}`), &fromJSON))
	s.NotNil(fromJSON.AllUnits)
}

// An export is written as YAML and served as JSON, so a field named differently by the two
// encodings would not survive the round trip.
func (s *ModelTestSuite) TestReplayablePolicyUsesTheSameCamelCaseNamesInBothEncodings() {
	asJSON, err := json.Marshal(ReplayablePolicy{InitiatingOUID: "ou-1", Request: fullRequest()})
	s.Require().NoError(err)
	asYAML, err := yaml.Marshal(ReplayablePolicy{InitiatingOUID: "ou-1", Request: fullRequest()})
	s.Require().NoError(err)

	var fromJSON map[string]interface{}
	s.Require().NoError(json.Unmarshal(asJSON, &fromJSON))

	s.ElementsMatch([]string{"initiatingOuId", "request"}, keysOf(fromJSON))
	s.Equal(fromJSON, s.normalize(asYAML))
}

// normalize re-encodes a yaml document through json, so the two can be compared on field names
// without their differing number types getting in the way.
func (s *ModelTestSuite) normalize(document []byte) map[string]interface{} {
	s.T().Helper()
	var intermediate map[string]interface{}
	s.Require().NoError(yaml.Unmarshal(document, &intermediate))
	encoded, err := json.Marshal(intermediate)
	s.Require().NoError(err)
	var out map[string]interface{}
	s.Require().NoError(json.Unmarshal(encoded, &out))
	return out
}

// The distinction the pointer types exist for has to survive the wire, or a rule that permits
// nothing comes back as one that permits everything.
func (s *ModelTestSuite) TestOverlayRuleRoundTripsAbsentSeparatelyFromEmpty() {
	tests := []struct {
		name  string
		rule  OverlayRule
		check func(got OverlayRule)
	}{
		{
			name: "omitted stays omitted",
			rule: OverlayRule{Editable: true},
			check: func(got OverlayRule) {
				s.Nil(got.AllowedValues, "an omitted bound must not come back as a constraint")
			},
		},
		{
			name: "explicitly empty stays empty",
			rule: OverlayRule{Editable: true, AllowedValues: members()},
			check: func(got OverlayRule) {
				s.Require().NotNil(got.AllowedValues, "an empty bound permits nothing and must survive")
				s.Empty(*got.AllowedValues)
			},
		},
		{
			name: "a populated bound survives",
			rule: OverlayRule{Editable: true, AllowedValues: members("a", "b")},
			check: func(got OverlayRule) {
				s.Require().NotNil(got.AllowedValues)
				s.Equal([]string{"a", "b"}, *got.AllowedValues)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name+" (json)", func() {
			data, err := json.Marshal(tt.rule)
			s.Require().NoError(err)
			var got OverlayRule
			s.Require().NoError(json.Unmarshal(data, &got))
			tt.check(got)
		})
		s.Run(tt.name+" (yaml)", func() {
			data, err := yaml.Marshal(tt.rule)
			s.Require().NoError(err)
			var got OverlayRule
			s.Require().NoError(yaml.Unmarshal(data, &got))
			tt.check(got)
		})
	}
}

// keysOf returns a map's keys, so field naming can be asserted independently of values.
func keysOf(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
