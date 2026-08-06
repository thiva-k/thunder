// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {getDisplayNameForCode, toFlagEmoji} from '@thunderid/i18n';
import {Alert, Box, Button, CircularProgress, IconButton, PageTitle, Typography} from '@wso2/oxygen-ui';
import {ArrowLeft} from '@wso2/oxygen-ui-icons-react';
import {type JSX} from 'react';
import {useTranslation} from 'react-i18next';

/**
 * Props for the {@link TranslationEditorHeader} component.
 *
 * @public
 */
export interface TranslationEditorHeaderProps {
  /** The currently selected language code, or null if none. */
  selectedLanguage: string | null;
  /** Whether there are unsaved local changes. */
  hasDirtyChanges: boolean;
  /** Number of dirty (unsaved) keys. */
  dirtyCount: number;
  /** Whether a save or reset operation is in progress. */
  isSaving: boolean;
  /** Whether the selected language is the fallback language (disables Reset to Default). */
  isFallbackLanguage: boolean;
  /** Whether a namespace is selected (required to enable Reset to Default). */
  hasNamespace: boolean;
  /** Resolved error message from the last save or reset-to-default attempt, if any. */
  error?: string;
  /** Called when the user clicks the back button. */
  onBack: () => void;
  /** Called when the user clicks Discard Changes. */
  onDiscard: () => void;
  /** Called when the user clicks Reset to Default. */
  onResetToDefault: () => void;
  /** Called when the user clicks Save Changes. */
  onSave: () => void;
}

/**
 * Page title bar for the translations editor. Renders a back button, the
 * current language name with its flag, and the action buttons (Discard,
 * Reset to Default, Save).
 *
 * @param props - The component props
 *
 * @returns JSX element rendering the editor header
 *
 * @public
 */
export default function TranslationEditorHeader({
  selectedLanguage,
  hasDirtyChanges,
  dirtyCount,
  isSaving,
  isFallbackLanguage,
  hasNamespace,
  error = undefined,
  onBack,
  onDiscard,
  onResetToDefault,
  onSave,
}: TranslationEditorHeaderProps): JSX.Element {
  const {t} = useTranslation('translations');

  return (
    <PageTitle>
      <PageTitle.Header>
        <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
          <IconButton onClick={onBack}>
            <ArrowLeft size={16} />
          </IconButton>
          {selectedLanguage ? (
            <Box sx={{display: 'flex', alignItems: 'center', gap: 1}}>
              <Typography component="span" sx={{fontSize: 'inherit', userSelect: 'none'}}>
                {toFlagEmoji(selectedLanguage)}
              </Typography>
              {getDisplayNameForCode(selectedLanguage)}
            </Box>
          ) : (
            t('page.title')
          )}
        </Box>
      </PageTitle.Header>
      <PageTitle.Actions>
        <Box sx={{display: 'flex', flexDirection: 'column', gap: 1, alignItems: 'flex-end'}}>
          {error && (
            <Alert severity="error" sx={{py: 0}}>
              {error}
            </Alert>
          )}
          <Box sx={{display: 'flex', gap: 1, alignItems: 'center'}}>
            {hasDirtyChanges && (
              <Typography variant="caption" color="warning.main" sx={{fontWeight: 500}}>
                {t('editor.unsavedCount', {count: dirtyCount})}
              </Typography>
            )}
            <Button size="small" onClick={onDiscard} disabled={!hasDirtyChanges || isSaving}>
              {t('actions.discardChanges')}
            </Button>
            {!isFallbackLanguage && (
              <Button size="small" onClick={onResetToDefault} disabled={!hasNamespace || isSaving}>
                {t('actions.resetToDefault')}
              </Button>
            )}
            <Button
              size="small"
              variant="contained"
              onClick={onSave}
              disabled={!hasDirtyChanges || isSaving}
              startIcon={isSaving ? <CircularProgress size={14} color="inherit" /> : undefined}
            >
              {t('actions.saveChanges')}
            </Button>
          </Box>
        </Box>
      </PageTitle.Actions>
    </PageTitle>
  );
}
