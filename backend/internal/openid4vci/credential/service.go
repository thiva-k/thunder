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

package credential

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

// CredentialConfigurationServiceInterface manages OpenID4VCI credential configurations in the configdb
// store, which the issuer engine reads on demand.
type CredentialConfigurationServiceInterface interface {
	Create(ctx context.Context, dto *CredentialConfigurationDTO) (
		*CredentialConfigurationDTO, *serviceerror.ServiceError)
	Get(ctx context.Context, id string) (*CredentialConfigurationDTO, *serviceerror.ServiceError)
	GetByHandle(ctx context.Context, handle string) (*CredentialConfigurationDTO, *serviceerror.ServiceError)
	List(ctx context.Context) ([]CredentialConfigurationDTO, *serviceerror.ServiceError)
	ListSummaries(ctx context.Context) ([]CredentialConfigurationSummary, *serviceerror.ServiceError)
	Update(ctx context.Context, id string, dto *CredentialConfigurationDTO) (
		*CredentialConfigurationDTO, *serviceerror.ServiceError)
	Delete(ctx context.Context, id string) *serviceerror.ServiceError
	IsConfigurationDeclarative(ctx context.Context, id string) (bool, *serviceerror.ServiceError)
}

type configurationService struct {
	store  configurationStoreInterface
	ou     OUResolver
	logger *log.Logger
	uuid   func() (string, error)
}

// newCredentialConfigurationService builds a credential-configuration service over the given store.
func newCredentialConfigurationService(
	store configurationStoreInterface, ouResolver OUResolver,
) CredentialConfigurationServiceInterface {
	return &configurationService{
		store:  store,
		ou:     ouResolver,
		logger: log.GetLogger().With(log.String(log.LoggerKeyComponentName, "OpenID4VCIConfigurationService")),
		uuid:   utils.GenerateUUIDv7,
	}
}

// resolveOU resolves ouHandle to ouId when needed and verifies the OU exists.
func (s *configurationService) resolveOU(
	ctx context.Context, dto *CredentialConfigurationDTO,
) *serviceerror.ServiceError {
	if s.ou == nil {
		return nil
	}
	if dto.OUID == "" && strings.TrimSpace(dto.OUHandle) != "" {
		resolved, svcErr := s.ou.GetOrganizationUnitByPath(ctx, dto.OUHandle)
		if svcErr != nil {
			return &ErrorConfigurationInvalidOU
		}
		dto.OUID = resolved.ID
	}
	if strings.TrimSpace(dto.OUID) == "" {
		return &ErrorConfigurationInvalidOU
	}
	exists, svcErr := s.ou.IsOrganizationUnitExists(ctx, dto.OUID)
	if svcErr != nil {
		s.logger.Error(ctx, "Failed to verify organization unit", log.Any("error", svcErr))
		return &serviceerror.InternalServerError
	}
	if !exists {
		return &ErrorConfigurationInvalidOU
	}
	return nil
}

// populateOUHandle sets each DTO's owning OU handle for display.
func (s *configurationService) populateOUHandle(ctx context.Context, dtos ...*CredentialConfigurationDTO) {
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
		s.logger.Warn(ctx, "Failed to resolve OU handles for credential configurations", log.Any("error", svcErr))
		return
	}
	for _, dto := range dtos {
		if h, ok := handles[dto.OUID]; ok {
			dto.OUHandle = h
		}
	}
}

