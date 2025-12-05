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

package brandingmgt

import (
	"encoding/json"

	"github.com/asgardeo/thunder/internal/branding/common"
)

// BrandingListItem represents a branding item in the list response.
type BrandingListItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// CreateBrandingRequest represents the request body for creating a branding configuration.
type CreateBrandingRequest struct {
	DisplayName string          `json:"displayName"`
	Preferences json.RawMessage `json:"preferences"`
}

// UpdateBrandingRequest represents the request body for updating a branding configuration.
type UpdateBrandingRequest struct {
	DisplayName string          `json:"displayName"`
	Preferences json.RawMessage `json:"preferences"`
}

// BrandingListResponse represents the response for listing branding configurations with pagination.
type BrandingListResponse struct {
	TotalResults int                `json:"totalResults"`
	StartIndex   int                `json:"startIndex"`
	Count        int                `json:"count"`
	Brandings    []BrandingListItem `json:"brandings"`
	Links        []LinkResponse     `json:"links"`
}

// LinkResponse represents a pagination link.
type LinkResponse struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// Link represents a pagination link.
type Link struct {
	Href string
	Rel  string
}

// BrandingList represents the result of listing branding configurations.
type BrandingList struct {
	TotalResults int
	StartIndex   int
	Count        int
	Brandings    []common.Branding
	Links        []Link
}
