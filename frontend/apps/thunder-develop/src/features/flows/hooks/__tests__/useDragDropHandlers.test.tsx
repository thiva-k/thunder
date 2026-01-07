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
import {renderHook, act} from '@testing-library/react';
import type {ReactNode} from 'react';
import {ReactFlowProvider} from '@xyflow/react';
import type {Node, Edge} from '@xyflow/react';
import useDragDropHandlers, {type UseDragDropHandlersProps} from '../useDragDropHandlers';
import type {DragSourceData, DragTargetData} from '../../models/drag-drop';
import type {Resource} from '../../models/resources';
import {ResourceTypes} from '../../models/resources';
import type {Element} from '../../models/elements';
import type {Step} from '../../models/steps';
import type {Widget} from '../../models/widget';

// Import the mocked module
import autoAssignConnections from '../../utils/autoAssignConnections';

// Use vi.hoisted to define mocks that need to be referenced in vi.mock
const {mockScreenToFlowPosition, mockUpdateNodeData, mockGetNodes, mockGetEdges, mockUpdateNodeInternals} = vi.hoisted(
  () => ({
    mockScreenToFlowPosition: vi.fn((pos: {x: number; y: number}) => ({x: pos.x, y: pos.y})),
    mockUpdateNodeData: vi.fn(),
    mockGetNodes: vi.fn().mockReturnValue([]),
    mockGetEdges: vi.fn().mockReturnValue([]),
    mockUpdateNodeInternals: vi.fn(),
  }),
);

// Mock @xyflow/react
vi.mock('@xyflow/react', async () => {
  const actual = await vi.importActual('@xyflow/react');
  return {
    ...actual,
    useReactFlow: () => ({
      screenToFlowPosition: mockScreenToFlowPosition,
      updateNodeData: mockUpdateNodeData,
      getNodes: mockGetNodes,
      getEdges: mockGetEdges,
    }),
    useUpdateNodeInternals: () => mockUpdateNodeInternals,
  };
});

// Mock @dnd-kit/helpers
vi.mock('@dnd-kit/helpers', () => ({
  move: vi.fn((items: unknown[]) => items),
}));

// Mock generateResourceId
vi.mock('../../utils/generateResourceId', () => ({
  default: vi.fn((prefix: string) => `${prefix}-generated-id`),
}));

// Mock autoAssignConnections
vi.mock('../../utils/autoAssignConnections', () => ({
  default: vi.fn(),
}));

