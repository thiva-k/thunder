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

package entity

import (
	"context"
	"encoding/json"

	"github.com/asgardeo/thunder/internal/system/transaction"
)

// EntityServiceInterface is the internal interface used by directory-layer consumers
// (user service, group service). This is a richer interface than the boundary
// EntityProviderInterface used by gateway-layer consumers.
type EntityServiceInterface interface {
	// Entity CRUD
	CreateEntity(ctx context.Context, entity *Entity,
		credentials json.RawMessage, systemCredentials json.RawMessage) (*Entity, error)
	GetEntity(ctx context.Context, entityID string) (*Entity, error)
	UpdateEntity(ctx context.Context, entityID string, entity *Entity) (*Entity, error)
	DeleteEntity(ctx context.Context, entityID string) error

	// Credentials (write-only, never returned via GetEntity)
	UpdateSystemCredentials(ctx context.Context, entityID string, creds json.RawMessage) error
	GetEntityWithCredentials(ctx context.Context, entityID string) (
		*Entity, json.RawMessage, json.RawMessage, error)

	// UpdateEntityWithCredentials atomically updates an entity and its system credentials.
	UpdateEntityWithCredentials(ctx context.Context, entityID string,
		entity *Entity, systemCreds json.RawMessage) (*Entity, error)

	// Identifiers
	IdentifyEntity(ctx context.Context, filters map[string]interface{}) (*string, error)

	// Lists (category-scoped)
	GetEntityListCount(ctx context.Context, category EntityCategory,
		filters map[string]interface{}) (int, error)
	GetEntityList(ctx context.Context, category EntityCategory,
		limit, offset int, filters map[string]interface{}) ([]Entity, error)
	GetEntityListCountByOUIDs(ctx context.Context, category EntityCategory,
		ouIDs []string, filters map[string]interface{}) (int, error)
	GetEntityListByOUIDs(ctx context.Context, category EntityCategory,
		ouIDs []string, limit, offset int, filters map[string]interface{}) ([]Entity, error)

	// Bulk
	ValidateEntityIDs(ctx context.Context, entityIDs []string) ([]string, error)
	GetEntitiesByIDs(ctx context.Context, entityIDs []string) ([]Entity, error)
	ValidateEntityIDsInOUs(ctx context.Context, entityIDs []string, ouIDs []string) ([]string, error)

	// Groups (queries GROUP_MEMBER_REFERENCE)
	GetGroupCountForEntity(ctx context.Context, entityID string) (int, error)
	GetEntityGroups(ctx context.Context, entityID string, limit, offset int) ([]EntityGroup, error)

	// Declarative
	IsEntityDeclarative(ctx context.Context, entityID string) (bool, error)
	LoadDeclarativeResources(config DeclarativeLoaderConfig) error

}

// entityService is the default implementation of EntityServiceInterface.
type entityService struct {
	store         entityStoreInterface
	transactioner transaction.Transactioner
}

// newEntityService creates a new entity service.
func newEntityService(
	store entityStoreInterface,
	transactioner transaction.Transactioner,
) *entityService {
	return &entityService{
		store:         store,
		transactioner: transactioner,
	}
}

