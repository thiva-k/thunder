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

import type {OrganizationUnit} from '@thunderid/configure-organization-units';
import {OrganizationUnitDefaultItem} from '../models/application-create-flow';

export interface OrganizationUnitDefaultAvailabilityFlowFlags {
  /** Whether the selected template's flow includes a Security step. */
  hasSecurityStep: boolean;
  /** Whether the selected template's flow includes a Design step. */
  hasDesignStep: boolean;
}

/**
 * Which organization-unit-defaultable items are actually offerable: the OU has a configured value
 * for the item, *and* the selected template's flow has a step for it in the first place (a
 * template with no Security step never presents a hosted sign-in screen to inherit one for, and
 * one with no Design step has no theme/layout of its own). Shared between
 * `OrganizationUnitDefaultsSection` (which item rows to show) and `ConfigureApplicationDetails`
 * (which items to pre-check the first time an OU resolves).
 */
const computeOrganizationUnitDefaultAvailability = (
  organizationUnit: OrganizationUnit | undefined,
  {hasSecurityStep, hasDesignStep}: OrganizationUnitDefaultAvailabilityFlowFlags,
): Record<OrganizationUnitDefaultItem, boolean> => ({
  [OrganizationUnitDefaultItem.SIGN_IN]: hasSecurityStep && Boolean(organizationUnit?.authFlowId),
  [OrganizationUnitDefaultItem.SIGN_UP]: hasSecurityStep && Boolean(organizationUnit?.registrationFlowId),
  [OrganizationUnitDefaultItem.RECOVERY]: hasSecurityStep && Boolean(organizationUnit?.recoveryFlowId),
  [OrganizationUnitDefaultItem.SIGN_OUT]: hasSecurityStep && Boolean(organizationUnit?.signOutFlowId),
  [OrganizationUnitDefaultItem.THEME]: hasDesignStep && Boolean(organizationUnit?.themeId),
  [OrganizationUnitDefaultItem.LAYOUT]: hasDesignStep && Boolean(organizationUnit?.layoutId),
});

export default computeOrganizationUnitDefaultAvailability;
