// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {renderWithProviders, screen, fireEvent, waitFor} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import type {ResourceServer} from '../../models/resource-server';
import CreateResourceServerPage from '../CreateResourceServerPage';

const mockNavigate = vi.fn();

vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('@thunderid/react', async (importOriginal) => {
  const actual = await importOriginal<Record<string, unknown>>();
  return {
    ...actual,
    useThunderID: () => ({http: {request: vi.fn()}}),
  };
});

const mockShowToast = vi.fn();

vi.mock('@thunderid/contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunderid/contexts')>();
  return {
    ...actual,
    useConfig: () => ({getServerUrl: () => 'http://localhost:8090'}),
    useToast: () => ({showToast: mockShowToast}),
  };
});

vi.mock('@thunderid/logger/react', () => ({
  useLogger: () => ({error: vi.fn(), info: vi.fn(), debug: vi.fn()}),
}));

vi.mock('@thunderid/utils', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunderid/utils')>();
  return {
    ...actual,
    generateRandomHumanReadableIdentifiers: () => ['Alpha Service', 'Beta Platform'],
  };
});

const mockCreateResourceServerMutate = vi.fn();
const mockCreateResourceServerReset = vi.fn();
const mockUseCreateResourceServer = vi.fn(() => ({
  mutate: mockCreateResourceServerMutate,
  isPending: false,
  isError: false,
  reset: mockCreateResourceServerReset,
}));

vi.mock('../../api/useCreateResourceServer', () => ({
  default: () => mockUseCreateResourceServer(),
}));

vi.mock('@thunderid/configure-organization-units', () => ({
  useHasMultipleOUs: () => ({
    hasMultipleOUs: false,
    isLoading: false,
    ouList: [{id: 'ou-1', name: 'Default', handle: 'default', parent: null}],
  }),
  OrganizationUnitTreePicker: () => <div data-testid="ou-tree-picker" />,
}));

