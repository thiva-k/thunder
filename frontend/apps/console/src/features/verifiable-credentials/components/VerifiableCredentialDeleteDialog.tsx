// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button, Alert} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useDeleteVerifiableCredential from '../api/useDeleteVerifiableCredential';

export interface VerifiableCredentialDeleteDialogProps {
  open: boolean;
  vcId: string | null;
  onClose: () => void;
  onSuccess?: () => void;
}

/**
 * Dialog to confirm deletion of a credential configuration.
 */
export default function VerifiableCredentialDeleteDialog({
  open,
  vcId,
  onClose,
  onSuccess = undefined,
}: VerifiableCredentialDeleteDialogProps): JSX.Element {
  const {t} = useTranslation();
  const deleteVC = useDeleteVerifiableCredential();
  const [error, setError] = useState<string | null>(null);

  const handleCancel = (): void => {
    if (deleteVC.isPending) return;
    setError(null);
    onClose();
  };

  const handleConfirm = (): void => {
    if (!vcId) return;
    setError(null);
    deleteVC.mutate(vcId, {
      onSuccess: (): void => {
        setError(null);
        onClose();
        onSuccess?.();
      },
      onError: (err: Error) => {
        setError(err.message ?? t('verifiable-credentials:delete.error'));
      },
    });
  };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>{t('verifiable-credentials:delete.title')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 2}}>{t('verifiable-credentials:delete.message')}</DialogContentText>
        <Alert severity="warning" sx={{mb: 2}}>
          {t('verifiable-credentials:delete.disclaimer')}
        </Alert>
        {error && (
          <Alert severity="error" sx={{mt: 2}}>
            {error}
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} disabled={deleteVC.isPending}>
          {t('common:actions.cancel')}
        </Button>
        <Button onClick={handleConfirm} color="error" variant="contained" disabled={deleteVC.isPending || !vcId}>
          {deleteVC.isPending ? t('common:status.deleting') : t('common:actions.delete')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
