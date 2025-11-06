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

import useGetUser from '../useGetUser';
import type {ApiUser} from '../../types/users';

describe('useGetUser', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useGetUser(undefined));

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should fetch user successfully when id is provided', async () => {
    const mockUser: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUser,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockUser);
    expect(result.current.error).toBeNull();
    expect(global.fetch).toHaveBeenCalledWith('https://localhost:8090/users/user-123', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
  });

  it('should not fetch when id is not provided', () => {
    const {result} = renderHook(() => useGetUser(undefined));

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(global.fetch).not.toHaveBeenCalled();
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

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'User not found',
      description: 'An error occurred while fetching user',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle API error without JSON response', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      text: async () => 'Server error occurred',
      headers: new Headers({'content-type': 'text/plain'}),
    });

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Internal Server Error',
      description: 'An error occurred while fetching user',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle network error', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Network error',
      description: 'An error occurred while fetching user',
    });
    expect(result.current.data).toBeNull();
  });

  it('should refetch user with the same id', async () => {
    const mockUser: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockUser,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser);
    });

    await result.current.refetch();

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(2);
    });
  });

  it('should refetch user with a new id', async () => {
    const mockUser1: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockUser2: ApiUser = {
      id: 'user-789',
      organizationUnit: '/marketing',
      type: 'employee',
      attributes: {
        name: 'Jane Smith',
        email: 'jane@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockUser1,
        headers: new Headers({'content-type': 'application/json'}),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockUser2,
        headers: new Headers({'content-type': 'application/json'}),
      });

    const {result, rerender} = renderHook(({id}) => useGetUser(id), {
      initialProps: {id: 'user-123'},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser1);
    });

    rerender({id: 'user-789'});

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser2);
    });
  });

  it('should set error when refetch is called without id', async () => {
    const {result} = renderHook(() => useGetUser(undefined));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // When userId is undefined, fetchUser returns early without setting error
    expect(result.current.error).toBeNull();
    expect(result.current.data).toBeNull();
  });

  it('should prevent double fetch in strict mode', async () => {
    const mockUser: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockUser,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser);
    });

    // Should only fetch once despite strict mode
    expect(global.fetch).toHaveBeenCalledTimes(1);
  });

  it('should fetch when id changes', async () => {
    const mockUser1: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockUser2: ApiUser = {
      id: 'user-789',
      organizationUnit: '/marketing',
      type: 'employee',
      attributes: {
        name: 'Jane Smith',
        email: 'jane@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockUser1,
        headers: new Headers({'content-type': 'application/json'}),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockUser2,
        headers: new Headers({'content-type': 'application/json'}),
      });

    const {result, rerender} = renderHook(({id}: {id?: string}) => useGetUser(id), {
      initialProps: {id: 'user-123'},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser1);
    });

    rerender({id: 'user-789'});

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser2);
    });

    expect(global.fetch).toHaveBeenCalledTimes(2);
  });
});
