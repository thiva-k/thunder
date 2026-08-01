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

import {NameSuggestion} from '@thunderid/components';
import {FormControl, FormLabel, Stack, TextField, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';

interface ConnectionNameStepProps {
  name: string;
  onNameChange: (name: string) => void;
  /** External error, e.g. a duplicate-name 409 bounced back from a later step. */
  nameError?: string | null;
}

/**
 * The name step of the "Add custom connection" wizard: a text field for the connection name plus
 * random name suggestions, mirroring the application creation wizard's name step.
 */
export default function ConnectionNameStep({
  name,
  onNameChange,
  nameError = null,
}: ConnectionNameStepProps): JSX.Element {
  const {t} = useTranslation('connections');

  return (
    <Stack direction="column" spacing={4} data-testid="connection-name-step">
      <Typography variant="h1" gutterBottom>
        {t('wizard.name.title', "Let's collect some details about your connection")}
      </Typography>

      <FormControl fullWidth required error={Boolean(nameError)}>
        <FormLabel htmlFor="connection-name-input">{t('wizard.name.fieldLabel', 'Connection name')}</FormLabel>
        <TextField
          fullWidth
          id="connection-name-input"
          value={name}
          error={Boolean(nameError)}
          helperText={nameError ?? undefined}
          onChange={(e) => onNameChange(e.target.value)}
          placeholder={t('wizard.name.placeholder', 'Enter your connection name')}
          inputProps={{'data-testid': 'connection-name-input'}}
        />

        <NameSuggestion onSelect={onNameChange} />
      </FormControl>
    </Stack>
  );
}
