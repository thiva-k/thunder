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

/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-explicit-any */

import type {JSX} from 'react';
import {
  Box,
  Button,
  FormLabel,
  FormControl,
  Alert,
  TextField,
  Typography,
  styled,
  AlertTitle,
  Paper,
  Stack,
  ColorSchemeImage,
  IconButton,
  InputAdornment,
  Select,
  MenuItem,
  CircularProgress,
  Avatar,
  useTheme,
} from '@wso2/oxygen-ui';
import {EmbeddedFlowComponentType, EmbeddedFlowEventType, AcceptInvite, type EmbeddedFlowComponent} from '@asgardeo/react';
import {Eye, EyeClosed} from '@wso2/oxygen-ui-icons-react';
import {useNavigate} from 'react-router';
import {useState} from 'react';
import {useTranslation} from 'react-i18next';
import {mapEmbeddedFlowTextVariant, useBranding} from '@thunder/shared-branding';
import {useConfig} from '@thunder/commons-contexts';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';
import ROUTES from '../../constants/routes';

/** Typed shape for flow sub-components */
type FlowSubComponent = EmbeddedFlowComponent & {
  placeholder?: string;
  required?: boolean;
  options?: any[];
  hint?: string;
  variant?: string;
  eventType?: string;
};

const StyledPaper = styled(Paper)(({theme}) => ({
  display: 'flex',
  flexDirection: 'column',
  alignSelf: 'center',
  width: '100%',
  padding: theme.spacing(4),
  gap: theme.spacing(2),
  [theme.breakpoints.up('sm')]: {
    width: '450px',
  },
}));

