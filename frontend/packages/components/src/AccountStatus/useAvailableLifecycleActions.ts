// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useQuery} from '@tanstack/react-query';
import {AdministrationModes, resolveAdministrationOperation, useAdministration, useConfig} from '@thunderid/contexts';
import {useThunderID} from '@thunderid/react';
import {
  AdministrationFlowConfigKey,
  resolveAdministrationFlows,
  type AdministrationFlowConfigKeyValue,
  type HttpLike,
} from '@thunderid/utils';
import {useMemo} from 'react';
import type {LifecycleCategory, LifecycleOperation} from './lifecycleAction';

/**
 * Server-config purpose naming the administration flow for each category and operation. Applications
 * are never locked, so they have no unlock.
 */
export const LIFECYCLE_FLOW_PURPOSES: Record<
  LifecycleCategory,
  Partial<Record<LifecycleOperation, AdministrationFlowConfigKeyValue>>
> = {
  agent: {
    suspend: AdministrationFlowConfigKey.AGENT_SUSPEND,
    unlock: AdministrationFlowConfigKey.AGENT_UNLOCK,
    unsuspend: AdministrationFlowConfigKey.AGENT_UNSUSPEND,
  },
  application: {
    suspend: AdministrationFlowConfigKey.APPLICATION_SUSPEND,
    unsuspend: AdministrationFlowConfigKey.APPLICATION_UNSUSPEND,
  },
  user: {
    suspend: AdministrationFlowConfigKey.USER_SUSPEND,
    unlock: AdministrationFlowConfigKey.USER_UNLOCK,
    unsuspend: AdministrationFlowConfigKey.USER_UNSUSPEND,
  },
};

/**
 * Administration feature name of each category, as the other operations of its package use it.
 */
const ADMINISTRATION_FEATURE: Record<LifecycleCategory, string> = {
  agent: 'agents',
  application: 'applications',
  user: 'users',
};

const ALL_PURPOSES: AdministrationFlowConfigKeyValue[] = Object.values(LIFECYCLE_FLOW_PURPOSES).flatMap((operations) =>
  Object.values(operations).filter((purpose): purpose is AdministrationFlowConfigKeyValue => Boolean(purpose)),
);

/**
 * Query key for the resolved lifecycle flows, one entry for every category.
 */
export const LIFECYCLE_FLOWS_QUERY_KEY = ['lifecycle-flows'] as const;

/**
 * How long a resolution stays fresh. The flow section and flows change rarely.
 */
const LIFECYCLE_FLOWS_STALE_TIME: number = 5 * 60 * 1000;

export interface AvailableLifecycleActions {
  /**
   * Flow id for each operation this category can perform now. An operation is absent when it does not
   * run as a flow, no flow is configured for it, or the configured handle names no existing flow.
   */
  available: Partial<Record<LifecycleOperation, string>>;
  /**
   * True while availability is being resolved.
   */
  isPending: boolean;
  /**
   * True when the flows could not be resolved, so nothing is offered.
   */
  isError: boolean;
  /**
   * Why the flows could not be resolved.
   */
  error: Error | null;
}

/**
 * Resolves which lifecycle operations a deployment offers for a category. The operations have no
 * native endpoint, so one is offered only in flow mode and only when its flow resolves.
 *
 * @param category - Identity category the caller is rendering.
 * @returns The operations available for that category, and the resolution state.
 */
export default function useAvailableLifecycleActions(category: LifecycleCategory): AvailableLifecycleActions {
  const {http} = useThunderID();
  const {getServerUrl} = useConfig();
  const administration = useAdministration();

  const purposes = useMemo(
    () =>
      Object.entries(LIFECYCLE_FLOW_PURPOSES[category]).filter(
        ([operation]) =>
          resolveAdministrationOperation(administration, ADMINISTRATION_FEATURE[category], operation) ===
          AdministrationModes.FLOW,
      ),
    [administration, category],
  );
  const runsAsFlow = purposes.length > 0;

  const query = useQuery<Record<string, string>, Error>({
    enabled: runsAsFlow,
    queryFn: async (): Promise<Record<string, string>> =>
      resolveAdministrationFlows(http as unknown as HttpLike, getServerUrl(), ALL_PURPOSES),
    queryKey: LIFECYCLE_FLOWS_QUERY_KEY,
    staleTime: LIFECYCLE_FLOWS_STALE_TIME,
  });

  // Memoized so callers' callbacks keyed on it hold across renders.
  const available = useMemo<Partial<Record<LifecycleOperation, string>>>(() => {
    const resolved = query.data ?? {};
    const result: Partial<Record<LifecycleOperation, string>> = {};

    purposes.forEach(([operation, purpose]) => {
      const flowId = purpose ? resolved[purpose] : undefined;

      if (flowId) {
        result[operation as LifecycleOperation] = flowId;
      }
    });

    return result;
  }, [purposes, query.data]);

  return {
    available,
    error: runsAsFlow ? query.error : null,
    isError: runsAsFlow && query.isError,
    isPending: runsAsFlow && query.isPending,
  };
}
