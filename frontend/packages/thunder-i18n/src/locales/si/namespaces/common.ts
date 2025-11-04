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
 * Common translations (Sinhala) shared across all Thunder applications
 */
export const common = {
  // Actions
  actions: {
    add: 'එකතු කරන්න',
    edit: 'සංස්කරණය කරන්න',
    delete: 'මකන්න',
    cancel: 'අවලංගු කරන්න',
    save: 'සුරකින්න',
    create: 'නිර්මාණය කරන්න',
    update: 'යාවත්කාලීන කරන්න',
    remove: 'ඉවත් කරන්න',
    search: 'සොයන්න',
    filter: 'පෙරීම',
    reset: 'යළි සකසන්න',
    submit: 'ඉදිරිපත් කරන්න',
    close: 'වසන්න',
    back: 'ආපසු',
    next: 'ඊළඟ',
    previous: 'පෙර',
    confirm: 'තහවුරු කරන්න',
    ok: 'හරි',
    yes: 'ඔව්',
    no: 'නැත',
    continue: 'ඉදිරියට යන්න',
    skip: 'මඟ හරින්න',
    finish: 'අවසන් කරන්න',
    refresh: 'නැවුම් කරන්න',
    copy: 'පිටපත් කරන්න',
    download: 'බාගන්න',
    upload: 'උඩුගත කරන්න',
    export: 'අපනයනය කරන්න',
    import: 'ආනයනය කරන්න',
    view: 'බලන්න',
    details: 'විස්තර',
    settings: 'සැකසුම්',
    logout: 'පිටවන්න',
    login: 'ඇතුල් වන්න',
  },

  // Status messages
  status: {
    loading: 'පූරණය වෙමින්...',
    saving: 'සුරකිමින්...',
    success: 'සාර්ථකයි',
    error: 'දෝෂයකි',
    warning: 'අවවාදයයි',
    info: 'තොරතුරු',
    pending: 'පොරොත්තුවෙන්',
    active: 'සක්‍රීයයි',
    inactive: 'අක්‍රීයයි',
    enabled: 'සක්‍රීය කර ඇත',
    disabled: 'අක්‍රීය කර ඇත',
    completed: 'සම්පූර්ණයි',
    failed: 'අසාර්ථකයි',
  },

  // Form labels
  form: {
    name: 'නම',
    description: 'විස්තරය',
    email: 'විද්‍යුත් තැපැල',
    password: 'මුරපදය',
    username: 'පරිශීලක නාමය',
    required: 'අවශ්‍යයි',
    optional: 'විකල්ප',
    requiredField: 'මෙම ක්ෂේත්‍රය අවශ්‍යයි',
    invalidEmail: 'අවලංගු විද්‍යුත් තැපැල් ලිපිනයක්',
    invalidFormat: 'අවලංගු ආකෘතිය',
    searchPlaceholder: 'සොයන්න...',
  },

  // Messages
  messages: {
    confirmDelete: 'ඔබට මෙම අයිතමය මැකීමට අවශ්‍ය බව විශ්වාසද?',
    deleteSuccess: 'අයිතමය සාර්ථකව මකා දමන ලදී',
    deleteError: 'අයිතමය මැකීමට අසමර්ථ විය',
    saveSuccess: 'සාර්ථකව සුරකින ලදී',
    saveError: 'සුරැකීමට අසමර්ථ විය',
    updateSuccess: 'සාර්ථකව යාවත්කාලීන කරන ලදී',
    updateError: 'යාවත්කාලීන කිරීමට අසමර්ථ විය',
    createSuccess: 'සාර්ථකව නිර්මාණය කරන ලදී',
    createError: 'නිර්මාණය කිරීමට අසමර්ථ විය',
    noData: 'දත්ත නොමැත',
    noResults: 'ප්‍රතිඵල හමු නොවිණි',
    somethingWentWrong: 'යමක් වැරදී ඇත',
    tryAgain: 'කරුණාකර නැවත උත්සාහ කරන්න',
  },

  // Navigation
  navigation: {
    home: 'මුල් පිටුව',
    dashboard: 'උපකරණ පුවරුව',
    profile: 'පැතිකඩ',
    help: 'උදව්',
    documentation: 'ලේඛන',
  },
} as const;

export default common;
