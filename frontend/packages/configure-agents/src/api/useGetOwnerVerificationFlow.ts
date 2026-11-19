// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useQuery, type UseQueryResult} from '@tanstack/react-query';
import {useConfig} from '@thunderid/contexts';
import {useThunderID} from '@thunderid/react';
import {AdministrationFlowConfigKey, resolveAdministrationFlowHandle, type HttpLike} from '@thunderid/utils';
import AgentQueryKeys from '../constants/agent-query-keys';

/**
 * How long a resolved handle stays fresh. The flow section changes rarely.
 */
const STALE_TIME: number = 5 * 60 * 1000;

/**
 * Reads the handle of the agent owner verification flow the server configures, or an empty string when
 * none is set. Reading it needs permission to read the server configuration.
 *
 * @param enabled - Whether to read it. Pass false where the action is not offered.
 * @returns TanStack Query result holding the handle
 */
export default function useGetOwnerVerificationFlow(enabled: boolean): UseQueryResult<string> {
  const {http} = useThunderID();
  const {getServerUrl} = useConfig();

  return useQuery<string>({
    enabled,
    queryFn: async (): Promise<string> =>
      resolveAdministrationFlowHandle(
        http as unknown as HttpLike,
        getServerUrl(),
        AdministrationFlowConfigKey.AGENT_OWNER_VERIFICATION,
      ),
    queryKey: [AgentQueryKeys.OWNER_VERIFICATION_FLOW],
    staleTime: STALE_TIME,
  });
}
