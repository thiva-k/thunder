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
 * Request body for creating a new user
 */
export interface CreateUserRequest {
  organizationUnit: string;
  type: string;
  groups?: string[];
  attributes: Record<string, unknown>;
}

/**
 * Response after creating a user
 */
export interface CreateUserResponse {
  id: string;
  organizationUnit: string;
  type: string;
  attributes: Record<string, unknown>;
}

/**
 * Hook to create a new user
 * POST https://localhost:8090/users
 */
export default function useCreateUser() {
  const {http} = useAsgardeo();
  const [data, setData] = useState<CreateUserResponse | null>(null);
  const [error, setError] = useState<ApiError | null>(null);
  const [loading, setLoading] = useState(false);

  const createUser = useCallback(async (userData: CreateUserRequest) => {
    try {
      setLoading(true);
      setError(null);
      setData(null);

      const url = `${API_BASE_URL}/users`;

      const response = await http.request({
        url,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        data: userData,
      } as unknown as Parameters<typeof http.request>[0]);

      const result = response.data as CreateUserResponse;
      setData(result);

      return result;
    } catch (err) {
      if (err instanceof Error) {
        const apiError: ApiError = {
          code: 'CREATE_USER_ERROR',
          message: err.message,
          description: 'An error occurred while creating the user',
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
    setData(null);
    setError(null);
  }, []);

  return {
    createUser,
    data,
    loading,
    error,
    reset,
  };
}
