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

package idp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/asgardeo/thunder/internal/system/config"
)

type IDPInitTestSuite struct {
	suite.Suite
}

func TestIDPInitTestSuite(t *testing.T) {
	suite.Run(t, new(IDPInitTestSuite))
}

func (s *IDPInitTestSuite) TestInitialize() {
	// Initialize runtime config for the test
	testConfig := &config.Config{
		Database: config.DatabaseConfig{
			Identity: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
			Runtime: config.DataSource{
				Type: "sqlite",
				Path: ":memory:",
			},
		},
	}
	_ = config.InitializeThunderRuntime("test", testConfig)
	mux := http.NewServeMux()

	service := Initialize(mux)
	s.NotNil(service)
	s.Implements((*IDPServiceInterface)(nil), service)
}

func (s *IDPInitTestSuite) TestRegisterRoutes() {
	mux := http.NewServeMux()
	handler := &idpHandler{}

	// This test mainly ensures registerRoutes doesn't panic
	s.NotPanics(func() {
		registerRoutes(mux, handler)
	})

	// Verify expected routes are registered on the mux without invoking handlers
	cases := []struct {
		method   string
		target   string
		expected string
	}{
		{method: http.MethodPost, target: "/identity-providers", expected: "POST /identity-providers"},
		{method: http.MethodGet, target: "/identity-providers", expected: "GET /identity-providers"},
		{method: http.MethodOptions, target: "/identity-providers", expected: "OPTIONS /identity-providers"},
		{method: http.MethodGet, target: "/identity-providers/123", expected: "GET /identity-providers/{id}"},
		{method: http.MethodPut, target: "/identity-providers/123", expected: "PUT /identity-providers/{id}"},
		{method: http.MethodDelete, target: "/identity-providers/123", expected: "DELETE /identity-providers/{id}"},
		{method: http.MethodOptions, target: "/identity-providers/123", expected: "OPTIONS /identity-providers/{id}"},
	}

	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.target, nil)
		_, pattern := mux.Handler(req)
		s.Equal(c.expected, pattern)
	}
}

func (s *IDPInitTestSuite) TestNewIDPHandler() {
	service := &idpService{}
	handler := newIDPHandler(service)

	s.NotNil(handler)
	s.Equal(service, handler.idpService)
}

func (s *IDPInitTestSuite) TestNewIDPService() {
	store := &idpStore{}
	service := newIDPService(store)

	s.NotNil(service)
	s.Implements((*IDPServiceInterface)(nil), service)

	// Verify store is set correctly
	idpSvc, ok := service.(*idpService)
	s.True(ok)
	s.Equal(store, idpSvc.idpStore)
}
