// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {SettingsCard} from '@thunderid/components';
import {Stack, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import ClientIdDomainsField from './ClientIdDomainsField';
import {validateClientIdDomains} from '../../utils/clientIdDomains';

/**
 * Props for the {@link ClientIdDomainsSection} component.
 *
 * @public
 */
export interface ClientIdDomainsSectionProps {
  /** The domains currently set, edited or stored. */
  domains: string[];
  /** Invoked with the new domains, to join the page's pending edits. */
  onDomainsChange: (domains: string[]) => void;
  /** What the clients sign in as, which words the description. */
  subject: 'application' | 'agent';
  /** Whether the domains can't be changed. */
  isReadOnly?: boolean;
}

/**
 * The edit-page section of an application or agent that admits clients by the domain of their client ID
 * URL. The domains are editable and saved with the page. What each client's metadata document supplies is
 * listed, since those settings are absent from this application or agent.
 *
 * @public
 */
export default function ClientIdDomainsSection({
  domains,
  onDomainsChange,
  subject,
  isReadOnly = false,
}: ClientIdDomainsSectionProps): JSX.Element {
  const {t} = useTranslation();
  const error = validateClientIdDomains(domains);

  return (
    <SettingsCard
      title={t('applications:cimd.domains.edit.title', 'Clients from trusted domains')}
      description={
        subject === 'agent'
          ? t(
              'applications:cimd.domains.edit.description.agent',
              "Any client that publishes a Client ID Metadata Document at a client ID URL on these domains signs in as this agent, with this agent's settings. Nothing is registered for each client.",
            )
          : t(
              'applications:cimd.domains.edit.description.application',
              "Any client that publishes a Client ID Metadata Document at a client ID URL on these domains signs in through this application, with this application's settings. Nothing is registered for each client.",
            )
      }
    >
      <Stack spacing={2} data-testid="cimd-client-id-domains-section">
        <ClientIdDomainsField
          value={domains}
          onChange={onDomainsChange}
          error={error ? t(error.key, error.defaultValue) : undefined}
          disabled={isReadOnly}
        />
        <Typography variant="body2" color="text.secondary">
          {t(
            'applications:cimd.domains.edit.fromDocument',
            "Each client's own document sets its client ID, redirect URIs, sign-in method and keys, and the name and logo users see when they sign in. The document is read when the client signs in and kept for as long as its cache headers allow.",
          )}
        </Typography>
      </Stack>
    </SettingsCard>
  );
}
