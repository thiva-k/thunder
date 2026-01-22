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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen} from '@testing-library/react';
import type {Resource} from '@/features/flows/models/resources';
import ButtonExtendedProperties from '../ButtonExtendedProperties';

// Mock dependencies
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('../ButtonExtendedProperties.scss', () => ({}));

vi.mock('@/features/flows/hooks/useFlowBuilderCore', () => ({
  default: () => ({
    lastInteractedResource: {action: {type: 'SUBMIT'}},
    setLastInteractedResource: vi.fn(),
  }),
}));

// Mock with error notification
vi.mock('@/features/flows/hooks/useValidationStatus', () => ({
  default: () => ({
    selectedNotification: {
      hasResourceFieldNotification: vi.fn().mockReturnValue(true),
      getResourceFieldNotification: vi.fn().mockReturnValue('Action is required'),
    },
  }),
}));

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual('@wso2/oxygen-ui');
  return {
    ...actual,
    useColorScheme: () => ({mode: 'light', systemMode: 'light'}),
  };
});

vi.mock('@/features/login-flow/api/useGetLoginFlowBuilderActions', () => ({
  default: () => ({
    data: [
      {
        id: 'action-group-1',
        type: 'ACTION_GROUP',
        display: {label: 'Primary Actions'},
        types: [
          {
            id: 'submit-action',
            type: 'ACTION',
            display: {label: 'Submit', image: '/submit.png'},
            action: {type: 'SUBMIT'},
          },
        ],
      },
    ],
    isLoading: false,
  }),
}));

vi.mock('@/features/flows/utils/resolveStaticResourcePath', () => ({
  default: (path: string) => path,
}));

describe('ButtonExtendedProperties - Error Handling', () => {
  const mockOnChange = vi.fn();

  const createMockResource = (overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'button-1',
      type: 'ACTION',
      category: 'ACTION',
      resourceType: 'ELEMENT',
      action: {type: 'SUBMIT'},
      ...overrides,
    }) as Resource;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should display error message when validation notification exists', () => {
    const resource = createMockResource();

    render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

    expect(screen.getByText('Action is required')).toBeInTheDocument();
  });

  it('should apply error class to card when there is an error', () => {
    const resource = createMockResource();

    const {container} = render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

    const errorCards = container.querySelectorAll('.error');
    expect(errorCards.length).toBeGreaterThan(0);
  });
});
