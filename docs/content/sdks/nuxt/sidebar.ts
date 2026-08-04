// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebar: SidebarsConfig = {
  nuxtSdkSidebar: [
    {
      type: 'doc',
      id: 'sdks/nuxt/overview',
    },
    {
      type: 'category',
      label: 'APIs',
      collapsed: false,
      items: [
        {
          type: 'category',
          label: 'Configuration',
          collapsed: false,
          items: [
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/configuration/module-configuration',
              label: 'Module Configuration',
            },
          ],
        },
        {
          type: 'category',
          label: 'Composables',
          collapsed: false,
          items: [
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/composables/use-thunderid',
              label: 'useThunderID()',
            },
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/composables/use-user',
              label: 'useUser()',
            },
          ],
        },
        {
          type: 'category',
          label: 'Components',
          collapsed: false,
          items: [
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/components/thunderid-root',
              label: '<ThunderIDRoot />',
            },
            {
              type: 'category',
              label: 'Control Components',
              collapsed: true,
              items: [
                {
                  type: 'doc',
                  id: 'sdks/nuxt/apis/components/signed-in',
                  label: '<ThunderIDSignedIn />',
                },
                {
                  type: 'doc',
                  id: 'sdks/nuxt/apis/components/signed-out',
                  label: '<ThunderIDSignedOut />',
                },
                {
                  type: 'doc',
                  id: 'sdks/nuxt/apis/components/loading',
                  label: '<ThunderIDLoading />',
                },
              ],
            },
            {
              type: 'category',
              label: 'Action Components',
              collapsed: true,
              items: [
                {
                  type: 'doc',
                  id: 'sdks/nuxt/apis/components/sign-in-button',
                  label: '<ThunderIDSignInButton />',
                },
                {
                  type: 'doc',
                  id: 'sdks/nuxt/apis/components/sign-out-button',
                  label: '<ThunderIDSignOutButton />',
                },
                {
                  type: 'doc',
                  id: 'sdks/nuxt/apis/components/sign-up-button',
                  label: '<ThunderIDSignUpButton />',
                },
              ],
            },
            {
              type: 'category',
              label: 'User Self-care Components',
              collapsed: true,
              items: [
                {
                  type: 'doc',
                  id: 'sdks/nuxt/apis/components/user',
                  label: '<ThunderIDUser />',
                },
                {
                  type: 'doc',
                  id: 'sdks/nuxt/apis/components/user-profile',
                  label: '<ThunderIDUserProfile />',
                },
                {
                  type: 'doc',
                  id: 'sdks/nuxt/apis/components/user-dropdown',
                  label: '<ThunderIDUserDropdown />',
                },
              ],
            },
          ],
        },
        {
          type: 'category',
          label: 'Middleware',
          collapsed: false,
          items: [
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/middleware/define-thunderid-middleware',
              label: 'defineThunderIDMiddleware()',
            },
          ],
        },
        {
          type: 'category',
          label: 'Server Utilities',
          collapsed: false,
          items: [
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/server/use-server-session',
              label: 'useServerSession()',
            },
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/server/require-server-session',
              label: 'requireServerSession()',
            },
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/server/get-valid-access-token',
              label: 'getValidAccessToken()',
            },
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/server/get-thunderid-context',
              label: 'getThunderIDContext()',
            },
          ],
        },
        {
          type: 'category',
          label: 'Utilities',
          collapsed: true,
          items: [
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/utils/create-route-matcher',
              label: 'createRouteMatcher()',
            },
          ],
        },
        {
          type: 'category',
          label: 'Errors',
          collapsed: true,
          items: [
            {
              type: 'doc',
              id: 'sdks/nuxt/apis/errors/thunderid-error',
              label: 'ThunderIDError',
            },
          ],
        },
      ],
    },
  ],
};

export default sidebar.nuxtSdkSidebar;
