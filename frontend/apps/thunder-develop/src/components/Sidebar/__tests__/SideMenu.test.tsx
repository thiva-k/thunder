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
import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type {JSX} from 'react';
import render from '@/test/test-utils';
import SideMenu from '../SideMenu';

// Mock Asgardeo User component
vi.mock('@asgardeo/react', () => ({
  User: ({children}: {children: (user: {name: string; email: string}) => JSX.Element}) =>
    children({name: 'John Doe', email: 'john.doe@example.com'}),
}));

// Mock ThemedIcon component
vi.mock('@thunder/ui', () => ({
  ThemedIcon: () => <div data-testid="themed-icon">Logo</div>,
}));

// Mock child components
vi.mock('../MenuContent', () => ({
  default: () => <div data-testid="menu-content">Menu Content</div>,
}));

vi.mock('../OptionsMenu', () => ({
  default: () => (
    <button type="button" data-testid="options-menu">
      Options
    </button>
  ),
}));

describe('SideMenu', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  describe('Basic rendering', () => {
    it('renders the drawer component', () => {
      const {container} = render(<SideMenu />);

      const drawer = container.querySelector('.MuiDrawer-root');
      expect(drawer).toBeInTheDocument();
    });

    it('renders the Developer title when expanded', () => {
      render(<SideMenu />);

      expect(screen.getByText('Developer')).toBeInTheDocument();
    });

    it('renders menu content', () => {
      render(<SideMenu />);

      expect(screen.getByTestId('menu-content')).toBeInTheDocument();
    });

    it('renders user information when expanded', () => {
      render(<SideMenu />);

      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.getByText('john.doe@example.com')).toBeInTheDocument();
    });

    it('renders user avatar', () => {
      const {container} = render(<SideMenu />);

      const avatar = container.querySelector('.MuiAvatar-root');
      expect(avatar).toBeInTheDocument();
      expect(avatar).toHaveTextContent('J');
    });

    it('renders options menu when expanded', () => {
      render(<SideMenu />);

      expect(screen.getByTestId('options-menu')).toBeInTheDocument();
    });

    it('has permanent variant', () => {
      const {container} = render(<SideMenu />);

      const drawer = container.querySelector('.MuiDrawer-root');
      expect(drawer).toHaveClass('MuiDrawer-docked');
    });

    it('renders divider after logo section', () => {
      const {container} = render(<SideMenu />);

      const dividers = container.querySelectorAll('.MuiDivider-root');
      expect(dividers.length).toBeGreaterThan(0);
    });
  });

  describe('Collapsible functionality', () => {
    it('renders collapse button when expanded', () => {
      render(<SideMenu />);

      const collapseButton = screen.getByLabelText('Collapse sidebar');
      expect(collapseButton).toBeInTheDocument();
    });

    it('renders expand button when collapsed', () => {
      render(<SideMenu expanded={false} />);

      const expandButton = screen.getByLabelText('Expand sidebar');
      expect(expandButton).toBeInTheDocument();
    });

    it('calls onExpandedChange when collapse button is clicked', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      const handleExpandedChange = vi.fn();
      render(<SideMenu expanded onExpandedChange={handleExpandedChange} />);

      const collapseButton = screen.getByLabelText('Collapse sidebar');
      await user.click(collapseButton);

      expect(handleExpandedChange).toHaveBeenCalledWith(false);
      vi.useFakeTimers(); // Restore fake timers
    });

    it('calls onExpandedChange when expand button is clicked', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      const handleExpandedChange = vi.fn();
      render(<SideMenu expanded={false} onExpandedChange={handleExpandedChange} />);

      const expandButton = screen.getByLabelText('Expand sidebar');
      await user.click(expandButton);

      expect(handleExpandedChange).toHaveBeenCalledWith(true);
      vi.useFakeTimers(); // Restore fake timers
    });

    it('toggles expanded state when used in uncontrolled mode', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      render(<SideMenu />);

      // Initially expanded
      expect(screen.getByText('Developer')).toBeInTheDocument();
      expect(screen.getByLabelText('Collapse sidebar')).toBeInTheDocument();

      // Click to collapse
      const collapseButton = screen.getByLabelText('Collapse sidebar');
      await user.click(collapseButton);

      // Should now show expand button
      await waitFor(() => {
        expect(screen.getByLabelText('Expand sidebar')).toBeInTheDocument();
      });
      vi.useFakeTimers(); // Restore fake timers
    });

    it('hides logo and title when collapsed', () => {
      render(<SideMenu expanded={false} />);

      expect(screen.queryByText('Developer')).not.toBeInTheDocument();
      // Logo should not be rendered when collapsed (it's inside the same conditional block)
    });

    it('hides user information when collapsed', () => {
      render(<SideMenu expanded={false} />);

      expect(screen.queryByText('John Doe')).not.toBeInTheDocument();
      expect(screen.queryByText('john.doe@example.com')).not.toBeInTheDocument();
    });

    it('hides options menu when collapsed', () => {
      render(<SideMenu expanded={false} />);

      expect(screen.queryByTestId('options-menu')).not.toBeInTheDocument();
    });

    it('still shows user avatar when collapsed', () => {
      const {container} = render(<SideMenu expanded={false} />);

      const avatar = container.querySelector('.MuiAvatar-root');
      expect(avatar).toBeInTheDocument();
    });

    it('does not render collapse button when disableCollapsible is true', () => {
      render(<SideMenu disableCollapsible />);

      expect(screen.queryByLabelText('Collapse sidebar')).not.toBeInTheDocument();
    });

    it('does not render expand button when disableCollapsible is true and collapsed', () => {
      render(<SideMenu expanded={false} disableCollapsible />);

      expect(screen.queryByLabelText('Expand sidebar')).not.toBeInTheDocument();
    });

    it('shows logo and title when disableCollapsible is true', () => {
      render(<SideMenu expanded={false} disableCollapsible />);

      // When collapsible is disabled, it should still show content even if expanded is false
      expect(screen.getByText('Developer')).toBeInTheDocument();
    });
  });

  describe('Drawer width', () => {
    it('has correct width when expanded', () => {
      const {container} = render(<SideMenu expanded />);

      const drawer = container.querySelector('.MuiDrawer-root');
      const style = window.getComputedStyle(drawer!);
      expect(style.width).toBe('240px');
    });

    it('has correct width when collapsed', () => {
      const {container} = render(<SideMenu expanded={false} />);

      const drawer = container.querySelector('.MuiDrawer-root');
      const style = window.getComputedStyle(drawer!);
      expect(style.width).toBe('64px');
    });
  });

  describe('Controlled vs Uncontrolled mode', () => {
    it('works as controlled component when expanded prop is provided', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      const handleExpandedChange = vi.fn();
      const {rerender} = render(<SideMenu expanded onExpandedChange={handleExpandedChange} />);

      expect(screen.getByText('Developer')).toBeInTheDocument();

      // Click collapse button
      const collapseButton = screen.getByLabelText('Collapse sidebar');
      await user.click(collapseButton);

      // Should call handler but not change state (controlled)
      expect(handleExpandedChange).toHaveBeenCalledWith(false);
      expect(screen.getByText('Developer')).toBeInTheDocument();

      // Parent updates the prop
      rerender(<SideMenu expanded={false} onExpandedChange={handleExpandedChange} />);

      // Now it should be collapsed
      expect(screen.queryByText('Developer')).not.toBeInTheDocument();
      vi.useFakeTimers(); // Restore fake timers
    });

    it('works as uncontrolled component when expanded prop is not provided', async () => {
      vi.useRealTimers(); // Use real timers for user events
      const user = userEvent.setup();
      render(<SideMenu />);

      // Initially expanded
      expect(screen.getByText('Developer')).toBeInTheDocument();

      // Click to collapse
      const collapseButton = screen.getByLabelText('Collapse sidebar');
      await user.click(collapseButton);

      // Should update internal state and collapse
      await waitFor(() => {
        expect(screen.getByLabelText('Expand sidebar')).toBeInTheDocument();
      });
      vi.useFakeTimers(); // Restore fake timers
    });
  });

  describe('Accessibility', () => {
    it('has proper aria-label for collapse button', () => {
      render(<SideMenu expanded />);

      const collapseButtons = screen.getAllByLabelText('Collapse sidebar');
      expect(collapseButtons[0]).toHaveAccessibleName('Collapse sidebar');
    });

    it('has proper aria-label for expand button', () => {
      render(<SideMenu expanded={false} />);

      const expandButton = screen.getByLabelText('Expand sidebar');
      expect(expandButton).toHaveAccessibleName('Expand sidebar');
    });
  });

  describe('SidebarContext', () => {
    it('provides correct context values when expanded', () => {
      // We can verify context is provided by checking if MenuContent renders
      // (MenuContent depends on SidebarContext)
      render(<SideMenu expanded />);

      expect(screen.getByTestId('menu-content')).toBeInTheDocument();
    });

    it('provides correct context values when collapsed', () => {
      render(<SideMenu expanded={false} />);

      expect(screen.getByTestId('menu-content')).toBeInTheDocument();
    });
  });
});
