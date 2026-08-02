// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen} from '@thunderid/test-utils';
import {describe, expect, it, vi} from 'vitest';
import ConnectionFullPageLayout from '../ConnectionFullPageLayout';

describe('ConnectionFullPageLayout', () => {
  it('renders children inside a left-aligned constrained content wrapper', () => {
    render(
      <ConnectionFullPageLayout label="Configure connection" onClose={vi.fn()} progress={50}>
        <div>Wizard content</div>
      </ConnectionFullPageLayout>,
    );

    const content = screen.getByTestId('connection-fullpage-content');
    const styles = window.getComputedStyle(content);

    expect(screen.getByText('Wizard content')).toBeInTheDocument();
    expect(styles.maxWidth).toBe('920px');
    expect(content).not.toHaveStyle({marginLeft: 'auto'});
    expect(content).not.toHaveStyle({marginRight: 'auto'});
  });

  it('lets the content column span the full width when fullWidthContent is set', () => {
    render(
      <ConnectionFullPageLayout label="Choose a type" onClose={vi.fn()} fullWidthContent>
        <div>Type gallery</div>
      </ConnectionFullPageLayout>,
    );

    const content = screen.getByTestId('connection-fullpage-content');
    const styles = window.getComputedStyle(content);

    expect(styles.maxWidth).toBe('none');
  });
});
