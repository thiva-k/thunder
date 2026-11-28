// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {ServerConfigLayers} from './responses';

/**
 * Value of the `cimd` server-config section: whether clients publishing a Client ID Metadata Document may
 * sign in without being registered, and the clients refused whatever application or agent allows their domain.
 *
 * @public
 */
export interface CimdSettings {
  enabled: boolean;
  /** Client ID URLs refused whatever application or agent allows their domain. */
  blockedClientIds: string[];
}

/**
 * The `cimd` server-config section, as its three layers.
 *
 * @public
 */
export type CimdConfigResponse = ServerConfigLayers<CimdSettings>;
