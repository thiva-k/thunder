// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useMutation, useQueryClient, type UseMutationResult} from '@tanstack/react-query';
import {VerifiablePresentationQueryKeys} from '@thunderid/configure-verifiable-presentations';
import type {VerifiablePresentation} from '@thunderid/configure-verifiable-presentations';
import {useConfig, useToast} from '@thunderid/contexts';
import {useThunderID} from '@thunderid/react';
import {useTranslation} from 'react-i18next';
import type {UpdateVerifiablePresentationRequest} from '../models/requests';

interface UpdateArgs {
  id: string;
  data: UpdateVerifiablePresentationRequest;
}

/**
 * Updates an existing OpenID4VP presentation definition.
 */
export default function useUpdateVerifiablePresentation(): UseMutationResult<
  VerifiablePresentation,
  Error,
  UpdateArgs
> {
  const {http} = useThunderID();
  const {getServerUrl} = useConfig();
  const queryClient: ReturnType<typeof useQueryClient> = useQueryClient();
  const {t} = useTranslation('verifiable-presentations');
  const {showToast} = useToast();

  return useMutation<VerifiablePresentation, Error, UpdateArgs>({
    mutationFn: async ({id, data}: UpdateArgs): Promise<VerifiablePresentation> => {
      const serverUrl: string = getServerUrl();
      const response: {data: VerifiablePresentation} = await http.request({
        url: `${serverUrl}/openid4vp/presentation-definitions/${id}`,
        method: 'PUT',
        headers: {'Content-Type': 'application/json'},
        data,
      } as unknown as Parameters<typeof http.request>[0]);

      return response.data;
    },
    onSuccess: (_result, {id}: UpdateArgs) => {
      queryClient.invalidateQueries({queryKey: [VerifiablePresentationQueryKeys.VPS]}).catch(() => {
        /* noop */
      });
      queryClient.invalidateQueries({queryKey: [VerifiablePresentationQueryKeys.VP, id]}).catch(() => {
        /* noop */
      });
      showToast(t('update.success'), 'success');
    },
    onError: () => {
      showToast(t('update.error'), 'error');
    },
  });
}
