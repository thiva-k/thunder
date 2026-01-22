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

import {useState, useEffect, useMemo, type JSX} from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  Box,
  Typography,
  Button,
  Alert,
  AlertTitle,
  CircularProgress,
  TextField,
  IconButton,
  FormControl,
  FormLabel,
  Select,
  MenuItem,
  Stepper,
  Step,
  StepLabel,
} from '@wso2/oxygen-ui';
import {X, Copy, Check} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useForm, Controller} from 'react-hook-form';
import {z} from 'zod';
import {zodResolver} from '@hookform/resolvers/zod';
import {EmbeddedFlowComponentType, EmbeddedFlowEventType, InviteUser, type EmbeddedFlowComponent, type InviteUserRenderProps} from '@asgardeo/react';
import {mapEmbeddedFlowTextVariant} from '@thunder/shared-branding';
import {useTemplateLiteralResolver} from '@thunder/shared-hooks';

/** Typed shape for flow sub-components */
type FlowSubComponent = EmbeddedFlowComponent & {
  placeholder?: string;
  required?: boolean;
  options?: any[];
  hint?: string;
  variant?: string;
  eventType?: string;
};

export interface InviteUserDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess?: (inviteLink: string) => void;
}

const STEPS = ['User Details', 'Invite Link'];

