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

package userinfo

import "github.com/asgardeo/thunder/internal/system/error/serviceerror"

// Error constants for UserInfo endpoint following OIDC specification.
var (
	ErrorMissingToken = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "invalid_request",
		Error:            "Invalid request",
		ErrorDescription: "Access token is required",
	}

	ErrorInvalidTokenFormat = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "invalid_token",
		Error:            "Invalid token",
		ErrorDescription: "The access token is malformed or invalid",
	}

	ErrorInvalidTokenSignature = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "invalid_token",
		Error:            "Invalid token",
		ErrorDescription: "The access token signature is invalid",
	}

	ErrorMissingSubject = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "invalid_token",
		Error:            "Invalid token",
		ErrorDescription: "The access token does not contain a subject claim",
	}

	ErrorInsufficientScope = serviceerror.ServiceError{
		Type:             serviceerror.ClientErrorType,
		Code:             "insufficient_scope",
		Error:            "Insufficient scope",
		ErrorDescription: "The access token does not have the required 'openid' scope",
	}

	ErrorUserNotFound = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "server_error",
		Error:            "Server error",
		ErrorDescription: "Failed to retrieve user information",
	}

	ErrorUserAttributesProcessing = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "server_error",
		Error:            "Server error",
		ErrorDescription: "Failed to process user attributes",
	}

	ErrorInternalServerError = serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "server_error",
		Error:            "Server error",
		ErrorDescription: "An unexpected error occurred while processing the request",
	}
)
