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

// Client
export {default as ThunderIDExpressClient} from './ThunderIDExpressClient';

// Middleware
export {default as thunderIDExpressAuth} from './middleware/authentication';
export {default as protectRoute} from './middleware/protectRoute';

// Models
export type {ExpressClientConfig, ThunderIDExpressConfig, StrictExpressClientConfig, CookieOptions} from './models/config';
export type {UnauthenticatedCallback} from './models/protectRoute';

// Constants
export {default as CookieConfig, SESSION_COOKIE_NAME, DEFAULT_LOGIN_PATH, DEFAULT_LOGOUT_PATH} from './constants/CookieConfig';

// Re-export everything from the Node SDK
export * from '@thunderid/node';
