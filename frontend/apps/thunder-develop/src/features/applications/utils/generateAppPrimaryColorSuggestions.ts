/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/**
 * A predefined collection of Material Design inspired hex color codes.
 *
 * These colors are carefully selected to provide good contrast and accessibility
 * for use as primary brand colors in applications.
 */
const PRIMARY_COLORS: string[] = [
  '#1976d2',
  '#dc004e',
  '#ed6c02',
  '#2e7d32',
  '#9c27b0',
  '#d32f2f',
  '#7b1fa2',
  '#303f9f',
  '#388e3c',
  '#f57c00',
  '#5d4037',
  '#616161',
  '#455a64',
  '#e91e63',
  '#673ab7',
  '#009688',
];

/**
 * Generates a curated list of primary color suggestions for application branding.
 *
 * This function returns a predefined array of Material Design inspired hex color codes
 * that are suitable for use as primary brand colors in applications. The colors are carefully
 * selected to provide good contrast and accessibility.
 *
 * @param count - The number of color suggestions to return. Defaults to all available colors (16).
 * @returns An array of hex color code strings (e.g., '#1976d2') representing primary color options.
 *
 * @example
 * ```typescript
 * // Get all color suggestions
 * const allColors = generateAppPrimaryColorSuggestions();
 * // Returns all 16 colors
 *
 * // Get only 5 color suggestions
 * const fiveColors = generateAppPrimaryColorSuggestions(5);
 * // Returns: ['#1976d2', '#dc004e', '#ed6c02', '#2e7d32', '#9c27b0']
 * ```
 */
export default function generateAppPrimaryColorSuggestions(count: number = PRIMARY_COLORS.length): string[] {
  return PRIMARY_COLORS.slice(0, count);
}
