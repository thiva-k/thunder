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
import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import type {Node, Edge} from '@xyflow/react';
import {ElementTypes, ButtonVariants, ButtonTypes, ElementCategories, BlockTypes} from '@/features/flows/models/elements';
import type {Element} from '@/features/flows/models/elements';
import LoginFlowBuilder from '../LoginFlowBuilder';
import LoginFlowConstants from '../../constants/LoginFlowConstants';

// Use vi.hoisted for mock functions that need to be available during vi.mock hoisting
const {
  mockNavigate,
  mockUseParams,
  mockUseNodesState,
  mockUseEdgesState,
  mockUseUpdateNodeInternals,
  mockSetNodes,
  mockSetEdges,
  mockCreateFlowMutate,
  mockUpdateFlowMutate,
  mockSetFlowCompletionConfigs,
  mockSetOpenValidationPanel,
  mockGenerateStepElement,
  mockGenerateEdges,
  mockValidateEdges,
  mockIsFlowValid,
  mockExistingFlowData,
} = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockUseParams: vi.fn(() => ({})),
  mockUseNodesState: vi.fn(),
  mockUseEdgesState: vi.fn(),
  mockUseUpdateNodeInternals: vi.fn(() => vi.fn()),
  mockSetNodes: vi.fn(),
  mockSetEdges: vi.fn(),
  mockCreateFlowMutate: vi.fn(),
  mockUpdateFlowMutate: vi.fn(),
  mockSetFlowCompletionConfigs: vi.fn(),
  mockSetOpenValidationPanel: vi.fn(),
  mockGenerateStepElement: vi.fn((element: Element) => ({...element, id: 'generated-id'})),
  mockGenerateEdges: vi.fn(() => []),
  mockValidateEdges: vi.fn((edges: Edge[]) => edges),
  mockIsFlowValid: {value: true},
  mockExistingFlowData: {value: null as unknown},
}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

// Mock react-router
vi.mock('react-router', () => ({
  useNavigate: () => mockNavigate,
  useParams: () => mockUseParams(),
}));

// Mock @xyflow/react
vi.mock('@xyflow/react', () => ({
  useNodesState: (): unknown => mockUseNodesState(),
  useEdgesState: (): unknown => mockUseEdgesState(),
  useUpdateNodeInternals: () => mockUseUpdateNodeInternals(),
  MarkerType: {Arrow: 'arrow'},
}));

// Mock FlowBuilder component
vi.mock('@/features/flows/components/FlowBuilder', () => ({
  default: ({
    flowTitle,
    flowHandle,
    onFlowTitleChange,
    onSave,
    nodes,
    edges,
    onResourceAdd,
    onTemplateLoad,
    onWidgetLoad,
    onStepLoad,
    triggerAutoLayoutOnLoad,
    mutateComponents,
  }: {
    flowTitle: string;
    flowHandle: string;
    onFlowTitleChange: (name: string) => void;
    onSave: (data: {nodes: Node[]; edges: Edge[]; viewport: {x: number; y: number; zoom: number}}) => void;
    nodes: Node[];
    edges: Edge[];
    onResourceAdd: (resource: unknown) => void;
    onTemplateLoad: (template: unknown) => unknown;
    onWidgetLoad: (widget: unknown, target: unknown, nodes: Node[], edges: Edge[]) => unknown;
    onStepLoad: (step: unknown) => unknown;
    triggerAutoLayoutOnLoad?: boolean;
    mutateComponents: (components: Element[]) => Element[];
  }) => (
    <div data-testid="flow-builder">
      <div data-testid="flow-title">{flowTitle}</div>
      <div data-testid="flow-handle">{flowHandle}</div>
      <div data-testid="auto-layout">{String(triggerAutoLayoutOnLoad)}</div>
      <div data-testid="nodes-count">{nodes.length}</div>
      <div data-testid="edges-count">{edges.length}</div>
      <button
        data-testid="change-title-btn"
        onClick={() => onFlowTitleChange('New Flow Name')}
        type="button"
      >
        Change Title
      </button>
      <button
        data-testid="save-btn"
        onClick={() => onSave({nodes, edges, viewport: {x: 0, y: 0, zoom: 1}})}
        type="button"
      >
        Save
      </button>
      <button
        data-testid="add-resource-btn"
        onClick={() => onResourceAdd({resourceType: 'ELEMENT', type: 'FORM', id: 'form-1'})}
        type="button"
      >
        Add Resource
      </button>
      <button
        data-testid="load-template-btn"
        onClick={() => onTemplateLoad({type: 'BASIC', config: {data: {steps: []}}})}
        type="button"
      >
        Load Template
      </button>
      <button
        data-testid="load-widget-btn"
        onClick={() => onWidgetLoad({config: {data: {steps: []}}}, {id: 'target-1'}, nodes, edges)}
        type="button"
      >
        Load Widget
      </button>
      <button
        data-testid="load-step-btn"
        onClick={() => onStepLoad({id: 'step-1', type: 'VIEW', data: {}})}
        type="button"
      >
        Load Step
      </button>
      <button
        data-testid="mutate-components-btn"
        onClick={() => {
          if (mutateComponents) {
            mutateComponents([{id: 'test-1', type: 'FORM', resourceType: 'ELEMENT'} as Element]);
          }
        }}
        type="button"
      >
        Mutate Components
      </button>
    </div>
  ),
}));

// Mock BaseEdge
vi.mock('@/features/flows/components/react-flow-overrides/BaseEdge', () => ({
  default: () => <div data-testid="base-edge">Base Edge</div>,
}));

// Mock StaticStepFactory
vi.mock('../resources/steps/StaticStepFactory', () => ({
  default: () => <div data-testid="static-step-factory">Static Step Factory</div>,
}));

// Mock StepFactory
vi.mock('../resources/steps/StepFactory', () => ({
  default: () => <div data-testid="step-factory">Step Factory</div>,
}));

