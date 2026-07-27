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

import {describe, expect, it} from 'vitest';
import type {OAuth2Config} from '../../models/oauth';
import resolveApplicationType, {isM2MApplication} from '../resolveApplicationType';

const makeConfig = (overrides: Partial<OAuth2Config>): OAuth2Config => ({
  publicClient: false,
  grantTypes: [],
  responseTypes: [],
  redirectUris: [],
  pkceRequired: false,
  scopes: [],
  ...overrides,
});

describe('resolveApplicationType', () => {
  it('returns the explicit type when it is a known non-custom type', () => {
    expect(resolveApplicationType('m2m')).toBe('m2m');
    expect(resolveApplicationType('mobile')).toBe('mobile');
    expect(resolveApplicationType('browser', makeConfig({grantTypes: ['client_credentials']}))).toBe('browser');
  });

  it('falls back to config inference for legacy (undefined) type', () => {
    expect(resolveApplicationType(undefined, makeConfig({grantTypes: ['client_credentials']}))).toBe('m2m');
    expect(
      resolveApplicationType(undefined, makeConfig({publicClient: true, grantTypes: ['authorization_code']})),
    ).toBe('browser');
    expect(resolveApplicationType(undefined, makeConfig({grantTypes: ['authorization_code', 'refresh_token']}))).toBe(
      'fullstack',
    );
  });

  it('falls back to config inference for custom type', () => {
    expect(resolveApplicationType('custom', makeConfig({grantTypes: ['client_credentials']}))).toBe('m2m');
  });

  it('returns custom when neither type nor a recognizable config is present', () => {
    expect(resolveApplicationType(undefined, makeConfig({grantTypes: []}))).toBe('custom');
    expect(resolveApplicationType('custom', makeConfig({grantTypes: []}))).toBe('custom');
  });
});

describe('isM2MApplication', () => {
  it('detects m2m from the explicit type', () => {
    expect(isM2MApplication('m2m')).toBe(true);
    expect(isM2MApplication('browser')).toBe(false);
  });

  it('detects m2m from config shape for legacy/custom apps', () => {
    expect(isM2MApplication(undefined, makeConfig({grantTypes: ['client_credentials']}))).toBe(true);
    expect(isM2MApplication('custom', makeConfig({grantTypes: ['authorization_code']}))).toBe(false);
  });
});
