// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import normalizedNonEmpty from './normalizedNonEmpty';

/**
 * Builds a stable key for a set of origins, used to compare a draft against its saved baseline.
 *
 * @param values - The origin values to key
 * @returns A stable string key derived from the normalized, non-empty origins
 */
export default function baselineKey(values: string[]): string {
  return JSON.stringify(normalizedNonEmpty(values));
}
