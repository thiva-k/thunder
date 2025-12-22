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
  Radio,
  RadioGroup,
  FormControlLabel,
  FormControl,
  Autocomplete,
  TextField,
} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {Workflow} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {type BasicFlowDefinition} from '../../../../flows/models/responses';

/**
 * Props for the FlowsListView component
 */
export interface FlowsListViewProps {
  /**
   * Available authentication flows
   */
  availableFlows: BasicFlowDefinition[];

  /**
   * Currently selected authentication flow
   */
  selectedAuthFlow: BasicFlowDefinition | null;

  /**
   * Whether the user has completed onboarding
   */
  hasCompletedOnboarding: boolean;

  /**
   * Callback when a flow is selected
   */
  onFlowSelect: (flowId: string) => void;

  /**
   * Callback when switching to toggles view
   */
  onSwitchToToggles: () => void;

  /**
   * Callback when clearing flow selection
   */
  onClearSelection: () => void;
}

/**
 * Component that renders the flows list view with either radio buttons or autocomplete
 */
export default function FlowsListView({
  availableFlows,
  selectedAuthFlow,
  onFlowSelect,
  onClearSelection,
}: FlowsListViewProps): JSX.Element {
  const {t} = useTranslation();

  return (
    <Stack direction="column" spacing={2}>
      {availableFlows.length > 10 ? (
        // Use Autocomplete for more than 10 flows
        <Stack direction="column" spacing={2}>
          <Autocomplete
            options={[
              {
                id: '',
                name: t('applications:onboarding.configure.SignInOptions.preConfiguredFlows.none'),
                activeVersion: null,
              },
              ...availableFlows,
            ]}
            getOptionLabel={(option) => option.name}
            value={
              availableFlows.find((flow) => flow.id === selectedAuthFlow?.id) ?? {
                id: '',
                name: t('applications:onboarding.configure.SignInOptions.preConfiguredFlows.none'),
                activeVersion: null,
              }
            }
            onChange={(_, newValue) => {
              if (newValue?.id) {
                onFlowSelect(newValue.id);
              } else {
                onClearSelection();
              }
            }}
            renderInput={(params) => (
              <TextField
                {...params}
                label={t('applications:onboarding.configure.SignInOptions.preConfiguredFlows.selectFlow')}
                placeholder={t('applications:onboarding.configure.SignInOptions.preConfiguredFlows.searchFlows')}
              />
            )}
            renderOption={(props, option) => (
              <Box component="li" {...props}>
                <Box sx={{display: 'flex', alignItems: 'center', gap: 1, width: '100%'}}>
                  <Workflow size={20} />
                  <Box>
                    <Typography variant="body2" fontWeight="medium">
                      {option.name}
                    </Typography>
                    {option.activeVersion && (
                      <Typography variant="caption" color="text.secondary">
                        {t('applications:onboarding.configure.SignInOptions.preConfiguredFlows.version', {
                          version: option.activeVersion,
                        })}
                      </Typography>
                    )}
                  </Box>
                </Box>
              </Box>
            )}
          />
        </Stack>
      ) : (
        // Use RadioGroup for 10 or fewer flows
        <FormControl component="fieldset">
          <RadioGroup value={selectedAuthFlow?.id ?? ''} onChange={(e) => onFlowSelect(e.target.value)}>
            <FormControlLabel
              value=""
              control={<Radio />}
              label={t('applications:onboarding.configure.SignInOptions.preConfiguredFlows.none')}
            />
            {availableFlows.map((flow: BasicFlowDefinition) => (
              <FormControlLabel
                key={flow.id}
                value={flow.id}
                control={<Radio />}
                label={
                  <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
                    <Workflow size={20} />
                    <Box>
                      <Typography variant="body2" fontWeight="medium">
                        {flow.name}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {t('applications:onboarding.configure.SignInOptions.preConfiguredFlows.version', {
                          version: flow.activeVersion,
                        })}
                      </Typography>
                    </Box>
                  </Box>
                }
              />
            ))}
          </RadioGroup>
        </FormControl>
      )}
    </Stack>
  );
}
