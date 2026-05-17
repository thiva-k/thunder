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

import {logger as Logger} from '@thunderid/node';
import express from 'express';
import ThunderIDExpressClient from '../ThunderIDExpressClient';
import {UnauthenticatedCallback} from '../models/protectRoute';
import {SESSION_COOKIE_NAME} from '../constants/CookieConfig';

/**
 * Returns Express middleware that blocks unauthenticated requests.
 * Invokes `callback` when the session is missing or invalid.
 *
 * @param client - The initialized ThunderIDExpressClient instance.
 * @param callback - Called with the response and error message when the request is unauthenticated.
 *   Return `true` to indicate the response is handled; `false` to call `next()`.
 */
const protectRoute = (
  client: ThunderIDExpressClient,
  callback: UnauthenticatedCallback,
): ((req: express.Request, res: express.Response, next: express.NextFunction) => Promise<void>) => {
  return async (req: express.Request, res: express.Response, next: express.NextFunction): Promise<void> => {
    const sessionId: string | undefined = req.cookies?.[SESSION_COOKIE_NAME];

    if (!sessionId) {
      Logger.error('No session ID found in the request cookies');
      if (callback(res, 'Unauthenticated')) {
        return;
      }
      return next();
    }

    const isValid: boolean = await client.isSignedIn(sessionId);

    if (isValid) {
      return next();
    }

    Logger.error('Invalid session ID found in the request cookies');
    if (callback(res, 'Invalid session cookie')) {
      return;
    }
    return next();
  };
};

export default protectRoute;
