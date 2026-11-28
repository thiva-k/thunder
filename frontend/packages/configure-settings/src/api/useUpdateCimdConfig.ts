// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useMutation, useQueryClient, type UseMutationResult} from '@tanstack/react-query';
import {useConfig} from '@thunderid/contexts';
import {useThunderID} from '@thunderid/react';
import SettingsQueryKeys from '../constants/settings-query-keys';
import type {CimdConfigResponse, CimdSettings} from '../models/cimd';

/**
 * Variables for the {@link useUpdateCimdConfig} mutation.
 *
 * @public
 */
export interface UpdateCimdConfigVariables {
  /**
   * The complete writable value of the section to persist.
   */
  data: CimdSettings;
}

/**
 * Replaces the writable layer of the `cimd` server-config section. Each caller decides how to report the
 * outcome, since the same write turns the feature on from a switch and blocks a client from a field.
 *
 * @returns TanStack Query mutation for updating the section.
 *
 * @public
 */
export default function useUpdateCimdConfig(): UseMutationResult<CimdConfigResponse, Error, UpdateCimdConfigVariables> {
  const {http} = useThunderID();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();

  return useMutation<CimdConfigResponse, Error, UpdateCimdConfigVariables>({
    mutationFn: async ({data}: UpdateCimdConfigVariables): Promise<CimdConfigResponse> => {
      const serverUrl: string = getServerUrl();

      const response: {
        data: CimdConfigResponse;
      } = await http.request({
        url: `${serverUrl}/server-config/cimd`,
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        data,
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: (data) => {
      // PUT returns the full config, so keep the read model in sync without a refetch.
      queryClient.setQueryData([SettingsQueryKeys.SERVER_CONFIG, SettingsQueryKeys.CIMD], data);
    },
  });
}
