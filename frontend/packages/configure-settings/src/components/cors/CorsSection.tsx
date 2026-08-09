// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {QueryErrorNotice, SettingsCard, UnsavedChangesBar} from '@thunderid/components';
import {getErrorMessage} from '@thunderid/utils';
import {Box, Button, Divider, Skeleton, Stack, TextField, Typography} from '@wso2/oxygen-ui';
import {InfoIcon, Plus} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useCallback} from 'react';
import {useTranslation} from 'react-i18next';
import OriginRow from './OriginRow';
import useGetCorsConfig from '../../api/useGetCorsConfig';
import useUpdateCorsConfig from '../../api/useUpdateCorsConfig';
import useAllowedOriginsDraft from '../../hooks/useAllowedOriginsDraft';
import type {AllowedOrigin} from '../../models/responses';

const ROW_ACTION_WIDTH = 40;

/** Renders an allowed origin for display: a literal string as-is, a regex entry as its pattern. */
function originText(entry: AllowedOrigin): string {
  return typeof entry === 'string' ? entry : entry.regex;
}

/** A single non-editable origin row: a muted read-only field plus a spacer that aligns with editable rows. */
function OriginDisplayRow({value}: {value: string}): JSX.Element {
  return (
    <Stack direction="row" spacing={1} alignItems="center">
      <TextField
        fullWidth
        size="small"
        value={value}
        slotProps={{input: {readOnly: true}}}
        sx={{flex: 1, opacity: 0.65}}
      />
      <Box aria-hidden sx={{width: ROW_ACTION_WIDTH, flex: 'none'}} />
    </Stack>
  );
}

export default function CorsSection(): JSX.Element {
  const {t} = useTranslation();
  const {data, isLoading, error, refetch} = useGetCorsConfig();
  const updateCors = useUpdateCorsConfig();
  const origins = useAllowedOriginsDraft(data);

  // Resolves an error through the `settings` catalog. `t` defaults to the `common` namespace, so
  // this forwards explicit `ns:` prefixes unchanged and prefixes bare keys with `settings:`, per
  // getErrorMessage's namespace-resolution contract.
  const tForErrors = useCallback(
    (key: string, options?: Record<string, unknown>): string => t(key.includes(':') ? key : `settings:${key}`, options),
    [t],
  );

  const readOnlyOrigins: AllowedOrigin[] = data?.readOnly.allowedOrigins ?? [];
  const hasReadOnlyOrigins: boolean = readOnlyOrigins.length > 0;

  // A previous save error is stale once the draft changes again.
  const clearSaveError = (): void => {
    if (updateCors.isError) {
      updateCors.reset();
    }
  };

  const handleSave = (): void => {
    if (!origins.validateAll()) {
      return;
    }
    updateCors.mutate(
      {data: origins.buildPayload()},
      {
        onSuccess: () => {
          origins.reset();
        },
      },
    );
  };

  let body: JSX.Element;
  if (isLoading) {
    body = (
      <Stack spacing={1}>
        <Skeleton variant="rounded" height={40} />
        <Skeleton variant="rounded" height={40} />
      </Stack>
    );
  } else if (error) {
    body = (
      <QueryErrorNotice
        error={error}
        t={tForErrors}
        variant="inline"
        fallbackKey="settings:cors.load.error"
        fallbackDefaultValue="Failed to load allowed origins."
        onRetry={() => void refetch()}
      />
    );
  } else {
    body = (
      <>
        <Stack spacing={1}>
          {readOnlyOrigins.map((entry, index) => (
            // eslint-disable-next-line react/no-array-index-key
            <OriginDisplayRow key={`readonly-${index}`} value={originText(entry)} />
          ))}
          {origins.draft.map((value, index) => (
            <OriginRow
              // eslint-disable-next-line react/no-array-index-key
              key={`origin-${index}`}
              value={value}
              error={origins.errors[index]}
              placeholder={t('settings:cors.originPlaceholder')}
              removeLabel={t('settings:cors.removeOrigin')}
              onChange={(next) => {
                clearSaveError();
                origins.changeRow(index, next);
              }}
              onBlur={() => origins.blurRow(index)}
              onRemove={() => {
                clearSaveError();
                origins.removeRow(index);
              }}
            />
          ))}
        </Stack>

        <Button
          variant="text"
          color="primary"
          startIcon={<Plus size={18} />}
          onClick={() => {
            clearSaveError();
            origins.addRow();
          }}
          sx={{mt: 2}}
        >
          {t('settings:cors.addOrigin')}
        </Button>

        {hasReadOnlyOrigins && (
          <>
            <Divider sx={{mt: 2, mb: 1.5}} />
            <Stack direction="row" spacing={1} alignItems="flex-start">
              <Box aria-hidden sx={{flex: 'none', display: 'inline-flex', mt: '2px', color: 'text.secondary'}}>
                <InfoIcon size={16} />
              </Box>
              <Typography variant="body2" color="text.secondary">
                {t('settings:cors.readOnlyHint')}
              </Typography>
            </Stack>
          </>
        )}
      </>
    );
  }

  return (
    <>
      <SettingsCard title={t('settings:cors.card.title')} description={t('settings:cors.card.description')}>
        {body}
      </SettingsCard>
      {origins.dirty && (
        <UnsavedChangesBar
          message={t('settings:cors.unsavedChanges', 'You have unsaved changes')}
          resetLabel={t('settings:cors.reset', 'Reset')}
          saveLabel={t('settings:cors.save', 'Save changes')}
          savingLabel={t('settings:cors.saving', 'Saving...')}
          isSaving={updateCors.isPending}
          saveDisabled={origins.hasErrors}
          error={
            updateCors.error
              ? getErrorMessage(updateCors.error, tForErrors, 'cors.save.error', 'Failed to update allowed origins.')
              : undefined
          }
          onReset={() => {
            clearSaveError();
            origins.reset();
          }}
          onSave={handleSave}
        />
      )}
    </>
  );
}
