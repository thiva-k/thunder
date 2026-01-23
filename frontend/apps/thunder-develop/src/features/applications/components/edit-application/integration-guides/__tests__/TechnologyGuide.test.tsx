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
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {LoggerProvider, LogLevel} from '@thunder/logger';
import TechnologyGuide from '../TechnologyGuide';
import type {IntegrationGuides} from '../../../../models/application-templates';

const mockIntegrationGuides: IntegrationGuides = {
  INBUILT: {
    llm_prompt: {
      id: 'llm-1',
      title: 'Use AI Assistant',
      description: 'Get AI-powered integration guidance',
      type: 'llm' as const,
      icon: 'sparkles',
      content: 'Integrate with clientId: {{clientId}} and applicationId: {{applicationId}}',
    },
    manual_steps: [
      {
        step: 1,
        title: 'Install dependencies',
        description: 'Install required packages for your application',
        subDescription: 'Run the following command in your terminal',
        bullets: ['npm for Node Package Manager', 'yarn for Yarn Package Manager'],
        code: {
          language: 'bash',
          filename: 'terminal',
          content: 'npm install @thunder/sdk',
        },
      },
      {
        step: 2,
        title: 'Configure client',
        description: 'Set up your application with the client ID',
        code: {
          language: 'typescript',
          filename: 'config.ts',
          content: 'const clientId = "{{clientId}}";',
        },
      },
    ],
  },
  EMBEDDED: {
    llm_prompt: {
      id: 'llm-2',
      title: 'Embedded Integration',
      description: 'Custom login UI integration',
      type: 'llm' as const,
      icon: 'sparkles',
      content: 'Embedded integration prompt',
    },
    manual_steps: [
      {
        step: 1,
        title: 'Setup custom UI',
        description: 'Create your custom login form',
      },
    ],
  },
};

const mockClipboard = {
  writeText: vi.fn(),
};

const renderWithProviders = (component: React.ReactElement) =>
  render(<LoggerProvider logger={{level: LogLevel.DEBUG}}>{component}</LoggerProvider>);

describe('TechnologyGuide', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('navigator', {
      clipboard: mockClipboard,
    });
    mockClipboard.writeText.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  describe('Rendering', () => {
    it('should return null when guides is null', () => {
      const {container} = renderWithProviders(<TechnologyGuide guides={null} />);

      expect(container.firstChild?.firstChild).toBeFalsy();
    });

    it('should return null when selected guide is not found', () => {
      const guidesWithoutInbuilt: IntegrationGuides = {
        OTHER: mockIntegrationGuides.INBUILT,
      };

      const {container} = renderWithProviders(<TechnologyGuide guides={guidesWithoutInbuilt} templateId="react" />);

      expect(container.firstChild?.firstChild).toBeFalsy();
    });

    it('should render inbuilt guide for non-embedded template', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      expect(screen.getByText('Use AI Assistant')).toBeInTheDocument();
      expect(screen.getByText('Get AI-powered integration guidance')).toBeInTheDocument();
    });

    it('should render embedded guide for embedded template', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react-embedded" />);

      expect(screen.getByText('Embedded Integration')).toBeInTheDocument();
    });

    it('should default to inbuilt guide when templateId is null', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId={null} />);

      expect(screen.getByText('Use AI Assistant')).toBeInTheDocument();
    });
  });

  describe('LLM Prompt Section', () => {
    it('should render LLM prompt card with title and description', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" clientId="client-123" />);

      expect(screen.getByText('Use AI Assistant')).toBeInTheDocument();
      expect(screen.getByText('Get AI-powered integration guidance')).toBeInTheDocument();
    });

    it('should render copy prompt button', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      expect(screen.getByTestId('copy-prompt-button')).toBeInTheDocument();
      expect(screen.getByText('Copy Prompt')).toBeInTheDocument();
    });
  });

  describe('Manual Steps Section', () => {
    it('should render divider with "or" text', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      expect(screen.getByText('or')).toBeInTheDocument();
    });

    it('should render all manual steps', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      expect(screen.getByText('Install dependencies')).toBeInTheDocument();
      expect(screen.getByText('Configure client')).toBeInTheDocument();
    });

    it('should render step numbers', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      expect(screen.getByText('1')).toBeInTheDocument();
      expect(screen.getByText('2')).toBeInTheDocument();
    });

    it('should render step descriptions', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      expect(screen.getByText('Install required packages for your application')).toBeInTheDocument();
      expect(screen.getByText('Set up your application with the client ID')).toBeInTheDocument();
    });

    it('should render sub-descriptions when provided', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      expect(screen.getByText('Run the following command in your terminal')).toBeInTheDocument();
    });

    it('should render bullet points when provided', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      expect(screen.getByText('npm for Node Package Manager')).toBeInTheDocument();
      expect(screen.getByText('yarn for Yarn Package Manager')).toBeInTheDocument();
    });
  });

  describe('Code Blocks', () => {
    it('should render code blocks for steps with code', () => {
      const {container} = renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      // Check that code blocks exist
      const codeBlocks = container.querySelectorAll('pre');
      expect(codeBlocks).toHaveLength(2);

      // Check code content is present
      expect(container.textContent).toContain('npm install @thunder/sdk');
      expect(container.textContent).toContain('const clientId = "{{clientId}}";');
    });

    it('should render filenames when provided', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      expect(screen.getByText('terminal')).toBeInTheDocument();
      expect(screen.getByText('config.ts')).toBeInTheDocument();
    });

    it('should render copy buttons for each code block', () => {
      renderWithProviders(<TechnologyGuide guides={mockIntegrationGuides} templateId="react" />);

      const copyButtons = screen.getAllByTestId(/copy-code-button-/);
      expect(copyButtons).toHaveLength(2);
    });
  });

  describe('Empty States', () => {
    it('should not render code block when step has no code', () => {
      const guidesWithoutCode: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [
            {
              step: 1,
              title: 'No code step',
              description: 'This step has no code',
            },
          ],
        },
      };

      const {container} = renderWithProviders(<TechnologyGuide guides={guidesWithoutCode} templateId="react" />);

      const codeBlocks = container.querySelectorAll('pre');
      expect(codeBlocks).toHaveLength(0);
    });

    it('should not render manual steps section when manual_steps is empty', () => {
      const guidesWithoutSteps: IntegrationGuides = {
        INBUILT: {
          llm_prompt: mockIntegrationGuides.INBUILT.llm_prompt,
          manual_steps: [],
        },
      };

      renderWithProviders(<TechnologyGuide guides={guidesWithoutSteps} templateId="react" />);

      expect(screen.queryByText('or')).not.toBeInTheDocument();
    });
  });
});
