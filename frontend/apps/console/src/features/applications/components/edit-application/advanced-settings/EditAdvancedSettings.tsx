// Copyright 2025-2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {
  Application,
  AttestationConfig,
  InboundAuthConfig,
  OAuth2Config,
  OAuth2Token,
} from '@thunderid/configure-applications';
import {Stack} from '@wso2/oxygen-ui';
import {useEffect, useState} from 'react';
import AttestationSection from './AttestationSection';
import AudienceSection from './AudienceSection';
import CertificateSection from './CertificateSection';
import IdentityAssertionsSection from './IdentityAssertionsSection';
import MetadataSection from './MetadataSection';
import OAuth2ConfigSection from './OAuth2ConfigSection';
import PasskeysSection from './PasskeysSection';
import type {ApplicationTemplate} from '../../../models/application-templates';

/**
 * Props for the {@link EditAdvancedSettings} component.
 */
interface EditAdvancedSettingsProps {
  /**
   * The application being edited
   */
  application: Application;
  /**
   * Partial application object containing edited fields
   */
  editedApp: Partial<Application>;
  /**
   * OAuth2 configuration for the application (optional)
   */
  oauth2Config?: OAuth2Config;
  /**
   * Template-driven field constraints for OAuth2 fields (optional)
   */
  oauth2Constraints?: NonNullable<ApplicationTemplate['fieldConstraints']>['oauth2'];
  /**
   * Callback function to handle field value changes
   * @param field - The application field being updated
   * @param value - The new value for the field
   */
  onFieldChange: (field: keyof Application, value: unknown) => void;
  /**
   * When set, restricts the offered grant types to the intersection of this list and
   * discovery's `grant_types_supported`. Omit to offer every discovery-advertised grant type.
   */
  allowedGrantTypes?: string[];
  /**
   * Whether the platform attestation section is shown. Driven by the template's `attestation`
   * capability, so it appears only for templates that support it (e.g. mobile).
   */
  showAttestation?: boolean;
  /**
   * Callback to report whether any child section (identity assertions / ID-JAG, or platform
   * attestation) currently has validation errors (feeds the page's Save bar).
   */
  onValidationChange?: (hasErrors: boolean) => void;
}

type OAuthCertificate = {type: string; value?: string} | null;

/**
 * Container component for advanced application settings.
 *
 * Displays sections for:
 * - OAuth2 configuration (grant types, response types, PKCE, public client)
 * - Certificate configuration (JWKS/JWKS URI)
 * - Application metadata (created/updated timestamps)
 *
 * @param props - Component props
 * @returns Advanced settings sections wrapped in a Stack
 */
