// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Stack} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import TechnologyGuide from './TechnologyGuide';
import type {IntegrationGuides} from '../../../models/application-templates';

/**
 * Props for the {@link IntegrationGuide} component.
 *
 * @public
 */
export interface IntegrationGuideProps {
  /**
   * The client ID (if OAuth was configured)
   */
  clientId?: string;
  /**
   * The ID of the created application
   */
  applicationId?: string | null;
  /**
   * Integration guides configuration (optional - if not provided, won't show guides)
   */
  integrationGuides?: IntegrationGuides | null;
  /**
   * The template ID used to create the application (e.g., 'react', 'react-embedded')
   */
  templateId?: string | null;
}

/**
 * React component that displays integration guides and setup instructions
 * for newly created applications.
 *
 * This component provides:
 * 1. Technology-specific integration guides with code snippets
 * 2. OAuth2 credentials (Client ID and Secret) when applicable
 * 3. Step-by-step instructions for integrating with various frameworks
 *
 * The component handles different scenarios:
 * - Applications with integration guides (shows TechnologyGuide)
 * - Applications with OAuth configuration (displays credentials)
 * - Public vs confidential client configurations
 *
 * @param props - The component props
 * @param props.appName - Name of the application
 * @param props.appLogo - URL of the application logo
 * @param props.selectedColor - Brand color for visual elements
 * @param props.clientId - OAuth2 client ID (if applicable)
 * @param props.clientSecret - OAuth2 client secret (if applicable)
 * @param props.hasOAuthConfig - Whether OAuth was configured
 * @param props.applicationId - ID of the application
 *
 * @returns JSX element displaying the integration guide
 *
 * @example
 * ```tsx
 * import IntegrationGuide from './IntegrationGuide';
 *
 * function ApplicationOverview() {
 *   return (
 *     <IntegrationGuide
 *       appName="My Application"
 *       appLogo="https://example.com/logo.png"
 *       selectedColor="#FF5733"
 *       clientId="abc123"
 *       clientSecret="secret456"
 *       hasOAuthConfig={true}
 *       applicationId="app-uuid"
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function IntegrationGuide({
  clientId = '',
  applicationId = null,
  integrationGuides = null,
  templateId = null,
}: IntegrationGuideProps): JSX.Element {
  return (
    <Stack direction="column" spacing={4} sx={{width: '100%', alignItems: 'center'}}>
      {integrationGuides && (
        <TechnologyGuide
          guides={integrationGuides}
          templateId={templateId}
          clientId={clientId}
          applicationId={applicationId ?? undefined}
        />
      )}
    </Stack>
  );
}
