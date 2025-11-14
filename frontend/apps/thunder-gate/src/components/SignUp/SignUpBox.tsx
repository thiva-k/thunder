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

// Tracker: https://github.com/asgardeo/javascript/issues/222
/* eslint-disable @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-return */

import type {JSX, Key} from 'react';
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
} from '@wso2/oxygen-ui';
import {SignUp} from '@asgardeo/react';
import {Smartphone, Google, Facebook, GitHub, Eye, EyeClosed} from '@wso2/oxygen-ui-icons-react';
import {useNavigate, useSearchParams} from 'react-router';
import {useState} from 'react';
import ROUTES from '../../constants/routes';

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

export default function SignUpBox(): JSX.Element {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [showPasswordMap, setShowPasswordMap] = useState<Record<string, boolean>>({});

  const currentParams = searchParams.toString();
  const signInUrl = currentParams ? `${ROUTES.AUTH.SIGN_IN}?${currentParams}` : ROUTES.AUTH.SIGN_IN;

  const togglePasswordVisibility = (identifier: string): void => {
    setShowPasswordMap((prev: Record<string, boolean>) => ({
      ...prev,
      [identifier]: !prev[identifier],
    }));
  };

  return (
    <Stack gap={5}>
      <ColorSchemeImage
        src={{
          light: `${import.meta.env.BASE_URL}/assets/images/logo.svg`,
          dark: `${import.meta.env.BASE_URL}/assets/images/logo-inverted.svg`,
        }}
        alt={{light: 'Logo (Light)', dark: 'Logo (Dark)'}}
        height={30}
        width="auto"
      />
      <StyledPaper variant="outlined">
        <SignUp
          shouldRedirectAfterSignUp={false}
          onComplete={() => {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            navigate(signInUrl);
          }}
        >
          {({errors, handleSubmit, isLoading, components}) => (
            <>
              <Typography component="h1" variant="h5" sx={{width: '100%', mb: 2}}>
                Create Account
              </Typography>

              {!components ? (
                <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                  <Typography>Loading registration...</Typography>
                </Box>
              ) : (
                <>
                  {errors && Object.keys(errors).length > 0 && (
                    <Alert severity="error" sx={{mb: 2}}>
                      <AlertTitle>Error</AlertTitle>
                      {Object.entries(errors).map(([key, err]) => (
                        <Typography key={key} variant="body2">
                          {String(err)}
                        </Typography>
                      ))}
                    </Alert>
                  )}

                  {/* Handle different flow types */}
                  {(() => {
                    if (components && components.length > 0 && components.some((c) => c.type === 'FORM')) {
                      return (
                        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                          {components
                            .filter((c) => c.type === 'FORM')
                            .map((form) => (
                              <Box
                                key={form.id}
                                component="form"
                                onSubmit={(event) => {
                                  event.preventDefault();
                                  const data = new FormData(event.currentTarget);
                                  const inputs: Record<string, string> = {};

                                  form.components
                                    ?.filter((c: {type: string}) => c.type === 'INPUT' || c.type === 'SELECT')
                                    .forEach((input: {config: {identifier: string}}) => {
                                      if (input.config?.identifier) {
                                        inputs[input.config.identifier] = data.get(input.config.identifier) as string;
                                      }
                                    });

                                  handleSubmit(form, inputs).catch(() => {
                                    // Error handled by onError callback
                                  });
                                }}
                                noValidate
                                sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
                              >
                                {form.components?.map(
                                  (component: {
                                    type: string;
                                    id: Key | null | undefined;
                                    config: {
                                      identifier: string;
                                      label: string;
                                      type: string;
                                      placeholder: string;
                                      required: boolean;
                                      hint: string;
                                      text: string;
                                      options?: string[];
                                    };
                                    variant: string;
                                  }) => {
                                    if (component.type === 'SELECT' && component.config?.options) {
                                      return (
                                        <FormControl key={component.id} fullWidth>
                                          <FormLabel htmlFor={component.config?.identifier}>
                                            {component.config?.label}
                                          </FormLabel>
                                          <Select
                                            displayEmpty
                                            size="small"
                                            id={component.config?.identifier}
                                            name={component.config?.identifier}
                                            required={component.config?.required}
                                            fullWidth
                                            disabled={isLoading}
                                            error={!!errors[component.config?.identifier]}
                                            defaultValue=""
                                          >
                                            <MenuItem value="" disabled>
                                              {component.config?.placeholder || 'Select an option'}
                                            </MenuItem>
                                            {component.config.options.map((option: string) => (
                                              <MenuItem key={option} value={option}>
                                                {option}
                                              </MenuItem>
                                            ))}
                                          </Select>
                                          {errors[component.config?.identifier] && (
                                            <Typography variant="caption" color="error.main" sx={{mt: 0.5}}>
                                              {errors[component.config?.identifier]}
                                            </Typography>
                                          )}
                                          {component.config?.hint && (
                                            <Typography variant="caption" color="text.secondary">
                                              {component.config.hint}
                                            </Typography>
                                          )}
                                        </FormControl>
                                      );
                                    }

                                    if (component.type === 'INPUT') {
                                      const isPasswordField: boolean = component.config?.type === 'password';
                                      const showPassword: boolean =
                                        showPasswordMap[component.config?.identifier] || false;

                                      let inputType: string =
                                        component.config?.type === 'text' ? 'text' : component.config?.type;

                                      if (isPasswordField) {
                                        inputType = showPassword ? 'text' : 'password';
                                      }

                                      return (
                                        <FormControl key={component.id}>
                                          <FormLabel htmlFor={component.config?.identifier}>
                                            {component.config?.label}
                                          </FormLabel>
                                          <TextField
                                            id={component.config?.identifier}
                                            name={component.config?.identifier}
                                            type={inputType}
                                            placeholder={component.config?.placeholder}
                                            required={component.config?.required}
                                            fullWidth
                                            variant="outlined"
                                            disabled={isLoading}
                                            error={!!errors[component.config?.identifier]}
                                            helperText={errors[component.config?.identifier]}
                                            slotProps={
                                              isPasswordField
                                                ? {
                                                    input: {
                                                      endAdornment: (
                                                        <InputAdornment position="end">
                                                          <IconButton
                                                            aria-label="toggle password visibility"
                                                            onClick={(): void =>
                                                              togglePasswordVisibility(component.config?.identifier)
                                                            }
                                                            edge="end"
                                                            disabled={isLoading}
                                                          >
                                                            {showPassword ? <EyeClosed /> : <Eye />}
                                                          </IconButton>
                                                        </InputAdornment>
                                                      ),
                                                    },
                                                  }
                                                : undefined
                                            }
                                          />
                                          {component.config?.hint && (
                                            <Typography variant="caption" color="text.secondary">
                                              {component.config.hint}
                                            </Typography>
                                          )}
                                        </FormControl>
                                      );
                                    }

                                    if (component.type === 'BUTTON' && component.config?.type === 'submit') {
                                      return (
                                        <Button
                                          key={component.id}
                                          type="submit"
                                          fullWidth
                                          variant={component.variant === 'PRIMARY' ? 'contained' : 'outlined'}
                                          disabled={isLoading}
                                          sx={{mt: 2}}
                                        >
                                          {isLoading ? 'Creating account...' : component.config?.text}
                                        </Button>
                                      );
                                    }
                                    return null;
                                  },
                                )}
                              </Box>
                            ))}
                        </Box>
                      );
                    }

                    // Check if we have button components (multi-option flow)
                    if (components && components.length > 0 && components.some((c) => c.type === 'BUTTON')) {
                      return (
                        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                          {components
                            .filter((c) => c.type === 'BUTTON')
                            .map((button) => {
                              const actionId = (button.config?.actionId as string) ?? '';

                              const getIcon = () => {
                                if (actionId.includes('google')) return <Google />;
                                if (actionId.includes('facebook')) return <Facebook />;
                                if (actionId.includes('github')) return <GitHub />;
                                if (actionId.includes('mobile')) return <Smartphone />;
                                return null;
                              };

                              const getLabel = () => {
                                if (button.config?.text && typeof button.config.text === 'string') {
                                  return button.config.text;
                                }

                                if (actionId.includes('google')) return 'Sign up with Google';
                                if (actionId.includes('github')) return 'Sign up with GitHub';
                                if (actionId.includes('facebook')) return 'Sign up with Facebook';
                                if (actionId.includes('mobile')) return 'Sign up with SMS OTP';
                                return `Sign up with ${actionId.replace('_auth', '').replace('_', ' ')}`;
                              };

                              return (
                                <Button
                                  key={button.id}
                                  fullWidth
                                  variant="outlined"
                                  onClick={() => {
                                    handleSubmit(button, {}).catch(() => {
                                      // Error handled by onError callback
                                    });
                                  }}
                                  disabled={isLoading}
                                  startIcon={getIcon()}
                                >
                                  {getLabel()}
                                </Button>
                              );
                            })}
                        </Box>
                      );
                    }

                    // SignUpBox fallback error
                    return (
                      <Alert severity="error" sx={{mb: 2}}>
                        <AlertTitle>Oops, that didn&apos;t work</AlertTitle>
                        We&apos;re sorry, we ran into a problem. Please try again!
                      </Alert>
                    );
                  })()}
                </>
              )}

              <Typography sx={{textAlign: 'center', mt: 3}}>
                Already have an account?{' '}
                <Button
                  variant="text"
                  onClick={() => {
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    navigate(signInUrl);
                  }}
                  sx={{
                    p: 0,
                    minWidth: 'auto',
                    textTransform: 'none',
                    color: 'primary.main',
                    textDecoration: 'underline',
                    '&:hover': {
                      textDecoration: 'underline',
                      backgroundColor: 'transparent',
                    },
                  }}
                >
                  Sign in
                </Button>
              </Typography>
            </>
          )}
        </SignUp>
      </StyledPaper>
    </Stack>
  );
}
