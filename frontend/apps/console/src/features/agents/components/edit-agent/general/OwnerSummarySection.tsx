// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {SettingsCard} from '@thunderid/components';
import {useGetUsers} from '@thunderid/configure-users';
import {Typography} from '@wso2/oxygen-ui';
import {useMemo, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import type {Agent} from '../../../models/agent';

interface OwnerSummarySectionProps {
  agent: Agent;
}

const formatUserLabel = (user: {id: string; display?: string; attributes?: Record<string, unknown>}): string => {
  if (user.display) return user.display;
  const attrs = user.attributes ?? {};
  const username = typeof attrs.username === 'string' ? attrs.username : undefined;
  const email = typeof attrs.email === 'string' ? attrs.email : undefined;
  return username ?? email ?? user.id;
};

/**
 * Read-only preview of this agent's owner, with no edit affordance — used on the General tab.
 * The Advanced tab is where the owner is actually assigned.
 */
export default function OwnerSummarySection({agent}: OwnerSummarySectionProps): JSX.Element {
  const {t} = useTranslation();
  const {data: usersData} = useGetUsers({limit: 100, offset: 0});

  const ownerLabel = useMemo(() => {
    if (!agent.owner) return undefined;
    const match = (usersData?.users ?? []).find((user) => user.id === agent.owner);
    return match ? formatUserLabel(match) : agent.owner;
  }, [usersData, agent.owner]);

  return (
    <SettingsCard
      title={t('agents:edit.general.sections.owner.title', 'Owner')}
      description={t(
        'agents:edit.general.sections.owner.summaryDescription',
        'The user who is accountable for this agent, shown in audit records and used as the contact point for questions about what this agent does. Assigning an owner does not give that user any special access to the agent. Manage this from the Advanced tab.',
      )}
    >
      {ownerLabel ? (
        <Typography variant="body1">{ownerLabel}</Typography>
      ) : (
        <Typography variant="body2" color="text.secondary">
          {t('agents:edit.general.owner.empty', 'No owner assigned')}
        </Typography>
      )}
    </SettingsCard>
  );
}
