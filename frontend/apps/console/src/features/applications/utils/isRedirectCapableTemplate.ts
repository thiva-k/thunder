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

import type {ApplicationTemplate} from '../models/application-templates';

/**
 * Whether the template's seeded OAuth2 config uses the authorization_code grant, meaning the
 * created application needs a real redirect URI. Templates often ship a placeholder redirectUris
 * value (e.g. a localhost dev URL) so the app has something to run against out of the box; that
 * placeholder should not be mistaken for "already configured, nothing to ask" — the admin still
 * needs to confirm or replace it with their actual redirect URI.
 */
const isRedirectCapableTemplate = (templateConfig: ApplicationTemplate | null): boolean => {
  return Boolean(
    templateConfig?.defaults?.inboundAuthConfig
      ?.find((config) => config.type === 'oauth2')
      ?.config?.grantTypes?.includes('authorization_code'),
  );
};

export default isRedirectCapableTemplate;
