// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import * as thunderIdReactModule from '@thunderid/react';
import {renderWithProviders, screen, fireEvent} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import type {DefaultResourceServerConfigResponse, ResourceServerListResponse} from '../../models/resource-server';
import ResourceServersList from '../ResourceServersList';

vi.mock('@thunderid/react', {spy: true});

// eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access -- vi.mock({spy:true}) type inference doesn't resolve for this package's conditional exports
vi.mocked(thunderIdReactModule.useThunderID).mockImplementation(() => ({http: {request: vi.fn()}}) as never);

vi.mock('@thunderid/contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunderid/contexts')>();
  return {
    ...actual,
    useConfig: () => ({getServerUrl: () => 'http://localhost:8090'}),
    useToast: () => ({showToast: vi.fn()}),
  };
});

vi.mock('@thunderid/logger/react', () => ({
  useLogger: () => ({error: vi.fn(), info: vi.fn(), debug: vi.fn()}),
}));

vi.mock('react-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-router')>();
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

vi.mock('../ResourceServerDeleteDialog', () => ({
  default: () => null,
}));

const mockUseGetResourceServers = vi.fn();
const mockRefetch = vi.fn();

vi.mock('../../api/useGetResourceServers', () => ({
  default: (...args: unknown[]) =>
    mockUseGetResourceServers(...args) as {
      data: ResourceServerListResponse | undefined;
      isLoading: boolean;
      error: Error | null;
      refetch: () => void;
    },
}));

const mockUseGetDefaultResourceServer = vi.fn();

vi.mock('../../api/useGetDefaultResourceServer', () => ({
  default: () => mockUseGetDefaultResourceServer() as {data: DefaultResourceServerConfigResponse | undefined},
}));

const twoRowsResponse: ResourceServerListResponse = {
  totalResults: 2,
  startIndex: 0,
  count: 2,
  resourceServers: [
    {
      id: 'rs-1',
      name: 'Payments API',
      identifier: 'https://api.example.com',
      ouId: 'ou-1',
      delimiter: ':',
      type: 'API',
    },
    {
      id: 'rs-2',
      name: 'System MCP',
      identifier: 'https://mcp.example.com',
      ouId: 'ou-1',
      delimiter: '/',
      type: 'MCP',
      isReadOnly: true,
    },
  ],
};

describe('ResourceServersList', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetResourceServers.mockReturnValue({
      data: twoRowsResponse,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });
    mockUseGetDefaultResourceServer.mockReturnValue({
      data: {readOnly: {}, writable: {}, merged: {resourceServerId: 'rs-1'}},
    });
  });

  it('renders the type chip label for a normal row', () => {
    renderWithProviders(<ResourceServersList />);

    expect(screen.getByText('API')).toBeInTheDocument();
  });

  it('renders the type chip label for a read-only row', () => {
    renderWithProviders(<ResourceServersList />);

    expect(screen.getByText('MCP')).toBeInTheDocument();
  });

  it('shows Edit and Delete buttons for a normal row', () => {
    renderWithProviders(<ResourceServersList />);

    expect(screen.getByRole('button', {name: 'Edit'})).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Delete'})).toBeInTheDocument();
  });

  it('shows a Read Only button for the read-only row', () => {
    renderWithProviders(<ResourceServersList />);

    expect(screen.getByRole('button', {name: 'Read Only'})).toBeInTheDocument();
  });

  it('does not show a Delete button for the read-only row', () => {
    renderWithProviders(<ResourceServersList />);

    const deleteButtons = screen.queryAllByRole('button', {name: 'Delete'});
    expect(deleteButtons).toHaveLength(1);
  });

  it('shows a Default badge on the current default row', () => {
    renderWithProviders(<ResourceServersList />);

    expect(screen.getByText('Default')).toBeInTheDocument();
  });

  it('does not show a Default badge when no default is set', () => {
    mockUseGetDefaultResourceServer.mockReturnValue({data: {readOnly: {}, writable: {}, merged: {}}});

    renderWithProviders(<ResourceServersList />);

    expect(screen.queryByText('Default')).not.toBeInTheDocument();
  });

  it('renders the resolved catalog message in the error state, never the raw server text, when the fetch fails', () => {
    mockUseGetResourceServers.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('raw backend list failure detail'),
      refetch: mockRefetch,
    });

    renderWithProviders(<ResourceServersList />);

    expect(screen.getByText('Failed to load resource servers')).toBeInTheDocument();
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.queryByText('raw backend list failure detail')).not.toBeInTheDocument();
  });

  it('retries the fetch when Refresh is clicked in the error state', () => {
    mockUseGetResourceServers.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('Network error'),
      refetch: mockRefetch,
    });

    renderWithProviders(<ResourceServersList />);

    fireEvent.click(screen.getByRole('button', {name: /Refresh/i}));

    expect(mockRefetch).toHaveBeenCalled();
  });
});
