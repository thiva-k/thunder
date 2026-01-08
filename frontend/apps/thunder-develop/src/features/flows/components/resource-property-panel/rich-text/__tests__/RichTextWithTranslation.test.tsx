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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen} from '@testing-library/react';
import type {ReactNode} from 'react';
import {ValidationContext, type ValidationContextProps} from '@/features/flows/context/ValidationContext';
import Notification from '@/features/flows/models/notification';
import type {Resource} from '@/features/flows/models/resources';
import RichTextWithTranslation from '../RichTextWithTranslation';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

// Mock the RichText component
vi.mock('../RichText', () => ({
  default: ({onChange, resource, disabled, ToolbarProps, className}: {
    onChange: (value: string) => void;
    resource: Resource;
    disabled?: boolean;
    ToolbarProps?: Record<string, unknown>;
    className?: string;
  }) => (
    <button
      type="button"
      aria-label="Rich text editor"
      data-testid="rich-text-component"
      data-resource-id={resource?.id}
      data-resource-label={(resource as Resource & {label?: string})?.label}
      data-disabled={disabled}
      data-classname={className}
      data-toolbar-props={JSON.stringify(ToolbarProps)}
      onClick={() => onChange('test-content')}
    />
  ),
}));

// Mock I18nConfigurationCard
vi.mock('../../I18nConfigurationCard', () => ({
  default: ({open, onClose, onChange, i18nKey}: {
    open: boolean;
    onClose: () => void;
    onChange: (key: string) => void;
    i18nKey: string;
  }) => (
    open ? (
      <div data-testid="i18n-config-card" data-i18n-key={i18nKey}>
        <button type="button" onClick={onClose} data-testid="close-i18n-card">Close</button>
        <button type="button" onClick={() => onChange('test.key')} data-testid="change-i18n-key">Change Key</button>
      </div>
    ) : null
  ),
}));

