/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {ThunderIDAuthException, Storage, TokenResponse, Logger} from '@thunderid/node';
import express from 'express';
import {ThunderIDExpressClient} from '../client';
import {DEFAULT_LOGIN_PATH, DEFAULT_LOGOUT_PATH} from '../constants';
import {ExpressClientConfig} from '../models';

export const thunderIDExpressAuth = (
  thunderIDExpressClient: ThunderIDExpressClient,
  config: ExpressClientConfig,
  onSignIn: (res: express.Response, tokenResponse: TokenResponse) => void,
  onSignOut: (res: express.Response) => void,
  onError: (res: express.Response, exception: ThunderIDAuthException) => void,
): any => {
  //Create the router
  const router = new express.Router();

  //Patch AuthClient to the request and the response
  router.use(async (req: express.Request, res: express.Response, next: express.nextFunction): Promise<void> => {
    req.thunderIDAuth = thunderIDExpressClient;
    res.thunderIDAuth = thunderIDExpressClient;
    next();
  });

  //Patch in '/login' route
  router.get(
    config.loginPath || DEFAULT_LOGIN_PATH,
    async (req: express.Request, res: express.Response, next: express.nextFunction): Promise<void> => {
      try {
        const response: TokenResponse = await thunderIDExpressClient.signIn(req, res, next, config.signInConfig);
        if (response.accessToken || response.idToken) {
          onSignIn(res, response);
        }
      } catch (e: any) {
        Logger.error(e.message);
        onError(res, e);
      }
    },
  );

  //Patch in '/logout' route
  router.get(
    config.logoutPath || DEFAULT_LOGOUT_PATH,
    async (req: express.Request, res: express.Response, next: express.nextFunction): Promise<void> => {
      //Check if it is a logout success response
      if (req.query.state === 'sign_out_success') {
        onSignOut(res);

        return;
      }

      //Check if the cookie exists
      if (req.cookies.THUNDERID_SESSION_ID === undefined) {
        onError(
          res,
          new ThunderIDAuthException(
            'EXPRESS-AUTH_MW-LOGOUT-NF01',
            'No cookie found in the request',
            'No cookie was sent with the request. The user may not have signed in yet.',
          ),
        );

        return;
      } else {
        //Get the signout URL
        try {
          const signOutURL = await req.thunderIDAuth.signOut(req.cookies.THUNDERID_SESSION_ID);
          if (signOutURL) {
            res.cookie('THUNDERID_SESSION_ID', null, {maxAge: 0});
            res.redirect(signOutURL);

            return;
          }
        } catch (e: any) {
          onError(res, e);

          return;
        }
      }
    },
  );

  return router;
};
