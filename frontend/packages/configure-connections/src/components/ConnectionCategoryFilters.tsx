// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Chip, Stack} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {AVAILABLE_CONNECTION_CATEGORIES} from '../config/connectionVendorMeta';
import type {ConnectionCategory} from '../models/connection';

export type CategoryFilterValue = ConnectionCategory | 'all';

interface ConnectionCategoryFiltersProps {
  selected: CategoryFilterValue;
  onSelect: (value: CategoryFilterValue) => void;
}

export default function ConnectionCategoryFilters({selected, onSelect}: ConnectionCategoryFiltersProps): JSX.Element {
  const {t} = useTranslation('connections');

  const values: CategoryFilterValue[] = ['all', ...AVAILABLE_CONNECTION_CATEGORIES];

  return (
    <Stack
      direction="row"
      spacing={1}
      alignItems="center"
      flexWrap="wrap"
      useFlexGap
      data-testid="connection-category-filters"
    >
      {values.map((value) => (
        <Chip
          key={value}
          label={t(`categories.${value}`)}
          color={selected === value ? 'primary' : 'default'}
          variant={selected === value ? 'filled' : 'outlined'}
          onClick={() => onSelect(value)}
          sx={{borderRadius: '20px', cursor: 'pointer'}}
        />
      ))}
    </Stack>
  );
}
