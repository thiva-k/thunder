// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/internal/ou"
	"github.com/thunder-id/thunderid/internal/system/resourcedependency"
	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
	"github.com/thunder-id/thunderid/tests/mocks/oumock"
	"github.com/thunder-id/thunderid/tests/mocks/transactionmock"
)

const (
	testProjectID = "019a0000-0000-7000-8000-000000000001"
	testRootOUID  = "019a0000-0000-7000-8000-0000000000aa"
)

var testCreatedAt = time.Date(2026, 10, 23, 10, 15, 0, 0, time.UTC)

type ProjectServiceTestSuite struct {
	suite.Suite
	mockStore     *projectStoreInterfaceMock
	mockOUService *oumock.OrganizationUnitServiceInterfaceMock
	service       *projectService
}

func TestProjectServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectServiceTestSuite))
}

func (suite *ProjectServiceTestSuite) SetupTest() {
	suite.mockStore = newProjectStoreInterfaceMock(suite.T())
	suite.mockOUService = oumock.NewOrganizationUnitServiceInterfaceMock(suite.T())
	mockTx := transactionmock.NewTransactionerMock(suite.T())
	mockTx.On("Transact", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }).Maybe()
	suite.service = newProjectService(suite.mockStore, mockTx, suite.mockOUService)
}

func (suite *ProjectServiceTestSuite) record() projectRecord {
	return projectRecord{ID: testProjectID, RootOUID: testRootOUID, CreatedAt: testCreatedAt}
}

func (suite *ProjectServiceTestSuite) rootOU() ou.OrganizationUnit {
	return ou.OrganizationUnit{
		ID: testRootOUID, Handle: "shop", Name: "Shop", Description: "Online shop",
		LogoURL: "https://example.com/logo.png",
	}
}

// expectedProject is the project the record and root OU above describe.
func (suite *ProjectServiceTestSuite) expectedProject() *Project {
	return &Project{
		ID: testProjectID, Handle: "shop", DisplayName: "Shop", Description: "Online shop", RootOUID: testRootOUID,
		CreatedAt: testCreatedAt,
	}
}

func (suite *ProjectServiceTestSuite) TestCreateProject_CreatesRootOUAndRecord() {
	suite.mockOUService.On("CreateOrganizationUnit", mock.Anything, ou.OrganizationUnitRequestWithID{
		Handle: "shop", Name: "Shop", Description: "Online shop",
	}).Return(suite.rootOU(), nil)
	suite.mockStore.On("CreateProject", mock.Anything, mock.MatchedBy(func(r projectRecord) bool {
		return r.ID != "" && r.RootOUID == testRootOUID
	})).Return(nil)

	created, svcErr := suite.service.CreateProject(context.Background(),
		ProjectRequest{Handle: "shop", DisplayName: "Shop", Description: "Online shop"})

	suite.Nil(svcErr)
	suite.NotEmpty(created.ID)
	suite.Equal("shop", created.Handle)
	suite.Equal("Shop", created.DisplayName)
	suite.Equal(testRootOUID, created.RootOUID)
}

func (suite *ProjectServiceTestSuite) TestCreateProject_KeepsSuppliedID() {
	suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(projectRecord{}, errProjectNotFound)
	suite.mockOUService.On("CreateOrganizationUnit", mock.Anything, mock.Anything).Return(suite.rootOU(), nil)
	suite.mockStore.On("CreateProject", mock.Anything, mock.MatchedBy(func(r projectRecord) bool {
		return r.ID == testProjectID && r.RootOUID == testRootOUID && !r.CreatedAt.IsZero()
	})).Return(nil)

	created, svcErr := suite.service.CreateProject(context.Background(),
		ProjectRequest{ID: testProjectID, Handle: "shop", DisplayName: "Shop", Description: "Online shop"})

	suite.Nil(svcErr)
	suite.Equal(testProjectID, created.ID)
	suite.Equal("shop", created.Handle)
	suite.False(created.CreatedAt.IsZero())
}

