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
import useUpdateBranding from '../useUpdateBranding';
import type {Branding} from '../../models/branding';
import type {UpdateBrandingRequest} from '../../models/requests';
import BrandingQueryKeys from '../../constants/branding-query-keys';

// Mock the dependencies
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(),
}));

vi.mock('@thunder/commons-contexts', () => ({
  useConfig: vi.fn(),
}));

describe('useUpdateBranding', () => {
  const mockBranding: Branding = {
    id: 'branding-123',
    displayName: 'Updated Branding',
    preferences: {
      theme: {
        activeColorScheme: 'light',
        colorSchemes: {
          light: {
            colors: {
              primary: {
                main: '#ff0000',
                dark: '#cc0000',
                contrastText: '#ffffff',
              },
              secondary: {
                main: '#00ff00',
                dark: '#00cc00',
                contrastText: '#000000',
              },
            },
          },
        },
      },
    },
  };

  const mockUpdateRequest: UpdateBrandingRequest = {
    displayName: 'Updated Branding',
    preferences: {
      theme: {
        activeColorScheme: 'light',
        colorSchemes: {
          light: {
            colors: {
              primary: {
                main: '#ff0000',
                dark: '#cc0000',
                contrastText: '#ffffff',
              },
              secondary: {
                main: '#00ff00',
                dark: '#00cc00',
                contrastText: '#000000',
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
        },
        mutations: {
          retry: false,
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
    it('should initialize in idle state', () => {
      const {result} = renderHook(() => useUpdateBranding(), {
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

  describe('Successful Update', () => {
    it('should update branding successfully', async () => {
      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockBranding);
      expect(result.current.error).toBeNull();
    });

    it('should call API with correct parameters', async () => {
      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'http://localhost:8090/branding/branding-123',
          method: 'PUT',
          data: JSON.stringify(mockUpdateRequest),
        }),
      );
    });

    it('should invalidate queries after successful update', async () => {
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
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

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(
          {
            brandingId: 'branding-123',
            data: mockUpdateRequest,
          },
          {onSuccess},
        );
      });

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalledWith(
          mockBranding,
          {brandingId: 'branding-123', data: mockUpdateRequest},
          undefined,
          expect.objectContaining({
            client: expect.any(Object),
          }),
        );
      });
    });

    it('should return updated branding data with mutateAsync', async () => {
      mockHttpRequest.mockResolvedValue({data: mockBranding});

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      let updatedData: Branding | undefined;

      await act(async () => {
        updatedData = await result.current.mutateAsync({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
      });

      expect(updatedData).toEqual(mockBranding);
    });
  });

  describe('Error Handling', () => {
    it('should handle update errors', async () => {
      const mockError = new Error('Failed to update branding');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
      expect(result.current.data).toBeUndefined();
    });

    it('should maintain error state after failed update', async () => {
      const mockError = new Error('Network error');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error?.message).toBe('Network error');
      expect(result.current.isSuccess).toBe(false);
    });

    it('should call onError callback when provided', async () => {
      const mockError = new Error('Failed to update branding');
      mockHttpRequest.mockRejectedValue(mockError);
      const onError = vi.fn();

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate(
          {
            brandingId: 'branding-123',
            data: mockUpdateRequest,
          },
          {onError},
        );
      });

      await waitFor(() => {
        expect(onError).toHaveBeenCalledWith(
          mockError,
          {brandingId: 'branding-123', data: mockUpdateRequest},
          undefined,
          expect.objectContaining({
            client: expect.any(Object),
          }),
        );
      });
    });

    it('should not invalidate queries on error', async () => {
      const mockError = new Error('Failed to update branding');
      mockHttpRequest.mockRejectedValue(mockError);
      const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(invalidateSpy).not.toHaveBeenCalled();
    });

    it('should throw error with mutateAsync', async () => {
      const mockError = new Error('Failed to update branding');
      mockHttpRequest.mockRejectedValue(mockError);

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      await expect(
        act(async () => {
          await result.current.mutateAsync({
            brandingId: 'branding-123',
            data: mockUpdateRequest,
          });
        }),
      ).rejects.toThrow('Failed to update branding');
    });
  });

  describe('Loading State', () => {
    it('should show pending state during update', async () => {
      mockHttpRequest.mockImplementation(
        () =>
          new Promise((resolve) => {
            setTimeout(() => resolve({data: mockBranding}), 100);
          }),
      );

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
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
    it('should handle multiple sequential updates', async () => {
      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      // First update
      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockBranding);

      // Second update
      const updatedBranding: Branding = {
        ...mockBranding,
        displayName: 'Second Update',
      };
      mockHttpRequest.mockResolvedValueOnce({data: updatedBranding});

      const secondRequest: UpdateBrandingRequest = {
        ...mockUpdateRequest,
        displayName: 'Second Update',
      };

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: secondRequest,
        });
      });

      await waitFor(() => {
        expect(result.current.data?.displayName).toBe('Second Update');
      });
    });

    it('should reset mutation state with reset function', async () => {
      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: mockUpdateRequest,
        });
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
    it('should handle partial updates', async () => {
      const partialRequest: UpdateBrandingRequest = {
        displayName: 'Only Name Update',
        preferences: {},
      };

      mockHttpRequest.mockResolvedValue({
        data: {...mockBranding, displayName: 'Only Name Update'},
      });

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: partialRequest,
        });
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          data: JSON.stringify(partialRequest),
        }),
      );
    });

    it('should handle updates with empty display name', async () => {
      const emptyNameRequest: UpdateBrandingRequest = {
        displayName: '',
        preferences: mockUpdateRequest.preferences,
      };

      mockHttpRequest.mockResolvedValue({
        data: {...mockBranding, displayName: ''},
      });

      const {result} = renderHook(() => useUpdateBranding(), {
        wrapper: createWrapper(),
      });

      act(() => {
        result.current.mutate({
          brandingId: 'branding-123',
          data: emptyNameRequest,
        });
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
  });
});
