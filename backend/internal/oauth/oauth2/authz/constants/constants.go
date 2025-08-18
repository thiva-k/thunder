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

// Package constants defines constants related to OAuth2 authorization.
package constants

import "errors"

// Authorization code states.
const (
	AuthCodeStateActive   = "ACTIVE"
	AuthCodeStateInactive = "INACTIVE"
	AuthCodeStateExpired  = "EXPIRED"
	AuthCodeStateRevoked  = "REVOKED"
)

// ErrAuthorizationCodeNotFound is returned when an authorization code is not found in the database.
var ErrAuthorizationCodeNotFound = errors.New("authorization code not found")
