// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import HomeFloatingLogos from '../HomeFloatingLogos';

let mockMode: 'light' | 'dark' | 'system' = 'light';

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof import('@wso2/oxygen-ui')>('@wso2/oxygen-ui');
  return {
    ...actual,
    useColorScheme: () => ({mode: mockMode}),
  };
});

describe('HomeFloatingLogos', () => {
  beforeEach(() => {
    mockMode = 'light';
  });

  it('renders without crashing in light mode', () => {
    const {container} = render(<HomeFloatingLogos />);
    expect(container.firstChild).not.toBeNull();
  });

  it('renders without crashing in dark mode', () => {
    mockMode = 'dark';
    const {container} = render(<HomeFloatingLogos />);
    expect(container.firstChild).not.toBeNull();
  });

  it('renders SVG logo elements', () => {
    const {container} = render(<HomeFloatingLogos />);
    const svgs = container.querySelectorAll('svg');
    expect(svgs.length).toBeGreaterThan(0);
  });

  it('renders in system mode without crashing', () => {
    mockMode = 'system';
    const {container} = render(<HomeFloatingLogos />);
    expect(container.firstChild).not.toBeNull();
  });
});
