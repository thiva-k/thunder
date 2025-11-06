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

import useGetUserTypes from '../useGetUserTypes';
import type {UserSchemaListResponse} from '../../types/user-types';

describe('useGetUserTypes', () => {
  const mockUserSchemaList: UserSchemaListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    schemas: [
      {id: '123', name: 'UserType1'},
      {id: '456', name: 'UserType2'},
    ],
    links: [
      {rel: 'self', href: 'https://localhost:8090/user-schemas'},
    ],
  };

  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize and start fetching', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchemaList,
    });

    const {result} = renderHook(() => useGetUserTypes());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should fetch user types on mount', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchemaList,
    });

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
      expect(result.current.error).toBeNull();
      expect(result.current.loading).toBe(false);
    });

    expect(global.fetch).toHaveBeenCalledWith('https://localhost:8090/user-schemas', expect.any(Object));
  });

  it('should fetch user types with limit parameter', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchemaList,
    });

    renderHook(() => useGetUserTypes({limit: 10}));

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        'https://localhost:8090/user-schemas?limit=10',
        expect.any(Object),
      );
    });
  });

  it('should fetch user types with offset parameter', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchemaList,
    });

    renderHook(() => useGetUserTypes({offset: 5}));

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        'https://localhost:8090/user-schemas?offset=5',
        expect.any(Object),
      );
    });
  });

  it('should fetch user types with both limit and offset parameters', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchemaList,
    });

    renderHook(() => useGetUserTypes({limit: 10, offset: 5}));

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        'https://localhost:8090/user-schemas?limit=10&offset=5',
        expect.any(Object),
      );
    });
  });

  it('should set loading state during fetch', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockImplementation(
      () =>
        new Promise((resolve) => {
          setTimeout(
            () =>
              resolve({
                ok: true,
                json: async () => mockUserSchemaList,
              }),
            100,
          );
        }),
    );

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle API error with JSON response', async () => {
    const mockError = {
      code: 'FETCH_USER_TYPES_ERROR',
      message: 'Failed to fetch user types',
      description: 'An error occurred while fetching user types',
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: async () => mockError,
    });

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.error).toEqual(mockError);
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

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPES_ERROR',
        message: 'HTTP error! status: 500',
        description: 'Internal Server Error',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle network error', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPES_ERROR',
        message: 'Network error',
        description: 'Failed to fetch user types',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should refetch when refetch is called', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchemaList,
    });

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
    });

    const updatedList = {...mockUserSchemaList, totalResults: 3};

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => updatedList,
    });

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.data).toEqual(updatedList);
    });

    expect(global.fetch).toHaveBeenCalledTimes(2);
  });

  it('should abort previous request when params change', async () => {
    let abortSignal: AbortSignal | undefined;

    (global.fetch as ReturnType<typeof vi.fn>).mockImplementation((_url, options) => {
      const requestInit = options as RequestInit | undefined;
      abortSignal = requestInit?.signal ?? undefined;
      return new Promise((resolve) => {
        setTimeout(
          () =>
            resolve({
              ok: true,
              json: async () => mockUserSchemaList,
            }),
          100,
        );
      });
    });

    const {rerender} = renderHook(({params}) => useGetUserTypes(params), {
      initialProps: {params: {limit: 10}},
    });

    // Wait a bit for the first request to start
    await new Promise((resolve) => {
      setTimeout(resolve, 10);
    });

    const firstAbortSignal = abortSignal;

    // Change params to trigger a new request
    rerender({params: {limit: 20}});

    await waitFor(() => {
      expect(firstAbortSignal?.aborted).toBe(true);
    });
  });

  it('should not set error for aborted requests', async () => {
    const abortError = new Error('Aborted');
    abortError.name = 'AbortError';

    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(abortError);

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBeNull();
  });

  it('should abort request on unmount', async () => {
    let abortSignal: AbortSignal | undefined;

    (global.fetch as ReturnType<typeof vi.fn>).mockImplementation((_url, options) => {
      const requestInit = options as RequestInit | undefined;
      abortSignal = requestInit?.signal ?? undefined;
      return new Promise((resolve) => {
        setTimeout(
          () =>
            resolve({
              ok: true,
              json: async () => mockUserSchemaList,
            }),
          100,
        );
      });
    });

    const {unmount} = renderHook(() => useGetUserTypes());

    // Wait a bit for the request to start
    await new Promise((resolve) => {
      setTimeout(resolve, 10);
    });

    unmount();

    await waitFor(() => {
      expect(abortSignal?.aborted).toBe(true);
    });
  });

  it('should refetch when params change', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockUserSchemaList,
    });

    const {rerender} = renderHook(({params}) => useGetUserTypes(params), {
      initialProps: {params: {limit: 10}},
    });

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        'https://localhost:8090/user-schemas?limit=10',
        expect.any(Object),
      );
    });

    rerender({params: {limit: 20}});

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        'https://localhost:8090/user-schemas?limit=20',
        expect.any(Object),
      );
    });

    expect(global.fetch).toHaveBeenCalledTimes(2);
  });
});
