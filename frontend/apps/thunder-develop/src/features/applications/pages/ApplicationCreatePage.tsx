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

import {Box, Stack, Button, IconButton, LinearProgress, Breadcrumbs, Typography, Alert} from '@wso2/oxygen-ui';
import {X, ChevronRight} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useNavigate} from 'react-router';
import {useState} from 'react';
import ConfigureSignInOptions from '../components/create-applications/ConfigureSignInOptions';
import ConfigureDesign from '../components/create-applications/ConfigureDesign';
import ConfigureName from '../components/create-applications/ConfigureName';
import ConfigureRedirectURIs from '../components/create-applications/ConfigureRedirectURIs';
import Preview from '../components/create-applications/Preview';
import useCreateApplication from '../api/useCreateApplication';
import resolveAuthFlowGraphId, {USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY} from '../utils/resolveAuthFlowGraphId';
import useIdentityProviders from '../../integrations/api/useIdentityProviders';
import type {CreateApplicationRequest} from '../models/requests';
import type {OAuth2Config} from '../models/oauth';
import useCreateBranding from '../../branding/api/useCreateBranding';
import type {CreateBrandingRequest} from '../../branding/models/requests';
import BrandingConstants from '../constants/branding-contants';

type Step = 'name' | 'design' | 'options' | 'configure';

const steps: Record<Step, {label: string; order: number}> = {
  name: {label: 'Create an Application', order: 1},
  design: {label: 'Design', order: 2},
  options: {label: 'Sign In Options', order: 3},
  configure: {label: 'Configure', order: 4},
};

