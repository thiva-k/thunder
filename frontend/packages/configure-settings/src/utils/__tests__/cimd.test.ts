// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {describe, expect, it} from 'vitest';
import {isValidClientIdUrl, toCimdSettings} from '../cimd';

describe('isValidClientIdUrl', () => {
  it('accepts an https URL with a path', () => {
    expect(isValidClientIdUrl('https://claude.ai/oauth/client.json')).toBe(true);
  });

  it.each(['claude.ai', 'http://claude.ai/client.json', 'https://claude.ai', 'https://claude.ai/', 'https://a.b/c#x'])(
    'refuses %s',
    (value) => {
      expect(isValidClientIdUrl(value)).toBe(false);
    },
  );
});

describe('toCimdSettings', () => {
  it('fills in absent values', () => {
    expect(toCimdSettings(undefined)).toEqual({enabled: false, blockedClientIds: []});
    expect(toCimdSettings({enabled: true})).toEqual({enabled: true, blockedClientIds: []});
  });
});