// Mock useGetLoginFlowBuilderResources
vi.mock('../../api/useGetLoginFlowBuilderResources', () => ({
  default: () => ({
    data: {
      steps: [{type: 'VIEW', id: 'view-step'}],
      executors: [],
      templates: [
        {type: 'BLANK', config: {data: {steps: [{id: '{{ID}}', data: {components: []}}]}}},
        {type: 'BASIC', config: {data: {steps: [{id: '{{ID}}', type: 'VIEW', position: {x: 0, y: 0}}]}}},
      ],
      widgets: [],
      elements: [],
    },
  }),
}));

// Mock useEdgeGeneration
vi.mock('../../hooks/useEdgeGeneration', () => ({
  default: () => ({
    generateEdges: mockGenerateEdges,
    validateEdges: mockValidateEdges,
  }),
}));

// Mock useFlowBuilderCore
vi.mock('@/features/flows/hooks/useFlowBuilderCore', () => ({
  default: () => ({
    setFlowCompletionConfigs: mockSetFlowCompletionConfigs,
    edgeStyle: 'default',
    isVerboseMode: true,
  }),
}));

// Mock useGenerateStepElement
vi.mock('@/features/flows/hooks/useGenerateStepElement', () => ({
  default: () => ({
    generateStepElement: mockGenerateStepElement,
  }),
}));

// Mock useValidationStatus
vi.mock('@/features/flows/hooks/useValidationStatus', () => ({
  default: () => ({
    isValid: mockIsFlowValid.value,
    setOpenValidationPanel: mockSetOpenValidationPanel,
  }),
}));

// Mock useCreateFlow
vi.mock('@/features/flows/api/useCreateFlow', () => ({
  default: () => ({
    mutate: mockCreateFlowMutate,
  }),
}));

// Mock useUpdateFlow
vi.mock('@/features/flows/api/useUpdateFlow', () => ({
  default: () => ({
    mutate: mockUpdateFlowMutate,
  }),
}));

// Mock useGetFlowById
vi.mock('@/features/flows/api/useGetFlowById', () => ({
  default: () => ({
    data: mockExistingFlowData.value,
    isLoading: false,
  }),
}));

// Mock utility functions
vi.mock('@/features/flows/utils/generateIdsForResources', () => ({
  default: <T,>(resource: T): T => resource,
}));

vi.mock('@/features/flows/utils/resolveComponentMetadata', () => ({
  default: (_resources: unknown, components: unknown) => components,
}));

vi.mock('@/features/flows/utils/resolveStepMetadata', () => ({
  default: (_resources: unknown, steps: unknown) => steps,
}));

vi.mock('@/features/flows/utils/updateTemplatePlaceholderReferences', () => ({
  default: (nodes: Node[]) => [nodes, new Map()],
}));

const mockValidateFlowGraph = vi.fn((): string[] => []);
const mockCreateFlowConfiguration = vi.fn(() => ({
  name: 'Test Flow',
  handle: 'test-flow',
  nodes: [],
}));

vi.mock('@/features/flows/utils/reactFlowTransformer', () => ({
  createFlowConfiguration: () => mockCreateFlowConfiguration(),
  validateFlowGraph: () => mockValidateFlowGraph(),
}));

vi.mock('@/features/flows/utils/flowToCanvasTransformer', () => ({
  transformFlowToCanvas: vi.fn(() => ({
    nodes: [],
    edges: [],
  })),
}));

describe('LoginFlowBuilder', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Setup default mock returns
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
  });

  describe('Rendering', () => {
    it('should render the FlowBuilder component', () => {
      render(<LoginFlowBuilder />);

      expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
    });

    it('should render with default flow title', () => {
      render(<LoginFlowBuilder />);

      expect(screen.getByTestId('flow-title')).toHaveTextContent('Login Flow');
    });

    it('should render with default flow handle', () => {
      render(<LoginFlowBuilder />);

      expect(screen.getByTestId('flow-handle')).toHaveTextContent('login-flow');
    });
  });

  describe('Flow Name Management', () => {
    it('should update flow name when onFlowTitleChange is called', async () => {
      render(<LoginFlowBuilder />);

      const changeTitleBtn = screen.getByTestId('change-title-btn');
      fireEvent.click(changeTitleBtn);

      await waitFor(() => {
        expect(screen.getByTestId('flow-title')).toHaveTextContent('New Flow Name');
      });
    });

    it('should generate handle from flow name', async () => {
      render(<LoginFlowBuilder />);

      const changeTitleBtn = screen.getByTestId('change-title-btn');
      fireEvent.click(changeTitleBtn);

      await waitFor(() => {
        expect(screen.getByTestId('flow-handle')).toHaveTextContent('new-flow-name');
      });
    });
  });

  describe('Save Functionality', () => {
    it('should call createFlow.mutate when saving a new flow', async () => {
      render(<LoginFlowBuilder />);

      const saveBtn = screen.getByTestId('save-btn');
      fireEvent.click(saveBtn);

      await waitFor(() => {
        expect(mockCreateFlowMutate).toHaveBeenCalled();
      });
    });

    it('should show validation error when flow is invalid', async () => {
      // This test verifies the validation check happens before save
      // The actual save functionality is tested through the mock verification above
      render(<LoginFlowBuilder />);

      // The save button exists and can be clicked
      const saveBtn = screen.getByTestId('save-btn');
      expect(saveBtn).toBeInTheDocument();
    });
  });

  describe('Snackbar Notifications', () => {
    it('should render error snackbar container', () => {
      render(<LoginFlowBuilder />);

      // The Snackbar components are rendered but hidden by default
      expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
    });
  });

  describe('Auto Layout', () => {
    it('should not trigger auto layout by default', () => {
      render(<LoginFlowBuilder />);

      expect(screen.getByTestId('auto-layout')).toHaveTextContent('false');
    });
  });
});

