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

import {AuthClientConfig} from '@thunderid/node';

/**
 * Express-specific cookie configuration options.
 */
export interface CookieOptions {
  /** Cookie max age in milliseconds. */
  maxAge?: number;
  /** Whether the cookie is HTTP-only. */
  httpOnly?: boolean;
  /** SameSite policy. */
  sameSite?: string;
  /** Whether the cookie requires HTTPS. */
  secure?: boolean;
}

/**
 * Express-specific configuration fields.
 */
export interface StrictExpressClientConfig {
  /** The base URL of the Express application (e.g. `http://localhost:3000`). */
  appURL: string;
  /** Cookie configuration for the session cookie. */
  cookieConfig?: CookieOptions;
  /** Whether to apply global authentication middleware. */
  globalAuth?: boolean;
  /** Custom login path. Defaults to `/login`. */
  loginPath?: string;
  /** Custom logout path. Defaults to `/logout`. */
  logoutPath?: string;
  /** Additional parameters to include in the sign-in request. */
  signInConfig?: Record<string, string | boolean>;
}

/**
 * Full configuration type for `ThunderIDExpressClient`.
 * Combines node-level auth config with Express-specific settings.
 */
export type ExpressClientConfig = Exclude<AuthClientConfig, 'afterSignInUrl' | 'afterSignOutUrl'> &
  StrictExpressClientConfig;

/**
 * Configuration type for the ThunderID Express.js SDK.
 */
export type ThunderIDExpressConfig = ExpressClientConfig;
