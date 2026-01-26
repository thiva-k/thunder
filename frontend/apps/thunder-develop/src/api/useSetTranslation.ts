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

import {useMutation, useQueryClient, type UseMutationResult} from '@tanstack/react-query';
import {useConfig} from '@thunder/commons-contexts';
import {useAsgardeo} from '@asgardeo/react';
import {invalidateI18nCache} from '../i18n/invalidate-i18n-cache';
import I18nQueryKeys from './I18nQueryKeys';

/**
 * Response from the translation API.
 */
export interface TranslationResponse {
  language: string;
  namespace: string;
  key: string;
  value: string;
}

/**
 * Variables for the set translation mutation.
 */
export interface SetTranslationVariables {
  language: string;
  namespace: string;
  key: string;
  value: string;
}

/**
 * Custom hook to create or update a single translation.
 *
 * @returns TanStack Query mutation object for setting translations
 *
 * @example
 * ```tsx
 * function CreateTranslationForm() {
 *   const setTranslation = useSetTranslation();
 *
 *   const handleSubmit = (data: SetTranslationVariables) => {
 *     setTranslation.mutate(data, {
 *       onSuccess: (translation) => {
 *         console.log('Translation created:', translation);
 *       },
 *       onError: (error) => {
 *         console.error('Failed to create translation:', error);
 *       }
 *     });
 *   };
 *
 *   return <form onSubmit={handleSubmit}>...</form>;
 * }
 * ```
 */
export default function useSetTranslation(): UseMutationResult<TranslationResponse, Error, SetTranslationVariables> {
  const {http} = useAsgardeo();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();

  return useMutation<TranslationResponse, Error, SetTranslationVariables>({
    mutationFn: async ({language, namespace, key, value}: SetTranslationVariables): Promise<TranslationResponse> => {
      const serverUrl: string = getServerUrl();
      const response: {
        data: TranslationResponse;
      } = await http.request({
        url: `${serverUrl}/i18n/languages/${language}/translations/ns/${namespace}/keys/${key}`,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        data: JSON.stringify({value}),
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: (_data, variables) => {
      // Invalidate translations cache after successful update
      queryClient.invalidateQueries({queryKey: [I18nQueryKeys.TRANSLATIONS]}).catch(() => {
        // Ignore invalidation errors
      });
      queryClient.invalidateQueries({queryKey: [I18nQueryKeys.TRANSLATIONS, variables.language]}).catch(() => {
        // Ignore invalidation errors
      });

      // Also invalidate the app-level i18n cache to refresh i18next resources
      invalidateI18nCache();
    },
  });
}
