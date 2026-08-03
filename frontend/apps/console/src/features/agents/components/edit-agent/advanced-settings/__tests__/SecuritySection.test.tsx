// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type {OAuth2Config} from '@thunderid/configure-applications';
import {describe, it, expect, vi} from 'vitest';
import SecuritySection from '../SecuritySection';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
  Trans: ({defaults = ''}: {defaults?: string}) => <span>{defaults}</span>,
}));

describe('SecuritySection (agent)', () => {
  it('returns null when oauth2Config is undefined', () => {
    const {container} = render(<SecuritySection />);
    expect(container.firstChild).toBeNull();
  });

  it('checks PKCE automatically when authorization_code is selected', () => {
    const oauth2Config: OAuth2Config = {
      grantTypes: ['authorization_code'],
      responseTypes: ['code'],
      pkceRequired: false,
      publicClient: false,
    };

    render(<SecuritySection oauth2Config={oauth2Config} />);

    const pkceSwitch = screen.getByLabelText('agents:edit.advanced.security.pkce.label');
    expect(pkceSwitch).toBeChecked();
  });

  it('unchecks PKCE when authorization_code is not selected', () => {
    const oauth2Config: OAuth2Config = {
      grantTypes: ['client_credentials'],
      responseTypes: [],
      pkceRequired: false,
      publicClient: false,
    };

    render(<SecuritySection oauth2Config={oauth2Config} />);

    const pkceSwitch = screen.getByLabelText('agents:edit.advanced.security.pkce.label');
    expect(pkceSwitch).not.toBeChecked();
  });

  it('checks PKCE when the client is public even without authorization_code', () => {
    const oauth2Config: OAuth2Config = {
      grantTypes: ['client_credentials'],
      responseTypes: [],
      pkceRequired: false,
      publicClient: true,
    };

    render(<SecuritySection oauth2Config={oauth2Config} />);

    const pkceSwitch = screen.getByLabelText('agents:edit.advanced.security.pkce.label');
    expect(pkceSwitch).toBeChecked();
  });

  it('is never directly editable by the user', () => {
    const oauth2Config: OAuth2Config = {
      grantTypes: ['authorization_code'],
      responseTypes: ['code'],
      pkceRequired: true,
      publicClient: false,
    };

    render(<SecuritySection oauth2Config={oauth2Config} />);

    expect(screen.getByLabelText('agents:edit.advanced.security.pkce.label')).toBeDisabled();
  });

  it('shows the not-applicable caption with the authorization_code grant called out when the grant is off', () => {
    const oauth2Config: OAuth2Config = {
      grantTypes: ['client_credentials'],
      responseTypes: [],
      pkceRequired: false,
      publicClient: false,
    };

    render(<SecuritySection oauth2Config={oauth2Config} />);

    expect(screen.getByText(/PKCE only applies to the/)).toBeInTheDocument();
    const pkceSwitch = screen.getByLabelText('agents:edit.advanced.security.pkce.label');
    expect(pkceSwitch).toBeDisabled();
  });

  describe('Pushed Authorization Requests', () => {
    const oauth2Config: OAuth2Config = {
      grantTypes: ['authorization_code'],
      responseTypes: ['code'],
      pkceRequired: true,
      publicClient: false,
    };

    it('reflects the configured value', () => {
      render(<SecuritySection oauth2Config={{...oauth2Config, requirePushedAuthorizationRequests: true}} />);

      expect(screen.getByLabelText('agents:edit.advanced.security.par.label')).toBeChecked();
    });

    it('treats an unset value as unchecked', () => {
      render(<SecuritySection oauth2Config={oauth2Config} />);

      expect(screen.getByLabelText('agents:edit.advanced.security.par.label')).not.toBeChecked();
    });

    it('calls onOAuth2ConfigChange when toggled', async () => {
      const user = userEvent.setup();
      const onOAuth2ConfigChange = vi.fn();
      render(<SecuritySection oauth2Config={oauth2Config} onOAuth2ConfigChange={onOAuth2ConfigChange} />);

      await user.click(screen.getByLabelText('agents:edit.advanced.security.par.label'));

      expect(onOAuth2ConfigChange).toHaveBeenCalledWith({requirePushedAuthorizationRequests: true});
    });

    it('is disabled when the section is read-only', () => {
      render(<SecuritySection oauth2Config={oauth2Config} onOAuth2ConfigChange={vi.fn()} disabled />);

      expect(screen.getByLabelText('agents:edit.advanced.security.par.label')).toBeDisabled();
    });

    it('is disabled when there is no onOAuth2ConfigChange handler', () => {
      render(<SecuritySection oauth2Config={oauth2Config} />);

      expect(screen.getByLabelText('agents:edit.advanced.security.par.label')).toBeDisabled();
    });

    it('is disabled and unchecked when the authorization_code grant is off', () => {
      render(
        <SecuritySection
          oauth2Config={{
            grantTypes: ['client_credentials'],
            responseTypes: [],
            requirePushedAuthorizationRequests: true,
          }}
          onOAuth2ConfigChange={vi.fn()}
        />,
      );

      const parSwitch = screen.getByLabelText('agents:edit.advanced.security.par.label');
      expect(parSwitch).toBeDisabled();
      expect(parSwitch).not.toBeChecked();
    });
  });
});
