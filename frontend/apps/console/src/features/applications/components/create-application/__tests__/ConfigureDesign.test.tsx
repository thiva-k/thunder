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

import userEvent from '@testing-library/user-event';
import {render, screen} from '@thunderid/test-utils';
import {describe, it, expect, beforeEach, vi} from 'vitest';
import {ApplicationCreateFlowSignInApproach} from '../../../models/application-create-flow';
import ConfigureDesign, {type ConfigureDesignProps} from '../ConfigureDesign';

// Mock the Packages
vi.mock('@thunderid/design');

vi.mock('@/features/applications/hooks/useApplicationCreateContext', () => ({
  default: () => ({
    ouDefaults: {signIn: false, signUp: false, recovery: false, signOut: false, theme: false, layout: false},
    selectedTemplateConfig: null,
  }),
}));

const {useGetThemes, useGetTheme, useGetLayouts, useGetLayout} = await import('@thunderid/design');

describe('ConfigureDesign', () => {
  const mockOnThemeSelect = vi.fn();
  const mockOnLayoutSelect = vi.fn();

  const defaultProps: ConfigureDesignProps = {
    onThemeSelect: mockOnThemeSelect,
    onLayoutSelect: mockOnLayoutSelect,
    selectedApproach: ApplicationCreateFlowSignInApproach.INBUILT,
    onApproachChange: vi.fn(),
    showApproachSection: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(useGetThemes).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetThemes>);

    vi.mocked(useGetTheme).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetTheme>);

    vi.mocked(useGetLayouts).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetLayouts>);

    vi.mocked(useGetLayout).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetLayout>);
  });

  const renderComponent = (props: Partial<ConfigureDesignProps> = {}) =>
    render(<ConfigureDesign {...defaultProps} {...props} />);

  it('should render the component with title', () => {
    renderComponent();

    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
  });

  it('should render subtitle', () => {
    renderComponent();

    expect(screen.getByText('Customize the appearance of your application')).toBeInTheDocument();
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true on mount', () => {
      const mockOnReadyChange = vi.fn();
      renderComponent({onReadyChange: mockOnReadyChange});

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });
  });

  describe('Sign-In Approach', () => {
    it('renders both approach options and reflects the selected one', () => {
      renderComponent({
        showApproachSection: true,
        selectedApproach: ApplicationCreateFlowSignInApproach.EMBEDDED,
      });

      expect(screen.getByText('Sign-In Approach')).toBeInTheDocument();
      expect(screen.getByText(/Redirect to .* Gate/)).toBeInTheDocument();
      expect(screen.getByText('Bring Your Own UI')).toBeInTheDocument();
      expect(screen.getAllByRole('radio')[1]).toBeChecked();
    });

    it('calls onApproachChange when a different card is clicked', async () => {
      const onApproachChange = vi.fn();
      const user = userEvent.setup();

      renderComponent({
        showApproachSection: true,
        selectedApproach: ApplicationCreateFlowSignInApproach.INBUILT,
        onApproachChange,
      });

      await user.click(screen.getAllByRole('radio')[1]);

      expect(onApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.EMBEDDED);
    });

    it('hides the embedded option when allowEmbeddedApproach is false', () => {
      renderComponent({
        showApproachSection: true,
        allowEmbeddedApproach: false,
        selectedApproach: ApplicationCreateFlowSignInApproach.INBUILT,
      });

      expect(screen.getByText(/Redirect to .* Gate/)).toBeInTheDocument();
      expect(screen.queryByText('Bring Your Own UI')).not.toBeInTheDocument();
      expect(screen.getAllByRole('radio')).toHaveLength(1);
    });
  });

  describe('Theme selection', () => {
    const mockThemeDetails = {
      id: 'theme-1',
      displayName: 'Corporate Blue',
      theme: {
        colorSchemes: {
          light: {
            colors: {
              primary: {
                main: '#123456',
              },
            },
          },
        },
      },
    };

    const mockThemesList = [
      {id: 'theme-1', displayName: 'Corporate Blue'},
      {id: 'theme-2', displayName: 'Sunset Orange'},
    ];

    it('should render theme cards when themes are available', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: mockThemesList},
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetThemes>);

      vi.mocked(useGetTheme).mockReturnValue({
        data: mockThemeDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetTheme>);

      renderComponent();

      expect(screen.getByText('Corporate Blue')).toBeInTheDocument();
      expect(screen.getByText('Sunset Orange')).toBeInTheDocument();
    });

    it('should render a card for each theme', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: mockThemesList},
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetThemes>);

      vi.mocked(useGetTheme).mockReturnValue({
        data: mockThemeDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetTheme>);

      renderComponent();

      expect(screen.getByTestId('theme-card-theme-1')).toBeInTheDocument();
      expect(screen.getByTestId('theme-card-theme-2')).toBeInTheDocument();
    });

    it('should call onThemeSelect with theme details when theme is loaded', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: mockThemesList},
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetThemes>);

      vi.mocked(useGetTheme).mockReturnValue({
        data: mockThemeDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetTheme>);

      renderComponent();

      expect(mockOnThemeSelect).toHaveBeenCalledWith('theme-1', mockThemeDetails.theme);
    });

    it('should show empty state when no themes are configured', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: []},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetThemes>);

      renderComponent();

      expect(screen.getByText('No themes configured')).toBeInTheDocument();
      expect(screen.getByText('You can configure themes later from the Design settings.')).toBeInTheDocument();
    });

    it('should select a different theme when clicking its card', async () => {
      const user = userEvent.setup();
      const mockOnThemeSelectLocal = vi.fn();

      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: mockThemesList},
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetThemes>);

      vi.mocked(useGetTheme).mockReturnValue({
        data: mockThemeDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetTheme>);

      renderComponent({onThemeSelect: mockOnThemeSelectLocal});

      const secondThemeCard = screen.getByTestId('theme-card-theme-2');
      await user.click(secondThemeCard);

      expect(mockOnThemeSelectLocal).toHaveBeenCalledWith('theme-1', mockThemeDetails.theme);
    });
  });

  describe('Layout selection', () => {
    const mockLayoutDetails = {
      id: 'layout-1',
      displayName: 'Split Screen',
      layout: {screens: {}},
    };

    const mockLayoutsList = [
      {id: 'layout-1', displayName: 'Split Screen'},
      {id: 'layout-2', displayName: 'Centered Card'},
    ];

    it('should call onLayoutSelect with layout details when layout is loaded, even though the picker UI is hidden', () => {
      vi.mocked(useGetLayouts).mockReturnValue({
        data: {layouts: mockLayoutsList},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetLayouts>);

      vi.mocked(useGetLayout).mockReturnValue({
        data: mockLayoutDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetLayout>);

      renderComponent();

      expect(mockOnLayoutSelect).toHaveBeenCalledWith('layout-1', mockLayoutDetails.layout);
    });

    it('should never render the layout section, even when layouts are available', () => {
      vi.mocked(useGetLayouts).mockReturnValue({
        data: {layouts: mockLayoutsList},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetLayouts>);

      vi.mocked(useGetLayout).mockReturnValue({
        data: mockLayoutDetails,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetLayout>);

      renderComponent();

      expect(screen.queryByText('Layout')).not.toBeInTheDocument();
      expect(screen.queryByText('Split Screen')).not.toBeInTheDocument();
      expect(screen.queryByTestId('layout-card-layout-1')).not.toBeInTheDocument();
    });
  });
});
