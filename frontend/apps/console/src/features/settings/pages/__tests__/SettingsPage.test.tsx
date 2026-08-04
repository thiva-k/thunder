// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {renderWithProviders} from '@thunderid/test-utils';
import {describe, it, expect, vi} from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string) => key}),
}));

vi.mock('../../components/cors/CorsSection', () => ({
  default: () => <div data-testid="cors-section" />,
}));

const {default: SettingsPage} = await import('../SettingsPage');

describe('SettingsPage', () => {
  it('renders the title, subtitle, the CORS tab, and the CORS panel', () => {
    renderWithProviders(<SettingsPage />);

    expect(screen.getByText('settings:page.title')).toBeInTheDocument();
    expect(screen.getByText('settings:page.subtitle')).toBeInTheDocument();
    expect(screen.getByRole('tab', {name: 'settings:tabs.cors'})).toBeInTheDocument();
    expect(screen.getByTestId('cors-section')).toBeInTheDocument();
  });

  it('keeps the CORS panel active when its tab is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<SettingsPage />);
    await user.click(screen.getByRole('tab', {name: 'settings:tabs.cors'}));
    expect(screen.getByTestId('cors-section')).toBeInTheDocument();
  });
});