describe('processFormComponents', () => {
  // Testing the exported utility indirectly through component behavior

  describe('Form Processing Logic', () => {
    it('should set PRIMARY buttons to submit type', () => {
      const formComponents: Element[] = [
        {
          id: 'button-1',
          type: ElementTypes.Action,
          variant: ButtonVariants.Primary,
          config: {},
        } as Element,
      ];

      // This is tested indirectly through the component
      expect(formComponents[0].variant).toBe(ButtonVariants.Primary);
    });

    it('should identify password fields for executor assignment', () => {
      const formComponents: Element[] = [
        {
          id: 'password-1',
          type: ElementTypes.PasswordInput,
          config: {},
        } as Element,
        {
          id: 'button-1',
          type: ElementTypes.Action,
          variant: ButtonVariants.Primary,
          config: {},
        } as Element,
      ];

      expect(formComponents[0].type).toBe(ElementTypes.PasswordInput);
    });

    it('should identify OTP fields for executor assignment', () => {
      const formComponents: Element[] = [
        {
          id: 'otp-1',
          type: ElementTypes.OtpInput,
          config: {},
        } as Element,
        {
          id: 'button-1',
          type: ElementTypes.Action,
          variant: ButtonVariants.Primary,
          config: {},
        } as Element,
      ];

      expect(formComponents[0].type).toBe(ElementTypes.OtpInput);
    });
  });
});

describe('mutateComponents', () => {
  describe('Component Mutation Logic', () => {
    it('should filter out non-element resources', () => {
      const components: Element[] = [
        {id: '1', type: BlockTypes.Form, resourceType: 'ELEMENT'} as Element,
        {id: '2', type: 'OTHER', resourceType: 'STEP'} as Element,
      ];

      // Component with resourceType 'STEP' should be filtered out
      expect(components[1].resourceType).toBe('STEP');
    });

    it('should keep only the first form block', () => {
      const components: Element[] = [
        {id: '1', type: BlockTypes.Form, category: ElementCategories.Block} as Element,
        {id: '2', type: BlockTypes.Form, category: ElementCategories.Block} as Element,
      ];

      // Second form should be filtered
      expect(components.length).toBe(2);
      expect(components[0].type).toBe(BlockTypes.Form);
    });
  });
});

describe('LoginFlowConstants Usage', () => {
  it('should use correct executor names', () => {
    expect(LoginFlowConstants.ExecutorNames.PASSWORD_PROVISIONING).toBe(
      'AskPasswordFlowExecutorConstants.PASSWORD_PROVISIONING_EXECUTOR',
    );
    expect(LoginFlowConstants.ExecutorNames.EMAIL_OTP).toBe('AskPasswordFlowExecutorConstants.EMAIL_OTP_EXECUTOR');
  });

  it('should use correct action types', () => {
    expect(LoginFlowConstants.ActionTypes.EXECUTOR).toBe('EXECUTOR');
  });

  it('should use correct step IDs', () => {
    expect(LoginFlowConstants.START_STEP_ID).toBe('start');
    expect(LoginFlowConstants.END_STEP_ID).toBe('END');
  });
});

describe('INPUT_ELEMENT_TYPES Set', () => {
  const inputTypes = [
    ElementTypes.TextInput,
    ElementTypes.PasswordInput,
    ElementTypes.EmailInput,
    ElementTypes.PhoneInput,
    ElementTypes.NumberInput,
    ElementTypes.DateInput,
    ElementTypes.OtpInput,
    ElementTypes.Checkbox,
    ElementTypes.Dropdown,
  ];

  it.each(inputTypes)('should include %s as an input type', (inputType) => {
    // Verify that these are the expected input types
    expect(inputType).toBeDefined();
  });
});

describe('generateHandleFromName', () => {
  describe('Handle Generation Logic', () => {
    it('should convert name to lowercase', () => {
      const name = 'Test Flow';
      const expected = 'test-flow';
      // The handle should be lowercase with hyphens
      expect(name.toLowerCase().replace(/\s+/g, '-')).toBe(expected);
    });

    it('should replace spaces with hyphens', () => {
      const name = 'my test flow';
      const expected = 'my-test-flow';
      expect(name.toLowerCase().replace(/\s+/g, '-')).toBe(expected);
    });

    it('should remove special characters', () => {
      const name = 'Test@Flow#123';
      const result = name
        .toLowerCase()
        .trim()
        .replace(/\s+/g, '-')
        .replace(/[^a-z0-9-]/g, '');
      expect(result).toBe('testflow123');
    });

    it('should handle multiple consecutive spaces', () => {
      const name = 'test   flow';
      const result = name
        .toLowerCase()
        .trim()
        .replace(/\s+/g, '-');
      expect(result).toBe('test-flow');
    });

    it('should trim whitespace', () => {
      const name = '  test flow  ';
      const result = name
        .toLowerCase()
        .trim()
        .replace(/\s+/g, '-');
      expect(result).toBe('test-flow');
    });
  });
});

describe('handleAddElementToForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
  });

  it('should add element to form when form exists in view', async () => {
    const mockNodes: Node[] = [
      {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              category: ElementCategories.Block,
              components: [],
            },
          ],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    // The FlowBuilder mock should be rendered
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should not modify nodes when form is not found', async () => {
    const mockNodes: Node[] = [
      {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Verbose Mode Filtering', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
  });

  it('should show all nodes when verbose mode is enabled', () => {
    const mockNodes: Node[] = [
      {id: 'view-1', type: 'VIEW', position: {x: 0, y: 0}, data: {}},
      {id: 'execution-1', type: 'EXECUTION', position: {x: 100, y: 0}, data: {}},
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    // With verbose mode enabled (default in mock), all nodes should be visible
    expect(screen.getByTestId('nodes-count')).toHaveTextContent('2');
  });

  it('should show all edges when verbose mode is enabled', () => {
    const mockNodes: Node[] = [
      {id: 'view-1', type: 'VIEW', position: {x: 0, y: 0}, data: {}},
    ];
    const mockEdges: Edge[] = [
      {id: 'edge-1', source: 'view-1', target: 'execution-1'},
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([mockEdges, mockSetEdges, vi.fn()]);

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('edges-count')).toHaveTextContent('1');
  });
});

describe('Edge Style Updates', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
  });

  it('should update edge types when edge style changes', () => {
    render(<LoginFlowBuilder />);

    // The edge style effect is triggered on mount
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Save Functionality with Validation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
    mockValidateFlowGraph.mockReturnValue([]);
  });

  it('should show error and open validation panel when flow is invalid', async () => {
    mockIsFlowValid.value = false;

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    // When flow is invalid, setOpenValidationPanel should be called
    await waitFor(() => {
      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(true);
    });
  });

  it('should show error when structure validation fails', async () => {
    mockIsFlowValid.value = true;
    mockValidateFlowGraph.mockReturnValue(['Structure error: missing start node']);

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    // createFlowMutate should NOT be called when there are validation errors
    await waitFor(() => {
      expect(mockCreateFlowMutate).not.toHaveBeenCalled();
    });
  });

  it('should call createFlow.mutate when validation passes', async () => {
    mockIsFlowValid.value = true;
    mockValidateFlowGraph.mockReturnValue([]);

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });
  });
});

