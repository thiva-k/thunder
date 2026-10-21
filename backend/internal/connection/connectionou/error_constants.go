// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package connectionou

import tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"

var (
	// ErrorInvalidOU is the error returned when the organization unit given for a connection does not exist.
	ErrorInvalidOU = tidcommon.ServiceError{
		Type: tidcommon.ClientErrorType,
		Code: "CON-1023",
		Error: tidcommon.I18nMessage{
			Key:          "error.connectionservice.invalid_connection_ou",
			DefaultValue: "Invalid organization unit",
		},
		ErrorDescription: tidcommon.I18nMessage{
			Key:          "error.connectionservice.invalid_connection_ou_description",
			DefaultValue: "The organization unit given for the connection does not exist",
		},
	}
	// ErrorOUUpdateNotAllowed is the error returned when attempting to change the organization unit of a connection.
	ErrorOUUpdateNotAllowed = tidcommon.ServiceError{
		Type: tidcommon.ClientErrorType,
		Code: "CON-1024",
		Error: tidcommon.I18nMessage{
			Key:          "error.connectionservice.connection_ou_update_not_allowed",
			DefaultValue: "Organization unit update not allowed",
		},
		ErrorDescription: tidcommon.I18nMessage{
			Key:          "error.connectionservice.connection_ou_update_not_allowed_description",
			DefaultValue: "The organization unit of a connection cannot be changed",
		},
	}
	// ErrorOURequired is the error returned when a connection is created without an organization unit.
	ErrorOURequired = tidcommon.ServiceError{
		Type: tidcommon.ClientErrorType,
		Code: "CON-1025",
		Error: tidcommon.I18nMessage{
			Key:          "error.connectionservice.connection_ou_required",
			DefaultValue: "Organization unit required",
		},
		ErrorDescription: tidcommon.I18nMessage{
			Key:          "error.connectionservice.connection_ou_required_description",
			DefaultValue: "A connection needs an organization unit. Give its ouId in the request",
		},
	}
)
