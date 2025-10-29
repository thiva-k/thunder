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

package discovery

import (
	"encoding/json"
	"net/http"

	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/system/utils"
)

// DiscoveryHandlerInterface defines the interface for discovery handlers
type discoveryHandlerInterface interface {
	HandleOAuth2AuthorizationServerMetadata(w http.ResponseWriter, r *http.Request)
	HandleOIDCDiscovery(w http.ResponseWriter, r *http.Request)
}

// discoveryHandler implements DiscoveryHandlerInterface
type discoveryHandler struct {
	discoveryService DiscoveryServiceInterface
}

// NewDiscoveryHandler creates a new discovery handler
func newDiscoveryHandler(discoveryService DiscoveryServiceInterface) discoveryHandlerInterface {
	return &discoveryHandler{
		discoveryService: discoveryService,
	}
}

// HandleOAuth2AuthorizationServerMetadata handles OAuth 2.0 Authorization Server Metadata requests
func (dh *discoveryHandler) HandleOAuth2AuthorizationServerMetadata(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DiscoveryHandler"))

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, "method_not_allowed",
			"Only GET method is allowed", http.StatusMethodNotAllowed, nil)
		return
	}

	metadata := dh.discoveryService.GetOAuth2AuthorizationServerMetadata()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		logger.Error("Failed to encode OAuth 2.0 Authorization Server Metadata response", log.Error(err))
		utils.WriteJSONError(w, "server_error",
			"Failed to generate discovery response", http.StatusInternalServerError, nil)
		return
	}
}

// HandleOIDCDiscovery handles OpenID Connect Discovery requests
func (dh *discoveryHandler) HandleOIDCDiscovery(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DiscoveryHandler"))

	if r.Method != http.MethodGet {
		utils.WriteJSONError(w, "method_not_allowed",
			"Only GET method is allowed", http.StatusMethodNotAllowed, nil)
		return
	}

	metadata := dh.discoveryService.GetOIDCMetadata()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		logger.Error("Failed to encode OIDC discovery response", log.Error(err))
		utils.WriteJSONError(w, "server_error",
			"Failed to generate discovery response", http.StatusInternalServerError, nil)
		return
	}
}
