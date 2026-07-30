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

import {describe, it, expect} from 'vitest';
import buildMfaPreviewMock from '../buildMfaPreviewMock';

type MockComponent = Record<string, unknown>;

const getComponentById = (components: MockComponent[], id: string): MockComponent | undefined => {
  for (const component of components) {
    if (component.id === id) {
      return component;
    }
    const nested = component.components as MockComponent[] | undefined;
    const found = nested && getComponentById(nested, id);
    if (found) {
      return found;
    }
  }
  return undefined;
};

describe('buildMfaPreviewMock', () => {
  it('should render the OTP-entry screen when only Email OTP is enabled', () => {
    const result = buildMfaPreviewMock({hasEmailOtpMfa: true, hasSmsOtpMfa: false}) as unknown as MockComponent[];

    expect(getComponentById(result, 'otp_input_mfa')).toBeDefined();
    expect(getComponentById(result, 'action_mfa_verify')).toBeDefined();
    expect(getComponentById(result, 'action_mfa_resend')).toBeDefined();
    expect(getComponentById(result, 'text_mfa_desc')!.label).toContain('email');

    // Not the channel-choice screen
    expect(getComponentById(result, 'action_mfa_choose_email')).toBeUndefined();
  });

  it('should render the OTP-entry screen when only SMS OTP is enabled', () => {
    const result = buildMfaPreviewMock({hasEmailOtpMfa: false, hasSmsOtpMfa: true}) as unknown as MockComponent[];

    expect(getComponentById(result, 'otp_input_mfa')).toBeDefined();
    expect(getComponentById(result, 'text_mfa_desc')!.label).toContain('phone');
  });

  it('should render the channel-choice screen when both Email and SMS OTP are enabled', () => {
    const result = buildMfaPreviewMock({hasEmailOtpMfa: true, hasSmsOtpMfa: true}) as unknown as MockComponent[];

    expect(getComponentById(result, 'action_mfa_choose_email')).toBeDefined();
    expect(getComponentById(result, 'action_mfa_choose_sms')).toBeDefined();

    // Not the direct OTP-entry screen
    expect(getComponentById(result, 'otp_input_mfa')).toBeUndefined();
  });
});
