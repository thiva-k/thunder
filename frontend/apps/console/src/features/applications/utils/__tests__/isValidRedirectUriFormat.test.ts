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

import {describe, it, expect} from 'vitest';
import isValidRedirectUriFormat from '../isValidRedirectUriFormat';

describe('isValidRedirectUriFormat', () => {
  it('accepts well-formed URIs', () => {
    expect(isValidRedirectUriFormat('https://example.com/callback')).toBe(true);
    expect(isValidRedirectUriFormat('http://localhost:3000/cb')).toBe(true);
  });

  it('accepts host wildcards', () => {
    expect(isValidRedirectUriFormat('https://*.example.com/callback')).toBe(true);
    expect(isValidRedirectUriFormat('https://app-*.example.com/cb')).toBe(true);
  });

  it('accepts path wildcards', () => {
    expect(isValidRedirectUriFormat('https://example.com/callback/*')).toBe(true);
  });

  it('rejects empty or whitespace-only input', () => {
    expect(isValidRedirectUriFormat('')).toBe(false);
    expect(isValidRedirectUriFormat('   ')).toBe(false);
  });

  it('rejects malformed URIs', () => {
    expect(isValidRedirectUriFormat('not a uri')).toBe(false);
    expect(isValidRedirectUriFormat('://missing-scheme')).toBe(false);
  });
});
