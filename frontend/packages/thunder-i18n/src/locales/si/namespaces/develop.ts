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
 * Translations (Sinhala) for Thunder Develop application
 */
export const develop = {
  // Page titles
  pages: {
    users: 'පරිශීලකයින්',
    userTypes: 'පරිශීලක වර්ග',
    integrations: 'ඒකාබද්ධ කිරීම්',
    applications: 'යෙදුම්',
    dashboard: 'උපකරණ පුවරුව',
  },

  // Users management
  users: {
    title: 'පරිශීලක කළමනාකරණය',
    addUser: 'පරිශීලකයා එකතු කරන්න',
    editUser: 'පරිශීලකයා සංස්කරණය කරන්න',
    deleteUser: 'පරිශීලකයා මකන්න',
    userDetails: 'පරිශීලක විස්තර',
    firstName: 'මුල් නම',
    lastName: 'අවසාන නම',
    email: 'විද්‍යුත් තැපැල් ලිපිනය',
    username: 'පරිශීලක නාමය',
    role: 'භූමිකාව',
    status: 'තත්ත්වය',
    createdAt: 'නිර්මාණය කළ දිනය',
    lastLogin: 'අවසාන පිවිසුම',
    actions: 'ක්‍රියාමාර්ග',
    noUsers: 'පරිශීලකයින් හමු නොවිණි',
    searchUsers: 'පරිශීලකයින් සොයන්න...',
    confirmDeleteUser: 'ඔබට මෙම පරිශීලකයා මැකීමට අවශ්‍ය බව විශ්වාසද?',
    userCreatedSuccess: 'පරිශීලකයා සාර්ථකව නිර්මාණය කරන ලදී',
    userUpdatedSuccess: 'පරිශීලකයා සාර්ථකව යාවත්කාලීන කරන ලදී',
    userDeletedSuccess: 'පරිශීලකයා සාර්ථකව මකා දමන ලදී',
  },

  // User Types
  userTypes: {
    title: 'පරිශීලක වර්ග',
    addUserType: 'පරිශීලක වර්ගය එකතු කරන්න',
    createUserType: 'පරිශීලක වර්ගය නිර්මාණය කරන්න',
    editUserType: 'පරිශීලක වර්ගය සංස්කරණය කරන්න',
    deleteUserType: 'පරිශීලක වර්ගය මකන්න',
    userTypeDetails: 'පරිශීලක වර්ග විස්තර',
    typeName: 'වර්ග නාමය',
    typeNamePlaceholder: 'උදා:, සේවකයා, පාරිභෝගිකයා, හවුල්කරු',
    description: 'විස්තරය',
    createDescription: 'ඔබේ සංවිධානය සඳහා නව පරිශීලක වර්ග ක්‍රමයක් නිර්වචනය කරන්න',
    permissions: 'අවසර',
    schemaProperties: 'ක්‍රම ගුණාංග',
    propertyName: 'ගුණාංගයේ නම',
    propertyNamePlaceholder: 'උදා:, විද්‍යුත් තැපෑල, වයස, ලිපිනය',
    propertyType: 'වර්ගය',
    addProperty: 'ගුණාංගය එකතු කරන්න',
    unique: 'අද්විතීය',
    regexPattern: 'නිතිමත් ප්‍රකාශන රටාව (විකල්ප)',
    regexPlaceholder: 'උදා:, ^[a-zA-Z0-9]+$',
    enumValues: 'අනුමත අගයන් (Enum) - විකල්ප',
    enumPlaceholder: 'අගය එකතු කර Enter ඔබන්න',
    types: {
      string: 'අක්ෂර මාලාව',
      number: 'අංකය',
      boolean: 'බූලියන්',
      enum: 'ගණනය',
      array: 'අරාව',
      object: 'වස්තුව',
    },
    validationErrors: {
      nameRequired: 'කරුණාකර පරිශීලක වර්ග නාමයක් ඇතුළත් කරන්න',
      propertiesRequired: 'කරුණාකර අවම වශයෙන් එක් ගුණාංගයක් එකතු කරන්න',
      duplicateProperties: 'අනුපිටපත් ගුණාංග නාම හමු විය: {{duplicates}}',
    },
    noUserTypes: 'පරිශීලක වර්ග හමු නොවිණි',
    confirmDeleteUserType: 'ඔබට මෙම පරිශීලක වර්ගය මැකීමට අවශ්‍ය බව විශ්වාසද?',
  },

  // Integrations
  integrations: {
    title: 'ඒකාබද්ධ කිරීම්',
    addIntegration: 'ඒකාබද්ධ කිරීම එකතු කරන්න',
    editIntegration: 'ඒකාබද්ධ කිරීම සංස්කරණය කරන්න',
    deleteIntegration: 'ඒකාබද්ධ කිරීම මකන්න',
    integrationDetails: 'ඒකාබද්ධ කිරීම් විස්තර',
    provider: 'සපයන්නා',
    apiKey: 'API යතුර',
    endpoint: 'අන්ත ලක්ෂ්‍යය',
    status: 'තත්ත්වය',
    connected: 'සම්බන්ධ කර ඇත',
    disconnected: 'විසන්ධි කර ඇත',
    testConnection: 'සම්බන්ධතාවය පරීක්ෂා කරන්න',
    noIntegrations: 'ඒකාබද්ධ කිරීම් හමු නොවිණි',
  },

  // Applications
  applications: {
    title: 'යෙදුම්',
    addApplication: 'යෙදුම එකතු කරන්න',
    editApplication: 'යෙදුම සංස්කරණය කරන්න',
    deleteApplication: 'යෙදුම මකන්න',
    applicationDetails: 'යෙදුම් විස්තර',
    appName: 'යෙදුමේ නම',
    appType: 'යෙදුමේ වර්ගය',
    clientId: 'සේවාදායක හැඳුනුම්පත',
    clientSecret: 'සේවාදායක රහස',
    redirectUri: 'නැවත යොමු කිරීමේ URI',
    allowedOrigins: 'අනුමත මූලාශ්‍ර',
    noApplications: 'යෙදුම් හමු නොවිණි',
  },

  // Dashboard
  dashboard: {
    welcomeMessage: 'Thunder Develop වෙත සාදරයෙන් පිළිගනිමු',
    totalUsers: 'සම්පූර්ණ පරිශීලකයින්',
    activeUsers: 'සක්‍රීය පරිශීලකයින්',
    totalApplications: 'සම්පූර්ණ යෙදුම්',
    recentActivity: 'මෑත ක්‍රියාකාරකම්',
    quickActions: 'ඉක්මන් ක්‍රියාමාර්ග',
  },
} as const;

export default develop;
