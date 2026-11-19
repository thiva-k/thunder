// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useMutation, type UseMutationResult} from '@tanstack/react-query';
import {useConfig, useToast} from '@thunderid/contexts';
import {useThunderID} from '@thunderid/react';
import {FlowExecutionFailure, type FlowExecutionResponse} from '@thunderid/utils';
import {useTranslation} from 'react-i18next';

export interface SendOwnerVerificationVariables {
  agentId: string;
  /**
   * Handle of the flow to start, from the server's `agentOwnerVerificationFlow` setting.
   */
  handle: string;
  /**
   * The owner the link goes to, as the page already shows it. The flow answers with a masked destination, which
   * hides nothing from an administrator who can read the agent, so the confirmation names the owner instead.
   */
  owner?: string;
}

/**
 * Starts the agent owner verification flow for an agent. The flow sends the agent's owner a link, by email or
 * SMS; the owner is verified once they open it and confirm, so nothing about the agent changes here. A failure
 * is left to the caller to show where the administrator sent the link from.
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
    mutationFn: async ({agentId, handle}: SendOwnerVerificationVariables): Promise<FlowExecutionResponse> => {
      const response: {data: FlowExecutionResponse} = await http.request({
        url: `${getServerUrl()}/flow/execute`,
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        data: {
          flowType: 'ADMIN_INITIATED',
          handle,
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
    onSuccess: (_data, {owner}) => {
      showToast(
        owner
          ? t('owner.verification.sent', 'Verification link sent to {{destination}}', {destination: owner})
          : t('owner.verification.sentNoDestination', 'Verification link sent to the owner'),
        'success',
      );
    },
  });
}
