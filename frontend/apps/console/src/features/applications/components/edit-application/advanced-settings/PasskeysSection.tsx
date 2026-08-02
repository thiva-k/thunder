// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {SettingsCard} from '@thunderid/components';
import {Box, Button, FormControl, IconButton, Stack, TextField, Tooltip} from '@wso2/oxygen-ui';
import {Plus, Trash} from '@wso2/oxygen-ui-icons-react';
import {useEffect, useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';

interface PasskeysSectionProps {
  /**
   * The current list of allowed origins for passkey operations.
   */
  allowedOrigins?: string[];
  /**
   * Callback invoked when the list changes. Omit to render read-only.
   */
  onPasskeysChange?: (origins: string[]) => void;
  /**
   * Whether inputs should be disabled (e.g. read-only resource).
   */
  disabled?: boolean;
  /**
   * Called whenever the section's validation state changes. `true` means at least one origin is
   * empty or not a valid URL, so the parent can block Save.
   */
  onValidationChange?: (hasErrors: boolean) => void;
}

const isValidURL = (value: string): boolean => {
  try {
    return Boolean(new URL(value));
  } catch {
    return false;
  }
};

export default function PasskeysSection({
  allowedOrigins = undefined,
  onPasskeysChange = undefined,
  disabled = false,
  onValidationChange = undefined,
}: PasskeysSectionProps): JSX.Element | null {
  const {t} = useTranslation();
  const [errors, setErrors] = useState<Record<number, string>>({});

  const origins = allowedOrigins ?? [];
  const isEditable = Boolean(onPasskeysChange) && !disabled;

  const hasInvalidOrigins = origins.length > 0 && origins.some((o) => !o.trim() || !isValidURL(o));

  useEffect(() => {
    onValidationChange?.(hasInvalidOrigins);
  }, [hasInvalidOrigins, onValidationChange]);

  const commit = (next: string[]): void => {
    if (!isEditable) return;
    onPasskeysChange?.(next);
  };

  const handleChange = (index: number, value: string): void => {
    const next = [...origins];
    next[index] = value;
    if (value.trim()) {
      setErrors((prev) => {
        const copy = {...prev};
        delete copy[index];
        return copy;
      });
    }
    commit(next);
  };

  const handleBlur = (index: number): void => {
    const value = origins[index] ?? '';
    if (!value.trim()) {
      setErrors((prev) => ({
        ...prev,
        [index]: t('applications:edit.advanced.passkeys.allowedOrigins.error.empty', 'Origin cannot be empty'),
      }));
      return;
    }
    if (!isValidURL(value)) {
      setErrors((prev) => ({
        ...prev,
        [index]: t('applications:edit.advanced.passkeys.allowedOrigins.error.invalid', 'Enter a valid URL'),
      }));
    }
  };

  const handleAdd = (): void => {
    commit([...origins, '']);
  };

  const handleRemove = (index: number): void => {
    const next = origins.filter((_, i) => i !== index);
    setErrors((prev) => {
      const copy: Record<number, string> = {};
      Object.entries(prev).forEach(([key, value]) => {
        const i = parseInt(key, 10);
        if (i < index) copy[i] = value;
        else if (i > index) copy[i - 1] = value;
      });
      return copy;
    });
    commit(next);
  };

  return (
    <SettingsCard
      title={t('applications:edit.advanced.labels.passkeys', 'Passkey Allowed Origins')}
      description={t(
        'applications:edit.advanced.passkeys.intro',
        'Allowed origins for passkey operations initiated through this application.',
      )}
    >
      <FormControl fullWidth>
        <Stack spacing={2}>
          {origins.map((origin, index) => (
            // eslint-disable-next-line react/no-array-index-key
            <Stack key={index} direction="row" spacing={1} alignItems="flex-start">
              <FormControl fullWidth sx={{flex: 1}}>
                <TextField
                  fullWidth
                  id={`app-passkey-origin-${index}`}
                  value={origin}
                  onChange={(e) => handleChange(index, e.target.value)}
                  onBlur={() => handleBlur(index)}
                  error={!!errors[index]}
                  helperText={errors[index]}
                  placeholder={t(
                    'applications:edit.advanced.passkeys.allowedOrigins.placeholder',
                    'https://app.example.com',
                  )}
                  disabled={!isEditable}
                />
              </FormControl>
              {isEditable && (
                <Tooltip title={t('common:actions.delete', 'Delete')}>
                  <IconButton onClick={() => handleRemove(index)} color="error" sx={{mt: 1}}>
                    <Trash size={20} />
                  </IconButton>
                </Tooltip>
              )}
            </Stack>
          ))}
          {isEditable && (
            <Box>
              <Button variant="outlined" startIcon={<Plus />} onClick={handleAdd} size="small">
                {t('applications:edit.advanced.passkeys.allowedOrigins.addOrigin', 'Add Origin')}
              </Button>
            </Box>
          )}
        </Stack>
      </FormControl>
    </SettingsCard>
  );
}
