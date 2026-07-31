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

import {createContext, type Context} from 'react';

/**
 * Wiring between the compact-mode canvas nodes and FlowBuilder's per-stack
 * expansion state: stack chips expand into their member chips, and the head
 * member chip of an expanded group offers restacking.
 */
export interface CompactStacksContextValue {
  /**
   * Collapse an expanded group back into its stack chip.
   */
  collapseStack: (stackId: string) => void;
  /**
   * Expand a stack chip into its individual member chips.
   */
  expandStack: (stackId: string, memberIds: string[]) => void;
  /**
   * Lookup from the head member node id of each expanded group to its stack
   * id, used by the head chip to offer the restack affordance.
   */
  expandedHeadIdToStackId: ReadonlyMap<string, string>;
}

const CompactStacksContext: Context<CompactStacksContextValue> = createContext<CompactStacksContextValue>({
  collapseStack: () => undefined,
  expandStack: () => undefined,
  expandedHeadIdToStackId: new Map<string, string>(),
});

CompactStacksContext.displayName = 'CompactStacksContext';

export default CompactStacksContext;
