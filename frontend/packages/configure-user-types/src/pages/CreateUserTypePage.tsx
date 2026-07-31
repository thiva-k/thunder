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

import {
  OrganizationUnitPickerScreen,
  useGetOrganizationUnit,
  useHasMultipleOUs,
} from '@thunderid/configure-organization-units';
import {useLogger} from '@thunderid/logger/react';
import {
  Box,
  Stack,
  Button,
  CircularProgress,
  IconButton,
  LinearProgress,
  Typography,
  Alert,
  Snackbar,
  AppBreadcrumbs,
} from '@wso2/oxygen-ui';
import {Home, X} from '@wso2/oxygen-ui-icons-react';
import {useState, useCallback, useEffect, useMemo} from 'react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import useCreateUserType from '../api/useCreateUserType';
import ConfigureName from '../components/create-user-type/ConfigureName';
import ConfigureProperties from '../components/create-user-type/ConfigureProperties';
import useUserTypeCreate from '../contexts/UserTypeCreate/useUserTypeCreate';
import useUserTypeRoutes from '../hooks/useUserTypeRoutes';
import {UserTypeCreateFlowStep} from '../models/user-type-create-flow';
import type {PropertyDefinition, UserTypeDefinition, CreateUserTypeRequest} from '../types/user-types';

