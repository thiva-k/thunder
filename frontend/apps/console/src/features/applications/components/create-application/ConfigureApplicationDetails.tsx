/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {useGetOrganizationUnit} from '@thunderid/configure-organization-units';
import type {UserTypeListItem} from '@thunderid/configure-user-types';
import {Stack, Typography, Box, Divider} from '@wso2/oxygen-ui';
import {Home} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useEffect, useRef} from 'react';
import {useTranslation} from 'react-i18next';
import ConfigureName from './ConfigureName';
import OrganizationUnitDefaultsSection from './OrganizationUnitDefaultsSection';
import UserAccessSection from './UserAccessSection';
import {
  type OrganizationUnitDefaultItem,
  type OrganizationUnitDefaultsSelection,
} from '../../models/application-create-flow';
import computeOrganizationUnitDefaultAvailability from '../../utils/computeOrganizationUnitDefaultAvailability';

export interface ConfigureApplicationDetailsProps {
  /**
   * Whether more than one organization unit exists. The OU picker only renders when true.
   */
  hasMultipleOUs: boolean;

  /**
   * The selected organization unit ID, picked on the pre-wizard organization unit screen (only
   * meaningful when `hasMultipleOUs` is true).
   */
  selectedOuId: string;

  /**
   * Callback invoked when the "Change" link is clicked, reopening the pre-wizard organization
   * unit screen.
   */
  onChangeOu: () => void;

  /**
   * The organization unit whose defaults should be offered: the picked one when there are
   * multiple, or the deployment's sole organization unit otherwise. Undefined until resolved.
   */
  resolvedOuId: string | undefined;

  /**
   * The current application name.
   */
  appName: string;

  /**
   * Callback invoked when the application name changes.
   */
  onAppNameChange: (name: string) => void;

  /**
   * URL or emoji of the currently selected application logo.
   */
  appLogo: string | null;

  /**
   * Callback invoked when a logo is selected.
   */
  onLogoSelect: (logoUrl: string) => void;

  /**
   * Current per-item "use organization unit default" selection.
   */
  ouDefaults: OrganizationUnitDefaultsSelection;

  /**
   * Callback invoked when the organization unit defaults selection changes.
   */
  onOuDefaultsChange: (selection: OrganizationUnitDefaultsSelection) => void;

  /**
   * Whether this template's flow includes a Security step. Templates without one (e.g.
   * machine-to-machine backends) never present a hosted sign-in screen, so the "Use organization
   * defaults" accordion hides the sign-in/sign-up/recovery/sign-out flows group for them.
   */
  hasSecurityStep: boolean;

  /**
   * Whether this template's flow includes a Design step. Templates without one have no theme or
   * layout of their own, so the "Use organization defaults" accordion hides the Design group.
   */
  hasDesignStep: boolean;

  /**
   * Available user types. The User Access section only renders when there are two or more.
   */
  userTypes: UserTypeListItem[];

  /**
   * Currently selected user type names.
   */
  selectedUserTypes: string[];

