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
 * Translations for Thunder Develop application
 */
export const develop = {
  // Page titles
  pages: {
    users: 'Users',
    userTypes: 'User Types',
    integrations: 'Integrations',
    applications: 'Applications',
    dashboard: 'Dashboard',
  },

  // Users management
  users: {
    title: 'User Management',
    addUser: 'Add User',
    editUser: 'Edit User',
    deleteUser: 'Delete User',
    userDetails: 'User Details',
    firstName: 'First Name',
    lastName: 'Last Name',
    email: 'Email Address',
    username: 'Username',
    role: 'Role',
    status: 'Status',
    createdAt: 'Created At',
    lastLogin: 'Last Login',
    actions: 'Actions',
    noUsers: 'No users found',
    searchUsers: 'Search users...',
    confirmDeleteUser: 'Are you sure you want to delete this user?',
    userCreatedSuccess: 'User created successfully',
    userUpdatedSuccess: 'User updated successfully',
    userDeletedSuccess: 'User deleted successfully',
  },

  // User Types
  userTypes: {
    title: 'User Types',
    addUserType: 'Add User Type',
    createUserType: 'Create User Type',
    editUserType: 'Edit User Type',
    deleteUserType: 'Delete User Type',
    userTypeDetails: 'User Type Details',
    typeName: 'Type Name',
    typeNamePlaceholder: 'e.g., Employee, Customer, Partner',
    description: 'Description',
    createDescription: 'Define a new user type schema for your organization',
    permissions: 'Permissions',
    schemaProperties: 'Schema Properties',
    propertyName: 'Property Name',
    propertyNamePlaceholder: 'e.g., email, age, address',
    propertyType: 'Type',
    addProperty: 'Add Property',
    unique: 'Unique',
    regexPattern: 'Regular Expression Pattern (Optional)',
    regexPlaceholder: 'e.g., ^[a-zA-Z0-9]+$',
    enumValues: 'Allowed Values (Enum) - Optional',
    enumPlaceholder: 'Add value and press Enter',
    types: {
      string: 'String',
      number: 'Number',
      boolean: 'Boolean',
      enum: 'Enum',
      array: 'Array',
      object: 'Object',
    },
    validationErrors: {
      nameRequired: 'Please enter a user type name',
      propertiesRequired: 'Please add at least one property',
      duplicateProperties: 'Duplicate property names found: {{duplicates}}',
    },
    noUserTypes: 'No user types found',
    confirmDeleteUserType: 'Are you sure you want to delete this user type?',
  },

  // Integrations
  integrations: {
    title: 'Integrations',
    addIntegration: 'Add Integration',
    editIntegration: 'Edit Integration',
    deleteIntegration: 'Delete Integration',
    integrationDetails: 'Integration Details',
    provider: 'Provider',
    apiKey: 'API Key',
    endpoint: 'Endpoint',
    status: 'Status',
    connected: 'Connected',
    disconnected: 'Disconnected',
    testConnection: 'Test Connection',
    noIntegrations: 'No integrations found',
  },

  // Applications
  applications: {
    title: 'Applications',
    addApplication: 'Add Application',
    editApplication: 'Edit Application',
    deleteApplication: 'Delete Application',
    applicationDetails: 'Application Details',
    appName: 'Application Name',
    appType: 'Application Type',
    clientId: 'Client ID',
    clientSecret: 'Client Secret',
    redirectUri: 'Redirect URI',
    allowedOrigins: 'Allowed Origins',
    noApplications: 'No applications found',
  },

  // Dashboard
  dashboard: {
    welcomeMessage: 'Welcome to Thunder Develop',
    totalUsers: 'Total Users',
    activeUsers: 'Active Users',
    totalApplications: 'Total Applications',
    recentActivity: 'Recent Activity',
    quickActions: 'Quick Actions',
  },
} as const;

export default develop;
