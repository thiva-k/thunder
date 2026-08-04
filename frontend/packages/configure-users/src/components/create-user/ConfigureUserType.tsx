// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Stack, Typography, FormControl, FormLabel, Select, MenuItem} from '@wso2/oxygen-ui';
import {useEffect} from 'react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import type {SchemaInterface} from '../../models/users';

/**
 * Props for the {@link ConfigureUserType} component.
 *
 * @public
 */
export interface ConfigureUserTypeProps {
  schemas: SchemaInterface[];
  selectedSchema: SchemaInterface | null;
  onSchemaChange: (schema: SchemaInterface | null) => void;
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Step 1 of the user creation wizard: select a user type (schema).
 *
 * @public
 */
export default function ConfigureUserType({
  schemas,
  selectedSchema,
  onSchemaChange,
  onReadyChange = undefined,
}: ConfigureUserTypeProps): JSX.Element {
  const {t} = useTranslation();

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(selectedSchema !== null);
    }
  }, [selectedSchema, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-user-type">
      <Typography variant="h1" gutterBottom>
        {t('users:createWizard.selectUserType.title')}
      </Typography>
      <Typography variant="body1" color="text.secondary">
        {t('users:createWizard.selectUserType.subtitle')}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel htmlFor="user-type-select">{t('users:createWizard.selectUserType.fieldLabel')}</FormLabel>
        <Select
          id="user-type-select"
          value={selectedSchema?.id ?? ''}
          onChange={(e) => {
            const schema = schemas.find((s) => s.id === e.target.value);
            onSchemaChange(schema ?? null);
          }}
          displayEmpty
          data-testid="user-type-select"
        >
          <MenuItem value="" disabled>
            <em>{t('users:createWizard.selectUserType.placeholder')}</em>
          </MenuItem>
          {schemas.map((schema) => (
            <MenuItem key={schema.id} value={schema.id}>
              {schema.name}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    </Stack>
  );
}