func (s *configurationService) Create(
	ctx context.Context, dto *CredentialConfigurationDTO,
) (*CredentialConfigurationDTO, *serviceerror.ServiceError) {
	if svcErr := validateConfiguration(dto); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := s.resolveOU(ctx, dto); svcErr != nil {
		return nil, svcErr
	}

	existing, err := s.store.GetByHandle(ctx, dto.Handle)
	if err != nil && !errors.Is(err, ErrNotFound) {
		s.logger.Error(ctx, "Failed to check existing configuration", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	if existing != nil {
		return nil, &ErrorConfigurationAlreadyExists
	}

	id := dto.ID
	if id == "" {
		var genErr error
		id, genErr = s.uuid()
		if genErr != nil {
			s.logger.Error(ctx, "Failed to generate configuration ID", log.Error(genErr))
			return nil, &serviceerror.InternalServerError
		}
	}
	dto.ID = id

	if err := s.store.Create(ctx, *dto); err != nil {
		s.logger.Error(ctx, "Failed to create credential configuration", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	return dto, nil
}

func (s *configurationService) Get(
	ctx context.Context, id string,
) (*CredentialConfigurationDTO, *serviceerror.ServiceError) {
	if strings.TrimSpace(id) == "" {
		return nil, &ErrorConfigurationInvalidRequest
	}
	dto, err := s.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, &ErrorConfigurationNotFound
		}
		s.logger.Error(ctx, "Failed to get credential configuration", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	s.populateOUHandle(ctx, dto)
	return dto, nil
}

func (s *configurationService) GetByHandle(
	ctx context.Context, handle string,
) (*CredentialConfigurationDTO, *serviceerror.ServiceError) {
	dto, err := s.store.GetByHandle(ctx, handle)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, &ErrorConfigurationNotFound
		}
		s.logger.Error(ctx, "Failed to get credential configuration by handle", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	return dto, nil
}

func (s *configurationService) List(
	ctx context.Context,
) ([]CredentialConfigurationDTO, *serviceerror.ServiceError) {
	configs, err := s.store.List(ctx)
	if err != nil {
		if errors.Is(err, ErrResultLimitExceededInCompositeMode) {
			return nil, &ErrorConfigurationResultLimitExceeded
		}
		s.logger.Error(ctx, "Failed to list credential configurations", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	ptrs := make([]*CredentialConfigurationDTO, len(configs))
	for i := range configs {
		ptrs[i] = &configs[i]
	}
	s.populateOUHandle(ctx, ptrs...)
	return configs, nil
}

func (s *configurationService) ListSummaries(
	ctx context.Context,
) ([]CredentialConfigurationSummary, *serviceerror.ServiceError) {
	summaries, err := s.store.ListSummaries(ctx)
	if err != nil {
		if errors.Is(err, ErrResultLimitExceededInCompositeMode) {
			return nil, &ErrorConfigurationResultLimitExceeded
		}
		s.logger.Error(ctx, "Failed to list credential configuration summaries", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	s.populateSummaryOUHandles(ctx, summaries)
	return summaries, nil
}

// populateSummaryOUHandles resolves each summary's owning OU handle for display.
func (s *configurationService) populateSummaryOUHandles(
	ctx context.Context, summaries []CredentialConfigurationSummary,
) {
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
		s.logger.Warn(ctx, "Failed to resolve OU handles for credential configuration summaries",
			log.Any("error", svcErr))
		return
	}
	for i := range summaries {
		if h, ok := handles[summaries[i].OUID]; ok {
			summaries[i].OUHandle = h
		}
	}
}

func (s *configurationService) Update(
	ctx context.Context, id string, dto *CredentialConfigurationDTO,
) (*CredentialConfigurationDTO, *serviceerror.ServiceError) {
	if strings.TrimSpace(id) == "" {
		return nil, &ErrorConfigurationInvalidRequest
	}
	if svcErr := validateConfiguration(dto); svcErr != nil {
		return nil, svcErr
	}
	if svcErr := s.resolveOU(ctx, dto); svcErr != nil {
		return nil, svcErr
	}

	existing, err := s.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, &ErrorConfigurationNotFound
		}
		s.logger.Error(ctx, "Failed to load credential configuration", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}

	if existing.Handle != dto.Handle {
		clash, err := s.store.GetByHandle(ctx, dto.Handle)
		if err != nil && !errors.Is(err, ErrNotFound) {
			s.logger.Error(ctx, "Failed to check handle uniqueness", log.Error(err))
			return nil, &serviceerror.InternalServerError
		}
		if clash != nil {
			return nil, &ErrorConfigurationAlreadyExists
		}
	}

	dto.ID = id
	if err := s.store.Update(ctx, *dto); err != nil {
		if errors.Is(err, ErrConfigurationIsImmutable) {
			return nil, &ErrorConfigurationImmutable
		}
		s.logger.Error(ctx, "Failed to update credential configuration", log.Error(err))
		return nil, &serviceerror.InternalServerError
	}
	return dto, nil
}

func (s *configurationService) Delete(ctx context.Context, id string) *serviceerror.ServiceError {
	if strings.TrimSpace(id) == "" {
		return &ErrorConfigurationInvalidRequest
	}
	if _, err := s.store.GetByID(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil // idempotent
		}
		s.logger.Error(ctx, "Failed to load credential configuration", log.Error(err))
		return &serviceerror.InternalServerError
	}
	if err := s.store.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrConfigurationIsImmutable) {
			return &ErrorConfigurationImmutable
		}
		s.logger.Error(ctx, "Failed to delete credential configuration", log.Error(err))
		return &serviceerror.InternalServerError
	}
	return nil
}

func (s *configurationService) IsConfigurationDeclarative(
	ctx context.Context, id string,
) (bool, *serviceerror.ServiceError) {
	isDeclarative, err := s.store.IsDeclarative(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "Failed to check if credential configuration is declarative", log.Error(err))
		return false, &serviceerror.InternalServerError
	}
	return isDeclarative, nil
}

// validateConfiguration enforces the required fields of a credential configuration.
func validateConfiguration(dto *CredentialConfigurationDTO) *serviceerror.ServiceError {
	if dto == nil || strings.TrimSpace(dto.Handle) == "" || strings.TrimSpace(dto.VCT) == "" {
		return &ErrorConfigurationInvalidRequest
	}
	if dto.Format == "" {
		dto.Format = DefaultCredentialFormat
	}
	if dto.Format != DefaultCredentialFormat {
		return &ErrorConfigurationUnsupportedFormat
	}
	if dto.ValiditySeconds != nil && *dto.ValiditySeconds <= 0 {
		return &ErrorConfigurationInvalidRequest
	}
	return nil
}
