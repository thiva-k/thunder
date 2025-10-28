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
import {screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import render from '@/test/test-utils';
import MenuContent from './MenuContent';

const mockSetCurrentPage = vi.fn();

// Mock useNavigation hook
vi.mock('@/layouts/contexts/useNavigation', () => ({
  default: vi.fn(() => ({
    currentPage: 'home',
    setCurrentPage: mockSetCurrentPage,
    sidebarOpen: false,
    setSidebarOpen: vi.fn(),
    toggleSidebar: vi.fn(),
  })),
}));

describe('MenuContent', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders main menu items', () => {
    render(<MenuContent />);

    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Users')).toBeInTheDocument();
    expect(screen.getByText('User Types')).toBeInTheDocument();
    expect(screen.getByText('Integrations')).toBeInTheDocument();
    expect(screen.getByText('Applications')).toBeInTheDocument();
  });

  it('renders secondary menu items', () => {
    render(<MenuContent />);

    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('renders icons for main menu items', () => {
    const {container} = render(<MenuContent />);

    expect(container.querySelector('svg.lucide-house')).toBeInTheDocument();
    expect(container.querySelector('svg.lucide-users-round')).toBeInTheDocument();
    expect(container.querySelector('svg.lucide-user')).toBeInTheDocument();
    expect(container.querySelector('svg.lucide-blocks')).toBeInTheDocument();
    expect(container.querySelector('svg.lucide-layout-grid')).toBeInTheDocument();
  });

  it('renders icons for secondary menu items', () => {
    const {container} = render(<MenuContent />);

    expect(container.querySelector('svg.lucide-settings')).toBeInTheDocument();
  });

  it('calls setCurrentPage when a menu item is clicked', async () => {
    const user = userEvent.setup();
    render(<MenuContent />);

    const usersLink = screen.getByText('Users');
    await user.click(usersLink);

    expect(mockSetCurrentPage).toHaveBeenCalledWith('users');
  });

  it('calls setCurrentPage with correct data for secondary items', async () => {
    const user = userEvent.setup();
    render(<MenuContent />);

    const settingsLink = screen.getByText('Settings');
    await user.click(settingsLink);

    expect(mockSetCurrentPage).toHaveBeenCalledWith('settings');
  });

  it('renders NavLink components with correct paths', () => {
    const {container} = render(<MenuContent />);

    const homeLink = container.querySelector('a[href="/"]');
    expect(homeLink).toBeInTheDocument();

    const usersLink = container.querySelector('a[href="/users"]');
    expect(usersLink).toBeInTheDocument();

    const settingsLink = container.querySelector('a[href="/settings"]');
    expect(settingsLink).toBeInTheDocument();
  });

  it('marks current page as selected', () => {
    render(<MenuContent />);

    // Home should be selected based on mock
    const homeButton = screen.getByText('Home').closest('a');
    expect(homeButton).toHaveClass('Mui-selected');
  });

  it('renders two separate lists', () => {
    const {container} = render(<MenuContent />);

    const lists = container.querySelectorAll('.MuiList-root');
    expect(lists.length).toBe(2);
  });

  it('all menu items are clickable links', () => {
    render(<MenuContent />);

    const homeLink = screen.getByText('Home').closest('a');
    expect(homeLink).toBeInTheDocument();

    const usersLink = screen.getByText('Users').closest('a');
    expect(usersLink).toBeInTheDocument();
  });
});
