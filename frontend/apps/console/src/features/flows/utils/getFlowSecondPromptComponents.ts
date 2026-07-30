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

import type {EmbeddedFlowComponent} from '@thunderid/react';
import {FlowNodeType} from '../models/flows';
import type {FlowDefinitionResponse, FlowNode} from '../models/responses';

/**
 * Follows a single outgoing edge from a node: whichever of `onSuccess`, or the first prompt
 * action's `nextNode`, is present. Sufficient for walking the linear/branching shapes this app
 * generates (see `generateFlowGraph`); nodes with multiple genuinely divergent branches (e.g. the
 * MFA channel choice) are walked via their first branch only.
 */
function nextNodeId(node: FlowNode): string | undefined {
  if (node.onSuccess) {
    return node.onSuccess;
  }
  return node.prompts?.[0]?.action?.nextNode;
}

/**
 * Extracts the UI components of a flow's *second* reachable sign-in screen: the screen a user
 * lands on after completing the first PROMPT (e.g. an MFA challenge shown after the primary
 * sign-in form). Walks from the first PROMPT node past its own success chain until a second
 * PROMPT node with renderable components is found.
 *
 * Mirrors {@link getFlowEntryComponents}, continuing one screen further, so it can be passed
 * directly to {@link GatePreview}'s `mock` prop to preview "what happens next."
 *
 * @param flow - The full flow definition, or undefined while loading
 * @returns The second PROMPT screen's components, or null if the flow has no second screen
 *
 * @public
 */
export default function getFlowSecondPromptComponents(
  flow: FlowDefinitionResponse | undefined,
): EmbeddedFlowComponent[] | null {
  if (!flow?.nodes?.length) {
    return null;
  }

  const start = flow.nodes.find((node) => node.type === FlowNodeType.START);
  const visited = new Set<string>();
  let currentId: string | undefined = start?.onSuccess;
  let promptsSeen = 0;

  while (currentId && !visited.has(currentId)) {
    visited.add(currentId);
    const node: FlowNode | undefined = flow.nodes.find((candidate) => candidate.id === currentId);
    if (!node) {
      break;
    }
    if (node.type === FlowNodeType.PROMPT && node.meta?.components?.length) {
      promptsSeen += 1;
      if (promptsSeen === 2) {
        return node.meta.components as EmbeddedFlowComponent[];
      }
    }
    currentId = nextNodeId(node);
  }

  return null;
}
