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
import useCreateBranding from '../useCreateBranding';
import type {Branding} from '../../models/branding';
import type {CreateBrandingRequest} from '../../models/requests';
import BrandingQueryKeys from '../../constants/branding-query-keys';

// Mock the dependencies
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(),
}));

vi.mock('@thunder/commons-contexts', () => ({
  useConfig: vi.fn(),
}));

describe('useCreateBranding', () => {
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

  const mockRequest: CreateBrandingRequest = {
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
    // Create a fresh QueryClient for each test
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

    // Mock HTTP request function
    mockHttpRequest = vi.fn();

    // Mock Asgardeo SDK
    vi.mocked(useAsgardeo).mockReturnValue({
      http: {
        request: mockHttpRequest,
      },
    } as unknown as ReturnType<typeof useAsgardeo>);

    // Mock useConfig hook
    vi.mocked(useConfig).mockReturnValue({
      getServerUrl: () => 'https://localhost:8090',
    } as ReturnType<typeof useConfig>);
  });

  afterEach(() => {
    vi.clearAllMocks();
    queryClient.clear();
  });

  /**
   * Helper function to create a wrapper with QueryClientProvider
   */
  const createWrapper = () => {
    function Wrapper({children}: {children: ReactNode}) {
      return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
    }
    return Wrapper;
  };

  describe('Initialization', () => {
    it('should initialize with idle state', () => {
      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      expect(result.current.data).toBeUndefined();
      expect(result.current.error).toBeNull();
      expect(result.current.isPending).toBe(false);
      expect(result.current.isIdle).toBe(true);
      expect(result.current.isSuccess).toBe(false);
      expect(result.current.isError).toBe(false);
      expect(typeof result.current.mutate).toBe('function');
      expect(typeof result.current.mutateAsync).toBe('function');
    });

    it('should expose mutation functions', () => {
      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      expect(result.current.mutate).toBeDefined();
      expect(result.current.mutateAsync).toBeDefined();
      expect(typeof result.current.mutate).toBe('function');
      expect(typeof result.current.mutateAsync).toBe('function');
    });
  });

  describe('Successful Creation', () => {
    it('should create branding successfully', async () => {
      mockHttpRequest.mockResolvedValue({data: mockBranding});

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest);
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockBranding);
      expect(result.current.error).toBeNull();
      expect(result.current.isPending).toBe(false);
    });

    it('should call the API with correct parameters', async () => {
      mockHttpRequest.mockResolvedValue({data: mockBranding});

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest);
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith({
        url: 'https://localhost:8090/branding',
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        data: JSON.stringify(mockRequest),
      });
    });

    it('should invalidate brandings query after successful creation', async () => {
      mockHttpRequest.mockResolvedValue({data: mockBranding});

      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest);
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(invalidateQueriesSpy).toHaveBeenCalledWith({
        queryKey: [BrandingQueryKeys.BRANDINGS],
      });
    });

    it('should call onSuccess callback when provided', async () => {
      mockHttpRequest.mockResolvedValue({data: mockBranding});
      const onSuccess = vi.fn();

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest, {onSuccess});
      });

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalledWith(
          mockBranding,
          mockRequest,
          undefined,
          expect.objectContaining({
            client: expect.any(Object),
          }),
        );
      });
    });

    it('should return created branding data with mutateAsync', async () => {
      mockHttpRequest.mockResolvedValue({data: mockBranding});

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      let createdBranding: Branding | undefined;

      await act(async () => {
        createdBranding = await result.current.mutateAsync(mockRequest);
      });

      expect(createdBranding).toEqual(mockBranding);
    });
  });

  describe('Error Handling', () => {
    it('should handle creation errors', async () => {
      const mockError = new Error('Failed to create branding');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest);
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
      expect(result.current.data).toBeUndefined();
      expect(result.current.isPending).toBe(false);
    });

    it('should call onError callback when provided', async () => {
      const mockError = new Error('Failed to create branding');
      mockHttpRequest.mockRejectedValue(mockError);
      const onError = vi.fn();

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest, {onError});
      });

      await waitFor(() => {
        expect(onError).toHaveBeenCalledWith(
          mockError,
          mockRequest,
          undefined,
          expect.objectContaining({
            client: expect.any(Object),
          }),
        );
      });
    });

    it('should not invalidate queries on error', async () => {
      const mockError = new Error('Failed to create branding');
      mockHttpRequest.mockRejectedValue(mockError);

      const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest);
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(invalidateQueriesSpy).not.toHaveBeenCalled();
    });

    it('should handle network errors', async () => {
      const networkError = new Error('Network error');
      mockHttpRequest.mockRejectedValue(networkError);

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest);
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error?.message).toBe('Network error');
    });

    it('should reject promise with mutateAsync on error', async () => {
      const mockError = new Error('Failed to create branding');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      await expect(
        act(async () => {
          await result.current.mutateAsync(mockRequest);
        }),
      ).rejects.toThrow('Failed to create branding');
    });
  });

  describe('Loading State', () => {
    it('should show pending state during creation', async () => {
      mockHttpRequest.mockImplementation(
        () =>
          new Promise((resolve) => {
            setTimeout(() => resolve({data: mockBranding}), 100);
          }),
      );

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest);
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
    it('should handle multiple sequential mutations', async () => {
      const mockBranding2: Branding = {
        ...mockBranding,
        id: 'branding-456',
        displayName: 'Second Branding',
      };

      mockHttpRequest.mockResolvedValueOnce({data: mockBranding}).mockResolvedValueOnce({data: mockBranding2});

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      // First mutation
      act(() => {
        result.current.mutate(mockRequest);
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
      expect(result.current.data).toEqual(mockBranding);

      // Second mutation
      const mockRequest2 = {...mockRequest, displayName: 'Second Branding'};
      act(() => {
        result.current.mutate(mockRequest2);
      });

      await waitFor(() => {
        expect(result.current.data).toEqual(mockBranding2);
      });
    });

    it('should reset mutation state with reset function', async () => {
      mockHttpRequest.mockResolvedValue({data: mockBranding});

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(mockRequest);
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
    it('should handle empty branding name', async () => {
      const emptyNameRequest = {...mockRequest, displayName: ''};
      mockHttpRequest.mockResolvedValue({
        data: {...mockBranding, displayName: ''},
      });

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(emptyNameRequest);
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          data: JSON.stringify(emptyNameRequest),
        }),
      );
    });

    it('should handle minimal branding preferences', async () => {
      const minimalRequest: CreateBrandingRequest = {
        displayName: 'Minimal Branding',
        preferences: {
          theme: {
            activeColorScheme: 'light',
            colorSchemes: {
              light: {
                colors: {
                  primary: {
                    main: '#000000',
                    dark: '#000000',
                    contrastText: '#ffffff',
                  },
                  secondary: {
                    main: '#000000',
                    dark: '#000000',
                    contrastText: '#ffffff',
                  },
                },
              },
            },
          },
        },
      };

      mockHttpRequest.mockResolvedValue({
        data: {...mockBranding, displayName: 'Minimal Branding'},
      });

      const {result} = renderHook(() => useCreateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(minimalRequest);
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          data: JSON.stringify(minimalRequest),
        }),
      );
    });
  });
});
