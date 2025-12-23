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
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import {type BasicFlowDefinition} from '@/features/flows/models/responses';
import IndividualMethodsToggleView, {type IndividualMethodsToggleViewProps} from '../IndividualMethodsToggleView';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:onboarding.configure.SignInOptions.usernamePassword': 'Username & Password',
        'applications:onboarding.configure.SignInOptions.google': 'Google',
        'applications:onboarding.configure.SignInOptions.github': 'GitHub',
        'applications:onboarding.configure.SignInOptions.smsOtp': 'SMS OTP',
        'applications:onboarding.configure.SignInOptions.notConfigured': 'Not configured',
      };
      return translations[key] || key;
    },
  }),
}));

// Mock the integration icon utility
vi.mock('@/features/integrations/utils/getIntegrationIcon', () => ({
  default: vi.fn((type: string) => {
    const iconMap: Record<string, string> = {
      [IdentityProviderTypes.OAUTH]: 'icon-oauth',
      [IdentityProviderTypes.OIDC]: 'icon-oidc',
    };
    const testId = iconMap[type] || `icon-${type}`;
    return <div data-testid={testId}>Mock Icon</div>;
  }),
}));

const {default: getIntegrationIcon} = await import('@/features/integrations/utils/getIntegrationIcon');

describe.skip('IndividualMethodsToggleView', () => {
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
    {
      id: 'oauth-idp-1',
      name: 'OAuth Provider',
      type: IdentityProviderTypes.OAUTH,
      description: 'Sign in with OAuth',
    },
    {
      id: 'oidc-idp-1',
      name: 'OIDC Provider',
      type: IdentityProviderTypes.OIDC,
      description: 'Sign in with OIDC',
    },
  ];

  const mockFlowsByType = {
    basic: {
      id: 'basic-flow',
      name: 'Basic Authentication Flow',
      activeVersion: 1,
      handle: 'basic-auth-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '2023-01-01T00:00:00Z',
      updatedAt: '2023-01-01T00:00:00Z',
    } as BasicFlowDefinition,
    google: {
      id: 'google-flow',
      name: 'Google OAuth Flow',
      activeVersion: 1,
      handle: 'google-oauth-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '2023-01-01T00:00:00Z',
      updatedAt: '2023-01-01T00:00:00Z',
    } as BasicFlowDefinition,
    github: {
      id: 'github-flow',
      name: 'GitHub OAuth Flow',
      activeVersion: 1,
      handle: 'github-oauth-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '2023-01-01T00:00:00Z',
      updatedAt: '2023-01-01T00:00:00Z',
    } as BasicFlowDefinition,
    smsOtp: {
      id: 'sms-flow',
      name: 'SMS OTP Flow',
      activeVersion: 1,
      handle: 'sms-otp-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '2023-01-01T00:00:00Z',
      updatedAt: '2023-01-01T00:00:00Z',
    } as BasicFlowDefinition,
    other: [
      {
        id: 'custom-flow',
        name: 'Custom Authentication Flow',
        activeVersion: 1,
        handle: 'custom-auth-flow',
        flowType: 'AUTHENTICATION',
        createdAt: '2023-01-01T00:00:00Z',
        updatedAt: '2023-01-01T00:00:00Z',
      } as BasicFlowDefinition,
    ],
  };

  const defaultProps: IndividualMethodsToggleViewProps = {
    integrations: {
      [AuthenticatorTypes.BASIC_AUTH]: false,
    },
    availableIntegrations: mockIdentityProviders,
    flowsByType: mockFlowsByType,
    onIntegrationToggle: mockOnIntegrationToggle,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getIntegrationIcon).mockImplementation((type: string) => (
      <div data-testid={`icon-${type}`}>Mock Icon</div>
    ));
  });

  const renderComponent = (props: Partial<IndividualMethodsToggleViewProps> = {}) =>
    render(<IndividualMethodsToggleView {...defaultProps} {...props} />);

  describe('core authentication methods', () => {
    it('should render Username & Password option first', () => {
      renderComponent();

      expect(screen.getByText('Username & Password')).toBeInTheDocument();

      // Should be the first item in the list
      const listItems = screen.getAllByRole('button');
      const firstItem = listItems[0];
      expect(firstItem).toHaveTextContent('Username & Password');
    });

    it('should render Google and GitHub options', () => {
      renderComponent();

      expect(screen.getByText('Google')).toBeInTheDocument();
      expect(screen.getByText('GitHub')).toBeInTheDocument();
    });

    it('should show correct icons for each method', () => {
      renderComponent();

      // Username & Password uses UserRound icon (built-in)
      // Google and GitHub use their specific icons (built-in, not mocked)
      // Verify the built-in components are rendered
      expect(screen.getByText('Username & Password')).toBeInTheDocument();
      expect(screen.getByText('Google')).toBeInTheDocument();
      expect(screen.getByText('GitHub')).toBeInTheDocument();

      // Other providers should only be rendered if they pass the filter logic
      // Since only OAUTH and OIDC are in otherProviders, verify they are rendered
      expect(screen.getByText('OAuth Provider')).toBeInTheDocument();
      expect(screen.getByText('OIDC Provider')).toBeInTheDocument();
    });
  });

  describe('integration states', () => {
    it('should show enabled state for selected integrations', () => {
      renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'google-idp': true,
          'github-idp': false,
        },
      });

      const switches = screen.getAllByRole('switch');

      // Username & Password should be checked
      expect(switches.find((s) => s.closest('[data-testid*="basic"]') ?? s)).toBeChecked();

      // Should have switches for each integration
      expect(switches.length).toBeGreaterThan(0);
    });

    it('should call onIntegrationToggle when an integration is toggled', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Find switch for Username & Password
      const usernamePasswordSwitch = screen.getAllByRole('switch')[0]; // First switch should be Username & Password
      await user.click(usernamePasswordSwitch);

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith(AuthenticatorTypes.BASIC_AUTH);
    });

    it('should handle Google provider toggle', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Find the Google switch by finding the text first, then locating its associated switch
      const googleItem = screen.getByText('Google').closest('li');
      const googleSwitch = googleItem?.querySelector('input[type="checkbox"]');

      if (googleSwitch) {
        await user.click(googleSwitch);
      }

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith('google-idp');
    });

    it('should handle GitHub provider toggle', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Find the GitHub switch by finding the text first, then locating its associated switch
      const githubItem = screen.getByText('GitHub').closest('li');
      const githubSwitch = githubItem?.querySelector('input[type="checkbox"]');

      if (githubSwitch) {
        await user.click(githubSwitch);
      }

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith('github-idp');
    });
  });

  describe('flow integration display', () => {
    it('should show flow information when integration is enabled and has associated flow', () => {
      renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'google-idp': true,
        },
      });

      expect(screen.getByText(/Flow: Basic Authentication Flow/)).toBeInTheDocument();
      expect(screen.getByText(/Flow: Google OAuth Flow/)).toBeInTheDocument();
    });

    it('should not show flow information when integration is disabled', () => {
      renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          'google-idp': false,
        },
      });

      expect(screen.queryByText(/Flow: Basic Authentication Flow/)).not.toBeInTheDocument();
      expect(screen.queryByText(/Flow: Google OAuth Flow/)).not.toBeInTheDocument();
    });

    it('should not show flow information when no flow is available', () => {
      renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
        flowsByType: {
          ...mockFlowsByType,
          basic: null,
        },
      });

      expect(screen.queryByText(/Flow: Basic Authentication Flow/)).not.toBeInTheDocument();
    });
  });

  describe('SMS OTP integration', () => {
    it('should render SMS OTP when flow is available', () => {
      renderComponent();

      expect(screen.getByText('SMS OTP')).toBeInTheDocument();
    });

    it('should not render SMS OTP when flow is not available', () => {
      renderComponent({
        flowsByType: {
          ...mockFlowsByType,
          smsOtp: null,
        },
      });

      expect(screen.queryByText('SMS OTP')).not.toBeInTheDocument();
    });

    it('should handle SMS OTP toggle', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Find the SMS OTP switch
      const smsOtpItem = screen.getByText('SMS OTP').closest('li');
      const smsOtpSwitch = smsOtpItem?.querySelector('input[type="checkbox"]');

      if (smsOtpSwitch) {
        await user.click(smsOtpSwitch);
      }

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith('sms-otp');
    });

    it('should show flow information for SMS OTP when enabled', () => {
      renderComponent({
        integrations: {
          'sms-otp': true,
        },
      });

      expect(screen.getByText(/Flow: SMS OTP Flow/)).toBeInTheDocument();
    });
  });

  describe('other social providers', () => {
    it('should render other social providers', () => {
      renderComponent();

      expect(screen.getByText('OAuth Provider')).toBeInTheDocument();
      expect(screen.getByText('OIDC Provider')).toBeInTheDocument();
    });

    it('should use getIntegrationIcon for other providers', () => {
      renderComponent();

      expect(getIntegrationIcon).toHaveBeenCalledWith(IdentityProviderTypes.OAUTH);
      expect(getIntegrationIcon).toHaveBeenCalledWith(IdentityProviderTypes.OIDC);
    });

    it('should handle other provider toggles', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Find the OAuth Provider switch
      const oauthItem = screen.getByText('OAuth Provider').closest('li');
      const oauthSwitch = oauthItem?.querySelector('input[type="checkbox"]');

      if (oauthSwitch) {
        await user.click(oauthSwitch);
      }

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith('oauth-idp-1');
    });
  });

  describe('custom flows', () => {
    it('should render custom flows from other category', () => {
      renderComponent();

      expect(screen.getByText('Custom Authentication Flow')).toBeInTheDocument();
    });

    it('should handle custom flow toggles', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Find the Custom Authentication Flow switch
      const customFlowItem = screen.getByText('Custom Authentication Flow').closest('li');
      const customFlowSwitch = customFlowItem?.querySelector('input[type="checkbox"]');

      if (customFlowSwitch) {
        await user.click(customFlowSwitch);
      }

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith('custom-auth-flow');
    });

    it('should show flow information for enabled custom flows', () => {
      renderComponent({
        integrations: {
          'custom-auth-flow': true,
        },
      });

      expect(screen.getByText(/Flow: Custom Authentication Flow/)).toBeInTheDocument();
    });
  });

  describe('unavailable providers', () => {
    it('should handle missing Google provider gracefully', () => {
      renderComponent({
        availableIntegrations: mockIdentityProviders.filter((idp) => idp.type !== IdentityProviderTypes.GOOGLE),
      });

      const googleElement = screen.getByText('Google');
      expect(googleElement).toBeInTheDocument();

      // Should show as not configured
      expect(screen.getByText('Not configured')).toBeInTheDocument();
    });

    it('should handle missing GitHub provider gracefully', () => {
      renderComponent({
        availableIntegrations: mockIdentityProviders.filter((idp) => idp.type !== IdentityProviderTypes.GITHUB),
      });

      const githubElement = screen.getByText('GitHub');
      expect(githubElement).toBeInTheDocument();

      // Should show as not configured
      expect(screen.getByText('Not configured')).toBeInTheDocument();
    });

    it('should disable unavailable provider buttons', () => {
      renderComponent({
        availableIntegrations: [],
      });

      // When providers are unavailable, Google and GitHub should show "Not configured"
      expect(screen.getByText('Google')).toBeInTheDocument();
      expect(screen.getByText('GitHub')).toBeInTheDocument();
      expect(screen.getAllByText('Not configured')).toHaveLength(2);

      // The buttons should not have switches since they're unavailable
      const googleItem = screen.getByText('Google').closest('li');
      const githubItem = screen.getByText('GitHub').closest('li');

      expect(googleItem).toBeInTheDocument();
      expect(githubItem).toBeInTheDocument();

      // These should not have switches since they're unavailable
      expect(googleItem?.querySelector('input[type="checkbox"]')).toBeNull();
      expect(githubItem?.querySelector('input[type="checkbox"]')).toBeNull();

      // When unavailable, the AuthenticationMethodItem renders disabled ListItemButton
      // We can verify by checking that the "Not configured" secondary text is present
      expect(screen.getAllByText('Not configured').length).toBe(2);
    });
  });

  describe('list structure', () => {
    it('should render items with proper dividers', () => {
      renderComponent();

      const list = screen.getByRole('list');
      expect(list).toBeInTheDocument();

      // Should have dividers between items (MUI Divider components)
      const dividers = list.querySelectorAll('hr, .MuiDivider-root');
      expect(dividers.length).toBeGreaterThan(0);
    });

    it('should apply proper styling to the list container', () => {
      renderComponent();

      const list = screen.getByRole('list');
      expect(list).toHaveClass('MuiList-root');
    });
  });

  describe('edge cases', () => {
    it('should handle empty integrations object', () => {
      renderComponent({
        integrations: {},
      });

      // Should not crash and should render all options as unchecked
      expect(screen.getByText('Username & Password')).toBeInTheDocument();
      expect(screen.getByText('Google')).toBeInTheDocument();
      expect(screen.getByText('GitHub')).toBeInTheDocument();
    });

    it('should handle empty available integrations array', () => {
      renderComponent({
        availableIntegrations: [],
      });

      // Should still render username/password, Google, and GitHub
      expect(screen.getByText('Username & Password')).toBeInTheDocument();
      expect(screen.getByText('Google')).toBeInTheDocument();
      expect(screen.getByText('GitHub')).toBeInTheDocument();

      // But should show them as not configured
      const notConfiguredElements = screen.getAllByText('Not configured');
      expect(notConfiguredElements.length).toBe(2); // Google and GitHub
    });

    it('should handle flows with missing or null values', () => {
      renderComponent({
        flowsByType: {
          basic: null,
          google: null,
          github: null,
          smsOtp: null,
          other: [],
        },
      });

      // Should still render basic options
      expect(screen.getByText('Username & Password')).toBeInTheDocument();
      expect(screen.getByText('Google')).toBeInTheDocument();
      expect(screen.getByText('GitHub')).toBeInTheDocument();

      // Should not render SMS OTP or other flows
      expect(screen.queryByText('SMS OTP')).not.toBeInTheDocument();
    });

    it('should handle providers with special characters in names', () => {
      const specialProviders: IdentityProvider[] = [
        {
          id: 'special-provider',
          name: 'Provider with Special & Characters',
          type: IdentityProviderTypes.OAUTH,
          description: 'Special provider',
        },
      ];

      renderComponent({
        availableIntegrations: [...mockIdentityProviders, ...specialProviders],
      });

      expect(screen.getByText('Provider with Special & Characters')).toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('should have proper list structure for screen readers', () => {
      renderComponent();

      const list = screen.getByRole('list');
      expect(list).toBeInTheDocument();

      const listItems = screen.getAllByRole('button');
      expect(listItems.length).toBeGreaterThan(0);
    });

    it('should have keyboard navigation support', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Focus first item
      await user.tab();
      const firstButton = screen.getAllByRole('button')[0];
      expect(firstButton).toHaveFocus();

      // Should be able to activate with Enter or Space
      await user.keyboard('{Enter}');
      expect(mockOnIntegrationToggle).toHaveBeenCalled();
    });
  });
});
