// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {CimdSettings} from '../models/cimd';

/**
 * Validates a client ID URL to block: an https URL with a host and a path, as a Client ID Metadata
 * Document's client ID is.
 *
 * @param raw - The typed URL
 * @returns Whether it is a client ID URL
 *
 * @public
 */
export function isValidClientIdUrl(raw: string): boolean {
  try {
    const url = new URL(raw);
    return url.protocol === 'https:' && url.hostname.length > 0 && url.pathname.length > 1 && !url.hash;
  } catch {
    return false;
  }
}

/**
 * Returns a layer of the section with an empty list in place of an absent one.
 *
 * @param settings - The layer, absent before the section is first saved
 * @returns The settings to edit
 *
 * @public
 */
export function toCimdSettings(settings: Partial<CimdSettings> | null | undefined): CimdSettings {
  return {
    enabled: settings?.enabled ?? false,
    blockedClientIds: settings?.blockedClientIds ?? [],
  };
}
