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

import {fireEvent, render, screen} from '@testing-library/react';
import {AuthenticatorTypes} from '@thunderid/configure-connections';
import {describe, it, expect, vi} from 'vitest';
import PasswordlessLoginGroup, {type PasswordlessLoginGroupProps} from '../PasswordlessLoginGroup';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string) => fallback ?? key,
  }),
}));

describe('PasswordlessLoginGroup', () => {
  const mockOnIntegrationToggle = vi.fn();

  const defaultProps: PasswordlessLoginGroupProps = {
    integrations: {[AuthenticatorTypes.PASSKEY]: false, [AuthenticatorTypes.MAGIC_LINK]: false},
    onIntegrationToggle: mockOnIntegrationToggle,
  };

  it('starts collapsed, same as always, when no flow is selected', () => {
    render(<PasswordlessLoginGroup {...defaultProps} />);

    expect(screen.queryByTestId(`auth-method-${AuthenticatorTypes.PASSKEY}`)).not.toBeInTheDocument();
    expect(screen.getByRole('checkbox')).not.toBeDisabled();
    expect(screen.getByRole('button', {name: 'Expand'})).not.toBeDisabled();
  });

  it('collapses an already-expanded list and disables the checkbox/chevron once a pre-configured flow is selected', () => {
    const {rerender} = render(<PasswordlessLoginGroup {...defaultProps} />);
    const chevron = screen.getByRole('button', {name: 'Expand'});
    fireEvent.click(chevron);
    expect(screen.getByTestId(`auth-method-${AuthenticatorTypes.PASSKEY}`)).toBeInTheDocument();
    expect(chevron).toHaveAttribute('aria-expanded', 'true');

    rerender(<PasswordlessLoginGroup {...defaultProps} disabled />);

    // The Collapse's own exit transition (not run by jsdom) governs when the row actually leaves
    // the DOM under `unmountOnExit` — asserted here via the chevron/checkbox's immediate state
    // instead, which the disabling effect updates synchronously regardless of that animation.
    expect(chevron).toHaveAttribute('aria-expanded', 'false');
    expect(screen.getByRole('checkbox')).toBeDisabled();
    expect(chevron).toBeDisabled();
  });
});
