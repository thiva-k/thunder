// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {SettingsCard} from '@thunderid/components';
import type {OAuth2Config} from '@thunderid/configure-applications';
import {Box, FormControlLabel, Switch, Typography} from '@wso2/oxygen-ui';
import type {ReactNode} from 'react';
import {Trans, useTranslation} from 'react-i18next';
import {deriveOAuth2Flags} from '../../../../applications/utils/oauth2Rules';
import {codeComponents} from '../shared/transCodeComponents';

interface SecuritySectionProps {
  oauth2Config?: OAuth2Config;
  onOAuth2ConfigChange?: (updates: Partial<OAuth2Config>) => void;
  disabled?: boolean;
}

export default function SecuritySection({
  oauth2Config = undefined,
  onOAuth2ConfigChange = undefined,
  disabled = false,
}: SecuritySectionProps) {
  const {t} = useTranslation();

  if (!oauth2Config) return null;

  const isEditable = Boolean(onOAuth2ConfigChange) && !disabled;
  const flags = deriveOAuth2Flags(oauth2Config);
  const isPkceForced = flags.isPkceForcedByPublicClient;
  // PKCE is fully derived from the authorization_code grant for agents — it is never a
  // manually editable setting, it just reflects whether that grant is selected.
  const isPkceRequired = isPkceForced || flags.hasAuthorizationCodeGrant;

  const getPkceCaption = (): ReactNode => {
    if (isPkceForced) {
      return t(
        'agents:edit.advanced.security.pkce.forced',
        'This agent is set up as a public client, so PKCE is required and cannot be turned off.',
      );
    }
    if (flags.isPkceDisabledByGrants) {
      return (
        <Trans
          i18nKey="agents:edit.advanced.security.pkce.notApplicable"
          defaults="PKCE only applies to the <code>authorization_code</code> grant. Turn that on to enable this setting."
          components={codeComponents}
        />
      );
    }
    return t(
      'agents:edit.advanced.security.pkce.on',
      'authorization_code is on for this agent, so PKCE is required automatically.',
    );
  };

  return (
    <SettingsCard
      title={t('agents:edit.advanced.security.title', 'Security')}
      description={t(
        'agents:edit.advanced.security.description',
        'Controls how this agent protects the authorization code exchange when a user signs in.',
      )}
    >
      <Box sx={{display: 'flex', flexDirection: 'column', gap: 3}}>
        <Box>
          <FormControlLabel
            control={
              <Switch
                checked={isPkceRequired}
                disabled
                inputProps={{'aria-label': t('agents:edit.advanced.security.pkce.label', 'Require PKCE')}}
              />
            }
            label={
              <Typography variant="subtitle2">
                {t('agents:edit.advanced.security.pkce.label', 'Require PKCE')}
              </Typography>
            }
          />
          <Typography variant="caption" color="text.secondary" sx={{display: 'block', ml: '52px'}}>
            {getPkceCaption()}
          </Typography>
        </Box>

        <Box>
          <FormControlLabel
            control={
              <Switch
                checked={
                  flags.isParDisabledByGrants ? false : (oauth2Config.requirePushedAuthorizationRequests ?? false)
                }
                disabled={!isEditable || flags.isParDisabledByGrants}
                onChange={(e) => onOAuth2ConfigChange?.({requirePushedAuthorizationRequests: e.target.checked})}
                inputProps={{
                  'aria-label': t('agents:edit.advanced.security.par.label', 'Require Pushed Authorization Requests'),
                }}
              />
            }
            label={
              <Typography variant="subtitle2">
                {t('agents:edit.advanced.security.par.label', 'Require Pushed Authorization Requests')}
              </Typography>
            }
          />
          <Typography variant="caption" color="text.secondary" sx={{display: 'block', ml: '52px'}}>
            {flags.isParDisabledByGrants ? (
              <Trans
                i18nKey="agents:edit.advanced.security.par.notApplicable"
                defaults="Pushed Authorization Requests only apply to the <code>authorization_code</code> grant. Turn that on to enable this setting."
                components={codeComponents}
              />
            ) : (
              t(
                'agents:edit.advanced.security.par.hint',
                'Require this agent to push its authorization request to the PAR endpoint before redirecting a user to sign in.',
              )
            )}
          </Typography>
        </Box>
      </Box>
    </SettingsCard>
  );
}
