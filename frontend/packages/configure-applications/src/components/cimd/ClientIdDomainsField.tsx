// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Autocomplete, Chip, TextField} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useState} from 'react';
import {useTranslation} from 'react-i18next';
import {isValidClientIdDomain, parseClientIdDomains} from '../../utils/clientIdDomains';

/**
 * Props for the {@link ClientIdDomainsField} component.
 *
 * @public
 */
export interface ClientIdDomainsFieldProps {
  /** The domains entered so far. */
  value: string[];
  /** Invoked with the normalized domains whenever they change. */
  onChange: (domains: string[]) => void;
  /** The validation message to show in place of the hint. */
  error?: string;
  /** Whether the field is read-only. */
  disabled?: boolean;
  /** The id of the input, also used to derive its test id. */
  id?: string;
}

/**
 * The domains whose clients may sign in by publishing a Client ID Metadata Document, entered as chips. A
 * space, comma, Enter or leaving the field ends a domain, so a pasted list becomes chips at once, and a
 * pasted URL keeps only its host. An invalid domain shows as an outlined error chip.
 *
 * @public
 */
export default function ClientIdDomainsField({
  value,
  onChange,
  error = undefined,
  disabled = false,
  id = 'cimd-client-id-domains',
}: ClientIdDomainsFieldProps): JSX.Element {
  const {t} = useTranslation();
  const [input, setInput] = useState('');

  const commit = (values: string[]): void => onChange(parseClientIdDomains(values));

  return (
    <Autocomplete
      multiple
      freeSolo
      options={[]}
      value={value}
      inputValue={input}
      disabled={disabled}
      onChange={(_event, next) => commit(next)}
      onInputChange={(_event, next, reason) => {
        if (reason === 'input' && /[\s,]/.test(next)) {
          const parts = next.split(/[\s,]+/);
          const pending = /[\s,]$/.test(next) ? '' : (parts.pop() ?? '');
          commit([...value, ...parts]);
          setInput(pending);
          return;
        }
        setInput(next);
      }}
      onBlur={() => {
        if (input.trim()) {
          commit([...value, input]);
          setInput('');
        }
      }}
      renderValue={(values, getItemProps) =>
        values.map((domain, index) => {
          const {key, onDelete, ...itemProps} = getItemProps({index});
          const invalid = !isValidClientIdDomain(domain);
          return (
            <Chip
              key={key}
              {...itemProps}
              onDelete={onDelete}
              label={domain}
              size="small"
              color={invalid ? 'error' : 'default'}
              variant={invalid ? 'outlined' : 'filled'}
              sx={{fontFamily: 'monospace'}}
            />
          );
        })
      }
      renderInput={(params) => (
        <TextField
          {...params}
          id={id}
          label={t('applications:cimd.domains.label', 'Trusted domains')}
          required={value.length === 0}
          placeholder={
            value.length === 0 ? t('applications:cimd.domains.placeholder', 'vscode.dev, claude.ai') : undefined
          }
          error={Boolean(error)}
          helperText={
            error ??
            t(
              'applications:cimd.domains.helper',
              'Clients whose client ID URL is on one of these domains can sign in. Use *.example.com to include every subdomain of example.com.',
            )
          }
          inputProps={{...params.inputProps, 'data-testid': `${id}-input`}}
        />
      )}
    />
  );
}
