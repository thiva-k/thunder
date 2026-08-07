// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import userEvent from '@testing-library/user-event';
import {render, screen} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import ThemeCreatePage from '../ThemeCreatePage';

const {mockNavigate, mockMutate, mockReset, mockCreateThemeState} = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockMutate: vi.fn(),
  mockReset: vi.fn(),
  mockCreateThemeState: {isPending: false, isError: false},
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
    t: (key: string, fallback?: string | {defaultValue?: string}) => {
      if (typeof fallback === 'string') return fallback || key;
      if (fallback && typeof fallback === 'object') return fallback.defaultValue ?? key;
      return key;
    },
  }),
}));

vi.mock('@thunderid/design', () => ({
  useCreateTheme: () => ({
    mutate: mockMutate,
    reset: mockReset,
    isPending: mockCreateThemeState.isPending,
    isError: mockCreateThemeState.isError,
  }),
  useGetTheme: () => ({data: undefined}),
  useGetThemes: () => ({data: {themes: []}}),
}));

vi.mock('../../../../components/GatePreview/GatePreview', () => ({
  default: () => <div data-testid="preview" />,
}));

vi.mock('../../components/create-theme/ConfigureThemeName', () => ({
  default: ({
    onThemeNameChange,
    onReadyChange,
  }: {
    onThemeNameChange: (name: string) => void;
    onReadyChange?: (isReady: boolean) => void;
  }) => (
    <div data-testid="step-name">
      <button type="button" onClick={() => onThemeNameChange('My Theme')}>
        Set Name
      </button>
      <button type="button" onClick={() => onReadyChange?.(true)}>
        Set Ready
      </button>
    </div>
  ),
}));

vi.mock('../../components/create-theme/ConfigureThemeColor', () => ({
  default: ({onPrimaryColorChange}: {onPrimaryColorChange: (color: string) => void}) => (
    <div data-testid="step-color">
      <button type="button" onClick={() => onPrimaryColorChange('#123456')}>
        Set Color
      </button>
    </div>
  ),
}));

describe('ThemeCreatePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateThemeState.isPending = false;
    mockCreateThemeState.isError = false;
  });

  it('renders the name step by default', () => {
    render(<ThemeCreatePage />);
    expect(screen.getByTestId('step-name')).toBeInTheDocument();
  });

  it('moves to the color step once the name step reports ready and continue is clicked', async () => {
    const user = userEvent.setup();
    render(<ThemeCreatePage />);

    await user.click(screen.getByText('Set Ready'));
    await user.click(screen.getByRole('button', {name: /Continue/i}));

    expect(screen.getByTestId('step-color')).toBeInTheDocument();
  });

  it('calls mutate with a kebab-cased handle when Create Theme is clicked', async () => {
    const user = userEvent.setup();
    render(<ThemeCreatePage />);

    await user.click(screen.getByText('Set Ready'));
    await user.click(screen.getByText('Set Name'));
    await user.click(screen.getByRole('button', {name: /Continue/i}));
    await user.click(screen.getByRole('button', {name: /Create Theme/i}));

    expect(mockMutate).toHaveBeenCalledWith(
      expect.objectContaining({handle: 'my-theme', displayName: 'My Theme'}),
      expect.any(Object),
    );
  });

  it('displays a resolved error message on create failure, never the raw server text', async () => {
    mockMutate.mockImplementation((_data, opts: {onError: (err: Error) => void}) => {
      opts.onError(new Error('Network error'));
    });

    const user = userEvent.setup();
    render(<ThemeCreatePage />);

    await user.click(screen.getByText('Set Ready'));
    await user.click(screen.getByRole('button', {name: /Continue/i}));
    await user.click(screen.getByRole('button', {name: /Create Theme/i}));

    expect(await screen.findByText('Failed to create theme. Please try again.')).toBeInTheDocument();
    expect(screen.queryByText('Network error')).not.toBeInTheDocument();
  });

  it('clears a stale create error and resets the mutation as soon as the name field changes', async () => {
    // Mirrors the mutation's own bookkeeping: a real TanStack mutation flips isError to true
    // before onError fires, so the next render already reflects it.
    mockMutate.mockImplementation((_data, opts: {onError: (err: Error) => void}) => {
      mockCreateThemeState.isError = true;
      opts.onError(new Error('Network error'));
    });

    const user = userEvent.setup();
    render(<ThemeCreatePage />);

    await user.click(screen.getByText('Set Ready'));
    await user.click(screen.getByRole('button', {name: /Continue/i}));
    await user.click(screen.getByRole('button', {name: /Create Theme/i}));

    expect(await screen.findByText('Failed to create theme. Please try again.')).toBeInTheDocument();

    // Go back to the name step to trigger the name field-change handler.
    await user.click(screen.getByRole('button', {name: /Back/i}));
    await user.click(screen.getByText('Set Name'));

    expect(mockReset).toHaveBeenCalledTimes(1);
    expect(screen.queryByText('Failed to create theme. Please try again.')).not.toBeInTheDocument();
  });

  it('clears a stale create error and resets the mutation as soon as the color field changes', async () => {
    // Mirrors the mutation's own bookkeeping: a real TanStack mutation flips isError to true
    // before onError fires, so the next render already reflects it.
    mockMutate.mockImplementation((_data, opts: {onError: (err: Error) => void}) => {
      mockCreateThemeState.isError = true;
      opts.onError(new Error('Network error'));
    });

    const user = userEvent.setup();
    render(<ThemeCreatePage />);

    await user.click(screen.getByText('Set Ready'));
    await user.click(screen.getByRole('button', {name: /Continue/i}));
    await user.click(screen.getByRole('button', {name: /Create Theme/i}));

    expect(await screen.findByText('Failed to create theme. Please try again.')).toBeInTheDocument();

    await user.click(screen.getByText('Set Color'));

    expect(mockReset).toHaveBeenCalledTimes(1);
    expect(screen.queryByText('Failed to create theme. Please try again.')).not.toBeInTheDocument();
  });

  it('does not reset the mutation on field change while a create is still pending', async () => {
    // A mutation cannot be isPending and isError simultaneously, so isError stays false here —
    // the same state a real in-flight TanStack mutation would report.
    mockCreateThemeState.isPending = true;

    const user = userEvent.setup();
    render(<ThemeCreatePage />);

    await user.click(screen.getByText('Set Ready'));
    await user.click(screen.getByRole('button', {name: /Continue/i}));
    await user.click(screen.getByText('Set Color'));

    expect(mockReset).not.toHaveBeenCalled();
  });
});