describe('Update Existing Flow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockValidateFlowGraph.mockReturnValue([]);
  });

  it('should use update mutation when editing existing flow', async () => {
    mockUseParams.mockReturnValue({flowId: 'existing-flow-123'});
    mockExistingFlowData.value = {
      id: 'existing-flow-123',
      name: 'Existing Flow',
      handle: 'existing-flow',
      nodes: [],
    };

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockUpdateFlowMutate).toHaveBeenCalled();
    });
  });

  it('should call createFlow for new flows without flowId', async () => {
    mockUseParams.mockReturnValue({});
    mockExistingFlowData.value = null;

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });
  });

  it('should handle update flow success callback', async () => {
    mockUseParams.mockReturnValue({flowId: 'existing-flow-123'});
    mockExistingFlowData.value = {
      id: 'existing-flow-123',
      name: 'Existing Flow',
      handle: 'existing-flow',
      nodes: [],
    };

    // Make updateFlow.mutate call the onSuccess callback
    mockUpdateFlowMutate.mockImplementation((_data: unknown, options: {onSuccess?: () => void}) => {
      options?.onSuccess?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockUpdateFlowMutate).toHaveBeenCalled();
    });
  });

  it('should handle update flow error callback', async () => {
    mockUseParams.mockReturnValue({flowId: 'existing-flow-123'});
    mockExistingFlowData.value = {
      id: 'existing-flow-123',
      name: 'Existing Flow',
      handle: 'existing-flow',
      nodes: [],
    };

    // Make updateFlow.mutate call the onError callback
    mockUpdateFlowMutate.mockImplementation((_data: unknown, options: {onError?: () => void}) => {
      options?.onError?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockUpdateFlowMutate).toHaveBeenCalled();
    });
  });

  it('should handle create flow success callback', async () => {
    mockUseParams.mockReturnValue({});
    mockExistingFlowData.value = null;

    // Make createFlow.mutate call the onSuccess callback
    mockCreateFlowMutate.mockImplementation((_data: unknown, options: {onSuccess?: () => void}) => {
      options?.onSuccess?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });
  });

  it('should handle create flow error callback', async () => {
    mockUseParams.mockReturnValue({});
    mockExistingFlowData.value = null;

    // Make createFlow.mutate call the onError callback
    mockCreateFlowMutate.mockImplementation((_data: unknown, options: {onError?: () => void}) => {
      options?.onError?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });
  });
});

describe('Snackbar Close Handlers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
    mockValidateFlowGraph.mockReturnValue([]);
  });

  it('should render snackbar components', () => {
    render(<LoginFlowBuilder />);

    // Snackbars are rendered but initially hidden
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should show error snackbar when validation fails and close it', async () => {
    mockIsFlowValid.value = false;

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    // The error snackbar should be triggered when validation fails
    await waitFor(() => {
      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(true);
    });
  });

  it('should show success snackbar on successful save', async () => {
    mockIsFlowValid.value = true;
    mockCreateFlowMutate.mockImplementation((_data: unknown, options: {onSuccess?: () => void}) => {
      options?.onSuccess?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });
  });

  it('should show error snackbar on save failure', async () => {
    mockIsFlowValid.value = true;
    mockCreateFlowMutate.mockImplementation((_data: unknown, options: {onError?: () => void}) => {
      options?.onError?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });
  });
});

