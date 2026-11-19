// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {waitFor, renderHook} from '@thunderid/test-utils';
import {FlowExecutionFailure} from '@thunderid/utils';
import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest';
import useSendOwnerVerification from '../useSendOwnerVerification';

vi.mock('@thunderid/react', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@thunderid/react')>()),
  useThunderID: vi.fn(),
}));

const {mockShowToast} = vi.hoisted(() => ({mockShowToast: vi.fn()}));

vi.mock('@thunderid/contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunderid/contexts')>();
  return {
    ...actual,
    useConfig: vi.fn(),
    useToast: vi.fn().mockReturnValue({showToast: mockShowToast}),
  };
});

const {useThunderID} = await import('@thunderid/react');
const {useConfig} = await import('@thunderid/contexts');

describe('useSendOwnerVerification', () => {
  let mockHttpRequest: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockHttpRequest = vi.fn();
    vi.mocked(useThunderID).mockReturnValue({
      http: {request: mockHttpRequest},
    } as unknown as ReturnType<typeof useThunderID>);
    vi.mocked(useConfig).mockReturnValue({
      getServerUrl: () => 'https://api.test.com',
    } as ReturnType<typeof useConfig>);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('starts the configured owner verification flow for the agent and names the owner the link went to', async () => {
    mockHttpRequest.mockResolvedValueOnce({
      data: {flowStatus: 'INCOMPLETE', data: {additionalData: {maskedDestination: 'j•••@example.com'}}},
    });
    const {result} = renderHook(() => useSendOwnerVerification());

    result.current.mutate({agentId: 'agent-1', handle: 'acme-owner-verification', owner: 'jane@example.com'});

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://api.test.com/flow/execute',
        method: 'POST',
        data: {
          flowType: 'ADMIN_INITIATED',
          handle: 'acme-owner-verification',
          inputs: {subject: 'agent-1'},
          verbose: true,
        },
      }),
    );
    // The flow's masked destination hides nothing from an administrator who can read the agent, so the owner is named.
    expect(mockShowToast).toHaveBeenCalledWith('Verification link sent to jane@example.com', 'success');
  });

  it('reports a flow that refused to send the link', async () => {
    mockHttpRequest.mockResolvedValueOnce({
      data: {flowStatus: 'ERROR', error: {code: 'FET-1112', message: {key: 'k', defaultValue: 'Refused'}}},
    });
    const {result} = renderHook(() => useSendOwnerVerification());

    result.current.mutate({agentId: 'agent-1', handle: 'default-agent-owner-verification-flow'});

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(FlowExecutionFailure);
    expect((result.current.error as FlowExecutionFailure).code).toBe('FET-1112');
    // The caller shows the failure where the link was sent from, so the hook raises no toast.
    expect(mockShowToast).not.toHaveBeenCalled();
  });

  it('reports a request that failed', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Network down'));
    const {result} = renderHook(() => useSendOwnerVerification());

    result.current.mutate({agentId: 'agent-1', handle: 'default-agent-owner-verification-flow'});

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(mockShowToast).not.toHaveBeenCalled();
  });
});
