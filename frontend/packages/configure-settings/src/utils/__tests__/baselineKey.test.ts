// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {describe, it, expect} from 'vitest';
import baselineKey from '../baselineKey';

describe('baselineKey', () => {
  it('is equal for inputs that normalize to the same origins', () => {
    expect(baselineKey(['HTTPS://App.IO/', ''])).toBe(baselineKey(['https://app.io']));
  });

  it('differs when the origins differ', () => {
    expect(baselineKey(['https://a.io'])).not.toBe(baselineKey(['https://b.io']));
  });

  it('is order-sensitive', () => {
    expect(baselineKey(['https://a.io', 'https://b.io'])).not.toBe(baselineKey(['https://b.io', 'https://a.io']));
  });
});
