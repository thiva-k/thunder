// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useQuery, type UseQueryResult} from '@tanstack/react-query';
import {useConfig} from '@thunderid/contexts';
import {useThunderID} from '@thunderid/react';
import SettingsQueryKeys from '../constants/settings-query-keys';
import type {CimdConfigResponse} from '../models/cimd';

/**
 * Fetches the `cimd` server-config section: whether unregistered clients publishing a Client ID Metadata
 * Document may sign in, and the blocked client IDs.
 *
 * @param enabled - Whether to read the section, so a caller without access to it reads nothing
 * @returns TanStack Query result containing the section's layers.
 *
 * @public
 */
export default function useGetCimdConfig(enabled = true): UseQueryResult<CimdConfigResponse> {
  const {http} = useThunderID();
  const {getServerUrl} = useConfig();

  return useQuery<CimdConfigResponse>({
    queryKey: [SettingsQueryKeys.SERVER_CONFIG, SettingsQueryKeys.CIMD],
    enabled,
    queryFn: async (): Promise<CimdConfigResponse> => {
      const serverUrl: string = getServerUrl();

      const response: {
        data: CimdConfigResponse;
      } = await http.request({
        url: `${serverUrl}/server-config/cimd`,
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
  });
}
