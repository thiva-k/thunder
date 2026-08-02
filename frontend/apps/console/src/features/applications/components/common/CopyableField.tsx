// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useCopyToClipboard} from '@thunderid/hooks';
import {FormControl, FormLabel, IconButton, InputAdornment, TextField} from '@wso2/oxygen-ui';
import {Check, Copy} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';

/**
 * Props for the {@link CopyableField} component.
 *
 * @public
 */
export interface CopyableFieldProps {
  /**
   * The `id`/`htmlFor` used to associate the field's label with its input
   */
  id: string;

  /**
   * The field's label
   */
  label: string;

  /**
   * The read-only value displayed and copied
   */
  value: string;

  /**
   * The `aria-label` for the copy button
   */
  copyAriaLabel: string;
}

/**
 * Read-only monospace field with a copy-to-clipboard affordance, matching
 * `QuickCopySection`'s field pattern. Shared by the mcp-client template's create-flow
 * Connect completion screen and edit-page Connect tab for the Application ID, Client ID,
 * client secret, and discovery endpoint fields.
 *
 * @param props - The component props
 * @param props.id - The `id`/`htmlFor` used to associate the field's label with its input
 * @param props.label - The field's label
 * @param props.value - The read-only value displayed and copied
 * @param props.copyAriaLabel - The `aria-label` for the copy button
 *
 * @returns JSX element displaying a read-only copyable field
 *
 * @example
 * ```tsx
 * <CopyableField id="mcp-connect-client-id" label="Client ID" value={clientId} copyAriaLabel="Copy Client ID" />
 * ```
 *
 * @public
 */
export default function CopyableField({id, label, value, copyAriaLabel}: CopyableFieldProps): JSX.Element {
  const {copied, copy} = useCopyToClipboard({resetDelay: 2000});

  return (
    <FormControl fullWidth>
      <FormLabel htmlFor={id}>{label}</FormLabel>
      <TextField
        fullWidth
        id={id}
        value={value}
        InputProps={{
          readOnly: true,
          endAdornment: (
            <InputAdornment position="end">
              <IconButton
                aria-label={copyAriaLabel}
                onClick={() => {
                  copy(value).catch(() => {
                    // Error already handled in copy
                  });
                }}
                edge="end"
                size="small"
              >
                {copied ? <Check size={16} /> : <Copy size={16} />}
              </IconButton>
            </InputAdornment>
          ),
        }}
        sx={{'& input': {fontFamily: 'monospace', fontSize: '0.875rem'}}}
      />
    </FormControl>
  );
}
