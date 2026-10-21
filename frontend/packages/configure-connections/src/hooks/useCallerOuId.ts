// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useQuery} from '@tanstack/react-query';
import {useThunderID} from '@thunderid/react';
import ConnectionQueryKeys from '../constants/query-keys';

/**
 * The organization unit of the signed-in user, read from the `ouId` claim of their ID token. A new connection is owned
 * by it, so the create wizards send it as `ouId` without asking.
 *
 * `ouId` is an empty string while the token is read, and stays empty when the token does not carry the claim.
 */
export default function useCallerOuId(): {ouId: string; isLoading: boolean} {
  const {getDecodedIdToken} = useThunderID();

  const {data, isLoading} = useQuery<string>({
    queryKey: [ConnectionQueryKeys.CALLER_OU_ID],
    queryFn: async (): Promise<string> => {
      const claims = (await getDecodedIdToken()) as {ouId?: unknown} | undefined;

      return typeof claims?.ouId === 'string' ? claims.ouId : '';
    },
  });

  return {ouId: data ?? '', isLoading};
}
