// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {OAuth2GrantTypes} from '@thunderid/configure-applications';
import type {OAuth2Config} from '@thunderid/configure-applications';
import {Stack} from '@wso2/oxygen-ui';
import {useState} from 'react';
import {useTranslation} from 'react-i18next';
import AllowedUserTypesSection from './AllowedUserTypesSection';
import DangerZoneSection from './DangerZoneSection';
import OperationModesSection from './OperationModesSection';
import OwnerSection from './OwnerSection';
import TokenAudienceSelector, {
  type TokenAudienceOption,
} from '../../../../applications/components/common/TokenAudienceSelector';
import {applyGrantTypesChange} from '../../../../applications/utils/oauth2Rules';
import {DELEGATED_ONLY_GRANTS} from '../../../constants/delegationGrants';
import type {Agent, AgentInboundAuthConfig, OAuthAgentConfig} from '../../../models/agent';
import AgentDeleteDialog from '../../AgentDeleteDialog';

const ON_OWN_BEHALF = 'onOwnBehalf';
const ON_BEHALF_OF_USER = 'onBehalfOfUser';

interface EditAdvancedSettingsProps {
  agent: Agent;
  editedAgent: Partial<Agent>;
  oauth2Config?: OAuthAgentConfig;
  onFieldChange: (field: keyof Agent, value: unknown) => void;
  onDeleteSuccess?: () => void;
}

export default function EditAdvancedSettings({
  agent,
  editedAgent,
  oauth2Config = undefined,
  onFieldChange,
  onDeleteSuccess = undefined,
}: EditAdvancedSettingsProps) {
  const {t} = useTranslation();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const isUnlocked = oauth2Config?.grantTypes?.includes(OAuth2GrantTypes.AUTHORIZATION_CODE) ?? false;

  const handleOAuth2ConfigChange = (updates: Partial<OAuth2Config>) => {
    const currentInboundAuth: AgentInboundAuthConfig[] = editedAgent.inboundAuthConfig ?? agent.inboundAuthConfig ?? [];
    const updatedInboundAuth = currentInboundAuth.map((auth) =>
      auth.type === 'oauth2' ? {...auth, config: {...auth.config, ...updates} as OAuthAgentConfig} : auth,
    );
    onFieldChange('inboundAuthConfig', updatedInboundAuth);
  };

  // The "On behalf of a user" mode unlocks the delegated-only grants below and the Flows/Tokens
  // tabs. Switching modes just flips authorization_code on/off; applyGrantTypesChange handles the
  // dependent grants.
  const handleDelegationToggle = (checked: boolean): void => {
    if (!oauth2Config || checked === isUnlocked) return;
    const grantTypes = oauth2Config.grantTypes ?? [];
    const nextGrantTypes = checked
      ? [...new Set([...grantTypes, OAuth2GrantTypes.AUTHORIZATION_CODE])]
      : grantTypes.filter((grant) => !DELEGATED_ONLY_GRANTS.includes(grant));
    const updates = applyGrantTypesChange(oauth2Config, nextGrantTypes);
    // PKCE is fully derived from authorization_code for agents.
    if (checked) {
      updates.pkceRequired = true;
    }
    handleOAuth2ConfigChange(updates);
  };

  const mode = isUnlocked ? ON_BEHALF_OF_USER : ON_OWN_BEHALF;
  const modeOptions: TokenAudienceOption[] = [
    {
      value: ON_OWN_BEHALF,
      label: t('agents:edit.advanced.mode.onOwnBehalf.label', 'On its own behalf'),
      description: t(
        'agents:edit.advanced.mode.onOwnBehalf.description',
        'This agent authenticates with its own credentials without user interaction, using Client Credentials.',
      ),
    },
    {
      value: ON_BEHALF_OF_USER,
      label: t('agents:edit.advanced.mode.onBehalfOfUser.label', 'On behalf of a user'),
      description: t(
        'agents:edit.advanced.mode.onBehalfOfUser.description',
        'This agent acts on behalf of a signed-in user, using Authorization Code with PKCE.',
      ),
    },
  ];

  const handleModeChange = (next: string): void => {
    handleDelegationToggle(next === ON_BEHALF_OF_USER);
  };

  return (
    <Stack spacing={3}>
      <TokenAudienceSelector
        title={t('agents:edit.advanced.mode.title', 'Operating Mode')}
        options={modeOptions}
        value={mode}
        onChange={handleModeChange}
        disabled={!oauth2Config || agent.isReadOnly === true}
      >
        <Stack spacing={3}>
          <OwnerSection agent={agent} editedAgent={editedAgent} onFieldChange={onFieldChange} />
          <AllowedUserTypesSection
            agent={agent}
            editedAgent={editedAgent}
            oauth2Config={oauth2Config}
            onFieldChange={onFieldChange}
          />
          <OperationModesSection
            oauth2Config={oauth2Config}
            onOAuth2ConfigChange={handleOAuth2ConfigChange}
            disabled={agent.isReadOnly}
          />
          {!agent.isReadOnly && <DangerZoneSection onDeleteClick={() => setDeleteDialogOpen(true)} />}
        </Stack>
      </TokenAudienceSelector>

      <AgentDeleteDialog
        open={deleteDialogOpen}
        agentId={agent.id}
        onClose={() => setDeleteDialogOpen(false)}
        onSuccess={onDeleteSuccess}
      />
    </Stack>
  );
}
