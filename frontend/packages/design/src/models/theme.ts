// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {CssVarsTheme, Theme as OxygenUITheme} from '@wso2/oxygen-ui';

/**
 * Theme configuration containing color schemes, shape, and typography
 */
export type Theme = OxygenUITheme & CssVarsTheme;

/**
 * Supported color scheme options for theme preview and default scheme selection.
 */
export type ColorSchemeOption = 'light' | 'dark' | 'system';
