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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/internal/ou"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
)

type DefinitionServiceTestSuite struct {
	suite.Suite
}

func TestDefinitionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(DefinitionServiceTestSuite))
}

// fakeDefinitionStore is an in-memory Store for service tests.
type fakeDefinitionStore struct {
	byID map[string]PresentationDefinitionDTO
}

func newFakeDefinitionStore() *fakeDefinitionStore {
	return &fakeDefinitionStore{byID: map[string]PresentationDefinitionDTO{}}
}

func (f *fakeDefinitionStore) Create(_ context.Context, dto PresentationDefinitionDTO) error {
	f.byID[dto.ID] = dto
	return nil
}

func (f *fakeDefinitionStore) GetByID(_ context.Context, id string) (*PresentationDefinitionDTO, error) {
	dto, ok := f.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &dto, nil
}

func (f *fakeDefinitionStore) GetByHandle(_ context.Context, handle string) (*PresentationDefinitionDTO, error) {
	for _, dto := range f.byID {
		if dto.Handle == handle {
			d := dto
			return &d, nil
		}
	}
	return nil, ErrNotFound
}

func (f *fakeDefinitionStore) List(_ context.Context) ([]PresentationDefinitionDTO, error) {
	out := make([]PresentationDefinitionDTO, 0, len(f.byID))
	for _, dto := range f.byID {
		out = append(out, dto)
	}
	return out, nil
}

func (f *fakeDefinitionStore) ListSummaries(_ context.Context) ([]PresentationDefinitionSummary, error) {
	out := make([]PresentationDefinitionSummary, 0, len(f.byID))
	for _, dto := range f.byID {
		out = append(out, toSummary(dto))
	}
	return out, nil
}

func (f *fakeDefinitionStore) Update(_ context.Context, dto PresentationDefinitionDTO) error {
	f.byID[dto.ID] = dto
	return nil
}

func (f *fakeDefinitionStore) Delete(_ context.Context, id string) error {
	delete(f.byID, id)
	return nil
}