export default function ApplicationCreatePage(): JSX.Element {
  const navigate = useNavigate();
  const createApplication = useCreateApplication();
  const createBranding = useCreateBranding();
  const {data: identityProviders} = useIdentityProviders();
  const [currentStep, setCurrentStep] = useState<Step>('name');
  const [appName, setAppName] = useState('');
  const [selectedColor, setSelectedColor] = useState('#1976d2');
  const [appLogo, setAppLogo] = useState<string | null>(null);
  const [integrations, setIntegrations] = useState<Record<string, boolean>>({
    [USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY]: true,
  });
  const [error, setError] = useState<string | null>(null);
  const [stepReady, setStepReady] = useState<Record<Step, boolean>>({
    name: false,
    design: true,
    options: true, // Start as true since username/password is enabled by default
    configure: false, // Start as false since no redirect URIs are added by default
  });
  const [useDefaultBranding, setUseDefaultBranding] = useState<boolean>(false);
  const [defaultBrandingId, setDefaultBrandingId] = useState<string | undefined>(undefined);
  const [redirectURIs, setRedirectURIs] = useState<string[]>([]);

  const handleClose = () => {
    (async () => {
      await navigate('/applications');
    })().catch(() => {
      // TODO: Log the errors
      // Tracker: https://github.com/asgardeo/thunder/issues/618
    });
  };

  const handleLogoSelect = (logoUrl: string) => {
    setAppLogo(logoUrl);
  };

  const handleIntegrationToggle = (integrationId: string) => {
    setIntegrations((prev) => ({
      ...prev,
      [integrationId]: !prev[integrationId],
    }));
  };

  const handleBrandingSelectionChange = (useDefault: boolean, brandingId?: string) => {
    setUseDefaultBranding(useDefault);
    setDefaultBrandingId(brandingId);
  };

  const handleNextStep = () => {
    switch (currentStep) {
      case 'name':
        setCurrentStep('design');
        break;
      case 'design':
        setCurrentStep('options');
        break;
      case 'options':
        setCurrentStep('configure');
        break;
      default:
        break;
    }
  };

  const handlePrevStep = () => {
    switch (currentStep) {
      case 'design':
        setCurrentStep('name');
        break;
      case 'options':
        setCurrentStep('design');
        break;
      case 'configure':
        setCurrentStep('options');
        break;
      default:
        break;
    }
  };

  const handleCreateApplication = () => {
    // Clear any previous errors
    setError(null);

    // Check if username/password is enabled
    const hasUsernamePassword = integrations[USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY] ?? false;

    // Get selected identity providers with their full data from API
    const selectedIdentityProviders = identityProviders?.filter((idp) => integrations[idp.id]) ?? [];

    // Resolve the appropriate auth flow graph ID based on selected options
    const authFlowGraphId = resolveAuthFlowGraphId({
      hasUsernamePassword,
      identityProviders: selectedIdentityProviders,
    });

    // Function to create the application with branding_id
    const createApplicationWithBranding = (brandingId: string) => {
      // Build the application request payload
      const applicationData: CreateApplicationRequest = {
        name: appName,
        logo_url: appLogo ?? undefined,
        auth_flow_graph_id: authFlowGraphId,
        user_attributes: ['given_name', 'family_name', 'email', 'groups'],
        branding_id: brandingId,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              redirect_uris: redirectURIs,
            } as OAuth2Config,
          },
        ],
      };

      createApplication.mutate(applicationData, {
        onSuccess: (): void => {
          // Navigate to the application details page
          (async () => {
            await navigate(`/applications`);
          })().catch(() => {
            // Fallback to applications list if navigation fails
            handleClose();
          });
        },
        onError: (err: Error) => {
          setError(err.message ?? 'Failed to create application. Please try again.');
        },
      });
    };

    // If using DEFAULT branding, use its ID directly
    if (useDefaultBranding && defaultBrandingId) {
      createApplicationWithBranding(defaultBrandingId);
      return;
    }

    // Otherwise, create a new branding with the selected color
    // If there's no Default branding preset, create one named "Default"
    // If Default branding exists but user customized, create one with the app name
    const brandingName = !defaultBrandingId ? BrandingConstants.DEFAULT_BRANDING_NAME : appName;
    const brandingData: CreateBrandingRequest = {
      displayName: brandingName,
      preferences: {
        theme: {
          activeColorScheme: 'light',
          colorSchemes: {
            light: {
              colors: {
                primary: {
                  main: selectedColor,
                  dark: selectedColor,
                  contrastText: '#ffffff',
                },
                secondary: {
                  main: selectedColor,
                  dark: selectedColor,
                  contrastText: '#ffffff',
                },
              },
            },
          },
        },
      },
    };

    createBranding.mutate(brandingData, {
      onSuccess: (branding): void => {
        // Create application with the new branding ID
        createApplicationWithBranding(branding.id);
      },
      onError: (err: Error) => {
        setError(err.message ?? 'Failed to create branding. Please try again.');
      },
    });
  };

  const handleStepReadyChange = (step: Step, isReady: boolean) => {
    setStepReady((prev) => ({
      ...prev,
      [step]: isReady,
    }));
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 'name':
        return (
          <ConfigureName
            appName={appName}
            onAppNameChange={setAppName}
            onReadyChange={(isReady) => handleStepReadyChange('name', isReady)}
          />
        );

      case 'design':
        return (
          <ConfigureDesign
            appLogo={appLogo}
            selectedColor={selectedColor}
            appName={appName}
            onLogoSelect={handleLogoSelect}
            onInitialLogoLoad={handleLogoSelect}
            onColorSelect={setSelectedColor}
            onReadyChange={(isReady) => handleStepReadyChange('design', isReady)}
            onBrandingSelectionChange={handleBrandingSelectionChange}
          />
        );

      case 'options':
        return (
          <ConfigureSignInOptions
            integrations={integrations}
            onIntegrationToggle={handleIntegrationToggle}
            onReadyChange={(isReady) => handleStepReadyChange('options', isReady)}
          />
        );

      case 'configure':
        return (
          <ConfigureRedirectURIs
            redirectURIs={redirectURIs}
            onRedirectURIsChange={setRedirectURIs}
            onReadyChange={(isReady) => handleStepReadyChange('configure', isReady)}
          />
        );

      default:
        return null;
    }
  };

  const getStepProgress = () => {
    const stepNames = Object.keys(steps) as Step[];
    return ((stepNames.indexOf(currentStep) + 1) / stepNames.length) * 100;
  };

  const getBreadcrumbSteps = () => {
    const stepNames = Object.keys(steps) as Step[];
    const currentIndex = stepNames.indexOf(currentStep);
    return stepNames.slice(0, currentIndex + 1);
  };

  return (
    <Box sx={{minHeight: '100vh', display: 'flex', flexDirection: 'column'}}>
      {/* Progress bar at the very top */}
      <LinearProgress variant="determinate" value={getStepProgress()} sx={{height: 6}} />

      <Box sx={{flex: 1, display: 'flex', flexDirection: 'row'}}>
        <Box sx={{flex: currentStep === 'name' ? 1 : '0 0 50%', display: 'flex', flexDirection: 'column'}}>
          {/* Header with close button and breadcrumb */}
          <Box sx={{p: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
            <Stack direction="row" alignItems="center" spacing={2}>
              <IconButton
                onClick={handleClose}
                sx={{
                  bgcolor: 'background.paper',
                  '&:hover': {bgcolor: 'action.hover'},
                  boxShadow: 1,
                }}
              >
                <X size={24} />
              </IconButton>
              <Breadcrumbs separator={<ChevronRight size={16} />} aria-label="breadcrumb">
                {getBreadcrumbSteps().map((step, index, array) => {
                  const isLast = index === array.length - 1;
                  const isClickable = !isLast;

                  return isClickable ? (
                    <Typography key={step} variant="h5" onClick={() => setCurrentStep(step)}>
                      {steps[step].label}
                    </Typography>
                  ) : (
                    <Typography key={step} variant="h5" color="text.primary">
                      {steps[step].label}
                    </Typography>
                  );
                })}
              </Breadcrumbs>
            </Stack>
          </Box>

          {/* Main content */}
          <Box sx={{flex: 1, display: 'flex', minHeight: 0}}>
            {/* Left side - Form content */}
            <Box
              sx={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                py: 8,
                px: 20,
                mx: currentStep === 'name' ? 'auto' : 0,
              }}
            >
              <Box sx={{width: '100%', maxWidth: 800, textAlign: 'left'}}>
                {renderStepContent()}

                {/* Error Alert */}
                {error && (
                  <Alert severity="error" sx={{mt: 3}} onClose={() => setError(null)}>
                    {error}
                  </Alert>
                )}

                {/* Navigation buttons */}
                <Box
                  sx={{
                    mt: 4,
                    display: 'flex',
                    justifyContent: currentStep === 'name' ? 'flex-start' : 'space-between',
                    gap: 2,
                  }}
                >
                  {currentStep !== 'name' && (
                    <Button
                      variant="outlined"
                      onClick={handlePrevStep}
                      sx={{minWidth: 100}}
                      disabled={createApplication.isPending || createBranding.isPending}
                    >
                      Back
                    </Button>
                  )}

                  {currentStep === 'configure' ? (
                    <Button
                      variant="contained"
                      sx={{minWidth: 150, bgcolor: selectedColor, '&:hover': {bgcolor: selectedColor}}}
                      onClick={handleCreateApplication}
                      disabled={createApplication.isPending || createBranding.isPending || !stepReady.configure}
                    >
                      {createApplication.isPending || createBranding.isPending ? 'Creating...' : 'Create Application'}
                    </Button>
                  ) : (
                    <Button
                      variant="contained"
                      disabled={!stepReady[currentStep]}
                      sx={{minWidth: 100}}
                      onClick={handleNextStep}
                    >
                      Continue
                    </Button>
                  )}
                </Box>
              </Box>
            </Box>
          </Box>
        </Box>
        {/* Right side - Preview (show from design step onwards) */}
        {currentStep !== 'name' && (
          <Box sx={{flex: '0 0 50%', display: 'flex', flexDirection: 'column', p: 5}}>
            <Preview appName={appName} appLogo={appLogo} selectedColor={selectedColor} integrations={integrations} />
          </Box>
        )}
      </Box>
    </Box>
  );
}
