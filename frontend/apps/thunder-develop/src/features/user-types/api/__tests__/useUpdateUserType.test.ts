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

import useUpdateUserType from '../useUpdateUserType';
import type {ApiUserSchema, UpdateUserSchemaRequest} from '../../types/user-types';

describe('useUpdateUserType', () => {
  const mockUserTypeId = '123';
  const mockUserSchema: ApiUserSchema = {
    id: mockUserTypeId,
    name: 'UpdatedUserType',
    schema: {
      username: {
        type: 'string',
        required: true,
      },
      email: {
        type: 'string',
        required: true,
      },
    },
  };

  const mockRequest: UpdateUserSchemaRequest = {
    name: 'UpdatedUserType',
    schema: {
      username: {
        type: 'string',
        required: true,
      },
      email: {
        type: 'string',
        required: true,
      },
    },
  };

  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with default state', () => {
    const {result} = renderHook(() => useUpdateUserType());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(typeof result.current.updateUserType).toBe('function');
    expect(typeof result.current.reset).toBe('function');
  });

  it('should successfully update a user type', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result} = renderHook(() => useUpdateUserType());

    await result.current.updateUserType(mockUserTypeId, mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
      expect(result.current.error).toBeNull();
      expect(result.current.loading).toBe(false);
    });

    expect(global.fetch).toHaveBeenCalledWith(`https://localhost:8090/user-schemas/${mockUserTypeId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(mockRequest),
    });
  });

  it('should set loading state during update', async () => {
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

    const {result} = renderHook(() => useUpdateUserType());

    const promise = result.current.updateUserType(mockUserTypeId, mockRequest);

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
      code: 'VALIDATION_ERROR',
      message: 'Validation failed',
      description: 'User type name already exists',
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: async () => apiErrorResponse,
    });

    const {result} = renderHook(() => useUpdateUserType());

    await expect(result.current.updateUserType(mockUserTypeId, mockRequest)).rejects.toThrow('Validation failed');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_TYPE_ERROR',
        message: 'Validation failed',
        description: 'Failed to update user type',
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

    const {result} = renderHook(() => useUpdateUserType());

    await expect(result.current.updateUserType(mockUserTypeId, mockRequest)).rejects.toThrow('HTTP error! status: 500');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_TYPE_ERROR',
        message: 'HTTP error! status: 500',
        description: 'Failed to update user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle network error', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useUpdateUserType());

    await expect(result.current.updateUserType(mockUserTypeId, mockRequest)).rejects.toThrow('Network error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_TYPE_ERROR',
        message: 'Network error',
        description: 'Failed to update user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle user type not found error', async () => {
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

    const {result} = renderHook(() => useUpdateUserType());

    await expect(result.current.updateUserType(mockUserTypeId, mockRequest)).rejects.toThrow('User type not found');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_TYPE_ERROR',
        message: 'User type not found',
        description: 'Failed to update user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should reset state when reset is called', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result} = renderHook(() => useUpdateUserType());

    await result.current.updateUserType(mockUserTypeId, mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    await waitFor(() => {
      result.current.reset();
    });

    await waitFor(() => {
      expect(result.current.data).toBeNull();
      expect(result.current.error).toBeNull();
    });
  });

  it('should clear previous data and error when starting new request', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result} = renderHook(() => useUpdateUserType());

    await result.current.updateUserType(mockUserTypeId, mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    const updatedSchema = {...mockUserSchema, name: 'AnotherUpdate'};

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => updatedSchema,
    });

    await result.current.updateUserType(mockUserTypeId, mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(updatedSchema);
    });
  });

  it('should update different user type with different id', async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockUserSchema,
    });

    const {result} = renderHook(() => useUpdateUserType());

    await result.current.updateUserType('123', mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    const anotherUserSchema = {...mockUserSchema, id: '456', name: 'AnotherUserType'};

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => anotherUserSchema,
    });

    await result.current.updateUserType('456', mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(anotherUserSchema);
    });

    expect(global.fetch).toHaveBeenNthCalledWith(
      1,
      'https://localhost:8090/user-schemas/123',
      expect.any(Object),
    );
    expect(global.fetch).toHaveBeenNthCalledWith(
      2,
      'https://localhost:8090/user-schemas/456',
      expect.any(Object),
    );
  });
});
