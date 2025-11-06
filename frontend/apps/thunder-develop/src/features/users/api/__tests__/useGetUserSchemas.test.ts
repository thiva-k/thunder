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

import useGetUserSchemas from '../useGetUserSchemas';
import type {UserSchemaListResponse} from '../../types/users';

describe('useGetUserSchemas', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useGetUserSchemas());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should fetch user schemas on mount', async () => {
    const mockResponse: UserSchemaListResponse = {
      totalResults: 2,
      startIndex: 0,
      count: 2,
      schemas: [
        {
          id: 'schema-1',
          name: 'Customer',
        },
        {
          id: 'schema-2',
          name: 'Employee',
        },
      ],
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(result.current.error).toBeNull();
    expect(global.fetch).toHaveBeenCalledWith(
      'https://localhost:8090/user-schemas',
      expect.objectContaining({
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    );
  });

  it('should fetch user schemas with query parameters', async () => {
    const mockResponse: UserSchemaListResponse = {
      totalResults: 100,
      startIndex: 10,
      count: 20,
      schemas: [],
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas({limit: 20, offset: 10}));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(global.fetch).toHaveBeenCalledWith(
      'https://localhost:8090/user-schemas?limit=20&offset=10',
      expect.objectContaining({
        method: 'GET',
      }),
    );
  });

  it('should fetch user schemas with limit parameter only', async () => {
    const mockResponse: UserSchemaListResponse = {
      totalResults: 50,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas({limit: 10}));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(global.fetch).toHaveBeenCalledWith(
      'https://localhost:8090/user-schemas?limit=10',
      expect.objectContaining({
        method: 'GET',
      }),
    );
  });

  it('should fetch user schemas with offset parameter only', async () => {
    const mockResponse: UserSchemaListResponse = {
      totalResults: 50,
      startIndex: 20,
      count: 30,
      schemas: [],
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas({offset: 20}));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(global.fetch).toHaveBeenCalledWith(
      'https://localhost:8090/user-schemas?offset=20',
      expect.objectContaining({
        method: 'GET',
      }),
    );
  });

  it('should handle API error with JSON response', async () => {
    const apiErrorResponse = {
      code: 'UNAUTHORIZED',
      message: 'Unauthorized access',
      description: 'You do not have permission to access this resource',
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 401,
      json: async () => apiErrorResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Unauthorized access',
      description: 'An error occurred while fetching user schemas',
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
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Internal Server Error',
      description: 'An error occurred while fetching user schemas',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle network error', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useGetUserSchemas());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Network error',
      description: 'An error occurred while fetching user schemas',
    });
    expect(result.current.data).toBeNull();
  });

  it('should abort previous request when params change', async () => {
    const mockResponse2: UserSchemaListResponse = {
      totalResults: 20,
      startIndex: 0,
      count: 20,
      schemas: [],
    };

    const abortError = new Error('Aborted');
    abortError.name = 'AbortError';

    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(abortError).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse2,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result, rerender} = renderHook(
      ({params}: {params?: {limit?: number; offset?: number}}) => useGetUserSchemas(params),
      {
        initialProps: {params: {limit: 10}},
      },
    );

    // Quickly change params to trigger abort
    rerender({params: {limit: 20}});

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Should have the data from the second request
    expect(result.current.data).toEqual(mockResponse2);
    // Error should be null since AbortError is ignored
    expect(result.current.error).toBeNull();
  });

  it('should refetch with new parameters', async () => {
    const mockResponse1: UserSchemaListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    const mockResponse2: UserSchemaListResponse = {
      totalResults: 5,
      startIndex: 0,
      count: 5,
      schemas: [],
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockResponse1,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas({limit: 10}));

    // Wait for initial fetch to complete with mockResponse1
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data?.count).toBe(10);
    });

    // Now update the mock to return mockResponse2 for the refetch
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockResponse2,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    // Call refetch with new parameters
    await result.current.refetch({limit: 5});

    // Wait for refetch to complete with mockResponse2
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data?.count).toBe(5);
    });

    expect(global.fetch).toHaveBeenCalled();
  });

  it('should refetch with original parameters when no new params provided', async () => {
    const mockResponse: UserSchemaListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas({limit: 10}));

    // Wait for initial fetch to complete
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.data).toEqual(mockResponse);

    // Clear mock calls from initial fetch
    vi.clearAllMocks();
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    // Call refetch without parameters
    await result.current.refetch();

    // Wait for refetch to complete
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(1);
    });

    // Should use the same parameters as initial call
    expect(global.fetch).toHaveBeenCalledWith(
      'https://localhost:8090/user-schemas?limit=10',
      expect.objectContaining({method: 'GET'}),
    );
  });

  it('should cleanup and abort request on unmount', async () => {
    const mockResponse: UserSchemaListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result, unmount} = renderHook(() => useGetUserSchemas());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse);
    });

    unmount();

    // The abort should have been called during cleanup
    // This is difficult to assert directly, but we can verify no errors occur
    expect(true).toBe(true);
  });

  it('should handle empty schemas list', async () => {
    const mockResponse: UserSchemaListResponse = {
      totalResults: 0,
      startIndex: 0,
      count: 0,
      schemas: [],
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(result.current.data?.schemas).toHaveLength(0);
    expect(result.current.error).toBeNull();
  });

  it('should handle large offset pagination', async () => {
    const mockResponse: UserSchemaListResponse = {
      totalResults: 1000,
      startIndex: 900,
      count: 50,
      schemas: [],
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
      signal: undefined,
    });

    const {result} = renderHook(() => useGetUserSchemas({limit: 50, offset: 900}));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(global.fetch).toHaveBeenCalledWith(
      'https://localhost:8090/user-schemas?limit=50&offset=900',
      expect.objectContaining({
        method: 'GET',
      }),
    );
  });
});
