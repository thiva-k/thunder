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

import {
  LegacyThunderIDNodeClient,
  AuthClientConfig,
  AuthURLCallback,
  TokenResponse,
  Storage,
  User,
  OIDCEndpoints,
  IdToken,
  TokenExchangeRequestConfig,
  ThunderIDAuthException,
  Logger,
} from '@thunderid/node';
import express from 'express';
import {v4 as uuidv4} from 'uuid';
import {CookieConfig, DEFAULT_LOGIN_PATH, DEFAULT_LOGOUT_PATH} from './constants';
import {thunderIDExpressAuth, protectRoute} from './middleware';
import {ExpressClientConfig, UnauthenticatedCallback} from './models';
import {ExpressUtils} from './utils/express-utils';

export class ThunderIDExpressClient {
  private _authClient: LegacyThunderIDNodeClient<AuthClientConfig>;
  private _storage?: Storage;
  private static _clientConfig: ExpressClientConfig;

  private static _instance: ThunderIDExpressClient;

  private constructor(config: ExpressClientConfig, storage?: Storage) {
    //Set the client config
    ThunderIDExpressClient._clientConfig = {...config};

    //Add the afterSignInUrl and afterSignOutUrl
    //Add custom paths if the user has already declared any or else use the defaults
    const nodeClientConfig: AuthClientConfig = {
      ...config,
      afterSignInUrl: config.appURL + (config.loginPath || DEFAULT_LOGIN_PATH),
      afterSignOutUrl: config.appURL + (config.logoutPath || DEFAULT_LOGOUT_PATH),
    };

    //Initialize the user provided storage if there is any
    if (storage) {
      Logger.debug('Initializing user provided storage');
      this._storage = storage;
    }

    //Initialize the Auth Client
    this._authClient = new LegacyThunderIDNodeClient();
    this._authClient.initialize(nodeClientConfig, this._storage);
  }

  public static getInstance(config: ExpressClientConfig, storage?: Storage): ThunderIDExpressClient;
  public static getInstance(): ThunderIDExpressClient;
  public static getInstance(config?: ExpressClientConfig, storage?: Storage): ThunderIDExpressClient {
    //Create a new instance if its not instantiated already
    if (!ThunderIDExpressClient._instance && config) {
      ThunderIDExpressClient._instance = new ThunderIDExpressClient(config, storage);
      Logger.debug('Initialized ThunderIDExpressClient successfully');
    }

    if (!ThunderIDExpressClient._instance && !config) {
      throw Error(
        new ThunderIDAuthException(
          'EXPRESS-CLIENT-GI1-NF01',
          'User configuration  is not found',
          'User config has not been passed to initialize ThunderIDExpressClient',
        ).toString(),
      );
    }

    return ThunderIDExpressClient._instance;
  }

  public async signIn(
    req: express.Request,
    res: express.Response,
    next: express.nextFunction,
    signInConfig?: Record<string, string | boolean>,
  ): Promise<TokenResponse> {
    if (ExpressUtils.hasErrorInURL(req.originalUrl)) {
      return Promise.reject(
        new ThunderIDAuthException(
          'EXPRESS-CLIENT-SI-IV01',
          'Invalid login request URL',
          'Login request contains an error query parameter in the URL',
        ),
      );
    }

    //Check if the user has a valid user ID and if not create one
    let userId = req.cookies.THUNDERID_SESSION_ID;
    if (!userId) {
      userId = uuidv4();
    }

    //Handle signIn() callback
    const authRedirectCallback = (url: string) => {
      if (url) {
        //DEBUG
        Logger.debug('Redirecting to: ' + url);
        res.cookie('THUNDERID_SESSION_ID', userId, {
          maxAge: ThunderIDExpressClient._clientConfig.cookieConfig?.maxAge
            ? ThunderIDExpressClient._clientConfig.cookieConfig.maxAge
            : CookieConfig.defaultMaxAge,
          httpOnly: ThunderIDExpressClient._clientConfig.cookieConfig?.httpOnly ?? CookieConfig.defaultHttpOnly,
          sameSite: ThunderIDExpressClient._clientConfig.cookieConfig?.sameSite ?? CookieConfig.defaultSameSite,
          secure: ThunderIDExpressClient._clientConfig.cookieConfig?.secure ?? CookieConfig.defaultSecure,
        });
        res.redirect(url);

        next && typeof next === 'function' && next();
      }
    };

    const authResponse: TokenResponse = await this._authClient.signIn(
      authRedirectCallback,
      userId,
      req.query.code,
      req.query.session_state,
      req.query.state,
      signInConfig,
    );

    if (authResponse.accessToken || authResponse.idToken) {
      return authResponse;
    } else {
      return {
        accessToken: '',
        createdAt: 0,
        expiresIn: '',
        idToken: '',
        refreshToken: '',
        scope: '',
        tokenType: '',
      };
    }
  }

  public async signOut(userId: string): Promise<string> {
    return this._authClient.signOut(userId);
  }

  public async isSignedIn(userId: string): Promise<boolean> {
    return this._authClient.isSignedIn(userId);
  }

  public async getIdToken(userId: string): Promise<string> {
    return this._authClient.getIdToken(userId);
  }

  public async getUser(userId: string): Promise<User> {
    return this._authClient.getUser(userId);
  }

  public async getOpenIDProviderEndpoints(): Promise<OIDCEndpoints> {
    return this._authClient.getOpenIDProviderEndpoints();
  }

  public async getDecodedIdToken(userId?: string): Promise<IdToken> {
    return this._authClient.getDecodedIdToken(userId);
  }

  public async getAccessToken(userId?: string): Promise<string> {
    return this._authClient.getAccessToken(userId);
  }

  public async exchangeToken(config: TokenExchangeRequestConfig, userId?: string): Promise<TokenResponse | Response> {
    return this._authClient.exchangeToken(config, userId);
  }

  public async reInitialize(config: Partial<AuthClientConfig>): Promise<void> {
    return this._authClient.reInitialize(config);
  }

  public async revokeAccessToken(userId?: string): Promise<Response> {
    return this._authClient.revokeAccessToken(userId);
  }

  public static didSignOutFail(afterSignOutUrl: string): boolean {
    return LegacyThunderIDNodeClient.didSignOutFail(afterSignOutUrl);
  }

  public static isSignOutSuccessful(afterSignOutUrl: string): boolean {
    return LegacyThunderIDNodeClient.isSignOutSuccessful(afterSignOutUrl);
  }

  public static protectRoute(
    callback: UnauthenticatedCallback,
  ): (req: express.Request, res: express.Response, next: express.nextFunction) => Promise<void> {
    if (!this._instance) {
      throw new ThunderIDAuthException(
        'EXPRESS-CLIENT-PR-NF01',
        'ThunderIDExpressClient is not instantiated',
        'Create an instance of ThunderIDExpressClient before using calling this method.',
      );
    }

    return protectRoute(this._instance, callback);
  }

  public static thunderIDExpressAuth(
    onSignIn: (response: TokenResponse) => void,
    onSignOut: () => void,
    onError: (exception: ThunderIDAuthException) => void,
  ): any {
    if (!this._instance) {
      throw new ThunderIDAuthException(
        'EXPRESS-CLIENT-AEA-NF01',
        'ThunderIDExpressClient is not instantiated',
        'Create an instance of ThunderIDExpressClient before using calling this method.',
      );
    }

    return thunderIDExpressAuth(this._instance, ThunderIDExpressClient._clientConfig, onSignIn, onSignOut, onError);
  }

  public async getStorageManager() {
    return this._authClient.getStorageManager();
  }
}
