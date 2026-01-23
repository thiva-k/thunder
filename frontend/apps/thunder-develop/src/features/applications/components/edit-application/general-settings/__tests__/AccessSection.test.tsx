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
});