export default function InviteUserDialog({open, onClose, onSuccess}: InviteUserDialogProps): JSX.Element {
  const {resolve} = useTemplateLiteralResolver();
  const {t} = useTranslation();
  const [copied, setCopied] = useState(false);
  const [activeStep, setActiveStep] = useState(0);

  const handleCopy = () => {
    setCopied(true);
    setTimeout(() => setCopied(false), 3000);
  };

  const handleClose = () => {
    setActiveStep(0);
    setCopied(false);
    onClose();
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

  /**
   * Build Zod schema dynamically based on flow components
   */
  const buildFormSchema = (components: EmbeddedFlowComponent[]): z.ZodObject<any> => {
    const shape: Record<string, z.ZodTypeAny> = {};

    const processComponents = (comps: EmbeddedFlowComponent[]) => {
      comps.forEach((comp) => {
        if (comp.type === EmbeddedFlowComponentType.Block && comp.components) {
          processComponents(comp.components);
        } else if (
          (comp.type === EmbeddedFlowComponentType.TextInput ||
            comp.type === 'EMAIL_INPUT' ||
            comp.type === 'SELECT') &&
          comp.ref
        ) {
          let fieldSchema: z.ZodTypeAny = z.string();

          if (comp.type === 'EMAIL_INPUT') {
            fieldSchema = z.string().email('Please enter a valid email address');
          }

          if (comp.required) {
            fieldSchema = (fieldSchema as z.ZodString).min(1, `${t(resolve(comp.label)!) || comp.ref} is required`);
          } else {
            fieldSchema = (fieldSchema as z.ZodString).optional();
          }

          shape[comp.ref] = fieldSchema;
        }
      });
    };

    processComponents(components);
    return z.object(shape);
  };

  /**
   * Render form field using react-hook-form Controller
   */
  const renderFormField = (
    component: FlowSubComponent,
    index: number,
    control: any,
    errors: any,
    isLoading: boolean,
    handleInputChange: (field: string, value: string) => void,
  ) => {
    const {type, ref, label, placeholder, required, options, hint} = component;
    if (!ref) return null;

    // TEXT_INPUT
    if (type === EmbeddedFlowComponentType.TextInput) {
      return (
        <FormControl key={component.id ?? index} required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(label)!)}</FormLabel>
          <Controller
            name={ref}
            control={control}
            rules={{
              required: required ? `${t(resolve(label)!)} is required` : false,
            }}
            render={({field}) => (
              <TextField
                {...field}
                fullWidth
                size="small"
                id={ref}
                type="text"
                placeholder={t(resolve(placeholder) ?? placeholder ?? '')}
                autoComplete="off"
                required={required}
                variant="outlined"
                disabled={isLoading}
                error={!!errors[ref]}
                helperText={errors[ref]?.message as string}
                color={errors[ref] ? 'error' : 'primary'}
                onChange={(e) => {
                  field.onChange(e);
                  handleInputChange(ref, e.target.value);
                }}
              />
            )}
          />
        </FormControl>
      );
    }

    // EMAIL_INPUT
    if (type === 'EMAIL_INPUT') {
      return (
        <FormControl key={component.id ?? index} required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(label)!)}</FormLabel>
          <Controller
            name={ref}
            control={control}
            rules={{
              required: required ? `${t(resolve(label)!)} is required` : false,
              pattern: {
                value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                message: 'Please enter a valid email address',
              },
            }}
            render={({field}) => (
              <TextField
                {...field}
                fullWidth
                size="small"
                id={ref}
                type="email"
                placeholder={t(resolve(placeholder) ?? placeholder ?? '')}
                autoComplete="email"
                required={required}
                variant="outlined"
                disabled={isLoading}
                error={!!errors[ref]}
                helperText={errors[ref]?.message as string}
                color={errors[ref] ? 'error' : 'primary'}
                onChange={(e) => {
                  field.onChange(e);
                  handleInputChange(ref, e.target.value);
                }}
              />
            )}
          />
        </FormControl>
      );
    }

    // SELECT
    if (type === 'SELECT' && options) {
      return (
        <FormControl key={component.id ?? index} fullWidth>
          <FormLabel htmlFor={ref}>{t(resolve(label)!)}</FormLabel>
          <Controller
            name={ref}
            control={control}
            rules={{
              required: required ? `${t(resolve(label)!)} is required` : false,
            }}
            render={({field}) => (
              <>
                <Select
                  {...field}
                  value={field.value || ''}
                  displayEmpty
                  size="small"
                  id={ref}
                  required={required}
                  fullWidth
                  disabled={isLoading}
                  error={!!errors[ref]}
                  onChange={(e) => {
                    field.onChange(e);
                    handleInputChange(ref, e.target.value);
                  }}
                  renderValue={(selected) => {
                    if (!selected || selected === '') {
                      return <Typography sx={{color: 'text.secondary'}}>{t(resolve(placeholder) ?? 'Select an option')}</Typography>;
                    }
                    const selectedOption = options.find((opt: any) => getOptionValue(opt) === selected);
                    return selectedOption ? getOptionLabel(selectedOption) : selected;
                  }}
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
                {errors[ref] && (
                  <Typography variant="caption" color="error.main" sx={{mt: 0.5}}>
                    {errors[ref]?.message as string}
                  </Typography>
                )}
                {hint && (
                  <Typography variant="caption" color="text.secondary">
                    {hint}
                  </Typography>
                )}
              </>
            )}
          />
        </FormControl>
      );
    }

    return null;
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', pb: 1}}>
        <Typography variant="h3" component="h2">
          {t('users:inviteUser', 'Invite User')}
        </Typography>
        <IconButton onClick={handleClose} size="small" aria-label="close">
          <X size={18} />
        </IconButton>
      </DialogTitle>
      <DialogContent>
        <DialogContentText sx={{mb: 3, fontSize: '1rem'}}>
          {t('users:inviteUserDescription', 'Send an invite link to a new user to complete their registration')}
        </DialogContentText>

        <Stepper activeStep={activeStep} sx={{mb: 3}}>
          {STEPS.map((label) => (
            <Step key={label}>
              <StepLabel>{t(`users:invite.steps.${label.toLowerCase().replace(' ', '')}`, label)}</StepLabel>
            </Step>
          ))}
        </Stepper>

        <InviteUser
          onInviteLinkGenerated={(link: string) => {
            setActiveStep(1);
            onSuccess?.(link);
          }}
          onError={(error: Error) => console.error('User onboarding error:', error)}
        >
          {({values, error, isLoading, components, handleInputChange, handleSubmit, isInviteGenerated, inviteLink, copyInviteLink, inviteLinkCopied, resetFlow}: InviteUserRenderProps) => {
            // Build form schema dynamically from components
            const formSchema = useMemo(() => {
              if (!components?.length) return z.object({});
              return buildFormSchema(components);
            }, [components]);

            // Initialize react-hook-form
            const {
              control,
              formState: {errors, isValid},
              setValue,
              reset,
            } = useForm({
              resolver: zodResolver(formSchema),
              mode: 'onChange',
              defaultValues: values || {},
            });

            // Sync react-hook-form with SDK component values
            useEffect(() => {
              if (values) {
                Object.entries(values).forEach(([key, value]) => {
                  setValue(key, value);
                });
              }
            }, [values, setValue]);

            // Reset form when flow resets
            useEffect(() => {
              if (!components?.length && Object.keys(values || {}).length === 0) {
                reset({});
              }
            }, [components, values, reset]);
            // Loading
            if (isLoading && !components?.length && !isInviteGenerated) {
              return (
                <Box sx={{display: 'flex', justifyContent: 'center', p: 4}}>
                  <CircularProgress />
                </Box>
              );
            }

            // Invite generated
            if (isInviteGenerated && inviteLink) {
              return (
                <Box>
                  <Alert severity="success" sx={{mb: 3}}>
                    <AlertTitle>{t('users:inviteLinkGenerated', 'Invite Link Generated!')}</AlertTitle>
                    {t('users:inviteLinkDescription', 'Share this link with the user to complete their registration.')}
                  </Alert>
                  <Box sx={{mb: 3}}>
                    <Typography variant="body2" sx={{mb: 1}}>
                      {t('users:inviteLink', 'Invite Link')}
                    </Typography>
                    <Box sx={{display: 'flex', gap: 1}}>
                      <TextField
                        fullWidth
                        value={inviteLink}
                        InputProps={{readOnly: true}}
                        size="small"
                        sx={{'& .MuiInputBase-root': {backgroundColor: 'background.default', fontFamily: 'monospace', fontSize: '0.85rem'}}}
                      />
                      <IconButton
                        onClick={() => {
                          copyInviteLink?.();
                          handleCopy();
                        }}
                        color={copied || inviteLinkCopied ? 'success' : 'primary'}
                        sx={{flexShrink: 0}}
                      >
                        {copied || inviteLinkCopied ? <Check size={18} /> : <Copy size={18} />}
                      </IconButton>
                    </Box>
                  </Box>
                  <Box sx={{display: 'flex', gap: 2, justifyContent: 'flex-end'}}>
                    <Button variant="outlined" onClick={handleClose}>
                      {t('common:actions.close', 'Close')}
                    </Button>
                    <Button
                      variant="contained"
                      onClick={() => {
                        setActiveStep(0);
                        resetFlow();
                      }}
                    >
                      {t('users:inviteAnother', 'Invite Another User')}
                    </Button>
                  </Box>
                </Box>
              );
            }

            // Error without components
            if (error && !components?.length) {
              return (
                <Box>
                  <Alert severity="error" sx={{mb: 2}}>
                    <AlertTitle>{t('users:errors.failed.title', 'Error')}</AlertTitle>
                    {error.message ?? t('users:errors.failed.description', 'An error occurred.')}
                  </Alert>
                  <Box sx={{display: 'flex', justifyContent: 'flex-end'}}>
                    <Button variant="outlined" onClick={handleClose}>
                      {t('common:actions.close', 'Close')}
                    </Button>
                  </Box>
                </Box>
              );
            }

            // Loading components
            if (!components?.length) {
              return (
                <Box sx={{display: 'flex', justifyContent: 'center', p: 4}}>
                  <CircularProgress />
                </Box>
              );
            }

            return (
              <>
                {error && (
                  <Alert severity="error" sx={{mb: 2}}>
                    <AlertTitle>{t('users:errors.failed.title', 'Error')}</AlertTitle>
                    {error.message ?? t('users:errors.failed.description', 'An error occurred.')}
                  </Alert>
                )}
                <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
                  {components.map((component: EmbeddedFlowComponent, index: number) => {
                    // TEXT - skip main headings in dialog
                    if (component.type === EmbeddedFlowComponentType.Text) {
                      if (component.variant === 'HEADING_1' || component.variant === 'HEADING_2') return null;
                      return (
                        <Typography key={component.id ?? index} variant={mapEmbeddedFlowTextVariant(component.variant)} sx={{mb: 1}}>
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

                      const isButtonDisabled = isLoading || !isValid;

                      return (
                        <Box
                          key={component.id ?? index}
                          component="form"
                          onSubmit={(e) => {
                            e.preventDefault();
                            if (!isButtonDisabled) {
                              handleSubmit(submitAction, values).catch(() => {});
                            }
                          }}
                          noValidate
                          sx={{display: 'flex', flexDirection: 'column', width: '100%', gap: 2}}
                        >
                          {blockComponents.map((subComponent, compIndex) => {
                            // Form fields using react-hook-form Controller
                            const field = renderFormField(subComponent, compIndex, control, errors, isLoading, handleInputChange);
                            if (field) return field;

                            // Submit button
                            if (subComponent.type === EmbeddedFlowComponentType.Action && subComponent.eventType === EmbeddedFlowEventType.Submit) {
                              return (
                                <Box key={subComponent.id ?? compIndex} sx={{display: 'flex', gap: 2, justifyContent: 'flex-end', mt: 2}}>
                                  <Button variant="outlined" onClick={handleClose} disabled={isLoading}>
                                    {t('common:actions.cancel', 'Cancel')}
                                  </Button>
                                  <Button type="submit" variant={subComponent.variant === 'PRIMARY' ? 'contained' : 'outlined'} disabled={isButtonDisabled}>
                                    {isLoading ? <CircularProgress size={20} color="inherit" /> : t(resolve(subComponent.label)!)}
                                  </Button>
                                </Box>
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
              </>
            );
          }}
        </InviteUser>
      </DialogContent>
    </Dialog>
  );
}
