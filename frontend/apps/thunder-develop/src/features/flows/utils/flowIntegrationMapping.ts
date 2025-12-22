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

import {AuthenticatorTypes} from '../../integrations/models/authenticators';

/**
 * Flow handle patterns to integration type mappings
 * Maps flow handle patterns to their corresponding integration identifiers
 *
 * @public
 */
const FLOW_HANDLE_TO_INTEGRATION_MAP: Record<string, string> = {
  // Basic authentication flows
  'default-basic-flow': AuthenticatorTypes.BASIC_AUTH,
  basic: AuthenticatorTypes.BASIC_AUTH,
  'login-flow': AuthenticatorTypes.BASIC_AUTH,
  'basic-with-prompt-flow': AuthenticatorTypes.BASIC_AUTH,

  // Google authentication flows
  'default-google-flow': 'google',
  google: 'google',

  // GitHub authentication flows
  'default-github-flow': 'github',
  github: 'github',

  // SMS OTP flows
  'default-sms-otp-flow': 'sms-otp',
  'sms-otp': 'sms-otp',
  'sms-otp-with-username-flow': 'sms-otp',

  // Combined flows - map to primary integration type
  'basic-google-flow': AuthenticatorTypes.BASIC_AUTH,
  'basic-google-github-flow': AuthenticatorTypes.BASIC_AUTH,
  'basic-google-github-sms-flow': AuthenticatorTypes.BASIC_AUTH,
} as const;

/**
 * Integration types that flows can map to
 *
 * @public
 */
export type FlowIntegrationType = typeof AuthenticatorTypes.BASIC_AUTH | 'google' | 'github' | 'sms-otp';

export {FLOW_HANDLE_TO_INTEGRATION_MAP};
export default FLOW_HANDLE_TO_INTEGRATION_MAP;
