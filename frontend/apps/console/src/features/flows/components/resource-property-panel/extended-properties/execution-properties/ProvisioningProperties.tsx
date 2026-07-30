/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import {Checkbox, FormControlLabel, FormHelperText, FormLabel, Stack, Typography} from '@wso2/oxygen-ui';
import {useCallback, useMemo, type ReactNode} from 'react';
import {useTranslation} from 'react-i18next';
import DraftTextField from './DraftTextField';
import type {CommonResourcePropertiesPropsInterface} from './types';
import {clampToInteger} from './utils';
import type {StepData} from '@/features/flows/models/steps';

function ProvisioningProperties({resource, onChange}: CommonResourcePropertiesPropsInterface): ReactNode {
  const {t} = useTranslation();

  const properties = useMemo(() => {
    const stepData = resource?.data as StepData | undefined;
    return stepData?.properties ?? {};
  }, [resource]);

  const handleBooleanPropertyChange = useCallback(
    (propertyName: string, value: boolean): void => {
      onChange(`data.properties.${propertyName}`, value, resource);
    },
    [resource, onChange],
  );

  const handleStringPropertyChange = useCallback(
    (propertyName: string, value: string): void => {
      onChange(`data.properties.${propertyName}`, value, resource);
    },
    [resource, onChange],
  );

  const handleNumberPropertyChange = useCallback(
    (propertyName: string, value: string): void => {
      onChange(`data.properties.${propertyName}`, Number(value), resource);
    },
    [resource, onChange],
  );

  const maxPerPromptValue = useMemo(() => {
    const rawValue = properties.maxPerPrompt;
    const numericValue = typeof rawValue === 'number' ? rawValue : Number(rawValue ?? 0);

    return Number.isFinite(numericValue) ? numericValue : 0;
  }, [properties.maxPerPrompt]);

  return (
    <Stack gap={2}>
      <Typography variant="body2" color="text.secondary">
        {t('flows:core.executions.provisioning.description')}
      </Typography>

      <FormControlLabel
        control={
          <Checkbox
            checked={!!properties.allowCrossOUProvisioning}
            onChange={(e) => handleBooleanPropertyChange('allowCrossOUProvisioning', e.target.checked)}
            size="small"
          />
        }
        label={t('flows:core.executions.federation.allowCrossOUProvisioning.label')}
      />
      <FormHelperText>{t('flows:core.executions.federation.allowCrossOUProvisioning.hint')}</FormHelperText>

      <FormControlLabel
        control={
          <Checkbox
            checked={!!properties.includeOptional}
            onChange={(e) => handleBooleanPropertyChange('includeOptional', e.target.checked)}
            size="small"
          />
        }
        label={t('flows:core.executions.provisioning.includeOptional.label')}
      />
      <FormHelperText>{t('flows:core.executions.provisioning.includeOptional.hint')}</FormHelperText>

      <FormControlLabel
        control={
          <Checkbox
            checked={!!properties.includeOptionalCredentials}
            onChange={(e) => handleBooleanPropertyChange('includeOptionalCredentials', e.target.checked)}
            size="small"
          />
        }
        label={t('flows:core.executions.provisioning.includeOptionalCredentials.label')}
      />
      <FormHelperText>{t('flows:core.executions.provisioning.includeOptionalCredentials.hint')}</FormHelperText>

      <div>
        <FormLabel htmlFor="max-per-prompt">{t('flows:core.executions.provisioning.maxPerPrompt.label')}</FormLabel>
        <DraftTextField
          id="max-per-prompt"
          type="number"
          value={String(maxPerPromptValue)}
          onCommit={(value) => handleNumberPropertyChange('maxPerPrompt', value)}
          normalize={(raw) => clampToInteger(raw, 0)}
          placeholder={t('flows:core.executions.provisioning.maxPerPrompt.placeholder')}
          fullWidth
          size="small"
          slotProps={{
            htmlInput: {
              min: 0,
            },
          }}
        />
        <FormHelperText>{t('flows:core.executions.provisioning.maxPerPrompt.hint')}</FormHelperText>
      </div>

      <div>
        <FormLabel htmlFor="assign-group">{t('flows:core.executions.provisioning.assignGroup.label')}</FormLabel>
        <DraftTextField
          id="assign-group"
          value={(properties.assignGroup as string) || ''}
          onCommit={(value) => handleStringPropertyChange('assignGroup', value)}
          placeholder={t('flows:core.executions.provisioning.assignGroup.placeholder')}
          fullWidth
          size="small"
        />
      </div>

      <div>
        <FormLabel htmlFor="assign-role">{t('flows:core.executions.provisioning.assignRole.label')}</FormLabel>
        <DraftTextField
          id="assign-role"
          value={(properties.assignRole as string) || ''}
          onCommit={(value) => handleStringPropertyChange('assignRole', value)}
          placeholder={t('flows:core.executions.provisioning.assignRole.placeholder')}
          fullWidth
          size="small"
        />
      </div>
    </Stack>
  );
}

export default ProvisioningProperties;
