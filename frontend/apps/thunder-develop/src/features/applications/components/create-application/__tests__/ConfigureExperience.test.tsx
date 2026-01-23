/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ConfigureExperience from '../ConfigureExperience';
import {ApplicationCreateFlowSignInApproach} from '../../../models/application-create-flow';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('ConfigureExperience', () => {
  const mockOnApproachChange = vi.fn();
  const mockOnReadyChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the component with both approach options', () => {
      render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      );

      expect(screen.getByText('applications:onboarding.configure.approach.inbuilt.title')).toBeInTheDocument();
      expect(screen.getByText('applications:onboarding.configure.approach.native.title')).toBeInTheDocument();
    });

    it('should select INBUILT approach by default', () => {
      render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      );

      const inbuiltRadio = screen.getAllByRole('radio')[0];
      expect(inbuiltRadio).toBeChecked();
    });

    it('should select EMBEDDED approach when prop is set', () => {
      render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.EMBEDDED}
          onApproachChange={mockOnApproachChange}
        />,
      );

      const embeddedRadio = screen.getAllByRole('radio')[1];
      expect(embeddedRadio).toBeChecked();
    });
  });

  describe('User Interactions', () => {
    it('should call onApproachChange when INBUILT is clicked', async () => {
      const user = userEvent.setup();
      render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.EMBEDDED}
          onApproachChange={mockOnApproachChange}
        />,
      );

      const inbuiltRadio = screen.getAllByRole('radio')[0];
      await user.click(inbuiltRadio);

      expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.INBUILT);
    });

    it('should call onApproachChange when EMBEDDED is clicked', async () => {
      const user = userEvent.setup();
      render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      );

      const embeddedRadio = screen.getAllByRole('radio')[1];
      await user.click(embeddedRadio);

      expect(mockOnApproachChange).toHaveBeenCalledWith(ApplicationCreateFlowSignInApproach.EMBEDDED);
    });
  });

  describe('Ready State', () => {
    it('should call onReadyChange with true on mount', () => {
      render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          onReadyChange={mockOnReadyChange}
        />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });
  });

  describe('User Types Selection', () => {
    const mockUserTypes = [
      {id: '1', name: 'Internal', ouId: 'INTERNAL', allowSelfRegistration: true},
      {id: '2', name: 'External', ouId: 'EXTERNAL', allowSelfRegistration: false},
    ];
    const mockOnUserTypesChange = vi.fn();

    it('should render user types selection when userTypes prop is provided', () => {
      render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
          userTypes={mockUserTypes}
          onUserTypesChange={mockOnUserTypesChange}
        />,
      );

      expect(
        screen.getByText('applications:onboarding.configure.experience.access.userTypes.title'),
      ).toBeInTheDocument();
    });

    it('should not render user types selection when userTypes prop is undefined', () => {
      render(
        <ConfigureExperience
          selectedApproach={ApplicationCreateFlowSignInApproach.INBUILT}
          onApproachChange={mockOnApproachChange}
        />,
      );

      expect(
        screen.queryByText('applications:onboarding.configure.experience.access.userTypes.title'),
      ).not.toBeInTheDocument();
    });
  });
});