func (suite *ProjectServiceTestSuite) TestCreateProject_ValidationErrors() {
	cases := []struct {
		name    string
		request ProjectRequest
		wantErr tidcommon.ServiceError
	}{
		{"non uuid id", ProjectRequest{ID: "not-a-uuid", Handle: "shop", DisplayName: "Shop"}, ErrorInvalidProjectID},
		{"blank display name", ProjectRequest{Handle: "shop", DisplayName: "  "}, ErrorInvalidDisplayName},
		{"long display name", ProjectRequest{Handle: "shop", DisplayName: strings.Repeat("a", 101)},
			ErrorInvalidDisplayName},
		{"long description", ProjectRequest{Handle: "shop", DisplayName: "Shop",
			Description: strings.Repeat("a", 256)}, ErrorInvalidDescription},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			created, svcErr := suite.service.CreateProject(context.Background(), tc.request)

			suite.Nil(created)
			suite.Equal(tc.wantErr, *svcErr)
		})
	}
}

func (suite *ProjectServiceTestSuite) TestCreateProject_IDConflict() {
	suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(suite.record(), nil)

	_, svcErr := suite.service.CreateProject(context.Background(),
		ProjectRequest{ID: testProjectID, Handle: "shop", DisplayName: "Shop"})

	suite.Equal(ErrorProjectIDConflict, *svcErr)
}

func (suite *ProjectServiceTestSuite) TestCreateProject_PassesThroughOUError() {
	suite.mockOUService.On("CreateOrganizationUnit", mock.Anything, mock.Anything).
		Return(ou.OrganizationUnit{}, &ou.ErrorOrganizationUnitHandleConflict)

	_, svcErr := suite.service.CreateProject(context.Background(), ProjectRequest{Handle: "shop", DisplayName: "Shop"})

	suite.Equal(ou.ErrorOrganizationUnitHandleConflict, *svcErr)
}

func (suite *ProjectServiceTestSuite) TestCreateProject_StoreErrors() {
	suite.Run("id lookup fails", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(projectRecord{}, errors.New("db"))

		_, svcErr := suite.service.CreateProject(context.Background(),
			ProjectRequest{ID: testProjectID, Handle: "shop", DisplayName: "Shop"})

		suite.Equal(tidcommon.InternalServerError, *svcErr)
	})
	suite.Run("insert fails", func() {
		suite.SetupTest()
		suite.mockOUService.On("CreateOrganizationUnit", mock.Anything, mock.Anything).Return(suite.rootOU(), nil)
		suite.mockStore.On("CreateProject", mock.Anything, mock.Anything).Return(errors.New("db"))

		_, svcErr := suite.service.CreateProject(context.Background(),
			ProjectRequest{Handle: "shop", DisplayName: "Shop"})

		suite.Equal(tidcommon.InternalServerError, *svcErr)
	})
}

func (suite *ProjectServiceTestSuite) TestGetProject() {
	suite.Run("found", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(suite.record(), nil)
		suite.mockOUService.On("GetOrganizationUnit", mock.Anything, testRootOUID).Return(suite.rootOU(), nil)

		project, svcErr := suite.service.GetProject(context.Background(), testProjectID)

		suite.Nil(svcErr)
		suite.Equal(suite.expectedProject(), project)
	})
	suite.Run("not found", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(projectRecord{}, errProjectNotFound)

		_, svcErr := suite.service.GetProject(context.Background(), testProjectID)

		suite.Equal(ErrorProjectNotFound, *svcErr)
	})
	suite.Run("store error", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(projectRecord{}, errors.New("db"))

		_, svcErr := suite.service.GetProject(context.Background(), testProjectID)

		suite.Equal(tidcommon.InternalServerError, *svcErr)
	})
}

func (suite *ProjectServiceTestSuite) TestGetProject_RootOULookupFails() {
	suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(suite.record(), nil)
	suite.mockOUService.On("GetOrganizationUnit", mock.Anything, testRootOUID).
		Return(ou.OrganizationUnit{}, &tidcommon.InternalServerError)

	_, svcErr := suite.service.GetProject(context.Background(), testProjectID)

	suite.Equal(tidcommon.InternalServerError, *svcErr)
}

