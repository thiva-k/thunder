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

// Package constants defines error constants for user management operations.
package constants

import (
	"errors"

	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// Client errors for user management operations.
var (
	// ErrorInvalidRequestFormat is the error returned when the request format is invalid.
	ErrorInvalidRequestFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1001",
		Error:            "Invalid request format",
		ErrorDescription: "The request body is malformed or contains invalid data",
	}
	// ErrorMissingUserID is the error returned when user ID is missing.
	ErrorMissingUserID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1002",
		Error:            "Invalid request format",
		ErrorDescription: "User ID is required",
	}
	// ErrorUserNotFound is the error returned when a user is not found.
	ErrorUserNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1003",
		Error:            "User not found",
		ErrorDescription: "The user with the specified id does not exist",
	}
	// ErrorOrganizationUnitNotFound is the error returned when an organization unit is not found.
	ErrorOrganizationUnitNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1005",
		Error:            "Organization unit not found",
		ErrorDescription: "The specified organization unit does not exist",
	}
	// ErrorInvalidGroupID is the error returned when group ID is invalid.
	ErrorInvalidGroupID = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1007",
		Error:            "Invalid group ID",
		ErrorDescription: "One or more group IDs in the request do not exist",
	}
	// ErrorHandlePathRequired is the error returned when handle path is missing.
	ErrorHandlePathRequired = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1008",
		Error:            "Handle path required",
		ErrorDescription: "Handle path is required for this operation",
	}
	// ErrorInvalidHandlePath is the error returned when handle path format is invalid.
	ErrorInvalidHandlePath = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1009",
		Error:            "Invalid handle path",
		ErrorDescription: "Handle path must contain valid organizational unit identifiers separated by forward slashes",
	}
	// ErrorInvalidLimit is the error returned when limit parameter is invalid.
	ErrorInvalidLimit = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1011",
		Error:            "Invalid pagination parameter",
		ErrorDescription: "The limit parameter must be a positive integer",
	}
	// ErrorInvalidOffset is the error returned when offset parameter is invalid.
	ErrorInvalidOffset = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1012",
		Error:            "Invalid pagination parameter",
		ErrorDescription: "The offset parameter must be a non-negative integer",
	}
	// ErrorUsernameConflict is the error returned when username already exists.
	ErrorUsernameConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1014",
		Error:            "Username conflict",
		ErrorDescription: "A user with the same username already exists",
	}
	// ErrorEmailConflict is the error returned when email already exists.
	ErrorEmailConflict = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1015",
		Error:            "Email conflict",
		ErrorDescription: "A user with the same email already exists",
	}
	// ErrorMissingRequiredFields is the error returned when required fields are missing.
	ErrorMissingRequiredFields = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1016",
		Error:            "Missing required fields",
		ErrorDescription: "At least one identifying attribute must be provided",
	}
	// ErrorMissingCredentials is the error returned when credentials are missing.
	ErrorMissingCredentials = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1017",
		Error:            "Missing credentials",
		ErrorDescription: "At least one credential field must be provided",
	}
	// ErrorAuthenticationFailed is the error returned when authentication fails.
	ErrorAuthenticationFailed = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "USR-1018",
		Error:            "Authentication failed",
		ErrorDescription: "Invalid credentials provided",
	}
)

// Server errors for user management operations.
var (
	// ErrorInternalServerError is the error returned when an internal server error occurs.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "USR-5000",
		Error:            "Internal server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)

// Error variables
var (
	// ErrUserNotFound is returned when the user is not found in the system.
	ErrUserNotFound = errors.New("user not found")
)
