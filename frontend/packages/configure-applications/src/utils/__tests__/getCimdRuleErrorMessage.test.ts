// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {describe, it, expect} from 'vitest';
import getApplicationErrorMessage from '../getApplicationErrorMessage';
import getCimdRuleErrorMessage from '../getCimdRuleErrorMessage';

// Returns the default value with its params filled in, as i18next does when the key has no translation.
const t = (key: string, options?: Record<string, unknown>): string => {
  const template = typeof options?.['defaultValue'] === 'string' ? options['defaultValue'] : key;
  return Object.entries(options ?? {}).reduce(
    (message, [name, value]) => (name === 'defaultValue' ? message : message.replaceAll(`{{${name}}}`, String(value))),
    template,
  );
};

function cimdError(code: string, rule: string, params?: Record<string, unknown>): Error {
  return {
    response: {data: {code, description: {key: `error.cimdservice.${rule}_description`, params}}},
  } as unknown as Error;
}

describe('getCimdRuleErrorMessage', () => {
  it('names the domain another application or agent already trusts', () => {
    expect(getCimdRuleErrorMessage(cimdError('APP-1050', 'client_id_domain_in_use', {domain: 'vscode.dev'}), t)).toBe(
      'vscode.dev is already trusted by another application or agent. A domain can be trusted by only one.',
    );
  });

  it('resolves the rule for an agent as for an application', () => {
    expect(getCimdRuleErrorMessage(cimdError('AGT-1070', 'invalid_client_id_domain'), t)).toBe(
      'List 1 to 50 trusted domains, each only once.',
    );
  });

  it('names a domain that is not a valid name', () => {
    expect(
      getCimdRuleErrorMessage(cimdError('APP-1050', 'invalid_client_id_domain_name', {domain: 'Claude.ai'}), t),
    ).toBe(
      'Claude.ai is not a valid domain. Use a lowercase name such as claude.ai, or *.example.com for its subdomains.',
    );
  });

  it('explains each trusted-domains rule', () => {
    for (const rule of [
      'client_id_domains_without_document',
      'domain_client_values',
      'invalid_response_types',
      'immutable',
    ]) {
      expect(getCimdRuleErrorMessage(cimdError('APP-1050', rule), t)).toBeTruthy();
    }
  });

  it('leaves other errors to the generic resolution', () => {
    expect(getCimdRuleErrorMessage(cimdError('APP-1001', 'client_id_domain_in_use'), t)).toBeUndefined();
    expect(getCimdRuleErrorMessage(cimdError('APP-1050', 'unknown_rule'), t)).toBeUndefined();
    expect(
      getCimdRuleErrorMessage(
        {response: {data: {code: 'APP-1050', description: {key: 'error.other.thing'}}}} as unknown as Error,
        t,
      ),
    ).toBeUndefined();
    expect(getCimdRuleErrorMessage(new Error('network'), t)).toBeUndefined();
  });

  it('is what an application save shows for a broken rule', () => {
    expect(
      getApplicationErrorMessage(
        cimdError('APP-1050', 'client_id_domain_in_use', {domain: 'claude.ai'}),
        t,
        'update.error',
        'Failed to update the application.',
      ),
    ).toBe('claude.ai is already trusted by another application or agent. A domain can be trusted by only one.');
  });
});
