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

// Package model defines the data structures and interfaces for organization unit management.
package model

// OrganizationUnitBasic represents the basic information of an organization unit.
type OrganizationUnitBasic struct {
	ID          string `json:"id"`
	Handle      string `json:"handle"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// OrganizationUnit represents an organization unit.
type OrganizationUnit struct {
	ID          string  `json:"id"`
	Handle      string  `json:"handle"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent"`
}

// OrganizationUnitRequest represents the request body for creating an organization unit.
type OrganizationUnitRequest struct {
	Handle      string  `json:"handle"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Parent      *string `json:"parent"`
}

// Link represents a pagination link.
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// OrganizationUnitListResponse represents the response for listing organization units with pagination.
type OrganizationUnitListResponse struct {
	TotalResults      int                     `json:"totalResults"`
	StartIndex        int                     `json:"startIndex"`
	Count             int                     `json:"count"`
	OrganizationUnits []OrganizationUnitBasic `json:"organizationUnits"`
	Links             []Link                  `json:"links"`
}

// User represents a user with basic information for OU endpoints.
type User struct {
	ID string `json:"id"`
}

// Group represents a group with basic information for OU endpoints.
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UserListResponse represents the response for listing users in an organization unit.
type UserListResponse struct {
	TotalResults int    `json:"totalResults"`
	StartIndex   int    `json:"startIndex"`
	Count        int    `json:"count"`
	Users        []User `json:"users"`
	Links        []Link `json:"links"`
}

// GroupListResponse represents the response for listing groups in an organization unit.
type GroupListResponse struct {
	TotalResults int     `json:"totalResults"`
	StartIndex   int     `json:"startIndex"`
	Count        int     `json:"count"`
	Groups       []Group `json:"groups"`
	Links        []Link  `json:"links"`
}
