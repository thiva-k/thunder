// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useLogger} from '@thunderid/logger/react';
import {Stack, Button, PageContent, PageTitle} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import RouteConfig from '../../../configs/RouteConfig';
import RolesList from '../components/RolesList';

export default function RolesListPage(): JSX.Element {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('RolesListPage');

  return (
    <PageContent>
      <PageTitle>
        <PageTitle.Header>{t('roles:listing.title')}</PageTitle.Header>
        <PageTitle.SubHeader>{t('roles:listing.subtitle')}</PageTitle.SubHeader>
        <PageTitle.Actions>
          <Stack direction="row" spacing={2}>
            <Button
              variant="contained"
              startIcon={<Plus size={18} />}
              onClick={() => {
                (async () => {
                  await navigate(RouteConfig.roles.create());
                })().catch((error: unknown) => {
                  logger.error('Failed to navigate to create role page', {error});
                });
              }}
            >
              {t('roles:listing.addRole')}
            </Button>
          </Stack>
        </PageTitle.Actions>
      </PageTitle>

      <RolesList />
    </PageContent>
  );
}
