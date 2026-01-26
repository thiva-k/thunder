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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import type {ReactNode} from 'react';
import I18nConfigurationCard from '../I18nConfigurationCard';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../../context/FlowBuilderCoreContext';
import {EdgeStyleTypes} from '../../../models/steps';
import {PreviewScreenType} from '../../../models/custom-text-preference';
import {ElementTypes} from '../../../models/elements';
import type {Base} from '../../../models/base';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      if (params) {
        return `${key} ${JSON.stringify(params)}`;
      }
      return key;
    },
  }),
}));

// Mock the API hooks
const mockMutate = vi.fn();
vi.mock('../../../../../api/useSetTranslation', () => ({
  default: () => ({
    mutate: mockMutate,
    isPending: false,
  }),
}));

vi.mock('../../../../../api/useGetLanguages', () => ({
  default: () => ({
    data: {languages: ['en', 'es', 'fr']},
  }),
}));

vi.mock('../../../../../api/useGetTranslations', () => ({
  default: () => ({
    data: {
      language: 'en-US',
      translations: {
        flowI18n: {
          'login.title': 'Sign In',
          'login.description': 'Enter your credentials',
          'login.button.submit': 'Submit',
          'common.continue': 'Continue',
          'common.cancel': 'Cancel',
        },
      },
    },
    isLoading: false,
  }),
}));

