// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {OrganizationUnitTreePicker} from '@thunderid/configure-organization-units';
import {Stack, Typography, FormControl, FormLabel} from '@wso2/oxygen-ui';
import {useEffect} from 'react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';

export interface ConfigureOrganizationUnitProps {
  rootOuId: string;
  selectedOuId: string;
  onOuIdChange: (ouId: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

export default function ConfigureOrganizationUnit({
  rootOuId,
  selectedOuId,
  onOuIdChange,
  onReadyChange = undefined,
}: ConfigureOrganizationUnitProps): JSX.Element {
  const {t} = useTranslation();

  useEffect(() => {
    if (!selectedOuId) {
      onOuIdChange(rootOuId);
    }
  }, [selectedOuId, rootOuId, onOuIdChange]);

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(selectedOuId.length > 0);
    }
  }, [selectedOuId, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-organization-unit">
      <Typography variant="h1" gutterBottom>
        {t('users:createWizard.selectOrganizationUnit.title')}
      </Typography>
      <Typography variant="body1" color="text.secondary">
        {t('users:createWizard.selectOrganizationUnit.subtitle')}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel>{t('users:createWizard.selectOrganizationUnit.fieldLabel')}</FormLabel>
        <OrganizationUnitTreePicker
          id="user-create-ou-picker"
          rootOuId={rootOuId}
          value={selectedOuId}
          onChange={onOuIdChange}
          maxHeight={500}
        />
      </FormControl>
    </Stack>
  );
}
