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

/**
 * Checks that a redirect URI is well formed, tolerating host wildcards. Wildcards in the host are
 * replaced with a placeholder so `new URL()` can parse it; path wildcards (e.g. `/callback/*`) parse
 * natively. The backend enforces the actual wildcard rules (allowed patterns, server config).
 *
 * Empty or whitespace-only input is rejected. Used by both the redirect and post-logout redirect
 * URI fields and the page-level save validation, so a valid wildcard URI never disables Save.
 *
 * @param uri - The redirect URI to check
 * @returns Whether the URI is a parseable, non-empty redirect URI
 *
 * @public
 */
export default function isValidRedirectUriFormat(uri: string): boolean {
  if (!uri.trim()) return false;

  try {
    const schemeEnd = uri.indexOf('://');
    let uriForValidation = uri;
    if (schemeEnd !== -1) {
      const pathStart = uri.indexOf('/', schemeEnd + 3);
      const hostPart = pathStart !== -1 ? uri.slice(schemeEnd + 3, pathStart) : uri.slice(schemeEnd + 3);
      if (hostPart.includes('*')) {
        const sanitizedHost = hostPart.replace(/\*/g, 'wildcard-placeholder');
        uriForValidation = uri.slice(0, schemeEnd + 3) + sanitizedHost + (pathStart !== -1 ? uri.slice(pathStart) : '');
      }
    }
    // eslint-disable-next-line no-new
    new URL(uriForValidation);
    return true;
  } catch {
    return false;
  }
}
