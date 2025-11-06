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

import useGetUserSchema from '../useGetUserSchema';
import type {ApiUserSchema} from '../../types/users';

describe('useGetUserSchema', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useGetUserSchema());

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should fetch user schema successfully when id is provided', async () => {
    const mockSchema: ApiUserSchema = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
        email: {
          type: 'string',
          required: true,
        },
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockSchema,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useGetUserSchema('schema-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockSchema);
    expect(result.current.error).toBeNull();
    expect(global.fetch).toHaveBeenCalledWith('https://localhost:8090/user-schemas/schema-123', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
  });

  it('should not fetch when id is not provided', () => {
    const {result} = renderHook(() => useGetUserSchema());

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(global.fetch).not.toHaveBeenCalled();
  });

  it('should handle API error with JSON response', async () => {
    const apiErrorResponse = {
      code: 'NOT_FOUND',
      message: 'Schema not found',
      description: 'The schema with the given ID does not exist',
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: async () => apiErrorResponse,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useGetUserSchema('schema-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Schema not found',
      description: 'An error occurred while fetching user schema',
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

    const {result} = renderHook(() => useGetUserSchema('schema-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Internal Server Error',
      description: 'An error occurred while fetching user schema',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle network error', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useGetUserSchema('schema-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Network error',
      description: 'An error occurred while fetching user schema',
    });
    expect(result.current.data).toBeNull();
  });

  it('should refetch user schema with the same id', async () => {
    const mockSchema: ApiUserSchema = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockSchema,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useGetUserSchema('schema-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema);
    });

    result.current.refetch();

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(2);
    });
  });

  it('should refetch user schema with a new id', async () => {
    const mockSchema1: ApiUserSchema = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    const mockSchema2: ApiUserSchema = {
      id: 'schema-456',
      name: 'Employee',
      schema: {
        employeeId: {
          type: 'string',
          required: true,
        },
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockSchema1,
        headers: new Headers({'content-type': 'application/json'}),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockSchema2,
        headers: new Headers({'content-type': 'application/json'}),
      });

    const {result} = renderHook(() => useGetUserSchema('schema-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema1);
    });

    result.current.refetch('schema-456');

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema2);
    });
  });

  it('should set error when refetch is called without id', async () => {
    const {result} = renderHook(() => useGetUserSchema());

    result.current.refetch();

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'INVALID_ID',
        message: 'Invalid schema ID',
        description: 'Schema ID is required',
      });
    });
  });

  it('should prevent double fetch in strict mode', async () => {
    const mockSchema: ApiUserSchema = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockSchema,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useGetUserSchema('schema-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema);
    });

    // Should only fetch once despite strict mode
    expect(global.fetch).toHaveBeenCalledTimes(1);
  });

  it('should fetch when id changes', async () => {
    const mockSchema1: ApiUserSchema = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    const mockSchema2: ApiUserSchema = {
      id: 'schema-456',
      name: 'Employee',
      schema: {
        employeeId: {
          type: 'string',
          required: true,
        },
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockSchema1,
        headers: new Headers({'content-type': 'application/json'}),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockSchema2,
        headers: new Headers({'content-type': 'application/json'}),
      });

    const {result, rerender} = renderHook(({id}: {id?: string}) => useGetUserSchema(id), {
      initialProps: {id: 'schema-123'},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema1);
    });

    rerender({id: 'schema-456'});

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema2);
    });

    expect(global.fetch).toHaveBeenCalledTimes(2);
  });

  it('should handle schema with complex properties', async () => {
    const mockSchema: ApiUserSchema = {
      id: 'schema-789',
      name: 'ComplexUser',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
        age: {
          type: 'number',
          required: false,
        },
        isActive: {
          type: 'boolean',
          required: true,
        },
        roles: {
          type: 'array',
          items: {
            type: 'string',
          },
          required: false,
        },
        address: {
          type: 'object',
          properties: {
            street: {
              type: 'string',
              required: true,
            },
            city: {
              type: 'string',
              required: true,
            },
          },
          required: false,
        },
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockSchema,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useGetUserSchema('schema-789'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockSchema);
    expect(result.current.error).toBeNull();
  });
});
