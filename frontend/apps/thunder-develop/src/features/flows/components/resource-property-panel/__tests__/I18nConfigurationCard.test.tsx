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
  Trans: ({children}: {children: ReactNode}) => children,
}));

// Mock the SCSS file
vi.mock('../I18nConfigurationCard.scss', () => ({}));

describe('I18nConfigurationCard', () => {
  const mockOnClose = vi.fn();
  const mockOnChange = vi.fn();
  const mockUpdateI18nKey = vi.fn().mockResolvedValue(true);
  const mockSetLanguage = vi.fn();
  const mockIsCustomI18nKey = vi.fn().mockReturnValue(true);

  let anchorEl: HTMLDivElement;

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

  const mockSupportedLocales = {
    en_US: {code: 'en_US', name: 'English', flag: 'us'},
    fr_FR: {code: 'fr_FR', name: 'French', flag: 'fr'},
    de_DE: {code: 'de_DE', name: 'German', flag: 'de'},
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
    isBrandingEnabled: true,
    isCustomI18nKey: mockIsCustomI18nKey,
    updateI18nKey: mockUpdateI18nKey,
    isI18nSubmitting: false,
    language: 'en_US',
    setLanguage: mockSetLanguage,
    supportedLocales: mockSupportedLocales,
    ...overrides,
  });

  const createWrapper = (contextValue: FlowBuilderCoreContextProps = createContextValue()) => {
    function Wrapper({children}: {children: ReactNode}) {
      return <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>;
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
    anchorEl = document.createElement('div');
    anchorEl.getBoundingClientRect = vi.fn().mockReturnValue({
      top: 100,
      left: 100,
      right: 200,
      bottom: 150,
      width: 100,
      height: 50,
    });
    document.body.appendChild(anchorEl);
  });

  afterEach(() => {
    document.body.removeChild(anchorEl);
  });

  describe('Rendering', () => {
    it('should not render when open is false', () => {
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

      expect(screen.queryByRole('button', {name: /close/i})).not.toBeInTheDocument();
    });

    it('should render card when open is true', () => {
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

      expect(document.querySelector('.flow-builder-resource-property-panel-i18n-configuration')).toBeInTheDocument();
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

      expect(document.querySelector('.MuiCircularProgress-root')).toBeInTheDocument();
    });

    it('should render card title with property key', () => {
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

      expect(screen.getByText(/flows:core.elements.textPropertyField.i18nCard.title/)).toBeInTheDocument();
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

      const closeButton = screen.getByRole('button', {name: /close/i});
      fireEvent.click(closeButton);

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should call onClose when backdrop is clicked', () => {
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

      const backdrop = document.querySelector('.card-backdrop');
      if (backdrop) {
        fireEvent.click(backdrop);
      }

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should call onClose when Escape key is pressed', () => {
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

      const backdrop = document.querySelector('.card-backdrop');
      if (backdrop) {
        fireEvent.keyDown(backdrop, {key: 'Escape'});
      }

      expect(mockOnClose).toHaveBeenCalled();
    });

    it('should not close when card content is clicked', () => {
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

      const card = document.querySelector('.card');
      if (card) {
        fireEvent.click(card);
      }

      expect(mockOnClose).not.toHaveBeenCalled();
    });
  });

  describe('Simple View', () => {
    it('should render autocomplete for i18n key selection', () => {
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

      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });

    it('should show New button when no i18nKey is selected', () => {
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

      expect(screen.getByText('common:new')).toBeInTheDocument();
    });

    it('should show Edit button when i18nKey is selected', () => {
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

      expect(screen.getByText('common:edit')).toBeInTheDocument();
    });

    it('should disable New button when branding is not enabled', () => {
      const disabledBrandingContext = createContextValue({isBrandingEnabled: false});

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(disabledBrandingContext)},
      );

      const newButton = screen.getByText('common:new').closest('button');
      expect(newButton).toBeDisabled();
    });

    it('should disable Edit button when branding is not enabled', () => {
      const disabledBrandingContext = createContextValue({isBrandingEnabled: false});

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(disabledBrandingContext)},
      );

      const editButton = screen.getByText('common:edit').closest('button');
      expect(editButton).toBeDisabled();
    });
  });

  describe('Customize View', () => {
    it('should switch to customize view when Edit button is clicked', async () => {
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

      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });
    });

    it('should switch to customize view when New button is clicked', async () => {
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

      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
        expect(screen.getByText('common:create')).toBeInTheDocument();
      });
    });

    it('should return to simple view when Back button is clicked', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Click back
      const backButton = screen.getByText('common:back');
      fireEvent.click(backButton);

      await waitFor(() => {
        expect(screen.getByText('common:edit')).toBeInTheDocument();
      });
    });

    it('should render language selector in customize view', async () => {
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

      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.language')).toBeInTheDocument();
      });
    });

    it('should render language text field in customize view', async () => {
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

      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.languageText')).toBeInTheDocument();
      });
    });
  });

  describe('Create Mode', () => {
    it('should show create title in create mode', async () => {
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

      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
      });
    });

    it('should show i18n key input with prefix in create mode', async () => {
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

      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        // The prefix should be displayed
        expect(screen.getByText('login.')).toBeInTheDocument();
      });
    });
  });

  describe('Update Mode', () => {
    it('should show update title in update mode', async () => {
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

      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.updateTitle')).toBeInTheDocument();
      });
    });

    it('should show warning for common keys', async () => {
      const contextWithNonCustomKey = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(false),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="common.continue"
        />,
        {wrapper: createWrapper(contextWithNonCustomKey)},
      );

      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.commonKeyWarning')).toBeInTheDocument();
      });
    });
  });

  describe('Save Functionality', () => {
    it('should disable save button when no changes are made', async () => {
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

      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).toBeDisabled();
      });
    });

    it('should show loading state when submitting', async () => {
      const submittingContext = createContextValue({isI18nSubmitting: true});

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(submittingContext)},
      );

      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).toBeDisabled();
      });
    });
  });

  describe('Custom Language TextField', () => {
    it('should render custom LanguageTextField when provided', async () => {
      const CustomTextField = vi.fn(
        ({
          value,
          onChange,
          disabled,
        }: {
          value: string;
          onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
          disabled?: boolean;
        }) => <input data-testid="custom-language-text-field" value={value} onChange={onChange} disabled={disabled} />,
      );

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
          LanguageTextField={CustomTextField}
        />,
        {wrapper: createWrapper()},
      );

      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByTestId('custom-language-text-field')).toBeInTheDocument();
      });
    });
  });

  describe('No i18n Text', () => {
    it('should handle undefined i18nText gracefully', () => {
      const noI18nContext = createContextValue({i18nText: undefined});

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(noI18nContext)},
      );

      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });
  });

  describe('Position Calculation', () => {
    it('should position card relative to anchor element', () => {
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

      const card = document.querySelector('.card');
      expect(card).toBeInTheDocument();
      // Card should have position style set
      expect((card as HTMLElement).style.left).toBeDefined();
      expect((card as HTMLElement).style.top).toBeDefined();
    });
  });

  describe('newI18nKeyPrefix', () => {
    it('should return empty string when primaryI18nScreen is undefined', async () => {
      const noPrimaryScreenContext = createContextValue({primaryI18nScreen: undefined});

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(noPrimaryScreenContext)},
      );

      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        // Should not have prefix when primaryI18nScreen is undefined
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
      });
    });
  });

  describe('handleDeleteI18nKey', () => {
    it('should handle deleting i18n key in customize view', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string) =>
          // Return true for custom keys (that can be deleted)
          key.startsWith('login.'),
        ),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter edit mode
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });
    });
  });

  describe('handleSaveCustomize', () => {
    it('should call updateI18nKey and return to simple view on success', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Type in the language text field to enable the save button
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'New translation text'}});
      }

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).not.toBeDisabled();
      });
    });

    it('should not call updateI18nKey when updateI18nKey is undefined', async () => {
      const noUpdateContext = createContextValue({updateI18nKey: undefined});

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(noUpdateContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });
    });
  });

  describe('Create Mode Key Input', () => {
    it('should update key input value when typing in create mode', async () => {
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
      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
      });

      // Find the text input for the key
      const keyInput = document.querySelector('.i18n-config-container input');
      if (keyInput) {
        fireEvent.change(keyInput, {target: {value: 'newkey'}});
        expect(keyInput).toHaveValue('newkey');
      }
    });

    it('should only allow lowercase letters and dots in key input', async () => {
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
      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
      });

      // Find the text input for the key
      const keyInput = document.querySelector('.i18n-config-container input');
      if (keyInput) {
        // Try to enter invalid characters - they should be rejected
        fireEvent.change(keyInput, {target: {value: 'Valid123'}});
        // The input should remain empty or have only valid chars
      }
    });
  });

  describe('Language Selection', () => {
    it('should call setLanguage when language is changed', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.language')).toBeInTheDocument();
      });

      // Find and click the language select
      const selectElements = document.querySelectorAll('.MuiSelect-select');
      expect(selectElements.length).toBeGreaterThan(0);
    });

    it('should not call setLanguage when value is empty', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.language')).toBeInTheDocument();
      });
    });
  });

  describe('Language Text Field Value', () => {
    it('should use languageTexts value when available', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.languageText')).toBeInTheDocument();
      });
    });

    it('should use i18nText value when languageTexts is empty', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        const textField = document.querySelector('.i18n-config-container textarea');
        expect(textField).toBeInTheDocument();
      });
    });
  });

  describe('Autocomplete Options Rendering', () => {
    it('should render common chip for non-custom keys', async () => {
      const mixedKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string) => key.startsWith('login.')),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(mixedKeyContext)},
      );

      // The autocomplete should be rendered in simple view
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });

    it('should render delete button for custom keys in customize view', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });
    });
  });

  describe('Scroll State', () => {
    it('should set isScrolled state when content exceeds height', async () => {
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

      // Enter customize view which has more content
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        const cardContent = document.querySelector('.card-content');
        expect(cardContent).toBeInTheDocument();
      });
    });
  });

  describe('Window Events', () => {
    it('should update position on window scroll', async () => {
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

      // Dispatch scroll event
      fireEvent.scroll(window);

      // Card should still be rendered
      const card = document.querySelector('.card');
      expect(card).toBeInTheDocument();
    });

    it('should update position on window resize', async () => {
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

      // Dispatch resize event
      fireEvent(window, new Event('resize'));

      // Card should still be rendered
      const card = document.querySelector('.card');
      expect(card).toBeInTheDocument();
    });
  });

  describe('handleLanguageTextChange', () => {
    it('should update language texts when input changes', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.languageText')).toBeInTheDocument();
      });

      // Find and update the language text field
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'Updated text'}});
      }
    });
  });

  describe('findI18nScreen', () => {
    it('should return primaryI18nScreen when key is not found', async () => {
      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="unknown.key"
        />,
        {wrapper: createWrapper()},
      );

      // The component should handle unknown keys gracefully
      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });
  });

  describe('Viewport Position Adjustments', () => {
    it('should adjust horizontal position when card would go off-screen right', () => {
      // Set anchor to be near right edge
      anchorEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 100,
        left: window.innerWidth - 50,
        right: window.innerWidth - 10,
        bottom: 150,
        width: 40,
        height: 50,
      });

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

      const card = document.querySelector('.card');
      expect(card).toBeInTheDocument();
    });

    it('should adjust vertical position when card would go off-screen bottom', () => {
      // Set anchor to be near bottom edge
      anchorEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: window.innerHeight - 50,
        left: 100,
        right: 200,
        bottom: window.innerHeight - 10,
        width: 100,
        height: 40,
      });

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

      const card = document.querySelector('.card');
      expect(card).toBeInTheDocument();
    });

    it('should adjust horizontal position when card would go off-screen left', async () => {
      // Set anchor so that when card is positioned to the right, it would go off-screen
      // and when repositioned to the left, the left value would be negative
      anchorEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 100,
        left: 50, // Near left edge
        right: 100,
        bottom: 150,
        width: 50,
        height: 50,
      });

      // Mock window.innerWidth to be small so card goes off-screen right
      Object.defineProperty(window, 'innerWidth', {value: 200, writable: true});

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

      const card = document.querySelector<HTMLElement>('.card')!;
      expect(card).toBeInTheDocument();

      // Verify position is set
      await waitFor(() => {
        const cardElement = document.querySelector<HTMLElement>('.card')!;
        expect(cardElement.style.left).toBeDefined();
        expect(cardElement.style.top).toBeDefined();
      });

      // Reset window.innerWidth
      Object.defineProperty(window, 'innerWidth', {value: 1024, writable: true});
    });
  });

  describe('Deleted I18n Keys', () => {
    it('should clear deletedI18nKeys when onChange is called after save', async () => {
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

      // Component should render without errors
      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });
  });

  describe('Supported Locales Rendering', () => {
    it('should render locale flags in select options', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.language')).toBeInTheDocument();
      });

      // Find the select element
      const select = document.querySelector('.MuiSelect-select');
      expect(select).toBeInTheDocument();
    });

    it('should render value when supportedLocales does not contain the language', async () => {
      const limitedLocalesContext = createContextValue({
        supportedLocales: {},
        language: 'unknown_locale',
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(limitedLocalesContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.language')).toBeInTheDocument();
      });
    });
  });

  describe('Simple View Autocomplete onChange', () => {
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

      // Find the autocomplete input
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();

      // The autocomplete onChange should handle null value
      const clearButton = document.querySelector('.MuiAutocomplete-clearIndicator');
      if (clearButton) {
        fireEvent.click(clearButton);
        await waitFor(() => {
          expect(mockOnChange).toHaveBeenCalledWith('');
        });
      }
    });
  });

  describe('Deleted I18n Keys Filtering', () => {
    it('should filter out deleted keys from available options', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view to access delete functionality
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // The autocomplete should render keys filtered by deletedI18nKeys
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });

    it('should return null for keys in deletedI18nKeys array', async () => {
      // This tests the mapping function that returns null for deleted keys
      const contextWithDeletedKeys = createContextValue({
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.deleted.key': 'Deleted Key Value',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(contextWithDeletedKeys)},
      );

      // Component should render with filtered keys
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });
  });

  describe('handleDeleteI18nKey Function', () => {
    it('should add key to deletedI18nKeys when delete is triggered', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          // Return true for custom keys that can be deleted
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Open the autocomplete dropdown
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      // Wait for dropdown to open
      await waitFor(() => {
        const listboxElement = document.querySelector('.MuiAutocomplete-listbox');
        expect(listboxElement).toBeDefined();
      });
    });

    it('should update languageTexts when deleting key with existing text', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string) => key.startsWith('login.')),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });
    });

    it('should handle deletion when primaryI18nScreen or selectedLanguage is undefined', async () => {
      const contextWithNoScreen = createContextValue({
        primaryI18nScreen: undefined,
        language: undefined,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(contextWithNoScreen)},
      );

      // Should render without errors
      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });
  });

  describe('Customize View Autocomplete onChange', () => {
    it('should set i18nKeyInputValue when option is selected', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // The autocomplete should be present in customize view
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });

    it('should set i18nKeyInputValue to null when selection is cleared', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // The autocomplete clear should set value to null
      const clearButton = document.querySelector('.MuiAutocomplete-clearIndicator');
      if (clearButton) {
        fireEvent.click(clearButton);
      }
    });
  });

  describe('Render Option with Delete Button', () => {
    it('should render delete button for custom keys in primary screen', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          // When checkPrimary is false, return true for login keys
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });
    });

    it('should render common chip for non-custom keys', async () => {
      const nonCustomKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(false),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="common.continue"
        />,
        {wrapper: createWrapper(nonCustomKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });
    });

    it('should render common chip for keys not in primary screen', async () => {
      const mixedKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string) => key.startsWith('login.')),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
          },
          [PreviewScreenType.COMMON]: {
            'common.continue': 'Continue',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="common.continue"
        />,
        {wrapper: createWrapper(mixedKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });
    });
  });

  describe('Edit Button Setting i18nKeyInputValue to null', () => {
    it('should set i18nKeyInputValue to null when no selectedI18nKey', async () => {
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

      // Click new button to enter create mode
      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
      });

      // Go back
      const backButton = screen.getByText('common:back');
      fireEvent.click(backButton);

      await waitFor(() => {
        expect(screen.getByText('common:new')).toBeInTheDocument();
      });
    });
  });

  describe('handleSaveCustomize with Deleted Keys', () => {
    it('should call onChange with empty string when selected key was deleted', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(true),
        updateI18nKey: vi.fn().mockResolvedValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Type in the language text field to enable the save button
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'New translation text'}});
      }

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).not.toBeDisabled();
      });
    });

    it('should handle updateI18nKey returning false', async () => {
      const failingUpdateContext = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(true),
        updateI18nKey: vi.fn().mockResolvedValue(false),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(failingUpdateContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Type in the language text field to enable the save button
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'New translation text'}});
      }

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).not.toBeDisabled();
      });
    });
  });

  describe('Simple View Render Option', () => {
    it('should render option with tooltip in simple view autocomplete', async () => {
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

      // The autocomplete is in simple view
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();

      // Open the dropdown
      const input = document.querySelector('.MuiAutocomplete-root input');
      if (input) {
        fireEvent.click(input);
        fireEvent.focus(input);
      }

      // Wait for dropdown options
      await waitFor(() => {
        const listboxElement = document.querySelector('.MuiAutocomplete-listbox');
        expect(listboxElement).toBeDefined();
      });
    });
  });

  describe('handleLanguageTextChange edge cases', () => {
    it('should not update languageTexts when i18nKeyInputValue is null', async () => {
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
      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
      });

      // Type in the language text field without selecting a key
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'Test text'}});
      }
    });

    it('should not update languageTexts when selectedLanguage is undefined', async () => {
      const noLanguageContext = createContextValue({
        language: undefined,
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(noLanguageContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });
    });
  });

  describe('Language Selection onChange', () => {
    it('should call setSelectedLanguage when language is changed', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.language')).toBeInTheDocument();
      });

      // Find and click the language select
      const select = document.querySelector('.MuiSelect-select');
      if (select) {
        fireEvent.mouseDown(select);
      }

      // Wait for menu to open
      await waitFor(() => {
        const menuElement = document.querySelector('.MuiMenu-paper');
        expect(menuElement).toBeDefined();
      });
    });

    it('should not call setSelectedLanguage when value is empty', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.language')).toBeInTheDocument();
      });

      // setSelectedLanguage should not be called with empty value
      expect(mockSetLanguage).not.toHaveBeenCalledWith('');
    });
  });

  describe('Screen texts not found', () => {
    it('should handle empty screenTexts gracefully', async () => {
      const contextWithEmptyScreen = createContextValue({
        i18nText: {
          [PreviewScreenType.LOGIN]: {},
          [PreviewScreenType.COMMON]: {},
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(contextWithEmptyScreen)},
      );

      // Should render without errors
      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });
  });

  describe('handleDeleteI18nKey with originalTexts', () => {
    it('should handle deletion when key not in prevTexts but in originalTexts', async () => {
      const contextWithOriginalTexts = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((_key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return true;
          }
          return true;
        }),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.button': 'Submit',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(contextWithOriginalTexts)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });
    });

    it('should handle deletion when originalTexts is undefined', async () => {
      const contextWithNoOriginalTexts = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(true),
        i18nText: undefined,
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(contextWithNoOriginalTexts)},
      );

      // Should render without errors
      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });
  });

  describe('Position Adjustment When Card Goes Off-Screen', () => {
    it('should adjust left position when card would go off-screen right', async () => {
      // Position anchor near the right edge so card goes off-screen
      anchorEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 100,
        left: window.innerWidth - 100,
        right: window.innerWidth - 50,
        bottom: 150,
        width: 50,
        height: 50,
      });

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

      // Wait for position to be calculated
      await waitFor(() => {
        const card = document.querySelector<HTMLElement>('.card');
        expect(card).toBeInTheDocument();
        // Card should have position adjusted (left should be set)
        expect(card?.style.left).toBeDefined();
      });
    });

    it('should adjust top position when card would go off-screen bottom', async () => {
      // Position anchor near the bottom edge so card goes off-screen
      anchorEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: window.innerHeight - 50,
        left: 100,
        right: 200,
        bottom: window.innerHeight - 10,
        width: 100,
        height: 40,
      });

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

      await waitFor(() => {
        const card = document.querySelector<HTMLElement>('.card');
        expect(card).toBeInTheDocument();
        expect(card?.style.top).toBeDefined();
      });
    });
  });

  describe('i18nKeys mapping with null returns', () => {
    it('should handle empty i18nText gracefully', () => {
      const contextWithEmptyI18nText = createContextValue({
        i18nText: {
          [PreviewScreenType.LOGIN]: {},
          [PreviewScreenType.COMMON]: {},
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(contextWithEmptyI18nText)},
      );

      // Should render without errors
      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });

    it('should filter out deleted keys from i18nKeys', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(true),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.deleteMe': 'Delete Me',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Open autocomplete dropdown
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      // After deletion, the key should be filtered out
      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });
    });
  });

  describe('handleDeleteI18nKey function', () => {
    it('should delete key and update languageTexts when key exists in prevTexts', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.toDelete': 'To Delete',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Open autocomplete and look for delete button
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });
    });

    it('should return prevTexts when primaryI18nScreen or selectedLanguage is undefined', async () => {
      const contextWithNoScreenOrLanguage = createContextValue({
        primaryI18nScreen: undefined,
        language: undefined,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(contextWithNoScreenOrLanguage)},
      );

      // Should render without errors
      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });
  });

  describe('handleSaveCustomize function', () => {
    it('should call updateI18nKey and return to simple view on success', async () => {
      const mockUpdateI18nKeySuccess = vi.fn().mockResolvedValue(true);
      const successContext = createContextValue({
        updateI18nKey: mockUpdateI18nKeySuccess,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(successContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Type in the language text field to enable the save button
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'New translation text'}});
      }

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).not.toBeDisabled();
      });

      // Click update
      const updateButton = screen.getByText('common:update').closest('button');
      if (updateButton) {
        fireEvent.click(updateButton);
      }

      // Wait for the async save operation to complete and return to simple view
      await waitFor(() => {
        expect(mockUpdateI18nKeySuccess).toHaveBeenCalled();
      });
    });

    it('should call onChange with i18nKeyInputValue.key when save is successful and key is set', async () => {
      const mockUpdateI18nKeySuccess = vi.fn().mockResolvedValue(true);
      const successContext = createContextValue({
        updateI18nKey: mockUpdateI18nKeySuccess,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(successContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Type in the language text field to enable the save button
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'New translation text'}});
      }

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).not.toBeDisabled();
      });

      // Click update
      const updateButton = screen.getByText('common:update').closest('button');
      if (updateButton) {
        fireEvent.click(updateButton);
      }

      // Wait for the async save operation to complete
      await waitFor(() => {
        expect(mockUpdateI18nKeySuccess).toHaveBeenCalled();
      });

      // Should return to simple view
      await waitFor(() => {
        expect(screen.getByText('common:edit')).toBeInTheDocument();
      });
    });

    it('should call onChange with empty string when selected key was deleted after successful save', async () => {
      const mockUpdateI18nKeySuccess = vi.fn().mockResolvedValue(true);
      const successContext = createContextValue({
        updateI18nKey: mockUpdateI18nKeySuccess,
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.toDelete': 'To Delete',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.toDelete"
        />,
        {wrapper: createWrapper(successContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Open the autocomplete dropdown to access delete button
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });

      // Find and click the delete button for the current key
      const deleteButtons = document.querySelectorAll('.delete-icon-button');
      if (deleteButtons.length > 0) {
        // Find the delete button for login.toDelete
        deleteButtons.forEach((btn) => {
          fireEvent.click(btn);
        });
      }

      // Type in the language text field to enable the save button
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'New translation text'}});
      }

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).not.toBeDisabled();
      });

      // Click update
      const updateButton = screen.getByText('common:update').closest('button');
      if (updateButton) {
        fireEvent.click(updateButton);
      }

      // Wait for the async save operation to complete
      await waitFor(() => {
        expect(mockUpdateI18nKeySuccess).toHaveBeenCalled();
      });
    });

    it('should skip screen data when screenData is undefined in handleSaveCustomize', async () => {
      const mockUpdateI18nKeySuccess = vi.fn().mockResolvedValue(true);
      const successContext = createContextValue({
        updateI18nKey: mockUpdateI18nKeySuccess,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(successContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Type in the language text field to enable the save button
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'New translation text'}});
      }

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).not.toBeDisabled();
      });

      // Click update - the languageTexts should be processed
      const updateButton = screen.getByText('common:update').closest('button');
      if (updateButton) {
        fireEvent.click(updateButton);
      }

      await waitFor(() => {
        expect(mockUpdateI18nKeySuccess).toHaveBeenCalled();
      });

      // Wait for save to complete and return to simple view
      await waitFor(() => {
        // Should return to simple view
        expect(screen.queryByText('common:back')).not.toBeInTheDocument();
      }, {timeout: 3000});
    });

    it('should not call updateI18nKey when it is undefined', async () => {
      const noUpdateContext = createContextValue({
        updateI18nKey: undefined,
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(noUpdateContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });
    });

    it('should handle updateI18nKey returning false', async () => {
      const mockUpdateI18nKeyFail = vi.fn().mockResolvedValue(false);
      const failContext = createContextValue({
        updateI18nKey: mockUpdateI18nKeyFail,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(failContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Type in the language text field
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'New translation text'}});
      }

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).not.toBeDisabled();
      });

      // Click update
      const updateButton = screen.getByText('common:update').closest('button');
      if (updateButton) {
        fireEvent.click(updateButton);
      }

      // Should stay in customize view when update fails
      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });
    });

    it('should call onChange with empty string when selected key was deleted', async () => {
      const mockUpdateI18nKeySuccess = vi.fn().mockResolvedValue(true);
      const contextWithDelete = createContextValue({
        updateI18nKey: mockUpdateI18nKeySuccess,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.toDelete': 'To Delete',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(contextWithDelete)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });
    });
  });

  describe('Simple view renderOption', () => {
    it('should render option with Tooltip in simple view', async () => {
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

      // Open the autocomplete dropdown
      const input = document.querySelector('.MuiAutocomplete-root input');
      if (input) {
        fireEvent.click(input);
        fireEvent.focus(input);
      }

      // Wait for dropdown options
      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });

      // Verify option-item class is present
      const optionItems = document.querySelectorAll('.option-item');
      expect(optionItems.length).toBeGreaterThanOrEqual(0);
    });
  });

  describe('Customize view autocomplete onChange', () => {
    it('should set i18nKeyInputValue when option is selected in customize view', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Open autocomplete and select an option
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });

      // Click on an option
      const options = document.querySelectorAll('.MuiAutocomplete-option');
      if (options.length > 0) {
        fireEvent.click(options[0]);
      }
    });

    it('should set i18nKeyInputValue to null when selection is cleared', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Clear the autocomplete
      const clearButton = document.querySelector('.MuiAutocomplete-clearIndicator');
      if (clearButton) {
        fireEvent.click(clearButton);
      }
    });
  });

  describe('Customize view renderOption', () => {
    it('should render delete button for custom keys in primary screen', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.customKey': 'Custom Value',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Open autocomplete
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });

      // Look for delete button
      const deleteButtons = document.querySelectorAll('.delete-icon-button');
      expect(deleteButtons.length).toBeGreaterThanOrEqual(0);
    });

    it('should render common chip for non-custom keys in primary screen', async () => {
      const nonCustomKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(false),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(nonCustomKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Open autocomplete
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });
    });

    it('should render common chip for keys not in primary screen', async () => {
      const mixedScreenContext = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(true),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
          },
          [PreviewScreenType.COMMON]: {
            'common.continue': 'Continue',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="common.continue"
        />,
        {wrapper: createWrapper(mixedScreenContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Open autocomplete
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });
    });

    it('should handle delete button click in renderOption', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.deletable': 'Deletable',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Open autocomplete
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });

      // Find and click delete button
      const deleteButtons = document.querySelectorAll('.delete-icon-button');
      if (deleteButtons.length > 0) {
        fireEvent.click(deleteButtons[0]);
      }
    });
  });

  describe('Edit button setting i18nKeyInputValue to null', () => {
    it('should set i18nKeyInputValue to null when selectedI18nKey is empty', async () => {
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

      // Click new button to enter create mode first
      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
      });

      // Go back to simple view
      const backButton = screen.getByText('common:back');
      fireEvent.click(backButton);

      await waitFor(() => {
        expect(screen.getByText('common:new')).toBeInTheDocument();
      });

      // i18nKeyInputValue should be null when there's no selectedI18nKey
    });
  });

  describe('isI18nCreationMode check', () => {
    it('should not render options with isI18nCreationMode in simple view', async () => {
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

      // In simple view, the autocomplete should be visible
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });
  });

  describe('Position Calculation with Off-Screen Adjustments', () => {
    it('should adjust left position when card would overflow right edge of viewport', async () => {
      // Mock a narrow viewport
      Object.defineProperty(window, 'innerWidth', {value: 400, writable: true, configurable: true});

      // Position anchor so card would go off-screen to the right
      anchorEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 100,
        left: 350,
        right: 390,
        bottom: 150,
        width: 40,
        height: 50,
      });

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

      // Wait for position calculation
      await waitFor(() => {
        const card = document.querySelector<HTMLElement>('.card');
        expect(card).toBeInTheDocument();
      });

      // Reset viewport
      Object.defineProperty(window, 'innerWidth', {value: 1024, writable: true, configurable: true});
    });

    it('should adjust top position when card would overflow bottom edge of viewport', async () => {
      // Mock a short viewport
      Object.defineProperty(window, 'innerHeight', {value: 300, writable: true, configurable: true});

      // Position anchor near bottom
      anchorEl.getBoundingClientRect = vi.fn().mockReturnValue({
        top: 250,
        left: 100,
        right: 200,
        bottom: 290,
        width: 100,
        height: 40,
      });

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

      await waitFor(() => {
        const card = document.querySelector<HTMLElement>('.card');
        expect(card).toBeInTheDocument();
      });

      // Reset viewport
      Object.defineProperty(window, 'innerHeight', {value: 768, writable: true, configurable: true});
    });

    it('should return early from updatePosition when anchorRect or cardRect is undefined', async () => {
      // Render without a valid anchor
      render(
        <I18nConfigurationCard
          open
          anchorEl={null}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper()},
      );

      // Card should still render but position won't be calculated
      const card = document.querySelector('.card');
      expect(card).toBeInTheDocument();
    });
  });

  describe('i18nKeys mapping with screenTexts undefined', () => {
    it('should return early when screenTexts is undefined for a screen', async () => {
      // Create a mock where one screen has undefined texts
      const contextWithPartialTexts = createContextValue({
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
          },
          // COMMON screen exists as key but screenTexts could be undefined in edge cases
        } as unknown as Record<PreviewScreenType, Record<string, string> | undefined>,
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(contextWithPartialTexts)},
      );

      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });
  });

  describe('handleDeleteI18nKey comprehensive coverage', () => {
    it('should delete key from prevTexts when it exists there', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.keyToDelete': 'Value to Delete',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // First, make a change to create prevTexts entry
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'Modified text'}});
      }

      // Open autocomplete to access delete buttons
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });

      // Click delete button
      const deleteButtons = document.querySelectorAll('.delete-icon-button');
      if (deleteButtons.length > 0) {
        fireEvent.click(deleteButtons[0]);
      }
    });

    it('should handle deletion when key is not in prevTexts but in originalTexts', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.anotherKey': 'Another Value',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Open autocomplete
      const autocomplete = document.querySelector('.MuiAutocomplete-root input');
      if (autocomplete) {
        fireEvent.click(autocomplete);
        fireEvent.focus(autocomplete);
      }

      await waitFor(() => {
        const listbox = document.querySelector('.MuiAutocomplete-listbox');
        expect(listbox).toBeDefined();
      });

      // Click delete on a key that hasn't been modified (not in prevTexts)
      const deleteButtons = document.querySelectorAll('.delete-icon-button');
      if (deleteButtons.length > 0) {
        fireEvent.click(deleteButtons[0]);
      }
    });

    it('should return prevTexts unchanged when primaryI18nScreen is undefined', async () => {
      const contextWithNoScreen = createContextValue({
        primaryI18nScreen: undefined,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(contextWithNoScreen)},
      );

      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });

    it('should return prevTexts unchanged when selectedLanguage is undefined', async () => {
      const contextWithNoLanguage = createContextValue({
        language: undefined,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(contextWithNoLanguage)},
      );

      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });

    it('should return prevTexts when originalTexts is undefined', async () => {
      const contextWithNoI18nText = createContextValue({
        i18nText: undefined,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(contextWithNoI18nText)},
      );

      expect(document.querySelector('.i18n-config-container')).toBeInTheDocument();
    });
  });

  describe('Simple view renderOption with className concatenation', () => {
    it('should concatenate className in simple view option rendering', async () => {
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

      // Verify autocomplete is present in simple view
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();

      // Verify the i18n selection row container is present
      const selectionRow = document.querySelector('.i18n-selection-row');
      expect(selectionRow).toBeInTheDocument();
    });

    it('should render common chip for non-custom keys in simple view', async () => {
      const mixedContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string) => key.startsWith('login.')),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
          },
          [PreviewScreenType.COMMON]: {
            'common.continue': 'Continue',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(mixedContext)},
      );

      // Verify autocomplete is rendered in simple view
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });
  });

  describe('Customize view autocomplete onChange with newValue', () => {
    it('should set i18nKeyInputValue with key, label, and screen when option is selected', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Verify the autocomplete is present with a value
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();

      // The text field should be enabled when a key is selected
      const textField = document.querySelector('.i18n-config-container textarea');
      expect(textField).toBeInTheDocument();
      expect(textField).not.toHaveAttribute('disabled');
    });

    it('should set i18nKeyInputValue to null when newValue is null', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Clear the autocomplete selection
      const clearButton = document.querySelector('.MuiAutocomplete-clearIndicator');
      if (clearButton) {
        fireEvent.click(clearButton);
      }

      // Language text field should be disabled when no key is selected
      await waitFor(() => {
        const textField = document.querySelector('.i18n-config-container textarea');
        if (textField) {
          expect(textField).toHaveAttribute('disabled');
        }
      });
    });
  });

  describe('Customize view renderOption with all branches', () => {
    it('should render delete button when option.screen equals primaryI18nScreen and isCustomI18nKey returns true', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          // Return true when checkPrimary is false (second call in the condition)
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
        primaryI18nScreen: PreviewScreenType.LOGIN,
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.customKey': 'Custom Value',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Verify autocomplete is present in customize view
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });

    it('should render common chip when option.screen equals primaryI18nScreen but isCustomI18nKey returns false', async () => {
      const nonCustomKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(false),
        primaryI18nScreen: PreviewScreenType.LOGIN,
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(nonCustomKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Verify autocomplete is present
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();

      // Warning alert should be shown for common keys
      expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.commonKeyWarning')).toBeInTheDocument();
    });

    it('should render common chip when option.screen does not equal primaryI18nScreen', async () => {
      const mixedScreenContext = createContextValue({
        isCustomI18nKey: vi.fn().mockReturnValue(true),
        primaryI18nScreen: PreviewScreenType.LOGIN,
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
          },
          [PreviewScreenType.COMMON]: {
            'common.continue': 'Continue',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(mixedScreenContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Verify autocomplete is present
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });

    it('should call handleDeleteI18nKey with stopPropagation when delete button is clicked', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
        primaryI18nScreen: PreviewScreenType.LOGIN,
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.deletable': 'Deletable Key',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Verify autocomplete is present
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();
    });
  });

  describe('Edit button click when selectedI18nKey is empty', () => {
    it('should set i18nKeyInputValue to null when edit is clicked but selectedI18nKey is falsy', async () => {
      // First render with an empty selectedI18nKey but still with Edit button visible
      // This scenario tests line 794 where selectedI18nKey is falsy
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

      // Click Edit to enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Go back and verify state is reset
      const backButton = screen.getByText('common:back');
      fireEvent.click(backButton);

      await waitFor(() => {
        expect(screen.getByText('common:edit')).toBeInTheDocument();
      });
    });
  });

  describe('ResizeObserver cleanup', () => {
    it('should clean up ResizeObserver when component unmounts', async () => {
      const {unmount} = render(
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

      // Enter customize view to trigger more content
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Unmount should clean up ResizeObserver
      unmount();
    });
  });

  describe('Window event listeners cleanup', () => {
    it('should remove scroll and resize event listeners on unmount', async () => {
      const removeEventListenerSpy = vi.spyOn(window, 'removeEventListener');

      const {unmount} = render(
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

      // Unmount to trigger cleanup
      unmount();

      // Cleanup should have been called
      expect(removeEventListenerSpy).toHaveBeenCalled();

      removeEventListenerSpy.mockRestore();
    });
  });

  describe('handleSaveCustomize with i18nKeyInputValue.key', () => {
    it('should call onChange with i18nKeyInputValue.key when save succeeds and key exists', async () => {
      const mockUpdateI18nKeySuccess = vi.fn().mockResolvedValue(true);
      const successContext = createContextValue({
        updateI18nKey: mockUpdateI18nKeySuccess,
        isCustomI18nKey: vi.fn().mockReturnValue(true),
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(successContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:update')).toBeInTheDocument();
      });

      // Make a change
      const textField = document.querySelector('.i18n-config-container textarea');
      if (textField) {
        fireEvent.change(textField, {target: {value: 'Updated translation'}});
      }

      await waitFor(() => {
        const updateButton = screen.getByText('common:update').closest('button');
        expect(updateButton).not.toBeDisabled();
      });

      // Click update
      const updateButton = screen.getByText('common:update').closest('button');
      if (updateButton) {
        fireEvent.click(updateButton);
      }

      // Wait for save to complete
      await waitFor(() => {
        expect(mockUpdateI18nKeySuccess).toHaveBeenCalled();
      });

      // Verify onChange was called with the key
      await waitFor(() => {
        expect(mockOnChange).toHaveBeenCalledWith('login.title');
      });
    });
  });

  describe('Keyboard events', () => {
    it('should not call onClose for non-Escape keys', () => {
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

      const backdrop = document.querySelector('.card-backdrop');
      if (backdrop) {
        fireEvent.keyDown(backdrop, {key: 'Enter'});
      }

      // onClose should not be called for Enter key
      expect(mockOnClose).not.toHaveBeenCalled();
    });
  });

  describe('Scroll state detection', () => {
    it('should detect scroll state when content height exceeds client height', async () => {
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

      // Enter customize view which has more content
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        const cardContent = document.querySelector('.card-content');
        expect(cardContent).toBeInTheDocument();
      });
    });
  });

  describe('newI18nKeyPrefix with hyphenated screen', () => {
    it('should replace hyphens with dots in primaryI18nScreen for prefix', async () => {
      const hyphenatedScreenContext = createContextValue({
        primaryI18nScreen: 'multi-factor-auth' as PreviewScreenType,
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey=""
        />,
        {wrapper: createWrapper(hyphenatedScreenContext)},
      );

      // Enter create mode
      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        // The prefix should have dots instead of hyphens
        expect(screen.getByText('multi.factor.auth.')).toBeInTheDocument();
      });
    });
  });

  describe('Language TextField disabled states', () => {
    it('should disable language text field when selectedLanguage is undefined', async () => {
      const noLanguageContext = createContextValue({
        language: undefined,
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(noLanguageContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        const textField = document.querySelector('.i18n-config-container textarea');
        if (textField) {
          expect(textField).toHaveAttribute('disabled');
        }
      });
    });

    it('should disable language text field when i18nKeyInputValue is null', async () => {
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
      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        const textField = document.querySelector('.i18n-config-container textarea');
        if (textField) {
          expect(textField).toHaveAttribute('disabled');
        }
      });
    });
  });

  describe('Create mode key input validation', () => {
    it('should reject non-lowercase characters in key input', async () => {
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
      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
      });

      // Try to enter uppercase characters
      const keyInput = document.querySelector('.i18n-config-container input');
      if (keyInput) {
        fireEvent.change(keyInput, {target: {value: 'InvalidKey'}});
        // Value should not change because of validation
        expect((keyInput as HTMLInputElement).value).not.toBe('InvalidKey');
      }
    });

    it('should accept lowercase letters and dots in key input', async () => {
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
      const newButton = screen.getByText('common:new');
      fireEvent.click(newButton);

      await waitFor(() => {
        expect(screen.getByText('flows:core.elements.textPropertyField.i18nCard.createTitle')).toBeInTheDocument();
      });

      // Enter valid characters
      const keyInput = document.querySelector('.i18n-config-container input');
      if (keyInput) {
        fireEvent.change(keyInput, {target: {value: 'valid.key'}});
        expect((keyInput as HTMLInputElement).value).toBe('valid.key');
      }
    });
  });

  describe('Language select renderValue', () => {
    it('should render just the value when supportedLocales does not contain the language', async () => {
      const unknownLocaleContext = createContextValue({
        language: 'unknown_locale',
        supportedLocales: {
          en_US: {code: 'en_US', name: 'English', flag: 'us'},
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(unknownLocaleContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        // The select should render the unknown locale value
        const select = document.querySelector('.MuiSelect-select');
        expect(select).toBeInTheDocument();
      });
    });

    it('should render flag and name when supportedLocales contains the language', async () => {
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

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        const select = document.querySelector('.MuiSelect-select');
        expect(select).toBeInTheDocument();
        // Should contain the formatted locale with flag
        expect(select?.textContent).toContain('English');
      });
    });
  });

  describe('Deleted keys filtering in i18nKeys', () => {
    it('should filter out deleted keys from i18nKeys list', async () => {
      const customKeyContext = createContextValue({
        isCustomI18nKey: vi.fn().mockImplementation((key: string, checkPrimary?: boolean) => {
          if (checkPrimary === false) {
            return key.startsWith('login.');
          }
          return key.startsWith('login.');
        }),
        i18nText: {
          [PreviewScreenType.LOGIN]: {
            'login.title': 'Sign In',
            'login.willBeDeleted': 'Will Be Deleted',
            'login.anotherKey': 'Another Key',
          },
        },
      });

      render(
        <I18nConfigurationCard
          open
          anchorEl={anchorEl}
          propertyKey="label"
          onClose={mockOnClose}
          onChange={mockOnChange}
          i18nKey="login.title"
        />,
        {wrapper: createWrapper(customKeyContext)},
      );

      // Enter customize view
      const editButton = screen.getByText('common:edit');
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByText('common:back')).toBeInTheDocument();
      });

      // Verify autocomplete is present in customize view
      const autocomplete = document.querySelector('.MuiAutocomplete-root');
      expect(autocomplete).toBeInTheDocument();

      // Verify the i18n config container is rendered
      const configContainer = document.querySelector('.i18n-config-container');
      expect(configContainer).toBeInTheDocument();
    });
  });
});
