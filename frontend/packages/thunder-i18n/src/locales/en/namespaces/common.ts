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
 * Common translations shared across all Thunder applications
 */
export const common = {
  // Actions
  actions: {
    add: 'Add',
    edit: 'Edit',
    delete: 'Delete',
    cancel: 'Cancel',
    save: 'Save',
    create: 'Create',
    update: 'Update',
    remove: 'Remove',
    search: 'Search',
    filter: 'Filter',
    reset: 'Reset',
    submit: 'Submit',
    close: 'Close',
    back: 'Back',
    next: 'Next',
    previous: 'Previous',
    confirm: 'Confirm',
    ok: 'OK',
    yes: 'Yes',
    no: 'No',
    continue: 'Continue',
    skip: 'Skip',
    finish: 'Finish',
    refresh: 'Refresh',
    copy: 'Copy',
    download: 'Download',
    upload: 'Upload',
    export: 'Export',
    import: 'Import',
    view: 'View',
    details: 'Details',
    settings: 'Settings',
    logout: 'Logout',
    login: 'Login',
  },

  // Status messages
  status: {
    loading: 'Loading...',
    saving: 'Saving...',
    success: 'Success',
    error: 'Error',
    warning: 'Warning',
    info: 'Info',
    pending: 'Pending',
    active: 'Active',
    inactive: 'Inactive',
    enabled: 'Enabled',
    disabled: 'Disabled',
    completed: 'Completed',
    failed: 'Failed',
  },

  // Form labels
  form: {
    name: 'Name',
    description: 'Description',
    email: 'Email',
    password: 'Password',
    username: 'Username',
    required: 'Required',
    optional: 'Optional',
    requiredField: 'This field is required',
    invalidEmail: 'Invalid email address',
    invalidFormat: 'Invalid format',
    searchPlaceholder: 'Search...',
  },

  // Messages
  messages: {
    confirmDelete: 'Are you sure you want to delete this item?',
    deleteSuccess: 'Item deleted successfully',
    deleteError: 'Failed to delete item',
    saveSuccess: 'Saved successfully',
    saveError: 'Failed to save',
    updateSuccess: 'Updated successfully',
    updateError: 'Failed to update',
    createSuccess: 'Created successfully',
    createError: 'Failed to create',
    noData: 'No data available',
    noResults: 'No results found',
    somethingWentWrong: 'Something went wrong',
    tryAgain: 'Please try again',
  },

  // Navigation
  navigation: {
    home: 'Home',
    dashboard: 'Dashboard',
    profile: 'Profile',
    help: 'Help',
    documentation: 'Documentation',
  },
} as const;

export default common;
