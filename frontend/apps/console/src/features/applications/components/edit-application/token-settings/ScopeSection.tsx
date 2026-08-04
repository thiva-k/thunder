// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {SettingsCard} from '@thunderid/components';
import type {ScopeClaims} from '@thunderid/configure-applications';
import {Box, Stack, Typography, Divider} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import ScopeMapper from './ScopeMapper';
import ScopeSelector from './ScopeSelector';

/**
 * Props for the {@link ScopeSection} component.
 */
interface ScopeSectionProps {
  /**
   * Current list of active OAuth2 scopes.
   */
  scopes: string[];
  /**
   * Current scope → attributes mapping from the top-level scope_claims field.
   */
  scopeClaims: ScopeClaims;
  /**
   * All available user attributes derived from user types.
   */
  userAttributes: string[];
  /**
   * Loading state for user attributes fetch.
   */
  isLoadingUserAttributes: boolean;
  /**
   * Callback fired when the scopes list changes.
   */
  onScopesChange: (scopes: string[]) => void;
  /**
   * Callback fired when the scope → attributes mapping changes.
   */
  onScopeClaimsChange: (scopeClaims: ScopeClaims) => void;
  /**
   * Singular noun used to refer to the entity in user-visible copy (default: 'application').
   */
  entityLabel?: string;
  /**
   * Whether inputs should be disabled (e.g. read-only resource).
   */
  disabled?: boolean;
}

/**
 * Settings card for managing OAuth2 scopes and their attribute mappings.
 *
 * Contains two sub-components:
 * - **{@link ScopeSelector}** — manage the active scope list (add/remove scopes).
 * - **{@link ScopeMapper}** — map user attributes to individual scopes via a
 *   two-panel interface (scope list on the left, attribute picker on the right).
 *
 * @param props - Component props
 * @returns Scopes and attribute mapping configuration within a SettingsCard
 */
export default function ScopeSection({
  scopes,
  scopeClaims,
  userAttributes,
  isLoadingUserAttributes,
  onScopesChange,
  onScopeClaimsChange,
  entityLabel = 'application',
  disabled = false,
}: ScopeSectionProps) {
  const {t} = useTranslation();

  const handleScopesChange = (newScopes: string[]) => {
    // Clean up attribute mappings for removed scopes
    const removedScopes = scopes.filter((s) => !newScopes.includes(s));
    if (removedScopes.length > 0) {
      const updatedClaims = {...scopeClaims};
      removedScopes.forEach((s) => delete updatedClaims[s]);
      onScopeClaimsChange(updatedClaims);
    }
    onScopesChange(newScopes);
  };

  return (
    <SettingsCard
      title={t('applications:edit.token.scopes_card.title', 'Scopes & User Attribute Mappings')}
      description={t(
        'applications:edit.token.scopes_card.description',
        'Configure the OAuth2 scopes and the user attributes exposed for each scope',
      )}
    >
      <Stack spacing={3}>
        {/* ── Scopes ─────────────────────────────────────────────────── */}
        <ScopeSelector
          scopes={scopes}
          onScopesChange={handleScopesChange}
          entityLabel={entityLabel}
          disabled={disabled}
        />

        <Divider />

        {/* ── Scope Attribute Mapper ───────────────────────────────── */}
        <Box>
          <Typography variant="subtitle2" gutterBottom>
            {t('applications:edit.token.scope_mapper.title', 'Attribute Mapping')}
          </Typography>
          <Typography variant="body2" color="text.disabled" sx={{mb: 2}}>
            {t(
              'applications:edit.token.scope_mapper.hint',
              'Select a scope to configure which user attributes are exposed when it is requested.',
            )}
          </Typography>

          <ScopeMapper
            scopes={scopes}
            scopeClaims={scopeClaims}
            userAttributes={userAttributes}
            isLoadingUserAttributes={isLoadingUserAttributes}
            onScopeClaimsChange={onScopeClaimsChange}
            disabled={disabled}
          />
        </Box>
      </Stack>
    </SettingsCard>
  );
}
