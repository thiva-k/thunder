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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/internal/ou"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
)

type ConfigurationServiceTestSuite struct {
	suite.Suite
}

func TestConfigurationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationServiceTestSuite))
}

// fakeStore is an in-memory Store for service tests.
type fakeStore struct {
	byID map[string]CredentialConfigurationDTO
}

func newFakeStore() *fakeStore {
	return &fakeStore{byID: map[string]CredentialConfigurationDTO{}}
}

func (f *fakeStore) Create(_ context.Context, dto CredentialConfigurationDTO) error {
	f.byID[dto.ID] = dto
	return nil
}

func (f *fakeStore) GetByID(_ context.Context, id string) (*CredentialConfigurationDTO, error) {
	dto, ok := f.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &dto, nil
}

func (f *fakeStore) GetByHandle(_ context.Context, handle string) (*CredentialConfigurationDTO, error) {
	for _, dto := range f.byID {
		if dto.Handle == handle {
			d := dto
			return &d, nil
		}
	}
	return nil, ErrNotFound
}

func (f *fakeStore) List(_ context.Context) ([]CredentialConfigurationDTO, error) {
	out := make([]CredentialConfigurationDTO, 0, len(f.byID))
	for _, dto := range f.byID {
		out = append(out, dto)
	}
	return out, nil
}

func (f *fakeStore) ListSummaries(_ context.Context) ([]CredentialConfigurationSummary, error) {
	out := make([]CredentialConfigurationSummary, 0, len(f.byID))
	for _, dto := range f.byID {
		out = append(out, toConfigSummary(dto))
	}
	return out, nil
}

func (f *fakeStore) Update(_ context.Context, dto CredentialConfigurationDTO) error {
	f.byID[dto.ID] = dto
	return nil
}

func (f *fakeStore) Delete(_ context.Context, id string) error {
	delete(f.byID, id)
	return nil
}

func (f *fakeStore) IsDeclarative(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (s *ConfigurationServiceTestSuite) newService() CredentialConfigurationServiceInterface {
	return newCredentialConfigurationService(newFakeStore(), nil)
}

// fakeOUResolver is a test OUResolver where every queried OU exists.
type fakeOUResolver struct {
	exists  map[string]bool
	byPath  map[string]string
	handles map[string]string
}

func (f *fakeOUResolver) IsOrganizationUnitExists(
	_ context.Context, id string,
) (bool, *serviceerror.ServiceError) {
	return f.exists[id], nil
}

func (f *fakeOUResolver) GetOrganizationUnitByPath(
	_ context.Context, handlePath string,
) (ou.OrganizationUnit, *serviceerror.ServiceError) {
	id, ok := f.byPath[handlePath]
	if !ok {
		return ou.OrganizationUnit{}, &serviceerror.InternalServerError
	}
	return ou.OrganizationUnit{ID: id}, nil
}

func (f *fakeOUResolver) GetOrganizationUnitHandlesByIDs(
	_ context.Context, ids []string,
) (map[string]string, *serviceerror.ServiceError) {
	out := make(map[string]string, len(ids))
	for _, id := range ids {
		if h, ok := f.handles[id]; ok {
			out[id] = h
		}
	}
	return out, nil
}

func (s *ConfigurationServiceTestSuite) validDTO() *CredentialConfigurationDTO {
	return &CredentialConfigurationDTO{
		Handle: "eudi-pid",
		VCT:    "urn:eudi:pid:de:1",
		Claims: []ClaimMapping{{Name: "given_name", DisplayName: "Given Name"}},
	}
}

func (s *ConfigurationServiceTestSuite) TestCreateDefaultsFormatAndAssignsID() {
	svc := s.newService()
	created, err := svc.Create(context.Background(), s.validDTO())
	s.Nil(err)
	s.NotEmpty(created.ID)
	s.Equal(DefaultCredentialFormat, created.Format)
}

func (s *ConfigurationServiceTestSuite) TestCreateRejectsMissingFields() {
	svc := s.newService()
	_, err := svc.Create(context.Background(), &CredentialConfigurationDTO{Handle: "x"})
	s.NotNil(err)
	s.Equal(ErrorConfigurationInvalidRequest.Code, err.Code)
}

func (s *ConfigurationServiceTestSuite) TestCreateRejectsUnsupportedFormat() {
	svc := s.newService()
	dto := s.validDTO()
	dto.Format = "mso_mdoc"
	_, err := svc.Create(context.Background(), dto)
	s.NotNil(err)
	s.Equal(ErrorConfigurationUnsupportedFormat.Code, err.Code)
}

func (s *ConfigurationServiceTestSuite) TestCreateRejectsDuplicateHandle() {
	svc := s.newService()
	_, err := svc.Create(context.Background(), s.validDTO())
	s.Nil(err)
	_, err = svc.Create(context.Background(), s.validDTO())
	s.NotNil(err)
	s.Equal(ErrorConfigurationAlreadyExists.Code, err.Code)
}

func (s *ConfigurationServiceTestSuite) TestGetByHandleAndDelete() {
	svc := s.newService()
	created, err := svc.Create(context.Background(), s.validDTO())
	s.Nil(err)

	got, err := svc.GetByHandle(context.Background(), "eudi-pid")
	s.Nil(err)
	s.Equal(created.ID, got.ID)

	s.Nil(svc.Delete(context.Background(), created.ID))
	_, err = svc.Get(context.Background(), created.ID)
	s.NotNil(err)
	s.Equal(ErrorConfigurationNotFound.Code, err.Code)
}

func (s *ConfigurationServiceTestSuite) TestDeleteIsIdempotent() {
	svc := s.newService()
	s.Nil(svc.Delete(context.Background(), "missing"))
}

func (s *ConfigurationServiceTestSuite) TestCreateResolvesAndValidatesOU() {
	resolver := &fakeOUResolver{
		exists:  map[string]bool{"ou-1": true},
		byPath:  map[string]string{"default": "ou-1"},
		handles: map[string]string{"ou-1": "default"},
	}
	svc := newCredentialConfigurationService(newFakeStore(), resolver)

	dto := s.validDTO()
	_, err := svc.Create(context.Background(), dto)
	s.Require().NotNil(err)
	s.Equal(ErrorConfigurationInvalidOU.Code, err.Code)

	dto = s.validDTO()
	dto.OUHandle = "default"
	created, err := svc.Create(context.Background(), dto)
	s.Require().Nil(err)
	s.Equal("ou-1", created.OUID)

	got, err := svc.Get(context.Background(), created.ID)
	s.Require().Nil(err)
	s.Equal("ou-1", got.OUID)
	s.Equal("default", got.OUHandle)
}
