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

import type {AllowedOrigin, CorsValue} from '../../settings/models/responses';
import {isValidOrigin, normalizeOrigin} from '../../settings/utils/origin';
import originValueText from '../../settings/utils/originValueText';

/**
 * Builds the CORS PUT payload for adding the Configuration step's origins to the deployment's
 * writable allow-list, without disturbing existing entries. Additions that are invalid, blank, or
 * already present (writable or read-only) are skipped rather than duplicated or overwritten.
 */
export default function mergeCorsOrigins(
  writable: AllowedOrigin[],
  readOnly: AllowedOrigin[],
  additions: string[],
): CorsValue {
  const existing = new Set([...writable, ...readOnly].map(originValueText).map(normalizeOrigin));
  const merged: AllowedOrigin[] = [...writable];

  additions.forEach((raw) => {
    const normalized = normalizeOrigin(raw);
    if (normalized === '' || !isValidOrigin(normalized) || existing.has(normalized)) return;
    existing.add(normalized);
    merged.push(normalized);
  });

  return {allowedOrigins: merged};
}
