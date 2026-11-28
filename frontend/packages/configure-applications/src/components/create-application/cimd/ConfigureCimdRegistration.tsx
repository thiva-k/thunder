// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {
  Box,
  Card,
  CardActionArea,
  CardContent,
  FormControlLabel,
  Radio,
  RadioGroup,
  Stack,
  Typography,
  useTheme,
} from '@wso2/oxygen-ui';
import {FileBadge, Globe, Lightbulb} from '@wso2/oxygen-ui-icons-react';
import type {ChangeEvent, JSX, ReactNode} from 'react';
import {useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import ConfigureMetadataDocument from './ConfigureMetadataDocument';
import {CimdRegistrationModes} from '../../../models/cimd';
import type {CimdPreview, CimdRegistrationMode} from '../../../models/cimd';
import {validateClientIdDomains} from '../../../utils/clientIdDomains';
import ClientIdDomainsField from '../../cimd/ClientIdDomainsField';

/**
 * Props for the {@link ConfigureCimdRegistration} component.
 *
 * @public
 */
export interface ConfigureCimdRegistrationProps {
  /** Which clients the application admits. */
  mode: CimdRegistrationMode;
  /** Invoked when the administrator picks the other option. */
  onModeChange: (mode: CimdRegistrationMode) => void;
  /** The approved document of a specific client, or null when none has been fetched for the current URL. */
  preview: CimdPreview | null;
  /** Invoked with the validated document of a specific client, or null when the URL changes. */
  onPreviewChange: (preview: CimdPreview | null) => void;
  /** The trusted domains entered so far. */
  domains: string[];
  /** Invoked with the trusted domains whenever they change. */
  onDomainsChange: (domains: string[]) => void;
  /** Invoked whenever the step's readiness changes. */
  onReadyChange?: (isReady: boolean) => void;
  /** Known clients offered as quick picks for a specific client, from the template's `metadataDocumentClients`. */
  knownClients?: {name: string; clientId: string}[];
}

/**
 * The CIMD template's first step. The administrator chooses between registering one specific client from its
 * metadata document, reviewed before it is saved, and letting any client whose client ID URL is on a trusted
 * domain sign in, with its document read when it signs in.
 *
 * @public
 */
export default function ConfigureCimdRegistration({
  mode,
  onModeChange,
  preview,
  onPreviewChange,
  domains,
  onDomainsChange,
  onReadyChange = undefined,
  knownClients = [],
}: ConfigureCimdRegistrationProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const domainsError = validateClientIdDomains(domains);

  useEffect((): void => {
    onReadyChange?.(mode === CimdRegistrationModes.TRUSTED_DOMAINS ? domainsError === undefined : preview !== null);
  }, [mode, preview, domainsError, onReadyChange]);

  const title = t('applications:cimd.registration.title', 'Which clients should sign in?');
  const options: {value: CimdRegistrationMode; icon: ReactNode; title: string; description: string}[] = [
    {
      value: CimdRegistrationModes.SPECIFIC_CLIENT,
      icon: <FileBadge size={20} />,
      title: t('applications:cimd.registration.specificClient.title', 'A specific client'),
      description: t(
        'applications:cimd.registration.specificClient.description',
        'Register one client from its metadata document URL. You review what the document declares before the application is created.',
      ),
    },
    {
      value: CimdRegistrationModes.TRUSTED_DOMAINS,
      icon: <Globe size={20} />,
      title: t('applications:cimd.registration.trustedDomains.title', 'Any client from trusted domains'),
      description: t(
        'applications:cimd.registration.trustedDomains.description',
        "Let every client whose client ID URL is on a domain you trust sign in, without registering each one. A client's document is read when it signs in.",
      ),
    },
  ];

  return (
    <Stack direction="column" spacing={3} data-testid="application-configure-cimd-registration">
      <Stack direction="column" spacing={0.5}>
        <Typography variant="h1" gutterBottom>
          {title}
        </Typography>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
          <Typography variant="body2" color="text.secondary">
            {t(
              'applications:cimd.registration.subtitle',
              'Clients that publish a Client ID Metadata Document identify themselves by its URL, so their sign-in details come from their vendor.',
            )}
          </Typography>
        </Stack>
      </Stack>

      <RadioGroup
        aria-label={title}
        value={mode}
        onChange={(event: ChangeEvent<HTMLInputElement>) => onModeChange(event.target.value as CimdRegistrationMode)}
      >
        <Stack direction={{xs: 'column', md: 'row'}} spacing={2}>
          {options.map((option) => {
            const isSelected = mode === option.value;
            return (
              <Card
                key={option.value}
                variant="outlined"
                sx={{flex: 1}}
                onClick={() => onModeChange(option.value)}
                data-testid={`cimd-registration-${option.value}`}
              >
                <CardActionArea
                  sx={{
                    height: '100%',
                    border: 1,
                    borderColor: isSelected ? 'primary.main' : 'divider',
                    '&:hover': {
                      borderColor: 'primary.main',
                      bgcolor: isSelected ? 'action.selected' : 'action.hover',
                    },
                  }}
                >
                  <CardContent>
                    <Stack direction="row" spacing={2} alignItems="flex-start">
                      <FormControlLabel
                        value={option.value}
                        control={<Radio inputProps={{'aria-label': option.title}} />}
                        label=""
                        sx={{m: 0}}
                        onClick={(e) => e.stopPropagation()}
                      />
                      <Box sx={{flex: 1}}>
                        <Stack direction="row" spacing={1} alignItems="center" sx={{mb: 1}}>
                          {option.icon}
                          <Typography variant="h6">{option.title}</Typography>
                        </Stack>
                        <Typography variant="body2" color="text.secondary">
                          {option.description}
                        </Typography>
                      </Box>
                    </Stack>
                  </CardContent>
                </CardActionArea>
              </Card>
            );
          })}
        </Stack>
      </RadioGroup>

      {mode === CimdRegistrationModes.TRUSTED_DOMAINS ? (
        <Stack spacing={1.5}>
          <ClientIdDomainsField
            value={domains}
            onChange={onDomainsChange}
            // An empty list is only incomplete, so its message waits until a domain has been entered.
            error={domains.length > 0 && domainsError ? t(domainsError.key, domainsError.defaultValue) : undefined}
          />
          <Typography variant="body2" color="text.secondary">
            {t(
              'applications:cimd.registration.trustedDomains.warning',
              'Anyone who can publish a document on these domains can sign in through this application. Add only domains whose publishers you trust, and prefer an exact domain to a wildcard.',
            )}
          </Typography>
        </Stack>
      ) : (
        <ConfigureMetadataDocument
          embedded
          knownClients={knownClients}
          preview={preview}
          onPreviewChange={onPreviewChange}
        />
      )}
    </Stack>
  );
}
