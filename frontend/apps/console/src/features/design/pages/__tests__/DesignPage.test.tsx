// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {fireEvent, render, screen, waitFor} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import DesignPage from '../DesignPage';

const {mockNavigate, mockShowToast, mockCreateLayout, mockRefetchThemes, mockRefetchLayouts} = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockShowToast: vi.fn(),
  mockCreateLayout: vi.fn(),
  mockRefetchThemes: vi.fn(),
  mockRefetchLayouts: vi.fn(),
}));

vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string | {defaultValue?: string} | Record<string, unknown>) => {
      if (typeof fallback === 'string') return fallback || key;
      if (fallback && typeof fallback === 'object' && 'defaultValue' in fallback) {
        return (fallback.defaultValue as string) ?? key;
      }
      return key;
    },
  }),
}));

vi.mock('@thunderid/contexts', async () => {
  const actual = await vi.importActual<typeof import('@thunderid/contexts')>('@thunderid/contexts');
  return {
    ...actual,
    useToast: () => ({showToast: mockShowToast}),
  };
});

let themesError: Error | null = null;
let layoutsError: Error | null = null;

vi.mock('@thunderid/design', () => ({
  useGetThemes: () => ({
    data: {themes: []},
    isLoading: false,
    error: themesError,
    refetch: mockRefetchThemes,
  }),
  useGetLayouts: () => ({
    data: {layouts: []},
    error: layoutsError,
    refetch: mockRefetchLayouts,
  }),
  useCreateLayout: () => ({mutateAsync: mockCreateLayout}),
  useDeleteTheme: () => ({mutate: vi.fn(), reset: vi.fn(), isPending: false, isError: false}),
  useGetThemeUsages: () => ({data: undefined, isLoading: false}),
}));

describe('DesignPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    themesError = null;
    layoutsError = null;
  });

  it('renders the themes and layouts section headers', () => {
    render(<DesignPage />);
    expect(screen.getByText('Themes')).toBeInTheDocument();
    expect(screen.getByText('Layouts')).toBeInTheDocument();
  });

  it('renders an inline read error for themes without affecting the layouts section', () => {
    themesError = new Error('Network error');

    render(<DesignPage />);

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.queryByText('Network error')).not.toBeInTheDocument();
    // Layouts section still renders normally
    expect(screen.getByText('Centered')).toBeInTheDocument();
  });

  it('retries the themes query when the read error action is clicked', () => {
    themesError = new Error('Network error');

    render(<DesignPage />);

    fireEvent.click(screen.getByRole('button', {name: /Refresh/i}));

    expect(mockRefetchThemes).toHaveBeenCalledTimes(1);
  });

  it('renders an inline read error for layouts without affecting the themes section', () => {
    layoutsError = new Error('Network error');

    render(<DesignPage />);

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.queryByText('Network error')).not.toBeInTheDocument();
    expect(screen.queryByText('Centered')).not.toBeInTheDocument();
  });

  it('shows an error toast, resolved rather than raw, when creating a layout preset fails', async () => {
    mockCreateLayout.mockRejectedValue(new Error('Network error'));

    render(<DesignPage />);

    fireEvent.click(screen.getByRole('button', {name: 'Centered'}));

    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith('Failed to create layout. Please try again.', 'error');
    });
  });

  it('navigates to the created layout on success', async () => {
    mockCreateLayout.mockResolvedValue({id: 'layout-new'});

    render(<DesignPage />);

    fireEvent.click(screen.getByRole('button', {name: 'Centered'}));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/design/layouts/layout-new');
    });
    expect(mockShowToast).not.toHaveBeenCalled();
  });
});
