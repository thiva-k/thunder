// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import userEvent from '@testing-library/user-event';
import {render, screen} from '@thunderid/test-utils';
import {describe, it, expect, vi} from 'vitest';
import ScreenListItem from '../ScreenListItem';

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next');
  return {
    ...actual,
    useTranslation: () => ({t: (key: string) => key}),
  };
});

describe('ScreenListItem', () => {
  describe('Rendering', () => {
    it('renders the screen name', () => {
      render(<ScreenListItem name="auth" isSelected={false} onClick={vi.fn()} />);
      expect(screen.getByText('auth')).toBeInTheDocument();
    });

    it('renders "base screen" label when no extendsBase is provided', () => {
      render(<ScreenListItem name="auth" isSelected={false} onClick={vi.fn()} />);
      // Translated key for "base screen"
      expect(
        screen.getByText((text) => text.toLowerCase().includes('base') || text.includes('design.layout')),
      ).toBeInTheDocument();
    });

    it('renders "extends X" text when extendsBase is provided', () => {
      render(<ScreenListItem name="password" extendsBase="auth" isSelected={false} onClick={vi.fn()} />);
      expect(screen.getByText(/auth/)).toBeInTheDocument();
    });

    it('renders a visual screen icon', () => {
      const {container} = render(<ScreenListItem name="login" isSelected={false} onClick={vi.fn()} />);
      // The screen icon is a CSS-styled box representation, not an SVG
      expect(container.querySelector('[class*="MuiCardContent"]')).toBeInTheDocument();
    });
  });

  describe('Selection state', () => {
    it('renders without errors when not selected', () => {
      render(<ScreenListItem name="auth" isSelected={false} onClick={vi.fn()} />);
      expect(screen.getByText('auth')).toBeInTheDocument();
    });

    it('renders without errors when selected', () => {
      render(<ScreenListItem name="auth" isSelected onClick={vi.fn()} />);
      expect(screen.getByText('auth')).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('calls onClick when the item is clicked', async () => {
      const onClick = vi.fn();
      const user = userEvent.setup();
      render(<ScreenListItem name="login" isSelected={false} onClick={onClick} />);

      await user.click(screen.getByText('login'));

      expect(onClick).toHaveBeenCalledOnce();
    });

    it('calls onClick on repeated clicks', async () => {
      const onClick = vi.fn();
      const user = userEvent.setup();
      render(<ScreenListItem name="signup" isSelected={false} onClick={onClick} />);

      await user.click(screen.getByText('signup'));
      await user.click(screen.getByText('signup'));

      expect(onClick).toHaveBeenCalledTimes(2);
    });
  });
});
