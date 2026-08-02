// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button, Alert} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useDeleteOrganizationUnit from '../api/useDeleteOrganizationUnit';
import type {ApiError, I18nMessage} from '../models/api-error';

export interface OrganizationUnitDeleteDialogProps {
  /**
   * Whether the dialog is open
   */
  open: boolean;
  /**
   * The ID of the organization unit to delete
   */
  organizationUnitId: string | null;
  /**
   * Callback when the dialog should be closed
   */
  onClose: () => void;
  /**
   * Callback when the organization unit is successfully deleted
   */
  onSuccess?: () => void;
  /**
   * Callback when the deletion fails, receives the error message
   */
  onError?: (message: string) => void;
}

/**
 * Resolves a backend error field to a display string. The field may be a plain
 * string or a translatable {@link I18nMessage} object; objects are reduced to
 * their `defaultValue` so they are never rendered directly as React children.
 */
function resolveField(value: I18nMessage | string | undefined): string | null {
  if (typeof value === 'string') {
    return value.trim() ? value : null;
  }

  const defaultValue = value?.defaultValue;

  return defaultValue?.trim() ? defaultValue : null;
}

/**
 * Extracts a user-friendly error message from the API error response.
 */
function getErrorMessage(err: Error, fallback: string): string {
  const {response} = err as Error & {response?: {data?: ApiError}};
  const description = resolveField(response?.data?.description);
  const message = err.message?.trim() ? err.message : null;

  return description ?? message ?? fallback;
}

/**
 * Dialog component for confirming organization unit deletion
 */
export default function OrganizationUnitDeleteDialog({
  open,
  organizationUnitId,
  onClose,
  onSuccess = undefined,
  onError = undefined,
}: OrganizationUnitDeleteDialogProps): JSX.Element {
  const {t} = useTranslation();
  const deleteOrganizationUnit = useDeleteOrganizationUnit();

  const handleCancel = (): void => {
    if (deleteOrganizationUnit.isPending) return;
    onClose();
  };

  const handleConfirm = (): void => {
    if (!organizationUnitId) return;

    deleteOrganizationUnit.mutate(organizationUnitId, {
      onSuccess: (): void => {
        onClose();
        onSuccess?.();
      },
      onError: (err: Error) => {
        const message = getErrorMessage(err, t('organizationUnits:delete.dialog.error'));
        onClose();
        onError?.(message);
      },
    });
  };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{t('organizationUnits:delete.dialog.title')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>{t('organizationUnits:delete.dialog.message')}</DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('organizationUnits:delete.dialog.disclaimer')}
        </Alert>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} disabled={deleteOrganizationUnit.isPending}>
          {t('common:actions.cancel')}
        </Button>
        <Button onClick={handleConfirm} color="error" variant="contained" disabled={deleteOrganizationUnit.isPending}>
          {deleteOrganizationUnit.isPending ? t('common:status.deleting') : t('common:actions.delete')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
