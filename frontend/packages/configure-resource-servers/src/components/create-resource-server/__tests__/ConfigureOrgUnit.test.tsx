// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen} from '@thunderid/test-utils';
import {describe, expect, it, vi} from 'vitest';
import ConfigureOrgUnit from '../ConfigureOrgUnit';

vi.mock('@thunderid/configure-organization-units', () => ({
  OrganizationUnitTreePicker: () => <div data-testid="ou-tree-picker" />,
}));

describe('ConfigureOrgUnit', () => {
  it('renders the resource server subtitle when selectedType is not provided', () => {
    render(<ConfigureOrgUnit selectedOuId="" onOuIdChange={vi.fn()} />);

    expect(screen.getByText('Select which organization unit this resource server belongs to.')).toBeInTheDocument();
  });

  it('renders the resource server subtitle when selectedType is API', () => {
    render(<ConfigureOrgUnit selectedOuId="" selectedType="API" onOuIdChange={vi.fn()} />);

    expect(screen.getByText('Select which organization unit this resource server belongs to.')).toBeInTheDocument();
  });

  it('renders the MCP server subtitle when selectedType is MCP', () => {
    render(<ConfigureOrgUnit selectedOuId="" selectedType="MCP" onOuIdChange={vi.fn()} />);

    expect(screen.getByText('Select which organization unit this MCP server belongs to.')).toBeInTheDocument();
  });

  it('renders the organization unit tree picker', () => {
    render(<ConfigureOrgUnit selectedOuId="" onOuIdChange={vi.fn()} />);

    expect(screen.getByTestId('ou-tree-picker')).toBeInTheDocument();
  });

  it('calls onReadyChange with true when selectedOuId is non-empty', () => {
    const onReadyChange = vi.fn();
    render(<ConfigureOrgUnit selectedOuId="ou-1" onOuIdChange={vi.fn()} onReadyChange={onReadyChange} />);

    expect(onReadyChange).toHaveBeenCalledWith(true);
  });

  it('calls onReadyChange with false when selectedOuId is empty', () => {
    const onReadyChange = vi.fn();
    render(<ConfigureOrgUnit selectedOuId="" onOuIdChange={vi.fn()} onReadyChange={onReadyChange} />);

    expect(onReadyChange).toHaveBeenCalledWith(false);
  });
});
