/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import AppConfigDefaults from './app.json';

interface AppConfig {
  productName: string;
  flowExecutionEndpoint: string;
  authorizationEndpoint: string;
}

/**
 * Application configuration that uses environment variables with fallbacks to JSON defaults.
 * Environment variables take precedence over JSON values when available.
 */
const getThunderHost = (): string => {
  return process.env.NEXT_PUBLIC_THUNDER_HOST || 'http://0.0.0.0:8090';
};

export const AppConfig: AppConfig = {
  productName: process.env.NEXT_PUBLIC_PRODUCT_NAME || AppConfigDefaults.productName,
  flowExecutionEndpoint: process.env.NEXT_PUBLIC_FLOW_EXECUTION_ENDPOINT || `${getThunderHost()}/flow/execute`,
  authorizationEndpoint: process.env.NEXT_PUBLIC_AUTHORIZATION_ENDPOINT || `${getThunderHost()}/oauth2/authorize`,
};

export default AppConfig;
