// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {SettingsCard} from '@thunderid/components';
import {
  CIMD_GRANT_TYPES,
  applyGrantTypesChange,
  getCimdGrantRestriction,
  getGrantTypeLabel,
  isGrantLockedByIdJag,
  isIdJagEnabled,
  type OAuth2Config,
} from '@thunderid/configure-applications';
import {
  Checkbox,
  Chip,
  FormControl,
  FormLabel,
  ListItemText,
  MenuItem,
  Select,
  Stack,
  Typography,
} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import applyAgentPkceRule from '../../../utils/applyAgentPkceRule';

/**
 * Props for the {@link TrustedDomainsClientSection} component.
 */
interface TrustedDomainsClientSectionProps {
  /** The agent's OAuth configuration, which admits clients from its trusted domains. */
  oauth2Config: OAuth2Config;
  /** Invoked with the OAuth configuration updates a grant change makes. */
  onOAuth2ConfigChange: (updates: Partial<OAuth2Config>) => void;
  /** Whether the agent can't be edited. */
  disabled?: boolean;
}

/**
 * The client settings of an agent that admits any client from its trusted domains. Each client's own metadata
 * document sets its redirect URIs and client authentication, so only the grant types are chosen here: an upper
 * bound that each client narrows to the grants its document declares. Token exchange and JWT bearer reach only a
 * client whose document uses private key JWT authentication, and token exchange follows identity assertions.
 */
export default function TrustedDomainsClientSection({
  oauth2Config,
  onOAuth2ConfigChange,
  disabled = false,
}: TrustedDomainsClientSectionProps): JSX.Element {
  const {t} = useTranslation();
  const grantTypes = oauth2Config.grantTypes ?? [];
  const idJagEnabled = isIdJagEnabled(oauth2Config);

  return (
    <SettingsCard title={t('agents:edit.advanced.trustedDomains.title', 'Client settings')}>
      <Stack spacing={2}>
        <Typography variant="body2" color="text.secondary">
          {t(
            'agents:edit.advanced.trustedDomains.description',
            "Each client's own metadata document sets its redirect URIs and client authentication. Manage the trusted domains on the Credentials tab.",
          )}
        </Typography>
        <FormControl fullWidth size="small">
          <FormLabel htmlFor="agent-domain-grant-types">
            {t('applications:edit.advanced.labels.grantTypes', 'Grant types')}
          </FormLabel>
          <Select
            id="agent-domain-grant-types"
            multiple
            displayEmpty
            disabled={disabled}
            value={grantTypes.filter((grant) => CIMD_GRANT_TYPES.includes(grant))}
            onChange={(e) =>
              onOAuth2ConfigChange(applyAgentPkceRule(applyGrantTypesChange(oauth2Config, e.target.value as string[])))
            }
            renderValue={(selected) => (
              <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap>
                {selected.map((grant) => (
                  <Chip key={grant} label={getGrantTypeLabel(grant, t)} size="small" />
                ))}
              </Stack>
            )}
          >
            {CIMD_GRANT_TYPES.map((grant) => {
              const restriction = getCimdGrantRestriction(grant, oauth2Config);
              const lockedByIdJag = isGrantLockedByIdJag(grant, grantTypes, idJagEnabled);
              let reason: string | undefined;
              if (lockedByIdJag) {
                reason = t(
                  'applications:edit.advanced.grantTypes.requiredByIdJag',
                  'Required by identity assertions (ID-JAG)',
                );
              } else if (restriction) {
                reason = t(...restriction);
              }
              return (
                <MenuItem key={grant} value={grant} disabled={lockedByIdJag || restriction !== undefined}>
                  <Checkbox checked={grantTypes.includes(grant)} size="small" />
                  <ListItemText primary={getGrantTypeLabel(grant, t)} secondary={reason} />
                </MenuItem>
              );
            })}
          </Select>
          <Typography variant="caption" color="text.secondary" sx={{mt: 0.5}}>
            {t(
              'agents:edit.advanced.trustedDomains.grantTypesHint',
              'Each client gets the grant types both this agent allows and its own metadata document declares. Token exchange and JWT bearer apply only to clients whose document uses private key JWT authentication, and token exchange comes with identity assertions (ID-JAG).',
            )}
          </Typography>
        </FormControl>
      </Stack>
    </SettingsCard>
  );
}
