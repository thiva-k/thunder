/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {renderHook, waitFor} from '@testing-library/react';

import useDeleteUser from '../useDeleteUser';

describe('useDeleteUser', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useDeleteUser());

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.deleteUser).toBe('function');
  });

  it('should delete a user successfully', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      status: 204,
    });

    const {result} = renderHook(() => useDeleteUser());

    const deleteResult = await result.current.deleteUser('user-123');

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(deleteResult).toBe(true);
    expect(result.current.error).toBeNull();
    expect(global.fetch).toHaveBeenCalledWith('https://localhost:8090/users/user-123', {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
    });
  });

  it('should handle API error with JSON response', async () => {
    const apiErrorResponse = {
      code: 'NOT_FOUND',
      message: 'User not found',
      description: 'The user with the given ID does not exist',
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: async () => apiErrorResponse,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useDeleteUser());

    try {
      await result.current.deleteUser('user-123');
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'User not found',
        description: 'An error occurred while deleting the user',
      });
    });
  });

  it('should handle API error without JSON response', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      text: async () => 'Server error occurred',
      headers: new Headers({'content-type': 'text/plain'}),
    });

    const {result} = renderHook(() => useDeleteUser());

    try {
      await result.current.deleteUser('user-123');
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'Internal Server Error',
        description: 'An error occurred while deleting the user',
      });
    });
  });

  it('should handle network error', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useDeleteUser());

    try {
      await result.current.deleteUser('user-123');
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'Network error',
        description: 'An error occurred while deleting the user',
      });
    });
  });

  it('should set loading state correctly during request', async () => {
    let resolveRequest: () => void;
    const requestPromise = new Promise<void>((resolve) => {
      resolveRequest = resolve;
    });

    (global.fetch as ReturnType<typeof vi.fn>).mockImplementationOnce(() =>
      requestPromise.then(() => ({
        ok: true,
        status: 204,
      })),
    );

    const {result} = renderHook(() => useDeleteUser());

    const promise = result.current.deleteUser('user-123');

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    resolveRequest!();
    await promise;

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should clear previous error on new delete attempt', async () => {
    const apiErrorResponse = {
      code: 'NOT_FOUND',
      message: 'User not found',
      description: 'The user with the given ID does not exist',
    };

    (global.fetch as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({
        ok: false,
        status: 404,
        json: async () => apiErrorResponse,
        headers: new Headers({'content-type': 'application/json'}),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 204,
      });

    const {result} = renderHook(() => useDeleteUser());

    try {
      await result.current.deleteUser('user-123');
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'User not found',
        description: 'An error occurred while deleting the user',
      });
    });

    await result.current.deleteUser('user-456');

    await waitFor(() => {
      expect(result.current.error).toBeNull();
    });
  });

  it('should handle multiple delete operations', async () => {
    (global.fetch as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({
        ok: true,
        status: 204,
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 204,
      });

    const {result} = renderHook(() => useDeleteUser());

    const deleteResult1 = await result.current.deleteUser('user-123');
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(deleteResult1).toBe(true);

    const deleteResult2 = await result.current.deleteUser('user-456');
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(deleteResult2).toBe(true);

    expect(global.fetch).toHaveBeenCalledTimes(2);
  });
});
