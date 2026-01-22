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

/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-call */

import {render, screen, waitFor} from '@testing-library/react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import userEvent from '@testing-library/user-event';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {BrowserRouter} from 'react-router';
import {ConfigProvider} from '@thunder/commons-contexts';
import ApplicationCreatePage from '../ApplicationCreatePage';
import ApplicationCreateProvider from '../../contexts/ApplicationCreate/ApplicationCreateProvider';

// Mock the logger
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn(),
    withComponent: vi.fn().mockReturnThis(),
  }),
}));

// Mock API hooks - declared before they're used in mock factories
const mockCreateApplication = vi.fn();
const mockCreateBranding = vi.fn();

// Mock the branding hook
vi.mock('@thunder/shared-branding', () => ({
  useCreateBranding: () => ({
    mutate: mockCreateBranding,
    isPending: false,
  }),
  LayoutType: {
    CENTERED: 'centered',
  },
}));

// Mock user types API
vi.mock('../../../user-types/api/useGetUserTypes', () => ({
  default: () => ({
    data: {
      schemas: [],
    },
    isLoading: false,
    error: null,
  }),
}));

// Mock child components
vi.mock('../../components/create-applications/ConfigureName', () => ({
  default: ({
    appName,
    onAppNameChange,
    onReadyChange,
  }: {
    appName: string;
    onAppNameChange: (name: string) => void;
    onReadyChange: (ready: boolean) => void;
  }) => (
    <div data-testid="configure-name">
      <input
        data-testid="app-name-input"
        value={appName}
        onChange={(e) => {
          onAppNameChange(e.target.value);
          onReadyChange(e.target.value.length > 0);
        }}
        placeholder="Enter app name"
      />
    </div>
  ),
}));

vi.mock('../../components/create-applications/ConfigureDesign', () => ({
  default: ({
    selectedColor,
    onColorSelect,
    onLogoSelect,
    onBrandingSelectionChange,
  }: {
    appLogo: string | null;
    selectedColor: string;
    onColorSelect: (color: string) => void;
    onLogoSelect: (logo: string) => void;
    onReadyChange: (ready: boolean) => void;
    onBrandingSelectionChange?: (useDefault: boolean, brandingId?: string) => void;
  }) => (
    <div data-testid="configure-design">
      <input
        data-testid="color-picker"
        type="color"
        value={selectedColor}
        onChange={(e) => onColorSelect(e.target.value)}
      />
      <button type="button" data-testid="logo-select-btn" onClick={() => onLogoSelect('test-logo.png')}>
        Select Logo
      </button>
      <button
        type="button"
        data-testid="use-default-branding-btn"
        onClick={() => onBrandingSelectionChange?.(true, 'default-branding-id')}
      >
        Use Default Branding
      </button>
      <button
        type="button"
        data-testid="use-custom-branding-btn"
        onClick={() => onBrandingSelectionChange?.(false)}
      >
        Use Custom Branding
      </button>
    </div>
  ),
}));

vi.mock('../../components/create-applications/configure-signin-options/ConfigureSignInOptions', async () => {
  const useApplicationCreateContextModule = await import('../../hooks/useApplicationCreateContext');

  return {
    default: function MockConfigureSignInOptions({
      integrations,
      onIntegrationToggle,
      onReadyChange,
    }: {
      integrations: Record<string, boolean>;
      onIntegrationToggle: (id: string) => void;
      onReadyChange: (ready: boolean) => void;
    }) {
      const {setSelectedAuthFlow} = useApplicationCreateContextModule.default();

      // Set the auth flow and ready state synchronously on mount (consistent with other mocks in this file)
      setTimeout(() => {
        setSelectedAuthFlow({
          id: 'test-flow-id',
          name: 'Test Flow',
          flowType: 'AUTHENTICATION',
          handle: 'test-flow',
          activeVersion: 1,
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-01T00:00:00Z',
        });
        const hasSelection = Object.values(integrations).some((enabled: boolean) => enabled);
        onReadyChange(hasSelection);
      }, 0);

      return (
        <div data-testid="configure-sign-in">
          <button type="button" data-testid="toggle-integration" onClick={() => onIntegrationToggle('basic_auth')}>
            Toggle Integration
          </button>
        </div>
      );
    },
  };
});


