// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {InboundAuthConfig} from './inbound-auth';
import type {OAuth2Config} from './oauth';

/**
 * How an MCP client identifies itself, chosen on the mcp-client template's Client identity step.
 *
 * @public
 */
export const McpClientIdentityModes = {
  /** The client publishes a Client ID Metadata Document at its client ID URL. */
  METADATA_DOCUMENT: 'metadataDocument',
  /** The administrator configures the client's redirect URIs and credentials. */
  MANUAL: 'manual',
  /** Any client whose client ID URL is on a trusted domain, read from its document when it signs in. */
  TRUSTED_DOMAINS: 'trustedDomains',
} as const;

/**
 * @public
 */
export type McpClientIdentityMode = (typeof McpClientIdentityModes)[keyof typeof McpClientIdentityModes];

/**
 * Which clients a Client ID Metadata Document application admits, chosen on the CIMD template's first step.
 *
 * @public
 */
export const CimdRegistrationModes = {
  /** One client, registered from its document URL after the administrator reviews the document. */
  SPECIFIC_CLIENT: 'specificClient',
  /** Any client whose client ID URL is on a trusted domain, read from its document when it signs in. */
  TRUSTED_DOMAINS: 'trustedDomains',
} as const;

/**
 * @public
 */
export type CimdRegistrationMode = (typeof CimdRegistrationModes)[keyof typeof CimdRegistrationModes];

/**
 * The response of `POST /cimd/preview`: the values taken from a Client ID Metadata Document, in the
 * shape of the fields an application create request takes.
 *
 * @public
 */
export interface CimdPreviewResponse {
  name: string;
  url?: string;
  tosUri?: string;
  policyUri?: string;
  contacts?: string[];
  inboundAuthConfig: InboundAuthConfig[];
}

/**
 * What the administrator approves: the previewed values, with the details the review shows.
 *
 * @public
 */
export interface CimdPreview {
  clientId: string;
  clientIdHost: string;
  clientName: string;
  clientUri?: string;
  tosUri?: string;
  policyUri?: string;
  contacts: string[];
  tokenEndpointAuthMethod: string;
  jwksUri?: string;
  hasInlineJwks: boolean;
  redirectUris: string[];
  grantTypes: string[];
  loopbackOnly: boolean;
  /** The OAuth configuration to store, exactly as the preview returned it. */
  oauth2Config: OAuth2Config;
}
