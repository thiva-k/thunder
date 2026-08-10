// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {fireEvent, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {renderWithProviders} from '@thunderid/test-utils';
import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest';
import type {CorsConfigResponse} from '../../../models/responses';

const mockUseGetCorsConfig =
  vi.fn<() => {data: CorsConfigResponse | undefined; isLoading: boolean; error: Error | null}>();
vi.mock('../../../api/useGetCorsConfig', () => ({
  default: () => mockUseGetCorsConfig(),
}));

const mockMutate = vi.fn();
vi.mock('../../../api/useUpdateCorsConfig', () => ({
  default: () => ({mutate: mockMutate, isPending: false}),
}));

const {default: CorsSection} = await import('../CorsSection');

function makeData(overrides?: Partial<CorsConfigResponse>): CorsConfigResponse {
  return {
    readOnly: {allowedOrigins: ['https://console.example.com']},
    writable: {allowedOrigins: ['https://app.acme.com']},
    merged: {allowedOrigins: []},
    ...overrides,
  };
}

describe('CorsSection', () => {
  beforeEach(() => {
    mockUseGetCorsConfig.mockReset();
    mockMutate.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('does not render the editor while loading', () => {
    mockUseGetCorsConfig.mockReturnValue({data: undefined, isLoading: true, error: null});
    renderWithProviders(<CorsSection />);
    expect(screen.queryByRole('button', {name: 'Add origin'})).toBeNull();
  });

  it('shows an alert on load error', () => {
    mockUseGetCorsConfig.mockReturnValue({data: undefined, isLoading: false, error: new Error('load failed')});
    renderWithProviders(<CorsSection />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });

  it('renders read-only origins (incl. regex patterns), editable origins, and the Add control', () => {
    mockUseGetCorsConfig.mockReturnValue({
      data: makeData({readOnly: {allowedOrigins: ['https://console.example.com', {regex: '^https://x$'}]}}),
      isLoading: false,
      error: null,
    });
    renderWithProviders(<CorsSection />);

    expect(screen.getByDisplayValue('https://console.example.com')).toHaveAttribute('readonly');
    expect(screen.getByDisplayValue('^https://x$')).toHaveAttribute('readonly');
    expect(screen.getByDisplayValue('https://app.acme.com')).not.toHaveAttribute('readonly');
    expect(screen.getByRole('button', {name: 'Add origin'})).toBeInTheDocument();
    expect(screen.getByText("Some origins are read-only because they're managed declaratively.")).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Remove origin'})).toBeInTheDocument();
  });

  it('removes an editable origin when its delete button is clicked', async () => {
    const user = userEvent.setup();
    mockUseGetCorsConfig.mockReturnValue({
      data: makeData({readOnly: {allowedOrigins: []}, writable: {allowedOrigins: ['https://remove.example.com']}}),
      isLoading: false,
      error: null,
    });
    renderWithProviders(<CorsSection />);

    expect(screen.getByDisplayValue('https://remove.example.com')).toBeInTheDocument();
    await user.click(screen.getByRole('button', {name: 'Remove origin'}));
    expect(screen.queryByDisplayValue('https://remove.example.com')).toBeNull();
  });

  it('adds an editable row when Add origin is clicked', async () => {
    const user = userEvent.setup();
    mockUseGetCorsConfig.mockReturnValue({
      data: makeData({writable: {allowedOrigins: []}}),
      isLoading: false,
      error: null,
    });
    renderWithProviders(<CorsSection />);

    expect(screen.queryByPlaceholderText('https://app.example.com')).toBeNull();
    await user.click(screen.getByRole('button', {name: 'Add origin'}));
    expect(screen.getByPlaceholderText('https://app.example.com')).toBeInTheDocument();
  });

  it('saves the edited origins via the update mutation and clears the unsaved bar on success', async () => {
    const user = userEvent.setup();
    mockMutate.mockImplementation((...args: unknown[]) => {
      const opts = args[1] as {onSuccess?: () => void} | undefined;
      opts?.onSuccess?.();
    });
    mockUseGetCorsConfig.mockReturnValue({
      data: makeData({readOnly: {allowedOrigins: []}, writable: {allowedOrigins: []}}),
      isLoading: false,
      error: null,
    });
    renderWithProviders(<CorsSection />);

    await user.click(screen.getByRole('button', {name: 'Add origin'}));
    await user.type(screen.getByPlaceholderText('https://app.example.com'), 'https://new.example.com');

    const saveButton = await screen.findByRole('button', {name: 'Save changes'});
    await user.click(saveButton);

    expect(mockMutate).toHaveBeenCalledWith(
      expect.objectContaining({data: {allowedOrigins: ['https://new.example.com']}}),
      expect.anything(),
    );
    // onSuccess → reset() clears the overlay, so the unsaved bar disappears.
    await waitFor(() => {
      expect(screen.queryByRole('button', {name: 'Save changes'})).toBeNull();
    });
  });

  it('does not save when a row fails the submit-time validation guard', async () => {
    const user = userEvent.setup();
    mockUseGetCorsConfig.mockReturnValue({
      data: makeData({readOnly: {allowedOrigins: []}, writable: {allowedOrigins: []}}),
      isLoading: false,
      error: null,
    });
    renderWithProviders(<CorsSection />);

    await user.click(screen.getByRole('button', {name: 'Add origin'}));
    // Change without blurring (fireEvent), so the row-level error isn't shown yet and Save is enabled.
    // '(bad' is neither a valid origin nor a compilable regex, so the submit-time guard must block it.
    fireEvent.change(screen.getByPlaceholderText('https://app.example.com'), {target: {value: '(bad'}});
    fireEvent.click(screen.getByRole('button', {name: 'Save changes'}));

    expect(mockMutate).not.toHaveBeenCalled();
  });

  it('blocks Save when a row is a duplicate', async () => {
    const user = userEvent.setup();
    mockUseGetCorsConfig.mockReturnValue({
      data: makeData({readOnly: {allowedOrigins: []}, writable: {allowedOrigins: []}}),
      isLoading: false,
      error: null,
    });
    renderWithProviders(<CorsSection />);

    await user.click(screen.getByRole('button', {name: 'Add origin'}));
    await user.click(screen.getByRole('button', {name: 'Add origin'}));
    const inputs = screen.getAllByPlaceholderText('https://app.example.com');
    await user.type(inputs[0], 'https://dup.example.com');
    await user.type(inputs[1], 'https://dup.example.com');

    const saveButton = await screen.findByRole('button', {name: 'Save changes'});
    expect(saveButton).toBeDisabled();
    expect(mockMutate).not.toHaveBeenCalled();
  });
});
