// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {Application, OAuth2Config} from '@thunderid/configure-applications';
import {Box, Tab, Tabs} from '@wso2/oxygen-ui';
import {useEffect, useState, type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';
import ClientAccessTokenSection from './ClientAccessTokenSection';
import EditTokenSettings from './EditTokenSettings';
import TokenConstants from '../../../constants/token-constants';
import {hasClientAccess, hasUserAccess} from '../../../utils/oauth2Rules';
import SettingsLockNotice from '../../common/SettingsLockNotice';

interface EditTokenSettingsTabsProps {
  application: Application;
  oauth2Config?: OAuth2Config;
  onFieldChange: (field: keyof Application, value: unknown) => void;
  onValidationChange?: (hasErrors: boolean) => void;
  sectionResetKey?: number;
}

/**
 * Token settings split into Application (client_credentials) and User sub-tabs. Each sub-tab is
 * frozen with a lock notice when its enabling grant type is not selected.
 */
export default function EditTokenSettingsTabs({
  application,
  oauth2Config = undefined,
  onFieldChange,
  onValidationChange = undefined,
  sectionResetKey = 0,
}: EditTokenSettingsTabsProps): JSX.Element {
  const {t} = useTranslation();
  const [subTab, setSubTab] = useState(0);
  const [clientTabHasError, setClientTabHasError] = useState(false);
  const [userTabHasError, setUserTabHasError] = useState(false);

  const grantTypes = oauth2Config?.grantTypes;
  const clientUnlocked = hasClientAccess(grantTypes);
  const userUnlocked = hasUserAccess(grantTypes);

  useEffect(() => {
    onValidationChange?.(clientTabHasError || userTabHasError);
  }, [clientTabHasError, userTabHasError, onValidationChange]);

  // Forcing isReadOnly disables every input via EditTokenSettings' existing
  // disabled={application.isReadOnly} wiring when the user tab is locked.
  const userTabApplication = {
    ...application,
    isReadOnly: (application.isReadOnly ?? false) || !userUnlocked,
  };

  const handleSubTabChange = (_event: SyntheticEvent, newValue: number): void => {
    setSubTab(newValue);
  };

  return (
    <Box>
      <Tabs value={subTab} onChange={handleSubTabChange} aria-label="application token settings sub-tabs">
        <Tab label={t('applications:edit.token.tabs.application', 'Application')} sx={{textTransform: 'none'}} />
        <Tab label={t('applications:edit.token.tabs.user', 'User')} sx={{textTransform: 'none'}} />
      </Tabs>
      <Box sx={{pt: 3}}>
        {subTab === 0 && (
          <SettingsLockNotice
            isUnlocked={clientUnlocked}
            message={t(
              'applications:edit.token.clientLock.message',
              'These settings apply only when the client credentials grant is enabled. Add it in the Advanced tab to configure them.',
            )}
          >
            <ClientAccessTokenSection
              key={sectionResetKey}
              oauth2Config={oauth2Config}
              inboundAuthConfig={application.inboundAuthConfig}
              onFieldChange={onFieldChange}
              availableAttributes={[...TokenConstants.CLIENT_TOKEN_OPTIONAL_CLAIMS]}
              disabled={(application.isReadOnly ?? false) || !clientUnlocked}
              onValidationChange={setClientTabHasError}
              subjectValue={application.id}
              copy={{
                attributesTitle: t('applications:edit.token.client.attributes.title', 'Access Token Claims'),
                attributesDescription: t(
                  'applications:edit.token.client.attributes.description',
                  'Optional claims to include in the access token this application receives for its own requests (client_credentials grant).',
                ),
                attributesLabel: t('applications:edit.token.client.attributes.label', 'Add or Remove Claims'),
                attributesHint: t(
                  'applications:edit.token.client.attributes.hint',
                  "Click a claim to include it in this application's client access token.",
                ),
                attributesEmpty: t('applications:edit.token.client.attributes.empty', 'No optional claims available.'),
                validityTitle: t('applications:edit.token.client.validity.title', 'Token Validity'),
                validityDescription: t(
                  'applications:edit.token.client.validity.description',
                  'How long this client access token remains valid before expiration.',
                ),
                validityLabel: t('applications:edit.token.client.validity.label', 'Token Validity'),
                validityHint: t(
                  'applications:edit.token.client.validity.hint',
                  'Token validity period in seconds (e.g., 3600 for 1 hour).',
                ),
                validityError: t(
                  'applications:edit.token.client.validity.error',
                  'Enter a validity period of at least 1 second.',
                ),
              }}
            />
          </SettingsLockNotice>
        )}
        {subTab === 1 && (
          <SettingsLockNotice
            isUnlocked={userUnlocked}
            message={t(
              'applications:edit.token.userLock.message',
              'These settings apply only when a user-facing grant is enabled. Add one in the Advanced tab to configure them.',
            )}
          >
            <EditTokenSettings
              application={userTabApplication}
              oauth2Config={oauth2Config}
              sectionResetKey={sectionResetKey}
              onFieldChange={onFieldChange}
              onValidationChange={setUserTabHasError}
            />
          </SettingsLockNotice>
        )}
      </Box>
    </Box>
  );
}
