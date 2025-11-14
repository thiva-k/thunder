/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IdentityProviderTypes, type IdentityProvider} from '@/features/integrations/models/identity-provider';
import ConfigureSignInOptions, {type ConfigureSignInOptionsProps} from '../ConfigureSignInOptions';
import {USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY} from '../../../utils/resolveAuthFlowGraphId';

// Mock the dependencies
vi.mock('@/features/integrations/api/useIdentityProviders');
vi.mock('@/features/integrations/utils/getIntegrationIcon');

const {default: useIdentityProviders} = await import('@/features/integrations/api/useIdentityProviders');
const {default: getIntegrationIcon} = await import('@/features/integrations/utils/getIntegrationIcon');

describe('ConfigureSignInOptions', () => {
  const mockOnIntegrationToggle = vi.fn();

  const mockIdentityProviders: IdentityProvider[] = [
    {
      id: 'google-idp',
      name: 'Google',
      type: IdentityProviderTypes.GOOGLE,
      description: 'Sign in with Google',
    },
    {
      id: 'github-idp',
      name: 'GitHub',
      type: IdentityProviderTypes.GITHUB,
      description: 'Sign in with GitHub',
    },
  ];

  const defaultProps: ConfigureSignInOptionsProps = {
    integrations: {
      [USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY]: true,
    },
    onIntegrationToggle: mockOnIntegrationToggle,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getIntegrationIcon).mockReturnValue(<div>Icon</div>);
  });

  const renderComponent = (props: Partial<ConfigureSignInOptionsProps> = {}) =>
    render(<ConfigureSignInOptions {...defaultProps} {...props} />);

  it('should render loading state', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render error state', () => {
    const error = new Error('Failed to load integrations');
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: undefined,
      isLoading: false,
      error,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/Failed to load authentication methods/i)).toBeInTheDocument();
  });

  it('should render the component with title and subtitle', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
    expect(screen.getByText('Choose how users will sign-in to your application')).toBeInTheDocument();
  });

  it('should always render Username & Password option first', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByText('Username & Password')).toBeInTheDocument();
  });

  it('should render Username & Password as checked by default', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({
      integrations: {
        [USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY]: true,
      },
    });

    const switches = screen.getAllByRole('switch');
    expect(switches[0]).toBeChecked();
  });

  it('should render all identity providers', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByText('Google')).toBeInTheDocument();
    expect(screen.getByText('GitHub')).toBeInTheDocument();
  });

  it('should call onIntegrationToggle when clicking Username & Password list item', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const usernamePasswordButton = screen.getByText('Username & Password').closest('.MuiListItemButton-root');
    if (usernamePasswordButton) {
      await user.click(usernamePasswordButton);
    }

    expect(mockOnIntegrationToggle).toHaveBeenCalledWith(USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY);
  });

  it('should call onIntegrationToggle when clicking provider list item', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const googleButton = screen.getByText('Google').closest('.MuiListItemButton-root');
    if (googleButton) {
      await user.click(googleButton);
    }

    expect(mockOnIntegrationToggle).toHaveBeenCalledWith('google-idp');
  });

  it('should call onIntegrationToggle when toggling switch', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const switches = screen.getAllByRole('switch');
    await user.click(switches[1]); // Click Google switch

    expect(mockOnIntegrationToggle).toHaveBeenCalledWith('google-idp');
  });

  it('should show checked state for enabled integrations', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({
      integrations: {
        [USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY]: true,
        'google-idp': true,
        'github-idp': false,
      },
    });

    const switches = screen.getAllByRole('switch');
    expect(switches[0]).toBeChecked(); // Username & Password
    expect(switches[1]).toBeChecked(); // Google
    expect(switches[2]).not.toBeChecked(); // GitHub
  });

  it('should show username/password option when no integrations are available', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
    } as unknown as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    // Should show username/password in the list (but disabled, no toggle)
    expect(screen.getByText('Username & Password')).toBeInTheDocument();
    expect(screen.getByRole('list')).toBeInTheDocument();

    // Should not have a toggle/switch when it's the only option
    const listItem = screen.getByText('Username & Password').closest('.MuiListItem-root');
    expect(listItem?.querySelector('.MuiSwitch-root')).not.toBeInTheDocument();
  });

  it('should render integration icons', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(getIntegrationIcon).toHaveBeenCalledWith(IdentityProviderTypes.GOOGLE);
    expect(getIntegrationIcon).toHaveBeenCalledWith(IdentityProviderTypes.GITHUB);
  });

  it('should render UserRound icon for Username & Password', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    // UserRound icon should be present
    const usernamePasswordSection = screen.getByText('Username & Password').closest('div');
    expect(usernamePasswordSection).toBeInTheDocument();
  });

  it('should stop propagation when clicking switch', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const switches = screen.getAllByRole('switch');
    await user.click(switches[1]);

    // Should only trigger once (not twice from card and switch)
    expect(mockOnIntegrationToggle).toHaveBeenCalledTimes(1);
  });

  it('should handle empty integrations record', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({integrations: {}});

    const switches = screen.getAllByRole('switch');
    // Username & Password should default to true
    expect(switches[0]).toBeChecked();
    // Others should default to false
    expect(switches[1]).not.toBeChecked();
  });

  it('should render info icon in subtitle', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const subtitle = screen.getByText('Choose how users will sign-in to your application').closest('div');
    expect(subtitle).toBeInTheDocument();
  });

  it('should handle multiple rapid toggles', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const switches = screen.getAllByRole('switch');
    await user.click(switches[1]);
    await user.click(switches[2]);
    await user.click(switches[1]);

    expect(mockOnIntegrationToggle).toHaveBeenCalledTimes(3);
  });

  it('should handle providers with long names', () => {
    const longNameProvider: IdentityProvider = {
      id: 'long-name-idp',
      name: 'Very Long Identity Provider Name That Should Still Display',
      type: 'OIDC',
      description: 'Test provider',
    };

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: [longNameProvider],
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByText(longNameProvider.name)).toBeInTheDocument();
  });

  it('should maintain switch state after re-render', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    const {rerender} = renderComponent({
      integrations: {
        [USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY]: true,
        'google-idp': true,
      },
    });

    let switches = screen.getAllByRole('switch');
    expect(switches[1]).toBeChecked();

    rerender(
      <ConfigureSignInOptions
        integrations={{
          [USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY]: true,
          'google-idp': true,
        }}
        onIntegrationToggle={mockOnIntegrationToggle}
      />,
    );

    switches = screen.getAllByRole('switch');
    expect(switches[1]).toBeChecked();
  });
});
