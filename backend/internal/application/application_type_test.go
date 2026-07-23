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

package application

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/internal/application/model"
	inboundmodel "github.com/thunder-id/thunderid/internal/inboundclient/model"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

type ApplicationTypeTestSuite struct {
	suite.Suite
}

func TestApplicationTypeTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationTypeTestSuite))
}

// TestToInboundClientPersistsType verifies the application type is packed into the inbound client
// properties for persistence.
func (s *ApplicationTypeTestSuite) TestToInboundClientPersistsType() {
	dto := &model.ApplicationProcessedDTO{ID: "app-1", Type: model.ApplicationTypeMobile}

	dao := toInboundClient(dto)

	s.Equal("mobile", dao.Properties[propType])
}

// TestToInboundClientOmitsEmptyType verifies an unset type is not written to properties.
func (s *ApplicationTypeTestSuite) TestToInboundClientOmitsEmptyType() {
	dto := &model.ApplicationProcessedDTO{ID: "app-1"}

	dao := toInboundClient(dto)

	_, ok := dao.Properties[propType]
	s.False(ok)
}

// TestToProcessedDTOReadsType verifies a persisted type is read back onto the DTO.
func (s *ApplicationTypeTestSuite) TestToProcessedDTOReadsType() {
	dao := &inboundmodel.InboundClient{
		ID:         "app-1",
		Properties: map[string]interface{}{propType: "browser"},
	}

	dto := toProcessedDTO(nil, dao, nil)

	s.Equal(model.ApplicationTypeBrowser, dto.Type)
}

// TestToProcessedDTOEmptyWhenTypeAbsent verifies applications without a stored type resolve to an
// empty type (no implicit default is applied).
func (s *ApplicationTypeTestSuite) TestToProcessedDTOEmptyWhenTypeAbsent() {
	withProps := toProcessedDTO(nil, &inboundmodel.InboundClient{
		ID:         "app-1",
		Properties: map[string]interface{}{},
	}, nil)
	s.Equal(model.ApplicationType(""), withProps.Type)

	nilProps := toProcessedDTO(nil, &inboundmodel.InboundClient{ID: "app-2"}, nil)
	s.Equal(model.ApplicationType(""), nilProps.Type)
}

// TestBuildBasicApplicationResponseType verifies the list-view response reads the stored type and
// leaves it empty when absent (no implicit default).
func (s *ApplicationTypeTestSuite) TestBuildBasicApplicationResponseType() {
	withType := buildBasicApplicationResponse(inboundmodel.InboundClient{
		ID:         "app-1",
		Properties: map[string]interface{}{propType: "m2m"},
	}, nil)
	s.Equal(model.ApplicationTypeM2M, withType.Type)

	absent := buildBasicApplicationResponse(inboundmodel.InboundClient{ID: "app-2"}, nil)
	s.Equal(model.ApplicationType(""), absent.Type)
}

// TestFlowSecretIneligibleByType verifies browser, mobile, and m2m apps are never issued a Flow
// Secret, decided by their type alone regardless of OAuth config shape.
func (s *ApplicationTypeTestSuite) TestFlowSecretIneligibleByType() {
	embedded := &providers.InboundAuthConfigWithSecret{
		Type: providers.OAuthInboundAuthType,
		OAuthConfig: &providers.OAuthConfigWithSecret{
			GrantTypes: []providers.GrantType{
				providers.GrantTypeClientCredentials,
				providers.GrantTypeTokenExchange,
			},
		},
	}
	for _, appType := range []model.ApplicationType{
		model.ApplicationTypeBrowser,
		model.ApplicationTypeMobile,
		model.ApplicationTypeM2M,
	} {
		s.False(isFlowSecretEligible(appType, nil), "type %q should not be eligible", appType)
		s.False(isFlowSecretEligible(appType, embedded), "type %q should not be eligible", appType)
	}
}

// TestFullStackAndCustomFlowSecretEligibility verifies full-stack and custom apps derive eligibility
// from the OAuth config shape: only confidential, non-redirect (embedded) clients are eligible.
func (s *ApplicationTypeTestSuite) TestFullStackAndCustomFlowSecretEligibility() {
	embedded := &providers.InboundAuthConfigWithSecret{
		Type: providers.OAuthInboundAuthType,
		OAuthConfig: &providers.OAuthConfigWithSecret{
			GrantTypes: []providers.GrantType{
				providers.GrantTypeClientCredentials,
				providers.GrantTypeTokenExchange,
			},
			TokenEndpointAuthMethod: providers.TokenEndpointAuthMethodClientSecretBasic,
		},
	}
	redirect := &providers.InboundAuthConfigWithSecret{
		Type: providers.OAuthInboundAuthType,
		OAuthConfig: &providers.OAuthConfigWithSecret{
			GrantTypes: []providers.GrantType{providers.GrantTypeAuthorizationCode},
		},
	}
	m2mShaped := &providers.InboundAuthConfigWithSecret{
		Type: providers.OAuthInboundAuthType,
		OAuthConfig: &providers.OAuthConfigWithSecret{
			GrantTypes: []providers.GrantType{providers.GrantTypeClientCredentials},
		},
	}

	for _, appType := range []model.ApplicationType{
		model.ApplicationTypeFullStack,
		model.ApplicationTypeCustom,
	} {
		// Embedded app with no OAuth config, and confidential non-redirect app, are eligible.
		s.True(isFlowSecretEligible(appType, nil), "type %q embedded should be eligible", appType)
		s.True(isFlowSecretEligible(appType, embedded), "type %q embedded should be eligible", appType)
		// Redirect and m2m-shaped apps are not eligible.
		s.False(isFlowSecretEligible(appType, redirect), "type %q redirect should not be eligible", appType)
		s.False(isFlowSecretEligible(appType, m2mShaped), "type %q m2m-shaped should not be eligible", appType)
	}
}