describe('useDragDropHandlers', () => {
  const mockOnStepLoad = vi.fn((step: Step) => step);
  const mockSetNodes = vi.fn();
  const mockSetEdges = vi.fn();
  const mockOnResourceDropOnCanvas = vi.fn();
  const mockGenerateStepElement = vi.fn((element: Element) => ({
    ...element,
    id: `generated-${element.type}`,
  }));
  const mockMutateComponents = vi.fn((components: Element[]) => components);
  const mockOnWidgetLoad = vi.fn((): [Node[], Edge[], Resource | null, string | null] => [[], [], null, null]);

  const defaultProps: UseDragDropHandlersProps = {
    onStepLoad: mockOnStepLoad,
    setNodes: mockSetNodes,
    setEdges: mockSetEdges,
    onResourceDropOnCanvas: mockOnResourceDropOnCanvas,
    generateStepElement: mockGenerateStepElement,
    mutateComponents: mockMutateComponents,
    onWidgetLoad: mockOnWidgetLoad,
  };

  const createWrapper = () => {
    function Wrapper({children}: {children: ReactNode}) {
      return <ReactFlowProvider>{children}</ReactFlowProvider>;
    }
    return Wrapper;
  };

  const createMockResource = (overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'resource-1',
      type: 'VIEW',
      resourceType: ResourceTypes.Step,
      category: 'STEP',
      version: '1.0.0',
      deprecated: false,
      deletable: true,
      display: {
        label: 'Test Resource',
        image: '',
        showOnResourcePanel: true,
      },
      config: {
        field: {name: '', type: {}},
        styles: {},
      },
      ...overrides,
    }) as Resource;

  const createMockDragEvent = (clientX: number, clientY: number) => ({
    nativeEvent: {
      clientX,
      clientY,
    } as MouseEvent,
  });

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetNodes.mockReturnValue([]);
    mockGetEdges.mockReturnValue([]);
  });

  describe('Hook Initialization', () => {
    it('should return stable handler functions', () => {
      const {result, rerender} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const initialHandlers = result.current;

      rerender();

      // Handlers should remain the same reference
      expect(result.current.addCanvasNode).toBe(initialHandlers.addCanvasNode);
      expect(result.current.addToView).toBe(initialHandlers.addToView);
      expect(result.current.addToForm).toBe(initialHandlers.addToForm);
      expect(result.current.addToViewAtIndex).toBe(initialHandlers.addToViewAtIndex);
      expect(result.current.addToFormAtIndex).toBe(initialHandlers.addToFormAtIndex);
    });

    it('should return all required handler functions', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      expect(result.current).toHaveProperty('addCanvasNode');
      expect(result.current).toHaveProperty('addToView');
      expect(result.current).toHaveProperty('addToForm');
      expect(result.current).toHaveProperty('addToViewAtIndex');
      expect(result.current).toHaveProperty('addToFormAtIndex');
    });
  });

  describe('addCanvasNode', () => {
    it('should add a new node to the canvas', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource(),
      };
      const targetData: DragTargetData = {};
      const event = createMockDragEvent(100, 200);

      act(() => {
        result.current.addCanvasNode(
          event as unknown as Parameters<typeof result.current.addCanvasNode>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockSetNodes).toHaveBeenCalled();
      expect(mockOnResourceDropOnCanvas).toHaveBeenCalled();
      expect(mockOnStepLoad).toHaveBeenCalled();
    });

    it('should not add node when source resource is missing', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {};
      const targetData: DragTargetData = {};
      const event = createMockDragEvent(100, 200);

      act(() => {
        result.current.addCanvasNode(
          event as unknown as Parameters<typeof result.current.addCanvasNode>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockSetNodes).not.toHaveBeenCalled();
    });

    it('should not add node when native event is missing', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource(),
      };
      const targetData: DragTargetData = {};
      const event = {nativeEvent: undefined};

      act(() => {
        result.current.addCanvasNode(
          event as unknown as Parameters<typeof result.current.addCanvasNode>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockSetNodes).not.toHaveBeenCalled();
    });

    it('should not add node when native event lacks clientX/clientY', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource(),
      };
      const targetData: DragTargetData = {};
      const event = {nativeEvent: new Event('custom')};

      act(() => {
        result.current.addCanvasNode(
          event as unknown as Parameters<typeof result.current.addCanvasNode>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockSetNodes).not.toHaveBeenCalled();
    });
  });

  describe('addToView', () => {
    it('should add element to view step', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource({resourceType: ResourceTypes.Element}),
      };
      const targetData: DragTargetData = {
        stepId: 'step-1',
        droppedOn: createMockResource(),
      };
      const event = createMockDragEvent(100, 200);

      act(() => {
        result.current.addToView(
          event as unknown as Parameters<typeof result.current.addToView>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockUpdateNodeData).toHaveBeenCalledWith('step-1', expect.any(Function));
      expect(mockGenerateStepElement).toHaveBeenCalled();
      expect(mockOnResourceDropOnCanvas).toHaveBeenCalled();
    });

    it('should handle widget drop on view', () => {
      const mockNodes: Node[] = [{id: 'node-1', position: {x: 0, y: 0}, data: {}}];
      const mockEdges: Edge[] = [];
      mockGetNodes.mockReturnValue(mockNodes);
      mockGetEdges.mockReturnValue(mockEdges);
      mockOnWidgetLoad.mockReturnValue([mockNodes, mockEdges, null, null]);

      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const widgetResource: Resource = createMockResource({
        resourceType: ResourceTypes.Widget,
        type: 'IDENTIFIER_PASSWORD',
      });

      const sourceData: DragSourceData = {
        dragged: widgetResource,
      };
      const targetData: DragTargetData = {
        stepId: 'step-1',
        droppedOn: createMockResource(),
      };
      const event = createMockDragEvent(100, 200);

      act(() => {
        result.current.addToView(
          event as unknown as Parameters<typeof result.current.addToView>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockOnWidgetLoad).toHaveBeenCalledWith(widgetResource as Widget, expect.any(Object), mockNodes, mockEdges);
      expect(mockSetNodes).toHaveBeenCalled();
      expect(mockSetEdges).toHaveBeenCalled();
    });

    it('should not add element when source resource is missing', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {};
      const targetData: DragTargetData = {
        stepId: 'step-1',
      };
      const event = createMockDragEvent(100, 200);

      act(() => {
        result.current.addToView(
          event as unknown as Parameters<typeof result.current.addToView>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });
  });

  describe('addToForm', () => {
    it('should add element to form within a step', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource({resourceType: ResourceTypes.Element}),
      };
      const targetData: DragTargetData = {
        stepId: 'step-1',
        droppedOn: createMockResource({id: 'form-1'}),
      };
      const event = createMockDragEvent(100, 200);

      act(() => {
        result.current.addToForm(
          event as unknown as Parameters<typeof result.current.addToForm>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockUpdateNodeData).toHaveBeenCalledWith('step-1', expect.any(Function));
      expect(mockGenerateStepElement).toHaveBeenCalled();
      expect(mockOnResourceDropOnCanvas).toHaveBeenCalled();
    });

    it('should not add element when target step is missing', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource({resourceType: ResourceTypes.Element}),
      };
      const targetData: DragTargetData = {
        droppedOn: createMockResource({id: 'form-1'}),
      };
      const event = createMockDragEvent(100, 200);

      act(() => {
        result.current.addToForm(
          event as unknown as Parameters<typeof result.current.addToForm>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });
  });

  describe('addToViewAtIndex', () => {
    it('should add element at specific index in view', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource({resourceType: ResourceTypes.Element}),
      };

      act(() => {
        result.current.addToViewAtIndex(sourceData, 'step-1', 'element-2');
      });

      expect(mockUpdateNodeData).toHaveBeenCalledWith('step-1', expect.any(Function));
      expect(mockGenerateStepElement).toHaveBeenCalled();
    });

    it('should handle widget drop at index', () => {
      const mockTargetNode: Node = {
        id: 'step-1',
        position: {x: 0, y: 0},
        data: {components: [{id: 'element-1'}, {id: 'element-2'}]},
      };
      mockGetNodes.mockReturnValue([mockTargetNode]);
      mockGetEdges.mockReturnValue([]);
      mockOnWidgetLoad.mockReturnValue([
        [{...mockTargetNode, data: {components: [{id: 'element-1'}, {id: 'element-2'}, {id: 'widget-button'}]}}],
        [],
        null,
        null,
      ]);

      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const widgetResource: Resource = createMockResource({
        resourceType: ResourceTypes.Widget,
        type: 'IDENTIFIER_PASSWORD',
      });

      const sourceData: DragSourceData = {
        dragged: widgetResource,
      };

      act(() => {
        result.current.addToViewAtIndex(sourceData, 'step-1', 'element-2');
      });

      expect(mockOnWidgetLoad).toHaveBeenCalled();
      expect(mockSetNodes).toHaveBeenCalled();
      expect(mockSetEdges).toHaveBeenCalled();
    });

    it('should not add when source resource is missing', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {};

      act(() => {
        result.current.addToViewAtIndex(sourceData, 'step-1', 'element-2');
      });

      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });

    it('should not add when target step is missing', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource({resourceType: ResourceTypes.Element}),
      };

      act(() => {
        result.current.addToViewAtIndex(sourceData, '', 'element-2');
      });

      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });
  });

  describe('addToFormAtIndex', () => {
    it('should add element at specific index in form', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource({resourceType: ResourceTypes.Element}),
      };

      act(() => {
        result.current.addToFormAtIndex(sourceData, 'step-1', 'form-1', 'element-2');
      });

      expect(mockUpdateNodeData).toHaveBeenCalledWith('step-1', expect.any(Function));
      expect(mockGenerateStepElement).toHaveBeenCalled();
    });

    it('should not add when source resource is missing', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {};

      act(() => {
        result.current.addToFormAtIndex(sourceData, 'step-1', 'form-1', 'element-2');
      });

      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });

    it('should not add when target step is missing', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource({resourceType: ResourceTypes.Element}),
      };

      act(() => {
        result.current.addToFormAtIndex(sourceData, '', 'form-1', 'element-2');
      });

      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });

    it('should not add when form id is missing', () => {
      const {result} = renderHook(() => useDragDropHandlers(defaultProps), {
        wrapper: createWrapper(),
      });

      const sourceData: DragSourceData = {
        dragged: createMockResource({resourceType: ResourceTypes.Element}),
      };

      act(() => {
        result.current.addToFormAtIndex(sourceData, 'step-1', '', 'element-2');
      });

      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });
  });

  describe('Metadata Handling', () => {
    it('should call autoAssignConnections when metadata has executorConnections', () => {
      const mockAutoAssignConnections = vi.mocked(autoAssignConnections);

      const propsWithMetadata: UseDragDropHandlersProps = {
        ...defaultProps,
        metadata: {
          flowType: 'LOGIN',
          supportedExecutors: [],
          connectorConfigs: {
            multiAttributeLoginEnabled: false,
            accountVerificationEnabled: false,
          },
          attributeProfile: 'default',
          attributeMetadata: [],
          executorConnections: [{executorName: 'executor1', connections: ['step-2']}],
        },
      };

      const mockNodes: Node[] = [{id: 'node-1', position: {x: 0, y: 0}, data: {}}];
      mockGetNodes.mockReturnValue(mockNodes);
      mockGetEdges.mockReturnValue([]);
      mockOnWidgetLoad.mockReturnValue([mockNodes, [], null, null]);

      const {result} = renderHook(() => useDragDropHandlers(propsWithMetadata), {
        wrapper: createWrapper(),
      });

      const widgetResource: Resource = createMockResource({
        resourceType: ResourceTypes.Widget,
        type: 'IDENTIFIER_PASSWORD',
      });

      const sourceData: DragSourceData = {
        dragged: widgetResource,
      };
      const targetData: DragTargetData = {
        stepId: 'step-1',
        droppedOn: createMockResource(),
      };
      const event = createMockDragEvent(100, 200);

      act(() => {
        result.current.addToView(
          event as unknown as Parameters<typeof result.current.addToView>[0],
          sourceData,
          targetData,
        );
      });

      expect(mockAutoAssignConnections).toHaveBeenCalledWith(
        mockNodes,
        propsWithMetadata.metadata?.executorConnections,
      );
    });
  });
});
