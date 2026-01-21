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

import type {Edge, Node} from '@xyflow/react';
import {MarkerType} from '@xyflow/react';
import type {Element} from '@/features/flows/models/elements';

/**
 * Generates edges for components that have actions with 'next' references
 * but don't have corresponding edges in the current edge set.
 *
 * @param currentEdges - The current set of edges in the flow.
 * @param currentNodes - The current set of nodes in the flow.
 * @param edgeStyle - The style to apply to generated edges.
 * @returns Array of missing edges that should be added to the flow.
 */
const generateUnconnectedEdges = (currentEdges: Edge[], currentNodes: Node[], edgeStyle: string): Edge[] => {
  const nodeIds = new Set<string>(currentNodes.map((node: Node) => node.id));
  const missingEdges: Edge[] = [];

  const processAction = (stepId: string, resourceId: string, action: unknown): void => {
    if (action && typeof action === 'object' && 'next' in action && action.next) {
      const buttonId: string = resourceId;
      const expectedTarget: string = action.next as string;

      // Ensure expected target exists in nodes
      if (!nodeIds.has(expectedTarget)) {
        // Target node doesn't exist, skip this edge
        return;
      }

      const existingEdge: Edge | undefined = currentEdges.find(
        (edge: Edge) => edge.source === stepId && edge.sourceHandle === `${buttonId}_NEXT`,
      );

      // If no edge exists or it's pointing to the wrong node, add a missing edge
      if (!existingEdge || existingEdge.target !== expectedTarget) {
        missingEdges.push({
          animated: false,
          id: `${buttonId}_MISSING_EDGE`,
          markerEnd: {
            type: MarkerType.Arrow,
          },
          source: stepId,
          sourceHandle: `${buttonId}_NEXT`,
          target: expectedTarget,
          type: edgeStyle,
        });
      }
    }
  };

  currentNodes.forEach((node: Node) => {
    if (!node.data) {
      return;
    }

    if (node.data?.components) {
      (node.data.components as Element[]).forEach((component: Element) => {
        processAction(node.id, component.id, component.action);

        // Process `FORM` components.
        if (component?.components) {
          component.components.forEach((nestedComponent: Element) =>
            processAction(node?.id, nestedComponent.id, nestedComponent.action),
          );
        }
      });
    }

    if (node.data?.action) {
      processAction(node.id, node.id, node.data.action);
    }
  });

  return missingEdges;
};

export default generateUnconnectedEdges;
