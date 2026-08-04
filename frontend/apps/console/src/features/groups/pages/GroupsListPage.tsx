// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {ExternalLink} from '@thunderid/components';
import {useLogger} from '@thunderid/logger/react';
import {Stack, Button, PageContent, PageTitle} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import RouteConfig from '../../../configs/RouteConfig';
import GroupsList from '../components/GroupsList';

export default function GroupsListPage(): JSX.Element {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('GroupsListPage');

  return (
    <PageContent>
      <PageTitle>
        <PageTitle.Header>{t('groups:listing.title')}</PageTitle.Header>
        <PageTitle.SubHeader>
          {t('groups:listing.subtitle')} <ExternalLink docKey="groups" confirmBeforeNavigate={false} />
        </PageTitle.SubHeader>
        <PageTitle.Actions>
          <Stack direction="row" spacing={2}>
            <Button
              variant="contained"
              startIcon={<Plus size={18} />}
              onClick={() => {
                (async () => {
                  await navigate(RouteConfig.groups.create());
                })().catch((error: unknown) => {
                  logger.error('Failed to navigate to create group page', {error});
                });
              }}
            >
              {t('groups:listing.addGroup')}
            </Button>
          </Stack>
        </PageTitle.Actions>
      </PageTitle>

      <GroupsList />
    </PageContent>
  );
}