export default function EditAdvancedSettings({
  application,
  editedApp,
  oauth2Config = undefined,
  oauth2Constraints = undefined,
  onFieldChange,
  allowedGrantTypes = undefined,
  showAttestation = false,
  onValidationChange = undefined,
}: EditAdvancedSettingsProps) {
  // Identity assertions and attestation validate independently; each is tracked separately so one
  // resolving doesn't clobber the other's still-invalid state when both report to the single
  // upward onValidationChange prop.
  // Identity assertions, attestation, and passkeys validate independently; each is tracked
  // separately so one resolving doesn't clobber the other's still-invalid state when both report
  // to the single upward onValidationChange prop.
  const [identityAssertionsInvalid, setIdentityAssertionsInvalid] = useState(false);
  const [attestationInvalid, setAttestationInvalid] = useState(false);
  const [passkeysInvalid, setPasskeysInvalid] = useState(false);

  useEffect(() => {
    onValidationChange?.(identityAssertionsInvalid || attestationInvalid || passkeysInvalid);
  }, [identityAssertionsInvalid, attestationInvalid, passkeysInvalid, onValidationChange]);

  const handleOAuth2ConfigChange = (updates: Partial<OAuth2Config>) => {
    const currentInboundAuth: InboundAuthConfig[] = editedApp.inboundAuthConfig ?? application.inboundAuthConfig ?? [];
    const updatedInboundAuth = currentInboundAuth.map((auth) =>
      auth.type === 'oauth2' ? {...auth, config: {...auth.config, ...updates}} : auth,
    );
    onFieldChange('inboundAuthConfig', updatedInboundAuth);
  };

  const handleCertificateChange = (cert: OAuthCertificate) => {
    handleOAuth2ConfigChange({certificate: cert});
  };

  // Attestation is a client-level (protocol-agnostic) setting, so it is stored at the top level of
  // the application rather than nested under the OAuth2 config. This lets any application type —
  // including embedded apps with no OAuth2 config — enable it.
  const handleAttestationChange = (attestation: AttestationConfig | null) => {
    onFieldChange('attestation', attestation);
  };

  // Prefer the edited value whenever it has been set — including an explicit null, which represents
  // the user clearing attestation. Only fall back to the stored value when the field is untouched.
  const currentAttestation = 'attestation' in editedApp ? editedApp.attestation : application.attestation;

  const handlePasskeysChange = (origins: string[]): void => {
    onFieldChange('passkeyAllowedOrigins', origins);
  };

  const currentPasskeyOrigins = editedApp.passkeyAllowedOrigins ?? application.passkeyAllowedOrigins ?? [];

  // Encrypted ID token / UserInfo formats are encrypted to the client certificate, so removing the
  // certificate while one is selected would produce an invalid config. Used to block that removal.
  const idTokenResponseType = oauth2Config?.token?.idToken?.responseType;
  const userInfoResponseType = oauth2Config?.userInfo?.responseType;
  const encryptionDependsOnCert =
    idTokenResponseType === 'JWE' ||
    idTokenResponseType === 'NESTED_JWT' ||
    userInfoResponseType === 'JWE' ||
    userInfoResponseType === 'NESTED_JWT';

  const handleTokenConfigChange = (tokenUpdates: Partial<OAuth2Token>, oauth2Updates: Partial<OAuth2Config> = {}) => {
    const currentInboundAuth: InboundAuthConfig[] = editedApp.inboundAuthConfig ?? application.inboundAuthConfig ?? [];
    const updatedInboundAuth = currentInboundAuth.map((auth) =>
      auth.type === 'oauth2'
        ? {
            ...auth,
            config: {
              ...auth.config,
              ...oauth2Updates,
              token: {...auth.config?.token, ...tokenUpdates},
            },
          }
        : auth,
    );
    onFieldChange('inboundAuthConfig', updatedInboundAuth);
  };

  const handleDefaultAudienceChange = (audience: string) => {
    handleTokenConfigChange({
      accessToken: {...oauth2Config?.token?.accessToken, defaultAudience: audience},
    });
  };

  return (
    <Stack spacing={3}>
      <OAuth2ConfigSection
        oauth2Config={oauth2Config}
        oauth2Constraints={oauth2Constraints}
        onOAuth2ConfigChange={handleOAuth2ConfigChange}
        disabled={application.isReadOnly}
        allowedGrantTypes={allowedGrantTypes}
      />
      {oauth2Config && (
        <IdentityAssertionsSection
          oauth2Config={oauth2Config}
          onTokenConfigChange={handleTokenConfigChange}
          disabled={application.isReadOnly}
          onValidationChange={setIdentityAssertionsInvalid}
        />
      )}
      {oauth2Config && (
        <AudienceSection
          audience={oauth2Config.token?.accessToken?.defaultAudience ?? ''}
          onAudienceChange={handleDefaultAudienceChange}
          disabled={application.isReadOnly}
        />
      )}
      <CertificateSection
        certificate={oauth2Config?.certificate}
        onCertificateChange={handleCertificateChange}
        required={oauth2Config?.tokenEndpointAuthMethod === 'private_key_jwt'}
        encryptionDependsOnCert={encryptionDependsOnCert}
        disabled={application.isReadOnly}
      />
      {showAttestation && (
        <AttestationSection
          attestation={currentAttestation}
          onAttestationChange={handleAttestationChange}
          disabled={application.isReadOnly}
          onValidationChange={setAttestationInvalid}
        />
      )}
      <PasskeysSection
        allowedOrigins={currentPasskeyOrigins}
        onPasskeysChange={application.isReadOnly ? undefined : handlePasskeysChange}
        disabled={application.isReadOnly}
        onValidationChange={setPasskeysInvalid}
      />
      <MetadataSection application={application} />
    </Stack>
  );
}
