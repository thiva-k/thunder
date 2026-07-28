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

import {render, screen} from '@testing-library/react';
import {describe, it, expect, vi} from 'vitest';
import FlowBuilderPage from '../FlowBuilderPage';

// Mock ReactFlowProvider
vi.mock('@xyflow/react', () => ({
  ReactFlowProvider: ({children}: {children: React.ReactNode}) => (
    <div data-testid="react-flow-provider">{children}</div>
  ),
}));

// Mock FlowBuilder
vi.mock('../../components/FlowBuilder', () => ({
  default: () => <div data-testid="flow-builder">Flow Builder</div>,
}));

// Mock FlowBuilderProvider
vi.mock('../../context/FlowBuilderProvider', () => ({
  default: ({children}: {children: React.ReactNode}) => <div data-testid="flow-builder-provider">{children}</div>,
}));

describe('FlowBuilderPage', () => {
  describe('Rendering', () => {
    it('should render the FlowBuilderProvider', () => {
      render(<FlowBuilderPage />);

      expect(screen.getByTestId('flow-builder-provider')).toBeInTheDocument();
    });

    it('should render the ReactFlowProvider inside FlowBuilderProvider', () => {
      render(<FlowBuilderPage />);

      const provider = screen.getByTestId('flow-builder-provider');
      const reactFlowProvider = screen.getByTestId('react-flow-provider');

      expect(provider).toContainElement(reactFlowProvider);
    });

    it('should render the FlowBuilder inside ReactFlowProvider', () => {
      render(<FlowBuilderPage />);

      const reactFlowProvider = screen.getByTestId('react-flow-provider');
      const flowBuilder = screen.getByTestId('flow-builder');

      expect(reactFlowProvider).toContainElement(flowBuilder);
    });

    it('should render components in correct nesting order', () => {
      render(<FlowBuilderPage />);

      const provider = screen.getByTestId('flow-builder-provider');
      const reactFlowProvider = screen.getByTestId('react-flow-provider');
      const flowBuilder = screen.getByTestId('flow-builder');

      // Verify proper nesting: FlowBuilderProvider > ReactFlowProvider > FlowBuilder
      expect(provider).toContainElement(reactFlowProvider);
      expect(reactFlowProvider).toContainElement(flowBuilder);
    });
  });

  describe('Component Integration', () => {
    it('should render FlowBuilder content', () => {
      render(<FlowBuilderPage />);

      expect(screen.getByText('Flow Builder')).toBeInTheDocument();
    });
  });

  describe('Page Structure', () => {
    it('should have all required provider wrappers', () => {
      render(<FlowBuilderPage />);

      expect(screen.getByTestId('flow-builder-provider')).toBeInTheDocument();
      expect(screen.getByTestId('react-flow-provider')).toBeInTheDocument();
      expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
    });

    it('should render without crashing', () => {
      expect(() => render(<FlowBuilderPage />)).not.toThrow();
    });
  });
});
