/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package entityprovider

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/asgardeo/thunder/internal/entity"
	"github.com/asgardeo/thunder/internal/system/security"
)

type defaultEntityProvider struct {
	entitySvc entity.EntityServiceInterface
}

// newDefaultEntityProvider creates a new default entity provider.
func newDefaultEntityProvider(entitySvc entity.EntityServiceInterface) EntityProviderInterface {
	return &defaultEntityProvider{entitySvc: entitySvc}
}

// IdentifyEntity identifies an entity based on the given filters.
func (p *defaultEntityProvider) IdentifyEntity(
	filters map[string]interface{}) (*string, *EntityProviderError) {

	ctx := security.WithRuntimeContext(context.Background())
	id, err := p.entitySvc.IdentifyEntity(ctx, filters)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, NewEntityProviderError(ErrorCodeEntityNotFound, "Entity not found", err.Error())
		}
		return nil, NewEntityProviderError(ErrorCodeSystemError, "Failed to identify entity", err.Error())
	}
	return id, nil
}

// GetEntity retrieves an entity by ID.
func (p *defaultEntityProvider) GetEntity(
	entityID string) (*Entity, *EntityProviderError) {

	ctx := security.WithRuntimeContext(context.Background())
	e, err := p.entitySvc.GetEntity(ctx, entityID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, NewEntityProviderError(ErrorCodeEntityNotFound, "Entity not found", err.Error())
		}
		return nil, NewEntityProviderError(ErrorCodeSystemError, "Failed to get entity", err.Error())
	}
	return toProviderEntity(e), nil
}

// GetEntityGroups retrieves the groups that an entity belongs to.
func (p *defaultEntityProvider) GetEntityGroups(
	entityID string, limit, offset int) (*EntityGroupListResponse, *EntityProviderError) {

	ctx := security.WithRuntimeContext(context.Background())
	count, err := p.entitySvc.GetGroupCountForEntity(ctx, entityID)
	if err != nil {
		return nil, NewEntityProviderError(ErrorCodeSystemError, "Failed to get group count", err.Error())
	}

	groups, err := p.entitySvc.GetEntityGroups(ctx, entityID, limit, offset)
	if err != nil {
		return nil, NewEntityProviderError(ErrorCodeSystemError, "Failed to get entity groups", err.Error())
	}

	providerGroups := make([]EntityGroup, len(groups))
	for i, g := range groups {
		providerGroups[i] = EntityGroup{
			ID:   g.ID,
			Name: g.Name,
			OUID: g.OUID,
		}
	}

	return &EntityGroupListResponse{
		TotalResults: count,
		Groups:       providerGroups,
	}, nil
}

// UpdateEntity updates an entity.
func (p *defaultEntityProvider) UpdateEntity(
	entityID string, providerEntity *Entity) (*Entity, *EntityProviderError) {

	if providerEntity == nil {
		return nil, NewEntityProviderError(
			ErrorCodeInvalidRequestFormat, "Invalid request", "Entity cannot be nil")
	}

	ctx := security.WithRuntimeContext(context.Background())
	internalEntity := fromProviderEntity(providerEntity)
	updated, err := p.entitySvc.UpdateEntity(ctx, entityID, internalEntity)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, NewEntityProviderError(ErrorCodeEntityNotFound, "Entity not found", err.Error())
		}
		return nil, NewEntityProviderError(ErrorCodeSystemError, "Failed to update entity", err.Error())
	}
	return toProviderEntity(updated), nil
}

// CreateEntity creates a new entity.
func (p *defaultEntityProvider) CreateEntity(
	providerEntity *Entity, systemCredentials json.RawMessage) (*Entity, *EntityProviderError) {

	if providerEntity == nil {
		return nil, NewEntityProviderError(
			ErrorCodeInvalidRequestFormat, "Invalid request", "Entity cannot be nil")
	}

	ctx := security.WithRuntimeContext(context.Background())
	internalEntity := fromProviderEntity(providerEntity)
	created, err := p.entitySvc.CreateEntity(ctx, internalEntity, nil, systemCredentials)
	if err != nil {
		return nil, NewEntityProviderError(ErrorCodeSystemError, "Failed to create entity", err.Error())
	}
	return toProviderEntity(created), nil
}

