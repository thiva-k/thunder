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

import {useMemo} from 'react';
import {useTranslation} from 'react-i18next';
import {isI18nPattern, resolveI18nValue} from '@/features/flows/utils/i18nPatternUtils';

/**
 * Hook to resolve i18n pattern values.
 * If the value matches the i18n pattern {{t(key)}}, it returns the translated value.
 * Otherwise, it returns the original value.
 *
 * @param value - The value to resolve.
 * @param stripHtml - Whether to strip HTML tags before checking the pattern. Default is false.
 * @returns The resolved value.
 */
function useResolveI18n(value: string | undefined, stripHtml = false): string {
  const {t} = useTranslation();

  return useMemo(() => {
    if (!value) return '';

    if (isI18nPattern(value, stripHtml)) {
      return resolveI18nValue(value, t, stripHtml);
    }

    return value;
  }, [value, stripHtml, t]);
}

export default useResolveI18n;
