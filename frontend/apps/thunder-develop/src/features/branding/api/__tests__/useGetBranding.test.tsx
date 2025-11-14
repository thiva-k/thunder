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
import useGetBranding from '../useGetBranding';
import type {Branding} from '../../models/branding';

// Mock the dependencies
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(),
}));

vi.mock('@thunder/commons-contexts', () => ({
  useConfig: vi.fn(),
}));

describe('useGetBranding', () => {
  const mockBranding: Branding = {
    id: 'branding-123',
    displayName: 'Test Branding',
    preferences: {
      theme: {
        activeColorScheme: 'light',
        colorSchemes: {
          light: {
            colors: {
              primary: {
                main: '#1976d2',
                dark: '#115293',
                contrastText: '#ffffff',
              },
              secondary: {
                main: '#dc004e',
                dark: '#9a0036',
                contrastText: '#ffffff',
              },
            },
          },
        },
      },
    },
  };

  let queryClient: QueryClient;
  let mockHttpRequest: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          gcTime: Infinity,
          staleTime: Infinity,
        },
      },
    });

    mockHttpRequest = vi.fn().mockResolvedValue({data: mockBranding});

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
    it('should initialize with loading state when brandingId is provided', () => {
      const {result} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
      expect(result.current.error).toBeNull();
    });

    it('should not fetch when brandingId is empty', () => {
      const {result} = renderHook(() => useGetBranding(''), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.data).toBeUndefined();
      expect(mockHttpRequest).not.toHaveBeenCalled();
    });
  });

  describe('Successful Fetch', () => {
    it('should fetch branding successfully', async () => {
      const {result} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockBranding);
      expect(result.current.error).toBeNull();
    });

    it('should call API with correct URL', async () => {
      const {result} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'http://localhost:8090/branding/branding-123',
          method: 'GET',
        }),
      );
    });

    it('should return correct branding structure', async () => {
      const {result} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toHaveProperty('id');
      expect(result.current.data).toHaveProperty('displayName');
      expect(result.current.data).toHaveProperty('preferences');
      expect(result.current.data?.id).toBe('branding-123');
    });

    it('should handle branding with minimal preferences', async () => {
      const minimalBranding: Branding = {
        id: 'branding-456',
        displayName: 'Minimal Branding',
        preferences: {},
      };

      mockHttpRequest.mockResolvedValue({data: minimalBranding});

      const {result} = renderHook(() => useGetBranding('branding-456'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(minimalBranding);
    });
  });

  describe('Error Handling', () => {
    it('should handle network errors', async () => {
      const mockError = new Error('Network error');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
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

      const {result} = renderHook(() => useGetBranding('non-existent-id'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error?.message).toBe('Branding not found');
    });

    it('should handle API error responses', async () => {
      const apiError = new Error('Failed to fetch branding');
      mockHttpRequest.mockRejectedValue(apiError);

      const {result} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error?.message).toBe('Failed to fetch branding');
    });
  });

  describe('Loading State', () => {
    it('should show loading state during fetch', async () => {
      mockHttpRequest.mockImplementation(
        () =>
          new Promise((resolve) => {
            setTimeout(() => resolve({data: mockBranding}), 100);
          }),
      );

      const {result} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.isLoading).toBe(false);
    });
  });

  describe('Refetching', () => {
    it('should allow manual refetch', async () => {
      const {result} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledTimes(1);

      await result.current.refetch();

      expect(mockHttpRequest).toHaveBeenCalledTimes(2);
    });

    it('should fetch fresh data on refetch', async () => {
      const {result} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.displayName).toBe('Test Branding');

      const updatedBranding: Branding = {
        ...mockBranding,
        displayName: 'Updated Branding',
      };

      mockHttpRequest.mockResolvedValue({data: updatedBranding});

      await act(async () => {
        await result.current.refetch();
      });

      await waitFor(() => {
        expect(result.current.data?.displayName).toBe('Updated Branding');
      });

      expect(mockHttpRequest).toHaveBeenCalledTimes(2);
    });
  });

  describe('Enabled State', () => {
    it('should not fetch when brandingId is falsy', () => {
      const {result} = renderHook(() => useGetBranding(''), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(false);
      expect(result.current.fetchStatus).toBe('idle');
      expect(mockHttpRequest).not.toHaveBeenCalled();
    });

    it('should fetch when brandingId becomes truthy', async () => {
      const {result, rerender} = renderHook(({id}: {id: string}) => useGetBranding(id), {
        wrapper: createWrapper(),
        initialProps: {id: ''},
      });

      expect(mockHttpRequest).not.toHaveBeenCalled();

      rerender({id: 'branding-123'});

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledTimes(1);
    });
  });

  describe('Caching', () => {
    it('should cache results for same brandingId', async () => {
      const wrapper = createWrapper();

      const {result: result1, unmount: unmount1} = renderHook(() => useGetBranding('branding-123'), {
        wrapper,
      });

      await waitFor(() => {
        expect(result1.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledTimes(1);

      // Unmount first hook
      unmount1();

      const {result: result2} = renderHook(() => useGetBranding('branding-123'), {
        wrapper,
      });

      await waitFor(() => {
        expect(result2.current.isSuccess).toBe(true);
      });

      // Should still be 1 API call, second hook uses cache
      expect(mockHttpRequest).toHaveBeenCalledTimes(1);
    });

    it('should make separate requests for different brandingIds', async () => {
      const {result: result1} = renderHook(() => useGetBranding('branding-123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result1.current.isSuccess).toBe(true);
      });

      const {result: result2} = renderHook(() => useGetBranding('branding-456'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result2.current.isSuccess).toBe(true);
      });

      // Should make 2 API calls for different IDs
      expect(mockHttpRequest).toHaveBeenCalledTimes(2);
    });
  });

  describe('Dynamic BrandingId Changes', () => {
    it('should fetch new data when brandingId changes', async () => {
      const {result, rerender} = renderHook(({id}: {id: string}) => useGetBranding(id), {
        wrapper: createWrapper(),
        initialProps: {id: 'branding-123'},
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.id).toBe('branding-123');

      const newBranding: Branding = {
        ...mockBranding,
        id: 'branding-456',
        displayName: 'New Branding',
      };

      mockHttpRequest.mockResolvedValue({data: newBranding});

      rerender({id: 'branding-456'});

      await waitFor(() => {
        expect(result.current.data?.id).toBe('branding-456');
      });

      expect(mockHttpRequest).toHaveBeenCalledTimes(2);
    });
  });
});
