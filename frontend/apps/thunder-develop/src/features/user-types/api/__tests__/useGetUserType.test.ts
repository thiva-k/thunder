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

import useGetUserType from '../useGetUserType';
import type {ApiUserSchema} from '../../types/user-types';

describe('useGetUserType', () => {
  const mockUserSchema: ApiUserSchema = {
    id: '123',
    name: 'TestUserType',
    schema: {
      username: {
        type: 'string',
        required: true,
      },
      email: {
        type: 'string',
        required: false,
      },
    },
  };

  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with default state when no id is provided', () => {
    const {result} = renderHook(() => useGetUserType());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should fetch user type when id is provided', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
      expect(result.current.error).toBeNull();
      expect(result.current.loading).toBe(false);
    });

    expect(global.fetch).toHaveBeenCalledWith('https://localhost:8090/user-schemas/123');
  });

  it('should set loading state during fetch', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockImplementation(
      () =>
        new Promise((resolve) => {
          setTimeout(
            () =>
              resolve({
                ok: true,
                json: async () => mockUserSchema,
              }),
            100,
          );
        }),
    );

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

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

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPE_ERROR',
        message: 'User type not found',
        description: 'Failed to fetch user type',
      });
      expect(result.current.data).toBeNull();
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

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPE_ERROR',
        message: 'HTTP error! status: 500',
        description: 'Failed to fetch user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle network error', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPE_ERROR',
        message: 'Network error',
        description: 'Failed to fetch user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should refetch when refetch is called', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => ({...mockUserSchema, name: 'UpdatedUserType'}),
    });

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.data).toEqual({...mockUserSchema, name: 'UpdatedUserType'});
    });

    expect(global.fetch).toHaveBeenCalledTimes(2);
  });

  it('should refetch with new id when provided to refetch', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    const newMockUserSchema = {...mockUserSchema, id: '456', name: 'NewUserType'};

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => newMockUserSchema,
    });

    await result.current.refetch('456');

    await waitFor(() => {
      expect(result.current.data).toEqual(newMockUserSchema);
    });

    expect(global.fetch).toHaveBeenCalledWith('https://localhost:8090/user-schemas/456');
  });

  it('should not fetch if refetch is called without id when no id is provided', async () => {
    const {result} = renderHook(() => useGetUserType());

    await result.current.refetch();

    expect(global.fetch).not.toHaveBeenCalled();
  });

  it('should prevent double-fetch in React Strict Mode', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result, rerender} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    // Simulate React Strict Mode double render
    rerender();

    await waitFor(() => {
      // Should only fetch once, not twice
      expect(global.fetch).toHaveBeenCalledTimes(1);
    });
  });

  it('should fetch again when id changes', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result, rerender} = renderHook(({id}) => useGetUserType(id), {
      initialProps: {id: '123'},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    const newMockUserSchema = {...mockUserSchema, id: '456', name: 'NewUserType'};

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => newMockUserSchema,
    });

    rerender({id: '456'});

    await waitFor(() => {
      expect(result.current.data).toEqual(newMockUserSchema);
    });

    expect(global.fetch).toHaveBeenCalledTimes(2);
    expect(global.fetch).toHaveBeenNthCalledWith(1, 'https://localhost:8090/user-schemas/123');
    expect(global.fetch).toHaveBeenNthCalledWith(2, 'https://localhost:8090/user-schemas/456');
  });

  it('should clear data when id changes to undefined', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result, rerender} = renderHook(({id}) => useGetUserType(id), {
      initialProps: {id: '123' as string | undefined},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    rerender({id: undefined});

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
  });
});
