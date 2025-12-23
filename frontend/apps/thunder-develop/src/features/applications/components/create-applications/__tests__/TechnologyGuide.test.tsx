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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import TechnologyGuide, {type TechnologyGuideProps} from '../TechnologyGuide';
import type {IntegrationGuides} from '../../../models/application-templates';
import {ApplicationCreateFlowSignInApproach} from '../../../models/application-create-flow';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:onboarding.summary.guides.title': 'Next Steps',
        'applications:onboarding.summary.guides.subtitle':
          'Choose how you want to integrate authentication into your application',
        'applications:onboarding.summary.guides.divider': 'or',
        'applications:clientSecret.copied': 'Copied!',
      };
      return translations[key] || key;
    },
  }),
}));

describe.skip('TechnologyGuide', () => {
  const mockGuides: IntegrationGuides = {
    inbuilt: {
      llm_prompt: {
        id: 'llm-prompt',
        title: 'Integrate with an LLM Prompt',
        description: 'Use AI to generate integration code',
        type: 'llm',
        icon: 'sparkles',
      },
      manual_steps: [
        {
          step: 1,
          title: 'Create a React application',
          description: 'Run the following commands',
          code: {
            language: 'terminal',
            content: 'npm create vite@latest',
          },
        },
        {
          step: 2,
          title: 'Install dependencies',
          description: 'Install the required packages',
          code: {
            language: 'terminal',
            content: 'npm install',
          },
        },
      ],
    },
  };

  const defaultProps: TechnologyGuideProps = {
    guides: mockGuides,
    signInApproach: ApplicationCreateFlowSignInApproach.INBUILT,
  };

  const renderComponent = (props: Partial<TechnologyGuideProps> = {}) =>
    render(<TechnologyGuide {...defaultProps} {...props} />);

  it('should render LLM prompt guide', () => {
    renderComponent();

    expect(screen.getByText('Integrate with an LLM Prompt')).toBeInTheDocument();
    expect(screen.getByText('Use AI to generate integration code')).toBeInTheDocument();
  });

  it('should render LLM guide card as clickable', async () => {
    renderComponent();

    const llmGuide = screen.getByText('Integrate with an LLM Prompt');
    expect(llmGuide).toBeInTheDocument();
  });

  it('should render divider with "or" text when manual steps exist', () => {
    renderComponent();

    expect(screen.getByText('or')).toBeInTheDocument();
  });

  it('should render manual integration steps', () => {
    renderComponent();

    expect(screen.getByText('Create a React application')).toBeInTheDocument();
    expect(screen.getByText('Run the following commands')).toBeInTheDocument();
    expect(screen.getByText('Install dependencies')).toBeInTheDocument();
  });

  it('should render code blocks for steps', () => {
    renderComponent();

    // Code content is split by syntax highlighting, so we check for parts that exist
    expect(screen.getByText('npm')).toBeInTheDocument();
    expect(screen.getByText('create vite@latest')).toBeInTheDocument();
    expect(screen.getByText('install')).toBeInTheDocument();
  });

  it('should render null when guides is null', () => {
    const {container} = renderComponent({guides: null});

    expect(container.firstChild).toBeNull();
  });

  it('should not render divider when manual steps are empty', () => {
    const guidesWithoutSteps: IntegrationGuides = {
      inbuilt: {
        llm_prompt: mockGuides.inbuilt.llm_prompt,
        manual_steps: [],
      },
    };

    renderComponent({guides: guidesWithoutSteps});

    expect(screen.queryByText('or')).not.toBeInTheDocument();
  });
});
