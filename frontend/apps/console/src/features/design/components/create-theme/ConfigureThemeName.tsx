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
import {type ChangeEvent, type JSX, useEffect} from 'react';
import {useTranslation} from 'react-i18next';

export interface ConfigureThemeNameProps {
  themeName: string;
  onThemeNameChange: (name: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

export default function ConfigureThemeName({
  themeName,
  onThemeNameChange,
  onReadyChange = () => null,
}: ConfigureThemeNameProps): JSX.Element {
  const {t} = useTranslation('design');

  useEffect(() => {
    onReadyChange?.(themeName.trim().length > 0);
  }, [themeName, onReadyChange]);

  return (
    <Stack direction="column" spacing={4}>
      <Typography variant="h1">
        {t('themes.forms.configure_name.title', "Let's collect some details about your theme")}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel htmlFor="theme-name-input">{t('themes.forms.configure_name.fieldLabel', 'Theme name')}</FormLabel>
        <TextField
          fullWidth
          id="theme-name-input"
          value={themeName}
          onChange={(e: ChangeEvent<HTMLInputElement>) => onThemeNameChange(e.target.value)}
          placeholder={t('themes.forms.configure_name.placeholder', 'e.g. Solarized Light')}
        />

        <NameSuggestion onSelect={onThemeNameChange} />
      </FormControl>
    </Stack>
  );
}
