/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import {
  Box,
  Typography,
  Stack,
  TextField,
  FormControl,
  FormLabel,
  Button,
  Chip,
  Alert,
  useTheme,
} from '@wso2/oxygen-ui';
import {Lightbulb, Plus, X} from '@wso2/oxygen-ui-icons-react';
import type {ChangeEvent, JSX} from 'react';
import {useState, useEffect} from 'react';
import {useTranslation} from 'react-i18next';

/**
 * Props for the ConfigureRedirectURIs component.
 */
export interface ConfigureRedirectURIsProps {
  /**
   * Array of redirect URIs
   */
  redirectURIs: string[];

  /**
   * Callback function when redirect URIs change
   */
  onRedirectURIsChange: (uris: string[]) => void;

  /**
   * Callback function to broadcast whether this step is ready to proceed
   */
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Validate if a string is a valid URL
 */
const isValidURL = (url: string): boolean => {
  try {
    const urlObj = new URL(url);
    return urlObj.protocol === 'http:' || urlObj.protocol === 'https:';
  } catch {
    return false;
  }
};

export default function ConfigureRedirectURIs({
  redirectURIs,
  onRedirectURIsChange,
  onReadyChange = undefined,
}: ConfigureRedirectURIsProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const [currentURI, setCurrentURI] = useState('');
  const [error, setError] = useState<string | null>(null);

  // Broadcast readiness whenever redirectURIs change
  useEffect(() => {
    // Check that we have at least one non-empty URI
    const isReady = redirectURIs.length > 0 && redirectURIs.every((uri) => uri.trim().length > 0);
    if (onReadyChange) {
      onReadyChange(isReady);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [redirectURIs]);

  const handleAddURI = (): void => {
    const trimmedURI = currentURI.trim();

    if (!trimmedURI) {
      setError(t('applications:onboarding.configure.redirectURIs.errors.empty'));
      return;
    }

    if (!isValidURL(trimmedURI)) {
      setError(t('applications:onboarding.configure.redirectURIs.errors.invalid'));
      return;
    }

    if (redirectURIs.includes(trimmedURI)) {
      setError(t('applications:onboarding.configure.redirectURIs.errors.duplicate'));
      return;
    }

    setError(null);
    onRedirectURIsChange([...redirectURIs, trimmedURI]);
    setCurrentURI('');
  };

  const handleRemoveURI = (uriToRemove: string): void => {
    onRedirectURIsChange(redirectURIs.filter((uri) => uri !== uriToRemove));
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>): void => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAddURI();
    }
  };

  const handleURIChange = (e: ChangeEvent<HTMLInputElement>): void => {
    setCurrentURI(e.target.value);
    if (error) {
      setError(null);
    }
  };

  return (
    <Stack direction="column" spacing={4}>
      <Stack direction="column" spacing={1}>
        <Typography variant="h4" component="h1" gutterBottom>
          {t('applications:onboarding.configure.redirectURIs.title')}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {t('applications:onboarding.configure.redirectURIs.subtitle')}
        </Typography>
      </Stack>

      <FormControl fullWidth>
        <FormLabel htmlFor="redirect-uri-input">{t('applications:onboarding.configure.redirectURIs.fieldLabel')}</FormLabel>
        <Stack direction="row" spacing={2}>
          <TextField
            fullWidth
            id="redirect-uri-input"
            value={currentURI}
            onChange={handleURIChange}
            onKeyPress={handleKeyPress}
            placeholder={t('applications:onboarding.configure.redirectURIs.placeholder')}
            error={!!error}
            helperText={error}
          />
          <Button
            variant="outlined"
            onClick={handleAddURI}
            startIcon={<Plus size={20} />}
            sx={{minWidth: 120}}
          >
            {t('applications:onboarding.configure.redirectURIs.addButton')}
          </Button>
        </Stack>
      </FormControl>

      {/* Display added URIs */}
      {redirectURIs.length > 0 && (
        <Box>
          <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
            {t('applications:onboarding.configure.redirectURIs.addedLabel')}
          </Typography>
          <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
            {redirectURIs.map(
              (uri: string): JSX.Element => (
                <Chip
                  key={uri}
                  label={uri}
                  onDelete={(): void => handleRemoveURI(uri)}
                  deleteIcon={<X size={16} />}
                  variant="outlined"
                  sx={{
                    '& .MuiChip-deleteIcon': {
                      color: 'text.secondary',
                      '&:hover': {
                        color: 'error.main',
                      },
                    },
                  }}
                />
              ),
            )}
          </Box>
        </Box>
      )}

      {/* Validation warning if no URIs added */}
      {redirectURIs.length === 0 && (
        <Alert severity="warning">
          {t('applications:onboarding.configure.redirectURIs.warning')}
        </Alert>
      )}

      <Stack direction="row" alignItems="center" spacing={1}>
        <Lightbulb size={20} color={theme?.vars?.palette.warning.main} />
        <Typography variant="body2" color="text.secondary">
          {t('applications:onboarding.configure.redirectURIs.hint')}
        </Typography>
      </Stack>
    </Stack>
  );
}

