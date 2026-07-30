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

export interface MfaPreviewMockOptions {
  hasEmailOtpMfa: boolean;
  hasSmsOtpMfa: boolean;
}

/**
 * Builds a preview mock for the screen a user sees right after their first sign-in factor
 * succeeds, when MFA is enabled. Mirrors the exact node shape `generateFlowGraph` produces for
 * the wizard's own generated flow (see `buildMfaChannelChain`/`mfa_choose_channel`), so this is a
 * faithful "what happens next" preview rather than a generic placeholder.
 *
 * When both Email and SMS OTP are enabled, the real next screen is the channel-choice prompt, so
 * that's what's previewed here rather than either OTP-entry screen individually.
 */
export default function buildMfaPreviewMock({
  hasEmailOtpMfa,
  hasSmsOtpMfa,
}: MfaPreviewMockOptions): EmbeddedFlowComponent[] {
  if (hasEmailOtpMfa && hasSmsOtpMfa) {
    return [
      {
        align: 'center',
        category: 'DISPLAY',
        id: 'text_mfa_choose_heading',
        label: 'Verify your identity',
        resourceType: 'ELEMENT',
        type: 'TEXT',
        variant: 'HEADING_3',
      },
      {
        category: 'BLOCK',
        components: [
          {
            category: 'ACTION',
            eventType: 'TRIGGER',
            id: 'action_mfa_choose_email',
            label: 'Email me a code',
            resourceType: 'ELEMENT',
            type: 'ACTION',
            variant: 'PRIMARY',
          },
          {
            category: 'ACTION',
            eventType: 'TRIGGER',
            id: 'action_mfa_choose_sms',
            label: 'Text me a code',
            resourceType: 'ELEMENT',
            type: 'ACTION',
            variant: 'SECONDARY',
          },
        ],
        id: 'block_mfa_choose',
        resourceType: 'ELEMENT',
        type: 'BLOCK',
      },
    ] as unknown as EmbeddedFlowComponent[];
  }

  const description = hasSmsOtpMfa ? 'Enter the code sent to your phone' : 'Enter the code sent to your email';

  return [
    {
      align: 'center',
      category: 'DISPLAY',
      id: 'text_mfa_heading',
      label: 'Verify your identity',
      resourceType: 'ELEMENT',
      type: 'TEXT',
      variant: 'HEADING_3',
    },
    {
      align: 'center',
      category: 'DISPLAY',
      id: 'text_mfa_desc',
      label: description,
      resourceType: 'ELEMENT',
      type: 'TEXT',
      variant: 'BODY',
    },
    {
      category: 'BLOCK',
      components: [
        {
          category: 'FIELD',
          hint: '',
          id: 'otp_input_mfa',
          inputType: 'text',
          label: 'One-time code',
          placeholder: '',
          ref: 'otp',
          required: true,
          resourceType: 'ELEMENT',
          type: 'OTP_INPUT',
        },
        {
          category: 'ACTION',
          eventType: 'SUBMIT',
          id: 'action_mfa_verify',
          label: 'Verify',
          resourceType: 'ELEMENT',
          type: 'ACTION',
          variant: 'PRIMARY',
        },
        {
          category: 'ACTION',
          eventType: 'SUBMIT',
          id: 'action_mfa_resend',
          label: 'Resend code',
          resourceType: 'ELEMENT',
          type: 'RESEND',
        },
      ],
      id: 'block_mfa',
      resourceType: 'ELEMENT',
      type: 'BLOCK',
    },
  ] as unknown as EmbeddedFlowComponent[];
}
