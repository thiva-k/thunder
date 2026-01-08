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

/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-return, @typescript-eslint/no-unsafe-assignment, react/button-has-type, react/require-default-props */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import type {ReactNode} from 'react';
import type {Node, Edge} from '@xyflow/react';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import DecoratedVisualFlow from '../DecoratedVisualFlow';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../../context/FlowBuilderCoreContext';
import {EdgeStyleTypes} from '../../../models/steps';
import {PreviewScreenType} from '../../../models/custom-text-preference';
import {ElementTypes} from '../../../models/elements';
import type {Base} from '../../../models/base';
import type {Resources} from '../../../models/resources';

// Mock hooks
vi.mock('../../../hooks/useFlowBuilderCore', () => ({
  default: () => ({
    isResourcePanelOpen: true,
    isResourcePropertiesPanelOpen: false,
    isFlowMetadataLoading: false,
    metadata: undefined,
    onResourceDropOnCanvas: vi.fn(),
  }),
}));

vi.mock('../../../hooks/useComponentDelete', () => ({
  default: () => ({
    deleteComponent: vi.fn(),
  }),
}));

vi.mock('../../../hooks/useResourceAdd', () => ({
  default: () => vi.fn(),
}));

vi.mock('../../../hooks/useGenerateStepElement', () => ({
  default: () => ({
    generateStepElement: vi.fn(),
  }),
}));

vi.mock('../../../hooks/useDeleteExecutionResource', () => ({
  default: () => {},
}));

vi.mock('../../../hooks/useStaticContentField', () => ({
  default: () => {},
}));

vi.mock('../../../hooks/useConfirmPasswordField', () => ({
  default: () => {},
}));

vi.mock('../../../hooks/useVisualFlowHandlers', () => ({
  default: () => ({
    handleConnect: vi.fn(),
    handleNodesDelete: vi.fn(),
    handleEdgesDelete: vi.fn(),
  }),
}));

vi.mock('../../../hooks/useDragDropHandlers', () => ({
  default: () => ({
    addCanvasNode: vi.fn(),
    addToView: vi.fn(),
    addToForm: vi.fn(),
    addToViewAtIndex: vi.fn(),
    addToFormAtIndex: vi.fn(),
  }),
}));

vi.mock('../../../hooks/useContainerDialogConfirm', () => ({
  default: () => vi.fn(),
}));

vi.mock('../../../utils/applyAutoLayout', () => ({
  default: vi.fn().mockResolvedValue([]),
}));

vi.mock('../../../utils/resolveCollisions', () => ({
  resolveCollisions: vi.fn((nodes) => nodes),
}));

vi.mock('../../../utils/computeExecutorConnections', () => ({
  default: vi.fn(() => []),
}));

vi.mock('@/features/integrations/api/useIdentityProviders', () => ({
  default: () => ({data: []}),
}));

vi.mock('@/features/notification-senders/api/useNotificationSenders', () => ({
  default: () => ({data: []}),
}));

// Use vi.hoisted for mocks that need to be referenced in vi.mock
const {mockToObject, mockGetNodes, mockGetEdges, mockUpdateNodeData, mockFitView, mockUpdateNodeInternals} =
  vi.hoisted(() => ({
    mockToObject: vi.fn(() => ({viewport: {x: 0, y: 0, zoom: 1}})),
    mockGetNodes: vi.fn((): Node[] => []),
    mockGetEdges: vi.fn((): Edge[] => []),
    mockUpdateNodeData: vi.fn(),
    mockFitView: vi.fn().mockResolvedValue(undefined),
    mockUpdateNodeInternals: vi.fn(),
  }));

vi.mock('@xyflow/react', () => ({
  useReactFlow: () => ({
    toObject: mockToObject,
    getNodes: mockGetNodes,
    getEdges: mockGetEdges,
    updateNodeData: mockUpdateNodeData,
    fitView: mockFitView,
  }),
  useUpdateNodeInternals: () => mockUpdateNodeInternals,
}));

// Mock @dnd-kit/react
vi.mock('@dnd-kit/react', () => ({
  DragDropProvider: ({children, onDragEnd, onDragOver}: any) => (
    <div data-testid="drag-drop-provider" data-ondragend={!!onDragEnd} data-ondragover={!!onDragOver}>
      {children}
    </div>
  ),
}));

// Mock @dnd-kit/helpers
vi.mock('@dnd-kit/helpers', () => ({
  move: vi.fn((items) => items),
}));

