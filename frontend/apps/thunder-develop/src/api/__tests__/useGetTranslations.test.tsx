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
import useGetTranslations, {type TranslationsResponse} from '../useGetTranslations';
import I18nQueryKeys from '../I18nQueryKeys';

vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(),
}));

vi.mock('@thunder/commons-contexts', () => ({
  useConfig: vi.fn(),
}));

describe('useGetTranslations', () => {
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

  it('should fetch translations for a language', async () => {
    const mockResponse: TranslationsResponse = {
      language: 'en-US',
      totalResults: 3,
      translations: {
        flowI18n: {
          'login.title': 'Sign In',
          'login.description': 'Enter your credentials',
          'login.button': 'Submit',
        },
      },
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(result.current.error).toBeNull();
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/i18n/languages/en-US/translations/resolve',
        method: 'GET',
        attachToken: false,
        withCredentials: false,
      }),
    );
  });

  it('should fetch translations with namespace filter', async () => {
    const mockResponse: TranslationsResponse = {
      language: 'en-US',
      translations: {
        flowI18n: {
          'login.title': 'Sign In',
        },
      },
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
          namespace: 'flowI18n',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/i18n/languages/en-US/translations/resolve?namespace=flowI18n',
        method: 'GET',
      }),
    );
  });

  it('should URL encode namespace parameter', async () => {
    const mockResponse: TranslationsResponse = {
      language: 'en-US',
      translations: {},
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
          namespace: 'flow i18n',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/i18n/languages/en-US/translations/resolve?namespace=flow%20i18n',
      }),
    );
  });

  it('should use correct query key without namespace', async () => {
    const mockResponse: TranslationsResponse = {
      language: 'en-US',
      translations: {},
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(I18nQueryKeys.TRANSLATIONS).toBe('i18n-translations');
  });

  it('should be disabled when enabled is false', async () => {
    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
          enabled: false,
        }),
      {
        wrapper: createWrapper(),
      },
    );

    // Wait a bit to ensure no request is made
    await new Promise((resolve) => {
      setTimeout(resolve, 100);
    });

    expect(mockHttpRequest).not.toHaveBeenCalled();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.data).toBeUndefined();
  });

  it('should be disabled when language is empty', async () => {
    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: '',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    // Wait a bit to ensure no request is made
    await new Promise((resolve) => {
      setTimeout(resolve, 100);
    });

    expect(mockHttpRequest).not.toHaveBeenCalled();
    expect(result.current.isLoading).toBe(false);
  });

  it('should handle empty translations', async () => {
    const mockResponse: TranslationsResponse = {
      language: 'en-US',
      translations: {},
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(Object.keys(result.current.data?.translations ?? {})).toHaveLength(0);
  });

  it('should handle multiple namespaces in response', async () => {
    const mockResponse: TranslationsResponse = {
      language: 'en-US',
      translations: {
        flowI18n: {
          'login.title': 'Sign In',
        },
        common: {
          cancel: 'Cancel',
          save: 'Save',
        },
      },
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data?.translations).toHaveProperty('flowI18n');
    expect(result.current.data?.translations).toHaveProperty('common');
    expect(Object.keys(result.current.data?.translations ?? {})).toHaveLength(2);
  });

  it('should handle API error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Failed to fetch translations'));

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Failed to fetch translations');
    expect(result.current.data).toBeUndefined();
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network error'));

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
        }),
      {
        wrapper: createWrapper(),
      },
    );

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
          setTimeout(() => resolve({data: {language: 'en-US', translations: {}}}), 100);
        }),
    );

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    expect(result.current.isLoading).toBe(true);
    expect(result.current.data).toBeUndefined();
  });

  it('should not send auth token with request', async () => {
    const mockResponse: TranslationsResponse = {
      language: 'en-US',
      translations: {},
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'en-US',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        attachToken: false,
        withCredentials: false,
      }),
    );
  });

  it('should fetch translations for different languages', async () => {
    const mockResponseES: TranslationsResponse = {
      language: 'es',
      translations: {
        flowI18n: {
          'login.title': 'Iniciar sesión',
        },
      },
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponseES,
    });

    const {result} = renderHook(
      () =>
        useGetTranslations({
          language: 'es',
        }),
      {
        wrapper: createWrapper(),
      },
    );

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data?.language).toBe('es');
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/i18n/languages/es/translations/resolve',
      }),
    );
  });
});
