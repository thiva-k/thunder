// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Box, Card, CardActionArea, CircularProgress, LinearProgress, Stack, Typography} from '@wso2/oxygen-ui';
import {type JSX} from 'react';
import {useTranslation} from 'react-i18next';

/**
 * Props for the {@link InitializeLanguage} component.
 *
 * @public
 */
export interface InitializeLanguageProps {
  /** Whether the user has chosen to pre-populate keys from English (en-US). */
  populateFromEnglish: boolean;
  /** Callback invoked when the user changes the initialization strategy. */
  onPopulateChange: (value: boolean) => void;
  /** Whether language creation is currently in progress. */
  isCreating: boolean;
  /** Creation progress percentage (0–100), displayed while creation is in progress. */
  progress: number;
}

/**
 * Step component for the language creation wizard that lets the user choose how
 * to initialize the new language's translation keys.
 *
 * Presents two card options — copying from English (en-US) or starting with
 * empty values — and shows a progress bar while keys are being written to the
 * server.
 *
 * @param props - The component props
 * @param props.populateFromEnglish - Whether the user has chosen to copy from English
 * @param props.onPopulateChange - Callback invoked when the initialization strategy changes
 * @param props.isCreating - Whether language creation is currently in progress
 * @param props.progress - Creation progress percentage (0–100)
 *
 * @returns JSX element rendering the initialization strategy selector
 *
 * @example
 * ```tsx
 * import InitializeLanguage from './InitializeLanguage';
 *
 * function Wizard() {
 *   const [populate, setPopulate] = useState(true);
 *   return (
 *     <InitializeLanguage
 *       populateFromEnglish={populate}
 *       onPopulateChange={setPopulate}
 *       isCreating={false}
 *       progress={0}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function InitializeLanguage({
  populateFromEnglish,
  onPopulateChange,
  isCreating,
  progress,
}: InitializeLanguageProps): JSX.Element {
  const {t} = useTranslation('translations');

  return (
    <Stack spacing={4}>
      <Box>
        <Typography variant="h1" gutterBottom>
          {t('language.create.initialize.title')}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {t('language.create.initialize.subtitle')}
        </Typography>
      </Box>

      <Stack spacing={2}>
        <Card
          variant="outlined"
          onClick={() => !isCreating && onPopulateChange(true)}
          sx={{
            borderColor: populateFromEnglish ? 'primary.main' : 'divider',
            borderWidth: populateFromEnglish ? 2 : 1,
            cursor: isCreating ? 'default' : 'pointer',
            opacity: isCreating ? 0.6 : 1,
          }}
        >
          <CardActionArea disabled={isCreating} sx={{p: 2.5}}>
            <Stack spacing={0.5}>
              <Typography variant="body1" fontWeight={600}>
                {t('language.create.initialize.copyFromEnglish.label')}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {t('language.create.initialize.copyFromEnglish.description')}
              </Typography>
            </Stack>
          </CardActionArea>
        </Card>

        <Card
          variant="outlined"
          onClick={() => !isCreating && onPopulateChange(false)}
          sx={{
            borderColor: !populateFromEnglish ? 'primary.main' : 'divider',
            borderWidth: !populateFromEnglish ? 2 : 1,
            cursor: isCreating ? 'default' : 'pointer',
            opacity: isCreating ? 0.6 : 1,
          }}
        >
          <CardActionArea disabled={isCreating} sx={{p: 2.5}}>
            <Stack spacing={0.5}>
              <Typography variant="body1" fontWeight={600}>
                {t('language.create.initialize.startEmpty.label')}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {t('language.create.initialize.startEmpty.description')}
              </Typography>
            </Stack>
          </CardActionArea>
        </Card>
      </Stack>

      {isCreating && (
        <Box sx={{display: 'flex', flexDirection: 'column', gap: 1}}>
          <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
            <CircularProgress size={16} />
            <Typography variant="body2" color="text.secondary">
              {t('language.add.adding')} ({progress}%)
            </Typography>
          </Box>
          <LinearProgress variant="determinate" value={progress} />
        </Box>
      )}
    </Stack>
  );
}
