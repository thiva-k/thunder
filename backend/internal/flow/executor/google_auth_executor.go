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

package executor

import (
	authngoogle "github.com/asgardeo/thunder/internal/authn/google"
	authnoidc "github.com/asgardeo/thunder/internal/authn/oidc"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/idp"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/internal/userschema"
)

// googleOIDCAuthExecutor implements the OIDC authentication executor for Google.
type googleOIDCAuthExecutor struct {
	oidcAuthExecutorInterface
	googleAuthService authngoogle.GoogleOIDCAuthnServiceInterface
}

var _ flowcore.ExecutorInterface = (*googleOIDCAuthExecutor)(nil)

// newGoogleOIDCAuthExecutor creates a new instance of GoogleOIDCAuthExecutor with the provided details.
func newGoogleOIDCAuthExecutor(
	flowFactory flowcore.FlowFactoryInterface,
	idpService idp.IDPServiceInterface,
	authService authngoogle.GoogleOIDCAuthnServiceInterface,
	userService user.UserServiceInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
) oidcAuthExecutorInterface {
	defaultInputs := []flowcm.InputData{
		{
			Name:     "code",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "nonce",
			Type:     "string",
			Required: false,
		},
	}

	oidcSvcCast, ok := authService.(authnoidc.OIDCAuthnCoreServiceInterface)
	if !ok {
		panic("failed to cast GoogleOIDCAuthnService to OIDCAuthnCoreServiceInterface")
	}

	base := newOIDCAuthExecutor(ExecutorNameGoogleAuth, defaultInputs, []flowcm.InputData{},
		flowFactory, idpService, oidcSvcCast, userService, userSchemaService)

	return &googleOIDCAuthExecutor{
		oidcAuthExecutorInterface: base,
		googleAuthService:         authService,
	}
}
