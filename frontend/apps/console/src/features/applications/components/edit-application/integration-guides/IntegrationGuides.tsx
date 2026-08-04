// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {Application, OAuth2Config} from '@thunderid/configure-applications';
import {Box, Stack, Typography} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import IntegrationGuide from './IntegrationGuide';
import getIntegrationGuidesForTemplate from '../../../utils/getIntegrationGuidesForTemplate';

/**
 * Props for the {@link IntegrationGuides} component.
 */
interface IntegrationGuidesProps {
  /**
   * The application to show integration guides for
   */
  application: Application;
  /**
   * OAuth2 configuration containing client credentials (optional)
   */
  oauth2Config?: OAuth2Config;
}

/**
 * Container component for displaying integration guides.
 *
 * Fetches integration guides based on the application's template and displays:
 * - Technology-specific integration guides with code snippets
 * - Setup instructions for various frameworks
 * - A fallback message if no guides are available
 *
 * @param props - Component props
 * @returns Integration guides UI or a message if no guides are available
 */
export default function IntegrationGuides({application, oauth2Config = undefined}: IntegrationGuidesProps) {
  const {t} = useTranslation();

  const integrationGuides = getIntegrationGuidesForTemplate(application.template ?? '');

  return (
    <Stack spacing={4}>
      {integrationGuides ? (
        <IntegrationGuide
          clientId={oauth2Config?.clientId ?? ''}
          applicationId={application?.id}
          integrationGuides={integrationGuides}
          templateId={application?.template ?? null}
        />
      ) : (
        <Box>
          <Typography variant="body1" color="text.secondary">
            {t('applications:edit.overview.noGuides')}
          </Typography>
        </Box>
      )}
    </Stack>
  );
}
