/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

// API Hooks
export {default as useGetApplication} from './api/useGetApplication';
export {default as useGetApplications} from './api/useGetApplications';
export type {UseGetApplicationsParams} from './api/useGetApplications';

// Models & Types
export type {Application, ApplicationType, BasicApplication} from './models/application';
export type {ApplicationListResponse} from './models/responses';
export {InboundAuthTypes} from './models/inbound-auth';
export type {InboundAuthConfig, InboundAuthType} from './models/inbound-auth';
export {
  OAuth2GrantTypes,
  OAuth2ResponseTypes,
  REFRESH_TOKEN_ISSUING_GRANTS,
  TokenEndpointAuthMethods,
} from './models/oauth';
export type {
  AndroidAttestationConfig,
  AppleAttestationConfig,
  AttestationConfig,
  IDJAGConfig,
  IDTokenConfig,
  IDTokenResponseType,
  OAuth2Config,
  OAuth2GrantType,
  OAuth2ResponseType,
  OAuth2Token,
  RefreshTokenConfig,
  ScopeClaims,
  TokenEndpointAuthMethod,
  UserInfoConfig,
  UserInfoResponseType,
} from './models/oauth';
export type {AccessTokenConfig, AccessTokenSubConfig, AssertionConfig, TokenConfig} from './models/token';

// Constants
export {default as ApplicationQueryKeys} from './constants/application-query-keys';
