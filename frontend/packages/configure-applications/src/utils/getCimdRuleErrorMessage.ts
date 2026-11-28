// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

/**
 * The subset of an API error this util reads: the code, and the description the server copies from
 * the Client ID Metadata Document rule an application or agent broke.
 */
interface CimdRuleApiError {
  code?: string;
  description?: {
    key?: string;
    params?: Record<string, unknown>;
  };
}

/** The codes under which the application and agent services report a broken metadata document rule. */
const CIMD_RULE_ERROR_CODES = ['APP-1050', 'AGT-1070'];

const RULE_KEY_PREFIX = 'error.cimdservice.';
const RULE_KEY_SUFFIX = '_description';

/**
 * Default messages for the metadata document rules, keyed by the rule name in the server's description key.
 */
const RULE_DEFAULTS: Record<string, string> = {
  client_id_domains_without_document:
    'Trusted domains can be set only on a client that uses Client ID Metadata Documents.',
  domain_client_values:
    "A client that trusts domains can't have its own client ID, client secret, redirect URIs, post-logout redirect URIs or certificate. Each client's metadata document sets them.",
  invalid_client_id_domain: 'List 1 to 50 trusted domains, each only once.',
  invalid_client_id_domain_name:
    '{{domain}} is not a valid domain. Use a lowercase name such as claude.ai, or *.example.com for its subdomains.',
  client_id_domain_in_use:
    '{{domain}} is already trusted by another application or agent. A domain can be trusted by only one.',
  invalid_response_types: "The client's metadata document must allow the code response type.",
  immutable:
    "How this client signs in, through one metadata document or through trusted domains, can't be changed after it is created.",
  invalid_grant_types:
    'The grant types must include authorization code. Token exchange and JWT bearer are allowed only with private key JWT authentication.',
  token_exchange_requires_idjag: 'The token exchange grant needs identity assertions (ID-JAG) to be turned on.',
  pkce_required: 'A metadata document client must require PKCE.',
  invalid_redirect_uri:
    "Each redirect URI must be on the client's own domain or the user's device, and at least one is required.",
  invalid_auth_method:
    'Only public clients, and private key JWT clients with exactly one key source, can use a metadata document.',
};

/**
 * Resolves the specific message for an application or agent save that broke a Client ID Metadata Document
 * rule. The server reports such a failure under one code (APP-1050 or AGT-1070) and carries the rule in its
 * description key, with any params the message names, such as the domain already in use.
 *
 * @param error - The error thrown by the mutation
 * @param t - The i18next translation function; keys are passed with an explicit namespace
 * @returns The localized message, or undefined when the error is not a known metadata document rule
 *
 * @public
 */
export default function getCimdRuleErrorMessage(
  error: Error,
  t: (key: string, options?: Record<string, unknown>) => string,
): string | undefined {
  const apiError = (error as {response?: {data?: CimdRuleApiError}}).response?.data;
  const key = apiError?.description?.key;
  if (!apiError?.code || !CIMD_RULE_ERROR_CODES.includes(apiError.code) || !key?.startsWith(RULE_KEY_PREFIX)) {
    return undefined;
  }
  const rule = key.slice(RULE_KEY_PREFIX.length, key.endsWith(RULE_KEY_SUFFIX) ? -RULE_KEY_SUFFIX.length : undefined);
  const defaultValue = RULE_DEFAULTS[rule];
  if (!defaultValue) {
    return undefined;
  }
  return t(`applications:cimd.rules.${rule}`, {...apiError.description?.params, defaultValue});
}
