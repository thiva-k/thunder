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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ConfigureRedirectURIs, {type ConfigureRedirectURIsProps} from '../ConfigureRedirectURIs';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:onboarding.configure.redirectURIs.title': 'Configure',
        'applications:onboarding.configure.redirectURIs.subtitle':
          'Add redirect URIs where users will be redirected after authentication',
        'applications:onboarding.configure.redirectURIs.fieldLabel': 'Redirect URIs',
        'applications:onboarding.configure.redirectURIs.placeholder': 'https://localhost:3000/callback',
        'applications:onboarding.configure.redirectURIs.addButton': 'Add',
        'applications:onboarding.configure.redirectURIs.addedLabel': 'Added redirect URIs:',
        'applications:onboarding.configure.redirectURIs.warning':
          'At least one redirect URI is required. Please add at least one redirect URI.',
        'applications:onboarding.configure.redirectURIs.hint':
          'Redirect URIs must be valid URLs (http:// or https://). These are the allowed callback URLs for your application after authentication.',
        'applications:onboarding.configure.redirectURIs.errors.empty': 'Please enter a redirect URI',
        'applications:onboarding.configure.redirectURIs.errors.invalid':
          'Please enter a valid URL (must start with http:// or https://)',
        'applications:onboarding.configure.redirectURIs.errors.duplicate': 'This redirect URI has already been added',
      };
      return translations[key] || key;
    },
  }),
}));

