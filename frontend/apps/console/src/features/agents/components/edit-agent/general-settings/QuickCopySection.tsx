// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {SettingsCard} from '@thunderid/components';
import {Stack, TextField, InputAdornment, Tooltip, IconButton, FormControl, FormLabel} from '@wso2/oxygen-ui';
import {Copy, Check} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {Agent} from '../../../models/agent';

interface QuickCopySectionProps {
  agent: Agent;
  copiedField: string | null;
  onCopyToClipboard: (text: string, fieldName: string) => Promise<void>;
}

export default function QuickCopySection({agent, copiedField, onCopyToClipboard}: QuickCopySectionProps) {
  const {t} = useTranslation();

  return (
    <SettingsCard
      title={t('agents:edit.general.sections.quickCopy.title', 'Identifier')}
      description={t('agents:edit.general.sections.quickCopy.description', 'The unique identifier for this agent.')}
    >
      <Stack spacing={3}>
        <FormControl fullWidth>
          <FormLabel htmlFor="agent-id-input">{t('agents:edit.general.labels.agentId', 'Agent ID')}</FormLabel>
          <TextField
            fullWidth
            id="agent-id-input"
            value={agent.id}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={copiedField === 'agent_id' ? t('common:actions.copied') : t('common:actions.copy')}>
                    <IconButton
                      onClick={() => {
                        onCopyToClipboard(agent.id, 'agent_id').catch(() => null);
                      }}
                      edge="end"
                    >
                      {copiedField === 'agent_id' ? <Check size={16} /> : <Copy size={16} />}
                    </IconButton>
                  </Tooltip>
                </InputAdornment>
              ),
            }}
            helperText={t('agents:edit.general.agentId.hint', 'Unique identifier for this agent')}
            sx={{
              '& input': {
                fontFamily: 'monospace',
                fontSize: '0.875rem',
              },
            }}
          />
        </FormControl>
      </Stack>
    </SettingsCard>
  );
}
