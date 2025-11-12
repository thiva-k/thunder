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

// Package branding provides error constants and service errors for branding management operations.
package branding

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client errors for branding management operations.
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
)

// Internal error constants for branding management operations.
var (
	// ErrBrandingNotFound is returned when the branding configuration is not found in the system.
	ErrBrandingNotFound = errors.New("branding not found")
)
