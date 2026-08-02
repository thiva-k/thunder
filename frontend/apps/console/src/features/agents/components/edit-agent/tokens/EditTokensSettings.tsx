// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {OAuth2GrantTypes} from '@thunderid/configure-applications';
import type {Application} from '@thunderid/configure-applications';
import {Box, Stack, Tab, Tabs} from '@wso2/oxygen-ui';
import {useEffect, useState, type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';
import AgentAccessTokenSection from './AgentAccessTokenSection';
import SettingsLockNotice from '../../../../applications/components/common/SettingsLockNotice';
import EditTokenSettings from '../../../../applications/components/edit-application/token-settings/EditTokenSettings';
import type {Agent, OAuthAgentConfig} from '../../../models/agent';

interface EditTokensSettingsProps {
  agent: Agent;
  editedAgent: Partial<Agent>;
  oauth2Config?: OAuthAgentConfig;
  onFieldChange: (field: keyof Agent, value: unknown) => void;
  onValidationChange?: (hasErrors: boolean) => void;
  sectionResetKey?: number;
}

export default function EditTokensSettings({
  agent,
  editedAgent,
  oauth2Config = undefined,
  onFieldChange,
  onValidationChange = undefined,
  sectionResetKey = 0,
}: EditTokensSettingsProps): JSX.Element {
  const {t} = useTranslation();
  const [subTab, setSubTab] = useState(0);
  const [userTabHasError, setUserTabHasError] = useState(false);
  const [agentTabHasError, setAgentTabHasError] = useState(false);

  const isUnlocked = oauth2Config?.grantTypes?.includes(OAuth2GrantTypes.AUTHORIZATION_CODE) ?? false;

  useEffect(() => {
    onValidationChange?.(userTabHasError || agentTabHasError);
  }, [userTabHasError, agentTabHasError, onValidationChange]);

  // Forcing isReadOnly disables every input via EditTokenSettings' existing
  // disabled={application.isReadOnly} wiring when Delegated mode isn't on.
  const appLikeAgent = {...agent, isReadOnly: (agent.isReadOnly ?? false) || !isUnlocked} as unknown as Application;
  const appHandleFieldChange = onFieldChange as unknown as (field: keyof Application, value: unknown) => void;

  const handleSubTabChange = (_event: SyntheticEvent, newValue: number): void => {
    setSubTab(newValue);
  };

  return (
    <Box>
      <Tabs value={subTab} onChange={handleSubTabChange} aria-label="agent token settings sub-tabs">
        <Tab label={t('agents:edit.tokens.tabs.agent', 'Agent')} sx={{textTransform: 'none'}} />
        <Tab label={t('agents:edit.tokens.tabs.user', 'User')} sx={{textTransform: 'none'}} />
      </Tabs>
      <Box sx={{pt: 3}}>
        {subTab === 0 && (
          <AgentAccessTokenSection
            key={sectionResetKey}
            agent={agent}
            editedAgent={editedAgent}
            oauth2Config={oauth2Config}
            onFieldChange={onFieldChange}
            onValidationChange={setAgentTabHasError}
          />
        )}
        {subTab === 1 && (
          <SettingsLockNotice
            isUnlocked={isUnlocked}
            message={t(
              'agents:edit.tokens.delegationLock.message',
              'These settings are frozen for this agent. Turn on Delegated mode in the Advanced tab to unlock and start using them.',
            )}
          >
            <Stack spacing={3}>
              <EditTokenSettings
                application={appLikeAgent}
                oauth2Config={oauth2Config}
                onFieldChange={appHandleFieldChange}
                onValidationChange={setUserTabHasError}
                entityLabel="agent"
                showUserInfoTab={false}
                showActorClaim
                actorSub={agent.id}
                certificateLocation="Credentials"
                sectionResetKey={sectionResetKey}
              />
            </Stack>
          </SettingsLockNotice>
        )}
      </Box>
    </Box>
  );
}
