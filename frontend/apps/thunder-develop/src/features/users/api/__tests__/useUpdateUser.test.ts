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

import useUpdateUser, {type UpdateUserRequest} from '../useUpdateUser';
import type {ApiUser} from '../../types/users';

describe('useUpdateUser', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useUpdateUser());

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.updateUser).toBe('function');
    expect(typeof result.current.reset).toBe('function');
  });

  it('should update a user successfully', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useUpdateUser());

    await result.current.updateUser('user-123', mockRequest);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data).toEqual(mockResponse);
      expect(result.current.error).toBeNull();
    });

    expect(global.fetch).toHaveBeenCalledWith('https://localhost:8090/users/user-123', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(mockRequest),
    });
  });

  it('should handle API error with JSON response', async () => {
    const apiErrorResponse = {
      code: 'VALIDATION_ERROR',
      message: 'Validation failed',
      description: 'Email already in use',
    };

    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: async () => apiErrorResponse,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useUpdateUser());

    try {
      await result.current.updateUser('user-123', mockRequest);
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_ERROR',
        message: 'Validation failed',
        description: 'An error occurred while updating the user',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should handle API error without JSON response', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      text: async () => 'Server error occurred',
      headers: new Headers({'content-type': 'text/plain'}),
    });

    const {result} = renderHook(() => useUpdateUser());

    try {
      await result.current.updateUser('user-123', mockRequest);
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_ERROR',
        message: 'Internal Server Error',
        description: 'An error occurred while updating the user',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should handle network error', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useUpdateUser());

    try {
      await result.current.updateUser('user-123', mockRequest);
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_ERROR',
        message: 'Network error',
        description: 'An error occurred while updating the user',
      });
    });
  });

  it('should set loading state correctly during request', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          setTimeout(
            () =>
              resolve({
                ok: true,
                json: async () => mockResponse,
                headers: new Headers({'content-type': 'application/json'}),
              } as Response),
            50,
          );
        }),
    );

    const {result} = renderHook(() => useUpdateUser());

    expect(result.current.loading).toBe(false);

    const promise = result.current.updateUser('user-123', mockRequest);

    // Loading should become true
    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    await promise;

    // Loading should become false after completion
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should reset state correctly', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useUpdateUser());

    await result.current.updateUser('user-123', mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse);
    });

    await waitFor(() => {
      result.current.reset();
    });

    await waitFor(() => {
      expect(result.current.data).toBeNull();
      expect(result.current.error).toBeNull();
    });
  });

  it('should clear previous data when updating', async () => {
    const mockRequest1: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockResponse1: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockRequest2: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'Jane Updated',
        email: 'jane.updated@example.com',
      },
    };

    const mockResponse2: ApiUser = {
      id: 'user-789',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'Jane Updated',
        email: 'jane.updated@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse1,
        headers: new Headers({'content-type': 'application/json'}),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse2,
        headers: new Headers({'content-type': 'application/json'}),
      });

    const {result} = renderHook(() => useUpdateUser());

    await result.current.updateUser('user-123', mockRequest1);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse1);
    });

    await result.current.updateUser('user-789', mockRequest2);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse2);
    });
  });

  it('should handle partial user attribute updates', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        email: 'newemail@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'newemail@example.com',
      },
    };

    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
      headers: new Headers({'content-type': 'application/json'}),
    });

    const {result} = renderHook(() => useUpdateUser());

    await result.current.updateUser('user-123', mockRequest);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse);
    });
    expect(result.current.error).toBeNull();
  });
});
