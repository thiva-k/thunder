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

/**
 * Sinhala translations for Thunder applications
 * Organized by namespaces for better tree-shaking and modularity
 */

import {common} from './namespaces/common';
import {develop} from './namespaces/develop';
import {gate} from './namespaces/gate';

export const si = {
  common,
  develop,
  gate,
} as const;

export default si;

// Export individual namespaces for selective imports
export {common} from './namespaces/common';
export {develop} from './namespaces/develop';
export {gate} from './namespaces/gate';
