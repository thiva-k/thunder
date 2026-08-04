// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

/**
 * Common browser-safe / system fonts that do not require loading from Google Fonts.
 * Used by GoogleFontLoader to skip loading and by the theme builder as autocomplete suggestions.
 */
export const BROWSER_SAFE_FONTS: string[] = [
  'Arial',
  'Arial Black',
  'Brush Script MT',
  'Comic Sans MS',
  'Courier New',
  'Georgia',
  'Helvetica',
  'Impact',
  'Lucida Console',
  'Lucida Sans Unicode',
  'Palatino Linotype',
  'system-ui',
  'Tahoma',
  'Times New Roman',
  'Trebuchet MS',
  'Verdana',
];

/** Lowercase set derived from BROWSER_SAFE_FONTS plus generic CSS font families. */
export const SYSTEM_FONTS = new Set([
  ...BROWSER_SAFE_FONTS.map((f) => f.toLowerCase()),
  'sans-serif',
  'serif',
  'monospace',
  'cursive',
  'fantasy',
]);
