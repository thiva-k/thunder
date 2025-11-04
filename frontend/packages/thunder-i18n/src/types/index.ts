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
 * TypeScript type definitions for Thunder i18n
 */

import type {en} from '../locales/en';
import type {si} from '../locales/si';

/**
 * Translation resources type derived from English translations
 */
export type TranslationResources = typeof en;

/**
 * Supported languages in Thunder applications
 */
export type SupportedLanguage = 'en' | 'si';

/**
 * Language configuration
 */
export interface LanguageConfig {
  code: SupportedLanguage;
  name: string;
  nativeName: string;
  direction: 'ltr' | 'rtl';
}

/**
 * i18n initialization options
 */
export interface I18nOptions {
  /**
   * Default language to use
   * @default 'en'
   */
  defaultLanguage?: SupportedLanguage;

  /**
   * Fallback language when translation is missing
   * @default 'en'
   */
  fallbackLanguage?: SupportedLanguage;

  /**
   * Enable debug mode
   * @default false
   */
  debug?: boolean;

  /**
   * Namespace to use for translations
   * @default 'translation'
   */
  namespace?: string;

  /**
   * Enable language detection from browser/storage
   * @default true
   */
  detectLanguage?: boolean;

  /**
   * Storage key for persisting language preference
   * @default 'thunder-language'
   */
  storageKey?: string;
}

/**
 * Available translation namespaces
 */
export type TranslationNamespace = 'translation';

/**
 * Translation key paths for type-safe translations
 * This creates a union type of all possible dot-notation paths in the translation object
 */
export type TranslationKey = RecursiveKeyOf<TranslationResources>;

/**
 * Utility type to get all nested keys as dot-notation strings
 */
type RecursiveKeyOf<TObj extends object> = {
  [TKey in keyof TObj & (string | number)]: TObj[TKey] extends object
    ? `${TKey}` | `${TKey}.${RecursiveKeyOf<TObj[TKey]>}`
    : `${TKey}`;
}[keyof TObj & (string | number)];

/**
 * Available translations by language
 */
export interface Translations {
  en: typeof en;
  si: typeof si;
}

/**
 * Language metadata
 */
export const LANGUAGE_CONFIGS: Record<SupportedLanguage, LanguageConfig> = {
  en: {
    code: 'en',
    name: 'English',
    nativeName: 'English',
    direction: 'ltr',
  },
  si: {
    code: 'si',
    name: 'Sinhala',
    nativeName: 'සිංහල',
    direction: 'ltr',
  },
};

/**
 * Type guard to check if a string is a supported language
 */
export function isSupportedLanguage(lang: string): lang is SupportedLanguage {
  return lang === 'en' || lang === 'si';
}
