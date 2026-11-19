// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useMutation, type UseMutationResult} from '@tanstack/react-query';
import {useConfig, useToast} from '@thunderid/contexts';
import {useThunderID} from '@thunderid/react';
import {FlowExecutionFailure, type FlowExecutionResponse} from '@thunderid/utils';
import {useTranslation} from 'react-i18next';

/** The administrator-initiated flow that sends an agent's external owner a link, by email or SMS, to verify it. */
export const AGENT_OWNER_VERIFICATION_FLOW_HANDLE = 'default-agent-owner-verification-flow';

export interface SendOwnerVerificationVariables {
  agentId: string;
}

/**
 * Starts the agent owner verification flow for an agent. The flow sends the agent's owner a link; the
 * owner is verified once they open it and confirm, so nothing about the agent changes here. A failure is left to
 * the caller to show where the administrator sent the link from.
 */
export default function useSendOwnerVerification(): UseMutationResult<
  FlowExecutionResponse,
  Error,
  SendOwnerVerificationVariables
> {
  const {http} = useThunderID();
  const {getServerUrl} = useConfig();
  const {t} = useTranslation('agents');
  const {showToast} = useToast();

  return useMutation<FlowExecutionResponse, Error, SendOwnerVerificationVariables>({
    mutationFn: async ({agentId}: SendOwnerVerificationVariables): Promise<FlowExecutionResponse> => {
      const response: {data: FlowExecutionResponse} = await http.request({
        url: `${getServerUrl()}/flow/execute`,
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        data: {
          flowType: 'ADMIN_INITIATED',
          handle: AGENT_OWNER_VERIFICATION_FLOW_HANDLE,
          inputs: {subject: agentId},
          // The owner continues the flow on Gate, which draws each step from its meta. A flow keeps the
          // verbosity it starts with, so it has to start verbose for the owner's steps to carry meta.
          verbose: true,
        },
      } as unknown as Parameters<typeof http.request>[0]);

      if (response.data.flowStatus === 'ERROR') {
        throw new FlowExecutionFailure(response.data.error ?? {}, 'The verification link could not be sent.');
      }

      return response.data;
    },
    onSuccess: (data) => {
      const destination = data.data?.additionalData?.['maskedDestination'];
      showToast(
        destination
          ? t('owner.verification.sent', 'Verification link sent to {{destination}}', {destination})
          : t('owner.verification.sentNoDestination', 'Verification link sent to the owner'),
        'success',
      );
    },
  });
}
