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

import {useTranslation} from 'react-i18next';
import {useCallback, useMemo} from 'react';
import type {SupportedLanguage} from '../types';
import {changeLanguage, getCurrentLanguage, getAvailableLanguages} from '../utils/i18n';

/**
 * Hook for managing language state and switching
 *
 * @example
 * ```tsx
 * function LanguageSwitcher() {
 *   const { currentLanguage, availableLanguages, setLanguage } = useLanguage();
 *
 *   return (
 *     <select value={currentLanguage} onChange={(e) => setLanguage(e.target.value as SupportedLanguage)}>
 *       {availableLanguages.map((lang) => (
 *         <option key={lang.code} value={lang.code}>
 *           {lang.nativeName}
 *         </option>
 *       ))}
 *     </select>
 *   );
 * }
 * ```
 */
export function useLanguage() {
  // useTranslation() ensures this component re-renders when language changes
  useTranslation();

  // getCurrentLanguage() is lightweight, no need for useMemo
  const currentLanguage = getCurrentLanguage();

  const availableLanguages = useMemo(() => getAvailableLanguages(), []);

  const setLanguage = useCallback(async (language: SupportedLanguage) => {
    await changeLanguage(language);
  }, []);

  const toggleLanguage = useCallback(async () => {
    const currentLang = getCurrentLanguage();
    const nextLang: SupportedLanguage = currentLang === 'en' ? 'si' : 'en';
    await changeLanguage(nextLang);
  }, []);

  return {
    currentLanguage,
    availableLanguages,
    setLanguage,
    toggleLanguage,
  };
}
