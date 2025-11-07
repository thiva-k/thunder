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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	oauth2const "github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
)

const (
	errRedirectURIFragment          = "redirect URI must not contain a fragment component"
	errRedirectURINotRegistered     = "your application's redirect URL does not match with the registered redirect URLs"
	errRedirectURIRequired          = "redirect URI is required in the authorization request"
	errRedirectURINotFullyQualified = "registered redirect URI is not fully qualified"
)

type OAuthAppConfigDTOTestSuite struct {
	suite.Suite
}

func TestOAuthAppConfigDTOTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthAppConfigDTOTestSuite))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_AuthorizationCode() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
			oauth2const.GrantTypeRefreshToken,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_ClientCredentials() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeClientCredentials,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeClientCredentials))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_RefreshToken() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeRefreshToken,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeRefreshToken))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_TokenExchange() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeTokenExchange,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeTokenExchange))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_NotAllowed() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
		},
	}
	assert.False(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeClientCredentials))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_EmptyGrantType() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
		},
	}
	assert.False(suite.T(), config.IsAllowedGrantType(""))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_EmptyGrantTypesList() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{},
	}
	assert.False(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_NilGrantTypesList() {
	config := &OAuthAppConfigDTO{
		GrantTypes: nil,
	}
	assert.False(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedGrantType_MultipleGrantTypes() {
	config := &OAuthAppConfigDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
			oauth2const.GrantTypeClientCredentials,
			oauth2const.GrantTypeRefreshToken,
			oauth2const.GrantTypeTokenExchange,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeClientCredentials))
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeRefreshToken))
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeTokenExchange))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_Code() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
		},
	}
	assert.True(suite.T(), config.IsAllowedResponseType("code"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_NotAllowed() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
		},
	}
	assert.False(suite.T(), config.IsAllowedResponseType("token"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_EmptyResponseType() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
		},
	}
	assert.False(suite.T(), config.IsAllowedResponseType(""))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_EmptyResponseTypesList() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{},
	}
	assert.False(suite.T(), config.IsAllowedResponseType("code"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_NilResponseTypesList() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: nil,
	}
	assert.False(suite.T(), config.IsAllowedResponseType("code"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedResponseType_MultipleResponseTypes() {
	config := &OAuthAppConfigDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
			"token",
			"id_token",
		},
	}
	assert.True(suite.T(), config.IsAllowedResponseType("code"))
	assert.True(suite.T(), config.IsAllowedResponseType("token"))
	assert.True(suite.T(), config.IsAllowedResponseType("id_token"))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_ClientSecretBasic() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}
	assert.True(suite.T(), config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodClientSecretBasic))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_ClientSecretPost() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
	}
	assert.True(suite.T(), config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodClientSecretPost))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_None() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
	}
	assert.True(suite.T(), config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodNone))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_NotAllowed() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}
	assert.False(suite.T(), config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodClientSecretPost))
}

func (suite *OAuthAppConfigDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_Empty() {
	config := &OAuthAppConfigDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}
	assert.False(suite.T(), config.IsAllowedTokenEndpointAuthMethod(""))
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidWithSingleRegisteredURI() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback"},
	}
	err := config.ValidateRedirectURI("https://example.com/callback")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidHTTPLocalhostWithPort() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"http://localhost:3000/callback"},
	}
	err := config.ValidateRedirectURI("http://localhost:3000/callback")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidHTTPSWithPath() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://app.example.com/oauth/callback"},
	}
	err := config.ValidateRedirectURI("https://app.example.com/oauth/callback")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidCustomScheme() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"myapp://callback"},
	}
	err := config.ValidateRedirectURI("myapp://callback")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_ValidWithQueryParameters() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback?param=value"},
	}
	err := config.ValidateRedirectURI("https://example.com/callback?param=value")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_InvalidWithFragment() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback#fragment"},
	}
	err := config.ValidateRedirectURI("https://example.com/callback#fragment")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIFragment, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_NotRegistered() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback"},
	}
	err := config.ValidateRedirectURI("https://different.com/callback")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotRegistered, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyWithSingleFullyQualifiedURI() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"https://example.com/callback"},
	}
	err := config.ValidateRedirectURI("")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyWithMultipleURIs() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{
			"https://example.com/callback",
			"https://example.com/callback2",
		},
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIRequired, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyWithPartialRegisteredURI() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"/callback"},
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotFullyQualified, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyWithInvalidRegisteredURI() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{"://invalid"},
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotFullyQualified, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_EmptyRedirectURIsList() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: []string{},
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIRequired, err.Error())
}

