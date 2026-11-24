// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

window.__IAM_RUNTIME_CONFIG__ = {
  brand: {
    product_name: 'Choreo Identity & Access Management',
    favicon: {
      light: 'assets/images/favicon.ico',
      dark: 'assets/images/favicon-inverted.ico',
    },
  },
  client: {
    base: '/gate',
  },
  documentation: {
    links: {},
  },
  // Defaults to the origin this app is served from. Add a `server` block with `public_url`
  // (or `hostname`, `port`, `http_only`) to target a different backend.
};
