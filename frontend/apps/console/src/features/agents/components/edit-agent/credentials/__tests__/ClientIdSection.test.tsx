// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {describe, it, expect, vi} from 'vitest';
import type {OAuthAgentConfig} from '../../../../models/agent';
import ClientIdSection from '../ClientIdSection';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string) => fallback ?? key,
  }),
}));

describe('ClientIdSection', () => {
  const mockOnCopyToClipboard = vi.fn().mockResolvedValue(undefined);

  it('renders the Client ID field when configured', () => {
    const oauth2Config = {clientId: 'client-123'} as OAuthAgentConfig;
    render(
      <ClientIdSection oauth2Config={oauth2Config} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
    );

    expect(screen.getByDisplayValue('client-123')).toBeInTheDocument();
  });

  it('renders nothing when there is no client ID', () => {
    const {container} = render(
      <ClientIdSection
        oauth2Config={{} as OAuthAgentConfig}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    expect(container.firstChild).toBeNull();
  });

  it('copies the client ID when the copy button is clicked', async () => {
    const user = userEvent.setup();
    const oauth2Config = {clientId: 'client-123'} as OAuthAgentConfig;
    render(
      <ClientIdSection oauth2Config={oauth2Config} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
    );

    await user.click(screen.getByRole('button', {name: 'common:actions.copy'}));

    expect(mockOnCopyToClipboard).toHaveBeenCalledWith('client-123', 'clientId');
  });

  it('shows the copied confirmation icon when this field was just copied', () => {
    const oauth2Config = {clientId: 'client-123'} as OAuthAgentConfig;
    render(
      <ClientIdSection oauth2Config={oauth2Config} copiedField="clientId" onCopyToClipboard={mockOnCopyToClipboard} />,
    );

    expect(screen.getByRole('button', {name: 'common:actions.copied'})).toBeInTheDocument();
  });
});
