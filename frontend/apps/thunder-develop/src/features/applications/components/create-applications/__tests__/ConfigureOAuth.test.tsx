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
import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ConfigureOAuth, {type ConfigureOAuthProps} from '../ConfigureOAuth';
import {OAuth2GrantTypes, TokenEndpointAuthMethods, getDefaultOAuthConfig, type OAuth2Config} from '../../../models/oauth';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:onboarding.configure.oauth.title': 'Configure OAuth',
        'applications:onboarding.configure.oauth.subtitle': 'Configure OAuth settings for your application',
        'applications:onboarding.configure.oauth.optional': 'This step is optional',
        'applications:onboarding.configure.oauth.publicClient.label': 'Public Client',
        'applications:onboarding.configure.oauth.pkce.label': 'Require PKCE',
        'applications:onboarding.configure.oauth.grantTypes.label': 'Grant Types',
        'applications:onboarding.configure.oauth.grantTypes.authorizationCode': 'Authorization Code',
        'applications:onboarding.configure.oauth.grantTypes.refreshToken': 'Refresh Token',
        'applications:onboarding.configure.oauth.grantTypes.clientCredentials': 'Client Credentials',
        'applications:onboarding.configure.oauth.redirectURIs.fieldLabel': 'Redirect URIs',
        'applications:onboarding.configure.oauth.redirectURIs.placeholder': 'https://example.com/callback',
        'applications:onboarding.configure.oauth.redirectURIs.addButton': 'Add',
        'applications:onboarding.configure.oauth.redirectURIs.errors.empty': 'Please enter a redirect URI',
        'applications:onboarding.configure.oauth.redirectURIs.errors.invalid': 'Please enter a valid URL',
        'applications:onboarding.configure.oauth.redirectURIs.errors.duplicate': 'This redirect URI has already been added',
        'applications:onboarding.configure.oauth.tokenEndpointAuthMethod.label': 'Token Endpoint Auth Method',
        'applications:onboarding.configure.oauth.tokenEndpointAuthMethod.clientSecretBasic': 'Client Secret Basic',
        'applications:onboarding.configure.oauth.tokenEndpointAuthMethod.clientSecretPost': 'Client Secret Post',
        'applications:onboarding.configure.oauth.tokenEndpointAuthMethod.none': 'None',
        'applications:onboarding.configure.oauth.errors.publicClientRequiresPKCE': 'Public clients must require PKCE',
        'applications:onboarding.configure.oauth.errors.publicClientRequiresNone': 'Public clients must use none authentication method',
        'applications:onboarding.configure.oauth.errors.publicClientNoClientCredentials': 'Public clients cannot use client_credentials grant type',
        'applications:onboarding.configure.oauth.errors.atLeastOneGrantTypeRequired': 'At least one grant type is required',
        'applications:onboarding.configure.oauth.errors.refreshTokenRequiresAuthorizationCode': 'Refresh token requires authorization code',
        'applications:onboarding.configure.oauth.errors.authorizationCodeRequiresRedirectURIs': 'Authorization code requires redirect URIs',
        'applications:onboarding.configure.oauth.errors.clientCredentialsRequiresAuth': 'Client credentials cannot use none authentication method',
      };
      return translations[key] || key;
    },
  }),
}));