describe('CreateResourceServerPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseCreateResourceServer.mockReturnValue({
      mutate: mockCreateResourceServerMutate,
      isPending: false,
      isError: false,
      reset: mockCreateResourceServerReset,
    });
  });

  it('renders the Type step initially', () => {
    renderWithProviders(<CreateResourceServerPage />);

    expect(screen.getByText(/What type of resource server are you adding/i)).toBeInTheDocument();
  });

  it('renders the type cards in the Type step', () => {
    renderWithProviders(<CreateResourceServerPage />);

    expect(screen.getAllByRole('button', {name: /API|MCP|Custom/i}).length).toBeGreaterThanOrEqual(1);
  });

  it('advances to the Name step after selecting a type', async () => {
    renderWithProviders(<CreateResourceServerPage />);

    const apiCard = screen.getByRole('button', {name: /API/i});
    fireEvent.click(apiCard);

    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /resource server name/i})).toBeInTheDocument();
    });
  });

  it('shows the Name step with name and identifier fields after navigating to it', async () => {
    renderWithProviders(<CreateResourceServerPage />);

    const apiCard = screen.getByRole('button', {name: /API/i});
    fireEvent.click(apiCard);
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /resource server name/i})).toBeInTheDocument();
      expect(screen.getByRole('textbox', {name: /identifier/i})).toBeInTheDocument();
    });
  });

  it('keeps Continue disabled on the Name step until both name and identifier are filled', async () => {
    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /API/i}));
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /resource server name/i})).toBeInTheDocument();
    });

    fireEvent.change(screen.getByRole('textbox', {name: /resource server name/i}), {
      target: {value: 'Payments API'},
    });

    expect(screen.getByRole('button', {name: /Continue/i})).toBeDisabled();

    fireEvent.change(screen.getByRole('textbox', {name: /identifier/i}), {
      target: {value: 'https://api.example.com'},
    });

    expect(screen.getByRole('button', {name: /Continue/i})).toBeEnabled();
  });

  it('advances to the Separator step after filling the name and identifier and clicking Next', async () => {
    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /API/i}));
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /resource server name/i})).toBeInTheDocument();
    });

    fireEvent.change(screen.getByRole('textbox', {name: /resource server name/i}), {
      target: {value: 'Payments API'},
    });
    fireEvent.change(screen.getByRole('textbox', {name: /identifier/i}), {
      target: {value: 'https://api.example.com'},
    });

    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });
  });

  it('shows the permission preview in the Separator step', async () => {
    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /API/i}));
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /resource server name/i})).toBeInTheDocument();
    });

    fireEvent.change(screen.getByRole('textbox', {name: /resource server name/i}), {
      target: {value: 'Payments API'},
    });
    fireEvent.change(screen.getByRole('textbox', {name: /identifier/i}), {
      target: {value: 'https://api.example.com'},
    });

    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByText('<resource>:<action>')).toBeInTheDocument();
    });
  });

  it('shows the MCP server name label after selecting the MCP type', async () => {
    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /MCP/i}));
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /mcp server name/i})).toBeInTheDocument();
    });
  });

  it('shows the Create MCP server submit button on the last step after selecting the MCP type', async () => {
    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /MCP/i}));
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /mcp server name/i})).toBeInTheDocument();
    });

    fireEvent.change(screen.getByRole('textbox', {name: /mcp server name/i}), {
      target: {value: 'Payments MCP'},
    });
    fireEvent.change(screen.getByRole('textbox', {name: /identifier/i}), {
      target: {value: 'https://mcp.example.com'},
    });

    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /Create MCP server/i})).toBeInTheDocument();
    });
  });

  it('shows the Create resource server submit button on the last step after selecting the API type', async () => {
    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /API/i}));
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /resource server name/i})).toBeInTheDocument();
    });

    fireEvent.change(screen.getByRole('textbox', {name: /resource server name/i}), {
      target: {value: 'Payments API'},
    });
    fireEvent.change(screen.getByRole('textbox', {name: /identifier/i}), {
      target: {value: 'https://api.example.com'},
    });

    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /Create resource server/i})).toBeInTheDocument();
    });
  });

  it('sends the identifier in the create payload', async () => {
    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /API/i}));
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /resource server name/i})).toBeInTheDocument();
    });

    fireEvent.change(screen.getByRole('textbox', {name: /resource server name/i}), {
      target: {value: 'Payments API'},
    });
    fireEvent.change(screen.getByRole('textbox', {name: /identifier/i}), {
      target: {value: 'https://api.example.com'},
    });

    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /Create resource server/i})).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', {name: /Create resource server/i}));

    expect(mockCreateResourceServerMutate).toHaveBeenCalledWith(
      expect.objectContaining({name: 'Payments API', identifier: 'https://api.example.com'}),
      expect.any(Object),
    );
    expect(mockCreateResourceServerMutate.mock.calls[0][0]).not.toHaveProperty('handle');
  });

  it('shows the MCP server created success toast after a successful MCP server creation', async () => {
    mockCreateResourceServerMutate.mockImplementationOnce(
      (_payload: unknown, options: {onSuccess: (created: ResourceServer) => void}) => {
        options.onSuccess({id: 'mcp-rs-1'} as ResourceServer);
      },
    );

    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /MCP/i}));
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /mcp server name/i})).toBeInTheDocument();
    });

    fireEvent.change(screen.getByRole('textbox', {name: /mcp server name/i}), {
      target: {value: 'Payments MCP'},
    });
    fireEvent.change(screen.getByRole('textbox', {name: /identifier/i}), {
      target: {value: 'https://mcp.example.com'},
    });

    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /Create MCP server/i})).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', {name: /Create MCP server/i}));

    await waitFor(() => {
      expect(mockShowToast).toHaveBeenCalledWith('MCP server created successfully.', 'success');
      expect(mockNavigate).toHaveBeenCalledWith('/resource-servers/mcp-rs-1?tab=resources');
    });
  });

  it('displays the resolved catalog message, never the raw server text, when create fails', async () => {
    const rawServerMessage = 'raw backend create failure detail';
    mockCreateResourceServerMutate.mockImplementation((_payload: unknown, options: {onError: (err: Error) => void}) => {
      options.onError(new Error(rawServerMessage));
    });

    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /API/i}));
    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('textbox', {name: /resource server name/i})).toBeInTheDocument();
    });

    fireEvent.change(screen.getByRole('textbox', {name: /resource server name/i}), {
      target: {value: 'Payments API'},
    });
    fireEvent.change(screen.getByRole('textbox', {name: /identifier/i}), {
      target: {value: 'https://api.example.com'},
    });

    fireEvent.click(screen.getByRole('button', {name: /Continue/i}));

    await waitFor(() => {
      expect(screen.getByRole('button', {name: /Create resource server/i})).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', {name: /Create resource server/i}));

    await waitFor(() => {
      expect(screen.getByText('Failed to create resource server. Please try again.')).toBeInTheDocument();
    });
    expect(screen.queryByText(rawServerMessage)).not.toBeInTheDocument();
  });

  it('resets the create error when the type is changed after a failed create', () => {
    mockUseCreateResourceServer.mockReturnValue({
      mutate: mockCreateResourceServerMutate,
      isPending: false,
      isError: true,
      reset: mockCreateResourceServerReset,
    });

    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /API/i}));

    expect(mockCreateResourceServerReset).toHaveBeenCalled();
  });

  it('does not reset a pending create mutation when the type is changed', () => {
    mockUseCreateResourceServer.mockReturnValue({
      mutate: mockCreateResourceServerMutate,
      isPending: true,
      isError: false,
      reset: mockCreateResourceServerReset,
    });

    renderWithProviders(<CreateResourceServerPage />);

    fireEvent.click(screen.getByRole('button', {name: /API/i}));

    expect(mockCreateResourceServerReset).not.toHaveBeenCalled();
  });
});
