// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen, userEvent} from '@thunderid/test-utils';
import {describe, expect, it, vi} from 'vitest';
import ClientIdDomainsSection from '../ClientIdDomainsSection';

describe('ClientIdDomainsSection', () => {
  it('shows the domains and reports the ones typed, normalized', async () => {
    const user = userEvent.setup();
    const onDomainsChange = vi.fn();
    render(<ClientIdDomainsSection subject="application" domains={['vscode.dev']} onDomainsChange={onDomainsChange} />);

    expect(screen.getByText('vscode.dev')).toBeInTheDocument();
    expect(screen.getByText(/signs in through this application/)).toBeInTheDocument();

    await user.type(screen.getByTestId('cimd-client-id-domains-input'), 'Claude.ai,');

    expect(onDomainsChange).toHaveBeenLastCalledWith(['vscode.dev', 'claude.ai']);
  });

  it('words the description for an agent', () => {
    render(<ClientIdDomainsSection subject="agent" domains={['claude.ai']} onDomainsChange={vi.fn()} />);

    expect(screen.getByText(/signs in as this agent/)).toBeInTheDocument();
  });

  it('asks for a domain when none is left', () => {
    render(<ClientIdDomainsSection subject="application" domains={[]} onDomainsChange={vi.fn()} />);

    expect(screen.getByText('Add at least one domain.')).toBeInTheDocument();
  });

  it('cannot be changed when read-only', () => {
    render(
      <ClientIdDomainsSection subject="application" domains={['vscode.dev']} onDomainsChange={vi.fn()} isReadOnly />,
    );

    expect(screen.getByTestId('cimd-client-id-domains-input')).toBeDisabled();
  });
});
