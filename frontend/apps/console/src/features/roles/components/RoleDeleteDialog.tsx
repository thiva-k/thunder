// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button, Alert} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useDeleteRole from '../api/useDeleteRole';

export interface RoleDeleteDialogProps {
  open: boolean;
  roleId: string | null;
  onClose: () => void;
  onSuccess?: () => void;
}

/**
 * Dialog component for confirming role deletion.
 */
export default function RoleDeleteDialog({
  open,
  roleId,
  onClose,
  onSuccess = undefined,
}: RoleDeleteDialogProps): JSX.Element {
  const {t} = useTranslation();
  const deleteRole = useDeleteRole();
  const [error, setError] = useState<string | null>(null);

  const handleCancel = (): void => {
    if (deleteRole.isPending) return;
    setError(null);
    onClose();
  };

  const handleConfirm = (): void => {
    if (!roleId) return;

    setError(null);
    deleteRole.mutate(roleId, {
      onSuccess: (): void => {
        setError(null);
        onClose();
        onSuccess?.();
      },
      onError: (err: Error) => {
        setError(err.message ?? t('roles:delete.error'));
      },
    });
  };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{t('roles:delete.title')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>{t('roles:delete.message')}</DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('roles:delete.disclaimer')}
        </Alert>
        {error && (
          <Alert severity="error" sx={{mt: 2}}>
            {error}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} disabled={deleteRole.isPending}>
          {t('common:actions.cancel')}
        </Button>
        <Button onClick={handleConfirm} color="error" variant="contained" disabled={deleteRole.isPending || !roleId}>
          {deleteRole.isPending ? t('common:status.deleting') : t('common:actions.delete')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
