// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {OrganizationUnitTreePicker} from '@thunderid/configure-organization-units';
import {FormControl, FormLabel, Stack, Typography} from '@wso2/oxygen-ui';
import {useEffect, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import type {ResourceServerType} from '../../models/resource-server';

interface ConfigureOrgUnitProps {
  selectedOuId: string;
  /** The resource server type selected in the previous step, used to tailor copy for MCP servers. */
  selectedType?: ResourceServerType;
  onOuIdChange: (ouId: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

export default function ConfigureOrgUnit({
  selectedOuId,
  selectedType = undefined,
  onOuIdChange,
  onReadyChange = undefined,
}: ConfigureOrgUnitProps): JSX.Element {
  const {t} = useTranslation();

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(selectedOuId.length > 0);
    }
  }, [selectedOuId, onReadyChange]);

  return (
    <Stack direction="column" spacing={4}>
      <Typography variant="h1" gutterBottom>
        {t('resourceServers:create.orgUnit.title', 'Choose an organization unit')}
      </Typography>
      <Typography variant="body1" color="text.secondary">
        {selectedType === 'MCP'
          ? t(
              'resourceServers:create.orgUnit.subtitleMcp',
              'Select which organization unit this MCP server belongs to.',
            )
          : t(
              'resourceServers:create.orgUnit.subtitle',
              'Select which organization unit this resource server belongs to.',
            )}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel>{t('resourceServers:create.orgUnit.fieldLabel', 'Organization Unit')}</FormLabel>
        <OrganizationUnitTreePicker
          id="resource-server-create-ou-picker"
          value={selectedOuId}
          onChange={onOuIdChange}
          maxHeight={400}
        />
      </FormControl>
    </Stack>
  );
}