export default function AcceptInviteBox(): JSX.Element {
  const navigate = useNavigate();
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();
  const {getServerUrl} = useConfig();
  const {images, theme: brandingTheme, isBrandingEnabled} = useBranding();
  const theme = useTheme();

  const [showPasswordMap, setShowPasswordMap] = useState<Record<string, boolean>>({});
  const baseUrl = getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string);

  const togglePasswordVisibility = (field: string): void => {
    setShowPasswordMap((prev) => ({...prev, [field]: !prev[field]}));
  };

  const handleGoToSignIn = () => {
    const result = navigate(ROUTES.AUTH.SIGN_IN);
    if (result instanceof Promise) result.catch(() => {});
  };

  const getOptionValue = (option: any): string => {
    if (typeof option === 'string') return option;
    if (typeof option?.value === 'string') return option.value;
    return JSON.stringify(option?.value ?? option);
  };

  const getOptionLabel = (option: any): string => {
    if (typeof option === 'string') return option;
    if (typeof option?.label === 'string') return option.label;
    return JSON.stringify(option?.label ?? option);
  };

  const renderFormField = (
    component: FlowSubComponent,
    index: number,
    values: Record<string, string>,
    touched: Record<string, boolean>,
    fieldErrors: Record<string, string>,
    isLoading: boolean,
    handleInputChange: (field: string, value: string) => void,
  ) => {
    const {type, ref, label, placeholder, required, options, hint} = component;
    if (!ref) return null;

    const hasError = touched?.[ref] && fieldErrors?.[ref];
    const value = values?.[ref] ?? '';

    // TEXT_INPUT
    if (type === EmbeddedFlowComponentType.TextInput) {
      return (
        <FormControl key={component.id ?? index} required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(label)!)}</FormLabel>
          <TextField
            fullWidth
            id={ref}
            name={ref}
            type="text"
            placeholder={t(resolve(placeholder) ?? placeholder ?? '')}
            autoComplete="off"
            required={required}
            variant="outlined"
            disabled={isLoading}
            error={!!hasError}
            helperText={hasError ? fieldErrors[ref] : undefined}
            color={hasError ? 'error' : 'primary'}
            value={value}
            onChange={(e) => handleInputChange(ref, e.target.value)}
          />
        </FormControl>
      );
    }

    // EMAIL_INPUT
    if (type === 'EMAIL_INPUT') {
      return (
        <FormControl key={component.id ?? index} required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(label)!)}</FormLabel>
          <TextField
            fullWidth
            id={ref}
            name={ref}
            type="email"
            placeholder={t(resolve(placeholder) ?? placeholder ?? '')}
            autoComplete="email"
            required={required}
            variant="outlined"
            disabled={isLoading}
            error={!!hasError}
            helperText={hasError ? fieldErrors[ref] : undefined}
            color={hasError ? 'error' : 'primary'}
            value={value}
            onChange={(e) => handleInputChange(ref, e.target.value)}
          />
        </FormControl>
      );
    }

    // PASSWORD_INPUT
    if (type === EmbeddedFlowComponentType.PasswordInput) {
      const showPassword = showPasswordMap[ref] ?? false;
      return (
        <FormControl key={component.id ?? index} required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(label)!)}</FormLabel>
          <TextField
            fullWidth
            id={ref}
            name={ref}
            type={showPassword ? 'text' : 'password'}
            placeholder={t(resolve(placeholder) ?? placeholder ?? '')}
            autoComplete="new-password"
            required={required}
            variant="outlined"
            disabled={isLoading}
            error={!!hasError}
            helperText={hasError ? fieldErrors[ref] : undefined}
            color={hasError ? 'error' : 'primary'}
            value={value}
            onChange={(e) => handleInputChange(ref, e.target.value)}
            slotProps={{
              input: {
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={() => togglePasswordVisibility(ref)}
                      edge="end"
                      disabled={isLoading}
                    >
                      {showPassword ? <EyeClosed /> : <Eye />}
                    </IconButton>
                  </InputAdornment>
                ),
              },
            }}
          />
        </FormControl>
      );
    }

    // SELECT
    if (type === 'SELECT' && options) {
      return (
        <FormControl key={component.id ?? index} fullWidth>
          <FormLabel htmlFor={ref}>{t(resolve(label)!)}</FormLabel>
          <Select
            displayEmpty
            size="small"
            id={ref}
            name={ref}
            required={required}
            fullWidth
            disabled={isLoading}
            error={!!hasError}
            value={value}
            onChange={(e) => handleInputChange(ref, e.target.value)}
          >
            <MenuItem value="" disabled>
              {t(resolve(placeholder) ?? 'Select an option')}
            </MenuItem>
            {options.map((option: any) => (
              <MenuItem key={getOptionValue(option)} value={getOptionValue(option)}>
                {getOptionLabel(option)}
              </MenuItem>
            ))}
          </Select>
          {hasError && (
            <Typography variant="caption" color="error.main" sx={{mt: 0.5}}>
              {fieldErrors[ref]}
            </Typography>
          )}
          {hint && (
            <Typography variant="caption" color="text.secondary">
              {hint}
            </Typography>
          )}
        </FormControl>
      );
    }

    return null;
  };

  return (
    <Stack gap={2}>
      <ColorSchemeImage
        src={{
          light: images?.logo?.primary?.url ?? `${import.meta.env.BASE_URL}/assets/images/logo.svg`,
          dark: images?.logo?.primary?.url ?? `${import.meta.env.BASE_URL}/assets/images/logo-inverted.svg`,
        }}
        alt={{
          light: images?.logo?.primary?.alt ?? 'Logo (Light)',
          dark: images?.logo?.primary?.alt ?? 'Logo (Dark)',
        }}
        height={images?.logo?.primary?.height ?? 30}
        width={images?.logo?.primary?.width ?? 'auto'}
        sx={{display: {xs: 'flex', md: 'none'}}}
      />
      {images?.logo?.primary?.url && (
        <Box sx={{display: {xs: 'none', md: 'flex'}, justifyContent: 'center', mb: 2}}>
          <Avatar
            src={images.logo.primary.url}
            alt={images.logo.primary.alt ?? 'Logo'}
            sx={{
              width: 64,
              height: 64,
              p: 1,
              ...theme.applyStyles('light', {
                backgroundColor: brandingTheme?.palette?.primary?.main ?? theme.palette.primary.main,
              }),
              ...theme.applyStyles('dark', {
                backgroundColor: brandingTheme?.palette?.primary?.main ?? theme.palette.primary.main,
              }),
            }}
          />
        </Box>
      )}
      <StyledPaper variant="outlined">
        <AcceptInvite
          baseUrl={baseUrl}
          onGoToSignIn={handleGoToSignIn}
          onError={(error: Error) => console.error('Invite acceptance error:', error)}
        >
          {({values, fieldErrors, error, touched, isLoading, components, handleInputChange, handleSubmit, isComplete, isValidatingToken, isTokenInvalid, isValid}) => {
            // Validating token
            if (isValidatingToken) {
              return (
                <Box sx={{display: 'flex', flexDirection: 'column', alignItems: 'center', p: 3, gap: 2}}>
                  <CircularProgress />
                  <Typography>{t('invite:validating', 'Validating your invite link...')}</Typography>
                </Box>
              );
            }

            // Invalid token
            if (isTokenInvalid) {
              return (
                <Alert severity="error">
                  <AlertTitle>{t('invite:errors.invalid.title', 'Unable to verify invite')}</AlertTitle>
                  {t('invite:errors.invalid.description', 'This invite link is invalid or has expired.')}
                </Alert>
              );
            }

            // Completed
            if (isComplete) {
              return (
                <Box sx={{textAlign: 'center', py: 2}}>
                  <Alert severity="success">
                    {t('invite:complete.description', 'Your account has been successfully set up.')}
                  </Alert>
                </Box>
              );
            }

            // Loading
            if (isLoading && !components?.length) {
              return (
                <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                  <CircularProgress />
                </Box>
              );
            }

            return (
              <>
                {error && (
                  <Alert severity="error" sx={{mb: 2}}>
                    <AlertTitle>{t('invite:errors.failed.title', 'Error')}</AlertTitle>
                    {error.message ?? t('invite:errors.failed.description', 'An error occurred.')}
                  </Alert>
                )}
                {components?.length > 0 && (
                  <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                    {components.map((component, index) => {
                      // TEXT
                      if (component.type === EmbeddedFlowComponentType.Text) {
                        return (
                          <Typography
                            key={component.id ?? index}
                            variant={mapEmbeddedFlowTextVariant(component.variant)}
                            sx={{mb: 1, textAlign: isBrandingEnabled ? 'center' : 'left'}}
                          >
                            {t(resolve(component.label)!)}
                          </Typography>
                        );
                      }

                      // BLOCK
                      if (component.type === EmbeddedFlowComponentType.Block) {
                        const blockComponents = (component.components ?? []) as FlowSubComponent[];
                        const submitAction = blockComponents.find(
                          (c) => c.type === EmbeddedFlowComponentType.Action && c.eventType === EmbeddedFlowEventType.Submit,
                        );

                        if (!submitAction) return null;

                        return (
                          <Box
                            key={component.id ?? index}
                            component="form"
                            onSubmit={(e) => {
                              e.preventDefault();
                              handleSubmit(submitAction, values).catch(() => {});
                            }}
                            noValidate
                            sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
                          >
                            {blockComponents.map((subComponent, compIndex) => {
                              // Form fields
                              const field = renderFormField(subComponent, compIndex, values, touched, fieldErrors, isLoading, handleInputChange);
                              if (field) return field;

                              // Submit button
                              if (subComponent.type === EmbeddedFlowComponentType.Action && subComponent.eventType === EmbeddedFlowEventType.Submit) {
                                return (
                                  <Button
                                    key={subComponent.id ?? compIndex}
                                    type="submit"
                                    fullWidth
                                    variant={subComponent.variant === 'PRIMARY' ? 'contained' : 'outlined'}
                                    disabled={isLoading || !isValid}
                                    sx={{mt: 2}}
                                  >
                                    {t(resolve(subComponent.label)!)}
                                  </Button>
                                );
                              }

                              return null;
                            })}
                          </Box>
                        );
                      }

                      return null;
                    })}
                  </Box>
                )}
              </>
            );
          }}
        </AcceptInvite>
      </StyledPaper>
    </Stack>
  );
}