// UpdateEntityCredentials updates the system credentials of an entity.
func (p *defaultEntityProvider) UpdateEntityCredentials(
	entityID string, credentials json.RawMessage) *EntityProviderError {

	ctx := security.WithRuntimeContext(context.Background())
	err := p.entitySvc.UpdateSystemCredentials(ctx, entityID, credentials)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return NewEntityProviderError(ErrorCodeEntityNotFound, "Entity not found", err.Error())
		}
		return NewEntityProviderError(ErrorCodeSystemError, "Failed to update credentials", err.Error())
	}
	return nil
}

// DeleteEntity deletes an entity.
func (p *defaultEntityProvider) DeleteEntity(entityID string) *EntityProviderError {
	ctx := security.WithRuntimeContext(context.Background())
	err := p.entitySvc.DeleteEntity(ctx, entityID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil // idempotent
		}
		return NewEntityProviderError(ErrorCodeSystemError, "Failed to delete entity", err.Error())
	}
	return nil
}

// AddSystemIdentifier adds a system-managed identifier for an entity.
func (p *defaultEntityProvider) AddSystemIdentifier(
	entityID string, idType string, value string) *EntityProviderError {

	ctx := security.WithRuntimeContext(context.Background())
	err := p.entitySvc.AddSystemIdentifier(ctx, entityID, idType, value)
	if err != nil {
		return NewEntityProviderError(ErrorCodeSystemError, "Failed to add identifier", err.Error())
	}
	return nil
}

// ValidateEntityIDs validates that the given entity IDs exist.
func (p *defaultEntityProvider) ValidateEntityIDs(
	entityIDs []string) ([]string, *EntityProviderError) {

	ctx := security.WithRuntimeContext(context.Background())
	invalidIDs, err := p.entitySvc.ValidateEntityIDs(ctx, entityIDs)
	if err != nil {
		return nil, NewEntityProviderError(
			ErrorCodeSystemError, "Failed to validate entity IDs", err.Error())
	}
	return invalidIDs, nil
}

// GetEntitiesByIDs retrieves entities by their IDs.
func (p *defaultEntityProvider) GetEntitiesByIDs(
	entityIDs []string) ([]*Entity, *EntityProviderError) {

	ctx := security.WithRuntimeContext(context.Background())
	entities, err := p.entitySvc.GetEntitiesByIDs(ctx, entityIDs)
	if err != nil {
		return nil, NewEntityProviderError(
			ErrorCodeSystemError, "Failed to get entities", err.Error())
	}

	result := make([]*Entity, len(entities))
	for i := range entities {
		result[i] = toProviderEntity(&entities[i])
	}
	return result, nil
}

// toProviderEntity converts an internal Entity to a provider Entity.
func toProviderEntity(e *entity.Entity) *Entity {
	return &Entity{
		EntityID:         e.EntityID,
		EntityCategory:   EntityCategory(e.EntityCategory),
		EntityType:       e.EntityType,
		OUID:             e.OrganizationUnitID,
		Attributes:       e.Attributes,
		SystemAttributes: e.SystemAttributes,
	}
}

// fromProviderEntity converts a provider Entity to an internal Entity.
func fromProviderEntity(e *Entity) *entity.Entity {
	return &entity.Entity{
		EntityID:           e.EntityID,
		EntityCategory:     entity.EntityCategory(e.EntityCategory),
		EntityType:         e.EntityType,
		OrganizationUnitID: e.OUID,
		Attributes:         e.Attributes,
		SystemAttributes:   e.SystemAttributes,
	}
}
