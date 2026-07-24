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

import userEvent from '@testing-library/user-event';
import {render, screen} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import type {InboundAuthConfig} from '../../../../models/inbound-auth';
import ClientAccessTokenSection, {type ClientAccessTokenCopy} from '../ClientAccessTokenSection';

vi.mock('../JwtPreview', () => ({
  default: ({payload}: {payload: Record<string, unknown>}) => (
    <pre data-testid="jwt-preview">{JSON.stringify(payload)}</pre>
  ),
}));

const copy: ClientAccessTokenCopy = {
  attributesTitle: 'Access Token Claims',
  attributesDescription: 'Optional claims for the client access token.',
  attributesLabel: 'Add or Remove Claims',
  attributesHint: 'Click a claim to include it.',
  attributesEmpty: 'No optional claims available.',
  validityTitle: 'Token Validity',
  validityDescription: 'How long the token is valid.',
  validityLabel: 'Token Validity',
  validityHint: 'Token validity period in seconds.',
  validityError: 'Enter a validity period of at least 1 second.',
};

describe('ClientAccessTokenSection', () => {
  const onFieldChange = vi.fn();
  const inboundAuthConfig: InboundAuthConfig[] = [
    {type: 'oauth2', config: {grantTypes: ['client_credentials'], responseTypes: []}},
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the supplied claims as selectable chips', () => {
    render(
      <ClientAccessTokenSection
        oauth2Config={{grantTypes: ['client_credentials'], responseTypes: []}}
        inboundAuthConfig={inboundAuthConfig}
        onFieldChange={onFieldChange}
        availableAttributes={['groups', 'roles']}
        copy={copy}
      />,
    );

    expect(screen.getByText('groups')).toBeInTheDocument();
    expect(screen.getByText('roles')).toBeInTheDocument();
  });

  it('adds a claim to clientConfig when its chip is clicked', async () => {
    const user = userEvent.setup();
    render(
      <ClientAccessTokenSection
        oauth2Config={{grantTypes: ['client_credentials'], responseTypes: []}}
        inboundAuthConfig={inboundAuthConfig}
        onFieldChange={onFieldChange}
        availableAttributes={['groups', 'roles']}
        copy={copy}
      />,
    );

    await user.click(screen.getByText('roles'));

    expect(onFieldChange).toHaveBeenCalledWith(
      'inboundAuthConfig',
      expect.arrayContaining([
        expect.objectContaining({
          type: 'oauth2',
          config: expect.objectContaining({
            token: expect.objectContaining({
              accessToken: expect.objectContaining({
                clientConfig: expect.objectContaining({attributes: ['roles']}) as Record<string, unknown>,
              }) as Record<string, unknown>,
            }) as Record<string, unknown>,
          }) as Record<string, unknown>,
        }),
      ]),
    );
  });

  it('defaults the validity to 3600 and commits a valid change', async () => {
    const user = userEvent.setup();
    render(
      <ClientAccessTokenSection
        oauth2Config={{grantTypes: ['client_credentials'], responseTypes: []}}
        inboundAuthConfig={inboundAuthConfig}
        onFieldChange={onFieldChange}
        availableAttributes={['groups']}
        copy={copy}
        inputId="client-access-token-validity"
      />,
    );

    const input = document.getElementById('client-access-token-validity')!;
    expect(input).toHaveValue(3600);

    await user.clear(input);
    await user.type(input, '7200');

    expect(onFieldChange).toHaveBeenLastCalledWith(
      'inboundAuthConfig',
      expect.arrayContaining([
        expect.objectContaining({
          config: expect.objectContaining({
            token: expect.objectContaining({
              accessToken: expect.objectContaining({
                clientConfig: expect.objectContaining({validityPeriod: 7200}) as Record<string, unknown>,
              }) as Record<string, unknown>,
            }) as Record<string, unknown>,
          }) as Record<string, unknown>,
        }),
      ]),
    );
  });

  it('reports a validation error without committing when the validity is cleared', async () => {
    const user = userEvent.setup();
    const onValidationChange = vi.fn();
    render(
      <ClientAccessTokenSection
        oauth2Config={{grantTypes: ['client_credentials'], responseTypes: []}}
        inboundAuthConfig={inboundAuthConfig}
        onFieldChange={onFieldChange}
        availableAttributes={['groups']}
        copy={copy}
        inputId="client-access-token-validity"
        onValidationChange={onValidationChange}
      />,
    );

    onFieldChange.mockClear();
    await user.clear(document.getElementById('client-access-token-validity')!);

    expect(screen.getByText('Enter a validity period of at least 1 second.')).toBeInTheDocument();
    expect(onValidationChange).toHaveBeenLastCalledWith(true);
    expect(onFieldChange).not.toHaveBeenCalled();
  });

  it('disables the validity input when disabled', () => {
    render(
      <ClientAccessTokenSection
        oauth2Config={{grantTypes: ['client_credentials'], responseTypes: []}}
        inboundAuthConfig={inboundAuthConfig}
        onFieldChange={onFieldChange}
        availableAttributes={['groups']}
        copy={copy}
        inputId="client-access-token-validity"
        disabled
      />,
    );

    expect(document.getElementById('client-access-token-validity')!).toBeDisabled();
  });
});
