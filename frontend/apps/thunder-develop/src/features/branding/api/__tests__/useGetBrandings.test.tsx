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
import {renderHook, waitFor} from '@testing-library/react';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import type {ReactNode} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/commons-contexts';
import useGetBrandings from '../useGetBrandings';
import type {BrandingListResponse} from '../../models/responses';

// Mock the dependencies
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(),
}));

vi.mock('@thunder/commons-contexts', () => ({
  useConfig: vi.fn(),
}));

describe('useGetBrandings', () => {
  const mockBrandingListResponse: BrandingListResponse = {
    totalResults: 3,
    startIndex: 1,
    count: 3,
    brandings: [
      {
        id: 'branding-1',
        displayName: 'Branding 1',
      },
      {
        id: 'branding-2',
        displayName: 'Branding 2',
      },
      {
        id: 'branding-3',
        displayName: 'Branding 3',
      },
    ],
    links: [
      {
        href: 'branding?offset=3&limit=3',
        rel: 'next',
      },
    ],
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

    mockHttpRequest = vi.fn().mockResolvedValue({data: mockBrandingListResponse});

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
    it('should initialize with loading state', () => {
      const {result} = renderHook(() => useGetBrandings(), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);
      expect(result.current.data).toBeUndefined();
      expect(result.current.error).toBeNull();
    });
  });

  describe('Successful Fetch', () => {
    it('should fetch brandings successfully', async () => {
      const {result} = renderHook(() => useGetBrandings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockBrandingListResponse);
      expect(result.current.error).toBeNull();
    });

    it('should call API with default pagination parameters', async () => {
      const {result} = renderHook(() => useGetBrandings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'http://localhost:8090/branding?limit=30&offset=0',
          method: 'GET',
        }),
      );
    });

    it('should call API with custom pagination parameters', async () => {
      const {result} = renderHook(() => useGetBrandings({limit: 10, offset: 20}), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'http://localhost:8090/branding?limit=10&offset=20',
          method: 'GET',
        }),
      );
    });

    it('should return correct branding list structure', async () => {
      const {result} = renderHook(() => useGetBrandings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toHaveProperty('totalResults');
      expect(result.current.data).toHaveProperty('startIndex');
      expect(result.current.data).toHaveProperty('count');
      expect(result.current.data).toHaveProperty('brandings');
      expect(result.current.data?.brandings).toHaveLength(3);
    });

    it('should handle empty branding list', async () => {
      const emptyResponse: BrandingListResponse = {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        brandings: [],
      };

      mockHttpRequest.mockResolvedValue({data: emptyResponse});

      const {result} = renderHook(() => useGetBrandings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.brandings).toHaveLength(0);
      expect(result.current.data?.totalResults).toBe(0);
    });
  });

  describe('Error Handling', () => {
    it('should handle network errors', async () => {
      const mockError = new Error('Network error');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useGetBrandings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
      expect(result.current.data).toBeUndefined();
    });

    it('should handle API error responses', async () => {
      const apiError = new Error('Failed to fetch brandings');
      mockHttpRequest.mockRejectedValue(apiError);

      const {result} = renderHook(() => useGetBrandings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error?.message).toBe('Failed to fetch brandings');
    });
  });

  describe('Loading State', () => {
    it('should show loading state during fetch', async () => {
      mockHttpRequest.mockImplementation(
        () =>
          new Promise((resolve) => {
            setTimeout(() => resolve({data: mockBrandingListResponse}), 100);
          }),
      );

      const {result} = renderHook(() => useGetBrandings(), {
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
      const {result} = renderHook(() => useGetBrandings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledTimes(1);

      await result.current.refetch();

      expect(mockHttpRequest).toHaveBeenCalledTimes(2);
    });
  });

  describe('Pagination', () => {
    it('should handle different limit values', async () => {
      const {result: result1} = renderHook(() => useGetBrandings({limit: 5}), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result1.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'http://localhost:8090/branding?limit=5&offset=0',
        }),
      );
    });

    it('should handle different offset values', async () => {
      const {result} = renderHook(() => useGetBrandings({offset: 10}), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'http://localhost:8090/branding?limit=30&offset=10',
        }),
      );
    });

    it('should handle pagination links in response', async () => {
      const {result} = renderHook(() => useGetBrandings(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.links).toBeDefined();
      expect(result.current.data?.links).toHaveLength(1);
      expect(result.current.data?.links?.[0].rel).toBe('next');
    });
  });

  describe('Caching', () => {
    it('should cache results with different pagination parameters', async () => {
      const {result: result1} = renderHook(() => useGetBrandings({limit: 10}), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result1.current.isSuccess).toBe(true);
      });

      const {result: result2} = renderHook(() => useGetBrandings({limit: 20}), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result2.current.isSuccess).toBe(true);
      });

      // Should have made 2 API calls for different cache keys
      expect(mockHttpRequest).toHaveBeenCalledTimes(2);
    });

    it('should use cached results for same pagination parameters', async () => {
      const wrapper = createWrapper();

      const {result: result1, unmount: unmount1} = renderHook(() => useGetBrandings({limit: 10}), {
        wrapper,
      });

      await waitFor(() => {
        expect(result1.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledTimes(1);

      // Unmount first hook
      unmount1();

      const {result: result2} = renderHook(() => useGetBrandings({limit: 10}), {
        wrapper,
      });

      await waitFor(() => {
        expect(result2.current.isSuccess).toBe(true);
      });

      // Should still be 1 API call, second hook uses cache
      expect(mockHttpRequest).toHaveBeenCalledTimes(1);
    });
  });
});
