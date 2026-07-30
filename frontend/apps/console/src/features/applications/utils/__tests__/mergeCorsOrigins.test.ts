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
import mergeCorsOrigins from '../mergeCorsOrigins';

describe('mergeCorsOrigins', () => {
  it('appends new valid origins to the existing writable list', () => {
    const result = mergeCorsOrigins(['https://existing.example.com'], [], ['https://new.example.com']);

    expect(result).toEqual({allowedOrigins: ['https://existing.example.com', 'https://new.example.com']});
  });

  it('skips additions that already exist in the writable list', () => {
    const result = mergeCorsOrigins(['https://existing.example.com'], [], ['https://existing.example.com']);

    expect(result).toEqual({allowedOrigins: ['https://existing.example.com']});
  });

  it('skips additions that already exist in the read-only list', () => {
    const result = mergeCorsOrigins([], ['https://readonly.example.com'], ['https://readonly.example.com']);

    expect(result).toEqual({allowedOrigins: []});
  });

  it('normalizes additions before comparing and storing (trailing slash, casing)', () => {
    const result = mergeCorsOrigins(['https://existing.example.com'], [], ['HTTPS://EXISTING.example.com/']);

    expect(result).toEqual({allowedOrigins: ['https://existing.example.com']});
  });

  it('skips blank and invalid additions', () => {
    const result = mergeCorsOrigins([], [], ['', '   ', 'not-a-valid-origin', 'https://valid.example.com/path']);

    expect(result).toEqual({allowedOrigins: []});
  });

  it('dedupes multiple identical additions in the same call', () => {
    const result = mergeCorsOrigins([], [], ['http://localhost:5173', 'http://localhost:5173']);

    expect(result).toEqual({allowedOrigins: ['http://localhost:5173']});
  });

  it('leaves regex read-only/writable entries untouched and still compares against their text', () => {
    const result = mergeCorsOrigins([{regex: '^https://.*\\.example\\.com$'}], [], ['http://localhost:3000']);

    expect(result).toEqual({
      allowedOrigins: [{regex: '^https://.*\\.example\\.com$'}, 'http://localhost:3000'],
    });
  });
});
