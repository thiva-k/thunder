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

import {render, screen} from '@testing-library/react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import ConfigureSecuritySettings from '../ConfigureSecuritySettings';
import {OrganizationUnitDefaultItem} from '@/features/applications/models/application-create-flow';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string) => fallback ?? key,
  }),
}));

let mockOuDefaults: Record<string, boolean>;
vi.mock('@/features/applications/hooks/useApplicationCreateContext', () => ({
  default: () => ({
    integrations: {},
    toggleIntegration: vi.fn(),
    ouDefaults: mockOuDefaults,
  }),
}));

// ConfigureMfaSettings (and the other sign-in categories) are now rendered internally by
// ConfigureSignInOptions itself (see ConfigureSignInOptions.tsx), so this stub only needs to
// stand in for the whole sign-in section; their own rendering is covered by
// ConfigureSignInOptions.test.tsx and ConfigureMfaSettings.test.tsx.
vi.mock('../../configure-signin-options/ConfigureSignInOptions', () => ({
  default: () => <div data-testid="sign-in-options" />,
}));

describe('ConfigureSecuritySettings', () => {
  beforeEach(() => {
    mockOuDefaults = {[OrganizationUnitDefaultItem.SIGN_IN]: false};
  });

  it('renders the sign-in section when sign-in is not snapshotted from the OU', () => {
    render(<ConfigureSecuritySettings />);

    expect(screen.getByTestId('sign-in-options')).toBeInTheDocument();
  });

  it('renders no sign-in section when sign-in is snapshotted from the organization unit default', () => {
    mockOuDefaults = {[OrganizationUnitDefaultItem.SIGN_IN]: true};
    render(<ConfigureSecuritySettings />);

    expect(screen.queryByTestId('sign-in-options')).not.toBeInTheDocument();
  });
});
