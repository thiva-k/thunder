// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

window.__IAM_RUNTIME_CONFIG__ = {
  brand: {
    product_name: 'Choreo Identity & Access Management',
    favicon: {
      light: 'assets/images/favicon.ico',
      dark: 'assets/images/favicon-inverted.ico',
    },
  },
  documentation: {
    links: {
      users: '',
    },
  },
  client: {
    base: '/console',
    client_id: 'CONSOLE',
    resource_identifier: 'https://localhost:8090/mcp',
    scopes: ['openid', 'profile', 'email', 'ou', 'system'],
  },
  // Defaults to the origin this app is served from. Add a `server` block with `public_url`
  // (or `hostname`, `port`, `http_only`) to target a different backend.

  // Optional: location of the login gate, used to build the OAuth redirect URI shown when
  // configuring social/OIDC connections. Omit to default to `${served origin}/gate/callback`.
  // gate_client: {
  //   public_url: 'https://gate.example.com',   // or hostname/port/scheme
  // },
};
