// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {CapabilitiesProvider} from '@thunderid/contexts';
import {renderWithProviders} from '@thunderid/test-utils';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import type {CimdConfigResponse, CimdSettings} from '../../../models/cimd';

const mockRefetch = vi.fn();
const mockUseGetCimdConfig =
  vi.fn<() => {data: CimdConfigResponse | undefined; isLoading: boolean; error: Error | null}>();
vi.mock('../../../api/useGetCimdConfig', () => ({
  default: () => ({refetch: mockRefetch, ...mockUseGetCimdConfig()}),
}));

const mockMutate = vi.fn();
vi.mock('../../../api/useUpdateCimdConfig', () => ({
  default: () => ({mutate: mockMutate, reset: vi.fn(), isPending: false, error: null}),
}));

const {default: CimdSection} = await import('../CimdSection');

const empty: CimdSettings = {enabled: false, blockedClientIds: []};

function makeData(writable: Partial<CimdSettings>, readOnly: Partial<CimdSettings> = {}): CimdConfigResponse {
  return {readOnly: {...empty, ...readOnly}, writable: {...empty, ...writable}, merged: {...empty, ...writable}};
}

const viewOnly = {can: (_resource: string, action = 'view') => action === 'view', isLoading: false};

describe('CimdSection', () => {
  beforeEach(() => {
    mockUseGetCimdConfig.mockReset();
    mockMutate.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('shows an alert when the settings fail to load', () => {
    mockUseGetCimdConfig.mockReturnValue({data: undefined, isLoading: false, error: new Error('failed')});
    renderWithProviders(<CimdSection />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });

  it('turns the feature on and points to where clients are admitted by domain', async () => {
    const user = userEvent.setup();
    mockUseGetCimdConfig.mockReturnValue({data: makeData({}), isLoading: false, error: null});
    renderWithProviders(<CimdSection />);

    expect(screen.getByTestId('cimd-domains-hint')).toHaveTextContent('Client ID Metadata Document settings');
    await user.click(screen.getByRole('switch', {name: 'Let unregistered clients sign in'}));

    expect(mockMutate).toHaveBeenCalledWith({data: {enabled: true, blockedClientIds: []}}, expect.any(Object));
  });

  it('blocks a client ID URL and validates the input', async () => {
    const user = userEvent.setup();
    mockUseGetCimdConfig.mockReturnValue({data: makeData({enabled: true}), isLoading: false, error: null});
    renderWithProviders(<CimdSection />);

    await user.type(screen.getByTestId('cimd-block-input'), 'claude.ai');
    await user.click(screen.getByRole('button', {name: 'Block'}));
    expect(screen.getByText(/Enter a client ID URL/)).toBeInTheDocument();
    expect(mockMutate).not.toHaveBeenCalled();

    await user.clear(screen.getByTestId('cimd-block-input'));
    await user.type(screen.getByTestId('cimd-block-input'), 'https://claude.ai/oauth/client.json{Enter}');
    await waitFor(() =>
      expect(mockMutate).toHaveBeenCalledWith(
        {data: {enabled: true, blockedClientIds: ['https://claude.ai/oauth/client.json']}},
        expect.any(Object),
      ),
    );
  });

  it('unblocks a client', async () => {
    const user = userEvent.setup();
    mockUseGetCimdConfig.mockReturnValue({
      data: makeData({enabled: true, blockedClientIds: ['https://claude.ai/oauth/client.json']}),
      isLoading: false,
      error: null,
    });
    renderWithProviders(<CimdSection />);

    await user.click(screen.getByRole('button', {name: 'Unblock'}));
    expect(mockMutate).toHaveBeenCalledWith({data: {enabled: true, blockedClientIds: []}}, expect.any(Object));
  });

  it('is read-only for a caller who cannot update server configuration', () => {
    mockUseGetCimdConfig.mockReturnValue({
      data: makeData({blockedClientIds: ['https://claude.ai/oauth/client.json']}),
      isLoading: false,
      error: null,
    });
    renderWithProviders(
      <CapabilitiesProvider capabilities={viewOnly}>
        <CimdSection />
      </CapabilitiesProvider>,
    );

    expect(screen.getByRole('switch', {name: 'Let unregistered clients sign in'})).toBeDisabled();
    expect(screen.queryByTestId('cimd-block-input')).toBeNull();
    expect(screen.queryByRole('button', {name: 'Unblock'})).toBeNull();
  });

  it('is read-only when a declarative document sets the section', () => {
    mockUseGetCimdConfig.mockReturnValue({
      data: makeData({}, {enabled: true}),
      isLoading: false,
      error: null,
    });
    renderWithProviders(<CimdSection />);

    expect(screen.getByTestId('cimd-read-only-hint')).toBeInTheDocument();
    expect(screen.getByRole('switch', {name: 'Let unregistered clients sign in'})).toBeChecked();
    expect(screen.queryByTestId('cimd-block-input')).toBeNull();
  });
});
