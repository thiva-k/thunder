// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {QueryErrorNotice, SettingsCard} from '@thunderid/components';
import {useCapabilitiesContext, useToast} from '@thunderid/contexts';
import {getErrorMessage} from '@thunderid/utils';
import {Box, Button, Skeleton, Stack, Switch, TextField, Typography} from '@wso2/oxygen-ui';
import {InfoIcon} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useCallback, useState} from 'react';
import {useTranslation} from 'react-i18next';
import useGetCimdConfig from '../../api/useGetCimdConfig';
import useUpdateCimdConfig from '../../api/useUpdateCimdConfig';
import type {CimdSettings} from '../../models/cimd';
import {isValidClientIdUrl, toCimdSettings} from '../../utils/cimd';

/**
 * The Client ID Metadata Documents tab. It holds the deployment-wide controls for clients that publish a
 * Client ID Metadata Document, such as MCP clients and AI agents: a switch that lets them sign in without
 * being registered, and a block list that refuses single clients. Which clients sign in, and how, is set by
 * the applications and agents that list the domains of their client ID URLs.
 *
 * Every change is written at once, so there is no unsaved state. A declarative `cimd` section makes the tab
 * read-only.
 *
 * @returns The rendered tab body
 *
 * @public
 */
export default function CimdSection(): JSX.Element {
  const {t} = useTranslation();
  const {showToast} = useToast();
  const {can} = useCapabilitiesContext();
  const {data, isLoading, error, refetch} = useGetCimdConfig();
  const updateConfig = useUpdateCimdConfig();
  const [blockInput, setBlockInput] = useState('');
  const [blockError, setBlockError] = useState<string | undefined>(undefined);

  const declarative = Boolean(data?.readOnly && (data.readOnly.enabled || data.readOnly.blockedClientIds?.length));
  const settings: CimdSettings = toCimdSettings(declarative ? data?.readOnly : data?.writable);
  const editable = can('serverconfig', 'update') && !declarative;

  const tForErrors = useCallback(
    (key: string, options?: Record<string, unknown>): string => t(key.includes(':') ? key : `settings:${key}`, options),
    [t],
  );

  const saveErrorMessage = (mutationError: Error): string =>
    getErrorMessage(mutationError, tForErrors, 'cimd.save.error', 'Failed to update the settings.');

  const handleToggle = (enabled: boolean): void => {
    updateConfig.mutate(
      {data: {...settings, enabled}},
      {
        onSuccess: () =>
          showToast(
            enabled
              ? t('settings:cimd.enabled.success', 'Unregistered clients can now sign in.')
              : t('settings:cimd.disabled.success', 'Unregistered clients can no longer sign in.'),
            'success',
          ),
        onError: (mutationError) => showToast(saveErrorMessage(mutationError), 'error'),
      },
    );
  };

  const handleBlock = (): void => {
    const clientId = blockInput.trim();
    if (!isValidClientIdUrl(clientId)) {
      setBlockError(
        t(
          'settings:cimd.blocked.validation.invalid',
          'Enter a client ID URL, for example https://example.com/client.json.',
        ),
      );
      return;
    }
    if (settings.blockedClientIds.includes(clientId)) {
      setBlockError(t('settings:cimd.blocked.validation.duplicate', 'This client is already blocked.'));
      return;
    }
    updateConfig.mutate(
      {data: {...settings, blockedClientIds: [...settings.blockedClientIds, clientId]}},
      {
        onSuccess: () => {
          setBlockInput('');
          showToast(t('settings:cimd.blocked.add.success', 'Client blocked.'), 'success');
        },
        onError: (mutationError) => setBlockError(saveErrorMessage(mutationError)),
      },
    );
  };

  const handleUnblock = (clientId: string): void => {
    updateConfig.mutate(
      {data: {...settings, blockedClientIds: settings.blockedClientIds.filter((id) => id !== clientId)}},
      {
        onSuccess: () => showToast(t('settings:cimd.blocked.remove.success', 'Client unblocked.'), 'success'),
        onError: (mutationError) => showToast(saveErrorMessage(mutationError), 'error'),
      },
    );
  };

  if (isLoading) {
    return (
      <Stack spacing={3}>
        <Skeleton variant="rounded" height={120} />
        <Skeleton variant="rounded" height={160} />
      </Stack>
    );
  }
  if (error) {
    return (
      <QueryErrorNotice
        error={error}
        t={tForErrors}
        variant="inline"
        fallbackKey="settings:cimd.load.error"
        fallbackDefaultValue="Failed to load the Client ID Metadata Document settings."
        onRetry={() => void refetch()}
      />
    );
  }

  const switchLabel = t('settings:cimd.toggle.ariaLabel', 'Let unregistered clients sign in');

  return (
    <Stack spacing={3}>
      {declarative && (
        <Stack direction="row" spacing={1} alignItems="flex-start" data-testid="cimd-read-only-hint">
          <Box aria-hidden sx={{flex: 'none', display: 'inline-flex', mt: 0.25, color: 'text.secondary'}}>
            <InfoIcon size={16} />
          </Box>
          <Typography variant="body2" color="text.secondary">
            {t(
              'settings:cimd.readOnlyHint',
              "These settings are read-only because they're managed declaratively. Change them in the declarative configuration.",
            )}
          </Typography>
        </Stack>
      )}

      <SettingsCard
        title={t('settings:cimd.card.title', 'Unregistered clients')}
        description={t(
          'settings:cimd.card.description',
          'Let MCP clients and AI agents, such as Claude, ChatGPT, and Visual Studio Code, sign in without registering each one. They identify themselves with a Client ID Metadata Document published at their client ID URL. A client registered with its own client ID always takes precedence.',
        )}
        headerAction={
          <Switch
            checked={settings.enabled}
            disabled={!editable || updateConfig.isPending}
            onChange={(event) => handleToggle(event.target.checked)}
            // Overriding slotProps.input drops MUI's default switch role, so restate it.
            slotProps={{input: {role: 'switch', 'aria-label': switchLabel}}}
            data-testid="cimd-enabled-switch"
          />
        }
      >
        <Stack spacing={1}>
          <Typography variant="body2" color="text.secondary" data-testid="cimd-domains-hint">
            {t(
              'settings:cimd.domains.hint',
              "An application or agent lets these clients sign in when its Client ID Metadata Document settings list the domain of their client ID URL. The clients then use that application or agent's settings.",
            )}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {t(
              'settings:cimd.card.hint',
              'Users still sign in and give consent as usual. Consent is asked for each client separately.',
            )}
          </Typography>
        </Stack>
      </SettingsCard>

      <SettingsCard
        title={t('settings:cimd.blocked.title', 'Blocked clients')}
        description={t(
          'settings:cimd.blocked.description',
          "Clients listed here can't sign in, even when an application or agent allows their domain.",
        )}
      >
        <Stack spacing={2}>
          {editable && (
            <Stack direction="row" spacing={1} alignItems="flex-start">
              <TextField
                fullWidth
                size="small"
                value={blockInput}
                placeholder={t('settings:cimd.blocked.placeholder', 'https://example.com/oauth/client.json')}
                onChange={(event) => {
                  setBlockInput(event.target.value);
                  setBlockError(undefined);
                }}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    event.preventDefault();
                    handleBlock();
                  }
                }}
                error={Boolean(blockError)}
                helperText={blockError}
                inputProps={{
                  'aria-label': t('settings:cimd.blocked.inputLabel', 'Client ID URL to block'),
                  'data-testid': 'cimd-block-input',
                }}
                sx={{'& input': {fontFamily: 'monospace'}}}
              />
              <Button
                variant="outlined"
                onClick={handleBlock}
                disabled={!blockInput.trim() || updateConfig.isPending}
                sx={{flex: 'none'}}
              >
                {t('settings:cimd.blocked.add', 'Block')}
              </Button>
            </Stack>
          )}
          {settings.blockedClientIds.length === 0 ? (
            <Typography variant="body2" color="text.secondary">
              {t('settings:cimd.blocked.empty', 'No clients are blocked.')}
            </Typography>
          ) : (
            <Stack spacing={1}>
              {settings.blockedClientIds.map((clientId) => (
                <Stack
                  key={clientId}
                  direction="row"
                  alignItems="center"
                  justifyContent="space-between"
                  spacing={2}
                  sx={{border: 1, borderColor: 'divider', borderRadius: 1, px: 2, py: 1}}
                  data-testid="cimd-blocked-row"
                >
                  <Typography variant="body2" sx={{fontFamily: 'monospace', wordBreak: 'break-all'}}>
                    {clientId}
                  </Typography>
                  {editable && (
                    <Button size="small" onClick={() => handleUnblock(clientId)} disabled={updateConfig.isPending}>
                      {t('settings:cimd.blocked.remove', 'Unblock')}
                    </Button>
                  )}
                </Stack>
              ))}
            </Stack>
          )}
        </Stack>
      </SettingsCard>
    </Stack>
  );
}
