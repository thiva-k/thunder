// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {
  isValidOrigin,
  normalizeOrigin,
  originValueText,
  type AllowedOrigin,
  type CorsValue,
} from '@thunderid/configure-settings';

/**
 * Builds the CORS PUT payload for adding the Configuration step's origins to the deployment's
 * writable allow-list, without disturbing existing entries. Additions that are invalid, blank, or
 * already present (writable or read-only) are skipped rather than duplicated or overwritten.
 */
export default function mergeCorsOrigins(
  writable: AllowedOrigin[],
  readOnly: AllowedOrigin[],
  additions: string[],
): CorsValue {
  const existing = new Set([...writable, ...readOnly].map(originValueText).map(normalizeOrigin));
  const merged: AllowedOrigin[] = [...writable];

  additions.forEach((raw) => {
    const normalized = normalizeOrigin(raw);
    if (normalized === '' || !isValidOrigin(normalized) || existing.has(normalized)) return;
    existing.add(normalized);
    merged.push(normalized);
  });

  return {allowedOrigins: merged};
}
