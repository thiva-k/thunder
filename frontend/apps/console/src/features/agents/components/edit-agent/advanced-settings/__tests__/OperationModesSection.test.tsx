// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type {OAuth2Config} from '@thunderid/configure-applications';
import {describe, it, expect, vi} from 'vitest';
import OperationModesSection from '../OperationModesSection';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string) => fallback ?? key,
  }),
}));

describe('OperationModesSection', () => {
  const autonomousOnlyConfig: OAuth2Config = {
    grantTypes: ['client_credentials'],
    responseTypes: [],
  };

  const delegatedConfig: OAuth2Config = {
    grantTypes: ['client_credentials', 'authorization_code'],
    responseTypes: ['code'],
    redirectUris: ['https://example.com/cb'],
  };

  it('returns null when oauth2Config is undefined', () => {
    const {container} = render(<OperationModesSection />);
    expect(container.firstChild).toBeNull();
  });

  it('renders the card title and description', () => {
    render(<OperationModesSection oauth2Config={autonomousOnlyConfig} />);

    expect(screen.getByText('OAuth Configuration')).toBeInTheDocument();
  });

  it('shows an info note explaining the Delegated-mode dependency, before the grant types selector', () => {
    render(<OperationModesSection oauth2Config={autonomousOnlyConfig} onOAuth2ConfigChange={vi.fn()} />);

    expect(
      screen.getByText(/greyed-out grants unlock once you turn on Delegated mode at the top of this tab/i),
    ).toBeInTheDocument();
  });

  describe('Grant type selection', () => {
    it('lists every grant type in Autonomous-only mode, but locks the delegated-only ones', async () => {
      const user = userEvent.setup();
      render(<OperationModesSection oauth2Config={autonomousOnlyConfig} onOAuth2ConfigChange={vi.fn()} />);

      await user.click(document.getElementById('agent-grant-types')!);

      const listbox = await screen.findByRole('listbox');
      expect(within(listbox).getByText('client_credentials')).toBeInTheDocument();
      expect(within(listbox).getByText('Token Exchange')).toBeInTheDocument();
      expect(within(listbox).getByText('authorization_code')).toBeInTheDocument();
      expect(within(listbox).getByText('CIBA (Client-Initiated Backchannel Authentication)')).toBeInTheDocument();
      expect(within(listbox).getByText('refresh_token')).toBeInTheDocument();

      expect(within(listbox).getByText('authorization_code').closest('li')).toHaveAttribute('aria-disabled', 'true');
      expect(
        within(listbox).getByText('CIBA (Client-Initiated Backchannel Authentication)').closest('li'),
      ).toHaveAttribute('aria-disabled', 'true');
      expect(within(listbox).getByText('refresh_token').closest('li')).toHaveAttribute('aria-disabled', 'true');
    });

    it('unlocks ciba and refresh_token (but keeps authorization_code locked) once Delegated mode is on', async () => {
      const user = userEvent.setup();
      render(<OperationModesSection oauth2Config={delegatedConfig} onOAuth2ConfigChange={vi.fn()} />);

      await user.click(document.getElementById('agent-grant-types')!);

      const listbox = await screen.findByRole('listbox');
      expect(within(listbox).getByText('authorization_code').closest('li')).toHaveAttribute('aria-disabled', 'true');
      expect(
        within(listbox).getByText('CIBA (Client-Initiated Backchannel Authentication)').closest('li'),
      ).not.toHaveAttribute('aria-disabled', 'true');
      expect(within(listbox).getByText('refresh_token').closest('li')).not.toHaveAttribute('aria-disabled', 'true');
    });

    it('locks client_credentials so it cannot be toggled off', async () => {
      const user = userEvent.setup();
      render(<OperationModesSection oauth2Config={autonomousOnlyConfig} onOAuth2ConfigChange={vi.fn()} />);

      await user.click(document.getElementById('agent-grant-types')!);
      const listbox = await screen.findByRole('listbox');

      expect(within(listbox).getByText('client_credentials').closest('li')).toHaveAttribute('aria-disabled', 'true');
    });

    it('toggles token_exchange on when selected', async () => {
      const user = userEvent.setup();
      const onOAuth2ConfigChange = vi.fn();
      render(<OperationModesSection oauth2Config={autonomousOnlyConfig} onOAuth2ConfigChange={onOAuth2ConfigChange} />);

      await user.click(document.getElementById('agent-grant-types')!);
      const listbox = await screen.findByRole('listbox');
      await user.click(within(listbox).getByText('Token Exchange'));

      expect(onOAuth2ConfigChange).toHaveBeenCalledWith(
        expect.objectContaining({
          grantTypes: expect.arrayContaining([
            'client_credentials',
            'urn:ietf:params:oauth:grant-type:token-exchange',
          ]) as string[],
        }),
      );
    });

    it('toggles ciba on within delegated grants', async () => {
      const user = userEvent.setup();
      const onOAuth2ConfigChange = vi.fn();
      render(<OperationModesSection oauth2Config={delegatedConfig} onOAuth2ConfigChange={onOAuth2ConfigChange} />);

      await user.click(document.getElementById('agent-grant-types')!);
      const listbox = await screen.findByRole('listbox');
      await user.click(within(listbox).getByText('CIBA (Client-Initiated Backchannel Authentication)'));

      expect(onOAuth2ConfigChange).toHaveBeenCalledWith(
        expect.objectContaining({
          grantTypes: expect.arrayContaining(['authorization_code', 'urn:openid:params:grant-type:ciba']) as string[],
        }),
      );
    });
  });

  describe('Redirect URIs', () => {
    it('renders the redirect URI section once authorization_code is selected', () => {
      render(<OperationModesSection oauth2Config={delegatedConfig} onOAuth2ConfigChange={vi.fn()} />);

      expect(screen.getByText('Authorized redirect URIs')).toBeInTheDocument();
    });

    it('hides the redirect URI section in Autonomous-only mode', () => {
      render(<OperationModesSection oauth2Config={autonomousOnlyConfig} onOAuth2ConfigChange={vi.fn()} />);

      expect(screen.queryByText('Authorized redirect URIs')).not.toBeInTheDocument();
    });
  });

  describe('read-only', () => {
    it('disables the grant types input when there is no onOAuth2ConfigChange handler', () => {
      render(<OperationModesSection oauth2Config={delegatedConfig} />);

      expect(document.getElementById('agent-grant-types')).toHaveAttribute('aria-disabled', 'true');
    });

    it('disables the grant types input when disabled is true', () => {
      render(<OperationModesSection oauth2Config={delegatedConfig} onOAuth2ConfigChange={vi.fn()} disabled />);

      expect(document.getElementById('agent-grant-types')).toHaveAttribute('aria-disabled', 'true');
    });
  });
});
