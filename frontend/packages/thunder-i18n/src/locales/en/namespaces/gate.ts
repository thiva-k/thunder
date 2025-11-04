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
 * Translations for Thunder Gate application
 */
export const gate = {
  // Authentication
  auth: {
    signIn: 'Sign In',
    signUp: 'Sign Up',
    signOut: 'Sign Out',
    forgotPassword: 'Forgot Password?',
    resetPassword: 'Reset Password',
    changePassword: 'Change Password',
    rememberMe: 'Remember Me',
    welcomeBack: 'Welcome Back',
    createAccount: 'Create Account',
    alreadyHaveAccount: 'Already have an account?',
    dontHaveAccount: "Don't have an account?",
    enterEmail: 'Enter your email',
    enterPassword: 'Enter your password',
    confirmPassword: 'Confirm Password',
    passwordMismatch: 'Passwords do not match',
    invalidCredentials: 'Invalid credentials',
    accountLocked: 'Account is locked',
    sessionExpired: 'Session expired. Please sign in again.',
    signInSuccess: 'Signed in successfully',
    signUpSuccess: 'Account created successfully',
    passwordResetSent: 'Password reset link sent to your email',
    passwordResetSuccess: 'Password reset successfully',
  },

  // Multi-factor authentication
  mfa: {
    title: 'Multi-Factor Authentication',
    setupMfa: 'Set Up MFA',
    enableMfa: 'Enable MFA',
    disableMfa: 'Disable MFA',
    verificationCode: 'Verification Code',
    enterCode: 'Enter verification code',
    sendCode: 'Send Code',
    resendCode: 'Resend Code',
    invalidCode: 'Invalid verification code',
    codeExpired: 'Verification code expired',
    scanQrCode: 'Scan this QR code with your authenticator app',
    backupCodes: 'Backup Codes',
    saveBackupCodes: 'Save these backup codes in a secure place',
  },

  // Social login
  social: {
    continueWith: 'Continue with',
    signInWith: 'Sign in with',
    google: 'Google',
    facebook: 'Facebook',
    github: 'GitHub',
    microsoft: 'Microsoft',
  },

  // Consent
  consent: {
    title: 'Consent Required',
    message: 'The application is requesting access to your information',
    requestedPermissions: 'Requested Permissions',
    allow: 'Allow',
    deny: 'Deny',
    learnMore: 'Learn More',
  },

  // Errors
  errors: {
    authenticationFailed: 'Authentication failed',
    unauthorizedAccess: 'Unauthorized access',
    accessDenied: 'Access denied',
    invalidRequest: 'Invalid request',
    serverError: 'Server error occurred',
    networkError: 'Network error. Please check your connection.',
    redirectFailed: 'Redirect failed',
  },
} as const;

export default gate;
