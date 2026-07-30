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

import {AuthenticatorTypes} from '@thunderid/configure-connections';
import {UserRound} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import AuthenticationMethodItem from './AuthenticationMethodItem';
import AuthMethodGroup from './AuthMethodGroup';

export interface PromptForCredentialsGroupProps {
  /**
   * Record of enabled authentication integrations
   */
  integrations: Record<string, boolean>;

  /**
   * Callback when an integration is toggled
   */
  onIntegrationToggle: (integrationId: string) => void;

  /**
   * Whether this group is disabled because a pre-configured flow is selected instead of
   * individual toggles.
   */
  disabled?: boolean;
}

/**
 * "Prompt for Credentials" category on the Sign-in Experience step: the classic username &
 * password prompt. The only method in this category, so the group's master checkbox always
 * mirrors the method's own toggle. Starts expanded — every other group starts collapsed — unless
 * a pre-configured flow is already selected, in which case it starts (and stays) collapsed like
 * the rest, since there's nothing to configure here while that's the case.
 */
export default function PromptForCredentialsGroup({
  integrations,
  onIntegrationToggle,
  disabled = false,
}: PromptForCredentialsGroupProps): JSX.Element {
  const {t} = useTranslation();

  const enabled = integrations[AuthenticatorTypes.CREDENTIALS_AUTH] ?? false;

  const handleCheckedChange = (checked: boolean): void => {
    if (checked !== enabled) onIntegrationToggle(AuthenticatorTypes.CREDENTIALS_AUTH);
  };

  return (
    <AuthMethodGroup
      title={t('applications:onboarding.configure.security.promptForCredentials.title', 'Prompt for Credentials')}
      checked={enabled}
      onCheckedChange={handleCheckedChange}
      disabled={disabled}
      defaultExpanded={!disabled}
    >
      <AuthenticationMethodItem
        id={AuthenticatorTypes.CREDENTIALS_AUTH}
        name={t('applications:onboarding.configure.SignInOptions.usernamePassword')}
        icon={<UserRound size={20} />}
        isEnabled={enabled}
        isAvailable
        isDisabled={disabled}
        onToggle={onIntegrationToggle}
      />
    </AuthMethodGroup>
  );
}
