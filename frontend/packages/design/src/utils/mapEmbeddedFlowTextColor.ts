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

import type {TypographyProps} from '@wso2/oxygen-ui';

/**
 * Maps flow typography color enum names to Material UI color values
 * for consistent text styling across embedded flow components.
 *
 * @param colorEnum - The color enum value (e.g., 'ERROR', 'SUCCESS')
 * @returns The mapped MUI color value (e.g., 'error', 'success'), or undefined if not found
 */
export function mapEmbeddedFlowTextColor(colorEnum: string | undefined): TypographyProps['color'] | undefined {
  switch (colorEnum) {
    case 'ERROR':
      return 'error';
    case 'WARNING':
      return 'warning';
    case 'SUCCESS':
      return 'success';
    case 'INFO':
      return 'info';
    case 'PRIMARY':
      return 'primary';
    case 'SECONDARY':
      return 'secondary';
    default:
      return undefined;
  }
}

export default mapEmbeddedFlowTextColor;
