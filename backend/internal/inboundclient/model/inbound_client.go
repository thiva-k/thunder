// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// Package model holds public data types for the inbound client subsystem.
//
//nolint:lll
package model

import "github.com/thunder-id/thunderid/pkg/thunderidengine/providers"

type (
	// InboundClient is the persistence shape for protocol-agnostic inbound client record.
	InboundClient = providers.InboundClient
	// AssertionConfig is the entity-level assertion config; token configs fall back to it.
	AssertionConfig = providers.AssertionConfig
	// LoginConsentConfig is the login consent configuration.
	LoginConsentConfig = providers.LoginConsentConfig
	// Certificate is a user-supplied certificate input.
	Certificate = providers.Certificate
)

// InboundClientAttributes is the flattened view of one inbound client's configured user attributes.
type InboundClientAttributes struct {
	InboundClientID string
	Attributes      []string
}

// DeclarativeLoaderConfig describes how to load inbound clients from a YAML resource directory.
type DeclarativeLoaderConfig struct {
	ResourceType  string
	DirectoryName string
	Parser        func(data []byte) (*InboundClient, error)
	Validator     func(*InboundClient) error
}
