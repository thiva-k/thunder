// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package role

// AssigneeType represents the type of assignee (user or group)
type AssigneeType string

const (
	AssigneeTypeUser  AssigneeType = "user"
	AssigneeTypeGroup AssigneeType = "group"
	AssigneeTypeApp   AssigneeType = "app"
)

// Assignment represents a role assignment
type Assignment struct {
	ID      string       `json:"id"`
	Type    AssigneeType `json:"type"`
	Display string       `json:"display,omitempty"` // Display name (only included with include=display parameter)
}

// CreateRoleRequest represents the request to create a role
type CreateRoleRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	OUID        string                `json:"ouId"`
	Permissions []ResourcePermissions `json:"permissions"`
	Assignments []Assignment          `json:"assignments,omitempty"`
}

// UpdateRoleRequest represents the request to update a role
type UpdateRoleRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	OUID        string                `json:"ouId"`
	Permissions []ResourcePermissions `json:"permissions"`
}

// Role represents a complete role resource
type Role struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	OUID        string                `json:"ouId"`
	Permissions []ResourcePermissions `json:"permissions"`
	Assignments []Assignment          `json:"assignments,omitempty"`
}

// RoleSummary represents a minimal role information
type RoleSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	OUID        string `json:"ouId"`
}

// Link represents a pagination link
type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

// RoleListResponse represents the paginated list of roles
type RoleListResponse struct {
	TotalResults int           `json:"totalResults"`
	StartIndex   int           `json:"startIndex"`
	Count        int           `json:"count"`
	Links        []Link        `json:"links,omitempty"`
	Roles        []RoleSummary `json:"roles"`
}

// AssignmentsRequest represents add/remove assignments request
type AssignmentsRequest struct {
	Assignments []Assignment `json:"assignments"`
}

// AssignmentListResponse represents the paginated list of assignments
type AssignmentListResponse struct {
	TotalResults int          `json:"totalResults"`
	StartIndex   int          `json:"startIndex"`
	Count        int          `json:"count"`
	Links        []Link       `json:"links,omitempty"`
	Assignments  []Assignment `json:"assignments"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}

// ResourcePermissions represents permissions grouped by resource server
type ResourcePermissions struct {
	ResourceServerID string   `json:"resourceServerId"`
	Permissions      []string `json:"permissions"`
}
