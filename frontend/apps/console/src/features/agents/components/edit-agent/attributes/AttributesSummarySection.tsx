// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {SettingsCard} from '@thunderid/components';
import {useGetAgentType, useGetAgentTypes} from '@thunderid/configure-agent-types';
import {useResolveDisplayName} from '@thunderid/hooks';
import {Box, Chip, CircularProgress, Stack, Typography} from '@wso2/oxygen-ui';
import {Tag} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import type {Agent} from '../../../models/agent';
import EmptyState from '../shared/EmptyState';

interface AttributesSummarySectionProps {
  agent: Agent;
  /**
   * 'settings' (default) wraps the content in a SettingsCard, matching the Attributes tab's
   * own fallback for read-only agents. 'bare' renders just the attribute values with no card
   * chrome or title, as a compact chip list — for embedding inside another card (e.g. the
   * Overview tab), so it reads as a quick preview rather than a second settings section.
   */
  variant?: 'settings' | 'bare';
}

const formatValue = (value: unknown): string => {
  if (value === null || value === undefined) return '-';
  if (Array.isArray(value)) return value.join(', ');
  if (typeof value === 'object') return JSON.stringify(value);
  if (typeof value === 'string' || typeof value === 'number') return String(value);
  return '-';
};

/**
 * Read-only preview of this agent's attribute values, with no edit affordance. The Attributes
 * tab is where values are actually edited.
 */
export default function AttributesSummarySection({
  agent,
  variant = 'settings',
}: AttributesSummarySectionProps): JSX.Element {
  const {t} = useTranslation();
  const {resolveDisplayName} = useResolveDisplayName({handlers: {t}});

  const {data: agentTypesData} = useGetAgentTypes();
  const matchedSchema = agentTypesData?.types?.find((s) => s.name === agent.type);
  const {data: schemaDetails, isLoading} = useGetAgentType(matchedSchema?.id);

  const attributes = agent.attributes ?? {};

  const labelFor = (key: string): string => {
    const fieldDef = schemaDetails?.schema?.[key];
    if (fieldDef?.displayName) {
      return resolveDisplayName(fieldDef.displayName) || key;
    }
    return key;
  };

  const formatChipValue = (value: unknown): string =>
    typeof value === 'boolean'
      ? value
        ? t('common:actions.yes', 'Yes')
        : t('common:actions.no', 'No')
      : formatValue(value);

  const content = isLoading ? (
    <Box sx={{display: 'flex', justifyContent: 'center', py: variant === 'bare' ? 2 : 4}}>
      <CircularProgress size={32} />
    </Box>
  ) : Object.keys(attributes).length > 0 ? (
    variant === 'bare' ? (
      <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
        {Object.entries(attributes).map(([key, value]) => (
          <Chip
            key={key}
            label={`${labelFor(key)}: ${formatChipValue(value)}`}
            size="small"
            variant="outlined"
            sx={{fontWeight: 500}}
          />
        ))}
      </Box>
    ) : (
      <Stack spacing={2}>
        {Object.entries(attributes).map(([key, value]) => (
          <Box key={key}>
            <Typography variant="caption" color="text.secondary">
              {labelFor(key)}
            </Typography>
            {typeof value === 'boolean' ? (
              <Box>
                <Chip
                  label={value ? t('common:actions.yes', 'Yes') : t('common:actions.no', 'No')}
                  size="small"
                  color={value ? 'success' : 'default'}
                  variant="outlined"
                />
              </Box>
            ) : (
              <Typography variant="body1">{formatValue(value)}</Typography>
            )}
          </Box>
        ))}
      </Stack>
    )
  ) : (
    <EmptyState icon={<Tag size={28} />} message={t('agents:edit.attributes.empty', 'No attributes available.')} />
  );

  if (variant === 'bare') {
    return content;
  }

  return (
    <SettingsCard
      title={t('agents:edit.general.sections.attributes.title', 'Attributes')}
      description={t(
        'agents:edit.general.sections.attributes.description',
        "A preview of this agent's attribute values. Manage them from the Attributes tab.",
      )}
    >
      {content}
    </SettingsCard>
  );
}
