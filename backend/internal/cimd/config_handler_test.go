// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package cimd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const testBlockedClientID = "https://claude.ai/oauth/blocked.json"

type ConfigHandlerTestSuite struct {
	suite.Suite
	handler *ConfigHandler
}

func TestConfigHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigHandlerTestSuite))
}

func (s *ConfigHandlerTestSuite) SetupTest() {
	s.handler = NewConfigHandler()
}

func (s *ConfigHandlerTestSuite) TestDecode() {
	empty, err := s.handler.Decode(nil)
	s.Require().NoError(err)
	s.Equal(unregisteredClientSettings{}, empty)

	decoded, err := s.handler.Decode(json.RawMessage(`{"enabled":true,"blockedClientIds":["` +
		testBlockedClientID + `"]}`))
	s.Require().NoError(err)
	settings := decoded.(unregisteredClientSettings)
	s.True(settings.Enabled)
	s.Equal([]string{testBlockedClientID}, settings.BlockedClientIDs)

	_, err = s.handler.Decode(json.RawMessage(`{"enabled":true,"policies":[]}`))
	s.Error(err)
	_, err = s.handler.Decode(json.RawMessage(`{`))
	s.Error(err)
}

func (s *ConfigHandlerTestSuite) TestValidate_AcceptsValidSettings() {
	settings := unregisteredClientSettings{Enabled: true, BlockedClientIDs: []string{testBlockedClientID}}

	s.NoError(s.handler.Validate(context.Background(), settings, unregisteredClientSettings{}, nil))
}

func (s *ConfigHandlerTestSuite) TestValidate_RefusesInvalidSettings() {
	cases := []struct {
		name     string
		settings unregisteredClientSettings
	}{
		{"blocked client ID not a URL", unregisteredClientSettings{BlockedClientIDs: []string{"claude.ai"}}},
		{"blocked client ID twice", unregisteredClientSettings{
			BlockedClientIDs: []string{testBlockedClientID, testBlockedClientID}}},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.Error(s.handler.Validate(context.Background(), tc.settings, unregisteredClientSettings{}, nil))
		})
	}
}

func (s *ConfigHandlerTestSuite) TestValidate_RefusesTooManyBlockedClientIDs() {
	tooMany := unregisteredClientSettings{}
	for i := 0; i <= maxBlockedClientIDs; i++ {
		tooMany.BlockedClientIDs = append(tooMany.BlockedClientIDs,
			"https://claude.ai/oauth/"+strings.Repeat("a", i+1))
	}

	s.Error(s.handler.Validate(context.Background(), tooMany, nil, nil))
}

func (s *ConfigHandlerTestSuite) TestValidate_RefusesWhileDeclarative() {
	readOnly := unregisteredClientSettings{Enabled: true}

	err := s.handler.Validate(context.Background(), unregisteredClientSettings{}, readOnly, nil)

	s.ErrorIs(err, errSettingsLocked)
}

func (s *ConfigHandlerTestSuite) TestValidate_RefusesUnexpectedType() {
	s.Error(s.handler.Validate(context.Background(), "settings", nil, nil))
}

func (s *ConfigHandlerTestSuite) TestMerge() {
	writable := unregisteredClientSettings{Enabled: true}

	merged := s.handler.Merge(unregisteredClientSettings{}, writable).(unregisteredClientSettings)
	s.True(merged.Enabled)
	s.NotNil(merged.BlockedClientIDs)

	readOnly := unregisteredClientSettings{BlockedClientIDs: []string{testBlockedClientID}}
	merged = s.handler.Merge(readOnly, writable).(unregisteredClientSettings)
	s.False(merged.Enabled)
	s.Equal([]string{testBlockedClientID}, merged.BlockedClientIDs)

	empty := s.handler.Merge(nil, nil).(unregisteredClientSettings)
	s.Equal(unregisteredClientSettings{BlockedClientIDs: []string{}}, empty)
}