func (suite *ProjectServiceTestSuite) TestGetProjectList() {
	suite.Run("returns page filled from root OUs", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProjectListCount", mock.Anything).Return(3, nil)
		suite.mockStore.On("GetProjectList", mock.Anything, 2, 0).Return([]projectRecord{suite.record()}, nil)
		suite.mockOUService.On("GetOrganizationUnit", mock.Anything, testRootOUID).Return(suite.rootOU(), nil)

		list, svcErr := suite.service.GetProjectList(context.Background(), 2, 0)

		suite.Nil(svcErr)
		suite.Equal(3, list.TotalResults)
		suite.Equal(1, list.StartIndex)
		suite.Equal([]Project{*suite.expectedProject()}, list.Projects)
		suite.NotEmpty(list.Links)
	})
	suite.Run("invalid limit", func() {
		suite.SetupTest()
		_, svcErr := suite.service.GetProjectList(context.Background(), 0, 0)
		suite.Equal(ErrorInvalidLimit, *svcErr)
	})
	suite.Run("invalid offset", func() {
		suite.SetupTest()
		_, svcErr := suite.service.GetProjectList(context.Background(), 10, -1)
		suite.Equal(ErrorInvalidOffset, *svcErr)
	})
	suite.Run("count fails", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProjectListCount", mock.Anything).Return(0, errors.New("db"))
		_, svcErr := suite.service.GetProjectList(context.Background(), 10, 0)
		suite.Equal(tidcommon.InternalServerError, *svcErr)
	})
	suite.Run("list fails", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProjectListCount", mock.Anything).Return(1, nil)
		suite.mockStore.On("GetProjectList", mock.Anything, 10, 0).Return(nil, errors.New("db"))
		_, svcErr := suite.service.GetProjectList(context.Background(), 10, 0)
		suite.Equal(tidcommon.InternalServerError, *svcErr)
	})
	suite.Run("root OU lookup fails", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProjectListCount", mock.Anything).Return(1, nil)
		suite.mockStore.On("GetProjectList", mock.Anything, 10, 0).Return([]projectRecord{suite.record()}, nil)
		suite.mockOUService.On("GetOrganizationUnit", mock.Anything, testRootOUID).
			Return(ou.OrganizationUnit{}, &tidcommon.InternalServerError)
		_, svcErr := suite.service.GetProjectList(context.Background(), 10, 0)
		suite.Equal(tidcommon.InternalServerError, *svcErr)
	})
}

func (suite *ProjectServiceTestSuite) TestUpdateProject() {
	suite.Run("updates the root OU name and description and keeps everything else", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(suite.record(), nil)
		suite.mockOUService.On("GetOrganizationUnit", mock.Anything, testRootOUID).Return(suite.rootOU(), nil)
		suite.mockOUService.On("UpdateOrganizationUnit", mock.Anything, testRootOUID,
			mock.MatchedBy(func(r ou.OrganizationUnitRequestWithID) bool {
				return r.Name == "Store" && r.Description == "Renamed" && r.Handle == "shop" &&
					r.Parent == nil && r.LogoURL == "https://example.com/logo.png"
			})).
			Return(ou.OrganizationUnit{ID: testRootOUID, Handle: "shop", Name: "Store", Description: "Renamed"}, nil)

		updated, svcErr := suite.service.UpdateProject(context.Background(), testProjectID,
			ProjectRequest{DisplayName: "Store", Description: "Renamed"})

		suite.Nil(svcErr)
		suite.Equal("Store", updated.DisplayName)
		suite.Equal(testRootOUID, updated.RootOUID)
	})
	suite.Run("handle change refused", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(suite.record(), nil)
		suite.mockOUService.On("GetOrganizationUnit", mock.Anything, testRootOUID).Return(suite.rootOU(), nil)

		_, svcErr := suite.service.UpdateProject(context.Background(), testProjectID,
			ProjectRequest{Handle: "other", DisplayName: "Shop"})

		suite.Equal(ErrorHandleChangeNotAllowed, *svcErr)
	})
	suite.Run("invalid display name", func() {
		suite.SetupTest()
		_, svcErr := suite.service.UpdateProject(context.Background(), testProjectID, ProjectRequest{})
		suite.Equal(ErrorInvalidDisplayName, *svcErr)
	})
	suite.Run("not found", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(projectRecord{}, errProjectNotFound)
		_, svcErr := suite.service.UpdateProject(context.Background(), testProjectID,
			ProjectRequest{DisplayName: "Shop"})
		suite.Equal(ErrorProjectNotFound, *svcErr)
	})
	suite.Run("root OU update error passes through", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(suite.record(), nil)
		suite.mockOUService.On("GetOrganizationUnit", mock.Anything, testRootOUID).Return(suite.rootOU(), nil)
		suite.mockOUService.On("UpdateOrganizationUnit", mock.Anything, testRootOUID, mock.Anything).
			Return(ou.OrganizationUnit{}, &ou.ErrorOrganizationUnitNameConflict)
		_, svcErr := suite.service.UpdateProject(context.Background(), testProjectID,
			ProjectRequest{DisplayName: "Taken"})
		suite.Equal(ou.ErrorOrganizationUnitNameConflict, *svcErr)
	})
}

