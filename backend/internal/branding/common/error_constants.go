/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package common

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client errors shared across branding operations.
var (
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1001",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
	// ErrorMissingBrandingID is the error returned when branding ID is missing.
	ErrorMissingBrandingID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1002",
		Error:            "Invalid request format",
		ErrorDescription: "Branding ID is required",
	}
	// ErrorBrandingNotFound is the error returned when a branding configuration is not found.
	ErrorBrandingNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1003",
		Error:            "Branding configuration not found",
		ErrorDescription: "The branding configuration with the specified id does not exist",
	}
	// ErrorCannotDeleteBranding is the error returned when branding cannot be deleted.
	ErrorCannotDeleteBranding = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1004",
		Error:            "Cannot delete branding",
		ErrorDescription: "Cannot delete branding configuration that is currently associated with one or more applications",
	}
	// ErrorMissingDisplayName is the error returned when displayName field is missing.
	ErrorMissingDisplayName = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1005",
		Error:            "Invalid request format",
		ErrorDescription: "The 'displayName' field is required",
	}
	// ErrorMissingPreferences is the error returned when preferences field is missing.
	ErrorMissingPreferences = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1006",
		Error:            "Invalid request format",
		ErrorDescription: "The 'preferences' field is required",
	}
	// ErrorInvalidPreferences is the error returned when preferences structure is invalid.
	ErrorInvalidPreferences = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1007",
		Error:            "Invalid preferences structure",
		ErrorDescription: "The preferences must be a valid JSON object",
	}
	// ErrorInvalidLimit is the error returned when limit parameter is invalid.
	ErrorInvalidLimit = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1008",
		Error:            "Invalid limit parameter",
		ErrorDescription: "The limit parameter must be a positive integer",
	}
	// ErrorInvalidOffset is the error returned when offset parameter is invalid.
	ErrorInvalidOffset = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1009",
		Error:            "Invalid offset parameter",
		ErrorDescription: "The offset parameter must be a non-negative integer",
	}
	// ErrorInvalidResolveType is the error returned when resolve type parameter is missing or invalid.
	ErrorInvalidResolveType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1010",
		Error:            "Invalid request format",
		ErrorDescription: "The 'type' query parameter is required and must be either 'APP' or 'OU'",
	}
	// ErrorMissingResolveID is the error returned when resolve id parameter is missing.
	ErrorMissingResolveID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1011",
		Error:            "Invalid request format",
		ErrorDescription: "The 'id' query parameter is required",
	}
	// ErrorUnsupportedResolveType is the error returned when resolve type is not yet supported.
	ErrorUnsupportedResolveType = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1012",
		Error:            "Unsupported resolve type",
		ErrorDescription: "The specified resolve type is not yet supported. Currently only 'APP' type is supported",
	}
	// ErrorApplicationHasNoBranding is the error returned when an application has no associated branding.
	ErrorApplicationHasNoBranding = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1013",
		Error:            "Application has no branding",
		ErrorDescription: "The specified application does not have an associated branding configuration",
	}
	// ErrorApplicationNotFound is the error returned when an application is not found.
	ErrorApplicationNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "BRD-1014",
		Error:            "Application not found",
		ErrorDescription: "The application with the specified id does not exist",
	}
)

// Internal error constants shared across branding operations.
var (
	// ErrBrandingNotFound is returned when the branding configuration is not found in the system.
	ErrBrandingNotFound = errors.New("branding not found")
)
