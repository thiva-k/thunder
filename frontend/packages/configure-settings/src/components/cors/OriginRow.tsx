// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {IconButton, Stack, TextField, Tooltip} from '@wso2/oxygen-ui';
import {Trash} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';

interface OriginRowProps {
  value: string;
  error?: string;
  placeholder: string;
  removeLabel: string;
  onChange: (value: string) => void;
  onBlur: () => void;
  onRemove: () => void;
}

/** A single editable allowed-origin row: a validated text field plus a remove action. */
export default function OriginRow({
  value,
  error = undefined,
  placeholder,
  removeLabel,
  onChange,
  onBlur,
  onRemove,
}: OriginRowProps): JSX.Element {
  return (
    <Stack direction="row" spacing={1} alignItems="flex-start">
      <TextField
        fullWidth
        size="small"
        value={value}
        placeholder={placeholder}
        error={Boolean(error)}
        helperText={error}
        onChange={(event) => onChange(event.target.value)}
        onBlur={onBlur}
        sx={{flex: 1}}
      />
      <Tooltip title={removeLabel}>
        <IconButton aria-label={removeLabel} color="error" onClick={onRemove} sx={{mt: 0.5}}>
          <Trash size={18} />
        </IconButton>
      </Tooltip>
    </Stack>
  );
}