// Mock @dnd-kit/abstract
vi.mock('@dnd-kit/abstract', () => ({
  CollisionPriority: {
    Low: 'low',
    High: 'high',
  },
}));

// Mock @wso2/oxygen-ui
vi.mock('@wso2/oxygen-ui', () => ({
  Box: ({children, className}: any) => (
    <div data-testid="box-component" className={className}>
      {children}
    </div>
  ),
}));

// Mock classnames
vi.mock('classnames', () => ({
  default: (...args: any[]) => args.filter(Boolean).join(' '),
}));

// Mock child components
vi.mock('../VisualFlow', () => ({
  default: ({nodes, edges, onSave, handleAutoLayout}: any) => (
    <div
      data-testid="visual-flow"
      data-nodes={JSON.stringify(nodes)}
      data-edges={JSON.stringify(edges)}
      data-has-save={!!onSave}
      data-has-auto-layout={!!handleAutoLayout}
    >
      <button data-testid="save-trigger" onClick={onSave}>
        Save
      </button>
      <button data-testid="auto-layout-trigger" onClick={handleAutoLayout}>
        Auto Layout
      </button>
    </div>
  ),
}));

vi.mock('../../dnd/Droppable', () => ({
  default: ({children, id, type}: any) => (
    <div data-testid="droppable" data-id={id} data-type={type}>
      {children}
    </div>
  ),
}));

vi.mock('../../resource-panel/ResourcePanel', () => ({
  default: ({children, open, disabled, flowTitle}: any) => (
    <div data-testid="resource-panel" data-open={open} data-disabled={disabled} data-title={flowTitle}>
      {children}
    </div>
  ),
}));

vi.mock('../../resource-property-panel/ResourcePropertyPanel', () => ({
  default: ({children, open}: any) => (
    <div data-testid="resource-property-panel" data-open={open}>
      {children}
    </div>
  ),
}));

vi.mock('../../validation-panel/ValidationPanel', () => ({
  default: () => <div data-testid="validation-panel" />,
}));

vi.mock('../FormRequiresViewDialog', () => ({
  default: ({open, scenario, onClose, onConfirm}: any) => (
    <div data-testid="form-requires-view-dialog" data-open={open} data-scenario={scenario}>
      <button data-testid="dialog-close" onClick={onClose}>
        Close
      </button>
      <button data-testid="dialog-confirm" onClick={onConfirm}>
        Confirm
      </button>
    </div>
  ),
}));

vi.mock('../../../utils/generateResourceId', () => ({
  default: (prefix: string) => `${prefix}_test123`,
}));

