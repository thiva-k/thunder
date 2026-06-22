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

package definition

import (
	"context"
	"errors"
	"strings"

	"github.com/thunder-id/thunderid/internal/ou"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/internal/system/utils"
)

// OUResolver is the subset of the organization-unit service used to validate the
// owning OU. The full OU service satisfies it; a nil resolver skips validation.
type OUResolver interface {
	IsOrganizationUnitExists(ctx context.Context, id string) (bool, *serviceerror.ServiceError)
	GetOrganizationUnitByPath(ctx context.Context, handlePath string) (ou.OrganizationUnit, *serviceerror.ServiceError)
	GetOrganizationUnitHandlesByIDs(ctx context.Context, ids []string) (map[string]string, *serviceerror.ServiceError)
}

// PresentationDefinitionServiceInterface manages OpenID4VP presentation
// definitions in the configdb store, which the verifier engine reads on demand.
type PresentationDefinitionServiceInterface interface {
	Create(ctx context.Context, dto *PresentationDefinitionDTO) (
		*PresentationDefinitionDTO, *serviceerror.ServiceError)
	Get(ctx context.Context, id string) (*PresentationDefinitionDTO, *serviceerror.ServiceError)
	GetByHandle(ctx context.Context, handle string) (*PresentationDefinitionDTO, *serviceerror.ServiceError)
	List(ctx context.Context) ([]PresentationDefinitionDTO, *serviceerror.ServiceError)
	ListSummaries(ctx context.Context) ([]PresentationDefinitionSummary, *serviceerror.ServiceError)
	Update(ctx context.Context, id string, dto *PresentationDefinitionDTO) (
		*PresentationDefinitionDTO, *serviceerror.ServiceError)
	Delete(ctx context.Context, id string) *serviceerror.ServiceError
	IsDefinitionDeclarative(ctx context.Context, id string) (bool, *serviceerror.ServiceError)
}

type definitionService struct {
	store  definitionStoreInterface
	ou     OUResolver
	logger *log.Logger
	uuid   func() (string, error)
}

// newPresentationDefinitionService builds a presentation-definition service over the given store.
func newPresentationDefinitionService(
	store definitionStoreInterface, ouResolver OUResolver,
) PresentationDefinitionServiceInterface {
	return &definitionService{
		store:  store,
		ou:     ouResolver,
		logger: log.GetLogger().With(log.String(log.LoggerKeyComponentName, "OpenID4VPDefinitionService")),
		uuid:   utils.GenerateUUIDv7,
	}
}

// resolveOU resolves ouHandle to ouId when needed and verifies the OU exists.
func (s *definitionService) resolveOU(
	ctx context.Context, dto *PresentationDefinitionDTO,
) *serviceerror.ServiceError {
	if s.ou == nil {
		return nil
	}
	if dto.OUID == "" && strings.TrimSpace(dto.OUHandle) != "" {
		resolved, svcErr := s.ou.GetOrganizationUnitByPath(ctx, dto.OUHandle)
		if svcErr != nil {
			return &ErrorDefinitionInvalidOU
		}
		dto.OUID = resolved.ID
	}
	if strings.TrimSpace(dto.OUID) == "" {
		return &ErrorDefinitionInvalidOU
	}
	exists, svcErr := s.ou.IsOrganizationUnitExists(ctx, dto.OUID)
	if svcErr != nil {
		s.logger.Error(ctx, "Failed to verify organization unit", log.Any("error", svcErr))
		return &serviceerror.InternalServerError
	}
	if !exists {
		return &ErrorDefinitionInvalidOU
	}
	return nil
}

// populateOUHandle sets each DTO's owning OU handle for display.
func (s *definitionService) populateOUHandle(ctx context.Context, dtos ...*PresentationDefinitionDTO) {
	if s.ou == nil {
		return
	}
	ids := make([]string, 0, len(dtos))
	seen := make(map[string]bool, len(dtos))
	for _, dto := range dtos {
		if dto.OUID != "" && !seen[dto.OUID] {
			seen[dto.OUID] = true
			ids = append(ids, dto.OUID)
		}
	}
	if len(ids) == 0 {
		return
	}
	handles, svcErr := s.ou.GetOrganizationUnitHandlesByIDs(ctx, ids)
	if svcErr != nil {
		s.logger.Warn(ctx, "Failed to resolve OU handles for presentation definitions", log.Any("error", svcErr))
		return
	}
	for _, dto := range dtos {
		if h, ok := handles[dto.OUID]; ok {
			dto.OUHandle = h
		}
	}
}

