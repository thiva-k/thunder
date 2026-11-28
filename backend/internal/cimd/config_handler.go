// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package cimd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// errSettingsLocked refuses a write to the cimd section while a declarative document sets it.
var errSettingsLocked = errors.New("cimd: the section is set by a declarative resource and is read-only")

// ConfigHandler decodes, validates and merges the cimd server configuration section.
type ConfigHandler struct{}

// NewConfigHandler creates the handler of the cimd server configuration section.
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// Decode parses a raw section value. Empty input yields empty settings. Unknown fields are rejected so a
// misspelled key does not silently admit or refuse clients.
func (h *ConfigHandler) Decode(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return unregisteredClientSettings{}, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var settings unregisteredClientSettings
	if err := decoder.Decode(&settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// Validate refuses a write while a declarative document sets the section, then checks the incoming
// settings.
func (h *ConfigHandler) Validate(_ context.Context, incoming, readOnly, _ any) error {
	if ro, ok := readOnly.(unregisteredClientSettings); ok && !ro.isEmpty() {
		return errSettingsLocked
	}
	settings, ok := incoming.(unregisteredClientSettings)
	if !ok {
		return fmt.Errorf("cimd: unexpected config type %T", incoming)
	}
	return validateSettings(settings)
}

// Merge returns the declarative document when one sets the section, and the writable layer otherwise.
func (h *ConfigHandler) Merge(readOnly, writable any) any {
	if ro, ok := readOnly.(unregisteredClientSettings); ok && !ro.isEmpty() {
		return ro.normalized()
	}
	wr, _ := writable.(unregisteredClientSettings)
	return wr.normalized()
}

// isEmpty reports whether the settings set nothing.
func (s unregisteredClientSettings) isEmpty() bool {
	return !s.Enabled && len(s.BlockedClientIDs) == 0
}

// normalized returns the settings with an empty list in place of an absent one.
func (s unregisteredClientSettings) normalized() unregisteredClientSettings {
	if s.BlockedClientIDs == nil {
		s.BlockedClientIDs = []string{}
	}
	return s
}

// validateSettings checks the limit and that each blocked Client Identifier URL is valid and listed once.
func validateSettings(s unregisteredClientSettings) error {
	if len(s.BlockedClientIDs) > maxBlockedClientIDs {
		return fmt.Errorf("cimd: at most %d blocked client IDs are allowed", maxBlockedClientIDs)
	}
	blocked := map[string]bool{}
	for i, clientID := range s.BlockedClientIDs {
		if _, err := parseClientIdentifierURL(clientID); err != nil {
			return fmt.Errorf("cimd: blockedClientIds[%d]: %w", i, err)
		}
		if blocked[clientID] {
			return fmt.Errorf("cimd: blockedClientIds[%d]: listed more than once", i)
		}
		blocked[clientID] = true
	}
	return nil
}