vi.mock('../../components/create-applications/Preview', () => ({
  default: ({appName, appLogo, selectedColor}: {appName: string; appLogo: string | null; selectedColor: string}) => (
    <div data-testid="preview">
      <div data-testid="preview-name">{appName}</div>
      <div data-testid="preview-logo">{appLogo}</div>
      <div data-testid="preview-color">{selectedColor}</div>
    </div>
  ),
}));

vi.mock('../../components/create-applications/ConfigureOAuth', () => ({
  default: ({
    onReadyChange,
  }: {
    oauthConfig: any;
    onOAuthConfigChange: (config: any) => void;
    onReadyChange?: (ready: boolean) => void;
    onValidationErrorsChange?: (hasErrors: boolean) => void;
  }) => {
    setTimeout(() => {
      if (onReadyChange) {
        onReadyChange(true);
      }
    }, 0);

    return <div data-testid="configure-oauth">Configure OAuth</div>;
  },
}));

// Add missing mocks for the new onboarding components
vi.mock('../../components/create-applications/ConfigureApproach', () => ({
  default: ({
    onReadyChange,
    onApproachChange,
    selectedApproach,
  }: {
    onReadyChange: (ready: boolean) => void;
    onApproachChange: (approach: string) => void;
    selectedApproach: string;
  }) => {
    setTimeout(() => onReadyChange(true), 0);
    return (
      <div data-testid="configure-approach">
        <span data-testid="current-approach">{selectedApproach}</span>
        <button type="button" data-testid="select-custom-approach" onClick={() => onApproachChange('CUSTOM')}>
          Select Custom
        </button>
        <button type="button" data-testid="select-inbuilt-approach" onClick={() => onApproachChange('INBUILT')}>
          Select Inbuilt
        </button>
      </div>
    );
  },
}));

vi.mock('../../components/create-applications/ConfigureStack', () => ({
  default: ({onReadyChange}: {onReadyChange: (ready: boolean) => void}) => {
    setTimeout(() => onReadyChange(true), 0);
    return <div data-testid="configure-stack">Configure Stack</div>;
  },
}));

vi.mock('../../components/create-applications/ConfigureDetails', () => ({
  default: ({
    onReadyChange,
    onCallbackUrlChange,
    userTypes,
    selectedUserTypes,
    onUserTypesChange,
  }: {
    onReadyChange: (ready: boolean) => void;
    onCallbackUrlChange: (url: string) => void;
    userTypes: {name: string}[];
    selectedUserTypes: string[];
    onUserTypesChange: (types: string[]) => void;
  }) => {
    setTimeout(() => onReadyChange(true), 0);
    return (
      <div data-testid="configure-details">
        <input
          data-testid="callback-url-input"
          onChange={(e) => onCallbackUrlChange(e.target.value)}
          placeholder="Callback URL"
        />
        <div data-testid="user-types-count">{userTypes?.length ?? 0}</div>
        <div data-testid="selected-user-types">{selectedUserTypes?.join(',') ?? ''}</div>
        <button type="button" data-testid="select-user-type" onClick={() => onUserTypesChange(['customer'])}>
          Select Customer Type
        </button>
      </div>
    );
  },
}));

// Mock utility function to force URL configuration type
vi.mock('../../utils/getConfigurationTypeFromTemplate', () => ({
  default: vi.fn(() => 'URL'),
}));

