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

import {createContext} from 'react';

/**
 * Context value for the sidebar component, providing state information
 * about the sidebar's expansion and animation states.
 */
export interface SidebarContextValue {
  /**
   * Indicates whether the sidebar is in mini (collapsed) mode.
   * When true, the sidebar shows only icons without text.
   * This is true when the sidebar is collapsed AND collapsible functionality is enabled.
   */
  mini: boolean;

  /**
   * Indicates whether the sidebar has fully completed its expansion animation.
   * Use this to conditionally render content that should only appear after
   * the expansion transition is complete (e.g., text labels with opacity transitions).
   */
  fullyExpanded: boolean;

  /**
   * Indicates whether the sidebar has fully completed its collapse animation.
   * Use this to conditionally render content that should only appear after
   * the collapse transition is complete.
   */
  fullyCollapsed: boolean;

  /**
   * Indicates whether drawer transitions are enabled.
   * Typically true for desktop viewports where smooth transitions enhance UX,
   * and may be false for mobile/tablet to avoid transition conflicts.
   */
  hasDrawerTransitions: boolean;
}

/**
 * Context for sharing sidebar state across child components.
 * Allows components like MenuContent to adapt their rendering based on
 * whether the sidebar is expanded, collapsed, or mid-transition.
 */
const SidebarContext = createContext<SidebarContextValue>({
  mini: false,
  fullyExpanded: true,
  fullyCollapsed: false,
  hasDrawerTransitions: true,
});

export default SidebarContext;