describe('I18nConfigurationCard', () => {
  const mockOnClose = vi.fn();
  const mockOnChange = vi.fn();

  let anchorEl: HTMLDivElement;

  // Reset mocks before each test
  beforeEach(() => {
    mockMutate.mockReset();
  });

  const mockBaseResource: Base = {
    id: 'resource-1',
    resourceType: 'ELEMENT',
    type: 'TEXT_INPUT',
    category: 'FIELD',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: 'Test Resource',
      image: '',
      showOnResourcePanel: false,
    },
    config: {
      field: {name: '', type: ElementTypes},
      styles: {},
    },
  };

  const mockI18nText = {
    [PreviewScreenType.LOGIN]: {
      'login.title': 'Sign In',
      'login.description': 'Enter your credentials',
      'login.button.submit': 'Submit',
    },
    [PreviewScreenType.COMMON]: {
      'common.continue': 'Continue',
      'common.cancel': 'Cancel',
    },
  };

  const createContextValue = (overrides: Partial<FlowBuilderCoreContextProps> = {}): FlowBuilderCoreContextProps => ({
    lastInteractedResource: mockBaseResource,
    lastInteractedStepId: 'step-1',
    ResourceProperties: () => null,
    resourcePropertiesPanelHeading: 'Test Panel Heading',
    primaryI18nScreen: PreviewScreenType.LOGIN,
    isResourcePanelOpen: true,
    isResourcePropertiesPanelOpen: false,
    isVersionHistoryPanelOpen: false,
    ElementFactory: () => null,
    onResourceDropOnCanvas: vi.fn(),
    selectedAttributes: {},
    setLastInteractedResource: vi.fn(),
    setLastInteractedStepId: vi.fn(),
    setResourcePropertiesPanelHeading: vi.fn(),
    setIsResourcePanelOpen: vi.fn(),
    setIsOpenResourcePropertiesPanel: vi.fn(),
    registerCloseValidationPanel: vi.fn(),
    setIsVersionHistoryPanelOpen: vi.fn(),
    setSelectedAttributes: vi.fn(),
    flowCompletionConfigs: {},
    setFlowCompletionConfigs: vi.fn(),
    flowNodeTypes: {},
    flowEdgeTypes: {},
    setFlowNodeTypes: vi.fn(),
    setFlowEdgeTypes: vi.fn(),
    isVerboseMode: false,
    setIsVerboseMode: vi.fn(),
    edgeStyle: EdgeStyleTypes.SmoothStep,
    setEdgeStyle: vi.fn(),
    i18nText: mockI18nText,
    i18nTextLoading: false,
    ...overrides,
  });

  const createWrapper = (contextValue: FlowBuilderCoreContextProps = createContextValue()) =>
    function Wrapper({children}: {children: ReactNode}) {
      return <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>;
    };

  beforeEach(() => {
    anchorEl = document.createElement('div');
    document.body.appendChild(anchorEl);
    vi.clearAllMocks();
  });

  afterEach(() => {
    document.body.removeChild(anchorEl);
  });

  describe('Rendering', () => {
    it('should render the popover when open is true', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByRole('presentation')).toBeInTheDocument();
    });

    it('should not render content when open is false', () => {
      render(
        <I18nConfigurationCard
          open={false}
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.queryByText('flows:core.elements.textPropertyField.i18nCard.i18nKey')).not.toBeInTheDocument();
    });

    it('should render the card title with formatted property key', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="buttonLabel"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      expect(
        screen.getByText('flows:core.elements.textPropertyField.i18nCard.title {"field":"Button Label"}'),
      ).toBeInTheDocument();
    });

    it('should render the i18n key label', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.i18nKey')).toBeInTheDocument();
    });

    it('should render loading state when i18nTextLoading is true', () => {
      const loadingContext = createContextValue({i18nTextLoading: true});

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(loadingContext)},
      );

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });
  });

  describe('Close Functionality', () => {
    it('should call onClose when close button is clicked', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      const closeButton = screen.getByLabelText('common:close');
      fireEvent.click(closeButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  describe('Autocomplete Options', () => {
    it('should display available i18n keys in autocomplete', async () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Open the autocomplete dropdown by clicking the Open button
      const openButton = screen.getByTitle('Open');
      fireEvent.click(openButton);

      // Keys are now prefixed with the namespace (flowI18n:)
      await waitFor(() => {
        expect(screen.getByText('flowI18n:login.title')).toBeInTheDocument();
        expect(screen.getByText('flowI18n:login.description')).toBeInTheDocument();
        expect(screen.getByText('flowI18n:common.continue')).toBeInTheDocument();
      });
    });

    it('should handle empty i18nText gracefully', () => {
      const emptyContext = createContextValue({i18nText: undefined});

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(emptyContext)},
      );

      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('Selection and onChange', () => {
    it('should call onChange with selected i18n key', async () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Open the autocomplete dropdown by clicking the Open button
      const openButton = screen.getByTitle('Open');
      fireEvent.click(openButton);

      // Keys are now prefixed with the namespace (flowI18n:)
      await waitFor(() => {
        expect(screen.getByText('flowI18n:login.title')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('flowI18n:login.title'));

      expect(mockOnChange).toHaveBeenCalledWith('flowI18n:login.title');
    });

    it('should call onChange with empty string when selection is cleared', async () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper()},
      );

      const clearButton = screen.getByLabelText('Clear');
      fireEvent.click(clearButton);

      expect(mockOnChange).toHaveBeenCalledWith('');
    });
  });

  describe('Resolved Value Display', () => {
    it('should display resolved value when i18n key is selected', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByText('flows:core.elements.textPropertyField.resolvedValue')).toBeInTheDocument();
      expect(screen.getByText('Sign In')).toBeInTheDocument();
    });

    it('should not display resolved value box when no i18n key is selected', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.queryByText('flows:core.elements.textPropertyField.resolvedValue')).not.toBeInTheDocument();
    });

    it('should not display resolved value box when key has no matching value', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="nonexistent.key"
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.queryByText('flows:core.elements.textPropertyField.resolvedValue')).not.toBeInTheDocument();
    });

    it('should display resolved value from common screen', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="common.continue"
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByText('Continue')).toBeInTheDocument();
    });
  });

  describe('Selected Value Display', () => {
    it('should show selected i18n key in autocomplete', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper()},
      );

      const input: HTMLInputElement = screen.getByRole('combobox');
      expect(input.value).toBe('login.title');
    });

    it('should show empty autocomplete when i18nKey is empty string', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      const input: HTMLInputElement = screen.getByRole('combobox');
      expect(input.value).toBe('');
    });
  });

  describe('Create Translation Mode', () => {
    it('should show create translation button', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
    });

    it('should enter create mode when create button is clicked', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      const createButton = screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle');
      fireEvent.click(createButton);

      // Should show language selector in create mode
      expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.language')).toBeInTheDocument();
      expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.languageText')).toBeInTheDocument();
    });

    it('should show cancel button in create mode', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      const createButton = screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle');
      fireEvent.click(createButton);

      expect(screen.getByText('common:cancel')).toBeInTheDocument();
    });

    it('should exit create mode when cancel is clicked', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createButton = screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle');
      fireEvent.click(createButton);

      // Click cancel
      const cancelButton = screen.getByText('common:cancel');
      fireEvent.click(cancelButton);

      // Should be back in select mode - the placeholder text indicates we're in select mode
      expect(screen.getByPlaceholderText('flows:core.elements.textPropertyField.i18nCard.selectI18nKey')).toBeInTheDocument();
    });

    it('should have disabled create button when key or value is empty', () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Enter create mode
      const createModeButton = screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle');
      fireEvent.click(createModeButton);

      // The create button should be disabled when fields are empty
      const submitButton = screen.getByText('common:create');
      expect(submitButton).toBeDisabled();
    });
  });
});
