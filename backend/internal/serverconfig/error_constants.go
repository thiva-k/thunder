// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package serverconfig

import (
	"github.com/thunder-id/thunderid/pkg/thunderidengine/common"
)

// Client errors for server-config operations.
var (
	// ErrorUnsupportedConfigName is the error returned when the config name is not supported.
	ErrorUnsupportedConfigName = common.ServiceError{
		Type: common.ClientErrorType,
		Code: "SCF-1001",
		Error: common.I18nMessage{
			Key:          "error.serverconfigservice.unsupported_config_name",
			DefaultValue: "Unsupported configuration name",
		},
		ErrorDescription: common.I18nMessage{
			Key:          "error.serverconfigservice.unsupported_config_name_description",
			DefaultValue: "The requested server configuration name is not supported",
		},
	}

	// ErrorConfigNotFound is the error returned when the requested config is not found.
	ErrorConfigNotFound = common.ServiceError{
		Type: common.ClientErrorType,
		Code: "SCF-1002",
		Error: common.I18nMessage{
			Key:          "error.serverconfigservice.config_not_found",
			DefaultValue: "Server configuration not found",
		},
		ErrorDescription: common.I18nMessage{
			Key:          "error.serverconfigservice.config_not_found_description",
			DefaultValue: "The requested server configuration does not exist",
		},
	}

	// ErrorInvalidConfigValue is the error returned when the config value is invalid.
	ErrorInvalidConfigValue = common.ServiceError{
		Type: common.ClientErrorType,
		Code: "SCF-1003",
		Error: common.I18nMessage{
			Key:          "error.serverconfigservice.invalid_config_value",
			DefaultValue: "Invalid server configuration value",
		},
		ErrorDescription: common.I18nMessage{
			Key:          "error.serverconfigservice.invalid_config_value_description",
			DefaultValue: "The provided server configuration value is invalid",
		},
	}

	// ErrorInvalidRequestFormat is the error returned when the request body is malformed.
	ErrorInvalidRequestFormat = common.ServiceError{
		Type: common.ClientErrorType,
		Code: "SCF-1004",
		Error: common.I18nMessage{
			Key:          "error.serverconfigservice.invalid_request_format",
			DefaultValue: "Invalid request format",
		},
		ErrorDescription: common.I18nMessage{
			Key:          "error.serverconfigservice.invalid_request_format_description",
			DefaultValue: "The request body is malformed or contains invalid data",
		},
	}
)
