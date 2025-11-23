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
import {useState, useCallback, useMemo} from 'react';
import {useTranslation} from 'react-i18next';
import ConfigureSignInOptions from '../components/create-applications/ConfigureSignInOptions';
import ConfigureDesign from '../components/create-applications/ConfigureDesign';
import ConfigureName from '../components/create-applications/ConfigureName';
import ConfigureOAuth from '../components/create-applications/ConfigureOAuth';
import {getDefaultOAuthConfig} from '../models/oauth';
import Preview from '../components/create-applications/Preview';
import ApplicationSummary from '../components/create-applications/ApplicationSummary';
import useCreateApplication from '../api/useCreateApplication';
import resolveAuthFlowGraphId, {USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY} from '../utils/resolveAuthFlowGraphId';
import useIdentityProviders from '../../integrations/api/useIdentityProviders';
import type {CreateApplicationRequest} from '../models/requests';
import type {OAuth2Config} from '../models/oauth';
import useCreateBranding from '../../branding/api/useCreateBranding';
import type {CreateBrandingRequest} from '../../branding/models/requests';
import BrandingConstants from '../constants/branding-contants';
import type {Application} from '../models/application';

type Step = 'name' | 'design' | 'options' | 'configure' | 'summary';

export default function ApplicationCreatePage(): JSX.Element {
  const {t} = useTranslation();
  
  // Memoize steps to prevent unnecessary re-renders
  const steps: Record<Step, {label: string; order: number}> = useMemo(() => ({
    name: {label: t('applications:onboarding.steps.name'), order: 1},
    design: {label: t('applications:onboarding.steps.design'), order: 2},
    options: {label: t('applications:onboarding.steps.options'), order: 3},
    configure: {label: t('applications:onboarding.steps.configure'), order: 4},
    summary: {label: t('applications:onboarding.steps.summary'), order: 5},
  }), [t]);
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
    configure: true, // OAuth config step is always initialized with sensible defaults (client_credentials grant type), so it is ready from the start
    summary: true, // Summary step is always ready
  });
  const [useDefaultBranding, setUseDefaultBranding] = useState<boolean>(false);
  const [defaultBrandingId, setDefaultBrandingId] = useState<string | undefined>(undefined);
  // Initialize with default OAuth config (client_credentials grant type) for convenience, but can be skipped during creation
  const [oauthConfig, setOAuthConfig] = useState<OAuth2Config | null>(getDefaultOAuthConfig());
  const [clientSecret, setClientSecret] = useState<string>('');
  const [clientId, setClientId] = useState<string>('');
  const [hasOAuthValidationErrors, setHasOAuthValidationErrors] = useState(false);
  const [hasOAuthConfig, setHasOAuthConfig] = useState(false);
  const [createdApplicationId, setCreatedApplicationId] = useState<string | null>(null);

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
      case 'summary':
        // Don't allow going back from summary - navigate to applications list instead
        handleClose();
        break;
      default:
        break;
    }
  };

  const handleCreateApplication = (skipOAuthConfig = false) => {
    // Clear any previous errors
    setError(null);

    // Track if OAuth configs were selected
    const oauthConfigSelected = !skipOAuthConfig && oauthConfig !== null;
    setHasOAuthConfig(oauthConfigSelected);

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
        // Only include OAuth config if not skipping
        ...(!skipOAuthConfig && {
          inbound_auth_config: [
            {
              type: 'oauth2',
              config: oauthConfig,
            },
          ],
        }),
      };

      createApplication.mutate(applicationData, {
        onSuccess: (createdApp: Application): void => {
          // Store the created application ID
          setCreatedApplicationId(createdApp.id);

          // Extract OAuth credentials from response
          const oauthConfigFromResponse = createdApp.inbound_auth_config?.find((config) => config.type === 'oauth2');
          const isPublicClient = oauthConfigFromResponse?.config?.public_client ?? false;
          const secret = oauthConfigFromResponse?.config?.client_secret;
          const cId = oauthConfigFromResponse?.config?.client_id;

          // Always store client_id if available (public clients still have client_id)
          if (cId) {
            setClientId(cId);
          }

          // Only store client_secret for confidential clients (public clients don't have secrets)
          if (!isPublicClient && secret) {
            setClientSecret(secret);
          }

          // Navigate to summary step instead of showing dialog
          setCurrentStep('summary');
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

  const handleStepReadyChange = useCallback((step: Step, isReady: boolean) => {
    setStepReady((prev) => ({
      ...prev,
      [step]: isReady,
    }));
  }, []);

  const handleOAuthValidationErrorsChange = useCallback((hasErrors: boolean) => {
    setHasOAuthValidationErrors(hasErrors);
  }, []);

  // Memoized callbacks for each step's onReadyChange
  const handleNameStepReadyChange = useCallback((isReady: boolean) => {
    handleStepReadyChange('name', isReady);
  }, [handleStepReadyChange]);

  const handleDesignStepReadyChange = useCallback((isReady: boolean) => {
    handleStepReadyChange('design', isReady);
  }, [handleStepReadyChange]);

  const handleOptionsStepReadyChange = useCallback((isReady: boolean) => {
    handleStepReadyChange('options', isReady);
  }, [handleStepReadyChange]);

  const handleConfigureStepReadyChange = useCallback((isReady: boolean) => {
    handleStepReadyChange('configure', isReady);
  }, [handleStepReadyChange]);

  const renderStepContent = () => {
    switch (currentStep) {
      case 'name':
        return (
          <ConfigureName
            appName={appName}
            onAppNameChange={setAppName}
            onReadyChange={handleNameStepReadyChange}
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
            onReadyChange={handleDesignStepReadyChange}
            onBrandingSelectionChange={handleBrandingSelectionChange}
          />
        );

      case 'options':
        return (
          <ConfigureSignInOptions
            integrations={integrations}
            onIntegrationToggle={handleIntegrationToggle}
            onReadyChange={handleOptionsStepReadyChange}
          />
        );

      case 'configure':
        return (
          <ConfigureOAuth
            oauthConfig={oauthConfig}
            onOAuthConfigChange={setOAuthConfig}
            onReadyChange={handleConfigureStepReadyChange}
            onValidationErrorsChange={handleOAuthValidationErrorsChange}
          />
        );

      case 'summary':
        return (
          <ApplicationSummary
            appName={appName}
            appLogo={appLogo}
            selectedColor={selectedColor}
            clientId={clientId}
            clientSecret={clientSecret}
            hasOAuthConfig={hasOAuthConfig}
            applicationId={createdApplicationId}
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
        <Box sx={{flex: currentStep === 'name' || currentStep === 'summary' ? 1 : '0 0 50%', display: 'flex', flexDirection: 'column'}}>
          {/* Header with close button and breadcrumb */}
          <Box sx={{p: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
            <Stack direction="row" alignItems="center" spacing={2}>
              {currentStep !== 'summary' && (
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
              )}
              {currentStep !== 'summary' && (
                <Breadcrumbs separator={<ChevronRight size={16} />} aria-label="breadcrumb">
                  {getBreadcrumbSteps().map((step, index, array) => {
                    const isLast = index === array.length - 1;
                    // Don't allow clicking on steps if we're on summary (app is already created)
                    const isClickable = !isLast && (currentStep as Step) !== 'summary';

                    return isClickable ? (
                      <Typography key={step} variant="h5" onClick={() => setCurrentStep(step)} sx={{cursor: 'pointer'}}>
                        {steps[step].label}
                      </Typography>
                    ) : (
                      <Typography key={step} variant="h5" color="text.primary">
                        {steps[step].label}
                      </Typography>
                    );
                  })}
                </Breadcrumbs>
              )}
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
                py: currentStep === 'summary' ? 2 : 8,
                px: 20,
                mx: currentStep === 'name' || currentStep === 'summary' ? 'auto' : 0,
                ...(currentStep === 'summary' && {
                  alignItems: 'center',
                }),
              }}
            >
              <Box
                sx={{
                  width: '100%',
                  maxWidth: currentStep === 'summary' ? 600 : 800,
                  textAlign: currentStep === 'summary' ? 'center' : 'left',
                  display: 'flex',
                  flexDirection: 'column',
                }}
              >
                {renderStepContent()}

                {/* Error Alert */}
                {error && (
                  <Alert severity="error" sx={{mt: 3}} onClose={() => setError(null)}>
                    {error}
                  </Alert>
                )}

                {/* Navigation buttons */}
                {currentStep === 'summary' ? (
                  <Box
                    sx={{
                      mt: 4,
                      display: 'flex',
                      justifyContent: 'flex-end',
                      gap: 2,
                    }}
                  >
                    {createdApplicationId && (
                      <Button
                        variant="outlined"
                        sx={{minWidth: 160, height: 36}}
                        onClick={() => {
                          (async () => {
                            await navigate(`/applications/${createdApplicationId}`);
                          })().catch(() => {
                            handleClose();
                          });
                        }}
                      >
                        {t('applications:onboarding.summary.viewApplication')}
                      </Button>
                    )}
                    <Button
                      variant="contained"
                      sx={{minWidth: 120, height: 36, bgcolor: selectedColor, '&:hover': {bgcolor: selectedColor}}}
                      onClick={handleClose}
                    >
                      {t('common:actions.done')}
                    </Button>
                  </Box>
                ) : (
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
                        {t('common:actions.back')}
                      </Button>
                    )}

                    {currentStep === 'configure' ? (
                      <Stack direction="row" spacing={1}>
                        <Button
                          variant="outlined"
                          onClick={() => handleCreateApplication(true)}
                          disabled={createApplication.isPending || createBranding.isPending}
                          sx={{minWidth: 140, height: 36, whiteSpace: 'nowrap'}}
                        >
                          {createApplication.isPending || createBranding.isPending
                            ? t('applications:onboarding.creating')
                            : t('applications:onboarding.skipAndCreate')}
                        </Button>
                        <Button
                          variant="contained"
                          sx={{minWidth: 160, height: 36, whiteSpace: 'nowrap', bgcolor: selectedColor, '&:hover': {bgcolor: selectedColor}}}
                          onClick={() => handleCreateApplication(false)}
                          disabled={createApplication.isPending || createBranding.isPending || hasOAuthValidationErrors}
                        >
                          {createApplication.isPending || createBranding.isPending
                            ? t('applications:onboarding.creating')
                            : t('applications:onboarding.createApplication')}
                        </Button>
                      </Stack>
                    ) : (
                      <Button
                        variant="contained"
                        disabled={!stepReady[currentStep]}
                        sx={{minWidth: 100}}
                        onClick={handleNextStep}
                      >
                        {t('common:actions.continue')}
                      </Button>
                    )}
                  </Box>
                )}
              </Box>
            </Box>
          </Box>
        </Box>
        {/* Right side - Preview (show from design step onwards, but not on summary) */}
        {currentStep !== 'name' && currentStep !== 'summary' && (
          <Box sx={{flex: '0 0 50%', display: 'flex', flexDirection: 'column', p: 5}}>
            <Preview appName={appName} appLogo={appLogo} selectedColor={selectedColor} integrations={integrations} />
          </Box>
        )}
      </Box>
    </Box>
  );
}
