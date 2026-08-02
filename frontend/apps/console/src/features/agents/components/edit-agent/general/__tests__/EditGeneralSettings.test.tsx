// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import userEvent from '@testing-library/user-event';
import {render, screen} from '@thunderid/test-utils';
import {describe, it, expect, vi} from 'vitest';
import type {Agent} from '../../../../models/agent';
import EditGeneralSettings from '../EditGeneralSettings';

vi.mock('../../attributes/AttributesSummarySection', () => ({default: () => <div data-testid="attributes-summary" />}));
vi.mock('../OwnerSummarySection', () => ({default: () => <div data-testid="owner-summary" />}));
vi.mock('../../general-settings/QuickCopySection', () => ({default: () => <div data-testid="quick-copy" />}));
vi.mock('../../general-settings/OrganizationUnitSection', () => ({
  default: () => <div data-testid="org-unit" />,
}));
vi.mock('../../general-settings/DangerZoneSection', () => ({
  default: ({onDeleteClick}: {onDeleteClick: () => void}) => (
    <button type="button" data-testid="danger-zone" onClick={onDeleteClick}>
      delete
    </button>
  ),
}));
vi.mock('../../../AgentDeleteDialog', () => ({
  default: ({open, onClose}: {open: boolean; onClose: () => void}) =>
    open ? (
      <div data-testid="delete-dialog">
        <button type="button" onClick={onClose}>
          cancel delete
        </button>
      </div>
    ) : null,
}));

describe('EditGeneralSettings', () => {
  const mockAgent: Agent = {id: 'agent-1', ouId: 'ou-1', type: 'default', name: 'Test Agent'};
  const mockOnCopyToClipboard = vi.fn();

  it('renders all general sections', () => {
    render(<EditGeneralSettings agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

    expect(screen.getByTestId('quick-copy')).toBeInTheDocument();
    expect(screen.getByTestId('owner-summary')).toBeInTheDocument();
    expect(screen.getByTestId('attributes-summary')).toBeInTheDocument();
    expect(screen.getByTestId('org-unit')).toBeInTheDocument();
    expect(screen.getByTestId('danger-zone')).toBeInTheDocument();
  });

  it('does not render the danger zone for read-only agents', () => {
    render(
      <EditGeneralSettings
        agent={{...mockAgent, isReadOnly: true}}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    expect(screen.queryByTestId('danger-zone')).not.toBeInTheDocument();
  });

  it('opens the delete dialog when danger zone reports a delete click', async () => {
    const user = userEvent.setup();
    render(<EditGeneralSettings agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

    await user.click(screen.getByTestId('danger-zone'));

    expect(screen.getByTestId('delete-dialog')).toBeInTheDocument();
  });

  it('closes the delete dialog when it reports a close', async () => {
    const user = userEvent.setup();
    render(<EditGeneralSettings agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

    await user.click(screen.getByTestId('danger-zone'));
    await user.click(screen.getByRole('button', {name: /cancel delete/i}));

    expect(screen.queryByTestId('delete-dialog')).not.toBeInTheDocument();
  });
});
