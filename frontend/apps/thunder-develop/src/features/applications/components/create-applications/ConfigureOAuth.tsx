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
  TextField,
  FormControl,
  FormLabel,
  Button,
  Chip,
  Alert,
  Switch,
  FormControlLabel,
  Select,
  MenuItem,
  FormHelperText,
} from '@wso2/oxygen-ui';
import {Plus, X} from '@wso2/oxygen-ui-icons-react';
import type {ChangeEvent, JSX} from 'react';
import {useState, useEffect, useMemo, useRef} from 'react';
import {useTranslation} from 'react-i18next';
import {
  OAuth2GrantTypes,
  TokenEndpointAuthMethods,
  getDefaultOAuthConfig,
  type OAuth2Config,
  type OAuth2GrantType,
} from '../../models/oauth';

/**
 * Props for the ConfigureOAuth component.
 */
export interface ConfigureOAuthProps {
  /**
   * OAuth2 configuration object
   */
  oauthConfig: OAuth2Config | null;

  /**
   * Callback function when OAuth configuration changes
   */
  onOAuthConfigChange: (config: OAuth2Config | null) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed
   */
  onReadyChange?: (isReady: boolean) => void;

  /**
   * Callback function when validation errors change
   */
  onValidationErrorsChange?: (hasErrors: boolean) => void;
}

/**
 * Validate if a string is a valid URL
 */
const isValidURL = (url: string): boolean => {
  try {
    const urlObj = new URL(url);
    return urlObj.protocol === 'http:' || urlObj.protocol === 'https:';
  } catch {
    return false;
  }
};


