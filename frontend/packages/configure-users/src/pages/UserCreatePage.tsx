// Copyright 2025-2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {OrganizationUnitTreePicker} from '@thunderid/configure-organization-units';
import {useLogger} from '@thunderid/logger/react';
import {
  EmbeddedFlowComponentType,
  InviteUser,
  useThunderID,
  type EmbeddedFlowComponent,
  type InviteUserRenderProps,
} from '@thunderid/react';
import {
  Box,
  Stack,
  Typography,
  Button,
  TextField,
  IconButton,
  FormControl,
  FormLabel,
  Select,
  MenuItem,
  LinearProgress,
  AppBreadcrumbs,
} from '@wso2/oxygen-ui';
import {X} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useState, useCallback, useEffect, useRef} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import useUserRoutes from '../hooks/useUserRoutes';

const ONBOARDING_MODE_CREATE_ACTION_ID = 'action_create_user_now';

/** Typed shape for flow sub-components */
type FlowSubComponent = EmbeddedFlowComponent & {
  components?: EmbeddedFlowComponent[];
  hint?: string;
  options?: unknown[];
  placeholder?: string;
  required?: boolean;
  variant?: string;
};

function UserCreateStepContent({
  renderProps,
  onStepLabelChange,
}: {
  renderProps: InviteUserRenderProps;
  onStepLabelChange: (label: string) => void;
}): JSX.Element {
  const {resolveFlowTemplateLiterals: rawResolve} = useThunderID();
  const resolve = useCallback((text?: string) => (text ? rawResolve(text) : undefined), [rawResolve]);
  const {t} = useTranslation();
  const components = renderProps.components as EmbeddedFlowComponent[] | undefined;

  // Auto-select create action if present
  const createAction = components?.find((c) => c.id === ONBOARDING_MODE_CREATE_ACTION_ID);

  const submittedActionIdRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    if (createAction && submittedActionIdRef.current !== createAction.id) {
      submittedActionIdRef.current = createAction.id;
      renderProps.handleSubmit(createAction, renderProps.values).catch(() => undefined);
    }
  }, [createAction, renderProps]);

  // Derive step label from HEADING_1 component
  const stepLabel = components?.find(
    (c) =>
      (String(c.type) === String(EmbeddedFlowComponentType.Text) || c.type === 'TEXT') &&
      (c as FlowSubComponent).variant === 'HEADING_1',
  );

  useEffect(() => {
    if (stepLabel && typeof stepLabel.label === 'string') {
      onStepLabelChange(t(resolve(stepLabel.label) ?? stepLabel.label));
    }
  }, [stepLabel, resolve, t, onStepLabelChange]);

  const {values, additionalData, handleInputChange} = renderProps;

  const renderComponent = (component: EmbeddedFlowComponent, index: number): JSX.Element | null => {
    // Render text components
    if (String(component.type) === String(EmbeddedFlowComponentType.Text) || component.type === 'TEXT') {
      const label = typeof component.label === 'string' ? component.label : '';
      const variant = (component as FlowSubComponent).variant;

      if (variant === 'HEADING_1') {
        return (
          <Typography key={component.id ?? index} variant="h1" gutterBottom>
            {t(resolve(label) ?? label)}
          </Typography>
        );
      }

      return (
        <Typography key={component.id ?? index} variant="body1" color="text.secondary">
          {t(resolve(label) ?? label)}
        </Typography>
      );
    }

    // Render form fields
    if (
      component.type === 'EMAIL_INPUT' ||
      component.type === 'TEXT_INPUT' ||
      component.type === 'PHONE_INPUT' ||
      component.type === 'PASSWORD_INPUT'
    ) {
      const ref = component.ref;
      const label = typeof component.label === 'string' ? component.label : '';
      const placeholder = (component as FlowSubComponent).placeholder ?? '';
      const required = (component as FlowSubComponent).required ?? false;

      if (!ref) return null;

      const value = ((values as Record<string, unknown>)?.[ref] as string) ?? '';
      let inputType = 'text';

      if (component.type === 'EMAIL_INPUT') inputType = 'email';
      if (component.type === 'PASSWORD_INPUT') inputType = 'password';
      if (component.type === 'PHONE_INPUT') inputType = 'tel';

      return (
        <FormControl key={component.id ?? index} fullWidth required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(label) ?? label)}</FormLabel>
          <TextField
            id={ref}
            type={inputType}
            size="small"
            variant="outlined"
            placeholder={resolve(placeholder) ?? placeholder}
            value={value}
            required={required}
            onChange={(e) => handleInputChange(ref, e.target.value)}
          />
        </FormControl>
      );
    }

    // Render SELECT
    if (component.type === 'SELECT') {
      const ref = component.ref;
      const label = typeof component.label === 'string' ? component.label : '';
      const options = (component as FlowSubComponent).options ?? [];
      const required = (component as FlowSubComponent).required ?? false;

      if (!ref || !options.length) return null;

      const value = ((values as Record<string, unknown>)?.[ref] as string) ?? '';

      return (
        <FormControl key={component.id ?? index} fullWidth required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(label) ?? label)}</FormLabel>
          <Select
            id={ref}
            value={value}
            size="small"
            displayEmpty
            required={required}
            onChange={(e) => handleInputChange(ref, String(e.target.value))}
          >
            <MenuItem value="">{t('Select an option')}</MenuItem>
            {options.map((opt: unknown) => {
              const optValue =
                typeof opt === 'object' && opt !== null ? (opt as Record<string, unknown>)['value'] : opt;
              const optLabel =
                typeof opt === 'object' && opt !== null ? (opt as Record<string, unknown>)['label'] : opt;
              const keyValue = String(optValue);
              return (
                <MenuItem key={keyValue} value={keyValue}>
                  {String(optLabel)}
                </MenuItem>
              );
            })}
          </Select>
        </FormControl>
      );
    }

    // Render OU_SELECT
    if (component.type === 'OU_SELECT') {
      const ref = component.ref;
      const label = typeof component.label === 'string' ? component.label : '';
      const required = (component as FlowSubComponent).required ?? false;

      if (!ref) return null;

      const value = ((values as Record<string, unknown>)?.[ref] as string) ?? '';
      const rootOuId = additionalData?.['rootOuId'] as string | undefined;

      return (
        <FormControl key={component.id ?? index} fullWidth required={required}>
          <FormLabel htmlFor={ref}>{t(resolve(label) ?? label)}</FormLabel>
          <OrganizationUnitTreePicker
            value={value}
            onChange={(ouId: string) => handleInputChange(ref, ouId)}
            rootOuId={rootOuId}
          />
        </FormControl>
      );
    }

    // Render ACTION buttons (submit buttons)
    if (component.type === 'ACTION') {
      const label = typeof component.label === 'string' ? component.label : '';
      const variant = (component as FlowSubComponent).variant;

      return (
        <Button
          key={component.id ?? index}
          variant={variant === 'OUTLINED' ? 'outlined' : 'contained'}
          onClick={() => {
            renderProps.handleSubmit(component, renderProps.values).catch(() => undefined);
          }}
          fullWidth
          size="large"
        >
          {t(resolve(label) ?? label)}
        </Button>
      );
    }

    // Render nested components (BLOCK, STACK)
    if ((component as FlowSubComponent).components) {
      const nested = (component as FlowSubComponent).components!;
      return (
        <Stack key={component.id ?? index} direction="column" spacing={2}>
          {nested.map((c, idx) => renderComponent(c, idx))}
        </Stack>
      );
    }

    return null;
  };

  return (
    <Stack direction="column" spacing={4}>
      {components?.map((component: EmbeddedFlowComponent, index: number) => renderComponent(component, index))}
    </Stack>
  );
}

