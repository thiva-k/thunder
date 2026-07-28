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

import {StaticStepTypes, StepTypes} from '../models/steps';

class FlowConstants {
  /**
   * Private constructor to avoid object instantiation from outside
   * the class.
   */
  // eslint-disable-next-line @typescript-eslint/no-empty-function
  private constructor() {}

  // 1 min
  public static readonly AUTO_SAVE_INTERVAL: number = 60000;

  // Maximum number of history items to keep
  public static readonly MAX_HISTORY_ITEMS: number = 20;

  /**
   * The ID for the start step in the flow.
   * Uses lowercase to match ReactFlow node type conventions.
   */
  public static readonly START_STEP_ID: string = StaticStepTypes.Start.toLowerCase();

  /**
   * The ID for the end/user onboard step in the flow.
   */
  public static readonly END_STEP_ID: string = StepTypes.End;

  /**
   * Default edge type for the flow canvas.
   */
  public static readonly DEFAULT_EDGE_TYPE: string = 'base-edge';

  /**
   * Executor names for auto-assignment based on field types.
   */
  public static readonly ExecutorNames = {
    PASSWORD_PROVISIONING: 'AskPasswordFlowExecutorConstants.PASSWORD_PROVISIONING_EXECUTOR',
    EMAIL_OTP: 'AskPasswordFlowExecutorConstants.EMAIL_OTP_EXECUTOR',
  } as const;

  /**
   * Action types for button actions.
   */
  public static readonly ActionTypes = {
    EXECUTOR: 'EXECUTOR',
  } as const;
}

export default FlowConstants;
