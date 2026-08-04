// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen} from '@testing-library/react';
import type {Application, OAuth2Config} from '@thunderid/configure-applications';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import getIntegrationGuidesForTemplate from '../../../../utils/getIntegrationGuidesForTemplate';
import IntegrationGuide from '../IntegrationGuide';
import IntegrationGuides from '../IntegrationGuides';

// Mock the integration guide utility
vi.mock('../../../../utils/getIntegrationGuidesForTemplate', () => ({
  default: vi.fn(),
}));

// Mock the child component
vi.mock('../IntegrationGuide', () => ({
  default: vi.fn(() => <div data-testid="integration-guide">Integration Guide</div>),
}));

const mockApplication: Application = {
  id: 'app-123',
  name: 'Test Application',
  template: 'react',
  description: 'Test description',
  allowedUserTypes: ['admin', 'user'],
};

const mockOAuth2Config: OAuth2Config = {
  clientId: 'client-123',
  clientSecret: 'secret-456',
  grantTypes: ['authorization_code'],
  responseTypes: ['code'],
  pkceRequired: true,
  publicClient: false,
  redirectUris: ['https://example.com/callback'],
};

const mockIntegrationGuides = {
  INBUILT: {
    llm_prompt: {
      id: 'llm-1',
      title: 'Use AI to integrate',
      description: 'Copy prompt for AI',
      type: 'llm' as const,
      icon: 'sparkles',
      content: 'LLM prompt content',
    },
    manual_steps: [
      {
        step: 1,
        title: 'Install dependencies',
        description: 'Install required packages',
        code: {
          language: 'bash',
          content: 'npm install',
        },
      },
    ],
  },
};

describe('IntegrationGuides', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render integration guide when guides are available', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} oauth2Config={mockOAuth2Config} />);

      expect(screen.getByTestId('integration-guide')).toBeInTheDocument();
    });

    it('should render fallback message when no guides are available', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(null);

      render(<IntegrationGuides application={mockApplication} />);

      expect(screen.getByText('No integration guides available for this application type.')).toBeInTheDocument();
    });
  });

  describe('Props Propagation', () => {
    it('should pass clientId from oauth2Config to IntegrationGuide', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} oauth2Config={mockOAuth2Config} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: 'client-123',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });

    it('should pass empty clientId when oauth2Config is not provided', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: '',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });

    it('should pass applicationId to IntegrationGuide', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: '',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });

    it('should pass integrationGuides to IntegrationGuide', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: '',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });

    it('should pass templateId from application to IntegrationGuide', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(IntegrationGuide).toHaveBeenCalledWith(
        {
          clientId: '',
          applicationId: 'app-123',
          integrationGuides: mockIntegrationGuides,
          templateId: 'react',
        },
        undefined,
      );
    });
  });

  describe('Template Utility Integration', () => {
    it('should call getIntegrationGuidesForTemplate with template from application', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(mockIntegrationGuides);

      render(<IntegrationGuides application={mockApplication} />);

      expect(getIntegrationGuidesForTemplate).toHaveBeenCalledWith('react');
    });

    it('should call getIntegrationGuidesForTemplate with empty string when template is not defined', () => {
      vi.mocked(getIntegrationGuidesForTemplate).mockReturnValue(null);
      const appWithoutTemplate = {...mockApplication, template: undefined};

      render(<IntegrationGuides application={appWithoutTemplate} />);

      expect(getIntegrationGuidesForTemplate).toHaveBeenCalledWith('');
    });
  });
});
