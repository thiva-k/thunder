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

import {Typography, Stack, CircularProgress, Alert, useTheme, Box} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useEffect, useMemo, useState, useCallback} from 'react';
import {useTranslation, Trans} from 'react-i18next';
import {Lightbulb} from '@wso2/oxygen-ui-icons-react';
import {type IdentityProvider, IdentityProviderTypes} from '@/features/integrations/models/identity-provider';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import useIdentityProviders from '../../../../integrations/api/useIdentityProviders';
import useGetFlows from '../../../../flows/api/useGetFlows';
import {FlowType} from '../../../../flows/models/flows';
import {type BasicFlowDefinition} from '../../../../flows/models/responses';
import useApplicationCreateContext from '../../../hooks/useApplicationCreateContext';
import findMatchingFlowForIntegrations from '../../../../flows/utils/findMatchingFlowForIntegrations';
import getFlowSupportedIntegrations from '../../../../flows/utils/getFlowSupportedIntegrations';
import FlowsListView from './FlowsListView';
import IndividualMethodsToggleView from './IndividualMethodsToggleView';

/**
 * Props for the {@link ConfigureSignInOptions} component.
 *
 * @public
 */
export interface ConfigureSignInOptionsProps {
  /**
   * Record of enabled authentication integrations
   * Keys are integration IDs, values indicate whether they are enabled
   */
  integrations: Record<string, boolean>;

  /**
   * Callback function when an integration toggle state changes
   */
  onIntegrationToggle: (connectionId: string) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed
   */
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Type definition for grouped flows by authentication type
 *
 * @internal
 */
interface FlowsByType {
  basic: BasicFlowDefinition | null;
  google: BasicFlowDefinition | null;
  github: BasicFlowDefinition | null;
  smsOtp: BasicFlowDefinition | null;
  other: BasicFlowDefinition[];
}

/**
 * Check if at least one authentication option is selected OR a flow is selected
 *
 * @param integrations - Record of integration states
 * @param selectedFlow - Selected authentication flow
 * @returns True if at least one integration is enabled or a flow is selected
 *
 * @internal
 */
const hasAtLeastOneSelected = (
  integrations: Record<string, boolean>,
  selectedFlow: BasicFlowDefinition | null,
): boolean => Object.values(integrations).some((isEnabled) => isEnabled) || selectedFlow !== null;

/**
 * React component that renders the sign-in options configuration step in the
 * application creation onboarding flow.
 *
 * This component allows users to configure authentication methods for their application
 * by choosing between:
 * 1. Individual authentication integrations (Username & Password, Google, GitHub, etc.)
 * 2. Pre-configured authentication flows that may combine multiple methods
 *
 * Users can either toggle individual integrations OR select a pre-configured flow,
 * but not both simultaneously. When a flow is selected, individual integrations are
 * disabled and vice versa.
 *
 * The component fetches available identity providers and displays them as toggleable
 * list items with appropriate icons. Users can enable/disable multiple authentication
 * methods. The step is marked as ready only when at least one authentication option
 * is selected, ensuring applications have a valid sign-in mechanism.
 *
 * @param props - The component props
 * @param props.integrations - Record of enabled integrations (key: integration ID, value: enabled state)
 * @param props.onIntegrationToggle - Callback invoked when an integration is toggled
 * @param props.onReadyChange - Optional callback to notify parent of step readiness
 *
 * @returns JSX element displaying the sign-in options configuration interface
 *
 * @example
 * ```tsx
 * import ConfigureSignInOptions from './ConfigureSignInOptions';
 *
 * function OnboardingFlow() {
 *   const [integrations, setIntegrations] = useState({
 *     'username-password': true,
 *     'google-idp-id': false,
 *   });
 *
 *   const handleToggle = (id: string) => {
 *     setIntegrations(prev => ({
 *       ...prev,
 *       [id]: !prev[id]
 *     }));
 *   };
 *
 *   return (
 *     <ConfigureSignInOptions
 *       integrations={integrations}
 *       onIntegrationToggle={handleToggle}
 *       onReadyChange={(isReady) => console.log('Ready:', isReady)}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ConfigureSignInOptions({
  integrations,
  onIntegrationToggle,
  onReadyChange = undefined,
}: ConfigureSignInOptionsProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const {selectedAuthFlow, setSelectedAuthFlow, hasCompletedOnboarding} = useApplicationCreateContext();

  // State to track which view to show: 'flows' or 'toggles'
  const [currentView, setCurrentView] = useState<'flows' | 'toggles'>(() =>
    hasCompletedOnboarding ? 'flows' : 'toggles',
  );

  // Update view when onboarding status changes - only on initial load
  useEffect(() => {
    // Only set initial view, don't override manual changes
    setCurrentView((prev) => {
      // If this is the first time hasCompletedOnboarding is set to true,
      // and we're still showing toggles, switch to flows
      if (hasCompletedOnboarding && prev === 'toggles') {
        return 'flows';
      }
      // If onboarding status changes to false, always show toggles
      if (!hasCompletedOnboarding) {
        return 'toggles';
      }
      // Otherwise, keep the current view
      return prev;
    });
  }, [hasCompletedOnboarding]);

  const {data, isLoading, error} = useIdentityProviders();
  const {
    data: flowsData,
    isLoading: isFlowsLoading,
    error: flowsError,
  } = useGetFlows({
    flowType: FlowType.AUTHENTICATION,
  });

  const availableIntegrations: IdentityProvider[] = useMemo(() => data ?? [], [data]);
  const availableFlows: BasicFlowDefinition[] = useMemo(
    (): BasicFlowDefinition[] =>
      flowsData?.flows?.filter((flow: BasicFlowDefinition) => flow.handle !== 'develop-app-flow') ?? [],
    [flowsData?.flows],
  );

  /**
   * Map enabled integrations to flow-compatible types and find matching flow
   */
  const getFlowForEnabledIntegrations = useCallback(
    (integrationsState: Record<string, boolean>): BasicFlowDefinition | null => {
      const enabledIntegrations: string[] = Object.entries(integrationsState)
        .filter(([, enabled]) => enabled)
        .map(([integrationId]) => {
          // Handle basic auth
          if (integrationId === AuthenticatorTypes.BASIC_AUTH) {
            return AuthenticatorTypes.BASIC_AUTH;
          }

          // Find the provider to get its type
          const provider: IdentityProvider | undefined = availableIntegrations.find((idp) => idp.id === integrationId);
          if (provider) {
            switch (provider.type) {
              case IdentityProviderTypes.GOOGLE:
                return 'google';
              case IdentityProviderTypes.GITHUB:
                return 'github';
              default:
                return integrationId;
            }
          }

          // For other special flow types (like sms-otp)
          return integrationId;
        });

      return findMatchingFlowForIntegrations(enabledIntegrations, availableFlows);
    },
    [availableIntegrations, availableFlows],
  );