describe('DecoratedVisualFlow', () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  const mockBaseResource: Base = {
    id: '',
    type: '',
    category: '',
    resourceType: '',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: '',
      image: '',
      showOnResourcePanel: false,
    },
    config: {
      field: {name: '', type: ElementTypes},
      styles: {},
    },
  };

  const defaultContextValue: FlowBuilderCoreContextProps = {
    lastInteractedResource: mockBaseResource,
    lastInteractedStepId: '',
    ResourceProperties: () => null,
    resourcePropertiesPanelHeading: '',
    primaryI18nScreen: PreviewScreenType.LOGIN,
    isResourcePanelOpen: true,
    isResourcePropertiesPanelOpen: false,
    isVersionHistoryPanelOpen: false,
    ElementFactory: () => null,
    onResourceDropOnCanvas: vi.fn(),
    selectedAttributes: {},
    setLastInteractedResource: vi.fn(),
    setLastInteractedStepId: vi.fn(),
    setResourcePropertiesPanelHeading: vi.fn(),
    setIsResourcePanelOpen: vi.fn(),
    setIsOpenResourcePropertiesPanel: vi.fn(),
    registerCloseValidationPanel: vi.fn(),
    setIsVersionHistoryPanelOpen: vi.fn(),
    setSelectedAttributes: vi.fn(),
    flowCompletionConfigs: {},
    setFlowCompletionConfigs: vi.fn(),
    flowNodeTypes: {},
    flowEdgeTypes: {},
    setFlowNodeTypes: vi.fn(),
    setFlowEdgeTypes: vi.fn(),
    isVerboseMode: false,
    setIsVerboseMode: vi.fn(),
    edgeStyle: EdgeStyleTypes.SmoothStep,
    setEdgeStyle: vi.fn(),
  };

  const mockResources: Resources = {
    steps: [],
    templates: [],
    elements: [],
    widgets: [],
    executors: [],
  };

  const createWrapper = (contextValue: FlowBuilderCoreContextProps = defaultContextValue) => {
    function Wrapper({children}: {children: ReactNode}) {
      return (
        <QueryClientProvider client={queryClient}>
          <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>
        </QueryClientProvider>
      );
    }
    return Wrapper;
  };

  const defaultProps = {
    resources: mockResources,
    nodes: [] as Node[],
    edges: [] as Edge[],
    setNodes: vi.fn(),
    setEdges: vi.fn(),
    onNodesChange: vi.fn(),
    onEdgesChange: vi.fn(),
    mutateComponents: vi.fn((components) => components),
    onTemplateLoad: vi.fn(() => [[], []] as [Node[], Edge[]]),
    onWidgetLoad: vi.fn(() => [[], [], null, null] as [Node[], Edge[], null, null]),
    onStepLoad: vi.fn((step) => step),
    onResourceAdd: vi.fn(),
    flowTitle: 'Test Flow',
    flowHandle: 'test-flow',
    onFlowTitleChange: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetNodes.mockReturnValue([]);
    mockGetEdges.mockReturnValue([]);
  });

  describe('Rendering', () => {
    it('should render the component structure', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getAllByTestId('box-component').length).toBeGreaterThan(0);
    });

    it('should render DragDropProvider', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('drag-drop-provider')).toBeInTheDocument();
    });

    it('should render ResourcePanel', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('resource-panel')).toBeInTheDocument();
    });

    it('should render ResourcePropertyPanel', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('resource-property-panel')).toBeInTheDocument();
    });

    it('should render Droppable canvas', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('droppable')).toBeInTheDocument();
    });

    it('should render VisualFlow', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('visual-flow')).toBeInTheDocument();
    });

    it('should render ValidationPanel', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('validation-panel')).toBeInTheDocument();
    });

    it('should render FormRequiresViewDialog', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('form-requires-view-dialog')).toBeInTheDocument();
    });
  });

  describe('Props Passing', () => {
    it('should pass nodes to VisualFlow', () => {
      const nodes: Node[] = [{id: 'node-1', position: {x: 0, y: 0}, data: {}}];

      render(<DecoratedVisualFlow {...defaultProps} nodes={nodes} />, {
        wrapper: createWrapper(),
      });

      const visualFlow = screen.getByTestId('visual-flow');
      expect(visualFlow).toHaveAttribute('data-nodes', JSON.stringify(nodes));
    });

    it('should pass edges to VisualFlow', () => {
      const edges: Edge[] = [{id: 'edge-1', source: 'node-1', target: 'node-2'}];

      render(<DecoratedVisualFlow {...defaultProps} edges={edges} />, {
        wrapper: createWrapper(),
      });

      const visualFlow = screen.getByTestId('visual-flow');
      expect(visualFlow).toHaveAttribute('data-edges', JSON.stringify(edges));
    });

    it('should pass flow title to ResourcePanel', () => {
      render(<DecoratedVisualFlow {...defaultProps} flowTitle="My Custom Flow" />, {
        wrapper: createWrapper(),
      });

      const resourcePanel = screen.getByTestId('resource-panel');
      expect(resourcePanel).toHaveAttribute('data-title', 'My Custom Flow');
    });

    it('should indicate save handler presence', () => {
      const mockOnSave = vi.fn();

      render(<DecoratedVisualFlow {...defaultProps} onSave={mockOnSave} />, {
        wrapper: createWrapper(),
      });

      const visualFlow = screen.getByTestId('visual-flow');
      expect(visualFlow).toHaveAttribute('data-has-save', 'true');
    });

    it('should indicate auto layout handler presence', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const visualFlow = screen.getByTestId('visual-flow');
      expect(visualFlow).toHaveAttribute('data-has-auto-layout', 'true');
    });
  });

  describe('Save Functionality', () => {
    it('should call onSave with canvas data when save is triggered', () => {
      const mockOnSave = vi.fn();
      mockToObject.mockReturnValue({viewport: {x: 10, y: 20, zoom: 1.5}});
      mockGetNodes.mockReturnValue([{id: 'node-1', position: {x: 0, y: 0}, data: {}}]);
      mockGetEdges.mockReturnValue([{id: 'edge-1', source: 'node-1', target: 'node-2'}]);

      render(<DecoratedVisualFlow {...defaultProps} onSave={mockOnSave} />, {
        wrapper: createWrapper(),
      });

      const saveButton = screen.getByTestId('save-trigger');
      fireEvent.click(saveButton);

      expect(mockOnSave).toHaveBeenCalledWith({
        nodes: [{id: 'node-1', position: {x: 0, y: 0}, data: {}}],
        edges: [{id: 'edge-1', source: 'node-1', target: 'node-2'}],
        viewport: {x: 10, y: 20, zoom: 1.5},
      });
    });

    it('should not throw when onSave is not provided', () => {
      render(<DecoratedVisualFlow {...defaultProps} onSave={undefined} />, {
        wrapper: createWrapper(),
      });

      const saveButton = screen.getByTestId('save-trigger');
      expect(() => fireEvent.click(saveButton)).not.toThrow();
    });
  });

  describe('Auto Layout', () => {
    it('should handle auto layout trigger', async () => {
      const applyAutoLayout = await import('../../../utils/applyAutoLayout');
      const mockApplyAutoLayout = vi.mocked(applyAutoLayout.default);
      mockApplyAutoLayout.mockResolvedValue([{id: 'node-1', position: {x: 100, y: 100}, data: {}}]);

      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const autoLayoutButton = screen.getByTestId('auto-layout-trigger');
      fireEvent.click(autoLayoutButton);

      await waitFor(() => {
        expect(mockApplyAutoLayout).toHaveBeenCalled();
      });
    });
  });

  describe('Form Requires View Dialog', () => {
    it('should render dialog in closed state initially', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const dialog = screen.getByTestId('form-requires-view-dialog');
      expect(dialog).toHaveAttribute('data-open', 'false');
    });

    it('should close dialog when close button is clicked', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const closeButton = screen.getByTestId('dialog-close');
      fireEvent.click(closeButton);

      const dialog = screen.getByTestId('form-requires-view-dialog');
      expect(dialog).toHaveAttribute('data-open', 'false');
    });
  });

  describe('Droppable Configuration', () => {
    it('should configure droppable with correct id prefix', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const droppable = screen.getByTestId('droppable');
      expect(droppable.getAttribute('data-id')).toContain('flow-builder-canvas');
    });

    it('should configure droppable with correct type', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const droppable = screen.getByTestId('droppable');
      expect(droppable).toHaveAttribute('data-type', 'flow-builder-droppable-canvas');
    });
  });

  describe('Resource Panel State', () => {
    it('should pass open state to resource panel', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const resourcePanel = screen.getByTestId('resource-panel');
      expect(resourcePanel).toHaveAttribute('data-open', 'true');
    });

    it('should pass disabled state based on loading', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const resourcePanel = screen.getByTestId('resource-panel');
      expect(resourcePanel).toHaveAttribute('data-disabled', 'false');
    });
  });

  describe('Auto Layout on Load', () => {
    it('should not trigger auto layout when triggerAutoLayoutOnLoad is false', () => {
      const applyAutoLayout = vi.fn().mockResolvedValue([]);
      vi.doMock('../../../utils/applyAutoLayout', () => ({default: applyAutoLayout}));

      render(<DecoratedVisualFlow {...defaultProps} triggerAutoLayoutOnLoad={false} />, {
        wrapper: createWrapper(),
      });

      // Auto layout should not be called on mount when flag is false
      expect(defaultProps.setNodes).not.toHaveBeenCalled();
    });

    it('should not trigger auto layout for single node', () => {
      mockGetNodes.mockReturnValue([{id: 'node-1', position: {x: 0, y: 0}, data: {}}]);

      render(<DecoratedVisualFlow {...defaultProps} triggerAutoLayoutOnLoad />, {
        wrapper: createWrapper(),
      });

      // Single node should not trigger auto layout
      expect(defaultProps.setNodes).not.toHaveBeenCalled();
    });
  });

  describe('Edge Types', () => {
    it('should accept custom edge types', () => {
      const customEdgeTypes = {
        custom: () => <div>Custom Edge</div>,
      };

      render(<DecoratedVisualFlow {...defaultProps} edgeTypes={customEdgeTypes} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('visual-flow')).toBeInTheDocument();
    });

    it('should use default empty object for edge types', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      expect(screen.getByTestId('visual-flow')).toBeInTheDocument();
    });
  });

  describe('DragDropProvider Configuration', () => {
    it('should configure drag end handler', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const provider = screen.getByTestId('drag-drop-provider');
      expect(provider).toHaveAttribute('data-ondragend', 'true');
    });

    it('should configure drag over handler', () => {
      render(<DecoratedVisualFlow {...defaultProps} />, {
        wrapper: createWrapper(),
      });

      const provider = screen.getByTestId('drag-drop-provider');
      expect(provider).toHaveAttribute('data-ondragover', 'true');
    });
  });
});
