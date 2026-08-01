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

import {generateRandomHumanReadableIdentifiers} from '@thunderid/utils';
import {IconButton, Stack, Tooltip, Typography} from '@wso2/oxygen-ui';
import {Shuffle} from '@wso2/oxygen-ui-icons-react';
import type {JSX, KeyboardEvent} from 'react';
import {useState} from 'react';
import {useTranslation} from 'react-i18next';

/**
 * Props for the {@link NameSuggestion} component.
 *
 * @public
 */
export interface NameSuggestionProps {
  /**
   * Callback invoked with the suggested name when it is picked.
   */
  onSelect: (name: string) => void;
}

/**
 * A single random name suggestion, reshuffled on demand, rendered inline below a name field.
 *
 * Used across the various "create X" wizards (applications, groups, roles, flows, etc.) so
 * naming a resource offers inspiration without a whole row of chips competing for attention.
 *
 * @public
 */
export default function NameSuggestion({onSelect}: NameSuggestionProps): JSX.Element {
  const {t} = useTranslation('elements');

  const [suggestion, setSuggestion] = useState<string>(() => generateRandomHumanReadableIdentifiers(1)[0]);

  const handleShuffleSuggestion = (): void => {
    setSuggestion(generateRandomHumanReadableIdentifiers(1)[0]);
  };

  const shuffleAriaLabel = t('name_suggestion.shuffle_aria_label', 'Try another suggestion');

  return (
    <Stack direction="row" alignItems="center" spacing={0.5} flexWrap="wrap">
      <Typography variant="body2" color="text.secondary">
        {t('name_suggestion.prefix', 'Need inspiration? How about')}
      </Typography>
      <Typography
        component="span"
        variant="body2"
        onClick={(): void => onSelect(suggestion)}
        role="button"
        tabIndex={0}
        onKeyDown={(e: KeyboardEvent<HTMLSpanElement>): void => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            onSelect(suggestion);
          }
        }}
        sx={{fontWeight: 600, color: 'primary.main', cursor: 'pointer'}}
      >
        {suggestion}
      </Typography>
      <Tooltip title={shuffleAriaLabel}>
        <IconButton size="small" onClick={handleShuffleSuggestion} aria-label={shuffleAriaLabel}>
          <Shuffle size={14} />
        </IconButton>
      </Tooltip>
      <Typography variant="body2" color="text.secondary">
        {t('name_suggestion.suffix', '?')}
      </Typography>
    </Stack>
  );
}