export default function CreateUserTypePage(): JSX.Element {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const logger = useLogger('CreateUserTypePage');
  const createUserTypeMutation = useCreateUserType();
  const routes = useUserTypeRoutes();

  const {
    currentStep,
    setCurrentStep,
    name,
    setName,
    ouId,
    setOuId,
    allowSelfRegistration,
    setAllowSelfRegistration,
    properties,
    setProperties,
    enumInput,
    setEnumInput,
    displayAttribute,
    setDisplayAttribute,
    error,
    setError,
  } = useUserTypeCreate();

  const {hasMultipleOUs, isLoading: isOuLoading, ouList} = useHasMultipleOUs();

  // The organization unit is the wizard's first step whenever there's a choice to make. Single-OU
  // deployments never need it, so once that's known, resolve it automatically and skip straight
  // past it.
  useEffect(() => {
    if (isOuLoading || hasMultipleOUs || currentStep !== UserTypeCreateFlowStep.ORGANIZATION_UNIT) return;
    setOuId(ouList[0]?.id ?? '');
    setCurrentStep(UserTypeCreateFlowStep.NAME);
  }, [isOuLoading, hasMultipleOUs, ouList, currentStep, setOuId, setCurrentStep]);

  // The organization unit whose name is shown in the Details step's summary chip.
  const resolvedOuId = hasMultipleOUs ? ouId : ouList[0]?.id;
  const {data: resolvedOrganizationUnit, isLoading: isResolvedOuLoading} = useGetOrganizationUnit(
    resolvedOuId,
    Boolean(resolvedOuId),
  );

  const activeSteps = useMemo((): UserTypeCreateFlowStep[] => {
    const base: UserTypeCreateFlowStep[] = [];
    if (hasMultipleOUs) base.push(UserTypeCreateFlowStep.ORGANIZATION_UNIT);
    base.push(UserTypeCreateFlowStep.NAME, UserTypeCreateFlowStep.PROPERTIES);
    return base;
  }, [hasMultipleOUs]);

  const steps: Partial<Record<UserTypeCreateFlowStep, {label: string}>> = useMemo(() => {
    const map: Partial<Record<UserTypeCreateFlowStep, {label: string}>> = {};
    if (hasMultipleOUs) {
      map.ORGANIZATION_UNIT = {label: t('userTypes:createWizard.steps.organizationUnit', 'Organization Unit')};
    }
    map.NAME = {label: t('userTypes:createWizard.steps.name', 'Details')};
    map.PROPERTIES = {label: t('userTypes:createWizard.steps.properties', 'Properties')};
    return map;
  }, [t, hasMultipleOUs]);

  const [validationError, setValidationError] = useState<string | null>(null);
  const [snackbarOpen, setSnackbarOpen] = useState(false);

  const [stepReady, setStepReady] = useState<Record<UserTypeCreateFlowStep, boolean>>({
    ORGANIZATION_UNIT: false,
    NAME: false,
    PROPERTIES: false,
  });

  const handleClose = (): void => {
    void navigate(routes.list());
  };

  const handleStepReadyChange = useCallback((step: UserTypeCreateFlowStep, isReady: boolean): void => {
    setStepReady((prev) => ({
      ...prev,
      [step]: isReady,
    }));
  }, []);

  const handleNameStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(UserTypeCreateFlowStep.NAME, isReady);
    },
    [handleStepReadyChange],
  );

  const handlePropertiesStepReadyChange = useCallback(
    (isReady: boolean): void => {
      handleStepReadyChange(UserTypeCreateFlowStep.PROPERTIES, isReady);
    },
    [handleStepReadyChange],
  );

  const handleSubmit = async (): Promise<void> => {
    setValidationError(null);
    setError(null);

    // Validate
    if (!name.trim()) {
      setValidationError(t('userTypes:validationErrors.nameRequired'));
      setSnackbarOpen(true);
      return;
    }

    const trimmedOuId = ouId.trim();
    if (!trimmedOuId) {
      setValidationError(t('userTypes:validationErrors.ouIdRequired'));
      setSnackbarOpen(true);
      return;
    }

    const validProperties = properties.filter((prop) => prop.name.trim());
    if (validProperties.length === 0) {
      setValidationError(t('userTypes:validationErrors.propertiesRequired'));
      setSnackbarOpen(true);
      return;
    }

    // Check for duplicate property names
    const propertyNames = validProperties.map((prop) => prop.name.trim());
    const duplicates = propertyNames.filter((propName, index) => propertyNames.indexOf(propName) !== index);
    if (duplicates.length > 0) {
      setValidationError(t('userTypes:validationErrors.duplicateProperties', {duplicates: duplicates.join(', ')}));
      setSnackbarOpen(true);
      return;
    }

    // Convert properties to schema definition
    const schema: UserTypeDefinition = {};
    validProperties.forEach((prop) => {
      const actualType = prop.type === 'enum' ? 'string' : prop.type;

      const propDef: Partial<PropertyDefinition> = {
        type: actualType,
        required: prop.required,
        ...(prop.displayName.trim() ? {displayName: prop.displayName.trim()} : {}),
      };

      if (actualType === 'string' || actualType === 'number') {
        if (prop.unique) {
          (propDef as {unique?: boolean}).unique = true;
        }
        if (prop.credential) {
          (propDef as {credential?: boolean}).credential = true;
        }
      }

      if (actualType === 'string') {
        if (prop.type === 'enum' || prop.enum.length > 0) {
          (propDef as {enum?: string[]}).enum = prop.enum;
        }
        if (prop.regex.trim()) {
          (propDef as {regex?: string}).regex = prop.regex;
        }
      }

      if (actualType === 'array') {
        (propDef as {items?: {type: string}}).items = {type: 'string'};
      } else if (actualType === 'object') {
        (propDef as {properties?: Record<string, PropertyDefinition>}).properties = {};
      }

      schema[prop.name.trim()] = propDef as PropertyDefinition;
    });

    const requestBody: CreateUserTypeRequest = {
      name: name.trim(),
      ouId: trimmedOuId,
      schema,
    };

    if (allowSelfRegistration) {
      requestBody.allowSelfRegistration = true;
    }

    if (displayAttribute) {
      requestBody.systemAttributes = {display: displayAttribute};
    }

    try {
      await createUserTypeMutation.mutateAsync(requestBody);
      await navigate(routes.list());
    } catch (submitError) {
      logger.error('Failed to create user type or navigate', {error: submitError, userTypeName: name});
    }
  };

  const handleNextStep = (): void => {
    switch (currentStep) {
      case UserTypeCreateFlowStep.ORGANIZATION_UNIT:
        setCurrentStep(UserTypeCreateFlowStep.NAME);
        break;
      case UserTypeCreateFlowStep.NAME:
        setCurrentStep(UserTypeCreateFlowStep.PROPERTIES);
        break;
      case UserTypeCreateFlowStep.PROPERTIES:
        handleSubmit().catch(() => {
          // Error handled in handleSubmit
        });
        break;
      default:
        break;
    }
  };

  const handlePrevStep = (): void => {
    switch (currentStep) {
      case UserTypeCreateFlowStep.NAME:
        if (hasMultipleOUs) setCurrentStep(UserTypeCreateFlowStep.ORGANIZATION_UNIT);
        break;
      case UserTypeCreateFlowStep.PROPERTIES:
        setCurrentStep(UserTypeCreateFlowStep.NAME);
        break;
      default:
        break;
    }
  };

  const renderStepContent = (): JSX.Element | null => {
    switch (currentStep) {
      case UserTypeCreateFlowStep.NAME:
        return (
          <ConfigureName
            name={name}
            onNameChange={setName}
            onReadyChange={handleNameStepReadyChange}
            hasMultipleOUs={hasMultipleOUs}
            organizationUnitName={resolvedOrganizationUnit?.name}
            organizationUnitLogoUrl={resolvedOrganizationUnit?.logoUrl}
            isOrganizationUnitLoading={isResolvedOuLoading}
            onChangeOu={() => setCurrentStep(UserTypeCreateFlowStep.ORGANIZATION_UNIT)}
            allowSelfRegistration={allowSelfRegistration}
            onAllowSelfRegistrationChange={setAllowSelfRegistration}
          />
        );
      case UserTypeCreateFlowStep.PROPERTIES:
        return (
          <ConfigureProperties
            properties={properties}
            onPropertiesChange={setProperties}
            enumInput={enumInput}
            onEnumInputChange={setEnumInput}
            displayAttribute={displayAttribute}
            onDisplayAttributeChange={setDisplayAttribute}
            onReadyChange={handlePropertiesStepReadyChange}
            userTypeName={name.trim()}
          />
        );
      default:
        return null;
    }
  };

  const getStepProgress = (): number => {
    const currentIndex = activeSteps.indexOf(currentStep);
    return ((currentIndex + 1) / activeSteps.length) * 100;
  };

  const getBreadcrumbSteps = (): UserTypeCreateFlowStep[] => {
    const currentIndex = activeSteps.indexOf(currentStep);
    return activeSteps.slice(0, currentIndex + 1);
  };

  const handleCloseSnackbar = () => {
    setSnackbarOpen(false);
  };

  const isLastStep = currentStep === UserTypeCreateFlowStep.PROPERTIES;
  // The Properties step uses a two-panel builder that needs more horizontal room
  // than the single-column Name/General forms.
  const isPropertiesStep = currentStep === UserTypeCreateFlowStep.PROPERTIES;

  if (currentStep === UserTypeCreateFlowStep.ORGANIZATION_UNIT) {
    if (isOuLoading || !hasMultipleOUs) {
      return (
        <Box sx={{minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center'}}>
          <CircularProgress />
        </Box>
      );
    }

    return (
      <OrganizationUnitPickerScreen
        icon={<Home size={26} />}
        title={t('userTypes:createWizard.organizationUnit.title', 'Where should this user type belong?')}
        subtitle={t(
          'userTypes:createWizard.organizationUnit.subtitle',
          "Choose the organization unit that will own this user type. You can't change this once created.",
        )}
        value={ouId}
        onChange={setOuId}
        onBack={handleClose}
        onContinue={handleNextStep}
        backLabel={t('common:actions.back', 'Back')}
        continueLabel={t('common:actions.continue', 'Continue')}
      />
    );
  }

  return (
    <Box sx={{minHeight: '100vh', display: 'flex', flexDirection: 'column'}}>
      {/* Progress bar at the very top */}
      <LinearProgress variant="determinate" value={getStepProgress()} sx={{height: 6}} />

      <Box sx={{flex: 1, display: 'flex', flexDirection: 'column'}}>
        {/* Header with close button and breadcrumb */}
        <Box sx={{p: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
          <Stack direction="row" alignItems="center" spacing={2}>
            <IconButton
              aria-label={t('common:actions.close')}
              onClick={handleClose}
              sx={{
                bgcolor: 'background.paper',
                '&:hover': {bgcolor: 'action.hover'},
                boxShadow: 1,
              }}
            >
              <X size={24} />
            </IconButton>
            <AppBreadcrumbs
              items={getBreadcrumbSteps().map((step, index, array) => ({
                key: step,
                label: steps[step]?.label ?? step,
                onClick: index < array.length - 1 ? () => setCurrentStep(step) : undefined,
              }))}
            />
          </Stack>
        </Box>

        {/* Main content */}
        <Box sx={{flex: 1, display: 'flex', minHeight: 0}}>
          <Box
            sx={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              py: 4,
              px: {xs: 4, md: 10},
              alignItems: 'flex-start',
            }}
          >
            <Box
              sx={{
                width: '100%',
                maxWidth: isPropertiesStep ? 1200 : 800,
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              {/* Error Alerts */}
              {error && (
                <Alert severity="error" sx={{my: 3}} onClose={() => setError(null)}>
                  {error}
                </Alert>
              )}

              {createUserTypeMutation.error && (
                <Alert severity="error" sx={{mb: 3}}>
                  <Typography variant="body2" sx={{fontWeight: 'bold', mb: 0.5}}>
                    {createUserTypeMutation.error.message}
                  </Typography>
                </Alert>
              )}

              {renderStepContent()}

              {/* Navigation buttons */}
              <Stack direction="row" justifyContent="flex-end" alignItems="center" spacing={2} sx={{mt: 4}}>
                {activeSteps.indexOf(currentStep) > 0 && (
                  <Button variant="text" onClick={handlePrevStep} disabled={createUserTypeMutation.isPending}>
                    {t('common:actions.back')}
                  </Button>
                )}

                <Button
                  variant="contained"
                  disabled={!stepReady[currentStep] || createUserTypeMutation.isPending}
                  sx={{minWidth: 140}}
                  onClick={handleNextStep}
                >
                  {(() => {
                    if (!isLastStep) return t('common:actions.continue');
                    if (createUserTypeMutation.isPending) return t('common:status.saving');
                    return t('userTypes:createUserType');
                  })()}
                </Button>
              </Stack>
            </Box>
          </Box>
        </Box>
      </Box>

      {/* Validation Error Snackbar */}
      <Snackbar
        open={snackbarOpen}
        autoHideDuration={6000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{vertical: 'top', horizontal: 'right'}}
      >
        <Alert onClose={handleCloseSnackbar} severity="error" sx={{width: '100%'}}>
          {validationError}
        </Alert>
      </Snackbar>
    </Box>
  );
}