// CreateEntity creates a new entity.
// Uses a transaction to ensure the entity row and its indexed identifiers are created atomically.
func (s *entityService) CreateEntity(ctx context.Context, entity *Entity,
	credentials json.RawMessage, systemCredentials json.RawMessage) (*Entity, error) {

	var created Entity
	err := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		if err := s.store.CreateEntity(txCtx, *entity, credentials, systemCredentials); err != nil {
			return err
		}

		result, err := s.store.GetEntity(txCtx, entity.EntityID)
		if err != nil {
			return err
		}
		created = result
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// GetEntity retrieves an entity by ID.
func (s *entityService) GetEntity(ctx context.Context, entityID string) (*Entity, error) {
	entity, err := s.store.GetEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// UpdateEntity updates an entity.
// Uses a transaction to ensure the entity update and identifier re-sync are atomic.
func (s *entityService) UpdateEntity(ctx context.Context, entityID string, entity *Entity) (*Entity, error) {
	var updated Entity
	err := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		entity.EntityID = entityID
		if err := s.store.UpdateEntity(txCtx, entity); err != nil {
			return err
		}

		result, err := s.store.GetEntity(txCtx, entityID)
		if err != nil {
			return err
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// DeleteEntity deletes an entity.
func (s *entityService) DeleteEntity(ctx context.Context, entityID string) error {
	return s.store.DeleteEntity(ctx, entityID)
}

// UpdateSystemCredentials updates the system credentials of an entity.
func (s *entityService) UpdateSystemCredentials(ctx context.Context, entityID string,
	creds json.RawMessage) error {
	return s.store.UpdateSystemCredentials(ctx, entityID, creds)
}

// GetEntityWithCredentials retrieves an entity with all credential columns.
func (s *entityService) GetEntityWithCredentials(ctx context.Context, entityID string) (
	*Entity, json.RawMessage, json.RawMessage, error) {

	entity, creds, sysCreds, err := s.store.GetEntityWithCredentials(ctx, entityID)
	if err != nil {
		return nil, nil, nil, err
	}
	return &entity, creds, sysCreds, nil
}

// IdentifyEntity identifies an entity using the given filters.
func (s *entityService) IdentifyEntity(ctx context.Context,
	filters map[string]interface{}) (*string, error) {
	return s.store.IdentifyEntity(ctx, filters)
}

// GetEntityListCount retrieves the total count of entities by category.
func (s *entityService) GetEntityListCount(ctx context.Context, category EntityCategory,
	filters map[string]interface{}) (int, error) {
	return s.store.GetEntityListCount(ctx, string(category), filters)
}

// GetEntityList retrieves a list of entities by category.
func (s *entityService) GetEntityList(ctx context.Context, category EntityCategory,
	limit, offset int, filters map[string]interface{}) ([]Entity, error) {
	return s.store.GetEntityList(ctx, string(category), limit, offset, filters)
}

// GetEntityListCountByOUIDs retrieves the total count of entities scoped to OU IDs.
func (s *entityService) GetEntityListCountByOUIDs(ctx context.Context, category EntityCategory,
	ouIDs []string, filters map[string]interface{}) (int, error) {
	return s.store.GetEntityListCountByOUIDs(ctx, string(category), ouIDs, filters)
}

// GetEntityListByOUIDs retrieves a list of entities scoped to OU IDs.
func (s *entityService) GetEntityListByOUIDs(ctx context.Context, category EntityCategory,
	ouIDs []string, limit, offset int, filters map[string]interface{}) ([]Entity, error) {
	return s.store.GetEntityListByOUIDs(ctx, string(category), ouIDs, limit, offset, filters)
}

// ValidateEntityIDs checks if all provided entity IDs exist.
func (s *entityService) ValidateEntityIDs(ctx context.Context, entityIDs []string) ([]string, error) {
	return s.store.ValidateEntityIDs(ctx, entityIDs)
}

// GetEntitiesByIDs retrieves entities by a list of IDs.
func (s *entityService) GetEntitiesByIDs(ctx context.Context, entityIDs []string) ([]Entity, error) {
	return s.store.GetEntitiesByIDs(ctx, entityIDs)
}

// ValidateEntityIDsInOUs checks which of the provided entity IDs belong to the given OU scope.
func (s *entityService) ValidateEntityIDsInOUs(ctx context.Context,
	entityIDs []string, ouIDs []string) ([]string, error) {
	return s.store.ValidateEntityIDsInOUs(ctx, entityIDs, ouIDs)
}

// GetGroupCountForEntity retrieves the total count of groups an entity belongs to.
func (s *entityService) GetGroupCountForEntity(ctx context.Context, entityID string) (int, error) {
	return s.store.GetGroupCountForEntity(ctx, entityID)
}

// GetEntityGroups retrieves groups that an entity belongs to with pagination.
func (s *entityService) GetEntityGroups(ctx context.Context, entityID string,
	limit, offset int) ([]EntityGroup, error) {
	return s.store.GetEntityGroups(ctx, entityID, limit, offset)
}

// IsEntityDeclarative checks if an entity is declarative (immutable).
func (s *entityService) IsEntityDeclarative(ctx context.Context, entityID string) (bool, error) {
	return s.store.IsEntityDeclarative(ctx, entityID)
}

// LoadDeclarativeResources loads declarative resources for a given entity category.
// Consumer packages provide parser/validator callbacks for type-specific YAML processing.
func (s *entityService) LoadDeclarativeResources(config DeclarativeLoaderConfig) error {
	return loadDeclarativeResources(s.store, s, config)
}

// UpdateEntityWithCredentials atomically updates an entity and its system credentials.
// If systemCreds is nil, only the entity is updated.
func (s *entityService) UpdateEntityWithCredentials(ctx context.Context, entityID string,
	entity *Entity, systemCreds json.RawMessage) (*Entity, error) {

	var updated Entity
	err := s.transactioner.Transact(ctx, func(txCtx context.Context) error {
		entity.EntityID = entityID
		if err := s.store.UpdateEntity(txCtx, entity); err != nil {
			return err
		}

		if systemCreds != nil {
			if err := s.store.UpdateSystemCredentials(txCtx, entityID, systemCreds); err != nil {
				return err
			}
		}

		result, err := s.store.GetEntity(txCtx, entityID)
		if err != nil {
			return err
		}
		updated = result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