  /**
   * Callback invoked when the user type selection changes.
   */
  onUserTypesChange: (userTypes: string[]) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed.
   */
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * First step of the application creation wizard: pick an organization unit (only shown when more
 * than one exists), name the application, choose which of that organization unit's configured
 * defaults (sign in / sign up / recovery / sign out flow, theme, layout) to snapshot into the new
 * application instead of building each one from scratch, and pick which user types can access it.
 */
export default function ConfigureApplicationDetails({
  hasMultipleOUs,
  selectedOuId,
  onChangeOu,
  resolvedOuId,
  appName,
  onAppNameChange,
  appLogo,
  onLogoSelect,
  ouDefaults,
  onOuDefaultsChange,
  hasSecurityStep,
  hasDesignStep,
  userTypes,
  selectedUserTypes,
  onUserTypesChange,
  onReadyChange = undefined,
}: ConfigureApplicationDetailsProps): JSX.Element {
  const {t} = useTranslation();
  const {data: organizationUnit, isLoading: isOuLoading} = useGetOrganizationUnit(resolvedOuId, Boolean(resolvedOuId));

  useEffect((): void => {
    if (!onReadyChange) return;
    const nameReady = appName.trim().length > 0;
    const ouReady = !hasMultipleOUs || selectedOuId.length > 0;
    const userAccessReady = userTypes.length < 2 || selectedUserTypes.length > 0;
    onReadyChange(nameReady && ouReady && userAccessReady);
  }, [appName, hasMultipleOUs, selectedOuId, userTypes, selectedUserTypes, onReadyChange]);

  // Default to "allow all" the first time user types load, mirroring the master checkbox's
  // default-checked state in UserAccessSection. Seeds at most once — otherwise this would fight
  // an admin who deliberately unchecks every type, snapping the selection back to "all" on the
  // very next render.
  const hasSeededUserTypesRef = useRef(false);
  useEffect((): void => {
    if (hasSeededUserTypesRef.current || userTypes.length < 2) return;
    hasSeededUserTypesRef.current = true;
    if (selectedUserTypes.length === 0) {
      onUserTypesChange(userTypes.map((userType) => userType.name));
    }
  }, [userTypes, selectedUserTypes.length, onUserTypesChange]);

  // Pre-check whichever organization-unit defaults are actually available, once per resolved OU
  // id, so admins don't have to manually opt into inheriting them. Keyed by OU id (not a boolean)
  // so picking a *different* OU via "Change" re-seeds, while a manual uncheck on the same OU is
  // never fought (the effect never fires twice for the same id).
  const seededOuIdRef = useRef<string | undefined>(undefined);
  useEffect((): void => {
    if (isOuLoading || !organizationUnit || !resolvedOuId) return;
    if (seededOuIdRef.current === resolvedOuId) return;
    seededOuIdRef.current = resolvedOuId;

    const availability = computeOrganizationUnitDefaultAvailability(organizationUnit, {hasSecurityStep, hasDesignStep});
    const seeded = {...ouDefaults};
    (Object.keys(availability) as OrganizationUnitDefaultItem[]).forEach((item) => {
      if (availability[item]) seeded[item] = true;
    });
    onOuDefaultsChange(seeded);
    // eslint-disable-next-line react-hooks/exhaustive-deps -- deliberately omits ouDefaults/onOuDefaultsChange: this must seed at most once per resolved OU id, not refire on every ouDefaults change.
  }, [isOuLoading, organizationUnit, resolvedOuId, hasSecurityStep, hasDesignStep]);

  const showOuDefaults = Boolean(
    !isOuLoading &&
      organizationUnit &&
      Object.values(
        computeOrganizationUnitDefaultAvailability(organizationUnit, {hasSecurityStep, hasDesignStep}),
      ).some(Boolean),
  );
  const showUserAccess = userTypes.length >= 2;

  return (
    <Stack direction="column" spacing={4} data-testid="application-configure-details">
      <Typography variant="h1" gutterBottom>
        {t(
          'applications:onboarding.configure.applicationDetails.title',
          "Let's collect some details about your application",
        )}
      </Typography>

      {hasMultipleOUs && (
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            gap: 1.5,
            p: '12px 14px',
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: '10px',
            bgcolor: 'background.paper',
          }}
        >
          <Box
            sx={{
              width: 32,
              height: 32,
              borderRadius: '8px',
              bgcolor: 'primary.main',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexShrink: 0,
            }}
          >
            <Home size={15} color="#fff" />
          </Box>
          <Box sx={{flex: 1, minWidth: 0}}>
            <Typography
              variant="overline"
              color="text.secondary"
              sx={{display: 'block', lineHeight: 1.2, fontWeight: 600}}
            >
              {t('applications:onboarding.organizationUnit.fieldLabel', 'Organization Unit')}
            </Typography>
            <Typography variant="body2" sx={{fontWeight: 600}}>
              {isOuLoading ? t('common:status.loading') : organizationUnit?.name}
            </Typography>
          </Box>
          <Typography
            component="span"
            variant="body2"
            onClick={onChangeOu}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                onChangeOu();
              }
            }}
            sx={{fontWeight: 500, color: 'primary.main', cursor: 'pointer', flexShrink: 0}}
          >
            {t('common:actions.change')}
          </Typography>
        </Box>
      )}

      <ConfigureName
        appName={appName}
        onAppNameChange={onAppNameChange}
        appLogo={appLogo}
        onLogoSelect={onLogoSelect}
        showTitle={false}
      />

      {(showOuDefaults || showUserAccess) && (
        <Box sx={{border: '1px solid', borderColor: 'divider', borderRadius: '10px', overflow: 'hidden'}}>
          {showOuDefaults && (
            <OrganizationUnitDefaultsSection
              organizationUnit={organizationUnit}
              isLoading={isOuLoading}
              ouDefaults={ouDefaults}
              onChange={onOuDefaultsChange}
              hasSecurityStep={hasSecurityStep}
              hasDesignStep={hasDesignStep}
            />
          )}

          {showOuDefaults && showUserAccess && <Divider />}

          {showUserAccess && (
            <UserAccessSection
              userTypes={userTypes}
              selectedUserTypes={selectedUserTypes}
              onUserTypesChange={onUserTypesChange}
            />
          )}
        </Box>
      )}
    </Stack>
  );
}
