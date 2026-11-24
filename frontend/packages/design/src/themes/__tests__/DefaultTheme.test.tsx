// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {OxygenTheme} from '@wso2/oxygen-ui';
import {describe, expect, it} from 'vitest';
import DefaultTheme from '../DefaultTheme';

describe('DefaultTheme', () => {
  it('is the Oxygen UI default theme', () => {
    expect(DefaultTheme).toBe(OxygenTheme);
  });
});