describe('RichTextWithTranslation', () => {
  const mockOnChange = vi.fn();

  const createMockResource = (overrides: Partial<Resource & {label?: string}> = {}): Resource => ({
    id: 'resource-1',
    resourceType: 'ELEMENT',
    type: 'RICH_TEXT',
    category: 'DISPLAY',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: 'Test Rich Text',
      image: '',
      showOnResourcePanel: true,
    },
    config: {
      field: {name: 'richText', type: 'RICH_TEXT'},
      styles: {},
    },
    ...overrides,
  } as unknown as Resource);

  const createMockNotification = (overrides: Partial<{
    hasResourceFieldNotification: (key: string) => boolean;
    getResourceFieldNotification: (key: string) => string;
  }> = {}): Notification => {
    const notification = new Notification(
      'notification-1',
      'Test notification',
      'error',
    );

    if (overrides.hasResourceFieldNotification) {
      notification.hasResourceFieldNotification = overrides.hasResourceFieldNotification;
    }
    if (overrides.getResourceFieldNotification) {
      notification.getResourceFieldNotification = overrides.getResourceFieldNotification;
    }

    return notification;
  };

  const createValidationContext = (
    selectedNotification: Notification | null = null,
  ): ValidationContextProps => ({
    isValid: true,
    notifications: [],
    selectedNotification,
    setSelectedNotification: vi.fn(),
    getNotification: vi.fn(),
  });

  const createWrapper = (validationContext: ValidationContextProps = createValidationContext()) => {
    function Wrapper({children}: {children: ReactNode}) {
      return (
        <ValidationContext.Provider value={validationContext}>
          {children}
        </ValidationContext.Provider>
      );
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the RichText component', () => {
      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-component')).toBeInTheDocument();
    });

    it('should pass resource to RichText component', () => {
      const resource = createMockResource({id: 'test-resource-id'});
      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={resource} />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-component')).toHaveAttribute('data-resource-id', 'test-resource-id');
    });

    it('should pass className to RichText component', () => {
      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} className="custom-class" />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-component')).toHaveAttribute('data-classname', 'custom-class');
    });

    it('should pass ToolbarProps to RichText component', () => {
      const toolbarProps = {bold: false, italic: true};
      render(
        <RichTextWithTranslation
          onChange={mockOnChange}
          resource={createMockResource()}
          ToolbarProps={toolbarProps}
        />,
        {wrapper: createWrapper()},
      );

      const richText = screen.getByTestId('rich-text-component');
      expect(richText).toHaveAttribute('data-toolbar-props', JSON.stringify(toolbarProps));
    });
  });

  describe('Error Handling', () => {
    it('should not show error message when there is no validation notification', () => {
      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper()},
      );

      expect(screen.queryByText(/error/i)).not.toBeInTheDocument();
    });

    it('should show error message when validation notification exists for resource field', () => {
      const notification = createMockNotification({
        hasResourceFieldNotification: (key: string) => key === 'resource-1_text',
        getResourceFieldNotification: () => 'This field has an error',
      });

      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper(createValidationContext(notification))},
      );

      expect(screen.getByText('This field has an error')).toBeInTheDocument();
    });

    it('should not show error message when notification exists but for different field', () => {
      const notification = createMockNotification({
        hasResourceFieldNotification: (key: string) => key === 'other-resource_text',
        getResourceFieldNotification: () => 'Other error',
      });

      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper(createValidationContext(notification))},
      );

      expect(screen.queryByText('Other error')).not.toBeInTheDocument();
    });

    it('should use correct key format for field notification check', () => {
      const hasResourceFieldNotification = vi.fn().mockReturnValue(false);
      const notification = createMockNotification({
        hasResourceFieldNotification,
        getResourceFieldNotification: () => '',
      });

      render(
        <RichTextWithTranslation
          onChange={mockOnChange}
          resource={createMockResource({id: 'my-resource'})}
        />,
        {wrapper: createWrapper(createValidationContext(notification))},
      );

      expect(hasResourceFieldNotification).toHaveBeenCalledWith('my-resource_text');
    });
  });

  describe('Default Props', () => {
    it('should default className to empty string', () => {
      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-component')).toHaveAttribute('data-classname', '');
    });

    it('should default ToolbarProps to empty object', () => {
      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-component')).toHaveAttribute('data-toolbar-props', '{}');
    });
  });

  describe('Edge Cases', () => {
    it('should handle resource without id gracefully', () => {
      const resource = createMockResource({id: undefined as unknown as string});
      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={resource} />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-component')).toBeInTheDocument();
    });

    it('should handle null selectedNotification', () => {
      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper(createValidationContext(null))},
      );

      expect(screen.getByTestId('rich-text-component')).toBeInTheDocument();
    });
  });

  describe('onChange Callback', () => {
    it('should call onChange when RichText content changes', () => {
      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper()},
      );

      const richText = screen.getByTestId('rich-text-component');
      richText.click();

      expect(mockOnChange).toHaveBeenCalledWith('test-content');
    });
  });

  describe('Resource Label Handling', () => {
    it('should pass resource label to RichText component', () => {
      const resource = createMockResource();
      (resource as Resource & {label?: string}).label = 'Test Label Content';

      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={resource} />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-component')).toHaveAttribute(
        'data-resource-label',
        'Test Label Content',
      );
    });

    it('should handle resource without label', () => {
      const resource = createMockResource();

      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={resource} />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-component')).toBeInTheDocument();
    });

    it('should handle i18n formatted label', () => {
      const resource = createMockResource();
      (resource as Resource & {label?: string}).label = '{{t(common.greeting)}}';

      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={resource} />,
        {wrapper: createWrapper()},
      );

      expect(screen.getByTestId('rich-text-component')).toHaveAttribute(
        'data-resource-label',
        '{{t(common.greeting)}}',
      );
    });
  });

  describe('Error Message Display', () => {
    it('should display error message with correct styling', () => {
      const notification = createMockNotification({
        hasResourceFieldNotification: (key: string) => key === 'resource-1_text',
        getResourceFieldNotification: () => 'Validation error message',
      });

      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper(createValidationContext(notification))},
      );

      const errorMessage = screen.getByText('Validation error message');
      expect(errorMessage).toBeInTheDocument();
    });

    it('should not display empty error message', () => {
      const notification = createMockNotification({
        hasResourceFieldNotification: () => false,
        getResourceFieldNotification: () => '',
      });

      render(
        <RichTextWithTranslation onChange={mockOnChange} resource={createMockResource()} />,
        {wrapper: createWrapper(createValidationContext(notification))},
      );

      // Should only have the rich text component, no error helper text
      expect(screen.queryByRole('paragraph')).not.toBeInTheDocument();
    });
  });
});
