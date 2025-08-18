/*
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
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

const response = await fetch('/runtime.json');
const runtimeConfig = await response.json();

const config = {
    applicationID: runtimeConfig.applicationID || import.meta.env.VITE_REACT_APP_AUTH_APP_ID,
    applicationsEndpoint: runtimeConfig.applicationsEndpoint || import.meta.env.VITE_REACT_APPLICATIONS_ENDPOINT,
    flowEndpoint: runtimeConfig.flowEndpoint || import.meta.env.VITE_REACT_APP_SERVER_FLOW_ENDPOINT,
    authorizationEndpoint: runtimeConfig.authorizationEndpoint || import.meta.env.VITE_REACT_APP_SERVER_AUTHORIZATION_ENDPOINT,
    tokenEndpoint: runtimeConfig.tokenEndpoint || import.meta.env.VITE_REACT_APP_SERVER_TOKEN_ENDPOINT,
    clientId: import.meta.env.VITE_REACT_APP_CLIENT_ID,
    clientSecret: import.meta.env.VITE_REACT_APP_CLIENT_SECRET,
    redirectUri: runtimeConfig.redirectUri || import.meta.env.VITE_REACT_APP_REDIRECT_URI,
    scope: import.meta.env.VITE_REACT_APP_SCOPE
};

export default config;
