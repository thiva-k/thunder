// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useMutation, useQueryClient, type UseMutationResult} from '@tanstack/react-query';
import {useConfig, useToast} from '@thunderid/contexts';
import {useThunderID} from '@thunderid/react';
import {getErrorMessage} from '@thunderid/utils';
import {useTranslation} from 'react-i18next';
import ConnectionQueryKeys from '../constants/query-keys';
import type {ConnectionRequest, ConnectionResponse, ConnectionType} from '../models/connection';
import isConflictError from '../utils/isConflictError';

/**
 * Update a connection instance (PUT /connections/{type}/{id}).
 *
 * Omit `clientSecret` from the payload to keep the stored value. Conflicts (409) are not
 * toasted here — the caller surfaces them inline next to the name field.
 */
export default function useUpdateConnection(
  type: ConnectionType,
  id: string,
): UseMutationResult<ConnectionResponse, Error, ConnectionRequest> {
  const {http} = useThunderID();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();
  const {t} = useTranslation('connections');
  const {showToast} = useToast();

  return useMutation<ConnectionResponse, Error, ConnectionRequest>({
    mutationFn: async (data: ConnectionRequest): Promise<ConnectionResponse> => {
      const serverUrl: string = getServerUrl();
      const response: {
        data: ConnectionResponse;
      } = await http.request({
        url: `${serverUrl}/connections/${type}/${id}`,
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        data,
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({queryKey: [ConnectionQueryKeys.CONNECTIONS]}).catch(() => {
        // Ignore invalidation errors
      });
      queryClient.invalidateQueries({queryKey: [ConnectionQueryKeys.CONNECTION_INSTANCES, type]}).catch(() => {
        // Ignore invalidation errors
      });
      queryClient.invalidateQueries({queryKey: [ConnectionQueryKeys.CONNECTION, type, id]}).catch(() => {
        // Ignore invalidation errors
      });
      showToast(t('update.success'), 'success');
    },
    onError: (error) => {
      if (!isConflictError(error)) {
        showToast(getErrorMessage(error, t, 'update.error'), 'error');
      }
    },
  });
}
