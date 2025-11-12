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

import {useState, useCallback, useEffect} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import type {ApiUser, ApiError} from '../types/users';

const API_BASE_URL = 'https://localhost:8090';

/**
 * Hook to fetch a single user by ID
 * GET https://localhost:8090/users/{userId}
 */
export default function useGetUser(userId: string | undefined) {
  const {http} = useAsgardeo();
  const [data, setData] = useState<ApiUser | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const fetchUser = useCallback(async () => {
    if (!userId) {
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const url = `${API_BASE_URL}/users/${userId}`;

      const response = await http.request({
        url,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      const result = response.data as ApiUser;
      setData(result);
    } catch (err) {
      if (err instanceof Error) {
        setError({
          code: 'FETCH_ERROR',
          message: err.message,
          description: 'An error occurred while fetching user',
        });
      }
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [userId, http]);

  useEffect(() => {
    fetchUser().catch(() => {
      // Error already handled
    });
  }, [fetchUser]);

  const refetch = useCallback(() => fetchUser(), [fetchUser]);

  return {
    data,
    loading: isLoading,
    error,
    refetch,
  };
}
