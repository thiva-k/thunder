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

import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';
import webpackPlugin from './plugins/webpackPlugin';
import thunderConfig from './thunder.config';
import fs from 'fs';

const resourcesHTML = fs.readFileSync('./src/snippets/resources.html', 'utf-8');

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: thunderConfig.project.name,
  tagline: thunderConfig.project.description,
  favicon: 'assets/images/favicon.ico',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  // Set the production url of your site here
  url: thunderConfig.project.documentation.deployment.production.url,
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: thunderConfig.project.source.github.owner.name, // Usually your GitHub org/user name.
  projectName: thunderConfig.project.source.github.name, // Usually your repo name.

  onBrokenLinks: 'throw',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  plugins: [
    webpackPlugin,
    [
      'docusaurus-plugin-openapi-docs',
      {
        id: 'openapi-api',
        docsPluginId: 'classic',
        config: {
          applications: {
            specPath: '../api/application.yaml',
            outputDir: 'docs/apis/application',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          authentication: {
            specPath: '../api/authentication.yaml',
            outputDir: 'docs/apis/authentication',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          branding: {
            specPath: '../api/branding.yaml',
            outputDir: 'docs/apis/branding',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          flowExecution: {
            specPath: '../api/flow-execution.yaml',
            outputDir: 'docs/apis/flow-execution',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          flowManagement: {
            specPath: '../api/flow-management.yaml',
            outputDir: 'docs/apis/flow-management',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          group: {
            specPath: '../api/group.yaml',
            outputDir: 'docs/apis/group',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          healthCheck: {
            specPath: '../api/healthcheck.yaml',
            outputDir: 'docs/apis/health-check',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          i18n: {
            specPath: '../api/i18n.yaml',
            outputDir: 'docs/apis/i18n',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          idp: {
            specPath: '../api/idp.yaml',
            outputDir: 'docs/apis/idp',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          notificationSender: {
            specPath: '../api/notification-sender.yaml',
            outputDir: 'docs/apis/notification-sender',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          ou: {
            specPath: '../api/ou.yaml',
            outputDir: 'docs/apis/ou',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          registration: {
            specPath: '../api/registration.yaml',
            outputDir: 'docs/apis/registration',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          resource: {
            specPath: '../api/resource.yaml',
            outputDir: 'docs/apis/resource',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          role: {
            specPath: '../api/role.yaml',
            outputDir: 'docs/apis/role',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
          user: {
            specPath: '../api/user.yaml',
            outputDir: 'docs/apis/user',
            sidebarOptions: {
              groupPathsBy: 'tag',
            },
          },
        },
      },
    ],
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          // Derived from docusaurus-theme-openapi
          docItemComponent: '@theme/ApiItem',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl: 'https://github.com/facebook/docusaurus/tree/main/packages/create-docusaurus/templates/shared/',
        },
        blog: {
          showReadingTime: true,
          feedOptions: {
            type: ['rss', 'atom'],
            xslt: true,
          },
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl: 'https://github.com/facebook/docusaurus/tree/main/packages/create-docusaurus/templates/shared/',
          // Useful options to enforce blogging best practices
          onInlineTags: 'warn',
          onInlineAuthors: 'warn',
          onUntruncatedBlogPosts: 'warn',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themes: ['docusaurus-theme-openapi-docs'],

  themeConfig: {
    image: 'assets/images/thunder-social-card.png',
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: '',
      logo: {
        href: '/',
        src: '/assets/images/logo.svg',
        srcDark: '/assets/images/logo-inverted.svg',
        alt: `${thunderConfig.project.name} Logo`,
        height: '40px',
        width: '101px',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
          className: 'navbar__link--docs',
        },
        {
          type: 'docSidebar',
          sidebarId: 'apisSidebar',
          position: 'left',
          label: 'APIs',
        },
        {
          to: '/sdks',
          position: 'left',
          label: 'SDKs',
        },
        {
          label: 'Resources',
          type: 'dropdown',
          className: 'navbar__link--dropdown',
          items: [
            {
              type: 'html',
              value: resourcesHTML
                .replace('{{ISSUES_URL}}', thunderConfig.project.source.github.issuesUrl)
                .replace('{{DISCUSSIONS_URL}}', thunderConfig.project.source.github.discussionsUrl)
                .replace('{{RELEASES_URL}}', thunderConfig.project.source.github.releasesUrl),
              className: 'navbar__link--dropdown-item',
            },
          ],
        },
        {
          type: 'docSidebar',
          sidebarId: 'communitySidebar',
          position: 'left',
          label: 'Community',
        },
        {
          href: `https://github.com/${thunderConfig.project.source.github.fullName}`,
          position: 'right',
          className: 'navbar__github--link',
          'aria-label': 'GitHub repository',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [],
      copyright: `Copyright © ${new Date().getFullYear()} ${thunderConfig.project.name}.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
