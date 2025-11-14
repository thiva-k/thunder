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

/* eslint-disable @typescript-eslint/no-unsafe-assignment */

import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest';
import {renderHook, waitFor, act} from '@testing-library/react';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import type {ReactNode} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/commons-contexts';
import useDeleteBranding from '../useDeleteBranding';
import BrandingQueryKeys from '../../constants/branding-query-keys';

// Mock the dependencies
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(),
}));

vi.mock('@thunder/commons-contexts', () => ({
  useConfig: vi.fn(),
}));

describe('useDeleteBranding', () => {
  let queryClient: QueryClient;
  let mockHttpRequest: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
        mutations: {
          retry: false,
        },
      },
    });

    mockHttpRequest = vi.fn().mockResolvedValue({data: undefined});

    vi.mocked(useAsgardeo).mockReturnValue({
      http: {
        request: mockHttpRequest,
      },
    } as unknown as ReturnType<typeof useAsgardeo>);

    vi.mocked(useConfig).mockReturnValue({
      getServerUrl: () => 'http://localhost:8090',
    } as ReturnType<typeof useConfig>);
  });

  afterEach(() => {
    queryClient.clear();
    vi.clearAllMocks();
  });

  const createWrapper = () => {
    function Wrapper({children}: {children: ReactNode}) {
      return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
    }
    return Wrapper;
  };

  describe('Initialization', () => {
    it('should initialize in idle state', () => {
      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isIdle).toBe(true);
      expect(result.current.isPending).toBe(false);
      expect(result.current.isSuccess).toBe(false);
      expect(result.current.isError).toBe(false);
      expect(result.current.data).toBeUndefined();
      expect(result.current.error).toBeNull();
    });
  });

  describe('Successful Deletion', () => {
    it('should delete branding successfully', async () => {
      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.error).toBeNull();
    });

    it('should call API with correct parameters', async () => {
      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'http://localhost:8090/branding/branding-123',
          method: 'DELETE',
        }),
      );
    });

    it('should invalidate queries after successful deletion', async () => {
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(invalidateSpy).toHaveBeenCalledWith({
        queryKey: [BrandingQueryKeys.BRANDING, 'branding-123'],
      });
      expect(invalidateSpy).toHaveBeenCalledWith({
        queryKey: [BrandingQueryKeys.BRANDINGS],
      });
    });

    it('should call onSuccess callback when provided', async () => {
      const onSuccess = vi.fn();

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123', {onSuccess});
      });

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalledWith(
          undefined,
          'branding-123',
          undefined,
          expect.objectContaining({
            client: expect.any(Object),
          }),
        );
      });
    });

    it('should complete successfully with mutateAsync', async () => {
      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      await act(async () => {
        await result.current.mutateAsync('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle deletion errors', async () => {
      const mockError = new Error('Failed to delete branding');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
      expect(result.current.data).toBeUndefined();
    });

    it('should handle 404 not found errors', async () => {
      const notFoundError = new Error('Branding not found');
      mockHttpRequest.mockRejectedValue(notFoundError);

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('non-existent-id');
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error?.message).toBe('Branding not found');
    });

    it('should maintain error state after failed deletion', async () => {
      const mockError = new Error('Network error');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error?.message).toBe('Network error');
      expect(result.current.isSuccess).toBe(false);
    });

    it('should call onError callback when provided', async () => {
      const mockError = new Error('Failed to delete branding');
      mockHttpRequest.mockRejectedValue(mockError);
      const onError = vi.fn();

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123', {onError});
      });

      await waitFor(() => {
        expect(onError).toHaveBeenCalledWith(
          mockError,
          'branding-123',
          undefined,
          expect.objectContaining({
            client: expect.any(Object),
          }),
        );
      });
    });

    it('should not invalidate queries on error', async () => {
      const mockError = new Error('Failed to delete branding');
      mockHttpRequest.mockRejectedValue(mockError);
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(invalidateSpy).not.toHaveBeenCalled();
    });

    it('should throw error with mutateAsync', async () => {
      const mockError = new Error('Failed to delete branding');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      await expect(
        act(async () => {
          await result.current.mutateAsync('branding-123');
        }),
      ).rejects.toThrow('Failed to delete branding');
    });
  });

  describe('Loading State', () => {
    it('should show pending state during deletion', async () => {
      mockHttpRequest.mockImplementation(
        () =>
          new Promise((resolve) => {
            setTimeout(() => resolve({data: undefined}), 100);
          }),
      );

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isPending).toBe(true);
      });

      expect(result.current.isIdle).toBe(false);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isPending).toBe(false);
    });
  });

  describe('Multiple Mutations', () => {
    it('should handle multiple sequential deletions', async () => {
      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      // First deletion
      act(() => {
        result.current.mutate('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Second deletion
      act(() => {
        result.current.mutate('branding-456');
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledTimes(2);
      expect(mockHttpRequest).toHaveBeenNthCalledWith(
        1,
        expect.objectContaining({
          url: 'http://localhost:8090/branding/branding-123',
        }),
      );
      expect(mockHttpRequest).toHaveBeenNthCalledWith(
        2,
        expect.objectContaining({
          url: 'http://localhost:8090/branding/branding-456',
        }),
      );
    });

    it('should reset mutation state with reset function', async () => {
      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate('branding-123');
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      act(() => {
        result.current.reset();
      });

      await waitFor(() => {
        expect(result.current.isIdle).toBe(true);
      });

      expect(result.current.data).toBeUndefined();
      expect(result.current.error).toBeNull();
      expect(result.current.isSuccess).toBe(false);
    });
  });

  describe('Edge Cases', () => {
    it('should handle deletion with special characters in ID', async () => {
      const specialId = 'branding-123-test@domain.com';

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(specialId);
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: `http://localhost:8090/branding/${specialId}`,
        }),
      );
    });

    it('should handle deletion with UUID format ID', async () => {
      const uuidId = '3fa85f64-5717-4562-b3fc-2c963f66afa6';

      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(uuidId);
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: `http://localhost:8090/branding/${uuidId}`,
        }),
      );
    });
  });

  describe('Concurrent Deletions', () => {
    it('should handle concurrent deletion attempts', async () => {
      const {result} = renderHook(() => useDeleteBranding(), {
        wrapper: createWrapper(),
      });

      // Trigger multiple deletions simultaneously
      act(() => {
        result.current.mutate('branding-123');
        result.current.mutate('branding-456');
        result.current.mutate('branding-789');
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      // Should have called the API for each deletion
      expect(mockHttpRequest).toHaveBeenCalledTimes(3);
    });
  });
});
