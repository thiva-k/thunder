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

import type {ApplicationType} from '../models/application';
import type {OAuth2Config} from '../models/oauth';
import {OAuth2GrantTypes} from '../models/oauth';

const KNOWN_TYPES: readonly ApplicationType[] = ['browser', 'fullstack', 'mobile', 'm2m', 'custom'];

/**
 * Resolves the canonical application type for behavior decisions.
 *
 * Prefers the explicit `type` set at creation. For legacy applications created before the type
 * attribute existed, or when the stored type is `custom` (no enforced shape), falls back to
 * inferring from the OAuth config so existing behavior is preserved. Note that `browser` and
 * `mobile` are indistinguishable from config shape alone, so the fallback resolves public
 * redirect clients to `browser`.
 *
 * @param type - The application's stored `type`, if any.
 * @param oauth2Config - The application's OAuth2 config, used only for the fallback path.
 * @returns The resolved application type.
 */
export default function resolveApplicationType(
  type: ApplicationType | undefined,
  oauth2Config?: OAuth2Config,
): ApplicationType {
  if (type && type !== 'custom' && KNOWN_TYPES.includes(type)) {
    return type;
  }

  const grantTypes = oauth2Config?.grantTypes ?? [];
  const isClientCredentialsOnly = grantTypes.length === 1 && grantTypes[0] === OAuth2GrantTypes.CLIENT_CREDENTIALS;
  if (isClientCredentialsOnly) {
    return 'm2m';
  }
  if (oauth2Config?.publicClient) {
    return 'browser';
  }
  if (grantTypes.includes(OAuth2GrantTypes.AUTHORIZATION_CODE)) {
    return 'fullstack';
  }

  return type ?? 'custom';
}

/**
 * Reports whether the application is a machine-to-machine (`m2m`) application, preferring the
 * explicit type and falling back to config-shape inference for legacy/custom applications.
 */
export function isM2MApplication(type: ApplicationType | undefined, oauth2Config?: OAuth2Config): boolean {
  return resolveApplicationType(type, oauth2Config) === 'm2m';
}
