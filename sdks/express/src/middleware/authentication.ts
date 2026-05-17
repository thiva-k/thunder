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

import {ThunderIDAuthException, TokenResponse, logger as Logger} from '@thunderid/node';
import express from 'express';
import ThunderIDExpressClient from '../ThunderIDExpressClient';
import {DEFAULT_LOGIN_PATH, DEFAULT_LOGOUT_PATH, SESSION_COOKIE_NAME} from '../constants/CookieConfig';
import {ExpressClientConfig} from '../models/config';

/**
 * Returns an Express router that wires the `/login` and `/logout` routes.
 * Patches `thunderIDAuth` onto `req` and `res` for use in downstream handlers.
 *
 * @param client - The initialized ThunderIDExpressClient instance.
 * @param config - The Express client configuration.
 * @param onSignIn - Called with the response and token response on successful sign-in.
 * @param onSignOut - Called with the response on successful sign-out.
 * @param onError - Called with the response and exception on error.
 */
const thunderIDExpressAuth = (
  client: ThunderIDExpressClient,
  config: ExpressClientConfig,
  onSignIn: (res: express.Response, tokenResponse: TokenResponse) => void,
  onSignOut: (res: express.Response) => void,
  onError: (res: express.Response, exception: ThunderIDAuthException) => void,
): express.Router => {
  const router: express.Router = express.Router();

  router.use(async (req: express.Request, res: express.Response, next: express.NextFunction): Promise<void> => {
    (req as any).thunderIDAuth = client;
    (res as any).thunderIDAuth = client;
    next();
  });

  router.get(
    config.loginPath || DEFAULT_LOGIN_PATH,
    async (req: express.Request, res: express.Response, next: express.NextFunction): Promise<void> => {
      try {
        const response: TokenResponse = await client.signIn(req, res, next, config.signInConfig);
        if (response.accessToken || response.idToken) {
          onSignIn(res, response);
        }
      } catch (e: any) {
        Logger.error(e.message);
        onError(res, e);
      }
    },
  );

  router.get(
    config.logoutPath || DEFAULT_LOGOUT_PATH,
    async (req: express.Request, res: express.Response, next: express.NextFunction): Promise<void> => {
      if ((req.query as any).state === 'sign_out_success') {
        onSignOut(res);
        return;
      }

      const sessionId: string | undefined = req.cookies?.[SESSION_COOKIE_NAME];

      if (!sessionId) {
        onError(
          res,
          new ThunderIDAuthException(
            'EXPRESS-AUTH_MW-LOGOUT-NF01',
            'No cookie found in the request',
            'No cookie was sent with the request. The user may not have signed in yet.',
          ),
        );
        return;
      }

      try {
        const signOutURL: string = await client.signOut(sessionId);
        if (signOutURL) {
          res.cookie(SESSION_COOKIE_NAME, null, {maxAge: 0});
          res.redirect(signOutURL);
        }
      } catch (e: any) {
        onError(res, e);
      }
    },
  );

  return router;
};

export default thunderIDExpressAuth;
