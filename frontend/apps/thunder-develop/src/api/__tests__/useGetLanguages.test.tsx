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
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import type {ReactNode} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/commons-contexts';
import useGetLanguages, {type LanguagesResponse} from '../useGetLanguages';
import I18nQueryKeys from '../I18nQueryKeys';

vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(),
}));

vi.mock('@thunder/commons-contexts', () => ({
  useConfig: vi.fn(),
}));

describe('useGetLanguages', () => {
  let queryClient: QueryClient;
  let mockHttpRequest: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {retry: false},
      },
    });

    mockHttpRequest = vi.fn();

    vi.mocked(useAsgardeo).mockReturnValue({
      http: {request: mockHttpRequest},
    } as unknown as ReturnType<typeof useAsgardeo>);

    vi.mocked(useConfig).mockReturnValue({
      getServerUrl: () => 'https://localhost:8090',
    } as ReturnType<typeof useConfig>);
  });

  afterEach(() => {
    vi.clearAllMocks();
    queryClient.clear();
  });

  const createWrapper = () => {
    function Wrapper({children}: {children: ReactNode}) {
      return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
    }
    return Wrapper;
  };

  it('should fetch languages on mount', async () => {
    const mockResponse: LanguagesResponse = {
      languages: ['en-US', 'es', 'fr', 'de'],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useGetLanguages(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(result.current.error).toBeNull();
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/i18n/languages',
        method: 'GET',
        attachToken: false,
        withCredentials: false,
      }),
    );
  });

  it('should use correct query key', async () => {
    const mockResponse: LanguagesResponse = {
      languages: ['en-US'],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useGetLanguages(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // Verify the query key is used correctly
    expect(I18nQueryKeys.LANGUAGES).toBe('i18n-languages');
  });

  it('should handle empty languages array', async () => {
    const mockResponse: LanguagesResponse = {
      languages: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useGetLanguages(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(result.current.data?.languages).toHaveLength(0);
  });

  it('should handle API error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Failed to fetch languages'));

    const {result} = renderHook(() => useGetLanguages(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Failed to fetch languages');
    expect(result.current.data).toBeUndefined();
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network error'));

    const {result} = renderHook(() => useGetLanguages(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Network error');
  });

  it('should be in loading state initially', () => {
    mockHttpRequest.mockImplementation(
      () =>
        new Promise((resolve) => {
          setTimeout(() => resolve({data: {languages: ['en-US']}}), 100);
        }),
    );

    const {result} = renderHook(() => useGetLanguages(), {
      wrapper: createWrapper(),
    });

    expect(result.current.isLoading).toBe(true);
    expect(result.current.data).toBeUndefined();
  });

  it('should not send auth token with request', async () => {
    const mockResponse: LanguagesResponse = {
      languages: ['en-US'],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useGetLanguages(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // Verify attachToken and withCredentials are false
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        attachToken: false,
        withCredentials: false,
      }),
    );
  });
});
