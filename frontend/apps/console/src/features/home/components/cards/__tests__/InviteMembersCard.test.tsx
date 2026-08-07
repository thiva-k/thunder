// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen, fireEvent} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import InviteMembersCard from '../InviteMembersCard';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string | object) => (typeof fallback === 'string' ? fallback : key),
  }),
}));

const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('framer-motion', async () => {
  const actual = await vi.importActual<typeof import('framer-motion')>('framer-motion');
  return {
    ...actual,
    motion: {
      ...((actual as {motion: object}).motion ?? {}),
      div: ({children, ...rest}: React.HTMLAttributes<HTMLDivElement>) => <div {...rest}>{children}</div>,
    },
  };
});

const mockUseGetUsers = vi.fn();
vi.mock('@thunderid/configure-users', () => ({
  useGetUsers: (args: unknown) => mockUseGetUsers(args) as unknown,
  UserConstants: {DEFAULT_AVATAR_PREFIX: 'avatar:shape=circle,variant=two_letter,colors=0,content='},
}));

/** The fallback avatar renders an <img> whose data-URI SVG embeds the initials as text content. */
function getRenderedInitials(img: HTMLImageElement): string {
  const match = /<text[^>]*>([^<]*)<\/text>/.exec(decodeURIComponent(img.src));
  return match?.[1] ?? '';
}

describe('InviteMembersCard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockReturnValue(undefined);
  });

  describe('Loading state', () => {
    it('renders skeleton placeholders while loading', () => {
      mockUseGetUsers.mockReturnValue({isLoading: true, data: undefined});

      render(<InviteMembersCard />);

      // MUI Skeleton renders with role="img" by default or just as a div
      // We verify skeletons indirectly by ensuring empty/avatar content is absent
      expect(screen.queryByText('No members yet — add collaborators')).not.toBeInTheDocument();
      expect(screen.queryByRole('img')).not.toBeInTheDocument();
    });
  });

  describe('Empty state (admin only)', () => {
    it('renders empty state message when only the admin exists (totalResults = 1)', () => {
      mockUseGetUsers.mockReturnValue({isLoading: false, data: {totalResults: 1, users: []}});

      render(<InviteMembersCard />);

      expect(screen.getByText('No members yet — add collaborators')).toBeInTheDocument();
    });

    it('renders empty state message when totalResults is 0', () => {
      mockUseGetUsers.mockReturnValue({isLoading: false, data: {totalResults: 0, users: []}});

      render(<InviteMembersCard />);

      expect(screen.getByText('No members yet — add collaborators')).toBeInTheDocument();
    });
  });

  describe('Members present', () => {
    const USERS = [
      {id: 'u1', display: 'Alice Smith'},
      {id: 'u2', display: 'Bob Jones'},
    ];

    it('renders member initials when users are present', () => {
      mockUseGetUsers.mockReturnValue({isLoading: false, data: {totalResults: 3, users: USERS}});

      render(<InviteMembersCard />);

      const initials = screen.getAllByRole<HTMLImageElement>('img').map(getRenderedInitials);
      expect(initials).toContain('AS');
      expect(initials).toContain('BJ');
    });

    it('renders first two characters for a single-word display name', () => {
      mockUseGetUsers.mockReturnValue({
        isLoading: false,
        data: {totalResults: 2, users: [{id: 'u1', display: 'Alice'}]},
      });

      render(<InviteMembersCard />);

      expect(getRenderedInitials(screen.getByRole<HTMLImageElement>('img'))).toBe('AL');
    });

    it('renders an extra count when totalResults exceeds the avatar limit', () => {
      const manyUsers = Array.from({length: 5}, (_, i) => ({id: `u${i}`, display: `User ${i}`}));
      mockUseGetUsers.mockReturnValue({isLoading: false, data: {totalResults: 8, users: manyUsers}});

      render(<InviteMembersCard />);

      expect(screen.getByText('+3')).toBeInTheDocument();
    });

    it('does not render extra count when totalResults is at or below the avatar limit', () => {
      mockUseGetUsers.mockReturnValue({isLoading: false, data: {totalResults: 4, users: USERS}});

      render(<InviteMembersCard />);

      expect(screen.queryByText(/^\+\d/)).not.toBeInTheDocument();
    });
  });

  describe('Action buttons', () => {
    beforeEach(() => {
      mockUseGetUsers.mockReturnValue({isLoading: false, data: {totalResults: 3, users: []}});
    });

    it('renders the "Add User" button', () => {
      render(<InviteMembersCard />);

      expect(screen.getByRole('button', {name: 'Add User'})).toBeInTheDocument();
    });

    it('navigates to /users/add when Add User is clicked', () => {
      render(<InviteMembersCard />);

      fireEvent.click(screen.getByRole('button', {name: 'Add User'}));

      expect(mockNavigate).toHaveBeenCalledWith('/users/add');
    });
  });
});
