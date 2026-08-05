// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import userEvent from '@testing-library/user-event';
import type {Application, OAuth2Config} from '@thunderid/configure-applications';
import {render, screen} from '@thunderid/test-utils';
import {useState} from 'react';
import {describe, it, expect, vi} from 'vitest';
import EditTokenSettingsTabs from '../EditTokenSettingsTabs';

vi.mock('../ClientAccessTokenSection', () => ({
  default: vi.fn(() => {
    const [clicks, setClicks] = useState(0);
    return (
      <div data-testid="client-token-section">
        Clicks: {clicks}
        <button type="button" data-testid="client-token-section-bump" onClick={() => setClicks((c) => c + 1)}>
          Bump
        </button>
      </div>
    );
  }),
}));
vi.mock('../EditTokenSettings', () => ({
  default: vi.fn(({sectionResetKey}: {sectionResetKey?: number}) => {
    const [clicks, setClicks] = useState(0);
    return (
      <div data-testid="user-token-section">
        Clicks: {clicks}, Key: {sectionResetKey}
        <button type="button" data-testid="user-token-section-bump" onClick={() => setClicks((c) => c + 1)}>
          Bump
        </button>
      </div>
    );
  }),
}));

const application: Application = {id: 'app-1', name: 'Test App'};
const clientLockMessage = /does not receive tokens for itself/i;
const userLockMessage = /does not receive tokens for signed-in users/i;

/**
 * Builds an OAuth2 config shaped like one loaded from the API, where the backend has always
 * materialized the access token and ID token blocks.
 */
const oauthConfig = (grantTypes: string[]): OAuth2Config => ({
  grantTypes,
  responseTypes: [],
  token: {
    accessToken: {userConfig: {validityPeriod: 3600, attributes: []}},
    idToken: {validityPeriod: 3600, userAttributes: []},
  },
});

