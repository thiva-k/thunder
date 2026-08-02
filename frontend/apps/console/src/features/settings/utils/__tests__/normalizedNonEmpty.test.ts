// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {describe, it, expect} from 'vitest';
import normalizedNonEmpty from '../normalizedNonEmpty';

describe('normalizedNonEmpty', () => {
  it('normalizes each value (lowercase + trailing slash) and preserves order', () => {
    expect(normalizedNonEmpty(['HTTPS://Example.COM/', 'https://app.io'])).toEqual([
      'https://example.com',
      'https://app.io',
    ]);
  });

  it('drops empty and whitespace-only entries', () => {
    expect(normalizedNonEmpty(['', '   ', 'https://app.io'])).toEqual(['https://app.io']);
  });

  it('returns an empty array when there are no non-empty values', () => {
    expect(normalizedNonEmpty(['', '  '])).toEqual([]);
  });
});
