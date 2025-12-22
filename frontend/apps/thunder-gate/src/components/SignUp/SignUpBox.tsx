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
import {EmbeddedFlowComponentType, EmbeddedFlowEventType, SignUp, type EmbeddedFlowComponent} from '@asgardeo/react';
import {Eye, EyeClosed} from '@wso2/oxygen-ui-icons-react';
import {useNavigate, useSearchParams} from 'react-router';
import {useState} from 'react';
import {Trans, useTranslation} from 'react-i18next';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';
import {mapEmbeddedFlowTextVariant, useBranding} from '@thunder/shared-branding';
import getIntegrationIcon from '../../utils/getIntegrationIcon';
import ROUTES from '../../constants/routes';

interface ComponentWithType {
  type: string;
  [key: string]: unknown;
}

interface ComponentWithRef {
  ref: string;
  [key: string]: unknown;
}

interface ComponentWithLabel {
  label: string;
  [key: string]: unknown;
}

interface ComponentWithId {
  id: string;
  [key: string]: unknown;
}

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
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();
  const {images, theme: brandingTheme, isBrandingEnabled} = useBranding();
  const theme = useTheme();

  const [showPasswordMap, setShowPasswordMap] = useState<Record<string, boolean>>({});
  const [formInputs, setFormInputs] = useState<Record<string, string>>({});
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  const currentParams = searchParams.toString();
  const signInUrl = currentParams ? `${ROUTES.AUTH.SIGN_IN}?${currentParams}` : ROUTES.AUTH.SIGN_IN;

  const togglePasswordVisibility = (identifier: string): void => {
    setShowPasswordMap((prev: Record<string, boolean>) => ({
      ...prev,
      [identifier]: !prev[identifier],
    }));
  };

  const validateForm = (components: EmbeddedFlowComponent[]): boolean => {
    const errors: Record<string, string> = {};
    let isValid = true;

    components.forEach((component) => {
      if (
        ((component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.TextInput ||
          (component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.PasswordInput ||
          component.type === 'PHONE_INPUT' ||
          component.type === 'OTP_INPUT') &&
        component.required &&
        component.ref &&
        typeof component.ref === 'string' &&
        typeof component.label === 'string'
      ) {
        const value = formInputs[component.ref] ?? '';
        if (!value.trim()) {
          errors[component.ref] = `${t('validations:form.field.required', {field: t(resolve(component.label)!)})}`;
          isValid = false;
        }
      }
    });

    setFormErrors(errors);
    return isValid;
  };

  const updateInput = (ref: string, value: string): void => {
    setFormInputs((prev) => ({...prev, [ref]: value}));

    // Clear error when user starts typing
    if (formErrors[ref]) {
      setFormErrors((prev) => ({...prev, [ref]: ''}));
    }
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
        sx={{
          display: {xs: 'flex', md: 'none'},
        }}
      />
      {/* Add branded logo above the sign-up box for desktop */}
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
        <SignUp
          shouldRedirectAfterSignUp={false}
          onComplete={() => {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            navigate(signInUrl);
          }}
        >
          {({errors, handleSubmit, isLoading, components}) => (
            <>
              {!components ? (
                <Box sx={{display: 'flex', justifyContent: 'center', p: 3}}>
                  <CircularProgress />
                </Box>
              ) : (
                <>
                  {errors && Object.keys(errors).length > 0 && (
                    <Alert severity="error" sx={{mb: 2}}>
                      <AlertTitle>{t('Error')}</AlertTitle>
                      {Object.entries(errors).map(([key, error]: [string, unknown]) => (
                        <Typography key={key} variant="body2">
                          {String(error)}
                        </Typography>
                      ))}
                    </Alert>
                  )}

                  {/* Handle different flow types */}
                  {(() => {
                    // Handle new component structure with TEXT and BLOCK types
                    if (components && components.length > 0) {
                      return (
                        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                          {components.map(
                            (
                              component: ComponentWithType &
                                ComponentWithId &
                                ComponentWithLabel & {
                                  variant?: string;
                                  components?: EmbeddedFlowComponent[];
                                },
                              index: number,
                            ) => {
                              // Handle TEXT components (headings)
                              if ((component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.Text) {
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

                              // Handle BLOCK components (form blocks or action blocks)
                              if ((component.type as EmbeddedFlowComponentType) === EmbeddedFlowComponentType.Block) {
                                const blockComponents: EmbeddedFlowComponent[] = component.components ?? [];

                                // Check if this block contains a SUBMIT action (form block) or TRIGGER actions (social login block)
                                const hasSubmitAction = blockComponents.some(
                                  (blockComponent: EmbeddedFlowComponent) =>
                                    ((blockComponent.type as EmbeddedFlowComponentType) ===
                                      EmbeddedFlowComponentType.Action &&
                                      blockComponent.eventType === EmbeddedFlowEventType.Submit) ||
                                    (blockComponent.type === 'RESEND' &&
                                      blockComponent.eventType === EmbeddedFlowEventType.Submit),
                                );

                                const hasTriggerAction = blockComponents.some(
                                  (blockComponent: EmbeddedFlowComponent) =>
                                    (blockComponent.type as EmbeddedFlowComponentType) ===
                                      EmbeddedFlowComponentType.Action &&
                                    blockComponent.eventType === EmbeddedFlowEventType.Trigger,
                                );

                                // Render form block with submission logic
                                if (hasSubmitAction) {
                                  return (
                                    <Box
                                      key={component.id ?? index}
                                      component="form"
                                      onSubmit={(event) => {
                                        event.preventDefault();
                                        if (validateForm(blockComponents)) {
                                          // Tracker: https://github.com/asgardeo/javascript/issues/222
                                          handleSubmit(component, formInputs).catch(() => {
                                            // Error handled by onError callback
                                          });
                                        }
                                      }}
                                      noValidate
                                      sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
                                    >
                                      {(component.components ?? []).map(
                                        (_subComponent: EmbeddedFlowComponent, compIndex: number) => {
                                          const subComponent: ComponentWithType &
                                            ComponentWithRef &
                                            ComponentWithLabel &
                                            ComponentWithId & {
                                              placeholder?: string;
                                              required?: boolean;
                                              options?: string[];
                                              hint?: string;
                                              variant?: string;
                                              eventType?: string;
                                            } = _subComponent as unknown as ComponentWithType &
                                            ComponentWithRef &
                                            ComponentWithLabel &
                                            ComponentWithId & {
                                              placeholder?: string;
                                              required?: boolean;
                                              options?: string[];
                                              hint?: string;
                                              variant?: string;
                                              eventType?: string;
                                            };

                                          // Handle TEXT_INPUT components
                                          if (
                                            (subComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.TextInput &&
                                            subComponent.ref
                                          ) {
                                            return (
                                              <FormControl key={subComponent.id ?? compIndex}>
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <TextField
                                                  error={!!formErrors[subComponent.ref]}
                                                  helperText={formErrors[subComponent.ref]}
                                                  id={subComponent.ref}
                                                  name={subComponent.ref}
                                                  type="text"
                                                  placeholder={t(resolve(subComponent.placeholder)!)}
                                                  autoComplete={(() => {
                                                    if (subComponent.ref === 'username') return 'username';
                                                    if (subComponent.ref === 'email') return 'email';
                                                    return 'off';
                                                  })()}
                                                  autoFocus={subComponent.ref === 'firstName'}
                                                  required={subComponent.required}
                                                  fullWidth
                                                  variant="outlined"
                                                  color={formErrors[subComponent.ref] ? 'error' : 'primary'}
                                                  disabled={isLoading}
                                                  value={formInputs[subComponent.ref] ?? ''}
                                                  onChange={(e) => updateInput(subComponent.ref, e.target.value)}
                                                />
                                              </FormControl>
                                            );
                                          }

                                          // Handle PASSWORD_INPUT components
                                          if (
                                            (subComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.PasswordInput &&
                                            subComponent.ref
                                          ) {
                                            const showPasswordForField = showPasswordMap[subComponent.ref] ?? false;

                                            return (
                                              <FormControl key={subComponent.id ?? compIndex}>
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <TextField
                                                  error={!!formErrors[subComponent.ref]}
                                                  helperText={formErrors[subComponent.ref]}
                                                  id={subComponent.ref}
                                                  name={subComponent.ref}
                                                  type={showPasswordForField ? 'text' : 'password'}
                                                  placeholder={t(resolve(subComponent.placeholder)!)}
                                                  autoComplete={
                                                    subComponent.ref === 'password' ? 'new-password' : 'off'
                                                  }
                                                  required={subComponent.required}
                                                  fullWidth
                                                  variant="outlined"
                                                  color={formErrors[subComponent.ref] ? 'error' : 'primary'}
                                                  disabled={isLoading}
                                                  value={formInputs[subComponent.ref] ?? ''}
                                                  onChange={(e) => updateInput(subComponent.ref, e.target.value)}
                                                  slotProps={{
                                                    input: {
                                                      endAdornment: (
                                                        <InputAdornment position="end">
                                                          <IconButton
                                                            aria-label="toggle password visibility"
                                                            onClick={() => togglePasswordVisibility(subComponent.ref)}
                                                            edge="end"
                                                            disabled={isLoading}
                                                          >
                                                            {showPasswordForField ? <EyeClosed /> : <Eye />}
                                                          </IconButton>
                                                        </InputAdornment>
                                                      ),
                                                    },
                                                  }}
                                                />
                                              </FormControl>
                                            );
                                          }

                                          // Handle SELECT components
                                          if (
                                            subComponent.type === 'SELECT' &&
                                            subComponent.options &&
                                            subComponent.ref
                                          ) {
                                            return (
                                              <FormControl key={subComponent.id ?? compIndex} fullWidth>
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <Select
                                                  displayEmpty
                                                  size="small"
                                                  id={subComponent.ref}
                                                  name={subComponent.ref}
                                                  required={subComponent.required}
                                                  fullWidth
                                                  disabled={isLoading}
                                                  error={!!formErrors[subComponent.ref]}
                                                  value={formInputs[subComponent.ref] ?? ''}
                                                  onChange={(e) => updateInput(subComponent.ref, e.target.value)}
                                                >
                                                  <MenuItem value="" disabled>
                                                    {subComponent.placeholder ?? 'Select an option'}
                                                  </MenuItem>
                                                  {subComponent.options.map((option: string) => (
                                                    <MenuItem key={option} value={option}>
                                                      {option}
                                                    </MenuItem>
                                                  ))}
                                                </Select>
                                                {formErrors[subComponent.ref] && (
                                                  <Typography variant="caption" color="error.main" sx={{mt: 0.5}}>
                                                    {formErrors[subComponent.ref]}
                                                  </Typography>
                                                )}
                                                {subComponent.hint && (
                                                  <Typography variant="caption" color="text.secondary">
                                                    {subComponent.hint}
                                                  </Typography>
                                                )}
                                              </FormControl>
                                            );
                                          }

                                          // Handle PHONE_INPUT components
                                          if (
                                            subComponent.type === 'PHONE_INPUT' &&
                                            subComponent.ref &&
                                            typeof subComponent.ref === 'string'
                                          ) {
                                            return (
                                              <FormControl
                                                key={subComponent.id ?? compIndex}
                                                required={subComponent.required}
                                              >
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <TextField
                                                  fullWidth
                                                  error={!!(subComponent.ref && formErrors[subComponent.ref])}
                                                  helperText={
                                                    subComponent.ref ? formErrors[subComponent.ref] : undefined
                                                  }
                                                  id={subComponent.ref}
                                                  name={subComponent.ref}
                                                  type="tel"
                                                  placeholder={t(
                                                    resolve(subComponent.placeholder) ?? subComponent.placeholder ?? '',
                                                  )}
                                                  autoComplete="tel"
                                                  required={subComponent.required}
                                                  variant="outlined"
                                                  color={
                                                    subComponent.ref && formErrors[subComponent.ref]
                                                      ? 'error'
                                                      : 'primary'
                                                  }
                                                  disabled={isLoading}
                                                  value={subComponent.ref ? (formInputs[subComponent.ref] ?? '') : ''}
                                                  onChange={(e) => updateInput(subComponent.ref, e.target.value)}
                                                />
                                              </FormControl>
                                            );
                                          }

                                          // Handle OTP_INPUT components
                                          if (
                                            subComponent.type === 'OTP_INPUT' &&
                                            subComponent.ref &&
                                            typeof subComponent.ref === 'string'
                                          ) {
                                            const otpLength = 6;
                                            const otpValue = formInputs[subComponent.ref] ?? '';
                                            const otpDigits = otpValue
                                              .padEnd(otpLength, ' ')
                                              .split('')
                                              .slice(0, otpLength);

                                            return (
                                              <FormControl
                                                key={subComponent.id ?? compIndex}
                                                required={subComponent.required}
                                              >
                                                <FormLabel htmlFor={subComponent.ref}>
                                                  {t(resolve(subComponent.label)!)}
                                                </FormLabel>
                                                <Box
                                                  sx={{
                                                    display: 'flex',
                                                    gap: 1,
                                                    justifyContent: 'center',
                                                    mt: 1,
                                                  }}
                                                >
                                                  {otpDigits.map((digit, idx) => (
                                                    <TextField
                                                      key={`otp-${subComponent.ref}`}
                                                      slotProps={{
                                                        htmlInput: {
                                                          maxLength: 1,
                                                          style: {textAlign: 'center', fontSize: '1.5rem'},
                                                          'aria-label': `OTP digit ${idx + 1}`,
                                                        },
                                                      }}
                                                      value={digit.trim()}
                                                      onChange={(e) => {
                                                        const {value} = e.target;
                                                        if (!/^\d*$/.test(value)) return;

                                                        const newOtp = otpDigits.map((d, i) =>
                                                          i === idx ? value : d.trim(),
                                                        );
                                                        updateInput(subComponent.ref, newOtp.join(''));

                                                        // Auto-focus next input
                                                        if (value && idx < otpLength - 1) {
                                                          // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
                                                          const nextInput = document.querySelector(
                                                            `input[aria-label="OTP digit ${idx + 2}"]`,
                                                          )! as HTMLInputElement;
                                                          nextInput.focus();
                                                        }
                                                      }}
                                                      onKeyDown={(e) => {
                                                        // Handle backspace to move to previous input
                                                        if (
                                                          e.key === 'Backspace' &&
                                                          !otpDigits[idx].trim() &&
                                                          idx > 0
                                                        ) {
                                                          // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
                                                          const prevInput = document.querySelector(
                                                            `input[aria-label="OTP digit ${idx}"]`,
                                                          )! as HTMLInputElement;
                                                          prevInput.focus();
                                                        }
                                                      }}
                                                      onPaste={(e) => {
                                                        e.preventDefault();
                                                        const pastedData = e.clipboardData.getData('text/plain');
                                                        const digits = pastedData
                                                          .replace(/\D/g, '')
                                                          .slice(0, otpLength);
                                                        updateInput(subComponent.ref, digits);

                                                        // Focus last filled input
                                                        const lastIdx = Math.min(digits.length, otpLength - 1);
                                                        // eslint-disable-next-line @typescript-eslint/no-unnecessary-type-assertion
                                                        const lastInput = document.querySelector(
                                                          `input[aria-label="OTP digit ${lastIdx + 1}"]`,
                                                        )! as HTMLInputElement;
                                                        lastInput.focus();
                                                      }}
                                                      error={!!(subComponent.ref && formErrors[subComponent.ref])}
                                                      disabled={isLoading}
                                                      variant="outlined"
                                                      sx={{
                                                        width: 48,
                                                        '& input': {
                                                          padding: '12px 8px',
                                                        },
                                                      }}
                                                    />
                                                  ))}
                                                </Box>
                                                {subComponent.ref && formErrors[subComponent.ref] && (
                                                  <Typography variant="caption" color="error" sx={{mt: 0.5, ml: 1.75}}>
                                                    {formErrors[subComponent.ref]}
                                                  </Typography>
                                                )}
                                              </FormControl>
                                            );
                                          }

                                          // Handle ACTION components (submit buttons)
                                          if (
                                            (subComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.Action &&
                                            subComponent.eventType === EmbeddedFlowEventType.Submit
                                          ) {
                                            return (
                                              <Button
                                                key={subComponent.id ?? compIndex}
                                                type="submit"
                                                fullWidth
                                                variant={subComponent.variant === 'PRIMARY' ? 'contained' : 'outlined'}
                                                disabled={isLoading}
                                                sx={{mt: 2}}
                                              >
                                                {isLoading ? t('Creating account...') : t(resolve(subComponent.label)!)}
                                              </Button>
                                            );
                                          }

                                          // Handle RESEND components
                                          if (
                                            subComponent.type === 'RESEND' &&
                                            subComponent.eventType === EmbeddedFlowEventType.Submit
                                          ) {
                                            return (
                                              <Button
                                                key={subComponent.id ?? compIndex}
                                                type="submit"
                                                fullWidth
                                                variant="text"
                                                disabled={isLoading}
                                                sx={{mt: 1}}
                                              >
                                                {t(resolve(subComponent.label)!)}
                                              </Button>
                                            );
                                          }

                                          // Handle TRIGGER action buttons within forms (e.g., OTP verify button)
                                          if (
                                            (subComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.Action &&
                                            subComponent.eventType === EmbeddedFlowEventType.Trigger
                                          ) {
                                            return (
                                              <Button
                                                key={subComponent.id ?? compIndex}
                                                fullWidth
                                                variant={subComponent.variant === 'PRIMARY' ? 'contained' : 'outlined'}
                                                disabled={isLoading}
                                                onClick={() => {
                                                  if (validateForm(blockComponents)) {
                                                    // Tracker: https://github.com/asgardeo/javascript/issues/222
                                                    handleSubmit(component, formInputs).catch(() => {
                                                      // Error handled by onError callback
                                                    });
                                                  }
                                                }}
                                                sx={{mt: 2}}
                                              >
                                                {t(resolve(subComponent.label)!)}
                                              </Button>
                                            );
                                          }

                                          return null;
                                        },
                                      )}
                                    </Box>
                                  );
                                }

                                // Render action block for social login buttons
                                if (hasTriggerAction) {
                                  return (
                                    <Box
                                      key={component.id ?? index}
                                      sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2, mt: 2}}
                                    >
                                      {blockComponents.map(
                                        (_actionComponent: EmbeddedFlowComponent, actionIndex: number) => {
                                          const actionComponent = _actionComponent as ComponentWithType &
                                            ComponentWithId &
                                            ComponentWithLabel & {
                                              eventType?: string;
                                              variant?: string;
                                              image?: string;
                                            };

                                          // Handle TRIGGER action buttons (social login)
                                          if (
                                            (actionComponent.type as EmbeddedFlowComponentType) ===
                                              EmbeddedFlowComponentType.Action &&
                                            actionComponent.eventType === EmbeddedFlowEventType.Trigger
                                          ) {
                                            return (
                                              <Button
                                                key={actionComponent.id ?? actionIndex}
                                                fullWidth
                                                variant="outlined"
                                                disabled={isLoading}
                                                onClick={() => {
                                                  // Tracker: https://github.com/asgardeo/javascript/issues/222
                                                  handleSubmit(component, {}).catch(() => {
                                                    // Error handled by onError callback
                                                  });
                                                }}
                                                startIcon={getIntegrationIcon(
                                                  actionComponent.label ?? '',
                                                  actionComponent.image ?? '',
                                                )}
                                              >
                                                {t(resolve(actionComponent.label)!)}
                                              </Button>
                                            );
                                          }

                                          return null;
                                        },
                                      )}
                                    </Box>
                                  );
                                }

                                return null;
                              }

                              return null;
                            },
                          )}
                        </Box>
                      );
                    }

                    // SignUpBox fallback error
                    return (
                      <Alert severity="error" sx={{mb: 2}}>
                        <AlertTitle>{t("Oops, that didn't work")}</AlertTitle>
                        {t("We're sorry, we ran into a problem. Please try again!")}
                      </Alert>
                    );
                  })()}
                </>
              )}

              <Typography sx={{textAlign: 'center', mt: 3}}>
                <Trans i18nKey="signup:redirect.to.signin">
                  Already have an account?
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
                </Trans>
              </Typography>
            </>
          )}
        </SignUp>
      </StyledPaper>
    </Stack>
  );
}