export default function ConfigureOAuth({
  oauthConfig,
  onOAuthConfigChange,
  onReadyChange = undefined,
  onValidationErrorsChange = undefined,
}: ConfigureOAuthProps): JSX.Element {
  const {t} = useTranslation();
  const [currentURI, setCurrentURI] = useState('');
  const [uriError, setUriError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

  // Initialize config if null - memoize to prevent creating new object on every render
  // This prevents infinite loops in useEffect hooks that depend on config
  const config = useMemo(() => oauthConfig ?? getDefaultOAuthConfig(), [oauthConfig]);

  // Track previous config to avoid unnecessary re-validations
  const prevConfigRef = useRef<string | null>(null);

  // Validate configuration
  useEffect(() => {
    const configString = JSON.stringify(config);
    
    // Skip validation if config hasn't changed since last validation
    // Always validate on first render (when prevConfigRef.current is null)
    if (prevConfigRef.current !== null && configString === prevConfigRef.current) {
      return;
    }
    
    prevConfigRef.current = configString;
    
    const errors: Record<string, string> = {};

    // If public client is enabled
    if (config.public_client) {
      // PKCE must be enabled for public clients
      if (!config.pkce_required) {
        errors.pkce = t('applications:onboarding.configure.oauth.errors.publicClientRequiresPKCE');
      }

      // Token endpoint auth method must be 'none' for public clients
      if (config.token_endpoint_auth_method !== TokenEndpointAuthMethods.NONE) {
        errors.tokenEndpointAuthMethod = t(
          'applications:onboarding.configure.oauth.errors.publicClientRequiresNone',
        );
      }

      // Public clients cannot use client_credentials grant type
      if (config.grant_types.includes(OAuth2GrantTypes.CLIENT_CREDENTIALS)) {
        errors.grantTypes = t('applications:onboarding.configure.oauth.errors.publicClientNoClientCredentials');
      }
    }

    // At least one grant type must be selected
    // Check after public client validation so we don't show conflicting error messages
    // (e.g., if public client removed client_credentials, we check if any grant types remain)
    if (config.grant_types.length === 0) {
      errors.grantTypes = t('applications:onboarding.configure.oauth.errors.atLeastOneGrantTypeRequired');
    }

    // Refresh token grant type cannot be selected alone - it requires authorization_code
    if (
      config.grant_types.length === 1 &&
      config.grant_types.includes(OAuth2GrantTypes.REFRESH_TOKEN)
    ) {
      errors.grantTypes = t('applications:onboarding.configure.oauth.errors.refreshTokenRequiresAuthorizationCode');
    }

    // authorization_code grant type requires redirect URIs
    if (
      config.grant_types.includes(OAuth2GrantTypes.AUTHORIZATION_CODE) &&
      config.redirect_uris.length === 0
    ) {
      errors.redirectURIs = t('applications:onboarding.configure.oauth.errors.authorizationCodeRequiresRedirectURIs');
    }

    // client_credentials grant type cannot use 'none' authentication method
    if (
      config.grant_types.includes(OAuth2GrantTypes.CLIENT_CREDENTIALS) &&
      config.token_endpoint_auth_method === TokenEndpointAuthMethods.NONE
    ) {
      errors.tokenEndpointAuthMethod = t(
        'applications:onboarding.configure.oauth.errors.clientCredentialsRequiresAuth',
      );
    }

    // Update validation errors (validation is deterministic, so same config = same errors)
    setValidationErrors(errors);
    
    // Broadcast validation errors
    if (onValidationErrorsChange) {
      const hasErrors = Object.keys(errors).length > 0;
      onValidationErrorsChange(hasErrors);
    }
  }, [config, t, onValidationErrorsChange]);

  // Track previous readiness to avoid unnecessary callbacks
  const prevIsReadyRef = useRef<boolean | undefined>(undefined);

  // Broadcast readiness whenever config changes
  // OAuth config step is optional, but if user starts configuring it,
  // at least one grant type must be selected for the step to be considered ready
  useEffect(() => {
    if (onReadyChange) {
      // Step is ready if at least one grant type is selected
      // (empty grant_types means step is not ready, but that's OK since it's optional)
      const isReady = config.grant_types.length > 0;
      
      // Only call callback if readiness actually changed
      if (prevIsReadyRef.current !== isReady) {
        prevIsReadyRef.current = isReady;
        onReadyChange(isReady);
      }
    }
  }, [config, onReadyChange]);

  const handleAddURI = (): void => {
    const trimmedURI = currentURI.trim();

    if (!trimmedURI) {
      setUriError(t('applications:onboarding.configure.oauth.redirectURIs.errors.empty'));
      return;
    }

    if (!isValidURL(trimmedURI)) {
      setUriError(t('applications:onboarding.configure.oauth.redirectURIs.errors.invalid'));
      return;
    }

    if (config.redirect_uris.includes(trimmedURI)) {
      setUriError(t('applications:onboarding.configure.oauth.redirectURIs.errors.duplicate'));
      return;
    }

    setUriError(null);
    onOAuthConfigChange({
      ...config,
      redirect_uris: [...config.redirect_uris, trimmedURI],
    });
    setCurrentURI('');
  };

  const handleRemoveURI = (uriToRemove: string): void => {
    onOAuthConfigChange({
      ...config,
      redirect_uris: config.redirect_uris.filter((uri) => uri !== uriToRemove),
    });
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>): void => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAddURI();
    }
  };

  const handleURIChange = (e: ChangeEvent<HTMLInputElement>): void => {
    setCurrentURI(e.target.value);
    if (uriError) {
      setUriError(null);
    }
  };

  const handlePublicClientChange = (event: ChangeEvent<HTMLInputElement>): void => {
    const isPublicClient = event.target.checked;
    let updatedGrantTypes = config.grant_types.filter((gt) => gt !== OAuth2GrantTypes.CLIENT_CREDENTIALS);
    
    // If enabling public client and no grant types remain, add authorization_code
    if (isPublicClient && updatedGrantTypes.length === 0) {
      updatedGrantTypes = [OAuth2GrantTypes.AUTHORIZATION_CODE];
    }
    
    const updatedConfig: OAuth2Config = {
      ...config,
      public_client: isPublicClient,
      // Auto-enable PKCE for public clients
      pkce_required: isPublicClient ? true : config.pkce_required,
      // Auto-set token endpoint auth method to 'none' for public clients
      token_endpoint_auth_method: isPublicClient
        ? TokenEndpointAuthMethods.NONE
        : TokenEndpointAuthMethods.CLIENT_SECRET_BASIC,
      grant_types: updatedGrantTypes,
    };
    onOAuthConfigChange(updatedConfig);
  };

  const handlePKCEChange = (event: ChangeEvent<HTMLInputElement>): void => {
    // If public client, PKCE cannot be disabled
    if (config.public_client && !event.target.checked) {
      return;
    }
    onOAuthConfigChange({
      ...config,
      pkce_required: event.target.checked,
    });
  };

  const handleGrantTypeToggle = (grantType: OAuth2GrantType): void => {
    const isSelected = config.grant_types.includes(grantType);
    let updatedGrantTypes: OAuth2GrantType[];

    if (isSelected) {
      updatedGrantTypes = config.grant_types.filter((gt) => gt !== grantType) as OAuth2GrantType[];
      
      // If removing authorization_code and refresh_token is selected, also remove refresh_token
      if (grantType === OAuth2GrantTypes.AUTHORIZATION_CODE && updatedGrantTypes.includes(OAuth2GrantTypes.REFRESH_TOKEN)) {
        updatedGrantTypes = updatedGrantTypes.filter((gt) => gt !== OAuth2GrantTypes.REFRESH_TOKEN);
      }
    } else {
      updatedGrantTypes = [...(config.grant_types as OAuth2GrantType[]), grantType];
      
      // If adding refresh_token, ensure authorization_code is also included
      if (grantType === OAuth2GrantTypes.REFRESH_TOKEN && !updatedGrantTypes.includes(OAuth2GrantTypes.AUTHORIZATION_CODE)) {
        updatedGrantTypes.push(OAuth2GrantTypes.AUTHORIZATION_CODE);
      }
    }

    // Backend will automatically handle response_types based on grant types
    onOAuthConfigChange({
      ...config,
      grant_types: updatedGrantTypes,
    });
  };

  const handleTokenEndpointAuthMethodChange = (event: {target: {value: string}}): void => {
    const method = event.target.value;
    // If public client, only 'none' is allowed
    if (config.public_client && method !== TokenEndpointAuthMethods.NONE) {
      return;
    }
    onOAuthConfigChange({
      ...config,
      token_endpoint_auth_method: method,
    });
  };

  const availableGrantTypes: {value: OAuth2GrantType; label: string}[] = [
    {value: OAuth2GrantTypes.AUTHORIZATION_CODE, label: t('applications:onboarding.configure.oauth.grantTypes.authorizationCode')},
    {value: OAuth2GrantTypes.REFRESH_TOKEN, label: t('applications:onboarding.configure.oauth.grantTypes.refreshToken')},
    {value: OAuth2GrantTypes.CLIENT_CREDENTIALS, label: t('applications:onboarding.configure.oauth.grantTypes.clientCredentials')},
  ];

  const availableAuthMethods = [
    {value: TokenEndpointAuthMethods.CLIENT_SECRET_BASIC, label: t('applications:onboarding.configure.oauth.tokenEndpointAuthMethod.clientSecretBasic')},
    {value: TokenEndpointAuthMethods.CLIENT_SECRET_POST, label: t('applications:onboarding.configure.oauth.tokenEndpointAuthMethod.clientSecretPost')},
    {value: TokenEndpointAuthMethods.NONE, label: t('applications:onboarding.configure.oauth.tokenEndpointAuthMethod.none')},
  ];

  return (
    <Stack direction="column" spacing={3}>
      <Stack direction="column" spacing={1}>
        <Typography variant="h1" gutterBottom>
          {t('applications:onboarding.configure.oauth.title')}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {t('applications:onboarding.configure.oauth.subtitle')}
        </Typography>
        <Alert severity="info" sx={{mt: 1}}>
          {t('applications:onboarding.configure.oauth.optional')}
        </Alert>
      </Stack>

      {/* Public Client and PKCE */}
      <Stack direction="column" spacing={1}>
        <FormControl fullWidth>
          <FormControlLabel
            control={
              <Switch
                checked={config.public_client ?? false}
                onChange={handlePublicClientChange}
                name="public_client"
              />
            }
            label={t('applications:onboarding.configure.oauth.publicClient.label')}
          />
        </FormControl>
        <FormControl fullWidth error={!!validationErrors.pkce}>
          <FormControlLabel
            control={
              <Switch
                checked={config.pkce_required ?? false}
                onChange={handlePKCEChange}
                disabled={config.public_client ?? false}
                name="pkce_required"
              />
            }
            label={t('applications:onboarding.configure.oauth.pkce.label')}
          />
          {validationErrors.pkce && (
            <FormHelperText error>{validationErrors.pkce}</FormHelperText>
          )}
        </FormControl>
      </Stack>

      {/* Grant Types */}
      <FormControl fullWidth error={!!validationErrors.grantTypes}>
        <FormLabel>{t('applications:onboarding.configure.oauth.grantTypes.label')}</FormLabel>
        <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1.5, mt: 1.5}}>
          {availableGrantTypes.map((grantType) => {
            const isSelected = config.grant_types.includes(grantType.value);
            const isDisabled =
              config.public_client && grantType.value === OAuth2GrantTypes.CLIENT_CREDENTIALS;
            return (
              <Chip
                key={grantType.value}
                label={grantType.label}
                onClick={() => !isDisabled && handleGrantTypeToggle(grantType.value)}
                color={isSelected ? 'primary' : 'default'}
                variant={isSelected ? 'filled' : 'outlined'}
                disabled={isDisabled}
                aria-label={
                  isDisabled
                    ? `${grantType.label} (not available for public clients)`
                    : undefined
                }
                sx={{cursor: isDisabled ? 'not-allowed' : 'pointer'}}
              />
            );
          })}
        </Box>
        {validationErrors.grantTypes && (
          <FormHelperText error>{validationErrors.grantTypes}</FormHelperText>
        )}
      </FormControl>

      {/* Redirect URIs */}
      <FormControl fullWidth error={!!validationErrors.redirectURIs}>
        <FormLabel htmlFor="redirect-uri-input">
          {t('applications:onboarding.configure.oauth.redirectURIs.fieldLabel')}
        </FormLabel>
        <Stack direction="row" spacing={2}>
          <TextField
            fullWidth
            id="redirect-uri-input"
            value={currentURI}
            onChange={handleURIChange}
            onKeyPress={handleKeyPress}
            placeholder={t('applications:onboarding.configure.oauth.redirectURIs.placeholder')}
            error={!!uriError || !!validationErrors.redirectURIs}
            helperText={uriError ?? validationErrors.redirectURIs}
          />
          <Button
            variant="outlined"
            onClick={handleAddURI}
            startIcon={<Plus size={20} />}
            sx={{minWidth: 120, height: 36}}
          >
            {t('applications:onboarding.configure.oauth.redirectURIs.addButton')}
          </Button>
        </Stack>
        {config.redirect_uris.length > 0 && (
          <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1, mt: 1.5}}>
            {config.redirect_uris.map(
              (uri: string): JSX.Element => (
                <Chip
                  key={uri}
                  label={uri}
                  onDelete={(): void => handleRemoveURI(uri)}
                  deleteIcon={<X size={16} />}
                  variant="outlined"
                  sx={{
                    maxWidth: '100%',
                    '& .MuiChip-label': {
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    },
                    '& .MuiChip-deleteIcon': {
                      color: 'text.secondary',
                      '&:hover': {
                        color: 'error.main',
                      },
                    },
                  }}
                />
              ),
            )}
          </Box>
        )}
      </FormControl>

      {/* Token Endpoint Auth Method */}
      <FormControl fullWidth error={!!validationErrors.tokenEndpointAuthMethod}>
        <FormLabel htmlFor="token-endpoint-auth-method">
          {t('applications:onboarding.configure.oauth.tokenEndpointAuthMethod.label')}
        </FormLabel>
        <Select
          id="token-endpoint-auth-method"
          value={config.token_endpoint_auth_method ?? TokenEndpointAuthMethods.CLIENT_SECRET_BASIC}
          onChange={handleTokenEndpointAuthMethodChange}
          disabled={config.public_client ?? false}
        >
          {availableAuthMethods.map((method) => (
            <MenuItem key={method.value} value={method.value}>
              {method.label}
            </MenuItem>
          ))}
        </Select>
        {validationErrors.tokenEndpointAuthMethod && (
          <FormHelperText error>{validationErrors.tokenEndpointAuthMethod}</FormHelperText>
        )}
      </FormControl>

    </Stack>
  );
}
