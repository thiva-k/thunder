// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useToast} from '@thunderid/contexts';
import {useLogger} from '@thunderid/logger/react';
import {Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, Stack, Typography} from '@wso2/oxygen-ui';
import {Star} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useSetDefaultResourceServer from '../api/useSetDefaultResourceServer';
import type {ResourceServer} from '../models/resource-server';

export interface SetDefaultResourceServerDialogProps {
  open: boolean;
  resourceServer: ResourceServer | null;
  onClose: () => void;
  onSuccess?: () => void;
}

export default function SetDefaultResourceServerDialog({
  open,
  resourceServer,
  onClose,
  onSuccess = undefined,
}: SetDefaultResourceServerDialogProps): JSX.Element {
  const {t} = useTranslation();
  const {showToast} = useToast();
  const logger = useLogger('SetDefaultResourceServerDialog');
  const setDefault = useSetDefaultResourceServer();

  const handleConfirm = (): void => {
    if (!resourceServer) return;

    setDefault.mutate(
      {resourceServerId: resourceServer.id},
      {
        onSuccess: () => {
          showToast(
            t('resourceServers:setDefault.success', '{{name}} is now the default resource server.', {
              name: resourceServer.name,
            }),
            'success',
          );
          onSuccess?.();
          onClose();
        },
        onError: (err: Error) => {
          logger.error('Failed to set default resource server', {error: err});
          showToast(t('resourceServers:setDefault.error', 'Failed to set the default resource server.'), 'error');
        },
      },
    );
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>
        <Stack direction="row" alignItems="center" spacing={1.5}>
          <Box
            sx={{
              width: 40,
              height: 40,
              borderRadius: 2,
              flexShrink: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'primary.main',
              bgcolor: 'primary.light',
            }}
          >
            <Star size={20} />
          </Box>
          <span>{t('resourceServers:setDefault.title', 'Set default resource server')}</span>
        </Stack>
      </DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="text.secondary">
          <strong>{resourceServer?.name}</strong>{' '}
          {t(
            'resourceServers:setDefault.message',
            'will become the default resource server. Requests without a resource parameter will fall back to it.',
          )}
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button variant="outlined" onClick={onClose} disabled={setDefault.isPending}>
          {t('common:cancel', 'Cancel')}
        </Button>
        <Button variant="contained" onClick={handleConfirm} disabled={setDefault.isPending}>
          {setDefault.isPending
            ? t('resourceServers:setDefault.setting', 'Setting…')
            : t('resourceServers:setDefault.confirm', 'Set as default')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