func (suite *OAuthAppConfigDTOTestSuite) TestValidateRedirectURI_NilRedirectURIsList() {
	config := &OAuthAppConfigDTO{
		RedirectURIs: nil,
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIRequired, err.Error())
}

type OAuthAppConfigProcessedDTOTestSuite struct {
	suite.Suite
}

func TestOAuthAppConfigProcessedDTOTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthAppConfigProcessedDTOTestSuite))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedGrantType_AuthorizationCode() {
	config := &OAuthAppConfigProcessedDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
			oauth2const.GrantTypeRefreshToken,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedGrantType_ClientCredentials() {
	config := &OAuthAppConfigProcessedDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeClientCredentials,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeClientCredentials))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedGrantType_RefreshToken() {
	config := &OAuthAppConfigProcessedDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeRefreshToken,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeRefreshToken))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedGrantType_TokenExchange() {
	config := &OAuthAppConfigProcessedDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeTokenExchange,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeTokenExchange))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedGrantType_NotAllowed() {
	config := &OAuthAppConfigProcessedDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
		},
	}
	assert.False(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeClientCredentials))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedGrantType_EmptyGrantType() {
	config := &OAuthAppConfigProcessedDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
		},
	}
	assert.False(suite.T(), config.IsAllowedGrantType(""))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedGrantType_EmptyGrantTypesList() {
	config := &OAuthAppConfigProcessedDTO{
		GrantTypes: []oauth2const.GrantType{},
	}
	assert.False(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedGrantType_NilGrantTypesList() {
	config := &OAuthAppConfigProcessedDTO{
		GrantTypes: nil,
	}
	assert.False(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedGrantType_MultipleGrantTypes() {
	config := &OAuthAppConfigProcessedDTO{
		GrantTypes: []oauth2const.GrantType{
			oauth2const.GrantTypeAuthorizationCode,
			oauth2const.GrantTypeClientCredentials,
			oauth2const.GrantTypeRefreshToken,
			oauth2const.GrantTypeTokenExchange,
		},
	}
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeAuthorizationCode))
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeClientCredentials))
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeRefreshToken))
	assert.True(suite.T(), config.IsAllowedGrantType(oauth2const.GrantTypeTokenExchange))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedResponseType_Code() {
	config := &OAuthAppConfigProcessedDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
		},
	}
	assert.True(suite.T(), config.IsAllowedResponseType("code"))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedResponseType_NotAllowed() {
	config := &OAuthAppConfigProcessedDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
		},
	}
	assert.False(suite.T(), config.IsAllowedResponseType("token"))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedResponseType_EmptyResponseType() {
	config := &OAuthAppConfigProcessedDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
		},
	}
	assert.False(suite.T(), config.IsAllowedResponseType(""))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedResponseType_EmptyResponseTypesList() {
	config := &OAuthAppConfigProcessedDTO{
		ResponseTypes: []oauth2const.ResponseType{},
	}
	assert.False(suite.T(), config.IsAllowedResponseType("code"))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedResponseType_NilResponseTypesList() {
	config := &OAuthAppConfigProcessedDTO{
		ResponseTypes: nil,
	}
	assert.False(suite.T(), config.IsAllowedResponseType("code"))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedResponseType_MultipleResponseTypes() {
	config := &OAuthAppConfigProcessedDTO{
		ResponseTypes: []oauth2const.ResponseType{
			oauth2const.ResponseTypeCode,
			"token",
			"id_token",
		},
	}
	assert.True(suite.T(), config.IsAllowedResponseType("code"))
	assert.True(suite.T(), config.IsAllowedResponseType("token"))
	assert.True(suite.T(), config.IsAllowedResponseType("id_token"))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_ClientSecretBasic() {
	config := &OAuthAppConfigProcessedDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}
	assert.True(suite.T(), config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodClientSecretBasic))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_ClientSecretPost() {
	config := &OAuthAppConfigProcessedDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretPost,
	}
	assert.True(suite.T(), config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodClientSecretPost))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_None() {
	config := &OAuthAppConfigProcessedDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodNone,
	}
	assert.True(suite.T(), config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodNone))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_NotAllowed() {
	config := &OAuthAppConfigProcessedDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}
	assert.False(suite.T(), config.IsAllowedTokenEndpointAuthMethod(oauth2const.TokenEndpointAuthMethodClientSecretPost))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestIsAllowedTokenEndpointAuthMethod_Empty() {
	config := &OAuthAppConfigProcessedDTO{
		TokenEndpointAuthMethod: oauth2const.TokenEndpointAuthMethodClientSecretBasic,
	}
	assert.False(suite.T(), config.IsAllowedTokenEndpointAuthMethod(""))
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_ValidWithSingleRegisteredURI() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"https://example.com/callback"},
	}
	err := config.ValidateRedirectURI("https://example.com/callback")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_ValidHTTPLocalhostWithPort() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"http://localhost:3000/callback"},
	}
	err := config.ValidateRedirectURI("http://localhost:3000/callback")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_ValidHTTPSWithPath() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"https://app.example.com/oauth/callback"},
	}
	err := config.ValidateRedirectURI("https://app.example.com/oauth/callback")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_ValidCustomScheme() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"myapp://callback"},
	}
	err := config.ValidateRedirectURI("myapp://callback")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_ValidWithQueryParameters() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"https://example.com/callback?param=value"},
	}
	err := config.ValidateRedirectURI("https://example.com/callback?param=value")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_InvalidWithFragment() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"https://example.com/callback#fragment"},
	}
	err := config.ValidateRedirectURI("https://example.com/callback#fragment")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIFragment, err.Error())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_NotRegistered() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"https://example.com/callback"},
	}
	err := config.ValidateRedirectURI("https://different.com/callback")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotRegistered, err.Error())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_EmptyWithSingleFullyQualifiedURI() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"https://example.com/callback"},
	}
	err := config.ValidateRedirectURI("")
	assert.NoError(suite.T(), err)
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_EmptyWithMultipleURIs() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{
			"https://example.com/callback",
			"https://example.com/callback2",
		},
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIRequired, err.Error())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_EmptyWithPartialRegisteredURI() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"/callback"},
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotFullyQualified, err.Error())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_EmptyWithInvalidRegisteredURI() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{"://invalid"},
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotFullyQualified, err.Error())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_EmptyRedirectURIsList() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: []string{},
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIRequired, err.Error())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestValidateRedirectURI_NilRedirectURIsList() {
	config := &OAuthAppConfigProcessedDTO{
		RedirectURIs: nil,
	}
	err := config.ValidateRedirectURI("")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIRequired, err.Error())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestRequiresPKCE_PKCERequiredTrue() {
	config := &OAuthAppConfigProcessedDTO{
		PKCERequired: true,
		PublicClient: false,
	}
	assert.True(suite.T(), config.RequiresPKCE())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestRequiresPKCE_PublicClientTrue() {
	config := &OAuthAppConfigProcessedDTO{
		PKCERequired: false,
		PublicClient: true,
	}
	assert.True(suite.T(), config.RequiresPKCE())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestRequiresPKCE_BothTrue() {
	config := &OAuthAppConfigProcessedDTO{
		PKCERequired: true,
		PublicClient: true,
	}
	assert.True(suite.T(), config.RequiresPKCE())
}

func (suite *OAuthAppConfigProcessedDTOTestSuite) TestRequiresPKCE_BothFalse() {
	config := &OAuthAppConfigProcessedDTO{
		PKCERequired: false,
		PublicClient: false,
	}
	assert.False(suite.T(), config.RequiresPKCE())
}

type HelperFunctionsTestSuite struct {
	suite.Suite
}

func TestHelperFunctionsTestSuite(t *testing.T) {
	suite.Run(t, new(HelperFunctionsTestSuite))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedGrantType_ValidGrantType() {
	grantTypes := []oauth2const.GrantType{
		oauth2const.GrantTypeAuthorizationCode,
		oauth2const.GrantTypeRefreshToken,
	}
	assert.True(suite.T(), isAllowedGrantType(grantTypes, oauth2const.GrantTypeAuthorizationCode))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedGrantType_InvalidGrantType() {
	grantTypes := []oauth2const.GrantType{
		oauth2const.GrantTypeAuthorizationCode,
	}
	assert.False(suite.T(), isAllowedGrantType(grantTypes, oauth2const.GrantTypeClientCredentials))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedGrantType_EmptyGrantType() {
	grantTypes := []oauth2const.GrantType{
		oauth2const.GrantTypeAuthorizationCode,
	}
	assert.False(suite.T(), isAllowedGrantType(grantTypes, ""))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedGrantType_EmptyList() {
	grantTypes := []oauth2const.GrantType{}
	assert.False(suite.T(), isAllowedGrantType(grantTypes, oauth2const.GrantTypeAuthorizationCode))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedGrantType_NilList() {
	assert.False(suite.T(), isAllowedGrantType(nil, oauth2const.GrantTypeAuthorizationCode))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedResponseType_ValidResponseType() {
	responseTypes := []oauth2const.ResponseType{
		oauth2const.ResponseTypeCode,
		"token",
	}
	assert.True(suite.T(), isAllowedResponseType(responseTypes, "code"))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedResponseType_InvalidResponseType() {
	responseTypes := []oauth2const.ResponseType{
		oauth2const.ResponseTypeCode,
	}
	assert.False(suite.T(), isAllowedResponseType(responseTypes, "token"))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedResponseType_EmptyResponseType() {
	responseTypes := []oauth2const.ResponseType{
		oauth2const.ResponseTypeCode,
	}
	assert.False(suite.T(), isAllowedResponseType(responseTypes, ""))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedResponseType_EmptyList() {
	responseTypes := []oauth2const.ResponseType{}
	assert.False(suite.T(), isAllowedResponseType(responseTypes, "code"))
}

func (suite *HelperFunctionsTestSuite) TestIsAllowedResponseType_NilList() {
	assert.False(suite.T(), isAllowedResponseType(nil, "code"))
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_ValidSingleURI() {
	redirectURIs := []string{"https://example.com/callback"}
	err := validateRedirectURI(redirectURIs, "https://example.com/callback")
	assert.NoError(suite.T(), err)
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_ValidMultipleURIs() {
	redirectURIs := []string{
		"https://example.com/callback",
		"https://example.com/callback2",
	}
	err := validateRedirectURI(redirectURIs, "https://example.com/callback2")
	assert.NoError(suite.T(), err)
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_InvalidNotRegistered() {
	redirectURIs := []string{"https://example.com/callback"}
	err := validateRedirectURI(redirectURIs, "https://different.com/callback")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotRegistered, err.Error())
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_InvalidWithFragment() {
	redirectURIs := []string{"https://example.com/callback#fragment"}
	err := validateRedirectURI(redirectURIs, "https://example.com/callback#fragment")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIFragment, err.Error())
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_EmptyURIWithSingleFullyQualified() {
	redirectURIs := []string{"https://example.com/callback"}
	err := validateRedirectURI(redirectURIs, "")
	assert.NoError(suite.T(), err)
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_EmptyURIWithMultiple() {
	redirectURIs := []string{
		"https://example.com/callback",
		"https://example.com/callback2",
	}
	err := validateRedirectURI(redirectURIs, "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIRequired, err.Error())
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_EmptyURIWithPartialRegistered() {
	redirectURIs := []string{"/callback"}
	err := validateRedirectURI(redirectURIs, "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotFullyQualified, err.Error())
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_EmptyURIWithNoScheme() {
	redirectURIs := []string{"example.com/callback"}
	err := validateRedirectURI(redirectURIs, "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotFullyQualified, err.Error())
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_EmptyURIWithNoHost() {
	redirectURIs := []string{"https:///callback"}
	err := validateRedirectURI(redirectURIs, "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURINotFullyQualified, err.Error())
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_EmptyURIList() {
	redirectURIs := []string{}
	err := validateRedirectURI(redirectURIs, "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIRequired, err.Error())
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_NilList() {
	err := validateRedirectURI(nil, "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), errRedirectURIRequired, err.Error())
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_CustomScheme() {
	redirectURIs := []string{"myapp://callback"}
	err := validateRedirectURI(redirectURIs, "myapp://callback")
	assert.NoError(suite.T(), err)
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_LocalhostHTTP() {
	redirectURIs := []string{"http://localhost:3000/callback"}
	err := validateRedirectURI(redirectURIs, "http://localhost:3000/callback")
	assert.NoError(suite.T(), err)
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_WithQueryParams() {
	redirectURIs := []string{"https://example.com/callback?foo=bar"}
	err := validateRedirectURI(redirectURIs, "https://example.com/callback?foo=bar")
	assert.NoError(suite.T(), err)
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_IPAddress() {
	redirectURIs := []string{"https://192.168.1.1/callback"}
	err := validateRedirectURI(redirectURIs, "https://192.168.1.1/callback")
	assert.NoError(suite.T(), err)
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_Localhost127() {
	redirectURIs := []string{"http://127.0.0.1:8080/callback"}
	err := validateRedirectURI(redirectURIs, "http://127.0.0.1:8080/callback")
	assert.NoError(suite.T(), err)
}

func (suite *HelperFunctionsTestSuite) TestValidateRedirectURI_InvalidURLFormat() {
	redirectURIs := []string{"http://example.com/callback\x00invalid"}
	err := validateRedirectURI(redirectURIs, "http://example.com/callback\x00invalid")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid redirect URI")
}
