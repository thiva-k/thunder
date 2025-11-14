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

import {describe, expect, it} from 'vitest';
import generateAppPrimaryColorSuggestions from '../generateAppPrimaryColorSuggestions';

describe('generateAppPrimaryColorSuggestions', () => {
  describe('Basic Functionality', () => {
    it('should return an array of color hex codes', () => {
      const colors = generateAppPrimaryColorSuggestions();

      expect(Array.isArray(colors)).toBe(true);
      expect(colors.length).toBeGreaterThan(0);
    });

    it('should return all 16 colors by default', () => {
      const colors = generateAppPrimaryColorSuggestions();

      expect(colors).toHaveLength(16);
    });

    it('should return strings', () => {
      const colors = generateAppPrimaryColorSuggestions();

      colors.forEach((color) => {
        expect(typeof color).toBe('string');
      });
    });

    it('should return valid hex color codes', () => {
      const colors = generateAppPrimaryColorSuggestions();

      colors.forEach((color) => {
        expect(color).toMatch(/^#[0-9a-f]{6}$/i);
      });
    });
  });

  describe('Count Parameter', () => {
    it('should return specified number of colors when count is provided', () => {
      const counts = [1, 3, 5, 10, 15];

      counts.forEach((count) => {
        const colors = generateAppPrimaryColorSuggestions(count);
        expect(colors).toHaveLength(count);
      });
    });

    it('should return 1 color when count is 1', () => {
      const colors = generateAppPrimaryColorSuggestions(1);

      expect(colors).toHaveLength(1);
      expect(colors[0]).toMatch(/^#[0-9a-f]{6}$/i);
    });

    it('should return 5 colors when count is 5', () => {
      const colors = generateAppPrimaryColorSuggestions(5);

      expect(colors).toHaveLength(5);
    });

    it('should return 10 colors when count is 10', () => {
      const colors = generateAppPrimaryColorSuggestions(10);

      expect(colors).toHaveLength(10);
    });

    it('should return all 16 colors when count is 16', () => {
      const colors = generateAppPrimaryColorSuggestions(16);

      expect(colors).toHaveLength(16);
    });

    it('should handle count of 0', () => {
      const colors = generateAppPrimaryColorSuggestions(0);

      expect(colors).toHaveLength(0);
      expect(colors).toEqual([]);
    });

    it('should handle count larger than available colors', () => {
      const colors = generateAppPrimaryColorSuggestions(100);

      // Should return maximum of 16 colors
      expect(colors.length).toBeLessThanOrEqual(16);
    });
  });

  describe('Color Values', () => {
    it('should include expected Material Design colors', () => {
      const colors = generateAppPrimaryColorSuggestions();

      // Test for some known Material Design colors
      expect(colors).toContain('#1976d2'); // Blue
      expect(colors).toContain('#dc004e'); // Pink
      expect(colors).toContain('#ed6c02'); // Orange
      expect(colors).toContain('#2e7d32'); // Green
    });

    it('should return colors in consistent order', () => {
      const colors1 = generateAppPrimaryColorSuggestions();
      const colors2 = generateAppPrimaryColorSuggestions();

      expect(colors1).toEqual(colors2);
    });

    it('should return same first N colors when sliced', () => {
      const allColors = generateAppPrimaryColorSuggestions();
      const fiveColors = generateAppPrimaryColorSuggestions(5);

      expect(fiveColors).toEqual(allColors.slice(0, 5));
    });

    it('should have all unique colors', () => {
      const colors = generateAppPrimaryColorSuggestions();
      const uniqueColors = new Set(colors);

      expect(uniqueColors.size).toBe(colors.length);
    });
  });

  describe('Hex Code Format', () => {
    it('should start with # symbol', () => {
      const colors = generateAppPrimaryColorSuggestions();

      colors.forEach((color) => {
        expect(color.startsWith('#')).toBe(true);
      });
    });

    it('should have 6 characters after # symbol', () => {
      const colors = generateAppPrimaryColorSuggestions();

      colors.forEach((color) => {
        expect(color.length).toBe(7); // # + 6 characters
      });
    });

    it('should use lowercase hex characters', () => {
      const colors = generateAppPrimaryColorSuggestions();

      colors.forEach((color) => {
        const hexPart = color.substring(1);
        expect(hexPart).toBe(hexPart.toLowerCase());
      });
    });

    it('should contain only valid hex characters', () => {
      const colors = generateAppPrimaryColorSuggestions();

      colors.forEach((color) => {
        const hexPart = color.substring(1);
        expect(/^[0-9a-f]{6}$/i.test(hexPart)).toBe(true);
      });
    });
  });

  describe('Predefined Color List', () => {
    it('should return the full predefined list without count parameter', () => {
      const colors = generateAppPrimaryColorSuggestions();

      const expectedColors = [
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

      expect(colors).toEqual(expectedColors);
    });

    it('should maintain order of predefined colors', () => {
      const colors = generateAppPrimaryColorSuggestions();

      expect(colors[0]).toBe('#1976d2');
      expect(colors[1]).toBe('#dc004e');
      expect(colors[2]).toBe('#ed6c02');
      expect(colors[colors.length - 1]).toBe('#009688');
    });

    it('should return subset of predefined colors in order', () => {
      const threeColors = generateAppPrimaryColorSuggestions(3);

      expect(threeColors).toEqual(['#1976d2', '#dc004e', '#ed6c02']);
    });
  });

  describe('Edge Cases', () => {
    it('should handle negative count gracefully', () => {
      // slice(0, -1) returns all elements except the last one
      const colors = generateAppPrimaryColorSuggestions(-1);

      expect(colors).toHaveLength(15); // Returns all colors except last
      colors.forEach((color) => {
        expect(color).toMatch(/^#[0-9a-f]{6}$/i);
      });
    });

    it('should handle undefined count parameter', () => {
      const colors = generateAppPrimaryColorSuggestions(undefined);

      expect(colors).toHaveLength(16);
    });

    it('should handle fractional count by truncating', () => {
      const colors = generateAppPrimaryColorSuggestions(5.7);

      expect(colors).toHaveLength(5);
    });
  });

  describe('Consistency', () => {
    it('should return same results on multiple calls with same parameter', () => {
      const colors1 = generateAppPrimaryColorSuggestions(10);
      const colors2 = generateAppPrimaryColorSuggestions(10);
      const colors3 = generateAppPrimaryColorSuggestions(10);

      expect(colors1).toEqual(colors2);
      expect(colors2).toEqual(colors3);
    });

    it('should be deterministic (no randomization)', () => {
      const calls = Array.from({length: 5}, () => generateAppPrimaryColorSuggestions(5));

      const firstCall = calls[0];
      calls.forEach((call) => {
        expect(call).toEqual(firstCall);
      });
    });
  });

  describe('Material Design Compliance', () => {
    it('should include Material Design primary color palette', () => {
      const colors = generateAppPrimaryColorSuggestions();

      // Check for various Material Design hues
      const hasMaterialBlue = colors.some((c) => c.includes('1976d2') || c.includes('303f9f'));
      const hasMaterialPink = colors.some((c) => c.includes('dc004e') || c.includes('e91e63'));
      const hasMaterialPurple = colors.some((c) => c.includes('9c27b0') || c.includes('673ab7'));
      const hasMaterialGreen = colors.some((c) => c.includes('2e7d32') || c.includes('388e3c'));

      expect(hasMaterialBlue).toBe(true);
      expect(hasMaterialPink).toBe(true);
      expect(hasMaterialPurple).toBe(true);
      expect(hasMaterialGreen).toBe(true);
    });

    it('should provide good variety of colors', () => {
      const colors = generateAppPrimaryColorSuggestions();

      // Colors should have different first characters (indicating different hue families)
      const firstChars = colors.map((c) => c[1]);
      const uniqueFirstChars = new Set(firstChars);

      // Should have variety in hue families
      expect(uniqueFirstChars.size).toBeGreaterThan(5);
    });
  });
});
