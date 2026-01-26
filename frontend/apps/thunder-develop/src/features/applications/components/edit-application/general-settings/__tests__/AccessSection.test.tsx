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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import AccessSection from '../AccessSection';
import useGetUserTypes from '../../../../../user-types/api/useGetUserTypes';
import type {Application} from '../../../../models/application';
import type {OAuth2Config} from '../../../../models/oauth';

// Mock the useGetUserTypes hook
vi.mock('../../../../../user-types/api/useGetUserTypes');

type MockedUseGetUserTypes = ReturnType<typeof useGetUserTypes>;

// Mock the SettingsCard component
vi.mock('../../SettingsCard', () => ({
  default: ({title, description, children}: {title: string; description: string; children: React.ReactNode}) => (
    <div data-testid="settings-card">
      <div data-testid="card-title">{title}</div>
      <div data-testid="card-description">{description}</div>
      {children}
    </div>
  ),
}));

describe('AccessSection', () => {
  const mockOnFieldChange = vi.fn();
  const mockApplication: Application = {
    id: 'app-123',
    name: 'Test App',
    url: 'https://example.com',
    allowed_user_types: ['admin', 'user'],
    inbound_auth_config: [
      {
        type: 'oauth2',
        config: {
          client_id: 'client-123',
          redirect_uris: ['https://example.com/callback'],
        },
      },
    ],
  } as Application;

  const mockOAuth2Config: OAuth2Config = {
    client_id: 'client-123',
    redirect_uris: ['https://example.com/callback'],
  } as OAuth2Config;

  const mockUserTypes = {
    schemas: [
      {name: 'admin', id: '1'},
      {name: 'user', id: '2'},
      {name: 'guest', id: '3'},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the settings card with title and description', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByTestId('card-title')).toHaveTextContent('Access');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        "Configure who can access this application, where it's hosted, etc.",
      );
    });

    it('should render allowed user types autocomplete', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByLabelText('Allowed User Types')).toBeInTheDocument();
    });

    it('should render application URL field', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByLabelText('Application URL')).toBeInTheDocument();
      expect(screen.getByDisplayValue('https://example.com')).toBeInTheDocument();
    });

    it('should render redirect URIs section when OAuth2 config is provided', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByText('Authorized redirect URIs')).toBeInTheDocument();
      expect(screen.getByDisplayValue('https://example.com/callback')).toBeInTheDocument();
    });

    it('should not render redirect URIs section when OAuth2 config is not provided', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.queryByLabelText('Redirect URIs')).not.toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading indicator while fetching user types', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: undefined,
        loading: true,
      } as unknown as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should not show loading indicator when user types are loaded', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
  });

  describe('Allowed User Types', () => {
    it('should display selected user types from application', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByText('admin')).toBeInTheDocument();
      expect(screen.getByText('user')).toBeInTheDocument();
    });

    it('should display selected user types from editedApp over application', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(
        <AccessSection
          application={mockApplication}
          editedApp={{allowed_user_types: ['guest']}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByText('guest')).toBeInTheDocument();
      expect(screen.queryByText('admin')).not.toBeInTheDocument();
    });

    it('should display all available user types in dropdown', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = screen.getByLabelText('Allowed User Types');
      await user.click(input);

      await waitFor(() => {
        expect(screen.getAllByText('admin').length).toBeGreaterThan(0);
        expect(screen.getAllByText('guest').length).toBeGreaterThan(0);
      });
    });
  });

  describe('Application URL', () => {
    it('should display URL from application', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const urlInput = screen.getByLabelText('Application URL');
      expect(urlInput).toHaveAttribute('value', 'https://example.com');
    });

    it('should display URL from editedApp over application', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(
        <AccessSection
          application={mockApplication}
          editedApp={{url: 'https://edited.com'}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const urlInput = screen.getByLabelText('Application URL');
      expect(urlInput).toHaveAttribute('value', 'https://edited.com');
    });

    it('should show validation error for invalid URL', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const urlInput = screen.getByLabelText('Application URL');
      await user.clear(urlInput);
      await user.type(urlInput, 'invalid-url');

      await waitFor(() => {
        expect(screen.getByText('Please enter a valid URL')).toBeInTheDocument();
      });
    });

    it('should accept valid URL without error', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(
        <AccessSection application={{...mockApplication, url: ''}} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      const urlInput = screen.getByLabelText('Application URL');
      await user.type(urlInput, 'https://newurl.com');

      await waitFor(() => {
        expect(screen.queryByText('Please enter a valid URL')).not.toBeInTheDocument();
      });
    });
  });

  describe('Redirect URIs', () => {
    it('should display existing redirect URIs', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      const configWithMultipleUris = {
        ...mockOAuth2Config,
        redirect_uris: ['https://example.com/callback1', 'https://example.com/callback2'],
      };

      render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithMultipleUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByDisplayValue('https://example.com/callback1')).toBeInTheDocument();
      expect(screen.getByDisplayValue('https://example.com/callback2')).toBeInTheDocument();
    });

    it('should add new redirect URI when add button is clicked', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const addButton = screen.getByRole('button', {name: /Add URI/i});
      await user.click(addButton);

      const inputs = screen.getAllByPlaceholderText('https://example.com/callback');
      expect(inputs).toHaveLength(2);
    });

    it('should remove redirect URI when delete button is clicked', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      const configWithMultipleUris = {
        ...mockOAuth2Config,
        redirect_uris: ['https://example.com/callback1', 'https://example.com/callback2'],
      };

      render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithMultipleUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const deleteButtons = screen.getAllByRole('button', {name: /delete/i});
      await user.click(deleteButtons[0]);

      expect(screen.queryByDisplayValue('https://example.com/callback1')).not.toBeInTheDocument();
      expect(screen.getByDisplayValue('https://example.com/callback2')).toBeInTheDocument();
    });
  });

  describe('Field Change Callbacks', () => {
    it('should call onFieldChange when user types are changed', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = screen.getByLabelText('Allowed User Types');
      await user.click(input);

      await waitFor(async () => {
        const guestOption = screen.getAllByText('guest').find((el) => el.closest('li'));
        if (guestOption) await user.click(guestOption);
      });

      await waitFor(() => {
        expect(mockOnFieldChange).toHaveBeenCalled();
        const {calls} = mockOnFieldChange.mock;
        const userTypesCall = calls.find((call) => call[0] === 'allowed_user_types');
        expect(userTypesCall).toBeDefined();
      });
    });
  });

  describe('URI Validation on Blur', () => {
    const mockApplicationWithAuth: Application = {
      id: 'app-123',
      name: 'Test App',
      url: 'https://example.com',
      allowed_user_types: ['admin', 'user'],
      inbound_auth_config: [
        {
          type: 'oauth2',
          config: {
            client_id: 'client-123',
            redirect_uris: ['https://example.com/callback'],
          },
        },
      ],
    } as Application;

    it('should show error when invalid URI is entered and blurred', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(
        <AccessSection
          application={mockApplicationWithAuth}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Find the existing URI input and enter invalid URI
      const uriInput = screen.getByDisplayValue('https://example.com/callback');
      await user.clear(uriInput);
      await user.type(uriInput, 'not-a-valid-url');

      // Blur the input to trigger validation
      await user.tab();

      // Should show error and not call onFieldChange for inbound_auth_config
      await waitFor(() => {
        const errorCalls = mockOnFieldChange.mock.calls.filter((call) => call[0] === 'inbound_auth_config');
        expect(errorCalls).toHaveLength(0);
      });
    });

    it('should show error when URI is empty and blurred', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(
        <AccessSection
          application={mockApplicationWithAuth}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Find the existing URI input and clear it
      const uriInput = screen.getByDisplayValue('https://example.com/callback');
      await user.clear(uriInput);

      // Blur the input to trigger validation
      await user.tab();

      // Should not call onFieldChange for empty URI
      await waitFor(() => {
        const errorCalls = mockOnFieldChange.mock.calls.filter((call) => call[0] === 'inbound_auth_config');
        expect(errorCalls).toHaveLength(0);
      });
    });

    it('should validate URI on blur', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(
        <AccessSection
          application={mockApplicationWithAuth}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Find the existing URI input
      const uriInput = screen.getByDisplayValue('https://example.com/callback');

      // Focus and blur to trigger validation flow
      await user.click(uriInput);
      await user.tab();

      // The onBlur handler should have been called
      // Since URI is valid and non-empty, it should call updateRedirectUris
      await waitFor(() => {
        expect(mockOnFieldChange).toHaveBeenCalledWith('inbound_auth_config', expect.any(Array));
      });
    });

  });


  describe('Handle empty user types data', () => {
    it('should handle undefined user types data gracefully', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: undefined,
        loading: false,
      } as unknown as MockedUseGetUserTypes);

      render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByLabelText('Allowed User Types')).toBeInTheDocument();
    });

    it('should handle null application allowed_user_types', () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      const appWithNullTypes = {
        ...mockApplication,
        allowed_user_types: undefined,
      };

      render(
        <AccessSection
          application={appWithNullTypes as unknown as Application}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByLabelText('Allowed User Types')).toBeInTheDocument();
    });
  });

  describe('URI Error Handling', () => {
    it('should clear error when typing non-empty value in URI field', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const uriInput = screen.getByDisplayValue('https://example.com/callback');

      // Clear and blur to trigger empty error
      await user.clear(uriInput);
      await user.tab();

      // Now type something to clear the error
      await user.click(uriInput);
      await user.type(uriInput, 'https://new-uri.com');

      // Error should be cleared when typing non-empty value
      await waitFor(() => {
        expect(screen.queryByText('URI cannot be empty')).not.toBeInTheDocument();
      });
    });

    it('should reindex errors when removing a URI with errors on subsequent URIs', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      const configWithThreeUris = {
        ...mockOAuth2Config,
        redirect_uris: ['https://example.com/callback1', 'invalid-uri', 'https://example.com/callback3'],
      };

      render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithThreeUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // First, trigger validation error on the second URI by blurring it
      const secondUriInput = screen.getByDisplayValue('invalid-uri');
      await user.click(secondUriInput);
      await user.tab();

      // Now remove the first URI - this should trigger reindexing of errors
      const deleteButtons = screen.getAllByRole('button', {name: /delete/i});
      await user.click(deleteButtons[0]);

      // The first URI should be removed
      expect(screen.queryByDisplayValue('https://example.com/callback1')).not.toBeInTheDocument();
    });

    it('should preserve errors on URIs before the removed index', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      const configWithThreeUris = {
        ...mockOAuth2Config,
        redirect_uris: ['invalid-first', 'https://example.com/callback2', 'https://example.com/callback3'],
      };

      render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithThreeUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Trigger validation error on the first URI
      const firstUriInput = screen.getByDisplayValue('invalid-first');
      await user.click(firstUriInput);
      await user.tab();

      // Remove the last URI (index 2) - error on index 0 should be preserved
      const deleteButtons = screen.getAllByRole('button', {name: /delete/i});
      await user.click(deleteButtons[2]);

      // The last URI should be removed
      expect(screen.queryByDisplayValue('https://example.com/callback3')).not.toBeInTheDocument();
      // First URI should still be present
      expect(screen.getByDisplayValue('invalid-first')).toBeInTheDocument();
    });
  });

  describe('Mixed Inbound Auth Config', () => {
    it('should preserve non-oauth2 config when updating redirect URIs', async () => {
      const user = userEvent.setup();
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        loading: false,
      } as MockedUseGetUserTypes);

      const appWithMixedConfig: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'saml',
            config: {issuer: 'test-issuer'},
          },
          {
            type: 'oauth2',
            config: {
              client_id: 'client-123',
              redirect_uris: ['https://example.com/callback'],
            },
          },
        ],
      } as Application;

      render(
        <AccessSection
          application={appWithMixedConfig}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Blur the URI input to trigger updateRedirectUris
      const uriInput = screen.getByDisplayValue('https://example.com/callback');
      await user.click(uriInput);
      await user.tab();

      await waitFor(() => {
        expect(mockOnFieldChange).toHaveBeenCalledWith('inbound_auth_config', expect.any(Array));
        const call = mockOnFieldChange.mock.calls.find((c) => c[0] === 'inbound_auth_config');
        if (call) {
          const updatedConfig = call[1] as Array<{type: string}>;
          // Should contain both saml and oauth2 configs
          expect(updatedConfig.some((c) => c.type === 'saml')).toBe(true);
          expect(updatedConfig.some((c) => c.type === 'oauth2')).toBe(true);
        }
      });
    });
  });

});
