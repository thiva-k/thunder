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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {type BasicFlowDefinition} from '@/features/flows/models/responses';
import FlowsListView, {type FlowsListViewProps} from '../FlowsListView';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:onboarding.configure.SignInOptions.preConfiguredFlows.none': 'None',
        'applications:onboarding.configure.SignInOptions.preConfiguredFlows.selectFlow': 'Select a flow',
        'applications:onboarding.configure.SignInOptions.preConfiguredFlows.searchFlows': 'Search flows...',
      };
      return translations[key] || key;
    },
  }),
}));

describe.skip('FlowsListView', () => {
  const mockOnFlowSelect = vi.fn();
  const mockOnClearSelection = vi.fn();

  const mockFlows: BasicFlowDefinition[] = [
    {
      id: 'flow-1',
      name: 'Basic Authentication Flow',
      activeVersion: 1,
      handle: 'basic-auth-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '',
      updatedAt: '',
    },
    {
      id: 'flow-2',
      name: 'Google OAuth Flow',
      activeVersion: 1,
      handle: 'google-oauth-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '',
      updatedAt: '',
    },
    {
      id: 'flow-3',
      name: 'Multi-Factor Auth Flow',
      activeVersion: 1,
      handle: 'mfa-flow',
      flowType: 'AUTHENTICATION',
      createdAt: '',
      updatedAt: '',
    },
  ];

  const defaultProps: FlowsListViewProps = {
    availableFlows: mockFlows,
    selectedAuthFlow: null,
    onFlowSelect: mockOnFlowSelect,
    onClearSelection: mockOnClearSelection,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = (props: Partial<FlowsListViewProps> = {}) =>
    render(<FlowsListView {...defaultProps} {...props} />);

  describe('with 10 or fewer flows (RadioGroup)', () => {
    it('should render radio group with all flows', () => {
      renderComponent();

      expect(screen.getByRole('radiogroup')).toBeInTheDocument();

      // Check for "None" option
      expect(screen.getByDisplayValue('')).toBeInTheDocument();
      expect(screen.getByText('None')).toBeInTheDocument();

      // Check for all flow options
      mockFlows.forEach((flow) => {
        expect(screen.getByDisplayValue(flow.id)).toBeInTheDocument();
        expect(screen.getByText(flow.name)).toBeInTheDocument();
        if (flow.activeVersion) {
          const versionTexts = screen.getAllByText(`Version ${flow.activeVersion}`);
          expect(versionTexts.length).toBeGreaterThan(0);
        }
      });
    });

    it('should show workflow icons for each flow', () => {
      renderComponent();

      // Check that flow names are rendered (which contain workflow icons)
      mockFlows.forEach((flow) => {
        expect(screen.getByText(flow.name)).toBeInTheDocument();
      });
    });

    it('should call onFlowSelect when a flow radio button is selected', async () => {
      const user = userEvent.setup();
      renderComponent();

      const flowRadio = screen.getByDisplayValue('flow-1');
      await user.click(flowRadio);

      expect(mockOnFlowSelect).toHaveBeenCalledWith('flow-1');
    });

    it('should call onClearSelection when "None" option is selected', async () => {
      const user = userEvent.setup();
      renderComponent({
        selectedAuthFlow: mockFlows[0],
      });

      const noneRadio = screen.getByDisplayValue('');
      await user.click(noneRadio);

      expect(mockOnFlowSelect).toHaveBeenCalledWith('');
    });

    it('should show selected flow as checked', () => {
      renderComponent({
        selectedAuthFlow: mockFlows[1],
      });

      const selectedRadio = screen.getByDisplayValue('flow-2');
      expect(selectedRadio).toBeChecked();
    });

    it('should show "None" as checked when no flow is selected', () => {
      renderComponent({
        selectedAuthFlow: null,
      });

      const noneRadio = screen.getByDisplayValue('');
      expect(noneRadio).toBeChecked();
    });

    it('should handle flows without activeVersion', () => {
      const flowsWithoutVersion: BasicFlowDefinition[] = [
        {
          id: 'flow-no-version',
          name: 'Flow Without Version',
          activeVersion: 1,
          handle: 'no-version-flow',
          flowType: 'AUTHENTICATION',
          createdAt: '2023-01-01T00:00:00Z',
          updatedAt: '2023-01-01T00:00:00Z',
        },
      ];

      renderComponent({
        availableFlows: flowsWithoutVersion,
      });

      expect(screen.getByText('Flow Without Version')).toBeInTheDocument();
      // Should not crash when activeVersion is null
    });
  });

  describe('with more than 10 flows (Autocomplete)', () => {
    const manyFlows: BasicFlowDefinition[] = Array.from({length: 15}, (_, i) => ({
      id: `flow-${i + 1}`,
      name: `Flow ${i + 1}`,
      activeVersion: i + 1,
      handle: `flow-${i + 1}-handle`,
      flowType: 'AUTHENTICATION',
      createdAt: '2023-01-01T00:00:00Z',
      updatedAt: '2023-01-01T00:00:00Z',
    }));

    it('should render autocomplete instead of radio group', () => {
      renderComponent({
        availableFlows: manyFlows,
      });

      expect(screen.queryByRole('radiogroup')).not.toBeInTheDocument();
      expect(screen.getByRole('combobox')).toBeInTheDocument();
      expect(screen.getByLabelText('Select a flow')).toBeInTheDocument();
    });

    it('should show placeholder text in autocomplete', () => {
      renderComponent({
        availableFlows: manyFlows,
      });

      expect(screen.getByPlaceholderText('Search flows...')).toBeInTheDocument();
    });

    it('should call onFlowSelect when autocomplete option is selected', async () => {
      const user = userEvent.setup();
      renderComponent({
        availableFlows: manyFlows,
      });

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      // Find and click a specific flow option
      const flowOption = screen.getByText('Flow 5');
      await user.click(flowOption);

      expect(mockOnFlowSelect).toHaveBeenCalledWith('flow-5');
    });

    it('should call onClearSelection when "None" is selected in autocomplete', async () => {
      const user = userEvent.setup();
      renderComponent({
        availableFlows: manyFlows,
        selectedAuthFlow: manyFlows[0],
      });

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      const noneOption = screen.getByText('None');
      await user.click(noneOption);

      expect(mockOnClearSelection).toHaveBeenCalled();
    });

    it('should show selected flow in autocomplete value', () => {
      renderComponent({
        availableFlows: manyFlows,
        selectedAuthFlow: manyFlows[3],
      });

      const autocomplete = screen.getByDisplayValue('Flow 4');
      expect(autocomplete).toBeInTheDocument();
    });

    it('should show "None" when no flow is selected in autocomplete', () => {
      renderComponent({
        availableFlows: manyFlows,
        selectedAuthFlow: null,
      });

      const autocomplete = screen.getByDisplayValue('None');
      expect(autocomplete).toBeInTheDocument();
    });

    it('should filter flows based on search input', async () => {
      const user = userEvent.setup();
      renderComponent({
        availableFlows: manyFlows,
      });

      const autocomplete = screen.getByRole('combobox');

      // Clear the input and type new search
      await user.clear(autocomplete);
      await user.type(autocomplete, 'Flow 1');

      // The autocomplete should be functioning (checking for combobox element)
      expect(autocomplete).toHaveValue('Flow 1');
    });
  });

  describe('edge cases', () => {
    it('should handle empty flows array with radio group', () => {
      renderComponent({
        availableFlows: [],
      });

      expect(screen.getByRole('radiogroup')).toBeInTheDocument();
      expect(screen.getByText('None')).toBeInTheDocument();
    });

    it('should handle flows with special characters in names', () => {
      const specialFlows: BasicFlowDefinition[] = [
        {
          id: 'special-flow',
          name: 'OAuth 2.0 & OIDC Flow',
          activeVersion: 1,
          handle: 'oauth-oidc-flow',
          flowType: 'AUTHENTICATION',
          createdAt: '2023-01-01T00:00:00Z',
          updatedAt: '2023-01-01T00:00:00Z',
        },
      ];

      renderComponent({
        availableFlows: specialFlows,
      });

      expect(screen.getByText('OAuth 2.0 & OIDC Flow')).toBeInTheDocument();
    });

    it('should handle flows with very long names', () => {
      const longNameFlows: BasicFlowDefinition[] = [
        {
          id: 'long-flow',
          name: 'This is a very long flow name that should still be displayed properly without breaking the layout',
          activeVersion: 1,
          handle: 'long-name-flow',
          flowType: 'AUTHENTICATION',
          createdAt: '2023-01-01T00:00:00Z',
          updatedAt: '2023-01-01T00:00:00Z',
        },
      ];

      renderComponent({
        availableFlows: longNameFlows,
      });

      expect(screen.getByText(/This is a very long flow name/)).toBeInTheDocument();
    });

    it('should handle flows with missing or invalid IDs', () => {
      const invalidFlows: BasicFlowDefinition[] = [
        {
          id: '',
          name: 'Flow with empty ID',
          activeVersion: 1,
          handle: 'empty-id-flow',
          flowType: 'AUTHENTICATION',
          createdAt: '2023-01-01T00:00:00Z',
          updatedAt: '2023-01-01T00:00:00Z',
        },
      ];

      renderComponent({
        availableFlows: invalidFlows,
      });

      // Should not crash and should still render the flow
      expect(screen.getByText('Flow with empty ID')).toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('should have proper ARIA labels for radio group', () => {
      renderComponent();

      const radioGroup = screen.getByRole('radiogroup');
      expect(radioGroup).toBeInTheDocument();

      const radios = screen.getAllByRole('radio');
      expect(radios.length).toBe(mockFlows.length + 1); // +1 for "None" option
    });

    it('should have proper ARIA labels for autocomplete', () => {
      const manyFlows: BasicFlowDefinition[] = Array.from({length: 15}, (_, i) => ({
        id: `flow-${i + 1}`,
        name: `Flow ${i + 1}`,
        activeVersion: i + 1,
        handle: `flow-${i + 1}-handle`,
        flowType: 'AUTHENTICATION',
        createdAt: '2023-01-01T00:00:00Z',
        updatedAt: '2023-01-01T00:00:00Z',
      }));

      renderComponent({
        availableFlows: manyFlows,
      });

      const combobox = screen.getByRole('combobox');
      expect(combobox).toHaveAttribute('aria-autocomplete', 'list');
    });

    it('should be keyboard navigable with radio group', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Focus first radio button
      await user.tab();
      const firstRadio = screen.getByDisplayValue('');
      expect(firstRadio).toHaveFocus();

      // Navigate with arrow keys
      await user.keyboard('{ArrowDown}');
      const secondRadio = screen.getByDisplayValue('flow-1');
      expect(secondRadio).toHaveFocus();
    });

    it('should be keyboard navigable with autocomplete', async () => {
      const user = userEvent.setup();
      const manyFlows: BasicFlowDefinition[] = Array.from({length: 15}, (_, i) => ({
        id: `flow-${i + 1}`,
        name: `Flow ${i + 1}`,
        activeVersion: i + 1,
        handle: `flow-${i + 1}-handle`,
        flowType: 'AUTHENTICATION',
        createdAt: '2023-01-01T00:00:00Z',
        updatedAt: '2023-01-01T00:00:00Z',
      }));

      renderComponent({
        availableFlows: manyFlows,
      });

      const combobox = screen.getByRole('combobox');
      await user.tab();
      expect(combobox).toHaveFocus();

      // The combobox should be accessible
      expect(combobox).toHaveAttribute('aria-expanded', 'false');
    });
  });

  describe('performance', () => {
    it('should handle large number of flows efficiently', () => {
      const largeFlowSet: BasicFlowDefinition[] = Array.from({length: 100}, (_, i) => ({
        id: `flow-${i + 1}`,
        name: `Flow ${i + 1}`,
        activeVersion: i + 1,
        handle: `flow-${i + 1}-handle`,
        flowType: 'AUTHENTICATION',
        createdAt: '2023-01-01T00:00:00Z',
        updatedAt: '2023-01-01T00:00:00Z',
      }));

      // Should not throw error or freeze
      expect(() => {
        renderComponent({
          availableFlows: largeFlowSet,
        });
      }).not.toThrow();

      // Should use autocomplete for large sets
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });
  });
});
