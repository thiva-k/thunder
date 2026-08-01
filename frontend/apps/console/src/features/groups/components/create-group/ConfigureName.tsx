/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {NameSuggestion, OrganizationUnitSummaryChip} from '@thunderid/components';
import {OrganizationUnitTreeConstants} from '@thunderid/configure-organization-units';
import {Typography, Stack, TextField, FormControl, FormLabel} from '@wso2/oxygen-ui';
import type {ChangeEvent, JSX} from 'react';
import {useEffect} from 'react';
import {useTranslation} from 'react-i18next';

/**
 * Props for the {@link ConfigureName} component.
 *
 * @public
 */
export interface ConfigureNameProps {
  name: string;
  onNameChange: (name: string) => void;
  onReadyChange?: (isReady: boolean) => void;

  /**
   * Whether the wizard's organization unit was picked on a dedicated earlier step (only then is
   * the summary chip shown).
   */
  hasMultipleOUs?: boolean;

  /**
   * The resolved organization unit's display name, shown in the summary chip.
   */
  organizationUnitName?: string;

  /**
   * The resolved organization unit's logo, shown in the summary chip.
   */
  organizationUnitLogoUrl?: string;

  /**
   * Whether the organization unit is still being resolved.
   */
  isOrganizationUnitLoading?: boolean;

  /**
   * Invoked when the chip's "Change" link is clicked, returning to the organization unit step.
   */
  onChangeOu?: () => void;
}

/**
 * Step 1 of the group creation wizard: configure the group name.
 *
 * @public
 */
export default function ConfigureName({
  name,
  onNameChange,
  onReadyChange = undefined,
  hasMultipleOUs = false,
  organizationUnitName = undefined,
  organizationUnitLogoUrl = undefined,
  isOrganizationUnitLoading = false,
  onChangeOu = undefined,
}: ConfigureNameProps): JSX.Element {
  const {t} = useTranslation();

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(name.trim().length > 0);
    }
  }, [name, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-name">
      <Typography variant="h1" gutterBottom>
        {t('groups:createWizard.name.title')}
      </Typography>

      {hasMultipleOUs && onChangeOu && (
        <OrganizationUnitSummaryChip
          logoUrl={organizationUnitLogoUrl}
          icon={OrganizationUnitTreeConstants.DEFAULT_AVATAR}
          label={t('groups:createWizard.organizationUnit.fieldLabel', 'Organization Unit')}
          value={isOrganizationUnitLoading ? t('common:status.loading', 'Loading...') : organizationUnitName}
          onChange={onChangeOu}
        />
      )}

      <FormControl fullWidth required>
        <FormLabel htmlFor="group-name-input">{t('groups:create.form.name.label')}</FormLabel>
        <TextField
          fullWidth
          id="group-name-input"
          value={name}
          onChange={(e: ChangeEvent<HTMLInputElement>): void => onNameChange(e.target.value)}
          placeholder={t('groups:create.form.name.placeholder')}
          inputProps={{
            'data-testid': 'group-name-input',
          }}
        />

        <NameSuggestion onSelect={onNameChange} />
      </FormControl>
    </Stack>
  );
}
