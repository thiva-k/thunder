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
import {renderHook, waitFor, act} from '@testing-library/react';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import type {ReactNode} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/commons-contexts';
import useSetTranslation, {type TranslationResponse, type SetTranslationVariables} from '../useSetTranslation';

vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(),
}));

vi.mock('@thunder/commons-contexts', () => ({
  useConfig: vi.fn(),
}));

// Mock invalidateI18nCache
const mockInvalidateI18nCache = vi.fn();
vi.mock('../../i18n/invalidate-i18n-cache', () => ({
  invalidateI18nCache: (): void => {
    mockInvalidateI18nCache();
  },
}));

describe('useSetTranslation', () => {
  let queryClient: QueryClient;
  let mockHttpRequest: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {retry: false},
        mutations: {retry: false},
      },
    });

    mockHttpRequest = vi.fn();
    mockInvalidateI18nCache.mockReset();

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

  it('should create a new translation', async () => {
    const mockResponse: TranslationResponse = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    await act(async () => {
      result.current.mutate(variables);
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/i18n/languages/en-US/translations/ns/flowI18n/keys/login.title',
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        data: JSON.stringify({value: 'Sign In'}),
      }),
    );
  });

  it('should update an existing translation', async () => {
    const mockResponse: TranslationResponse = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Log In',
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Log In',
    };

    await act(async () => {
      result.current.mutate(variables);
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.value).toBe('Log In');
  });

  it('should invalidate cache on success', async () => {
    const mockResponse: TranslationResponse = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    await act(async () => {
      result.current.mutate(variables);
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    // Verify invalidateI18nCache was called
    expect(mockInvalidateI18nCache).toHaveBeenCalled();
  });

  it('should handle API error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Failed to create translation'));

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    await act(async () => {
      result.current.mutate(variables);
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Failed to create translation');
    expect(result.current.data).toBeUndefined();
  });

  it('should not invalidate cache on error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Failed to create translation'));

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    await act(async () => {
      result.current.mutate(variables);
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    // invalidateI18nCache should not be called on error
    expect(mockInvalidateI18nCache).not.toHaveBeenCalled();
  });

  it('should be in pending state during mutation', async () => {
    mockHttpRequest.mockImplementation(
      () =>
        new Promise((resolve) => {
          setTimeout(
            () =>
              resolve({
                data: {
                  language: 'en-US',
                  namespace: 'flowI18n',
                  key: 'login.title',
                  value: 'Sign In',
                },
              }),
            100,
          );
        }),
    );

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    act(() => {
      result.current.mutate(variables);
    });

    // Check pending state using waitFor since the state update happens asynchronously
    await waitFor(() => {
      expect(result.current.isPending).toBe(true);
    });

    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
  });

  it('should handle different languages', async () => {
    const mockResponse: TranslationResponse = {
      language: 'es',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Iniciar sesión',
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'es',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Iniciar sesión',
    };

    await act(async () => {
      result.current.mutate(variables);
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/i18n/languages/es/translations/ns/flowI18n/keys/login.title',
      }),
    );
  });

  it('should handle different namespaces', async () => {
    const mockResponse: TranslationResponse = {
      language: 'en-US',
      namespace: 'customNs',
      key: 'test.key',
      value: 'Test Value',
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'customNs',
      key: 'test.key',
      value: 'Test Value',
    };

    await act(async () => {
      result.current.mutate(variables);
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/i18n/languages/en-US/translations/ns/customNs/keys/test.key',
      }),
    );
  });

  it('should handle nested keys', async () => {
    const mockResponse: TranslationResponse = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'auth.login.form.submit.button',
      value: 'Submit',
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'auth.login.form.submit.button',
      value: 'Submit',
    };

    await act(async () => {
      result.current.mutate(variables);
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/i18n/languages/en-US/translations/ns/flowI18n/keys/auth.login.form.submit.button',
      }),
    );
  });

  it('should handle special characters in value', async () => {
    const valueWithSpecialChars = 'Sign In with "OAuth" & <SSO>';
    const mockResponse: TranslationResponse = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: valueWithSpecialChars,
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: valueWithSpecialChars,
    };

    await act(async () => {
      result.current.mutate(variables);
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        data: JSON.stringify({value: valueWithSpecialChars}),
      }),
    );
  });

  it('should call onSuccess callback when provided', async () => {
    const mockResponse: TranslationResponse = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const onSuccess = vi.fn();
    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    await act(async () => {
      result.current.mutate(variables, {onSuccess});
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(onSuccess).toHaveBeenCalled();
  });

  it('should call onError callback when provided', async () => {
    const error = new Error('Failed to create translation');
    mockHttpRequest.mockRejectedValue(error);

    const onError = vi.fn();
    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    const variables: SetTranslationVariables = {
      language: 'en-US',
      namespace: 'flowI18n',
      key: 'login.title',
      value: 'Sign In',
    };

    await act(async () => {
      result.current.mutate(variables, {onError});
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(onError).toHaveBeenCalled();
  });

  it('should initialize with idle state', () => {
    const {result} = renderHook(() => useSetTranslation(), {
      wrapper: createWrapper(),
    });

    expect(result.current.data).toBeUndefined();
    expect(result.current.error).toBeNull();
    expect(result.current.isPending).toBe(false);
    expect(result.current.isIdle).toBe(true);
    expect(result.current.isSuccess).toBe(false);
    expect(result.current.isError).toBe(false);
  });
});