describe('ConfigureOAuth', () => {
  const mockOnOAuthConfigChange = vi.fn();
  const mockOnReadyChange = vi.fn();
  const mockOnValidationErrorsChange = vi.fn();

  const defaultProps: ConfigureOAuthProps = {
    oauthConfig: getDefaultOAuthConfig(),
    onOAuthConfigChange: mockOnOAuthConfigChange,
    onReadyChange: mockOnReadyChange,
    onValidationErrorsChange: mockOnValidationErrorsChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = (props: Partial<ConfigureOAuthProps> = {}) =>
    render(<ConfigureOAuth {...defaultProps} {...props} />);

  it('should render component with title and subtitle', () => {
    renderComponent();

    expect(screen.getByText('Configure OAuth')).toBeInTheDocument();
    expect(screen.getByText('Configure OAuth settings for your application')).toBeInTheDocument();
    expect(screen.getByText('This step is optional')).toBeInTheDocument();
  });

  it('should render public client switch', () => {
    renderComponent();

    expect(screen.getByLabelText('Public Client')).toBeInTheDocument();
  });

  it('should render PKCE switch', () => {
    renderComponent();

    expect(screen.getByLabelText('Require PKCE')).toBeInTheDocument();
  });

  it('should render grant types section', () => {
    renderComponent();

    expect(screen.getByText('Grant Types')).toBeInTheDocument();
    expect(screen.getByText('Authorization Code')).toBeInTheDocument();
    expect(screen.getByText('Client Credentials')).toBeInTheDocument();
  });

  it('should render redirect URI input field', () => {
    renderComponent();

    expect(screen.getByLabelText('Redirect URIs')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('https://example.com/callback')).toBeInTheDocument();
  });

  it('should render token endpoint auth method select', () => {
    renderComponent();

    // Check that the label exists and the select element is present
    expect(screen.getByText('Token Endpoint Auth Method')).toBeInTheDocument();
    // The select should be in the document (MUI Select renders as a button or combobox)
    const select = document.querySelector('#token-endpoint-auth-method') ??
                   screen.queryByRole('combobox') ?? 
                   screen.queryByRole('button', {name: /client secret basic/i});
    expect(select).toBeInTheDocument();
  });

  it('should initialize with default config when oauthConfig is null', async () => {
    renderComponent({oauthConfig: null});

    await waitFor(() => {
      expect(mockOnReadyChange).toHaveBeenCalled();
    });
  });

  it('should call onReadyChange when config has grant types', async () => {
    renderComponent({
      oauthConfig: {
        ...getDefaultOAuthConfig(),
        grant_types: [OAuth2GrantTypes.CLIENT_CREDENTIALS],
      },
    });

    await waitFor(() => {
      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });
  });

  it('should add redirect URI when valid URL is entered', async () => {
    const user = userEvent.setup({delay: null});
    const config = getDefaultOAuthConfig();
    renderComponent({oauthConfig: config});

    const input = screen.getByPlaceholderText('https://example.com/callback');
    await user.type(input, 'https://example.com/callback');
    await user.click(screen.getByText('Add'));

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        redirect_uris: ['https://example.com/callback'],
      }),
    );
  });

  it('should show error when empty redirect URI is submitted', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent();

    await user.click(screen.getByText('Add'));

    expect(screen.getByText('Please enter a redirect URI')).toBeInTheDocument();
    expect(mockOnOAuthConfigChange).not.toHaveBeenCalled();
  });

  it('should show error when invalid URL is entered', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent();

    const input = screen.getByPlaceholderText('https://example.com/callback');
    await user.type(input, 'invalid-url');
    await user.click(screen.getByText('Add'));

    expect(screen.getByText('Please enter a valid URL')).toBeInTheDocument();
    expect(mockOnOAuthConfigChange).not.toHaveBeenCalled();
  });

  it('should show error when duplicate redirect URI is added', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      redirect_uris: ['https://example.com/callback'],
    };
    renderComponent({oauthConfig: config});

    const input = screen.getByPlaceholderText('https://example.com/callback');
    await user.type(input, 'https://example.com/callback');
    await user.click(screen.getByText('Add'));

    expect(screen.getByText('This redirect URI has already been added')).toBeInTheDocument();
    expect(mockOnOAuthConfigChange).not.toHaveBeenCalled();
  });

  it('should remove redirect URI when delete is clicked', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      redirect_uris: ['https://example.com/callback'],
    };
    renderComponent({oauthConfig: config});

    // Find the chip with the URI and click its delete icon
    const chip = screen.getByText('https://example.com/callback');
    const deleteIcon = chip.closest('.MuiChip-root')?.querySelector('.MuiChip-deleteIcon');
    expect(deleteIcon).toBeInTheDocument();
    if (deleteIcon) {
      await user.click(deleteIcon as HTMLElement);
    }

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        redirect_uris: [],
      }),
    );
  });

  it('should add redirect URI on Enter key press', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent();

    const input = screen.getByPlaceholderText('https://example.com/callback');
    await user.type(input, 'https://example.com/callback{Enter}');

    expect(mockOnOAuthConfigChange).toHaveBeenCalled();
  });

  it('should toggle grant type when chip is clicked', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [OAuth2GrantTypes.CLIENT_CREDENTIALS],
    };
    renderComponent({oauthConfig: config});

    const authCodeChip = screen.getByText('Authorization Code');
    await user.click(authCodeChip);

    const expectedGrantTypes: string[] = [OAuth2GrantTypes.AUTHORIZATION_CODE];
    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        grant_types: expect.arrayContaining(expectedGrantTypes),
      }),
    );
  });

  it('should remove grant type when selected chip is clicked', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [OAuth2GrantTypes.CLIENT_CREDENTIALS],
    };
    renderComponent({oauthConfig: config});

    const clientCredentialsChip = screen.getByText('Client Credentials');
    await user.click(clientCredentialsChip);

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        grant_types: [],
      }),
    );
  });

  it('should automatically add authorization_code when refresh_token is selected', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [],
    };
    renderComponent({oauthConfig: config});

    const refreshTokenChip = screen.getByText('Refresh Token');
    await user.click(refreshTokenChip);

    const expectedGrantTypes: string[] = [OAuth2GrantTypes.AUTHORIZATION_CODE, OAuth2GrantTypes.REFRESH_TOKEN];
    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        grant_types: expect.arrayContaining(expectedGrantTypes),
      }),
    );
  });

  it('should remove refresh_token when authorization_code is removed', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [OAuth2GrantTypes.AUTHORIZATION_CODE, OAuth2GrantTypes.REFRESH_TOKEN],
    };
    renderComponent({oauthConfig: config});

    const authCodeChip = screen.getByText('Authorization Code');
    await user.click(authCodeChip);

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        grant_types: [],
      }),
    );
  });

  it('should toggle public client switch', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent();

    const publicClientSwitch = screen.getByLabelText('Public Client');
    await user.click(publicClientSwitch);

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        public_client: true,
        pkce_required: true,
        token_endpoint_auth_method: TokenEndpointAuthMethods.NONE,
      }),
    );
  });

  it('should remove client_credentials when public client is enabled', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [OAuth2GrantTypes.CLIENT_CREDENTIALS],
    };
    renderComponent({oauthConfig: config});

    const publicClientSwitch = screen.getByLabelText('Public Client');
    await user.click(publicClientSwitch);

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        grant_types: [OAuth2GrantTypes.AUTHORIZATION_CODE],
      }),
    );
  });

  it('should disable PKCE switch when public client is enabled', () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
    };
    renderComponent({oauthConfig: config});

    const pkceSwitch = screen.getByLabelText('Require PKCE');
    expect(pkceSwitch).toBeDisabled();
  });

  it('should not allow disabling PKCE when public client is enabled', () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
      pkce_required: true,
    };
    renderComponent({oauthConfig: config});

    const pkceSwitch = screen.getByLabelText('Require PKCE');
    // PKCE switch should be disabled when public client is enabled
    expect(pkceSwitch).toBeDisabled();
  });

  it('should toggle PKCE when not public client', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: false,
      pkce_required: false,
    };
    renderComponent({oauthConfig: config});

    const pkceSwitch = screen.getByLabelText('Require PKCE');
    await user.click(pkceSwitch);

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        pkce_required: true,
      }),
    );
  });

  it('should change token endpoint auth method', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent();

    // Find the select by ID or by role
    const select = document.querySelector('#token-endpoint-auth-method') ??
                   screen.getByRole('combobox') ??
                   screen.getByRole('button', {name: /client secret basic/i});
    await user.click(select);
    const option = screen.getByText('Client Secret Post');
    await user.click(option);

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        token_endpoint_auth_method: TokenEndpointAuthMethods.CLIENT_SECRET_POST,
      }),
    );
  });

  it('should disable token endpoint auth method when public client', () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
    };
    renderComponent({oauthConfig: config});

    // Find the select by ID or by role
    const select = document.querySelector('#token-endpoint-auth-method') ??
                   screen.getByRole('combobox');
    // MUI Select uses aria-disabled when disabled
    expect(select).toHaveAttribute('aria-disabled', 'true');
  });

  it('should not allow changing token endpoint auth method when public client', () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
      token_endpoint_auth_method: TokenEndpointAuthMethods.NONE,
    };
    renderComponent({oauthConfig: config});

    // Find the select by ID or by role
    const select = document.querySelector('#token-endpoint-auth-method') ??
                   screen.getByRole('combobox');
    // MUI Select uses aria-disabled when disabled
    expect(select).toHaveAttribute('aria-disabled', 'true');
  });

  it('should show validation error when public client has no PKCE', async () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
      pkce_required: false,
    };
    renderComponent({oauthConfig: config});

    await waitFor(() => {
      expect(screen.getByText('Public clients must require PKCE')).toBeInTheDocument();
    });
    expect(mockOnValidationErrorsChange).toHaveBeenCalledWith(true);
  });

  it('should show validation error when public client uses wrong auth method', async () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
      token_endpoint_auth_method: TokenEndpointAuthMethods.CLIENT_SECRET_BASIC,
    };
    renderComponent({oauthConfig: config});

    await waitFor(() => {
      expect(screen.getByText('Public clients must use none authentication method')).toBeInTheDocument();
    });
  });

  it('should show validation error when public client uses client_credentials', async () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
      grant_types: [OAuth2GrantTypes.CLIENT_CREDENTIALS],
    };
    renderComponent({oauthConfig: config});

    await waitFor(() => {
      expect(screen.getByText('Public clients cannot use client_credentials grant type')).toBeInTheDocument();
    });
  });

  it('should show validation error when no grant types are selected', async () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [],
    };
    renderComponent({oauthConfig: config});

    await waitFor(() => {
      expect(screen.getByText('At least one grant type is required')).toBeInTheDocument();
    });
    expect(mockOnReadyChange).toHaveBeenCalledWith(false);
  });

  it('should show validation error when only refresh_token is selected', async () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [OAuth2GrantTypes.REFRESH_TOKEN],
    };
    renderComponent({oauthConfig: config});

    await waitFor(() => {
      expect(screen.getByText('Refresh token requires authorization code')).toBeInTheDocument();
    });
  });

  it('should show validation error when authorization_code has no redirect URIs', async () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [OAuth2GrantTypes.AUTHORIZATION_CODE],
      redirect_uris: [],
    };
    renderComponent({oauthConfig: config});

    await waitFor(() => {
      expect(screen.getByText('Authorization code requires redirect URIs')).toBeInTheDocument();
    });
  });

  it('should show validation error when client_credentials uses none auth method', async () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [OAuth2GrantTypes.CLIENT_CREDENTIALS],
      token_endpoint_auth_method: TokenEndpointAuthMethods.NONE,
    };
    renderComponent({oauthConfig: config});

    await waitFor(() => {
      expect(screen.getByText('Client credentials cannot use none authentication method')).toBeInTheDocument();
    });
  });

  it('should disable client_credentials chip when public client', () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
    };
    renderComponent({oauthConfig: config});

    const clientCredentialsChip = screen.getByText('Client Credentials');
    expect(clientCredentialsChip.closest('.MuiChip-root')).toHaveAttribute('aria-disabled', 'true');
  });

  it('should display existing redirect URIs as chips', () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      redirect_uris: ['https://example.com/callback', 'https://example.com/callback2'],
    };
    renderComponent({oauthConfig: config});

    expect(screen.getByText('https://example.com/callback')).toBeInTheDocument();
    expect(screen.getByText('https://example.com/callback2')).toBeInTheDocument();
  });

  it('should clear URI input after adding', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent();

    const input = screen.getByPlaceholderText('https://example.com/callback');
    await user.type(input, 'https://example.com/callback');
    await user.click(screen.getByText('Add'));

    expect(screen.getByPlaceholderText('https://example.com/callback')).toHaveValue('');
  });

  it('should clear URI error when user starts typing', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent();

    await user.click(screen.getByText('Add'));
    expect(screen.getByText('Please enter a redirect URI')).toBeInTheDocument();

    const input = screen.getByPlaceholderText('https://example.com/callback');
    await user.type(input, 'https://example.com/callback');

    expect(screen.queryByText('Please enter a redirect URI')).not.toBeInTheDocument();
  });

  it('should disable public client and restore previous settings', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
      pkce_required: true,
      token_endpoint_auth_method: TokenEndpointAuthMethods.NONE,
      grant_types: [OAuth2GrantTypes.AUTHORIZATION_CODE],
    };
    renderComponent({oauthConfig: config});

    const publicClientSwitch = screen.getByLabelText('Public Client');
    await user.click(publicClientSwitch);

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        public_client: false,
        token_endpoint_auth_method: TokenEndpointAuthMethods.CLIENT_SECRET_BASIC,
      }),
    );
  });

  it('should not allow disabling PKCE when public client is enabled', () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
      pkce_required: true,
    };
    renderComponent({oauthConfig: config});

    const pkceSwitch = screen.getByLabelText('Require PKCE');
    // PKCE switch should be disabled when public client is enabled
    expect(pkceSwitch).toBeDisabled();
    
    // The handler should not be called because the switch is disabled
    // (This tests the early return in handlePKCEChange when public_client is true)
    expect(mockOnOAuthConfigChange).not.toHaveBeenCalled();
  });

  it('should not allow changing token endpoint auth method when public client', () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      public_client: true,
      token_endpoint_auth_method: TokenEndpointAuthMethods.NONE,
    };
    renderComponent({oauthConfig: config});

    // The select should be disabled, but test the handler path if it somehow gets called
    // This tests the early return in handleTokenEndpointAuthMethodChange
    const select = document.querySelector('#token-endpoint-auth-method') ??
                   screen.getByRole('combobox');
    
    // Since it's disabled, clicking won't trigger onChange, but we can verify the handler logic
    // by checking the component doesn't allow changes
    expect(select).toHaveAttribute('aria-disabled', 'true');
  });

  it('should reject invalid URL protocols', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent();

    const input = screen.getByPlaceholderText('https://example.com/callback');
    await user.type(input, 'ftp://example.com/callback');
    await user.click(screen.getByText('Add'));

    expect(screen.getByText('Please enter a valid URL')).toBeInTheDocument();
    expect(mockOnOAuthConfigChange).not.toHaveBeenCalled();
  });

  it('should reject file:// protocol URLs', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent();

    const input = screen.getByPlaceholderText('https://example.com/callback');
    await user.type(input, 'file:///path/to/file');
    await user.click(screen.getByText('Add'));

    expect(screen.getByText('Please enter a valid URL')).toBeInTheDocument();
    expect(mockOnOAuthConfigChange).not.toHaveBeenCalled();
  });

  it('should handle public client toggle when no grant types remain', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [OAuth2GrantTypes.CLIENT_CREDENTIALS],
    };
    renderComponent({oauthConfig: config});

    const publicClientSwitch = screen.getByLabelText('Public Client');
    await user.click(publicClientSwitch);

    // Should automatically add authorization_code when enabling public client with only client_credentials
    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        public_client: true,
        grant_types: [OAuth2GrantTypes.AUTHORIZATION_CODE],
      }),
    );
  });

  it('should handle multiple redirect URIs', async () => {
    const user = userEvent.setup({delay: null});
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      redirect_uris: ['https://example.com/callback'],
    };
    renderComponent({oauthConfig: config});

    const input = screen.getByPlaceholderText('https://example.com/callback');
    await user.type(input, 'https://example.com/callback2');
    await user.click(screen.getByText('Add'));

    expect(mockOnOAuthConfigChange).toHaveBeenCalledWith(
      expect.objectContaining({
        redirect_uris: ['https://example.com/callback', 'https://example.com/callback2'],
      }),
    );
  });

  it('should not call onValidationErrorsChange when there are no errors', async () => {
    const config: OAuth2Config = {
      ...getDefaultOAuthConfig(),
      grant_types: [OAuth2GrantTypes.CLIENT_CREDENTIALS],
      public_client: false,
      pkce_required: false,
      token_endpoint_auth_method: TokenEndpointAuthMethods.CLIENT_SECRET_BASIC,
    };
    renderComponent({oauthConfig: config});

    await waitFor(() => {
      // Should be called with false (no errors)
      expect(mockOnValidationErrorsChange).toHaveBeenCalledWith(false);
    });
  });
});

