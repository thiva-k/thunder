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

import type {JSX} from 'react';
import {
  Box,
  Button,
  Checkbox,
  Divider,
  FormLabel,
  FormControl,
  Alert,
  FormControlLabel,
  TextField,
  Typography,
  styled,
  AlertTitle,
  Paper,
} from '@wso2/oxygen-ui';
import {useState} from 'react';
import {SignIn} from '@asgardeo/react';
import {Smartphone} from 'lucide-react';
import {Google, Facebook, GitHub} from '@thunder/ui';

const LoginBox = styled(Paper)(({theme}) => ({
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

export default function SignInBox(): JSX.Element {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [emailError, setEmailError] = useState(false);
  const [emailErrorMessage, setEmailErrorMessage] = useState('');
  const [passwordError, setPasswordError] = useState(false);
  const [passwordErrorMessage, setPasswordErrorMessage] = useState('');

  const validateInputs = () => {
    let isValid = true;

    if (!email || !/\S+@\S+\.\S+/.test(email)) {
      setEmailError(true);
      setEmailErrorMessage('Please enter a valid email address.');
      isValid = false;
    } else {
      setEmailError(false);
      setEmailErrorMessage('');
    }

    if (!password || password.length < 6) {
      setPasswordError(true);
      setPasswordErrorMessage('Password must be at least 6 characters long.');
      isValid = false;
    } else {
      setPasswordError(false);
      setPasswordErrorMessage('');
    }

    return isValid;
  };

  return (
    <LoginBox variant="outlined">
      <SignIn
        onError={(error) => {
          setEmailError(true);
          setEmailErrorMessage(error?.message || 'Authentication failed');
        }}
      >
        {({onSubmit, isLoading, components, error, isInitialized}) => (
          <>
            <Typography component="h1" variant="h5" sx={{width: '100%', mb: 2}}>
              Sign in
            </Typography>

            {!isInitialized ? (
              <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                <Typography>Loading authentication...</Typography>
              </Box>
            ) : (
              <>
                {error && (
                  <Alert severity="error" sx={{mb: 2}}>
                    <AlertTitle>Error</AlertTitle>
                    {error.message || error.toString()}
                  </Alert>
                )}

                {/* Handle different flow types */}
                {(() => {
                  // Check if we have button components (multi-option flow)
                  if (components && components.length > 0 && components.some((c) => c.type === 'BUTTON')) {
                    return (
                      <>
                        {/* Always render BasicAuth form first if basic_auth is available */}
                        {components.some((c) => c.config?.actionId === 'basic_auth') && (
                          <Box
                            component="form"
                            onSubmit={(event) => {
                              event.preventDefault();
                              if (validateInputs()) {
                                // Tracker: https://github.com/asgardeo/javascript/issues/222
                                // eslint-disable-next-line @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-explicit-any
                                onSubmit({
                                  actionId: 'basic_auth',
                                  inputs: {
                                    username: email,
                                    password,
                                  },
                                  // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                } as any).catch(() => {
                                  // Error handled by onError callback
                                });
                              }
                            }}
                            noValidate
                            sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2, mb: 2, p: 3}}
                          >
                            <FormControl>
                              <FormLabel htmlFor="email">Email</FormLabel>
                              <TextField
                                error={emailError}
                                helperText={emailErrorMessage}
                                id="email"
                                type="email"
                                name="email"
                                placeholder="your@email.com"
                                autoComplete="email"
                                autoFocus
                                required
                                fullWidth
                                variant="outlined"
                                color={emailError ? 'error' : 'primary'}
                                disabled={isLoading}
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                              />
                            </FormControl>
                            <FormControl>
                              <FormLabel htmlFor="password">Password</FormLabel>
                              <TextField
                                error={passwordError}
                                helperText={passwordErrorMessage}
                                name="password"
                                placeholder="••••••"
                                type="password"
                                id="password"
                                autoComplete="current-password"
                                required
                                fullWidth
                                variant="outlined"
                                color={passwordError ? 'error' : 'primary'}
                                disabled={isLoading}
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                              />
                            </FormControl>
                            <FormControlLabel
                              control={<Checkbox value="remember" color="primary" disabled={isLoading} />}
                              label="Remember me"
                            />
                            <Button type="submit" fullWidth variant="contained" disabled={isLoading} sx={{mt: 2}}>
                              {isLoading ? 'Signing in...' : 'Sign in'}
                            </Button>
                          </Box>
                        )}

                        {/* Show divider if there are other auth options besides basic_auth */}
                        {components.some((c) => c.type === 'BUTTON' && c.config?.actionId !== 'basic_auth') && (
                          <Divider>or</Divider>
                        )}

                        {/* Show other authentication options as buttons */}
                        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                          {components
                            .filter((c) => c.type === 'BUTTON')
                            .filter((c) => c.config?.actionId !== 'basic_auth')
                            .map((button) => {
                              const actionId = button.config?.actionId as string ?? '';

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
                                if (actionId.includes('google')) return 'Sign in with Google';
                                if (actionId.includes('github')) return 'Sign in with GitHub';
                                if (actionId.includes('facebook')) return 'Sign in with Facebook';
                                if (actionId.includes('mobile')) return 'Sign in with SMS OTP';
                                return `Sign in with ${actionId.replace('_auth', '').replace('_', ' ')}`;
                              };

                              return (
                                <Button
                                  key={button.id}
                                  fullWidth
                                  variant="outlined"
                                  onClick={() => {
                                    // Tracker: https://github.com/asgardeo/javascript/issues/222
                                    // eslint-disable-next-line @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-explicit-any
                                    onSubmit({actionId} as any).catch(() => {
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
                      </>
                    );
                  }

                  // Handle FORM components
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

                                // Extract inputs from form components
                                form.components
                                  ?.filter((c) => c.type === 'INPUT')
                                  .forEach((input) => {
                                    if (input.config?.identifier) {
                                      inputs[input.config.identifier as string] = data.get(
                                        input.config.identifier as string,
                                      ) as string;
                                    }
                                  });

                                // Tracker: https://github.com/asgardeo/javascript/issues/222
                                // eslint-disable-next-line @typescript-eslint/no-unsafe-argument, @typescript-eslint/no-explicit-any
                                onSubmit({inputs} as any).catch(() => {
                                  // Error handled by onError callback
                                });
                              }}
                              noValidate
                              sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
                            >
                              {form.components?.map((component) => {
                                if (component.type === 'INPUT') {
                                  return (
                                    <FormControl key={component.id}>
                                      <FormLabel htmlFor={component.config?.identifier as string}>
                                        {component.config?.label as string}
                                      </FormLabel>
                                      <TextField
                                        id={component.config?.identifier as string}
                                        name={component.config?.identifier as string}
                                        type={
                                          (component.config?.type as string) === 'text'
                                            ? 'text'
                                            : (component.config?.type as string)
                                        }
                                        placeholder={component.config?.placeholder as string}
                                        required={component.config?.required as boolean}
                                        fullWidth
                                        variant="outlined"
                                        disabled={isLoading}
                                      />
                                      {(component.config?.hint as string) && (
                                        <Typography variant="caption" color="text.secondary">
                                          {component.config.hint as string}
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
                                      {isLoading ? 'Submitting...' : (component.config?.text as string)}
                                    </Button>
                                  );
                                }
                                return null;
                              })}
                            </Box>
                          ))}
                      </Box>
                    );
                  }

                  // SignInBox fallback error
                  return (
                    <Alert severity="error" sx={{mb: 2}}>
                      <AlertTitle>Oops, that didn&apos;t work</AlertTitle>
                      We&apos;re sorry, we ran into a problem. Please try again!
                    </Alert>
                  );
                })()}
              </>
            )}
          </>
        )}
      </SignIn>
    </LoginBox>
  );
}
