// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import type {Agent} from '../../../../models/agent';
import OwnerSummarySection from '../OwnerSummarySection';

const {mockUseGetUsers} = vi.hoisted(() => ({
  mockUseGetUsers: vi.fn(),
}));

vi.mock('@thunderid/configure-users', () => ({
  useGetUsers: (...args: unknown[]): unknown => mockUseGetUsers(...args) as unknown,
}));

describe('OwnerSummarySection', () => {
  const mockAgent: Agent = {id: 'agent-1', ouId: 'ou-1', type: 'default', name: 'Test Agent', owner: 'user-1'};

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetUsers.mockReturnValue({
      data: {users: [{id: 'user-1', display: 'Alice'}]},
      isLoading: false,
    });
  });

  it('shows the resolved owner label', () => {
    render(<OwnerSummarySection agent={mockAgent} />);

    expect(screen.getByText('Alice')).toBeInTheDocument();
  });

  it('shows a placeholder when no owner is assigned', () => {
    render(<OwnerSummarySection agent={{...mockAgent, owner: undefined}} />);

    expect(screen.getByText('No owner assigned')).toBeInTheDocument();
  });

  it('falls back to the raw owner id when the user list has not resolved it', () => {
    mockUseGetUsers.mockReturnValue({data: {users: []}, isLoading: false});
    render(<OwnerSummarySection agent={mockAgent} />);

    expect(screen.getByText('user-1')).toBeInTheDocument();
  });

  it('never shows an editable control', () => {
    render(<OwnerSummarySection agent={mockAgent} />);

    expect(screen.queryByRole('combobox')).not.toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });
});