func (f *fakeDefinitionStore) IsDeclarative(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func newTestDefinitionService() (*definitionService, *fakeDefinitionStore) {
	store := newFakeDefinitionStore()
	svc := newPresentationDefinitionService(store, nil).(*definitionService)
	return svc, store
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

func (suite *DefinitionServiceTestSuite) TestDefinitionServiceResolvesAndValidatesOU() {
	resolver := &fakeOUResolver{
		exists:  map[string]bool{"ou-1": true},
		byPath:  map[string]string{"default": "ou-1"},
		handles: map[string]string{"ou-1": "default"},
	}
	svc := newPresentationDefinitionService(newFakeDefinitionStore(), resolver)
	ctx := context.Background()

	_, err := svc.Create(ctx, &PresentationDefinitionDTO{Handle: "eudi-pid", VCT: "urn:eudi:pid:de:1"})
	suite.Require().NotNil(err)
	suite.Equal(ErrorDefinitionInvalidOU.Code, err.Code)

	created, err := svc.Create(ctx, &PresentationDefinitionDTO{
		Handle: "eudi-pid", VCT: "urn:eudi:pid:de:1", OUHandle: "default",
	})
	suite.Require().Nil(err)
	suite.Equal("ou-1", created.OUID)

	got, err := svc.Get(ctx, created.ID)
	suite.Require().Nil(err)
	suite.Equal("ou-1", got.OUID)
	suite.Equal("default", got.OUHandle)
}

func (suite *DefinitionServiceTestSuite) TestDefinitionServiceCreatePersists() {
	svc, store := newTestDefinitionService()
	ctx := context.Background()

	created, svcErr := svc.Create(ctx, &PresentationDefinitionDTO{
		Handle:          "eudi-pid",
		DisplayName:     "EUDI PID",
		VCT:             "urn:eudi:pid:de:1",
		MandatoryClaims: []string{"given_name", "family_name"},
		OptionalClaims:  []string{"birthdate"},
	})
	suite.Require().Nil(svcErr)
	suite.Require().NotEmpty(created.ID)
	suite.Equal(DefaultCredentialFormat, created.Format)

	// The verifier engine resolves definitions from the same store on demand.
	stored, err := store.GetByHandle(ctx, "eudi-pid")
	suite.Require().NoError(err)
	suite.Equal(created.ID, stored.ID)
	suite.Equal("urn:eudi:pid:de:1", stored.VCT)
}

func (suite *DefinitionServiceTestSuite) TestDefinitionServiceCreateValidationAndConflict() {
	svc, _ := newTestDefinitionService()
	ctx := context.Background()

	_, svcErr := svc.Create(ctx, &PresentationDefinitionDTO{Handle: "", VCT: "x"})
	suite.Require().NotNil(svcErr)
	suite.Equal(ErrorDefinitionInvalidRequest.Code, svcErr.Code)

	_, svcErr = svc.Create(ctx, &PresentationDefinitionDTO{Handle: "h", VCT: ""})
	suite.Require().NotNil(svcErr)
	suite.Equal(ErrorDefinitionInvalidRequest.Code, svcErr.Code)

	_, svcErr = svc.Create(ctx, &PresentationDefinitionDTO{Handle: "dup", VCT: "v"})
	suite.Require().Nil(svcErr)
	_, svcErr = svc.Create(ctx, &PresentationDefinitionDTO{Handle: "dup", VCT: "v2"})
	suite.Require().NotNil(svcErr)
	suite.Equal(ErrorDefinitionAlreadyExists.Code, svcErr.Code)
}

func (suite *DefinitionServiceTestSuite) TestDefinitionServiceRejectsUnsupportedFormat() {
	svc, _ := newTestDefinitionService()
	ctx := context.Background()

	// An empty format defaults to the supported SD-JWT VC format.
	created, svcErr := svc.Create(ctx, &PresentationDefinitionDTO{Handle: "default-fmt", VCT: "v"})
	suite.Require().Nil(svcErr)
	suite.Equal(DefaultCredentialFormat, created.Format)

	// An unsupported format is rejected.
	_, svcErr = svc.Create(ctx, &PresentationDefinitionDTO{Handle: "mdoc", VCT: "v", Format: "mso_mdoc"})
	suite.Require().NotNil(svcErr)
	suite.Equal(ErrorDefinitionUnsupportedFormat.Code, svcErr.Code)
}

func (suite *DefinitionServiceTestSuite) TestDefinitionServiceUpdateRehandles() {
	svc, store := newTestDefinitionService()
	ctx := context.Background()

	created, svcErr := svc.Create(ctx, &PresentationDefinitionDTO{Handle: "old", VCT: "v"})
	suite.Require().Nil(svcErr)
	_, err := store.GetByHandle(ctx, "old")
	suite.Require().NoError(err)

	_, svcErr = svc.Update(ctx, created.ID, &PresentationDefinitionDTO{Handle: "new", VCT: "v2"})
	suite.Require().Nil(svcErr)

	_, err = store.GetByHandle(ctx, "old")
	suite.ErrorIs(err, ErrNotFound, "old handle should no longer resolve")
	stored, err := store.GetByHandle(ctx, "new")
	suite.Require().NoError(err)
	suite.Equal("v2", stored.VCT)
}

func (suite *DefinitionServiceTestSuite) TestDefinitionServiceGetUpdateDeleteNotFound() {
	svc, store := newTestDefinitionService()
	ctx := context.Background()

	_, svcErr := svc.Get(ctx, "missing")
	suite.Require().NotNil(svcErr)
	suite.Equal(ErrorDefinitionNotFound.Code, svcErr.Code)

	_, svcErr = svc.Update(ctx, "missing", &PresentationDefinitionDTO{Handle: "h", VCT: "v"})
	suite.Require().NotNil(svcErr)
	suite.Equal(ErrorDefinitionNotFound.Code, svcErr.Code)

	// Delete of a missing definition is idempotent.
	suite.Require().Nil(svc.Delete(ctx, "missing"))

	created, _ := svc.Create(ctx, &PresentationDefinitionDTO{Handle: "todelete", VCT: "v"})
	suite.Require().Nil(svc.Delete(ctx, created.ID))
	_, err := store.GetByHandle(ctx, "todelete")
	suite.ErrorIs(err, ErrNotFound)
}
