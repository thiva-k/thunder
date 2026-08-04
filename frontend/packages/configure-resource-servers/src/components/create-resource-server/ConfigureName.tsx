// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {NameSuggestion} from '@thunderid/components';
import {FormControl, FormLabel, Stack, TextField, Typography} from '@wso2/oxygen-ui';
import {useEffect, type ChangeEvent, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import type {ResourceServerType} from '../../models/resource-server';

interface ConfigureNameProps {
  name: string;
  identifier: string;
  /** The resource server type selected in the previous step, used to tailor copy for MCP servers. */
  selectedType?: ResourceServerType;
  onNameChange: (name: string) => void;
  onIdentifierChange: (identifier: string) => void;
  onReadyChange?: (isReady: boolean) => void;
}

export default function ConfigureName({
  name,
  identifier,
  selectedType = undefined,
  onNameChange,
  onIdentifierChange,
  onReadyChange = undefined,
}: ConfigureNameProps): JSX.Element {
  const {t} = useTranslation();

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(name.trim().length > 0 && identifier.trim().length > 0);
    }
  }, [name, identifier, onReadyChange]);

  const handleNameChange = (e: ChangeEvent<HTMLInputElement>): void => {
    onNameChange(e.target.value);
  };

  const handleIdentifierChange = (e: ChangeEvent<HTMLInputElement>): void => {
    onIdentifierChange(e.target.value);
  };

  return (
    <Stack direction="column" spacing={4}>
      <Typography variant="h1" gutterBottom>
        {selectedType === 'MCP'
          ? t('resourceServers:create.name.titleMcp', "Let's collect some details about your MCP server")
          : t('resourceServers:create.name.title', "Let's collect some details about your resource server")}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel htmlFor="resource-server-name-input">
          {selectedType === 'MCP'
            ? t('resourceServers:create.name.nameLabelMcp', 'MCP Server Name')
            : t('resourceServers:create.name.nameLabel', 'Resource Server Name')}
        </FormLabel>
        <TextField
          id="resource-server-name-input"
          fullWidth
          value={name}
          onChange={handleNameChange}
          placeholder={t('resourceServers:create.name.namePlaceholder', 'e.g. Payments API')}
        />

        <NameSuggestion onSelect={onNameChange} />
      </FormControl>

      <FormControl fullWidth required>
        <FormLabel htmlFor="resource-server-identifier-input">
          {t('resourceServers:create.name.identifierLabel', 'Identifier')}
        </FormLabel>
        <TextField
          id="resource-server-identifier-input"
          fullWidth
          value={identifier}
          onChange={handleIdentifierChange}
          placeholder={
            selectedType === 'MCP'
              ? t('resourceServers:create.name.identifierPlaceholderMcp', 'https://mcp.example.com')
              : t('resourceServers:create.name.identifierPlaceholder', 'https://api.example.com')
          }
          helperText={
            selectedType === 'MCP'
              ? t(
                  'resourceServers:create.name.identifierHintMcp',
                  'A unique identifier for this MCP server. When set as an absolute URI, it becomes the token audience for RFC 8707 resource indicators.',
                )
              : t(
                  'resourceServers:create.name.identifierHint',
                  'A unique identifier for this resource server. When set as an absolute URI, it becomes the token audience for RFC 8707 resource indicators.',
                )
          }
        />
      </FormControl>
    </Stack>
  );
}
