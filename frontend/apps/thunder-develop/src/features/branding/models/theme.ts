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
 * Color configuration for a specific color type (primary, secondary, tertiary)
 */
export interface ThemeColor {
  /**
   * Main color value
   * @example "#1976d2"
   */
  main: string;

  /**
   * Dark variant of the color
   * @example "#0d47a1"
   */
  dark: string;

  /**
   * Text color that contrasts with the main color
   * @example "#ffffff"
   */
  contrastText: string;
}

/**
 * Color palette for a specific color scheme (light or dark)
 */
export interface ColorPalette {
  /**
   * Primary color configuration
   */
  primary: ThemeColor;

  /**
   * Secondary color configuration
   */
  secondary: ThemeColor;

  /**
   * Tertiary color configuration (optional)
   */
  tertiary?: ThemeColor;
}

/**
 * Color scheme configuration for light or dark mode
 */
export interface ColorScheme {
  /**
   * Color palette for this color scheme
   */
  colors: ColorPalette;
}

/**
 * Theme configuration containing color schemes
 */
export interface Theme {
  /**
   * Active color scheme ("light" or "dark")
   */
  activeColorScheme: 'light' | 'dark';

  /**
   * Available color schemes
   */
  colorSchemes: {
    light?: ColorScheme;
    dark?: ColorScheme;
  };
}
