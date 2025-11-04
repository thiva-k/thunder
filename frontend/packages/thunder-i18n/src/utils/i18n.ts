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
 * Core i18n configuration and initialization utilities
 */

import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import type {I18nOptions, SupportedLanguage, TranslationResources} from '../types';
import {isSupportedLanguage, LANGUAGE_CONFIGS} from '../types';

/**
 * Default i18n configuration
 */
const DEFAULT_OPTIONS: Required<I18nOptions> = {
  defaultLanguage: 'en',
  fallbackLanguage: 'en',
  debug: false,
  namespace: 'translation',
  detectLanguage: true,
  storageKey: 'thunder-language',
};

/**
 * Detects the user's preferred language from browser and localStorage
 */
export function detectLanguage(storageKey: string): SupportedLanguage {
  // Check localStorage first
  if (typeof window !== 'undefined') {
    const stored = localStorage.getItem(storageKey);
    if (stored && isSupportedLanguage(stored)) {
      return stored;
    }

    // Check browser language
    const browserLang = navigator.language.split('-')[0];
    if (isSupportedLanguage(browserLang)) {
      return browserLang;
    }
  }

  return 'en';
}

/**
 * Saves the selected language to localStorage
 */
export function saveLanguage(language: SupportedLanguage, storageKey: string): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem(storageKey, language);
  }
}

/**
 * Initializes i18next with the provided translations and options
 *
 * @example
 * ```tsx
 * import { initI18n } from '@thunder/i18n';
 * import en from '@thunder/i18n/locales/en';
 * import si from '@thunder/i18n/locales/si';
 *
 * const i18nInstance = await initI18n({
 *   translations: { en, si },
 *   options: {
 *     defaultLanguage: 'en',
 *     debug: process.env.NODE_ENV === 'development'
 *   }
 * });
 * ```
 */
export async function initI18n(config: {
  translations: Record<SupportedLanguage, TranslationResources>;
  options?: I18nOptions;
}) {
  const options = {...DEFAULT_OPTIONS, ...config.options};

  const language = options.detectLanguage ? detectLanguage(options.storageKey) : options.defaultLanguage;

  // Create resources object for i18next with namespaces
  // Resources structure: { [language]: { [namespace]: translations } }
  type ResourceValue = string | {[key: string]: ResourceValue};
  type NamespaceResources = {[key: string]: ResourceValue};
  const resources: Record<string, NamespaceResources> = {};

  Object.entries(config.translations).forEach(([lang, translations]) => {
    // If translations are already namespaced (e.g., { common: {...}, develop: {...} })
    if (
      translations &&
      typeof translations === 'object' &&
      ('common' in translations || 'develop' in translations || 'gate' in translations)
    ) {
      resources[lang] = translations;
    } else {
      // Fallback to single namespace for backward compatibility
      resources[lang] = {
        [options.namespace]: translations,
      };
    }
  });

  // Extract all unique namespaces from resources
  const allNamespaces = new Set<string>();
  Object.values(resources).forEach((langResource) => {
    Object.keys(langResource).forEach((ns) => allNamespaces.add(ns));
  });
  const namespaces = Array.from(allNamespaces);

  await i18n.use(initReactI18next).init({
    resources,
    lng: language,
    fallbackLng: options.fallbackLanguage,
    ns: namespaces.length > 0 ? namespaces : [options.namespace],
    defaultNS: namespaces.includes('common') ? 'common' : options.namespace,
    fallbackNS: 'common',
    debug: options.debug,

    interpolation: {
      escapeValue: false, // React already escapes values
    },

    react: {
      useSuspense: true,
    },

    // Return key if translation is missing (useful for development)
    returnNull: false,
    returnEmptyString: false,
  });

  // Save the detected/selected language
  saveLanguage(language as SupportedLanguage, options.storageKey);

  return i18n;
}

/**
 * Changes the current language and persists the choice
 */
export async function changeLanguage(
  language: SupportedLanguage,
  storageKey: string = DEFAULT_OPTIONS.storageKey,
): Promise<void> {
  await i18n.changeLanguage(language);
  saveLanguage(language, storageKey);

  // Update HTML lang attribute for accessibility
  if (typeof document !== 'undefined') {
    document.documentElement.lang = language;
    document.documentElement.dir = LANGUAGE_CONFIGS[language].direction;
  }
}

/**
 * Gets the current language
 */
export function getCurrentLanguage(): SupportedLanguage {
  const lang = i18n.language;
  return isSupportedLanguage(lang) ? lang : 'en';
}

/**
 * Gets all available languages with metadata
 */
export function getAvailableLanguages() {
  return Object.values(LANGUAGE_CONFIGS);
}

/**
 * Formats a date according to the current language
 */
export function formatDate(date: Date, options?: Intl.DateTimeFormatOptions): string {
  const currentLang = getCurrentLanguage();
  return new Intl.DateTimeFormat(currentLang, options).format(date);
}

/**
 * Formats a number according to the current language
 */
export function formatNumber(value: number, options?: Intl.NumberFormatOptions): string {
  const currentLang = getCurrentLanguage();
  return new Intl.NumberFormat(currentLang, options).format(value);
}

/**
 * Formats a currency according to the current language
 */
export function formatCurrency(value: number, currency: string = 'USD', options?: Intl.NumberFormatOptions): string {
  const currentLang = getCurrentLanguage();
  return new Intl.NumberFormat(currentLang, {
    style: 'currency',
    currency,
    ...options,
  }).format(value);
}