func (suite *ProjectServiceTestSuite) TestDeleteProject() {
	suite.Run("deletes project then root OU", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(suite.record(), nil)
		var order []string
		suite.mockStore.On("DeleteProject", mock.Anything, testProjectID).Return(nil).
			Run(func(mock.Arguments) { order = append(order, "project") })
		suite.mockOUService.On("DeleteOrganizationUnit", mock.Anything, testRootOUID).Return(nil).
			Run(func(mock.Arguments) { order = append(order, "ou") })

		suite.Nil(suite.service.DeleteProject(context.Background(), testProjectID))
		suite.Equal([]string{"project", "ou"}, order)
	})
	suite.Run("missing project succeeds", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(projectRecord{}, errProjectNotFound)

		suite.Nil(suite.service.DeleteProject(context.Background(), testProjectID))
	})
	suite.Run("root OU still in use", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(suite.record(), nil)
		suite.mockStore.On("DeleteProject", mock.Anything, testProjectID).Return(nil)
		suite.mockOUService.On("DeleteOrganizationUnit", mock.Anything, testRootOUID).
			Return(&ou.ErrorCannotDeleteOrganizationUnit)

		svcErr := suite.service.DeleteProject(context.Background(), testProjectID)

		suite.Equal(ou.ErrorCannotDeleteOrganizationUnit, *svcErr)
	})
	suite.Run("lookup fails", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(projectRecord{}, errors.New("db"))

		suite.Equal(tidcommon.InternalServerError, *suite.service.DeleteProject(context.Background(), testProjectID))
	})
	suite.Run("delete fails", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProject", mock.Anything, testProjectID).Return(suite.record(), nil)
		suite.mockStore.On("DeleteProject", mock.Anything, testProjectID).Return(errors.New("db"))

		suite.Equal(tidcommon.InternalServerError, *suite.service.DeleteProject(context.Background(), testProjectID))
	})
}

func (suite *ProjectServiceTestSuite) TestGetResourceDependencies() {
	suite.Run("other resource type", func() {
		suite.SetupTest()
		deps, err := suite.service.GetResourceDependencies(context.Background(),
			resourcedependency.ResourceTypeGroup, testRootOUID)
		suite.NoError(err)
		suite.Empty(deps)
	})
	suite.Run("root OU of a project blocks deletion", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProjectByRootOUID", mock.Anything, testRootOUID).Return(suite.record(), nil)

		deps, err := suite.service.GetResourceDependencies(context.Background(),
			resourcedependency.ResourceTypeOU, testRootOUID)

		suite.NoError(err)
		suite.Equal([]resourcedependency.ResourceDependency{{
			ResourceType:     resourcedependency.ResourceTypeProject,
			ID:               testProjectID,
			DisplayName:      testProjectID,
			BehaviorOnDelete: resourcedependency.BehaviorRestrict,
		}}, deps)
	})
	suite.Run("OU without project", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProjectByRootOUID", mock.Anything, testRootOUID).
			Return(projectRecord{}, errProjectNotFound)

		deps, err := suite.service.GetResourceDependencies(context.Background(),
			resourcedependency.ResourceTypeOU, testRootOUID)

		suite.NoError(err)
		suite.Empty(deps)
	})
	suite.Run("store error", func() {
		suite.SetupTest()
		suite.mockStore.On("GetProjectByRootOUID", mock.Anything, testRootOUID).
			Return(projectRecord{}, errors.New("db"))

		_, err := suite.service.GetResourceDependencies(context.Background(),
			resourcedependency.ResourceTypeOU, testRootOUID)

		suite.Error(err)
	})
}
