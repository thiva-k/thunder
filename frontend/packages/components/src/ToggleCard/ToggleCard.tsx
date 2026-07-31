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

import {Box, Checkbox, Typography} from '@wso2/oxygen-ui';
import type {JSX, ReactNode} from 'react';

export interface ToggleCardProps {
  /**
   * Whether the checkbox is checked.
   */
  checked: boolean;

  /**
   * Renders the checkbox in an indeterminate state (e.g. some, but not all, of an underlying
   * list is selected).
   */
  indeterminate?: boolean;

  /**
   * Invoked when the checkbox is toggled.
   */
  onChange: (checked: boolean) => void;

  /**
   * Bold title shown next to the checkbox.
   */
  title: string;

  /**
   * Secondary description shown below the title.
   */
  subtitle?: string;

  /**
   * Validation message shown below the subtitle in the error color.
   */
  error?: string;

  /**
   * Accessible label for the checkbox. Defaults to `title`.
   */
  ariaLabel?: string;

  /**
   * Wraps the row in a rounded bordered box. Set to `false` when the card is already nested
   * inside another bordered container (e.g. stacked with other toggle rows under one shared
   * border), so the border isn't doubled up.
   */
  bordered?: boolean;

  /**
   * Optional trailing content rendered at the end of the row (e.g. an expand/collapse button).
   */
  action?: ReactNode;
}

/**
 * A single checkbox presented as a titled, described row — optionally wrapped in its own
 * bordered card. Shared by any wizard step that offers a single yes/no toggle (e.g. "Allow all
 * user types to access this application", "Allow Self Registration").
 */
export default function ToggleCard({
  checked,
  indeterminate = false,
  onChange,
  title,
  subtitle = undefined,
  error = undefined,
  ariaLabel = undefined,
  bordered = true,
  action = undefined,
}: ToggleCardProps): JSX.Element {
  return (
    <Box
      sx={
        bordered ? {border: '1px solid', borderColor: 'divider', borderRadius: '10px', overflow: 'hidden'} : undefined
      }
    >
      <Box sx={{display: 'flex', alignItems: 'flex-start', gap: 1, p: 2}}>
        <Checkbox
          checked={checked}
          indeterminate={indeterminate}
          onChange={(_event, isChecked) => onChange(isChecked)}
          inputProps={{'aria-label': ariaLabel ?? title}}
          sx={{mt: -0.5}}
        />
        <Box sx={{flex: 1, minWidth: 0}}>
          <Typography variant="body1" fontWeight={600}>
            {title}
          </Typography>
          {subtitle && (
            <Typography variant="body2" color="text.secondary">
              {subtitle}
            </Typography>
          )}
          {error && (
            <Typography variant="caption" color="error" sx={{display: 'block', mt: 0.5}}>
              {error}
            </Typography>
          )}
        </Box>
        {action}
      </Box>
    </Box>
  );
}
