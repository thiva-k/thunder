// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {SettingsCard} from '@thunderid/components';
import {FormControl, FormLabel, IconButton, InputAdornment, TextField, Tooltip} from '@wso2/oxygen-ui';
import {Check, Copy} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import type {OAuthAgentConfig} from '../../../models/agent';

interface ClientIdSectionProps {
  oauth2Config?: OAuthAgentConfig;
  copiedField: string | null;
  onCopyToClipboard: (text: string, fieldName: string) => Promise<void>;
}

export default function ClientIdSection({
  oauth2Config = undefined,
  copiedField,
  onCopyToClipboard,
}: ClientIdSectionProps): JSX.Element | null {
  const {t} = useTranslation();

  if (!oauth2Config?.clientId) return null;

  return (
    <SettingsCard
      title={t('agents:edit.credentials.clientId.title', 'Client ID')}
      description={t(
        'agents:edit.credentials.clientId.description',
        'The public identifier this agent uses to authenticate as a client.',
      )}
    >
      <FormControl fullWidth>
        <FormLabel htmlFor="agent-client-id-input">
          {t('agents:edit.credentials.clientSecret.clientIdLabel', 'Client ID')}
        </FormLabel>
        <TextField
          fullWidth
          id="agent-client-id-input"
          value={oauth2Config.clientId}
          InputProps={{
            readOnly: true,
            endAdornment: (
              <InputAdornment position="end">
                <Tooltip title={copiedField === 'clientId' ? t('common:actions.copied') : t('common:actions.copy')}>
                  <IconButton
                    onClick={() => {
                      if (oauth2Config.clientId) {
                        onCopyToClipboard(oauth2Config.clientId, 'clientId').catch(() => null);
                      }
                    }}
                    edge="end"
                  >
                    {copiedField === 'clientId' ? <Check size={16} /> : <Copy size={16} />}
                  </IconButton>
                </Tooltip>
              </InputAdornment>
            ),
          }}
          sx={{'& input': {fontFamily: 'monospace', fontSize: '0.875rem'}}}
        />
      </FormControl>
    </SettingsCard>
  );
}