describe('EditTokenSettingsTabs', () => {
  const onFieldChange = vi.fn();

  it('renders Application and User audiences in a side selector', () => {
    render(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={oauthConfig(['client_credentials'])}
        onFieldChange={onFieldChange}
      />,
    );

    expect(screen.getByRole('tablist', {name: 'Issued to'})).toBeInTheDocument();
    expect(screen.getByRole('tab', {name: 'Application'})).toBeInTheDocument();
    expect(screen.getByRole('tab', {name: 'User'})).toBeInTheDocument();
    expect(screen.getByText('M2M access token')).toBeInTheDocument();
    expect(screen.getByText('Tokens for a signed-in user')).toBeInTheDocument();
    expect(screen.getByText(/configured independently for each audience/i)).toBeInTheDocument();
  });

  it('unlocks the Application tab when client_credentials is granted', () => {
    render(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={oauthConfig(['client_credentials'])}
        onFieldChange={onFieldChange}
      />,
    );

    expect(screen.getByRole('tab', {name: 'Application'})).toBeEnabled();
    expect(screen.getByTestId('client-token-section')).toBeInTheDocument();
    expect(screen.queryByText(clientLockMessage)).not.toBeInTheDocument();
  });

  it('shows the Application notice in place of its settings when client_credentials is not granted', async () => {
    const user = userEvent.setup();
    render(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={oauthConfig(['authorization_code'])}
        onFieldChange={onFieldChange}
      />,
    );

    // The notice belongs to the Application sub-tab, so it stays out of the way until opened.
    expect(screen.queryByText(clientLockMessage)).not.toBeInTheDocument();

    await user.click(screen.getByRole('tab', {name: 'Application'}));

    expect(screen.getByText(clientLockMessage)).toBeInTheDocument();
    expect(screen.queryByTestId('client-token-section')).not.toBeInTheDocument();
  });

  it('shows the User notice in place of its settings when no user-facing grant is present', async () => {
    const user = userEvent.setup();
    render(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={oauthConfig(['client_credentials'])}
        onFieldChange={onFieldChange}
      />,
    );

    expect(screen.queryByText(userLockMessage)).not.toBeInTheDocument();

    await user.click(screen.getByRole('tab', {name: 'User'}));

    expect(screen.getByText(userLockMessage)).toBeInTheDocument();
    expect(screen.queryByTestId('user-token-section')).not.toBeInTheDocument();
  });

  it('keeps a sub-tab with no applicable settings selectable', () => {
    render(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={oauthConfig(['authorization_code'])}
        onFieldChange={onFieldChange}
      />,
    );

    expect(screen.getByRole('tab', {name: 'Application'})).toBeEnabled();
  });

  it('unlocks the User tab when a user-facing grant is present', async () => {
    const user = userEvent.setup();
    render(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={oauthConfig(['authorization_code'])}
        onFieldChange={onFieldChange}
      />,
    );

    await user.click(screen.getByRole('tab', {name: 'User'}));

    expect(screen.getByTestId('user-token-section')).toBeInTheDocument();
    expect(screen.queryByText(userLockMessage)).not.toBeInTheDocument();
  });

  it('selects the User tab by default when the Application tab is locked', () => {
    render(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={oauthConfig(['authorization_code'])}
        onFieldChange={onFieldChange}
      />,
    );

    expect(screen.getByRole('tab', {name: 'User'})).toHaveAttribute('aria-selected', 'true');
    expect(screen.getByTestId('user-token-section')).toBeInTheDocument();
  });

  describe('app-native application (no OAuth2 configuration)', () => {
    it('unlocks the User tab so the assertion config can be edited', () => {
      render(
        <EditTokenSettingsTabs
          application={application}
          editedApp={{}}
          oauth2Config={undefined}
          onFieldChange={onFieldChange}
        />,
      );

      expect(screen.getByRole('tab', {name: 'User'})).toBeEnabled();
      expect(screen.getByTestId('user-token-section')).toBeInTheDocument();
      expect(screen.queryByText(userLockMessage)).not.toBeInTheDocument();
    });

    it('opens on the User sub-tab and keeps the Application notice behind its own tab', async () => {
      const user = userEvent.setup();
      render(
        <EditTokenSettingsTabs
          application={application}
          editedApp={{}}
          oauth2Config={undefined}
          onFieldChange={onFieldChange}
        />,
      );

      expect(screen.getByRole('tab', {name: 'User'})).toHaveAttribute('aria-selected', 'true');
      expect(screen.queryByText(clientLockMessage)).not.toBeInTheDocument();

      await user.click(screen.getByRole('tab', {name: 'Application'}));

      expect(screen.getByText(clientLockMessage)).toBeInTheDocument();
    });
  });

  it('remounts ClientAccessTokenSection when sectionResetKey changes', async () => {
    const user = userEvent.setup();
    const {rerender} = render(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={{grantTypes: ['client_credentials'], responseTypes: []}}
        onFieldChange={onFieldChange}
        sectionResetKey={0}
      />,
    );

    await user.click(screen.getByTestId('client-token-section-bump'));
    expect(screen.getByTestId('client-token-section')).toHaveTextContent('Clicks: 1');

    rerender(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={{grantTypes: ['client_credentials'], responseTypes: []}}
        onFieldChange={onFieldChange}
        sectionResetKey={1}
      />,
    );

    // A fresh key means a fresh mount, so the local click count is gone.
    expect(screen.getByTestId('client-token-section')).toHaveTextContent('Clicks: 0');
  });

  it('does not remount EditTokenSettings, but forwards the updated sectionResetKey, when it changes', async () => {
    const user = userEvent.setup();
    const {rerender} = render(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={{grantTypes: ['authorization_code'], responseTypes: []}}
        onFieldChange={onFieldChange}
        sectionResetKey={0}
      />,
    );

    await user.click(screen.getByRole('tab', {name: 'User'}));
    await user.click(screen.getByTestId('user-token-section-bump'));
    expect(screen.getByTestId('user-token-section')).toHaveTextContent('Clicks: 1');

    rerender(
      <EditTokenSettingsTabs
        application={application}
        editedApp={{}}
        oauth2Config={{grantTypes: ['authorization_code'], responseTypes: []}}
        onFieldChange={onFieldChange}
        sectionResetKey={1}
      />,
    );

    // EditTokenSettings resets its own form in place on a new key, so it must not remount here
    // (the click count survives) while still receiving the updated key as a prop.
    expect(screen.getByTestId('user-token-section')).toHaveTextContent('Clicks: 1, Key: 1');
  });
});