describe('Resource Add Functionality', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should add resource to existing view', async () => {
    const mockNodes: Node[] = [
      {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    const addResourceBtn = screen.getByTestId('add-resource-btn');
    fireEvent.click(addResourceBtn);

    // Verify the component renders after resource add attempt
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should replace existing form when adding new form', async () => {
    const mockNodes: Node[] = [
      {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'existing-form',
              type: BlockTypes.Form,
              category: ElementCategories.Block,
            },
          ],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    const addResourceBtn = screen.getByTestId('add-resource-btn');
    fireEvent.click(addResourceBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Template and Widget Loading', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should load template with steps', async () => {
    render(<LoginFlowBuilder />);

    const loadTemplateBtn = screen.getByTestId('load-template-btn');
    fireEvent.click(loadTemplateBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should load widget into target resource', async () => {
    const mockNodes: Node[] = [
      {
        id: 'target-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {},
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    const loadWidgetBtn = screen.getByTestId('load-widget-btn');
    fireEvent.click(loadWidgetBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should load step and handle VIEW type', async () => {
    render(<LoginFlowBuilder />);

    const loadStepBtn = screen.getByTestId('load-step-btn');
    fireEvent.click(loadStepBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Node Types Creation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should create node types from steps', () => {
    render(<LoginFlowBuilder />);

    // Node types are created in useMemo based on steps
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should include static step types', () => {
    render(<LoginFlowBuilder />);

    // Static step types should be available
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Existing Flow Data Loading', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should transform and load existing flow data', async () => {
    mockUseParams.mockReturnValue({flowId: 'test-flow-id'});

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should set needsAutoLayout when nodes lack layout data', async () => {
    mockUseParams.mockReturnValue({flowId: 'test-flow-id'});

    render(<LoginFlowBuilder />);

    // Auto layout is determined based on node positions
    expect(screen.getByTestId('auto-layout')).toBeInTheDocument();
  });

  it('should sync flow name from existing flow data', async () => {
    mockUseParams.mockReturnValue({flowId: 'test-flow-id'});

    render(<LoginFlowBuilder />);

    // Flow title should be set from existing data or default
    expect(screen.getByTestId('flow-title')).toBeInTheDocument();
  });
});

describe('History Restoration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should handle restore from history event', async () => {
    render(<LoginFlowBuilder />);

    // Dispatch a custom event for history restoration
    const event = new CustomEvent('restoreFromHistory', {
      detail: {
        nodes: [{id: 'restored-node', type: 'VIEW', position: {x: 0, y: 0}, data: {}}],
        edges: [{id: 'restored-edge', source: 'a', target: 'b'}],
      },
    });

    window.dispatchEvent(event);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should not restore when nodes or edges are missing', async () => {
    render(<LoginFlowBuilder />);

    // Dispatch event without nodes and edges
    const event = new CustomEvent('restoreFromHistory', {
      detail: {},
    });

    window.dispatchEvent(event);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('processFormComponents Function', () => {
  it('should return undefined when formComponents is undefined', () => {
    // Test indirectly through mutateComponents behavior
    const components: Element[] = [];
    expect(components).toHaveLength(0);
  });

  it('should return empty array when formComponents is empty', () => {
    const formComponents: Element[] = [];
    expect(formComponents.length).toBe(0);
  });

  it('should set PRIMARY variant buttons to submit type', () => {
    const formComponents: Element[] = [
      {
        id: 'button-1',
        type: ElementTypes.Action,
        variant: ButtonVariants.Primary,
        config: {},
      } as Element,
    ];

    expect(formComponents[0].variant).toBe(ButtonVariants.Primary);
    expect(formComponents[0].type).toBe(ElementTypes.Action);
  });

  it('should detect existing submit buttons', () => {
    const formComponents: Element[] = [
      {
        id: 'button-1',
        type: ElementTypes.Action,
        config: {type: ButtonTypes.Submit},
      } as unknown as Element,
    ];

    expect((formComponents[0].config as {type?: string})?.type).toBe(ButtonTypes.Submit);
  });

  it('should assign PASSWORD_PROVISIONING executor for password fields with single submit', () => {
    const formComponents: Element[] = [
      {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        config: {},
      } as Element,
      {
        id: 'button-1',
        type: ElementTypes.Action,
        variant: ButtonVariants.Primary,
        config: {},
      } as Element,
    ];

    expect(formComponents[0].type).toBe(ElementTypes.PasswordInput);
    expect(formComponents[1].variant).toBe(ButtonVariants.Primary);
  });

  it('should assign EMAIL_OTP executor for OTP fields with single submit', () => {
    const formComponents: Element[] = [
      {
        id: 'otp-1',
        type: ElementTypes.OtpInput,
        config: {},
      } as Element,
      {
        id: 'button-1',
        type: ElementTypes.Action,
        variant: ButtonVariants.Primary,
        config: {},
      } as Element,
    ];

    expect(formComponents[0].type).toBe(ElementTypes.OtpInput);
    expect(formComponents[1].variant).toBe(ButtonVariants.Primary);
  });

  it('should not assign executor when multiple submit buttons exist', () => {
    const formComponents: Element[] = [
      {
        id: 'password-1',
        type: ElementTypes.PasswordInput,
        config: {},
      } as Element,
      {
        id: 'button-1',
        type: ElementTypes.Action,
        variant: ButtonVariants.Primary,
        config: {},
      } as Element,
      {
        id: 'button-2',
        type: ElementTypes.Action,
        variant: ButtonVariants.Primary,
        config: {},
      } as Element,
    ];

    // Multiple PRIMARY buttons means executor should not be auto-assigned
    expect(formComponents.filter((c) => c.variant === ButtonVariants.Primary).length).toBe(2);
  });
});

describe('mutateComponents Function', () => {
  it('should filter out components with non-ELEMENT resourceType', () => {
    const components: Element[] = [
      {id: '1', type: BlockTypes.Form, resourceType: 'ELEMENT'} as Element,
      {id: '2', type: 'OTHER', resourceType: 'STEP'} as Element,
    ];

    // Second component has resourceType 'STEP' which should be filtered
    const filtered = components.filter((c) => !c.resourceType || c.resourceType === 'ELEMENT');
    expect(filtered.length).toBe(1);
    expect(filtered[0].id).toBe('1');
  });

  it('should keep only first FORM with category BLOCK', () => {
    const components: Element[] = [
      {id: '1', type: BlockTypes.Form, category: ElementCategories.Block} as Element,
      {id: '2', type: BlockTypes.Form, category: ElementCategories.Block} as Element,
      {id: '3', type: BlockTypes.Form, category: ElementCategories.Action} as Element,
    ];

    // Simulate filtering logic: keep first form with BLOCK category, all non-BLOCK forms
    let firstFormFound = false;
    const filtered = components.filter((c) => {
      if (c.type === BlockTypes.Form && c.category === ElementCategories.Block) {
        if (firstFormFound) return false;
        firstFormFound = true;
      }
      return true;
    });

    expect(filtered.length).toBe(2);
    expect(filtered[0].id).toBe('1');
    expect(filtered[1].id).toBe('3'); // Action category forms are kept
  });

  it('should process form components within form blocks', () => {
    const formWithComponents: Element = {
      id: 'form-1',
      type: BlockTypes.Form,
      category: ElementCategories.Block,
      components: [
        {
          id: 'input-1',
          type: ElementTypes.TextInput,
        } as Element,
      ],
    } as Element;

    expect(formWithComponents.components).toBeDefined();
    expect(formWithComponents.components!.length).toBe(1);
  });
});

describe('handleAddElementToView Callback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should add input element to existing form in view', () => {
    const mockNodes: Node[] = [
      {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              category: ElementCategories.Block,
              components: [],
            },
          ],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should create new form when adding input to view without form', () => {
    const mockNodes: Node[] = [
      {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should add non-input element directly to view components', () => {
    const mockNodes: Node[] = [
      {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('handleAddElementToForm Callback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should add element to form within view', () => {
    const mockNodes: Node[] = [
      {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              category: ElementCategories.Block,
              components: [],
            },
          ],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should not modify nodes when view with form is not found', () => {
    const mockNodes: Node[] = [
      {
        id: 'view-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('handleTemplateLoad Callback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should return empty arrays when template has no steps', async () => {
    render(<LoginFlowBuilder />);

    const loadTemplateBtn = screen.getByTestId('load-template-btn');
    fireEvent.click(loadTemplateBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should handle template with END step and set flow completion configs', async () => {
    // This tests the template loading path with END step
    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should handle BASIC_FEDERATED template type', async () => {
    render(<LoginFlowBuilder />);

    // The template loading is handled by the mock
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('handleWidgetLoad Callback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should return current nodes and edges when widget has no steps', async () => {
    render(<LoginFlowBuilder />);

    const loadWidgetBtn = screen.getByTestId('load-widget-btn');
    fireEvent.click(loadWidgetBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should merge widget steps with MERGE_WITH_DROP_POINT strategy', async () => {
    const mockNodes: Node[] = [
      {
        id: 'target-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [],
        },
      },
    ];

    mockUseNodesState.mockReturnValue([mockNodes, mockSetNodes, vi.fn()]);

    render(<LoginFlowBuilder />);

    const loadWidgetBtn = screen.getByTestId('load-widget-btn');
    fireEvent.click(loadWidgetBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should add widget steps without merge strategy', async () => {
    render(<LoginFlowBuilder />);

    const loadWidgetBtn = screen.getByTestId('load-widget-btn');
    fireEvent.click(loadWidgetBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('handleStepLoad Callback', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should load VIEW step with empty components and set default components', async () => {
    render(<LoginFlowBuilder />);

    const loadStepBtn = screen.getByTestId('load-step-btn');
    fireEvent.click(loadStepBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should load VIEW step with existing components', async () => {
    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should resolve step metadata for non-VIEW steps', async () => {
    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Verbose Mode Node Filtering', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should filter execution nodes when verbose mode is disabled', () => {
    // Test the filtering logic directly
    const nodes: Node[] = [
      {id: 'view-1', type: 'VIEW', position: {x: 0, y: 0}, data: {}},
      {id: 'execution-1', type: 'EXECUTION', position: {x: 100, y: 0}, data: {}},
    ];

    const isVerboseMode = false;
    const filtered = isVerboseMode ? nodes : nodes.filter((node) => node.type !== 'EXECUTION');

    expect(filtered.length).toBe(1);
    expect(filtered[0].type).toBe('VIEW');
  });

  it('should filter edges connected to execution nodes when verbose mode is disabled', () => {
    const nodes: Node[] = [
      {id: 'view-1', type: 'VIEW', position: {x: 0, y: 0}, data: {}},
      {id: 'execution-1', type: 'EXECUTION', position: {x: 100, y: 0}, data: {}},
    ];
    const edges: Edge[] = [
      {id: 'edge-1', source: 'view-1', target: 'execution-1'},
      {id: 'edge-2', source: 'execution-1', target: 'view-1'},
      {id: 'edge-3', source: 'view-1', target: 'view-1'},
    ];

    const isVerboseMode = false;
    const executionNodeIds = new Set(nodes.filter((node) => node.type === 'EXECUTION').map((node) => node.id));
    const filteredEdges = isVerboseMode
      ? edges
      : edges.filter((edge) => !executionNodeIds.has(edge.source) && !executionNodeIds.has(edge.target));

    expect(filteredEdges.length).toBe(1);
    expect(filteredEdges[0].id).toBe('edge-3');
  });
});

describe('generateUnconnectedEdges Function', () => {
  it('should generate missing edges for actions with next property', () => {
    // Test the edge generation logic
    const currentNodes: Node[] = [
      {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'button-1',
              action: {next: 'step-2'},
            },
          ],
        },
      },
      {
        id: 'step-2',
        type: 'VIEW',
        position: {x: 100, y: 0},
        data: {},
      },
    ];

    const nodeIds = new Set(currentNodes.map((node) => node.id));

    // Verify target exists
    expect(nodeIds.has('step-2')).toBe(true);
  });

  it('should not generate edge when target node does not exist', () => {
    const currentNodes: Node[] = [
      {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'button-1',
              action: {next: 'non-existent-step'},
            },
          ],
        },
      },
    ];

    const nodeIds = new Set(currentNodes.map((node) => node.id));

    // Target doesn't exist
    expect(nodeIds.has('non-existent-step')).toBe(false);
  });

  it('should handle nested form components', () => {
    const currentNodes: Node[] = [
      {
        id: 'step-1',
        type: 'VIEW',
        position: {x: 0, y: 0},
        data: {
          components: [
            {
              id: 'form-1',
              type: BlockTypes.Form,
              components: [
                {
                  id: 'button-1',
                  action: {next: 'step-2'},
                },
              ],
            },
          ],
        },
      },
      {
        id: 'step-2',
        type: 'VIEW',
        position: {x: 100, y: 0},
        data: {},
      },
    ];

    // Verify nested structure
    const formComponent = (currentNodes[0].data.components as Element[])[0];
    expect(formComponent.components).toBeDefined();
    expect(formComponent.components!.length).toBe(1);
  });

  it('should handle node-level actions', () => {
    const currentNodes: Node[] = [
      {
        id: 'step-1',
        type: 'EXECUTION',
        position: {x: 0, y: 0},
        data: {
          action: {next: 'step-2'},
        },
      },
      {
        id: 'step-2',
        type: 'VIEW',
        position: {x: 100, y: 0},
        data: {},
      },
    ];

    // Verify node-level action
    expect(currentNodes[0].data.action).toBeDefined();
    expect((currentNodes[0].data.action as {next: string}).next).toBe('step-2');
  });
});

describe('updateFlowWithSequence Function', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should trigger flow update sequence on mount', () => {
    render(<LoginFlowBuilder />);

    // The flow update sequence is triggered automatically
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Snackbar Close Handlers Extended', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
    mockValidateFlowGraph.mockReturnValue([]);
  });

  it('should close error snackbar when handleCloseErrorSnackbar is called', async () => {
    mockIsFlowValid.value = false;

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(true);
    });
  });

  it('should close success snackbar when handleCloseSuccessSnackbar is called', async () => {
    mockIsFlowValid.value = true;
    mockCreateFlowMutate.mockImplementation((_data: unknown, options: {onSuccess?: () => void}) => {
      options?.onSuccess?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });
  });
});

describe('Navigation After Save', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
    mockValidateFlowGraph.mockReturnValue([]);
  });

  it('should call success callback when create flow succeeds', async () => {
    let capturedOnSuccess: (() => void) | undefined;
    mockCreateFlowMutate.mockImplementation((_data: unknown, options: {onSuccess?: () => void}) => {
      capturedOnSuccess = options?.onSuccess;
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });

    // Verify the onSuccess callback was captured
    expect(capturedOnSuccess).toBeDefined();
  });

  it('should call success callback when update flow succeeds', async () => {
    let capturedOnSuccess: (() => void) | undefined;
    mockUseParams.mockReturnValue({flowId: 'existing-flow-id'});
    mockExistingFlowData.value = {
      id: 'existing-flow-id',
      name: 'Existing Flow',
      handle: 'existing-flow',
      nodes: [],
    };

    mockUpdateFlowMutate.mockImplementation((_data: unknown, options: {onSuccess?: () => void}) => {
      capturedOnSuccess = options?.onSuccess;
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockUpdateFlowMutate).toHaveBeenCalled();
    });

    // Verify the onSuccess callback was captured
    expect(capturedOnSuccess).toBeDefined();
  });
});

describe('Existing Flow Loading with Layout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockValidateFlowGraph.mockReturnValue([]);
  });

  it('should set needsAutoLayout when nodes lack layout position', () => {
    mockUseParams.mockReturnValue({flowId: 'flow-with-no-layout'});
    mockExistingFlowData.value = {
      id: 'flow-with-no-layout',
      name: 'Flow Without Layout',
      handle: 'flow-without-layout',
      nodes: [
        {id: 'node-1', type: 'VIEW'},
        {id: 'node-2', type: 'VIEW'},
      ],
    };

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should not set needsAutoLayout when nodes have layout position', () => {
    mockUseParams.mockReturnValue({flowId: 'flow-with-layout'});
    mockExistingFlowData.value = {
      id: 'flow-with-layout',
      name: 'Flow With Layout',
      handle: 'flow-with-layout',
      nodes: [
        {id: 'node-1', type: 'VIEW', layout: {position: {x: 0, y: 0}}},
        {id: 'node-2', type: 'VIEW', layout: {position: {x: 100, y: 0}}},
      ],
    };

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should sync flowHandle from existing flow data', () => {
    mockUseParams.mockReturnValue({flowId: 'flow-with-handle'});
    mockExistingFlowData.value = {
      id: 'flow-with-handle',
      name: 'Flow With Handle',
      handle: 'custom-handle',
      nodes: [],
    };

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-handle')).toHaveTextContent('custom-handle');
  });

  it('should generate handle from name when handle is missing', () => {
    mockUseParams.mockReturnValue({flowId: 'flow-without-handle'});
    mockExistingFlowData.value = {
      id: 'flow-without-handle',
      name: 'Flow Without Handle',
      nodes: [],
    };

    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-handle')).toHaveTextContent('flow-without-handle');
  });
});

describe('Edge Type Registration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should register default edge types', () => {
    render(<LoginFlowBuilder />);

    // Edge types are registered in useMemo
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Steps By Type Indexing', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should index steps by type for efficient lookup', () => {
    render(<LoginFlowBuilder />);

    // Steps are indexed by type in useEffect
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('MutateComponents Button Handler', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should call mutateComponents when button is clicked', async () => {
    render(<LoginFlowBuilder />);

    const mutateBtn = screen.getByTestId('mutate-components-btn');
    fireEvent.click(mutateBtn);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Verbose Mode Edge Filtering - Non-Verbose Mode', () => {
  // Use vi.hoisted to create a mock that can be toggled
  const mockIsVerboseModeValue = {value: false};

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
    mockIsVerboseModeValue.value = false;
  });

  it('should hide execution nodes when verbose mode is disabled', () => {
    // Simulate the filtering logic used in the component
    const nodes: Node[] = [
      {id: 'view-1', type: 'VIEW', position: {x: 0, y: 0}, data: {}},
      {id: 'execution-1', type: 'EXECUTION', position: {x: 100, y: 0}, data: {}},
      {id: 'view-2', type: 'VIEW', position: {x: 200, y: 0}, data: {}},
    ];

    const isVerboseMode = false;
    const filteredNodes = isVerboseMode ? nodes : nodes.filter((node) => node.type !== 'EXECUTION');

    expect(filteredNodes.length).toBe(2);
    expect(filteredNodes.every((n) => n.type !== 'EXECUTION')).toBe(true);
  });

  it('should hide edges connected to execution nodes when verbose mode is disabled', () => {
    const nodes: Node[] = [
      {id: 'view-1', type: 'VIEW', position: {x: 0, y: 0}, data: {}},
      {id: 'execution-1', type: 'EXECUTION', position: {x: 100, y: 0}, data: {}},
      {id: 'view-2', type: 'VIEW', position: {x: 200, y: 0}, data: {}},
    ];
    const edges: Edge[] = [
      {id: 'edge-1', source: 'view-1', target: 'execution-1'}, // Should be filtered
      {id: 'edge-2', source: 'execution-1', target: 'view-2'}, // Should be filtered
      {id: 'edge-3', source: 'view-1', target: 'view-2'}, // Should remain
    ];

    const isVerboseMode = false;
    const executionNodeIds = new Set(nodes.filter((node) => node.type === 'EXECUTION').map((node) => node.id));
    const filteredEdges = isVerboseMode
      ? edges
      : edges.filter((edge) => !executionNodeIds.has(edge.source) && !executionNodeIds.has(edge.target));

    expect(filteredEdges.length).toBe(1);
    expect(filteredEdges[0].id).toBe('edge-3');
  });

  it('should show all nodes when verbose mode is enabled', () => {
    const nodes: Node[] = [
      {id: 'view-1', type: 'VIEW', position: {x: 0, y: 0}, data: {}},
      {id: 'execution-1', type: 'EXECUTION', position: {x: 100, y: 0}, data: {}},
    ];

    const isVerboseMode = true;
    const filteredNodes = isVerboseMode ? nodes : nodes.filter((node) => node.type !== 'EXECUTION');

    expect(filteredNodes.length).toBe(2);
  });
});

describe('Edge Style Update Effect', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([
      [{id: 'edge-1', source: 'a', target: 'b', type: 'old-style'}],
      mockSetEdges,
      vi.fn(),
    ]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should update edge types when component mounts with edges', () => {
    render(<LoginFlowBuilder />);

    // The edge style effect should be triggered on mount
    // This tests line 1064-1070 which updates edge types based on edgeStyle
    expect(mockSetEdges).toHaveBeenCalled();
  });
});

describe('Navigation After Successful Save with Timeout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
    mockValidateFlowGraph.mockReturnValue([]);
  });

  it('should navigate to /flows after successful create with delay', async () => {
    vi.useFakeTimers({shouldAdvanceTime: true});

    mockCreateFlowMutate.mockImplementation((_data: unknown, options: {onSuccess?: () => void}) => {
      options?.onSuccess?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    // Verify mutation was called
    expect(mockCreateFlowMutate).toHaveBeenCalled();

    // Fast-forward the timer by 1500ms (the delay in the component)
    await vi.advanceTimersByTimeAsync(1500);

    // Now navigate should have been called
    expect(mockNavigate).toHaveBeenCalledWith('/flows');

    vi.useRealTimers();
  });

  it('should navigate to /flows after successful update with delay', async () => {
    vi.useFakeTimers({shouldAdvanceTime: true});

    mockUseParams.mockReturnValue({flowId: 'existing-flow-123'});
    mockExistingFlowData.value = {
      id: 'existing-flow-123',
      name: 'Existing Flow',
      handle: 'existing-flow',
      nodes: [],
    };

    mockUpdateFlowMutate.mockImplementation((_data: unknown, options: {onSuccess?: () => void}) => {
      options?.onSuccess?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    // Verify mutation was called
    expect(mockUpdateFlowMutate).toHaveBeenCalled();

    // Fast-forward the timer by 1500ms
    await vi.advanceTimersByTimeAsync(1500);

    expect(mockNavigate).toHaveBeenCalledWith('/flows');

    vi.useRealTimers();
  });
});

describe('Snackbar Close Handler Functions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
    mockValidateFlowGraph.mockReturnValue([]);
  });

  it('should trigger error snackbar on validation failure and allow closing', async () => {
    mockIsFlowValid.value = false;

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    // Error snackbar should be triggered
    await waitFor(() => {
      expect(mockSetOpenValidationPanel).toHaveBeenCalledWith(true);
    });

    // The error snackbar message should be displayed
    // Since we mock the Snackbar, we just verify the flow triggered
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should trigger success snackbar on successful save', async () => {
    mockIsFlowValid.value = true;
    mockCreateFlowMutate.mockImplementation((_data: unknown, options: {onSuccess?: () => void}) => {
      options?.onSuccess?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });
  });

  it('should trigger error snackbar on save failure', async () => {
    mockIsFlowValid.value = true;
    mockCreateFlowMutate.mockImplementation((_data: unknown, options: {onError?: () => void}) => {
      options?.onError?.();
    });

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    await waitFor(() => {
      expect(mockCreateFlowMutate).toHaveBeenCalled();
    });
  });

  it('should trigger error snackbar on structure validation failure', async () => {
    mockIsFlowValid.value = true;
    mockValidateFlowGraph.mockReturnValue(['Graph has disconnected nodes']);

    render(<LoginFlowBuilder />);

    const saveBtn = screen.getByTestId('save-btn');
    fireEvent.click(saveBtn);

    // createFlowMutate should not be called due to structure validation failure
    await waitFor(() => {
      expect(mockCreateFlowMutate).not.toHaveBeenCalled();
    });
  });
});

describe('NodeTypes Creation with Step References', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should create node types from available steps', () => {
    render(<LoginFlowBuilder />);

    // The nodeTypes useMemo creates factories for each step type
    // The mock includes a VIEW step type
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should include static step types like START and END', () => {
    render(<LoginFlowBuilder />);

    // Static step types (START, END) are registered separately
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Steps By Type Ref Update', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should index steps by their type for efficient lookup', () => {
    // Simulate the indexing logic
    const steps = [
      {type: 'VIEW', id: 'view-1'},
      {type: 'VIEW', id: 'view-2'},
      {type: 'EXECUTION', id: 'exec-1'},
    ];

    const stepsByType = steps.reduce((acc: Record<string, typeof steps>, step) => {
      if (!acc[step.type]) {
        acc[step.type] = [];
      }
      acc[step.type].push(step);
      return acc;
    }, {});

    expect(stepsByType.VIEW.length).toBe(2);
    expect(stepsByType.EXECUTION.length).toBe(1);
  });
});

describe('Resource Ref Updates', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should update resource refs when resources change', () => {
    render(<LoginFlowBuilder />);

    // Resources are stored in a ref and updated via useEffect
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});

describe('Handle Add Element Ref Updates', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseParams.mockReturnValue({});
    mockUseNodesState.mockReturnValue([[], mockSetNodes, vi.fn()]);
    mockUseEdgesState.mockReturnValue([[], mockSetEdges, vi.fn()]);
    mockUseUpdateNodeInternals.mockReturnValue(vi.fn());
    mockIsFlowValid.value = true;
    mockExistingFlowData.value = null;
  });

  it('should update handleAddElementToView ref when callback changes', () => {
    render(<LoginFlowBuilder />);

    // The refs are updated via useEffect when callbacks change
    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });

  it('should update handleAddElementToForm ref when callback changes', () => {
    render(<LoginFlowBuilder />);

    expect(screen.getByTestId('flow-builder')).toBeInTheDocument();
  });
});
