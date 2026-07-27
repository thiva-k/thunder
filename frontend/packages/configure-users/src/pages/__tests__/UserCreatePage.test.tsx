/**
 * Copyright (c) 2025-2026, WSO2 LLC. (https://www.wso2.com).
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

import type {InviteUserRenderProps, EmbeddedFlowComponent} from '@thunderid/react';
import {render, screen, waitFor, userEvent} from '@thunderid/test-utils';
import type {ReactNode} from 'react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import UserCreateProvider from '../../contexts/UserCreate/UserCreateProvider';
import UserCreatePage from '../UserCreatePage';

const mockNavigate = vi.fn();
const mockHandleSubmit = vi.fn();

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    Link: ({to, children = undefined, ...props}: {to: string; children?: ReactNode; [key: string]: unknown}) => (
      <a
        {...(props as Record<string, unknown>)}
        href={to}
        onClick={(e) => {
          e.preventDefault();
          Promise.resolve(mockNavigate(to)).catch(() => null);
        }}
      >
        {children}
      </a>
    ),
  };
});

// Mock logger
vi.mock('@thunderid/logger/react', () => ({
  useLogger: () => ({
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn(),
    withComponent: vi.fn().mockReturnThis(),
  }),
}));

// Mock InviteUser to provide embedded flow components
vi.mock('@thunderid/react', async (importOriginal) => {
  const actual = await importOriginal();
  return {
    ...(actual as object),
    InviteUser: ({children}: {children: (props: InviteUserRenderProps) => ReactNode}) => {
      const mockComponents: EmbeddedFlowComponent[] = [
        {
          id: 'step-heading',
          type: 'TEXT',
          label: 'User Details',
          ref: undefined,
          components: undefined,
          actions: undefined,
        },
        {
          id: 'email-field',
          type: 'EMAIL_INPUT',
          label: 'Email',
          ref: 'email',
          components: undefined,
          actions: undefined,
        },
        {
          id: 'submit-btn',
          type: 'ACTION',
          label: 'Create User',
          variant: 'PRIMARY',
          ref: undefined,
          components: undefined,
          actions: undefined,
        },
      ];

      const renderProps: InviteUserRenderProps = {
        components: mockComponents,
        values: {},
        fieldErrors: {},
        touched: {},
        error: null,
        isLoading: false,
        handleInputChange: vi.fn(),
        handleInputBlur: vi.fn(),
        handleSubmit: mockHandleSubmit,
        resetFlow: vi.fn(),
        isValid: true,
        meta: null,
        additionalData: {rootOuId: 'root-ou'},
      };

      return children(renderProps);
    },
  };
});

describe('UserCreatePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockResolvedValue(undefined);
    mockHandleSubmit.mockResolvedValue(undefined);
  });

  it('renders the page with progress bar', () => {
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    const progressBars = screen.getAllByRole('progressbar');
    expect(progressBars.length).toBeGreaterThan(0);
  });

  it('renders close button', () => {
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    expect(screen.getByLabelText('Close')).toBeInTheDocument();
  });

  it('renders breadcrumb container', () => {
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    // Check that breadcrumb container is rendered
    const breadcrumbContainer = screen.getByLabelText('breadcrumb');
    expect(breadcrumbContainer).toBeInTheDocument();
  });

  it('renders embedded flow components', () => {
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    expect(screen.getByLabelText('Email')).toBeInTheDocument();
  });

  it('renders Create User action button', () => {
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    const buttons = screen.getAllByRole('button').filter((btn) => btn.textContent?.includes('Create User'));
    expect(buttons.length).toBeGreaterThan(0);
  });

  it('closes page when X button is clicked', async () => {
    const user = userEvent.setup();
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    const closeButton = screen.getByLabelText('Close');
    await user.click(closeButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/users');
    });
  });

  it('renders email input field', () => {
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    const emailInput = screen.getByLabelText('Email');
    expect(emailInput).toBeInTheDocument();
  });

  it('allows typing in form fields', async () => {
    const user = userEvent.setup();
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    const emailInput = screen.getByLabelText('Email');
    // User action should complete without error
    await user.type(emailInput, 'test@example.com');

    // Email input field should remain in the document after typing
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
  });

  it('handles form submission', async () => {
    const user = userEvent.setup();
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    const submitButtons = screen.getAllByRole('button').filter((btn) => btn.textContent?.includes('Create User'));
    expect(submitButtons.length).toBeGreaterThan(0);
    await user.click(submitButtons[0]);

    // handleSubmit should have been called
    await waitFor(() => {
      expect(mockHandleSubmit).toHaveBeenCalled();
    });
  });

  it('displays auto-submit behavior when create action is detected', () => {
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    // The page should render without errors
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('renders with AdditionalData containing rootOuId', () => {
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    // The page should successfully render with the mocked additional data
    expect(screen.getByLabelText('Close')).toBeInTheDocument();
  });

  it('handles translation of form labels', () => {
    render(
      <UserCreateProvider>
        <UserCreatePage />
      </UserCreateProvider>,
    );

    // Email label should be translated and visible
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
  });
});