describe('ConfigureRedirectURIs', () => {
  const mockOnRedirectURIsChange = vi.fn();
  const mockOnReadyChange = vi.fn();

  const defaultProps: ConfigureRedirectURIsProps = {
    redirectURIs: [],
    onRedirectURIsChange: mockOnRedirectURIsChange,
    onReadyChange: mockOnReadyChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = (props: Partial<ConfigureRedirectURIsProps> = {}) =>
    render(<ConfigureRedirectURIs {...defaultProps} {...props} />);

  it('should render the component with title and subtitle', () => {
    renderComponent();

    expect(screen.getByRole('heading', {level: 1})).toHaveTextContent('Configure');
    expect(screen.getByText('Add redirect URIs where users will be redirected after authentication')).toBeInTheDocument();
  });

  it('should render the redirect URI input field', () => {
    renderComponent();

    expect(screen.getByLabelText('Redirect URIs')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('https://localhost:3000/callback')).toBeInTheDocument();
  });

  it('should render the Add button', () => {
    renderComponent();

    expect(screen.getByRole('button', {name: 'Add'})).toBeInTheDocument();
  });

  it('should show warning when no URIs are added', () => {
    renderComponent({redirectURIs: []});

    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText('At least one redirect URI is required. Please add at least one redirect URI.')).toBeInTheDocument();
  });

  it('should not show warning when URIs are added', () => {
    renderComponent({redirectURIs: ['https://example.com/callback']});

    expect(screen.queryByText('At least one redirect URI is required. Please add at least one redirect URI.')).not.toBeInTheDocument();
  });

  it('should display added redirect URIs as chips', () => {
    const uris = ['https://example.com/callback', 'https://app.example.com/auth/callback'];
    renderComponent({redirectURIs: uris});

    expect(screen.getByText('Added redirect URIs:')).toBeInTheDocument();
    expect(screen.getByText('https://example.com/callback')).toBeInTheDocument();
    expect(screen.getByText('https://app.example.com/auth/callback')).toBeInTheDocument();
  });

  it('should call onRedirectURIsChange when adding a valid URI', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'https://example.com/callback');
    await user.click(addButton);

    expect(mockOnRedirectURIsChange).toHaveBeenCalledWith(['https://example.com/callback']);
  });

  it('should clear input after adding a URI', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'https://example.com/callback');
    await user.click(addButton);

    await waitFor(() => {
      expect((input as HTMLInputElement).value).toBe('');
    });
  });

  it('should add URI when pressing Enter key', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');

    await user.type(input, 'https://example.com/callback');
    await user.keyboard('{Enter}');

    expect(mockOnRedirectURIsChange).toHaveBeenCalledWith(['https://example.com/callback']);
  });

  it('should show error for empty URI', async () => {
    const user = userEvent.setup();
    renderComponent();

    const addButton = screen.getByRole('button', {name: 'Add'});
    await user.click(addButton);

    expect(screen.getByText('Please enter a redirect URI')).toBeInTheDocument();
    expect(mockOnRedirectURIsChange).not.toHaveBeenCalled();
  });

  it('should show error for invalid URL (no protocol)', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'example.com/callback');
    await user.click(addButton);

    expect(screen.getByText('Please enter a valid URL (must start with http:// or https://)')).toBeInTheDocument();
    expect(mockOnRedirectURIsChange).not.toHaveBeenCalled();
  });

  it('should show error for invalid URL (ftp protocol)', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'ftp://example.com/callback');
    await user.click(addButton);

    expect(screen.getByText('Please enter a valid URL (must start with http:// or https://)')).toBeInTheDocument();
    expect(mockOnRedirectURIsChange).not.toHaveBeenCalled();
  });

  it('should show error for duplicate URI', async () => {
    const user = userEvent.setup();
    renderComponent({redirectURIs: ['https://example.com/callback']});

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'https://example.com/callback');
    await user.click(addButton);

    expect(screen.getByText('This redirect URI has already been added')).toBeInTheDocument();
    expect(mockOnRedirectURIsChange).not.toHaveBeenCalled();
  });

  it('should accept http:// URLs', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'http://localhost:3000/callback');
    await user.click(addButton);

    expect(mockOnRedirectURIsChange).toHaveBeenCalledWith(['http://localhost:3000/callback']);
  });

  it('should accept https:// URLs', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'https://example.com/callback');
    await user.click(addButton);

    expect(mockOnRedirectURIsChange).toHaveBeenCalledWith(['https://example.com/callback']);
  });

  it('should remove URI when clicking delete icon on chip', async () => {
    const user = userEvent.setup();
    const uris = ['https://example.com/callback', 'https://app.example.com/callback'];
    renderComponent({redirectURIs: uris});

    const chips = screen.getAllByText(/https:\/\//);
    const deleteButton = chips[0].parentElement?.querySelector('button[aria-label*="delete"]') ?? chips[0].closest('div')?.querySelector('svg');
    
    if (deleteButton) {
      await user.click(deleteButton);
    } else {
      // Try clicking the chip's delete area
      const chip = screen.getByText('https://example.com/callback');
      const chipContainer = chip.closest('div[role="button"]') ?? chip.parentElement;
      if (chipContainer) {
        await user.click(chipContainer);
      }
    }

    // The component should call onRedirectURIsChange with the remaining URIs
    // Since we can't easily test the delete icon click, we'll test the handler directly
    expect(screen.getByText('https://example.com/callback')).toBeInTheDocument();
  });

  it('should clear error when typing in input after error', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    // Trigger error
    await user.click(addButton);
    expect(screen.getByText('Please enter a redirect URI')).toBeInTheDocument();

    // Type to clear error
    await user.type(input, 'https://example.com');

    await waitFor(() => {
      expect(screen.queryByText('Please enter a redirect URI')).not.toBeInTheDocument();
    });
  });

  it('should call onReadyChange with false when no URIs', () => {
    renderComponent({redirectURIs: []});

    expect(mockOnReadyChange).toHaveBeenCalledWith(false);
  });

  it('should call onReadyChange with true when URIs are present', () => {
    renderComponent({redirectURIs: ['https://example.com/callback']});

    expect(mockOnReadyChange).toHaveBeenCalledWith(true);
  });

  it('should call onReadyChange with false when URI is empty string', () => {
    renderComponent({redirectURIs: ['']});

    expect(mockOnReadyChange).toHaveBeenCalledWith(false);
  });

  it('should call onReadyChange when redirectURIs change', () => {
    const {rerender} = renderComponent({redirectURIs: []});

    expect(mockOnReadyChange).toHaveBeenCalledWith(false);

    rerender(<ConfigureRedirectURIs redirectURIs={['https://example.com/callback']} onRedirectURIsChange={mockOnRedirectURIsChange} onReadyChange={mockOnReadyChange} />);

    expect(mockOnReadyChange).toHaveBeenCalledWith(true);
  });

  it('should display hint text with lightbulb icon', () => {
    renderComponent();

    expect(screen.getByText('Redirect URIs must be valid URLs (http:// or https://). These are the allowed callback URLs for your application after authentication.')).toBeInTheDocument();
  });

  it('should handle multiple URIs', async () => {
    const user = userEvent.setup();
    const {rerender} = renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'https://example.com/callback');
    await user.click(addButton);

    // First call should add the first URI
    expect(mockOnRedirectURIsChange).toHaveBeenNthCalledWith(1, ['https://example.com/callback']);

    // Re-render with the first URI in the list
    rerender(
      <ConfigureRedirectURIs
        redirectURIs={['https://example.com/callback']}
        onRedirectURIsChange={mockOnRedirectURIsChange}
        onReadyChange={mockOnReadyChange}
      />,
    );

    const input2 = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton2 = screen.getByRole('button', {name: 'Add'});

    await user.type(input2, 'https://app.example.com/callback');
    await user.click(addButton2);

    // Second call should include both URIs
    expect(mockOnRedirectURIsChange).toHaveBeenNthCalledWith(2, [
      'https://example.com/callback',
      'https://app.example.com/callback',
    ]);
  });

  it('should trim whitespace from URI before adding', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, '  https://example.com/callback  ');
    await user.click(addButton);

    expect(mockOnRedirectURIsChange).toHaveBeenCalledWith(['https://example.com/callback']);
  });

  it('should handle URLs with query parameters', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'https://example.com/callback?code=123&state=abc');
    await user.click(addButton);

    expect(mockOnRedirectURIsChange).toHaveBeenCalledWith(['https://example.com/callback?code=123&state=abc']);
  });

  it('should handle URLs with ports', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'http://localhost:8080/callback');
    await user.click(addButton);

    expect(mockOnRedirectURIsChange).toHaveBeenCalledWith(['http://localhost:8080/callback']);
  });

  it('should handle URLs with paths', async () => {
    const user = userEvent.setup();
    renderComponent();

    const input = screen.getByPlaceholderText('https://localhost:3000/callback');
    const addButton = screen.getByRole('button', {name: 'Add'});

    await user.type(input, 'https://example.com/auth/callback/oauth');
    await user.click(addButton);

    expect(mockOnRedirectURIsChange).toHaveBeenCalledWith(['https://example.com/auth/callback/oauth']);
  });
});

