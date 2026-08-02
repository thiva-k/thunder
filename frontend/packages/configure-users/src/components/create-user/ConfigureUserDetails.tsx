// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useResolveDisplayName} from '@thunderid/hooks';
import {Box, Stack, Typography} from '@wso2/oxygen-ui';
import {useEffect} from 'react';
import type {JSX} from 'react';
import {useForm} from 'react-hook-form';
import {useTranslation} from 'react-i18next';
import type {ApiUserType} from '../../models/users';
import renderSchemaField from '../../utils/renderSchemaField';

/**
 * Props for the {@link ConfigureUserDetails} component.
 *
 * @public
 */
export interface ConfigureUserDetailsProps {
  schema: ApiUserType;
  defaultValues: Record<string, unknown>;
  onFormValuesChange: (values: Record<string, unknown>) => void;
  onReadyChange?: (isReady: boolean) => void;
}

type UserDetailsFormData = Record<string, unknown>;

/**
 * Step 2 of the user creation wizard: fill in the dynamic form fields.
 *
 * @public
 */
export default function ConfigureUserDetails({
  schema,
  defaultValues,
  onFormValuesChange,
  onReadyChange = undefined,
}: ConfigureUserDetailsProps): JSX.Element {
  const {t} = useTranslation();
  const {resolveDisplayName} = useResolveDisplayName({handlers: {t}});

  const {
    control,
    watch,
    formState: {errors, isValid},
  } = useForm<UserDetailsFormData>({
    defaultValues,
    mode: 'onChange',
  });

  useEffect(() => {
    // eslint-disable-next-line react-hooks/incompatible-library
    const subscription = watch((values) => {
      onFormValuesChange(values as Record<string, unknown>);
    });

    return () => subscription.unsubscribe();
  }, [watch, onFormValuesChange]);

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(isValid);
    }
  }, [isValid, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-user-details">
      <Typography variant="h1" gutterBottom>
        {t('users:createWizard.userDetails.title')}
      </Typography>
      <Typography variant="body1" color="text.secondary">
        {t('users:createWizard.userDetails.subtitle')}
      </Typography>

      <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
        {schema.schema &&
          Object.entries(schema.schema).map(([fieldName, fieldDef]) =>
            renderSchemaField(fieldName, fieldDef, control, errors, resolveDisplayName),
          )}
      </Box>
    </Stack>
  );
}
