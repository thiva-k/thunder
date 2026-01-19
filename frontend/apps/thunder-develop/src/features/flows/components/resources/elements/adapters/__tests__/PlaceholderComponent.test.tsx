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
import PlaceholderComponent from '../PlaceholderComponent';

// Mock the SCSS file
vi.mock('../PlaceholderComponent.scss', () => ({}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated_${key}`,
  }),
}));

describe('PlaceholderComponent', () => {
  describe('Regular Text Rendering', () => {
    it('should render plain text value in a span', () => {
      render(<PlaceholderComponent value="Hello World" />);

      expect(screen.getByText('Hello World')).toBeInTheDocument();
    });

    it('should render empty string', () => {
      const {container} = render(<PlaceholderComponent value="" />);

      const span = container.querySelector('span');
      expect(span).toBeInTheDocument();
      expect(span?.textContent).toBe('');
    });

    it('should render text with special characters', () => {
      render(<PlaceholderComponent value="Hello & <World>" />);

      expect(screen.getByText('Hello & <World>')).toBeInTheDocument();
    });
  });

  describe('i18n Pattern Detection', () => {
    it('should detect i18n pattern {{t(key)}}', () => {
      render(<PlaceholderComponent value="{{t(i18n.key)}}" />);

      // Should render the resolved value
      expect(screen.getByText('translated_i18n.key')).toBeInTheDocument();
    });

    it('should detect i18n pattern with simple key', () => {
      render(<PlaceholderComponent value="{{t(label)}}" />);

      expect(screen.getByText('translated_label')).toBeInTheDocument();
    });

    it('should display the resolved value for i18n pattern', () => {
      render(<PlaceholderComponent value="{{t(my.translation.key)}}" />);

      expect(screen.getByText('translated_my.translation.key')).toBeInTheDocument();
    });

    it('should detect i18n pattern with whitespace around', () => {
      render(<PlaceholderComponent value="  {{t(key)}}  " />);

      expect(screen.getByText('translated_key')).toBeInTheDocument();
    });

    it('should not detect i18n pattern for partial match', () => {
      render(<PlaceholderComponent value="text {{t(key)}} more text" />);

      // Should render the raw value since it's not a pure i18n pattern
      expect(screen.getByText('text {{t(key)}} more text')).toBeInTheDocument();
    });

    it('should not detect old i18n pattern with double braces only', () => {
      render(<PlaceholderComponent value="{{key}}" />);

      // The old {{key}} pattern should NOT be detected, so it renders as-is
      expect(screen.getByText('{{key}}')).toBeInTheDocument();
    });

    it('should not detect i18n pattern with empty key', () => {
      render(<PlaceholderComponent value="{{t()}}" />);

      // Empty key should not match
      expect(screen.getByText('{{t()}}')).toBeInTheDocument();
    });
  });

  describe('Children Rendering', () => {
    it('should render children when provided and value is not i18n pattern', () => {
      render(
        <PlaceholderComponent value="regular text">
          <div data-testid="child">Child Content</div>
        </PlaceholderComponent>,
      );

      expect(screen.getByTestId('child')).toBeInTheDocument();
      expect(screen.getByText('Child Content')).toBeInTheDocument();
    });

    it('should not render children when value is i18n pattern', () => {
      render(
        <PlaceholderComponent value="{{t(i18n.key)}}">
          <div data-testid="child">Child Content</div>
        </PlaceholderComponent>,
      );

      expect(screen.queryByTestId('child')).not.toBeInTheDocument();
      expect(screen.getByText('translated_i18n.key')).toBeInTheDocument();
    });

    it('should prefer children over value when not i18n pattern', () => {
      render(
        <PlaceholderComponent value="value text">
          <span>Children text</span>
        </PlaceholderComponent>,
      );

      expect(screen.getByText('Children text')).toBeInTheDocument();
      expect(screen.queryByText('value text')).not.toBeInTheDocument();
    });

    it('should handle null children', () => {
      render(<PlaceholderComponent value="text">{null}</PlaceholderComponent>);

      expect(screen.getByText('text')).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle undefined-like empty value', () => {
      const {container} = render(<PlaceholderComponent value="" />);

      const span = container.querySelector('span');
      expect(span).toBeInTheDocument();
      expect(span?.textContent).toBe('');
    });

    it('should handle numeric-like strings', () => {
      render(<PlaceholderComponent value="12345" />);

      expect(screen.getByText('12345')).toBeInTheDocument();
    });

    it('should handle newlines in text', () => {
      render(<PlaceholderComponent value={'Line1\nLine2'} />);

      // The component renders the value as-is, so we check for the exact text content
      expect(screen.getByText((content) => content.includes('Line1') && content.includes('Line2'))).toBeInTheDocument();
    });
  });
});
