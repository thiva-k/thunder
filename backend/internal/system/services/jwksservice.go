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

package services

import (
	"net/http"

	"github.com/asgardeo/thunder/internal/oauth/jwks/handler"
	"github.com/asgardeo/thunder/internal/system/server"
)

// JWKSAPIService defines the API service for handling JWKS requests.
type JWKSAPIService struct {
	ServerOpsService server.ServerOperationServiceInterface
	jwksHandler      *handler.JWKSHandler
}

// NewJWKSAPIService creates a new instance of JWKSAPIService.
func NewJWKSAPIService(mux *http.ServeMux) ServiceInterface {
	instance := &JWKSAPIService{
		ServerOpsService: server.NewServerOperationService(),
		jwksHandler:      handler.NewJWKSHandler(),
	}
	instance.RegisterRoutes(mux)

	return instance
}

// RegisterRoutes registers the routes for the JWKSAPIService.
func (s *JWKSAPIService) RegisterRoutes(mux *http.ServeMux) {
	opts := server.RequestWrapOptions{
		Cors: &server.Cors{
			AllowedMethods:   "GET, OPTIONS",
			AllowedHeaders:   "Content-Type, Authorization",
			AllowCredentials: true,
		},
	}
	s.ServerOpsService.WrapHandleFunction(mux, "GET /oauth2/jwks", &opts,
		s.jwksHandler.HandleJWKSRequest)
	s.ServerOpsService.WrapHandleFunction(mux, "OPTIONS /oauth2/jwks", &opts,
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
}
