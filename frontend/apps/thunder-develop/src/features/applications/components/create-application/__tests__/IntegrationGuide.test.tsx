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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import {BrowserRouter} from 'react-router';
import IntegrationGuide from '../IntegrationGuide';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

// Mock TechnologyGuide component
vi.mock('../../edit-application/integration-guides/TechnologyGuide', () => ({
  default: () => <div data-testid="technology-guide">Technology Guide</div>,
}));

describe('IntegrationGuide', () => {
  const defaultProps = {
    appName: 'Test Application',
    appLogo: 'https://example.com/logo.png',
    selectedColor: '#FF5733',
    hasOAuthConfig: false,
    applicationId: 'app-123',
  };

  const renderWithRouter = (ui: React.ReactElement) => render(<BrowserRouter>{ui}</BrowserRouter>);

  describe('Rendering', () => {
    it('should render the component', () => {
      renderWithRouter(<IntegrationGuide {...defaultProps} />);

      expect(screen.getByText('applications:onboarding.summary.title')).toBeInTheDocument();
    });

    it('should display app name', () => {
      renderWithRouter(<IntegrationGuide {...defaultProps} />);

      expect(screen.getByText(defaultProps.appName)).toBeInTheDocument();
    });

    it('should display success message when OAuth not configured', () => {
      renderWithRouter(<IntegrationGuide {...defaultProps} hasOAuthConfig={false} />);

      expect(screen.getByText('applications:onboarding.summary.subtitle')).toBeInTheDocument();
    });
  });

  describe('OAuth Configuration', () => {
    it('should display client ID when hasOAuthConfig is true', () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
      };

      renderWithRouter(<IntegrationGuide {...props} />);

      expect(screen.getByDisplayValue('test_client_id')).toBeInTheDocument();
    });

    it('should display client secret when provided', () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
        clientSecret: 'test_secret',
      };

      renderWithRouter(<IntegrationGuide {...props} />);

      // Secret should be hidden by default
      const secretInput = screen.getByDisplayValue('test_secret');
      expect(secretInput).toHaveAttribute('type', 'password');
    });

    it('should not display OAuth credentials when hasOAuthConfig is false', () => {
      renderWithRouter(<IntegrationGuide {...defaultProps} hasOAuthConfig={false} />);

      expect(screen.queryByText('applications:create.integrationGuide.clientId')).not.toBeInTheDocument();
    });
  });

  describe('Integration Guides', () => {
    it('should render TechnologyGuide when integrationGuides are provided', () => {
      const props = {
        ...defaultProps,
        integrationGuides: {
          react: {
            llm_prompt: {
              id: 'test-guide',
              title: 'Test Guide',
              description: 'Test description',
              type: 'llm' as const,
              icon: 'test-icon',
              overview: 'Test overview',
              prerequisites: [],
              steps: [],
            },
            manual_steps: [],
          },
        },
      };

      renderWithRouter(<IntegrationGuide {...props} />);

      expect(screen.getByTestId('technology-guide')).toBeInTheDocument();
    });

    it('should not render TechnologyGuide when integrationGuides are null', () => {
      const props = {
        ...defaultProps,
        integrationGuides: null,
      };

      renderWithRouter(<IntegrationGuide {...props} />);

      expect(screen.queryByTestId('technology-guide')).not.toBeInTheDocument();
    });

    it('should not render TechnologyGuide when integrationGuides are undefined', () => {
      renderWithRouter(<IntegrationGuide {...defaultProps} />);

      expect(screen.queryByTestId('technology-guide')).not.toBeInTheDocument();
    });
  });

  describe('App Logo', () => {
    it('should display app logo when provided', () => {
      renderWithRouter(<IntegrationGuide {...defaultProps} />);

      const logo = screen.getByAltText('Test Application logo');
      expect(logo).toHaveAttribute('src', defaultProps.appLogo);
    });

    it('should handle null logo gracefully', () => {
      const props = {
        ...defaultProps,
        appLogo: null,
      };

      renderWithRouter(<IntegrationGuide {...props} />);

      // Should still render without crashing
      expect(screen.getByText(defaultProps.appName)).toBeInTheDocument();
    });
  });
});
