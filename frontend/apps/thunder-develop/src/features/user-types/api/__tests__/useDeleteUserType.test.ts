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

import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest';
import {renderHook, waitFor} from '@testing-library/react';

import useDeleteUserType from '../useDeleteUserType';

describe('useDeleteUserType', () => {
  const mockUserTypeId = '123';

  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with default state', () => {
    const {result} = renderHook(() => useDeleteUserType());

    expect(result.current.error).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(typeof result.current.deleteUserType).toBe('function');
    expect(typeof result.current.reset).toBe('function');
  });

  it('should successfully delete a user type', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
    });

    const {result} = renderHook(() => useDeleteUserType());

    const deleteResult = await result.current.deleteUserType(mockUserTypeId);

    expect(deleteResult).toBe(true);

    await waitFor(() => {
      expect(result.current.error).toBeNull();
      expect(result.current.loading).toBe(false);
    });

    expect(global.fetch).toHaveBeenCalledWith(`https://localhost:8090/user-schemas/${mockUserTypeId}`, {
      method: 'DELETE',
    });
  });

  it('should set loading state during deletion', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockImplementation(
      () =>
        new Promise((resolve) => {
          setTimeout(
            () =>
              resolve({
                ok: true,
              }),
            100,
          );
        }),
    );

    const {result} = renderHook(() => useDeleteUserType());

    const promise = result.current.deleteUserType(mockUserTypeId);

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    await promise;

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle API error with JSON response', async () => {
    const apiErrorResponse = {
      code: 'NOT_FOUND',
      message: 'User type not found',
      description: 'The specified user type does not exist',
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: async () => apiErrorResponse,
    });

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toThrow('User type not found');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_TYPE_ERROR',
        message: 'User type not found',
        description: 'Failed to delete user type',
      });
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle API error without JSON response', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: async () => {
        throw new Error('Not JSON');
      },
      text: async () => 'Internal Server Error',
    });

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toThrow('HTTP error! status: 500');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_TYPE_ERROR',
        message: 'HTTP error! status: 500',
        description: 'Failed to delete user type',
      });
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle network error', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toThrow('Network error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_TYPE_ERROR',
        message: 'Network error',
        description: 'Failed to delete user type',
      });
      expect(result.current.loading).toBe(false);
    });
  });

  it('should reset error state when reset is called', async () => {
    const apiErrorResponse = {
      code: 'NOT_FOUND',
      message: 'User type not found',
      description: 'The specified user type does not exist',
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: async () => apiErrorResponse,
    });

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toThrow('User type not found');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_TYPE_ERROR',
        message: 'User type not found',
        description: 'Failed to delete user type',
      });
    });

    await waitFor(() => {
      result.current.reset();
    });

    await waitFor(() => {
      expect(result.current.error).toBeNull();
    });
  });

  it('should clear previous error when starting new request', async () => {
    const apiErrorResponse = {
      code: 'NOT_FOUND',
      message: 'Previous error',
      description: 'Previous error description',
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: async () => apiErrorResponse,
    });

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toThrow('Previous error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_TYPE_ERROR',
        message: 'Previous error',
        description: 'Failed to delete user type',
      });
    });

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
    });

    await result.current.deleteUserType(mockUserTypeId);

    await waitFor(() => {
      expect(result.current.error).toBeNull();
    });
  });
});
