// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {renderHook, waitFor} from '@thunderid/test-utils';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import useCallerOuId from '../useCallerOuId';

const getDecodedIdTokenMock = vi.fn();

vi.mock('@thunderid/react', () => ({
  useThunderID: () => ({getDecodedIdToken: getDecodedIdTokenMock}),
}));

describe('useCallerOuId', () => {
  beforeEach(() => {
    getDecodedIdTokenMock.mockReset();
  });

  it('is loading with an empty id until the token is read', async () => {
    getDecodedIdTokenMock.mockResolvedValue({ouId: 'ou-root'});

    const {result} = renderHook(() => useCallerOuId());

    expect(result.current).toEqual({ouId: '', isLoading: true});
    await waitFor(() => expect(result.current.isLoading).toBe(false));
  });

  it('returns the ouId claim of the ID token', async () => {
    getDecodedIdTokenMock.mockResolvedValue({sub: 'admin', ouId: 'ou-root'});

    const {result} = renderHook(() => useCallerOuId());

    await waitFor(() => expect(result.current.ouId).toBe('ou-root'));
    expect(result.current.isLoading).toBe(false);
  });

  it('returns an empty id when the token has no ouId claim', async () => {
    getDecodedIdTokenMock.mockResolvedValue({sub: 'admin'});

    const {result} = renderHook(() => useCallerOuId());

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.ouId).toBe('');
  });

  it('returns an empty id when the claim is not a string', async () => {
    getDecodedIdTokenMock.mockResolvedValue({ouId: 42});

    const {result} = renderHook(() => useCallerOuId());

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.ouId).toBe('');
  });
});
