// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import generateResourceId from './generateResourceId';
import type {JsonArray, JsonObject, JsonValue} from '../models/json';

interface Replacer {
  key?: string;
  placeholder?: string;
  type?: string;
  value?: string;
  /**
   * Prefix for generated ids. Node ids surface as the step title on the
   * canvas, so a purposeful prefix (e.g. `recovery_call`) keeps otherwise
   * identical nodes tellable apart.
   */
  prefix?: string;
  [key: string]: unknown;
}

const updateTemplatePlaceholderReferences = <T = JsonValue>(
  obj: T,
  replacers: Replacer[],
): [T, Map<string, string>] => {
  const placeholderCache = new Map<string, string>();

  const replacePlaceholders = (input: JsonValue): JsonValue => {
    if (Array.isArray(input)) {
      return input.map((value) => replacePlaceholders(value)) as JsonArray;
    }
    if (typeof input === 'object' && input !== null) {
      return Object.fromEntries(
        Object.entries(input).map(([key, value]) => {
          if (typeof value === 'string') {
            // Extract placeholder key (remove {{ }})
            const placeholderKey = value.replace(/[{}]/g, '');

            // Check if we already replaced this placeholder
            if (placeholderCache.has(placeholderKey)) {
              return [key, placeholderCache.get(placeholderKey)];
            }

            // Find the matching replacer
            const replacer = replacers?.find((r) => r.key === placeholderKey || r.placeholder === placeholderKey);

            if (replacer) {
              let replacementValue: string;

              if (replacer.type === 'ID') {
                replacementValue = generateResourceId(replacer.prefix ?? replacer.type);
              } else {
                replacementValue = replacer.value ?? value; // Default to original if no value
              }

              // Store the replacement in the cache
              placeholderCache.set(placeholderKey, replacementValue);

              return [key, replacementValue];
            }
          }

          return [key, replacePlaceholders(value)];
        }),
      ) as JsonObject;
    }

    return input;
  };

  return [replacePlaceholders(obj as unknown as JsonValue) as unknown as T, placeholderCache];
};

export default updateTemplatePlaceholderReferences;
