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

import {useState, useCallback} from 'react';
import {useAsgardeo} from '@asgardeo/react';
import type {ApiError} from '../types/users';

const API_BASE_URL = 'https://localhost:8090';

/**
 * Hook to delete a user by ID
 * DELETE https://localhost:8090/users/{userId}
 */
export default function useDeleteUser() {
  const {http} = useAsgardeo();
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const deleteUser = useCallback(async (userId: string) => {
    try {
      setLoading(true);
      setError(null);

      const url = `${API_BASE_URL}/users/${userId}`;

      await http.request({
        url,
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      // 204 No Content - successful deletion
      return true;
    } catch (err) {
      if (err instanceof Error) {
        const apiError: ApiError = {
          code: 'DELETE_USER_ERROR',
          message: err.message,
          description: 'An error occurred while deleting the user',
        };
        setError(apiError);
        throw err;
      }
      throw err;
    } finally {
      setLoading(false);
    }
  }, [http]);

  const reset = useCallback(() => {
    setError(null);
  }, []);

  return {
    deleteUser,
    loading,
    error,
    reset,
  };
}
