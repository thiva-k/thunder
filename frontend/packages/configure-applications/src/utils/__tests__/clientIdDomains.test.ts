// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {describe, expect, it} from 'vitest';
import {
  CLIENT_ID_DOMAIN_ERRORS,
  MAX_CLIENT_ID_DOMAINS,
  admitsClientsByDomain,
  buildClientIdDomainsConfig,
  isValidClientIdDomain,
  normalizeClientIdDomain,
  parseClientIdDomains,
  validateClientIdDomains,
} from '../clientIdDomains';

describe('clientIdDomains', () => {
  it('normalizes a typed or pasted domain', () => {
    expect(normalizeClientIdDomain('  https://Claude.AI/oauth/client.json ')).toBe('claude.ai');
    expect(normalizeClientIdDomain('vscode.dev:443')).toBe('vscode.dev');
    expect(normalizeClientIdDomain('example.com.')).toBe('example.com');
    expect(normalizeClientIdDomain('*.Example.com')).toBe('*.example.com');
  });

  it('accepts DNS names of two or more labels, optionally with a leading wildcard', () => {
    for (const domain of ['claude.ai', '*.example.com', 'a-b.example.co', 'xn--bcher-kva.example']) {
      expect(isValidClientIdDomain(domain)).toBe(true);
    }
    for (const domain of ['', '*.', 'localhost', '10.0.0.1', 'a.*.example.com', '-a.example', 'a_b.example']) {
      expect(isValidClientIdDomain(domain)).toBe(false);
    }
  });

  it('splits, normalizes and deduplicates typed values', () => {
    expect(parseClientIdDomains(['vscode.dev, Claude.ai', 'vscode.dev', ' ', 'https://cursor.com/x'])).toEqual([
      'vscode.dev',
      'claude.ai',
      'cursor.com',
    ]);
  });

  it('validates a domain list', () => {
    expect(validateClientIdDomains([])).toBe(CLIENT_ID_DOMAIN_ERRORS.required);
    expect(validateClientIdDomains(['localhost'])).toBe(CLIENT_ID_DOMAIN_ERRORS.invalid);
    const tooMany = Array.from({length: MAX_CLIENT_ID_DOMAINS + 1}, (_, i) => `d${i}.example.com`);
    expect(validateClientIdDomains(tooMany)).toBe(CLIENT_ID_DOMAIN_ERRORS.tooMany);
    expect(validateClientIdDomains(['vscode.dev', '*.example.com'])).toBeUndefined();
  });

  it('builds the configuration of a client that admits clients by domain, and recognizes it', () => {
    const config = buildClientIdDomainsConfig(['vscode.dev']);

    expect(config).toEqual({
      clientIdMetadataDocument: true,
      clientIdDomains: ['vscode.dev'],
      grantTypes: ['authorization_code', 'refresh_token'],
      responseTypes: ['code'],
      tokenEndpointAuthMethod: 'none',
      publicClient: true,
      pkceRequired: true,
    });
    expect(admitsClientsByDomain(config)).toBe(true);
    // A domain client whose domains are all removed while editing is still one: the mode is fixed at creation.
    expect(admitsClientsByDomain({...config, clientIdDomains: []})).toBe(true);
    expect(admitsClientsByDomain({...config, clientIdDomains: undefined})).toBe(false);
    expect(admitsClientsByDomain({...config, clientIdMetadataDocument: false})).toBe(false);
    expect(admitsClientsByDomain(undefined)).toBe(false);
  });
});
