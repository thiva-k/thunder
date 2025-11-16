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

import {useCallback, useEffect, useMemo, useRef, useState} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import {useConfig} from '@thunder/commons-contexts';
import type {
  ApiError,
  OrganizationUnitListParams,
  OrganizationUnitListResponse,
} from '../types/organization-units';

/**
 * Fetches organization units list with optional pagination.
 */
export default function useGetOrganizationUnits(params?: OrganizationUnitListParams) {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const [data, setData] = useState<OrganizationUnitListResponse | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);
  const abortControllerRef = useRef<AbortController | null>(null);

  const API_BASE_URL: string = useMemo(
    () => getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string),
    [getServerUrl],
  );
  const paramsKey = useMemo(() => JSON.stringify(params ?? {}), [params]);

  const executeFetch = useCallback(
    async (finalParams?: OrganizationUnitListParams): Promise<void> => {
      abortControllerRef.current?.abort();
      abortControllerRef.current = new AbortController();

      try {
        setLoading(true);
        setError(null);

        const searchParams = new URLSearchParams();
        const resolvedParams = finalParams ?? params;

        if (resolvedParams?.limit !== undefined) {
          searchParams.append('limit', String(resolvedParams.limit));
        }
        if (resolvedParams?.offset !== undefined) {
          searchParams.append('offset', String(resolvedParams.offset));
        }

        const queryString = searchParams.toString();

        const response = await http.request({
          url: `${API_BASE_URL}/organization-units${queryString ? `?${queryString}` : ''}`,
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
          signal: abortControllerRef.current?.signal,
        } as unknown as Parameters<typeof http.request>[0]);

        setData(response.data as OrganizationUnitListResponse);
        setError(null);
      } catch (err) {
        if (err instanceof Error && err.name === 'AbortError') {
          return;
        }

        const apiError: ApiError = {
          code: 'FETCH_ERROR',
          message: err instanceof Error ? err.message : 'An unknown error occurred',
          description: 'Failed to fetch organization units',
        };
        setError(apiError);
      } finally {
        setLoading(false);
      }
    },
    [API_BASE_URL, http, params],
  );

  useEffect(() => {
    executeFetch().catch(() => {
      // Error already handled
    });

    return () => {
      abortControllerRef.current?.abort();
    };
  }, [executeFetch, paramsKey]);

  const refetch = async (newParams?: OrganizationUnitListParams): Promise<void> => {
    await executeFetch(newParams).catch((err) => {
      // propagate errors except abort to match previous behavior
      if (!(err instanceof Error && err.name === 'AbortError')) {
        throw err;
      }
    });
  };

  return {
    data,
    loading,
    error,
    refetch,
  };
}