  /**
   * Broadcast readiness whenever integrations or selected flow change.
   */
  useEffect((): void => {
    const isReady: boolean = hasAtLeastOneSelected(integrations, selectedAuthFlow);
    if (onReadyChange) {
      onReadyChange(isReady);
    }
  }, [integrations, selectedAuthFlow, onReadyChange]);

  /**
   * Auto-select matching flow when integrations change
   */
  useEffect((): void => {
    if (!selectedAuthFlow && availableFlows.length > 0) {
      const matchingFlow: BasicFlowDefinition | null = getFlowForEnabledIntegrations(integrations);
      if (matchingFlow) {
        setSelectedAuthFlow(matchingFlow);
      }
    }
  }, [integrations, availableFlows, selectedAuthFlow, setSelectedAuthFlow, getFlowForEnabledIntegrations]);

  const handleIntegrationToggle = (integrationId: string): void => {
    // Toggle the integration first
    onIntegrationToggle(integrationId);

    // Create the new integrations state
    const newIntegrations: Record<string, boolean> = {
      ...integrations,
      [integrationId]: !integrations[integrationId],
    };

    // Find matching flow for the new integration state
    const matchingFlow: BasicFlowDefinition | null = getFlowForEnabledIntegrations(newIntegrations);
    setSelectedAuthFlow(matchingFlow);
  };

  if (isLoading || isFlowsLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', py: 8}}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || flowsError) {
    return (
      <Alert severity="error" sx={{mb: 4}}>
        {t('applications:onboarding.configure.SignInOptions.error', {
          error: error?.message ?? flowsError?.message ?? 'Unknown error',
        })}
      </Alert>
    );
  }

  const hasAtLeastOneSelectedOption: boolean = hasAtLeastOneSelected(integrations, selectedAuthFlow);