export default function UserCreatePage(): JSX.Element {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const routes = useUserRoutes();
  const logger = useLogger('UserCreatePage');
  const [breadcrumbs, setBreadcrumbs] = useState<string[]>([]);

  const handleClose = useCallback(() => {
    Promise.resolve(navigate(routes.list())).catch((err: unknown) => {
      logger.error('Failed to navigate to users page', {error: err});
    });
  }, [navigate, routes, logger]);

  const handleStepLabelChange = useCallback(
    (label: string) => {
      setBreadcrumbs([t('users:addUser', 'Add User'), label]);
    },
    [t],
  );

  return (
    <Box sx={{minHeight: '100vh', display: 'flex', flexDirection: 'column'}}>
      <LinearProgress variant="determinate" value={0} sx={{height: 6}} />

      <Box sx={{flex: 1, display: 'flex', flexDirection: 'column'}}>
        {/* Header with close button and breadcrumb */}
        <Box sx={{p: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
          <Stack direction="row" alignItems="center" spacing={2}>
            <IconButton
              aria-label={t('common:actions.close', 'Close')}
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
              items={breadcrumbs.map((label, idx) => ({
                key: `breadcrumb-${idx}`,
                label,
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
              pt: 8,
              pb: 8,
              px: 20,
              mx: 'auto',
              alignItems: 'flex-start',
              justifyContent: 'flex-start',
            }}
          >
            <Box
              sx={{
                width: '100%',
                maxWidth: 800,
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <InviteUser>
                {(renderProps: InviteUserRenderProps) => (
                  <UserCreateStepContent renderProps={renderProps} onStepLabelChange={handleStepLabelChange} />
                )}
              </InviteUser>
            </Box>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}
