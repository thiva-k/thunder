// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useQuery, type UseQueryResult} from '@tanstack/react-query';
import {VerifiablePresentationQueryKeys} from '@thunderid/configure-verifiable-presentations';
import type {VerifiablePresentation} from '@thunderid/configure-verifiable-presentations';
import {useConfig} from '@thunderid/contexts';
import {useThunderID} from '@thunderid/react';

/**
 * Fetches a single OpenID4VP presentation definition by id.
 */
export default function useGetVerifiablePresentation(id: string): UseQueryResult<VerifiablePresentation> {
  const {http} = useThunderID();
  const {getServerUrl} = useConfig();

  return useQuery<VerifiablePresentation>({
    queryKey: [VerifiablePresentationQueryKeys.VP, id],
    enabled: Boolean(id),
    queryFn: async (): Promise<VerifiablePresentation> => {
      const serverUrl: string = getServerUrl();
      const response: {data: VerifiablePresentation} = await http.request({
        url: `${serverUrl}/openid4vp/presentation-definitions/${id}`,
        method: 'GET',
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
  });
}
