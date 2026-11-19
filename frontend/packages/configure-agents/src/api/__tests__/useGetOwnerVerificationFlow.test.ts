// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {renderHook, waitFor} from '@thunderid/test-utils';
import {describe, it, expect, beforeEach, vi} from 'vitest';
import useGetOwnerVerificationFlow from '../useGetOwnerVerificationFlow';

const mockHttpRequest = vi.fn();

vi.mock('@thunderid/react', () => ({
  useThunderID: () => ({http: {request: mockHttpRequest}}),
}));

vi.mock('@thunderid/contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunderid/contexts')>();
  return {
    ...actual,
    useConfig: () => ({getServerUrl: () => 'https://api.test.com'}),
  };
});

describe('useGetOwnerVerificationFlow', () => {
  beforeEach(() => {
    mockHttpRequest.mockReset();
  });

  it.each([
    {
      name: 'the merged handle',
      section: {
        readOnly: {agentOwnerVerificationFlow: {defaultHandle: 'default-agent-owner-verification-flow'}},
        merged: {agentOwnerVerificationFlow: {defaultHandle: 'acme-owner-verification'}},
      },
      expected: 'acme-owner-verification',
    },
    {
      name: 'an empty handle when only the user verification flow is set',
      section: {merged: {adminAttributeVerificationFlow: {defaultHandle: 'default-admin-attribute-verification-flow'}}},
      expected: '',
    },
  ])('reads $name from the flow section', async ({section, expected}) => {
    mockHttpRequest.mockResolvedValue({data: section});

    const {result} = renderHook(() => useGetOwnerVerificationFlow(true));

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe(expected);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({url: 'https://api.test.com/server-config/flow', method: 'GET'}),
    );
  });

  it('reads nothing when disabled', () => {
    renderHook(() => useGetOwnerVerificationFlow(false));

    expect(mockHttpRequest).not.toHaveBeenCalled();
  });
});