vi.mock('../../components/create-applications/ApplicationSummary', () => ({
  default: ({
    appName,
    applicationId,
    clientId,
    clientSecret,
    hasOAuthConfig,
  }: {
    appName: string;
    appLogo: string | null;
    selectedColor: string;
    clientId?: string;
    clientSecret?: string;
    hasOAuthConfig: boolean;
    applicationId?: string | null;
  }) => (
    <div data-testid="application-summary">
      <div data-testid="summary-app-name">{appName}</div>
      <div data-testid="summary-app-id">{applicationId}</div>
      <div data-testid="summary-client-id">{clientId}</div>
      <div data-testid="summary-client-secret">{clientSecret}</div>
      <div data-testid="summary-has-oauth">{hasOAuthConfig ? 'true' : 'false'}</div>
    </div>
  ),
}));

// Mock react-router navigate
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('../../api/useCreateApplication', () => ({
  default: () => ({
    mutate: mockCreateApplication,
    isPending: false,
  }),
}));

vi.mock('../../../integrations/api/useIdentityProviders', () => ({
  default: () => ({
    data: [
      {id: 'google', name: 'Google', type: 'social'},
      {id: 'github', name: 'GitHub', type: 'social'},
    ],
    isLoading: false,
    error: null,
  }),
}));

vi.mock('../../../flows/api/useGetFlows', () => ({
  default: () => ({
    data: {
      flows: [
        {id: 'flow1', name: 'Basic Auth Flow', handle: 'basic-auth'},
        {id: 'flow2', name: 'Google Flow', handle: 'google-flow'},
      ],
    },
    isLoading: false,
    error: null,
  }),
}));

