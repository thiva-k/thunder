// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {OAuth2Config} from '../models/oauth';
import {OAuth2GrantTypes, TokenEndpointAuthMethods} from '../models/oauth';

/** A domain starting with this prefix matches every subdomain of the rest. */
export const WILDCARD_DOMAIN_PREFIX = '*.';

/** The most domains one application or agent can list. The server applies the same limit. */
export const MAX_CLIENT_ID_DOMAINS = 50;

const DOMAIN_LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
const IPV4 = /^\d{1,3}(?:\.\d{1,3}){3}$/;
const MAX_DOMAIN_LENGTH = 253;

/**
 * Normalizes a typed domain: trims it, lowercases it, and drops a scheme, path or port pasted with it, so
 * `https://Claude.ai/oauth` becomes `claude.ai`.
 *
 * @param raw - The typed value
 * @returns The domain to validate and store
 *
 * @public
 */
export function normalizeClientIdDomain(raw: string): string {
  let value = raw.trim().toLowerCase();
  value = value.replace(/^[a-z][a-z0-9+.-]*:\/\//, '');
  value = value.split(/[/?#]/)[0] ?? '';
  return value.replace(/:\d+$/, '').replace(/\.$/, '');
}

/**
 * Reports whether a domain is a DNS name of at least two labels, optionally starting with `*.` to match its
 * subdomains. IP addresses are refused, since a client ID URL never has one. The server applies the same rule.
 *
 * @param domain - A normalized domain
 * @returns Whether the server accepts it
 *
 * @public
 */
export function isValidClientIdDomain(domain: string): boolean {
  const name = domain.startsWith(WILDCARD_DOMAIN_PREFIX) ? domain.slice(WILDCARD_DOMAIN_PREFIX.length) : domain;
  if (name.length === 0 || name.length > MAX_DOMAIN_LENGTH || IPV4.test(name)) {
    return false;
  }
  const labels = name.split('.');
  return labels.length >= 2 && labels.every((label) => DOMAIN_LABEL.test(label));
}

/**
 * Turns typed values into domains: each value may hold several domains separated by spaces or commas. The
 * result is normalized, without empty entries or repeats, in the order typed.
 *
 * @param values - The typed values
 * @returns The domains
 *
 * @public
 */
export function parseClientIdDomains(values: string[]): string[] {
  return [
    ...new Set(
      values
        .flatMap((value) => value.split(/[\s,]+/))
        .map(normalizeClientIdDomain)
        .filter((domain) => domain.length > 0),
    ),
  ];
}

/**
 * The validation messages of a domain list, as keys in the `applications` namespace with their English text.
 *
 * @public
 */
export const CLIENT_ID_DOMAIN_ERRORS = {
  required: {key: 'applications:cimd.domains.validation.required', defaultValue: 'Add at least one domain.'},
  tooMany: {
    key: 'applications:cimd.domains.validation.tooMany',
    defaultValue: 'You can list at most 50 domains.',
  },
  invalid: {
    key: 'applications:cimd.domains.validation.invalid',
    defaultValue:
      'Use domain names such as claude.ai or *.example.com. IP addresses and single-label names are not allowed.',
  },
} as const;

/**
 * Checks a domain list: at least one domain, at most the limit, and each one valid.
 *
 * @param domains - The domains
 * @returns The message to show, or undefined when the list is valid
 *
 * @public
 */
export function validateClientIdDomains(
  domains: string[],
): (typeof CLIENT_ID_DOMAIN_ERRORS)[keyof typeof CLIENT_ID_DOMAIN_ERRORS] | undefined {
  if (domains.length === 0) return CLIENT_ID_DOMAIN_ERRORS.required;
  if (domains.length > MAX_CLIENT_ID_DOMAINS) return CLIENT_ID_DOMAIN_ERRORS.tooMany;
  if (!domains.every(isValidClientIdDomain)) return CLIENT_ID_DOMAIN_ERRORS.invalid;
  return undefined;
}

/**
 * Whether an OAuth configuration admits any client publishing a Client ID Metadata Document on its domains,
 * rather than one client.
 *
 * @param config - The OAuth configuration
 * @returns Whether it lists Client ID domains
 *
 * @public
 */
export function admitsClientsByDomain(config?: OAuth2Config | null): boolean {
  return config?.clientIdMetadataDocument === true && Array.isArray(config.clientIdDomains);
}

/**
 * The OAuth configuration of an application or agent that admits clients by the domain of their client ID
 * URL. It has no client ID, secret, redirect URIs or keys: each client's metadata document supplies them.
 *
 * @param domains - The domains a client ID URL may be on
 * @returns The configuration to send on create
 *
 * @public
 */
export function buildClientIdDomainsConfig(domains: string[]): OAuth2Config {
  return {
    clientIdMetadataDocument: true,
    clientIdDomains: domains,
    grantTypes: [OAuth2GrantTypes.AUTHORIZATION_CODE, OAuth2GrantTypes.REFRESH_TOKEN],
    responseTypes: ['code'],
    tokenEndpointAuthMethod: TokenEndpointAuthMethods.NONE,
    publicClient: true,
    pkceRequired: true,
  };
}
