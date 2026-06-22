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
)

type CompositeDefinitionStoreTestSuite struct {
	suite.Suite
	composite *compositeDefinitionStore
	fileStore *definitionFileBasedStore
	dbStore   *fakeDefinitionStore
	ctx       context.Context
}

func TestCompositeDefinitionStoreTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeDefinitionStoreTestSuite))
}

func (s *CompositeDefinitionStoreTestSuite) SetupTest() {
	fileStore := newDefinitionFileBasedStore()
	s.Require().NoError(fileStore.GenericFileBasedStore.ClearByType())
	s.fileStore = fileStore
	s.dbStore = newFakeDefinitionStore()
	s.composite = newCompositeDefinitionStore(fileStore, s.dbStore)
	s.ctx = context.Background()
}

func (s *CompositeDefinitionStoreTestSuite) seedFile(id, handle, vct string) {
	storer := &definitionStorer{store: s.fileStore}
	s.Require().NoError(storer.Create(id, &PresentationDefinitionDTO{
		ID: id, Handle: handle, VCT: vct, Format: DefaultCredentialFormat,
	}))
}

func (s *CompositeDefinitionStoreTestSuite) seedDB(id, handle, vct string) {
	s.Require().NoError(s.dbStore.Create(s.ctx, PresentationDefinitionDTO{
		ID: id, Handle: handle, VCT: vct, Format: DefaultCredentialFormat,
	}))
}

func (s *CompositeDefinitionStoreTestSuite) TestGetByIDFallsBackToFileStore() {
	s.seedFile("file-1", "file-handle", "v")

	got, err := s.composite.GetByID(s.ctx, "file-1")
	s.Require().NoError(err)
	s.Equal("file-handle", got.Handle)
}

func (s *CompositeDefinitionStoreTestSuite) TestGetByIDPrefersDBStore() {
	s.seedDB("dup", "db-handle", "v-db")
	s.seedFile("dup", "file-handle", "v-file")

	got, err := s.composite.GetByID(s.ctx, "dup")
	s.Require().NoError(err)
	s.Equal("db-handle", got.Handle)
}

func (s *CompositeDefinitionStoreTestSuite) TestGetByIDNotFound() {
	_, err := s.composite.GetByID(s.ctx, "missing")
	s.ErrorIs(err, ErrNotFound)
}

func (s *CompositeDefinitionStoreTestSuite) TestGetByHandleFallsBackToFileStore() {
	s.seedFile("file-1", "file-handle", "v")

	got, err := s.composite.GetByHandle(s.ctx, "file-handle")
	s.Require().NoError(err)
	s.Equal("file-1", got.ID)
}

func (s *CompositeDefinitionStoreTestSuite) TestListMergesAndDeduplicates() {
	s.seedDB("db-1", "db-handle", "v")
	s.seedFile("file-1", "file-handle", "v")
	s.seedDB("dup", "db-dup", "v-db")
	s.seedFile("dup", "file-dup", "v-file")

	defs, err := s.composite.List(s.ctx)
	s.Require().NoError(err)
	s.Len(defs, 3)

	byID := make(map[string]PresentationDefinitionDTO, len(defs))
	for _, d := range defs {
		byID[d.ID] = d
	}
	// DB wins on duplicate ID.
	s.Equal("db-dup", byID["dup"].Handle)
}

func (s *CompositeDefinitionStoreTestSuite) TestCreateRoutesToDBStore() {
	err := s.composite.Create(s.ctx, PresentationDefinitionDTO{ID: "new", Handle: "h", VCT: "v"})
	s.Require().NoError(err)

	_, err = s.dbStore.GetByID(s.ctx, "new")
	s.Require().NoError(err)
}

func (s *CompositeDefinitionStoreTestSuite) TestUpdateRoutesToDBStore() {
	s.seedDB("db-1", "h", "v")

	err := s.composite.Update(s.ctx, PresentationDefinitionDTO{ID: "db-1", Handle: "h", VCT: "v2"})
	s.Require().NoError(err)

	got, _ := s.dbStore.GetByID(s.ctx, "db-1")
	s.Equal("v2", got.VCT)
}

func (s *CompositeDefinitionStoreTestSuite) TestUpdateImmutableDefinition() {
	s.seedFile("file-1", "file-handle", "v")

	err := s.composite.Update(s.ctx, PresentationDefinitionDTO{ID: "file-1", Handle: "x", VCT: "v"})
	s.ErrorIs(err, ErrDefinitionIsImmutable)
}

func (s *CompositeDefinitionStoreTestSuite) TestDeleteRoutesToDBStore() {
	s.seedDB("db-1", "h", "v")

	err := s.composite.Delete(s.ctx, "db-1")
	s.Require().NoError(err)

	_, err = s.dbStore.GetByID(s.ctx, "db-1")
	s.ErrorIs(err, ErrNotFound)
}

func (s *CompositeDefinitionStoreTestSuite) TestDeleteImmutableDefinition() {
	s.seedFile("file-1", "file-handle", "v")

	err := s.composite.Delete(s.ctx, "file-1")
	s.ErrorIs(err, ErrDefinitionIsImmutable)
}