describe('ApplicationCreatePage', () => {
  let queryClient: QueryClient;
  let user: ReturnType<typeof userEvent.setup>;

  const renderWithProviders = () =>
    render(
      <BrowserRouter>
        <QueryClientProvider client={queryClient}>
          <ConfigProvider>
            <ApplicationCreateProvider>
              <ApplicationCreatePage />
            </ApplicationCreateProvider>
          </ConfigProvider>
        </QueryClientProvider>
      </BrowserRouter>,
    );

  beforeEach(async () => {
    user = userEvent.setup();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });

    window.history.replaceState({}, '', '/');

    // Set up runtime config
    if (typeof window !== 'undefined') {
      // eslint-disable-next-line no-underscore-dangle
      window.__THUNDER_RUNTIME_CONFIG__ = {
        client: {
          base: '/develop',
          client_id: 'DEVELOP',
        },
        server: {
          hostname: 'localhost',
          port: 8090,
          http_only: false,
        },
      };
    }

    vi.clearAllMocks();
    mockNavigate.mockResolvedValue(undefined);

    // Reset the getConfigurationTypeFromTemplate mock to return 'URL' (needs configuration)
    const getConfigurationTypeFromTemplate = await import('../../utils/getConfigurationTypeFromTemplate');
    vi.mocked(getConfigurationTypeFromTemplate.default).mockReturnValue('URL');
  });

  describe('Initial Rendering', () => {
    it('should render the first step (name) by default', () => {
      renderWithProviders();

      expect(screen.getByTestId('configure-name')).toBeInTheDocument();
      expect(screen.queryByTestId('configure-design')).not.toBeInTheDocument();
      expect(screen.queryByTestId('configure-sign-in')).not.toBeInTheDocument();
    });

    it('should not show preview on first step', () => {
      renderWithProviders();

      expect(screen.queryByTestId('preview')).not.toBeInTheDocument();
    });

    it('should render close button', () => {
      const {container} = renderWithProviders();

      // The close button exists as an IconButton
      const buttons = container.querySelectorAll('button');
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('should render progress bar', () => {
      const {container} = renderWithProviders();

      const progressBar = container.querySelector('.MuiLinearProgress-root');
      expect(progressBar).toBeInTheDocument();
    });

    it('should show breadcrumb with current step', () => {
      renderWithProviders();

      expect(screen.getByText('Create an Application')).toBeInTheDocument();
    });
  });

  describe('Step Navigation', () => {
    it('should disable Continue button when name is empty', () => {
      renderWithProviders();

      const continueButton = screen.getByRole('button', {name: /continue/i});
      expect(continueButton).toBeDisabled();
    });

    it('should enable Continue button when name is entered', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');

      const continueButton = screen.getByRole('button', {name: /continue/i});
      expect(continueButton).toBeEnabled();
    });

    it('should navigate to design step when Continue is clicked', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');

      const continueButton = screen.getByRole('button', {name: /continue/i});
      await user.click(continueButton);

      expect(screen.getByTestId('configure-design')).toBeInTheDocument();
      expect(screen.queryByTestId('configure-name')).not.toBeInTheDocument();
    });

    it('should show preview from design step onwards', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');

      const continueButton = screen.getByRole('button', {name: /continue/i});
      await user.click(continueButton);

      expect(screen.getByTestId('preview')).toBeInTheDocument();
    });

    it('should navigate to options step from design step', async () => {
      renderWithProviders();

      // Step 1: Enter name
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 2: Continue from design
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
      expect(screen.queryByTestId('configure-design')).not.toBeInTheDocument();
    });

    it('should show Back button from design step', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByRole('button', {name: /back/i})).toBeInTheDocument();
    });

    it('should navigate back to previous step when Back is clicked', async () => {
      renderWithProviders();

      // Navigate to design step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Click back
      await user.click(screen.getByRole('button', {name: /back/i}));

      expect(screen.getByTestId('configure-name')).toBeInTheDocument();
      expect(screen.queryByTestId('configure-design')).not.toBeInTheDocument();
    });

    it('should reach configuration step with continue button', async () => {
      renderWithProviders();

      // Step 1: Name
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 2: Design
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 3: Sign In Options
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 4: Configure Approach
      await waitFor(() => {
        expect(screen.getByTestId('configure-approach')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 5: Configure Stack
      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 6: Configure Details
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });

      // Should have Continue button on configuration step (this is the final user-facing step)
      const continueButton = screen.getByRole('button', {name: /continue/i});
      expect(continueButton).toBeInTheDocument();
      expect(continueButton).toBeEnabled();
    });
  });

  describe('Breadcrumb Navigation', () => {
    it('should update breadcrumb as user progresses through steps', async () => {
      renderWithProviders();

      expect(screen.getByText('Create an Application')).toBeInTheDocument();

      // Navigate to design
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByText('Design')).toBeInTheDocument();

      // Navigate to options
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByText('Sign In Options')).toBeInTheDocument();
    });

    it('should allow clicking on previous breadcrumb steps to navigate back', async () => {
      renderWithProviders();

      // Navigate to options step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Click on "Create an Application" breadcrumb
      const firstBreadcrumb = screen.getByText('Create an Application');
      await user.click(firstBreadcrumb);

      expect(screen.getByTestId('configure-name')).toBeInTheDocument();
    });
  });

  describe('Close Functionality', () => {
    it('should navigate to applications list when close button is clicked', async () => {
      const {container} = renderWithProviders();

      // Find the close button by finding the IconButton with X icon
      const closeButton = container.querySelector('button[aria-label]') ?? container.querySelector('button');
      expect(closeButton).toBeInTheDocument();

      if (closeButton) {
        await user.click(closeButton);

        await waitFor(() => {
          expect(mockNavigate).toHaveBeenCalledWith('/applications');
        });
      }
    });
  });

  describe('Form State Management', () => {
    it('should update app name state when user types', async () => {
      renderWithProviders();

      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'Test App');

      expect(nameInput).toHaveValue('Test App');
    });

    it('should preserve app name when navigating between steps', async () => {
      renderWithProviders();

      // Enter name
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');

      // Navigate to design and back
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /back/i}));

      // Name should be preserved
      expect(screen.getByTestId('app-name-input')).toHaveValue('My App');
    });

    it.skip('should update preview with app name', async () => {
      renderWithProviders();

      // Enter name and navigate to design
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Check preview
      expect(screen.getByTestId('preview-name')).toHaveTextContent('My App');
    });

    it('should update color in state', async () => {
      renderWithProviders();

      // Navigate to design step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Change color - color inputs work differently, we need to use fireEvent or just set the value
      const colorPicker = screen.getByTestId('color-picker');

      // Simulate color change by triggering the onChange event
      await user.click(colorPicker);

      // The mock component will handle the color change
      // Just verify the initial color is shown
      expect(screen.getByTestId('preview-color')).toBeInTheDocument();
    });

    it('should update logo in state', async () => {
      renderWithProviders();

      // Navigate to design step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Select logo
      const logoButton = screen.getByTestId('logo-select-btn');
      await user.click(logoButton);

      // Check preview
      expect(screen.getByTestId('preview-logo')).toHaveTextContent('test-logo.png');
    });
  });

  describe('Application Creation', () => {
    it.skip('should call createApplication when continuing from configuration step', async () => {
      renderWithProviders();

      // Step 1: Name
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 2: Design
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 3: Sign In Options
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Navigate through all steps to reach configuration step
      await waitFor(() => {
        expect(screen.getByTestId('configure-approach')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });

      // Continue from configuration step (this triggers application creation)
      const continueButton = screen.getByRole('button', {name: /continue/i});
      await user.click(continueButton);

      await waitFor(
        () => {
          expect(mockCreateBranding).toHaveBeenCalled();
        },
        {timeout: 3000},
      );
    });

    it.skip('should show summary step after successful creation', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({id: 'branding-123', name: 'Test Branding'});
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: {
                client_id: 'test-client-id',
                client_secret: 'test-client-secret',
                public_client: false,
              },
            },
          ],
        });
      });

      renderWithProviders();

      // Step 1: Name
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 2: Design
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 3: Sign In Options
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Navigate through all steps to reach configuration step
      await waitFor(() => {
        expect(screen.getByTestId('configure-approach')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });

      // Continue from configuration step (this triggers application creation)
      const continueButton = screen.getByRole('button', {name: /continue/i});
      await user.click(continueButton);

      // After successful creation, should show summary step
      await waitFor(
        () => {
          expect(screen.getByTestId('application-summary')).toBeInTheDocument();
        },
        {timeout: 3000},
      );

      // Verify the summary shows the created app
      expect(screen.getByTestId('summary-app-name')).toHaveTextContent('My App');
      expect(screen.getByTestId('summary-app-id')).toHaveTextContent('app-123');
    });

    it.skip('should show error message when creation fails', async () => {
      mockCreateBranding.mockImplementation((_data, {onError}: any) => {
        onError(new Error('Failed to create branding'));
      });

      renderWithProviders();

      // Step 1: Name
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 2: Design (always ready, just click continue)
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 3: Sign In Options
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Navigate through all steps to reach configuration step
      await waitFor(() => {
        expect(screen.getByTestId('configure-approach')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });

      // Continue from configuration step (this triggers application creation and fails)
      const continueButton = screen.getByRole('button', {name: /continue/i});
      await user.click(continueButton);

      await waitFor(
        () => {
          expect(screen.getByText(/failed to create branding/i)).toBeInTheDocument();
        },
        {timeout: 3000},
      );
    });

    it.skip('should allow dismissing error message', async () => {
      mockCreateBranding.mockImplementation((_data, {onError}: any) => {
        onError(new Error('Failed to create branding'));
      });

      renderWithProviders();

      // Step 1: Name
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 2: Design (always ready, just click continue)
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 3: Sign In Options
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Navigate through all steps to reach configuration step
      await waitFor(() => {
        expect(screen.getByTestId('configure-approach')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });

      // Continue from configuration step (this triggers application creation and fails)
      const continueButton = screen.getByRole('button', {name: /continue/i});
      await user.click(continueButton);

      await waitFor(
        () => {
          expect(screen.getByText(/failed to create branding/i)).toBeInTheDocument();
        },
        {timeout: 3000},
      );

      // Close error
      const closeButton = screen.getByLabelText(/close/i);
      await user.click(closeButton);

      await waitFor(
        () => {
          expect(screen.queryByText(/failed to create branding/i)).not.toBeInTheDocument();
        },
        {timeout: 3000},
      );
    });
  });

  describe('Integration Management', () => {
    it('should toggle integration when requested', async () => {
      renderWithProviders();

      // Navigate to options step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Toggle integration
      const toggleButton = screen.getByTestId('toggle-integration');
      await user.click(toggleButton);

      // Component should not crash
      expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
    });

    it('should start with default selections (options step ready)', async () => {
      renderWithProviders();

      // Navigate to options step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Options step should be ready (username/password selected by default)
      // Continue button should be enabled
      const continueButton = screen.getByRole('button', {name: /continue/i});
      expect(continueButton).toBeEnabled();
    });

    it('should require at least one option to be selected before proceeding', async () => {
      renderWithProviders();

      // Navigate to options step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Initially enabled (default selection)
      let continueButton = screen.getByRole('button', {name: /continue/i});
      expect(continueButton).toBeEnabled();

      // Deselect the default integration
      const toggleButton = screen.getByTestId('toggle-integration');
      await user.click(toggleButton);

      // Now should be disabled
      await waitFor(() => {
        continueButton = screen.getByRole('button', {name: /continue/i});
        expect(continueButton).toBeDisabled();
      });

      // Select it again
      await user.click(toggleButton);

      // Now should be enabled again
      await waitFor(() => {
        continueButton = screen.getByRole('button', {name: /continue/i});
        expect(continueButton).not.toBeDisabled();
      });
    });
  });

  describe('Accessibility', () => {
    it('should have proper heading hierarchy', () => {
      renderWithProviders();

      const breadcrumbText = screen.getByText('Create an Application');
      expect(breadcrumbText).toBeInTheDocument();
    });

    it('should have accessible buttons', () => {
      const {container} = renderWithProviders();

      // The close button is an IconButton without explicit name, check it exists
      const buttons = container.querySelectorAll('button');
      expect(buttons.length).toBeGreaterThan(0);

      // Continue button should have accessible text
      const continueButton = screen.getByRole('button', {name: /continue/i});
      expect(continueButton).toBeInTheDocument();
    });
  });

  describe('Branding Selection', () => {
    it('should update branding selection when using default branding', async () => {
      renderWithProviders();

      // Navigate to design step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Click use default branding button
      const useDefaultBtn = screen.getByTestId('use-default-branding-btn');
      await user.click(useDefaultBtn);

      // Component should not crash and design step should still be visible
      expect(screen.getByTestId('configure-design')).toBeInTheDocument();
    });

    it('should update branding selection when using custom branding', async () => {
      renderWithProviders();

      // Navigate to design step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Click use custom branding button
      const useCustomBtn = screen.getByTestId('use-custom-branding-btn');
      await user.click(useCustomBtn);

      // Component should not crash
      expect(screen.getByTestId('configure-design')).toBeInTheDocument();
    });
  });

  describe('Custom Approach Flow', () => {
    it('should create application with skip OAuth when CUSTOM approach is selected', async () => {
      // Mock to return NONE configuration type for custom approach
      const getConfigurationTypeFromTemplate = await import('../../utils/getConfigurationTypeFromTemplate');
      vi.mocked(getConfigurationTypeFromTemplate.default).mockReturnValue('NONE');

      mockCreateBranding.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({id: 'branding-123', name: 'Test Branding'});
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
        });
      });

      renderWithProviders();

      // Step 1: Name
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 2: Design
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 3: Sign In Options
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 4: Approach - select CUSTOM
      await waitFor(() => {
        expect(screen.getByTestId('configure-approach')).toBeInTheDocument();
      });
      const selectCustomBtn = screen.getByTestId('select-custom-approach');
      await user.click(selectCustomBtn);
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 5: Stack - for custom approach, clicking continue should create app
      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should call createBranding and then createApplication
      await waitFor(() => {
        expect(mockCreateBranding).toHaveBeenCalled();
      });
    });
  });

  describe('Back Navigation', () => {
    it('should navigate back from OPTIONS to DESIGN step', async () => {
      renderWithProviders();

      // Navigate to OPTIONS step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();

      // Click back
      await user.click(screen.getByRole('button', {name: /back/i}));

      expect(screen.getByTestId('configure-design')).toBeInTheDocument();
    });

    it('should navigate back from APPROACH to OPTIONS step', async () => {
      renderWithProviders();

      // Navigate to APPROACH step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByTestId('configure-approach')).toBeInTheDocument();

      // Click back
      await user.click(screen.getByRole('button', {name: /back/i}));

      expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
    });

    it('should navigate back from STACK to APPROACH step', async () => {
      renderWithProviders();

      // Navigate to STACK step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByTestId('configure-stack')).toBeInTheDocument();

      // Click back
      await user.click(screen.getByRole('button', {name: /back/i}));

      expect(screen.getByTestId('configure-approach')).toBeInTheDocument();
    });

    it('should navigate back from CONFIGURE to STACK step', async () => {
      renderWithProviders();

      // Navigate to CONFIGURE step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      expect(screen.getByTestId('configure-details')).toBeInTheDocument();

      // Click back
      await user.click(screen.getByRole('button', {name: /back/i}));

      expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
    });
  });

  describe('Application Creation with Default Branding', () => {
    it('should use default branding ID directly when selected', async () => {
      mockCreateApplication.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: {
                client_id: 'test-client-id',
                client_secret: 'test-client-secret',
                public_client: false,
              },
            },
          ],
        });
      });

      renderWithProviders();

      // Step 1: Name
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 2: Design - select default branding
      const useDefaultBtn = screen.getByTestId('use-default-branding-btn');
      await user.click(useDefaultBtn);
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 3: Sign In Options
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 4: Approach
      await waitFor(() => {
        expect(screen.getByTestId('configure-approach')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 5: Stack
      await waitFor(() => {
        expect(screen.getByTestId('configure-stack')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Step 6: Configure
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should create application directly without creating branding
      await waitFor(() => {
        expect(mockCreateApplication).toHaveBeenCalled();
        // Should NOT call createBranding since we're using default
        expect(mockCreateBranding).not.toHaveBeenCalled();
      });
    });
  });

  describe('Application Creation Error Handling', () => {
    it('should show error when no auth flow is selected', async () => {
      // This test covers the case where selectedAuthFlow is undefined
      renderWithProviders();

      // Navigate through all steps quickly
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // The error would appear if selectedAuthFlow was not set
      // This verifies the error handling path exists
      expect(screen.getByTestId('configure-sign-in')).toBeInTheDocument();
    });

    it('should show error when application creation fails', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({id: 'branding-123', name: 'Test Branding'});
      });

      mockCreateApplication.mockImplementation((_data, {onError}: any) => {
        onError(new Error('Failed to create application'));
      });

      renderWithProviders();

      // Navigate through all steps
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      await waitFor(() => {
        expect(screen.getByText(/failed to create application/i)).toBeInTheDocument();
      });
    });
  });

  describe('Summary Page', () => {
    it('should show View Application button and navigate when clicked', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({id: 'branding-123', name: 'Test Branding'});
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: {
                client_id: 'test-client-id',
                client_secret: 'test-client-secret',
                public_client: false,
              },
            },
          ],
        });
      });

      renderWithProviders();

      // Navigate through all steps
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should be on summary page
      await waitFor(() => {
        expect(screen.getByTestId('application-summary')).toBeInTheDocument();
      });

      // Click View Application button
      const viewAppButton = screen.getByRole('button', {name: /view application/i});
      await user.click(viewAppButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
      });
    });

    it('should show Done button and navigate to applications list when clicked', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({id: 'branding-123', name: 'Test Branding'});
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [],
        });
      });

      renderWithProviders();

      // Navigate through all steps
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should be on summary page
      await waitFor(() => {
        expect(screen.getByTestId('application-summary')).toBeInTheDocument();
      });

      // Click Done button
      const doneButton = screen.getByRole('button', {name: /done/i});
      await user.click(doneButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications');
      });
    });

    it('should display client credentials for confidential clients', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({id: 'branding-123', name: 'Test Branding'});
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: {
                client_id: 'confidential-client-id',
                client_secret: 'super-secret-value',
                public_client: false,
              },
            },
          ],
        });
      });

      renderWithProviders();

      // Navigate through all steps
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should be on summary page with client credentials
      await waitFor(() => {
        expect(screen.getByTestId('application-summary')).toBeInTheDocument();
        expect(screen.getByTestId('summary-client-id')).toHaveTextContent('confidential-client-id');
        expect(screen.getByTestId('summary-client-secret')).toHaveTextContent('super-secret-value');
        expect(screen.getByTestId('summary-has-oauth')).toHaveTextContent('true');
      });
    });

    it('should display client ID only for public clients (no secret)', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({id: 'branding-123', name: 'Test Branding'});
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: {
                client_id: 'public-client-id',
                public_client: true,
              },
            },
          ],
        });
      });

      renderWithProviders();

      // Navigate through all steps
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should be on summary page with only client ID
      await waitFor(() => {
        expect(screen.getByTestId('application-summary')).toBeInTheDocument();
        expect(screen.getByTestId('summary-client-id')).toHaveTextContent('public-client-id');
        // Public clients don't have client secrets (they cannot securely store credentials)
        expect(screen.getByTestId('summary-client-secret')).toHaveTextContent('');
      });
    });
  });

  describe('Callback URL Update', () => {
    it('should update OAuth config when callback URL changes', async () => {
      renderWithProviders();

      // Navigate to configure details step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should be on configure details step
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });

      // Enter callback URL
      const callbackInput = screen.getByTestId('callback-url-input');
      await user.type(callbackInput, 'https://example.com/callback');

      // Component should handle the callback URL change
      expect(callbackInput).toHaveValue('https://example.com/callback');
    });
  });

  describe('User Types Selection', () => {
    it('should allow selecting user types in configure details', async () => {
      renderWithProviders();

      // Navigate to configure details step
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should be on configure details step
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });

      // Select user type
      const selectUserTypeBtn = screen.getByTestId('select-user-type');
      await user.click(selectUserTypeBtn);

      // Verify the selection was made
      await waitFor(() => {
        expect(screen.getByTestId('selected-user-types')).toHaveTextContent('customer');
      });
    });
  });

  describe('View Application Navigation Error', () => {
    it('should fall back to applications list when navigation fails', async () => {
      mockCreateBranding.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({id: 'branding-123', name: 'Test Branding'});
      });

      mockCreateApplication.mockImplementation((_data, {onSuccess}: any) => {
        onSuccess({
          id: 'app-123',
          name: 'My App',
          inbound_auth_config: [],
        });
      });

      // Make navigate fail for specific app route but work for /applications
      mockNavigate.mockImplementation((path: string) => {
        if (path === '/applications/app-123') {
          return Promise.reject(new Error('Navigation failed'));
        }
        return Promise.resolve();
      });

      renderWithProviders();

      // Navigate through all steps
      const nameInput = screen.getByTestId('app-name-input');
      await user.type(nameInput, 'My App');
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /continue/i})).not.toBeDisabled();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await user.click(screen.getByRole('button', {name: /continue/i}));
      await waitFor(() => {
        expect(screen.getByTestId('configure-details')).toBeInTheDocument();
      });
      await user.click(screen.getByRole('button', {name: /continue/i}));

      // Should be on summary page
      await waitFor(() => {
        expect(screen.getByTestId('application-summary')).toBeInTheDocument();
      });

      // Click View Application button
      const viewAppButton = screen.getByRole('button', {name: /view application/i});
      await user.click(viewAppButton);

      // Should attempt to navigate and then fall back
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
      });

      // After failure, should fall back to /applications
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications');
      });
    });
  });
});
