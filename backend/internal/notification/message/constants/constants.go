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

// Package constants contains the constants for message notification service.
package constants

// MessageProviderType defines the type of messaging provider.
type MessageProviderType string

const (
	// MessageProviderTypeVonage represents the Vonage messaging provider.
	MessageProviderTypeVonage MessageProviderType = "vonage"
	// MessageProviderTypeTwilio represents the Twilio messaging provider.
	MessageProviderTypeTwilio MessageProviderType = "twilio"
	// MessageProviderTypeCustom represents a custom messaging provider.
	MessageProviderTypeCustom MessageProviderType = "custom"
)

const (
	// VonagePropKeyAPIKey is the property key for the Vonage API key.
	VonagePropKeyAPIKey = "api_key"
	// VonagePropKeyAPISecret is the property key for the Vonage API secret.
	VonagePropKeyAPISecret = "api_secret"
	// VonagePropKeySenderID is the property key for the Vonage sender ID.
	VonagePropKeySenderID = "sender_id"
)

const (
	// TwilioPropKeyAccountSID is the property key for the Twilio account SID.
	TwilioPropKeyAccountSID = "account_sid"
	// TwilioPropKeyAuthToken is the property key for the Twilio auth token.
	TwilioPropKeyAuthToken = "auth_token"
	// TwilioPropKeySenderID is the property key for the Twilio sender ID.
	TwilioPropKeySenderID = "sender_id"
)

const (
	// CustomPropKeyURL is the property key for the custom URL.
	CustomPropKeyURL = "url"
	// CustomPropKeyHTTPMethod is the property key for the HTTP method.
	CustomPropKeyHTTPMethod = "http_method"
	// CustomPropKeyHTTPHeaders is the property key for the HTTP headers.
	CustomPropKeyHTTPHeaders = "http_headers"
	// CustomPropKeyContentType is the property key for the content type.
	CustomPropKeyContentType = "content_type"
)