  const flowsByType: FlowsByType = availableFlows.reduce(
    (acc: FlowsByType, flow: BasicFlowDefinition) => {
      if (!flow.handle) {
        acc.other.push(flow);
        return acc;
      }

      const supportedIntegrations = getFlowSupportedIntegrations(flow.handle);

      // Prioritize flows based on their primary integration type
      if (
        supportedIntegrations.includes(AuthenticatorTypes.BASIC_AUTH) &&
        (flow.handle.includes('basic') || flow.handle === 'login-flow')
      ) {
        // Prefer single basic auth flows over combined flows
        if (!acc.basic || supportedIntegrations.length === 1) {
          acc.basic = flow;
        }
      } else if (supportedIntegrations.includes('google') && flow.handle.includes('google')) {
        // Prefer single Google flows
        if (!acc.google || supportedIntegrations.length === 1) {
          acc.google = flow;
        }
      } else if (supportedIntegrations.includes('github') && flow.handle.includes('github')) {
        // Prefer single GitHub flows
        if (!acc.github || supportedIntegrations.length === 1) {
          acc.github = flow;
        }
      } else if (supportedIntegrations.includes('sms-otp')) {
        acc.smsOtp ??= flow;
      } else {
        acc.other.push(flow);
      }

      return acc;
    },
    {basic: null, google: null, github: null, smsOtp: null, other: [] as BasicFlowDefinition[]},
  );

  // Event handlers
  const handleFlowSelect = (flowId: string): void => {
    const selectedFlow: BasicFlowDefinition | null =
      availableFlows?.find((flow: BasicFlowDefinition) => flow.id === flowId) ?? null;
    setSelectedAuthFlow(selectedFlow);
  };

  const handleClearFlowSelection = (): void => {
    setSelectedAuthFlow(null);
  };

  const handleSwitchToFlows = (): void => {
    setCurrentView('flows');
  };

  const handleSwitchToToggles = (): void => {
    setCurrentView('toggles');
  };

  return (
    <Stack direction="column" spacing={4}>
      <Stack direction="column" spacing={1}>
        <Typography variant="h1" gutterBottom>
          {t('applications:onboarding.configure.SignInOptions.title')}
        </Typography>
        <Typography variant="subtitle1" gutterBottom>
          {t('applications:onboarding.configure.SignInOptions.subtitle')}
        </Typography>
      </Stack>

      {/* Validation warning if no options selected */}
      {!hasAtLeastOneSelectedOption && (
        <Alert severity="warning" sx={{mb: 2}}>
          {t('applications:onboarding.configure.SignInOptions.noSelectionWarning')}
        </Alert>
      )}

      {/* Conditional rendering based on view and onboarding status */}
      {(() => {
        if (currentView === 'flows') {
          // Show flows view
          return (
            <FlowsListView
              availableFlows={availableFlows}
              selectedAuthFlow={selectedAuthFlow}
              hasCompletedOnboarding={hasCompletedOnboarding}
              onFlowSelect={handleFlowSelect}
              onSwitchToToggles={handleSwitchToToggles}
              onClearSelection={handleClearFlowSelection}
            />
          );
        }
        // Show toggles view
        return (
          <IndividualMethodsToggleView
            integrations={integrations}
            availableIntegrations={availableIntegrations}
            flowsByType={flowsByType}
            onIntegrationToggle={handleIntegrationToggle}
          />
        );
      })()}

      {/* Hint for view switching */}
      <Stack direction="row" alignItems="center" spacing={1} flexWrap="wrap">
        <Stack direction="row" alignItems="center" spacing={1}>
          <Lightbulb size={20} color={theme?.vars?.palette.warning.main} />
          <Typography variant="body2" color="text.secondary">
            <Trans
              i18nKey={
                currentView === 'toggles'
                  ? 'applications:onboarding.configure.SignInOptions.individualMethods.hint'
                  : 'applications:onboarding.configure.SignInOptions.preConfiguredFlows.hint'
              }
              defaults={
                currentView === 'toggles'
                  ? 'You can also pick from an <switchLink>already defined flow</switchLink>. These can be always changed later in the application settings.'
                  : 'You can also configure <switchLink>individual authentication methods</switchLink>. These can be always changed later in the application settings.'
              }
              components={{
                switchLink: (
                  <Box
                    component="span"
                    role="button"
                    tabIndex={0}
                    onClick={currentView === 'toggles' ? handleSwitchToFlows : handleSwitchToToggles}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        (currentView === 'toggles' ? handleSwitchToFlows : handleSwitchToToggles)();
                      }
                    }}
                    sx={{
                      cursor: 'pointer',
                      color: 'primary.main',
                      textDecoration: 'underline',
                      '&:hover': {
                        textDecoration: 'none',
                      },
                    }}
                  />
                ),
              }}
            />
          </Typography>
        </Stack>
      </Stack>
    </Stack>
  );
}
