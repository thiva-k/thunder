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
 * Translations (Sinhala) for Thunder Gate application
 */
export const gate = {
  // Authentication
  auth: {
    signIn: 'පුරනය වන්න',
    signUp: 'ලියාපදිංචිය',
    signOut: 'පිටවන්න',
    forgotPassword: 'මුරපදය අමතකද?',
    resetPassword: 'මුරපදය යළි සකසන්න',
    changePassword: 'මුරපදය වෙනස් කරන්න',
    rememberMe: 'මා මතක තබා ගන්න',
    welcomeBack: 'නැවත සාදරයෙන් පිළිගනිමු',
    createAccount: 'ගිණුමක් සාදන්න',
    alreadyHaveAccount: 'දැනටමත් ගිණුමක් තිබේද?',
    dontHaveAccount: 'ගිණුමක් නොමැතිද?',
    enterEmail: 'ඔබේ විද්‍යුත් තැපෑල ඇතුළත් කරන්න',
    enterPassword: 'ඔබේ මුරපදය ඇතුළත් කරන්න',
    confirmPassword: 'මුරපදය තහවුරු කරන්න',
    passwordMismatch: 'මුරපද නොගැළපේ',
    invalidCredentials: 'අවලංගු අක්තපත්‍ර',
    accountLocked: 'ගිණුම අගුළු දමා ඇත',
    sessionExpired: 'සැසිය කල් ඉකුත් වී ඇත. කරුණාකර නැවත පුරනය වන්න.',
    signInSuccess: 'සාර්ථකව පුරනය විය',
    signUpSuccess: 'ගිණුම සාර්ථකව නිර්මාණය කරන ලදී',
    passwordResetSent: 'මුරපදය යළි සැකසීමේ සබැඳිය ඔබගේ විද්‍යුත් තැපෑලට යවන ලදී',
    passwordResetSuccess: 'මුරපදය සාර්ථකව යළි සකසන ලදී',
  },

  // Multi-factor authentication
  mfa: {
    title: 'බහු-සාධක සත්‍යාපනය',
    setupMfa: 'MFA පිහිටුවන්න',
    enableMfa: 'MFA සක්‍රීය කරන්න',
    disableMfa: 'MFA අක්‍රීය කරන්න',
    verificationCode: 'සත්‍යාපන කේතය',
    enterCode: 'සත්‍යාපන කේතය ඇතුළත් කරන්න',
    sendCode: 'කේතය යවන්න',
    resendCode: 'කේතය නැවත යවන්න',
    invalidCode: 'අවලංගු සත්‍යාපන කේතයකි',
    codeExpired: 'සත්‍යාපන කේතය කල් ඉකුත් වී ඇත',
    scanQrCode: 'ඔබගේ සත්‍යාපක යෙදුම සමඟ මෙම QR කේතය ස්කෑන් කරන්න',
    backupCodes: 'උපස්ථ කේත',
    saveBackupCodes: 'මෙම උපස්ථ කේත ආරක්ෂිත ස්ථානයක සුරකින්න',
  },

  // Social login
  social: {
    continueWith: 'සමඟ ඉදිරියට යන්න',
    signInWith: 'සමඟ පුරනය වන්න',
    google: 'Google',
    facebook: 'Facebook',
    github: 'GitHub',
    microsoft: 'Microsoft',
  },

  // Consent
  consent: {
    title: 'එකඟතාව අවශ්‍යයි',
    message: 'යෙදුම ඔබගේ තොරතුරු වෙත ප්‍රවේශ වීමට ඉල්ලා සිටී',
    requestedPermissions: 'ඉල්ලා සිටින අවසර',
    allow: 'ඉඩ දෙන්න',
    deny: 'ප්‍රතික්ෂේප කරන්න',
    learnMore: 'වැඩිදුර ඉගෙන ගන්න',
  },

  // Errors
  errors: {
    authenticationFailed: 'සත්‍යාපනය අසාර්ථක විය',
    unauthorizedAccess: 'අනවසර ප්‍රවේශය',
    accessDenied: 'ප්‍රවේශය ප්‍රතික්ෂේප විය',
    invalidRequest: 'අවලංගු ඉල්ලීමකි',
    serverError: 'සේවාදායක දෝෂයක් සිදු විය',
    networkError: 'ජාල දෝෂයකි. කරුණාකර ඔබේ සම්බන්ධතාවය පරීක්ෂා කරන්න.',
    redirectFailed: 'නැවත යොමු කිරීම අසාර්ථක විය',
  },
} as const;

export default gate;
