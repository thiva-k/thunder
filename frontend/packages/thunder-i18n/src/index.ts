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

/**
 * Thunder i18n - Internationalization package for Thunder applications
 *
 * This package provides a complete i18n solution using react-i18next with:
 * - Tree-shakable translation files
 * - TypeScript support with type-safe translation keys
 * - Language detection and persistence
 * - React hooks for easy integration
 * - Formatting utilities for dates, numbers, and currency
 *
 * @example Basic usage in an app
 * ```tsx
 * import { initI18n } from '@thunder/i18n';
 * import en from '@thunder/i18n/locales/en';
 * import si from '@thunder/i18n/locales/si';
 *
 * // Initialize i18n
 * await initI18n({
 *   translations: { en, si },
 *   options: {
 *     defaultLanguage: 'en',
 *     debug: process.env.NODE_ENV === 'development'
 *   }
 * });
 * ```
 *
 * @example Using in components
 * ```tsx
 * import { useTranslation } from 'react-i18next';
 * import { useLanguage } from '@thunder/i18n';
 *
 * function MyComponent() {
 *   const { t } = useTranslation();
 *   const { currentLanguage, setLanguage } = useLanguage();
 *
 *   return (
 *     <div>
 *       <h1>{t('common.navigation.home')}</h1>
 *       <button onClick={() => setLanguage('si')}>සිංහල</button>
 *     </div>
 *   );
 * }
 * ```
 *
 * @packageDocumentation
 */

// Core initialization and utilities
export {
  initI18n,
  changeLanguage,
  getCurrentLanguage,
  getAvailableLanguages,
  detectLanguage,
  saveLanguage,
  formatDate,
  formatNumber,
  formatCurrency,
} from './utils/i18n';

// React hooks
export {useLanguage} from './hooks';

// TypeScript types
export type {
  SupportedLanguage,
  TranslationResources,
  TranslationKey,
  TranslationNamespace,
  I18nOptions,
  LanguageConfig,
  Translations,
} from './types';

export {LANGUAGE_CONFIGS, isSupportedLanguage} from './types';

// Re-export commonly used react-i18next hooks and components
export {useTranslation, Trans, Translation} from 'react-i18next';
export type {TFunction} from 'i18next';
