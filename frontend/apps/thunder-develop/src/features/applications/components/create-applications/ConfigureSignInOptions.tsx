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
  Box,
  Typography,
  Stack,
  Switch,
  CircularProgress,
  Alert,
  useTheme,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Divider,
} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useEffect} from 'react';
import {Info, Lightbulb, UserRound, Google, GitHub} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {type IdentityProvider, IdentityProviderTypes} from '@/features/integrations/models/identity-provider';
import getIntegrationIcon from '@/features/integrations/utils/getIntegrationIcon';
import useIdentityProviders from '../../../integrations/api/useIdentityProviders';
import {USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY} from '../../utils/resolveAuthFlowGraphId';

/**
 * Props for the ConfigureSignInOptions component.
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
 * Check if at least one authentication option is selected
 */
const hasAtLeastOneSelected = (integrations: Record<string, boolean>): boolean =>
  Object.values(integrations).some((isEnabled) => isEnabled);

export default function ConfigureSignInOptions({
  integrations,
  onIntegrationToggle,
  onReadyChange = undefined,
}: ConfigureSignInOptionsProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const {data, isLoading, error} = useIdentityProviders();

  // Broadcast readiness whenever integrations change
  useEffect(() => {
    const isReady = hasAtLeastOneSelected(integrations);
    if (onReadyChange) {
      onReadyChange(isReady);
    }
  }, [integrations, onReadyChange]);

  if (isLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', py: 8}}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error" sx={{mb: 4}}>
        {t('applications:onboarding.configure.SignInOptions.error', {error: error.message ?? 'Unknown error'})}
      </Alert>
    );
  }

  const availableIntegrations: IdentityProvider[] = data ?? [];
  
  // Check if Google and GitHub exist in the API data
  const googleProvider = availableIntegrations.find((idp) => idp.type === IdentityProviderTypes.GOOGLE);
  const githubProvider = availableIntegrations.find((idp) => idp.type === IdentityProviderTypes.GITHUB);
  
  const hasAtLeastOneSelectedOption = hasAtLeastOneSelected(integrations);
  const hasUsernamePassword = integrations[USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY] ?? false;

  return (
    <Stack direction="column" spacing={4}>
      <Stack direction="column" spacing={1}>
        <Typography variant="h1" gutterBottom>
          {t('applications:onboarding.configure.SignInOptions.title')}
        </Typography>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Info size={14} />
          <Typography variant="body1" color="text.secondary" sx={{mb: 4}}>
            {t('applications:onboarding.configure.SignInOptions.subtitle')}
          </Typography>
        </Stack>
      </Stack>

      {/* Validation warning if no options selected */}
      {!hasAtLeastOneSelectedOption && (
        <Alert severity="warning" sx={{mb: 2}}>
          {t('applications:onboarding.configure.SignInOptions.noSelectionWarning')}
        </Alert>
      )}

      <List sx={{bgcolor: 'background.paper', borderRadius: 1, border: 1, borderColor: 'divider'}}>
        {/* Username & Password Option - Always shown first, always toggleable */}
        <ListItem
          disablePadding
          secondaryAction={
            <Switch
              edge="end"
              checked={hasUsernamePassword}
              onChange={(): void => onIntegrationToggle(USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY)}
              color="primary"
            />
          }
        >
          <ListItemButton onClick={(): void => onIntegrationToggle(USERNAME_PASSWORD_AUTHENTICATION_OPTION_KEY)}>
            <ListItemIcon>
              <UserRound size={24} />
            </ListItemIcon>
            <ListItemText primary={t('applications:onboarding.configure.SignInOptions.usernamePassword')} />
          </ListItemButton>
        </ListItem>
        
        <Divider component="li" />

        {/* Google Option - Always shown, enabled if configured */}
        {googleProvider ? (
          <ListItem
            disablePadding
            secondaryAction={
              <Switch
                edge="end"
                checked={integrations[googleProvider.id] ?? false}
                onChange={(): void => onIntegrationToggle(googleProvider.id)}
                color="primary"
              />
            }
          >
            <ListItemButton onClick={(): void => onIntegrationToggle(googleProvider.id)}>
              <ListItemIcon>
                <Google size={24} />
              </ListItemIcon>
              <ListItemText primary={t('applications:onboarding.configure.SignInOptions.google')} />
            </ListItemButton>
          </ListItem>
        ) : (
          <ListItem disablePadding>
            <ListItemButton disabled>
              <ListItemIcon>
                <Google size={24} />
              </ListItemIcon>
              <ListItemText 
                primary={t('applications:onboarding.configure.SignInOptions.google')} 
                secondary={t('applications:onboarding.configure.SignInOptions.notConfigured')}
              />
            </ListItemButton>
          </ListItem>
        )}
        <Divider component="li" />

        {/* GitHub Option - Always shown, enabled if configured */}
        {githubProvider ? (
          <ListItem
            disablePadding
            secondaryAction={
              <Switch
                edge="end"
                checked={integrations[githubProvider.id] ?? false}
                onChange={(): void => onIntegrationToggle(githubProvider.id)}
                color="primary"
              />
            }
          >
            <ListItemButton onClick={(): void => onIntegrationToggle(githubProvider.id)}>
              <ListItemIcon>
                <GitHub size={24} />
              </ListItemIcon>
              <ListItemText primary={t('applications:onboarding.configure.SignInOptions.github')} />
            </ListItemButton>
          </ListItem>
        ) : (
          <ListItem disablePadding>
            <ListItemButton disabled>
              <ListItemIcon>
                <GitHub size={24} />
              </ListItemIcon>
              <ListItemText 
                primary={t('applications:onboarding.configure.SignInOptions.github')} 
                secondary={t('applications:onboarding.configure.SignInOptions.notConfigured')}
              />
            </ListItemButton>
          </ListItem>
        )}

        {/* Other Social Login Providers (if any) */}
        {availableIntegrations
          .filter((provider) => provider.type !== IdentityProviderTypes.GOOGLE && provider.type !== IdentityProviderTypes.GITHUB)
          .map(
            (provider: IdentityProvider, index: number, filteredProviders): JSX.Element => (
              <>
                <ListItem
                  key={provider.id}
                  disablePadding
                  secondaryAction={
                    <Switch
                      edge="end"
                      checked={integrations[provider.id] ?? false}
                      onChange={(): void => onIntegrationToggle(provider.id)}
                      color="primary"
                    />
                  }
                >
                  <ListItemButton onClick={(): void => onIntegrationToggle(provider.id)}>
                    <ListItemIcon>{getIntegrationIcon(provider.type)}</ListItemIcon>
                    <ListItemText primary={provider.name} />
                  </ListItemButton>
                </ListItem>
                {index < filteredProviders.length - 1 && <Divider component="li" />}
              </>
            ),
          )}
      </List>

      <Stack direction="row" alignItems="center" spacing={1}>
        <Lightbulb size={20} color={theme?.vars?.palette.warning.main} />
        <Typography variant="body2" color="text.secondary">
          {t('applications:onboarding.configure.SignInOptions.hint')}
        </Typography>
      </Stack>
    </Stack>
  );
}