func (s *definitionService) Create(
	ctx context.Context, dto *PresentationDefinitionDTO,
) (*PresentationDefinitionDTO, *serviceerror.ServiceError) {
	if svcErr := validateDefinition(dto); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := s.resolveOU(ctx, dto); svcErr != nil {
		return nil, svcErr
	}

	existing, err := s.store.GetByHandle(ctx, dto.Handle)
	if err != nil && !errors.Is(err, ErrNotFound) {
		s.logger.Error(ctx, "Failed to check existing definition", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if existing != nil {
		return nil, &ErrorDefinitionAlreadyExists
	}

	id := dto.ID
	if id == "" {
		var genErr error
		id, genErr = s.uuid()
		if genErr != nil {
			s.logger.Error(ctx, "Failed to generate definition ID", log.Error(genErr))
			return nil, &serviceerror.InternalServerError
		}
	}
	dto.ID = id

	if err := s.store.Create(ctx, *dto); err != nil {
		s.logger.Error(ctx, "Failed to create presentation definition", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	return dto, nil
}

func (s *definitionService) Get(
	ctx context.Context, id string,
) (*PresentationDefinitionDTO, *serviceerror.ServiceError) {
	if strings.TrimSpace(id) == "" {
		return nil, &ErrorDefinitionInvalidRequest
	}
	dto, err := s.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, &ErrorDefinitionNotFound
		}
		s.logger.Error(ctx, "Failed to get presentation definition", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	s.populateOUHandle(ctx, dto)
	return dto, nil
}

func (s *definitionService) GetByHandle(
	ctx context.Context, handle string,
) (*PresentationDefinitionDTO, *serviceerror.ServiceError) {
	dto, err := s.store.GetByHandle(ctx, handle)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, &ErrorDefinitionNotFound
		}
		s.logger.Error(ctx, "Failed to get presentation definition by handle", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	return dto, nil
}

func (s *definitionService) List(
	ctx context.Context,
) ([]PresentationDefinitionDTO, *serviceerror.ServiceError) {
	defs, err := s.store.List(ctx)
	if err != nil {
		if errors.Is(err, ErrResultLimitExceededInCompositeMode) {
			return nil, &ErrorDefinitionResultLimitExceeded
		}
		s.logger.Error(ctx, "Failed to list presentation definitions", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	ptrs := make([]*PresentationDefinitionDTO, len(defs))
	for i := range defs {
		ptrs[i] = &defs[i]
	}
	s.populateOUHandle(ctx, ptrs...)
	return defs, nil
}

func (s *definitionService) ListSummaries(
	ctx context.Context,
) ([]PresentationDefinitionSummary, *serviceerror.ServiceError) {
	summaries, err := s.store.ListSummaries(ctx)
	if err != nil {
		if errors.Is(err, ErrResultLimitExceededInCompositeMode) {
			return nil, &ErrorDefinitionResultLimitExceeded
		}
		s.logger.Error(ctx, "Failed to list presentation definition summaries", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	s.populateSummaryOUHandles(ctx, summaries)
	return summaries, nil
}

// populateSummaryOUHandles resolves each summary's owning OU handle for display.
func (s *definitionService) populateSummaryOUHandles(ctx context.Context, summaries []PresentationDefinitionSummary) {
	if s.ou == nil {
		return
	}
	ids := make([]string, 0, len(summaries))
	seen := make(map[string]bool, len(summaries))
	for _, sm := range summaries {
		if sm.OUID != "" && !seen[sm.OUID] {
			seen[sm.OUID] = true
			ids = append(ids, sm.OUID)
		}
	}
	if len(ids) == 0 {
		return
	}
	handles, svcErr := s.ou.GetOrganizationUnitHandlesByIDs(ctx, ids)
	if svcErr != nil {
		s.logger.Warn(ctx, "Failed to resolve OU handles for presentation definition summaries",
			log.Any("error", svcErr))
		return
	}
	for i := range summaries {
		if h, ok := handles[summaries[i].OUID]; ok {
			summaries[i].OUHandle = h
		}
	}
}

func (s *definitionService) Update(
	ctx context.Context, id string, dto *PresentationDefinitionDTO,
) (*PresentationDefinitionDTO, *serviceerror.ServiceError) {
	if strings.TrimSpace(id) == "" {
		return nil, &ErrorDefinitionInvalidRequest
	}
	if svcErr := validateDefinition(dto); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := s.resolveOU(ctx, dto); svcErr != nil {
		return nil, svcErr
	}

	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, &ErrorDefinitionNotFound
		}
		s.logger.Error(ctx, "Failed to load presentation definition", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if existing.Handle != dto.Handle {
		clash, err := s.store.GetByHandle(ctx, dto.Handle)
		if err != nil && !errors.Is(err, ErrNotFound) {
			s.logger.Error(ctx, "Failed to check handle uniqueness", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		if clash != nil {
			return nil, &ErrorDefinitionAlreadyExists
		}
	}

	dto.ID = id
	if err := s.store.Update(ctx, *dto); err != nil {
		if errors.Is(err, ErrDefinitionIsImmutable) {
			return nil, &ErrorDefinitionImmutable
		}
		s.logger.Error(ctx, "Failed to update presentation definition", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	return dto, nil
}

func (s *definitionService) Delete(ctx context.Context, id string) *serviceerror.ServiceError {
	if strings.TrimSpace(id) == "" {
		return &ErrorDefinitionInvalidRequest
	}
	if _, err := s.store.GetByID(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil // idempotent
		}
		s.logger.Error(ctx, "Failed to load presentation definition", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if err := s.store.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrDefinitionIsImmutable) {
			return &ErrorDefinitionImmutable
		}
		s.logger.Error(ctx, "Failed to delete presentation definition", log.Error(err))
		return &serviceerror.InternalServerError
	}
	return nil
}

func (s *definitionService) IsDefinitionDeclarative(
	ctx context.Context, id string,
) (bool, *serviceerror.ServiceError) {
	isDeclarative, err := s.store.IsDeclarative(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to check if presentation definition is declarative", log.Error(err))
		return false, &serviceerror.InternalServerError
	}
	return isDeclarative, nil
}

// validateDefinition enforces the required fields of a presentation definition.
func validateDefinition(dto *PresentationDefinitionDTO) *serviceerror.ServiceError {
	if dto == nil || strings.TrimSpace(dto.Handle) == "" || strings.TrimSpace(dto.VCT) == "" {
		return &ErrorDefinitionInvalidRequest
	}
	if dto.Format == "" {
		dto.Format = DefaultCredentialFormat
	}
	if dto.Format != DefaultCredentialFormat {
		return &ErrorDefinitionUnsupportedFormat
	}
	return nil
}
