# WSO2 Thunder ⚡
### Identity Management Suite

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub last commit](https://img.shields.io/github/last-commit/asgardeo/thunder.svg)](https://github.com/asgardeo/thunder/commits/main)
[![GitHub issues](https://img.shields.io/github/issues/asgardeo/thunder.svg)](https://github.com/asgardeo/thunder/issues)

Thunder is a modern, open-source identity management service designed for teams building secure, customizable authentication experiences across applications, services, and AI agents. It enables developers to design and orchestrate login, registration, and recovery flows using a flexible identity flow designer.

Designed for extensibility, scalability, and seamless containerized deployment, Thunder integrates naturally with microservices and DevOps environments—serving as the core identity layer for your cloud platform.

---

## 🚀 Features

- ✅ **Standards-Based**
  - OAuth 2/ OpenID Connect (OIDC): Client Credentials, Authorization Code, Refresh Token
- 🔗 **Login Options:**
  - Basic Authentication (Username/Password)
  - Social Logins: Google, Github
  - SMS OTP
- 📝 **Registration Options:**
  - Username/Password
  - Social Registration: Google, Github
  - SMS OTP
- 🌐 **RESTful APIs:**
  - App Native Login/Registration
  - User Management
  - Application Management
  - Identity Provider Management
  - Message Notification Sender Management

---

## ⚡ Quickstart

### Download and Run WSO2 Thunder

You can run WSO2 Thunder either by downloading the release artifact or using the official Docker image.

#### Option 1: Run from Release Artifact

Follow these steps to download the latest release of WSO2 Thunder and run it locally.

1. **Download the distribution from the latest release**

    Download `thunder_<os>_<arch>-<version>.zip` from the [latest release](https://github.com/asgardeo/thunder/releases/latest) for your operating system and architecture.

    For example, if you are using a MacOS machine with a Apple Silicon (ARM64) processor, you would download `thunder_macos_arm64-<version>.zip`.

2. **Unzip the product**

    Unzip the downloaded file using the following command:

    ```bash
    unzip thunder-<os>_<arch>-<version>.zip
    ```

    Navigate to the unzipped directory:

    ```bash
    cd thunder-<os>_<arch>-<version>/
    ```

3. **Start the product**

    Start the product using the following command:

    If you are using a Linux or macOS machine:

    ```bash
    bash start.sh
    ```

    If you are using a Windows machine:

    ```bash
    start.bat
    ```

#### Option 2: Run with Docker

Follow these steps to run WSO2 Thunder using Docker.

1. **Pull the Docker image**

    ```bash
    docker pull ghcr.io/asgardeo/thunder:latest
    ```

2. **Run the container**

    ```bash
    docker run --rm \
      -p 8090:8090 \
      ghcr.io/asgardeo/thunder:latest
    ```

    Optionally if you want to modify the server configurations, you can mount a custom `deployment.yaml` file. Create a `deployment.yaml` file in your working directory similar to the [deployment.yaml](https://github.com/asgardeo/thunder/blob/main/backend/cmd/server/repository/conf/deployment.yaml), and mount it as below:

    ```bash
    docker run --rm \
      -p 8090:8090 \
      -v $(pwd)/deployment.yaml:/opt/thunder/repository/conf/deployment.yaml \
      ghcr.io/asgardeo/thunder:latest
    ```

    Optionally if you want to use custom certificates, you can mount them as follows:

    ```bash
    docker run --rm \
      -p 8090:8090 \
      -v $(pwd)/deployment.yaml:/opt/thunder/repository/conf/deployment.yaml \
      -v $(pwd)/certs/server.cert:/opt/thunder/repository/resources/security/server.cert \
      -v $(pwd)/certs/server.key:/opt/thunder/repository/resources/security/server.key \
      ghcr.io/asgardeo/thunder:latest
    ```

### Download and Run the Sample App

To quickly get started with WSO2 Thunder, you can use the sample app provided with the product. Follow these steps to download and run the sample app.

#### Step 1: Download the sample app

Download `thunder-sample-app-<version>.zip` from the [latest release](https://github.com/asgardeo/thunder/releases/latest).

#### Step 2: Unzip the sample app and install dependencies

```bash
unzip thunder-sample-app-<version>.zip
cd thunder-sample-app-<version>/
npm install
```

#### (Optional) Step 3: Configure the sample app

Open the `runtime.json` file in the thunder-sample-app-<version>/dist directory and update the configurations as per your setup. The default configurations should work for most cases, but you can customize the following properties:

- `applicationID`: The ID of the application you want to use for authentication. By default, it is set to `550e8400-e29b-41d4-a716-446655440000`.
- `flowEndpoint`: The endpoint for the flow execution API. By default, it is set to `https://localhost:8090/flow/execute`.

#### Step 4: Start the sample app

```bash
npm start
```

Open your browser and navigate to `https://localhost:3000` to see the sample app in action.

### Try Out the Product

#### 1️⃣ Create a User

Create a user in the system to tryout the authentication flows. You can use the following cURL command to create a user with the required attributes.

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/users \
  -d '{
      "organizationUnit": "456e8400-e29b-41d4-a716-446655440001",
      "type": "superhuman",
      "attributes": {
          "username": "thor",
          "password": "thor@123",
          "email": "thor@thunder.sky",
          "firstName": "Thor",
          "lastName": "Odinson",
          "age": 1534,
          "abilities": [
              "strength",
              "speed",
              "healing"
          ],
          "address": {
              "city": "Asgard",
              "zip": "00100"
          }
      }
  }'
  ```

#### 2️⃣ Try Out Client Credentials Flow

```bash
curl -k -X POST https://localhost:8090/oauth2/token \
  -d 'grant_type=client_credentials' \
  -u 'client123:secret123'
```

#### 3️⃣ Try Out Authorization Code Flow

Authorization code flow requires you to setup a gate client to handle the login and error redirection. Add the following configurations to the `deployment.yaml` file to configure your own gate client.

  ```yaml
  gate_client:
    hostname: "localhost"
    port: 9090
    scheme: "https"
    login_path: "/login"
    error_path: "/error"
  ```

- Open the following URL in your browser:

  ```bash
  https://localhost:8090/oauth2/authorize?response_type=code&client_id=client123&redirect_uri=https://localhost:3000&scope=openid&state=state_1
  ```

- Enter the credentials of the user you created in the first step.

- After successful authentication, you will be redirected to the redirect URI with the authorization code and state.

  ```bash
  https://localhost:3000/?code=<code>&state=state_1
  ```

- Copy the authorization code and exchange it for an access token using the following cURL command:

  ```bash
  curl -k -X POST 'https://localhost:8090/oauth2/token' \
  -u 'client123:secret123' \
  -d 'grant_type=authorization_code' \
  -d 'redirect_uri=https://localhost:3000' \
  -d 'code=<code>'
  ```

  - **Client ID:** `client123`
  - **Client Secret:** `secret123`

##### Refresh an Access Token

If the refresh grant is enabled for the application, a refresh token will be issued along with the access token. The refresh token can be used to obtain a new access token without requiring the user to log in again. Use the following cURL command to refresh an access token:

  ```bash
  curl -k -X POST 'https://localhost:8090/oauth2/token' \
  -u 'client123:secret123' \
  -d 'grant_type=refresh_token' \
  -d 'refresh_token=<refresh_token>'
  ```

#### 4️⃣ Try Username and Password Login

Open the sample app in your browser and enter the username and password you created in the first step. If the login is successful, you will be redirected to the home page of the sample app with the access token.

#### 5️⃣ Try Google Login

- Create an OAuth application in your Google account following the instructions given in the [Google documentation](https://developers.google.com/identity/protocols/oauth2/web-server#creatingcred).
  - Configure the Authorized origin and Redirect URI as per your application.
  - Copy the **Client ID** and **Client Secret**.

- Update the system created Google IDP by invoking the IDP management API with the following cURL command. Make sure to replace `<client_id>`, `<client_secret>`, and `<app_callback_url>` with the values you copied from your Google OAuth application.

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers/550e8400-e29b-41d4-a716-446655440002 \
  -d '{
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Google",
      "description": "Login with Google",
      "properties": [
        {
            "name": "client_id",
            "value": "<client_id>",
            "is_secret": false
        },
        {
            "name": "client_secret",
            "value": "<client_secret>",
            "is_secret": true
        },
        {
            "name": "redirect_uri",
            "value": "<app_callback_url>",
            "is_secret": false
        },
        {
            "name": "scopes",
            "value": "openid,email,profile",
            "is_secret": false
        }
    ]
  }'
  ```

  > Note: Refer [Identity Provider Configuration Guide](/docs/content/configure-identity-providers.md) for more details on configuring identity providers.

- Update the system default application to use the Google login template by invoking the application management API with the following cURL command.

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications/550e8400-e29b-41d4-a716-446655440000 \
  --data '{
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Test SPA",
      "description": "Initial testing App",
      "client_id": "client123",
      "client_secret": "***",
      "redirect_uris": [
          "https://localhost:3000"
      ],
      "grant_types": [
          "client_credentials",
          "authorization_code"
      ],
      "auth_flow_graph_id": "auth_flow_config_google"
  }'
  ```

  > Note: Refer [Authentication Flow Customization Guide](/docs/content/customize-auth-flow.md) for more details on customizing authentication flows.

- Open the sample app in your browser and click on the "Continue with Google" button. You will be redirected to the Google login page. Enter your Google credentials and authorize the application.

- If the login is successful, you will be redirected to the home page of the sample app with the access token.

#### 6️⃣ Try GitHub Login

- Create an OAuth application in your Github account following the instructions given in the [Github documentation](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app).
  - Configure home page and callback URLs as per your application.
  - Copy the **Client ID** and **Client Secret**.

- Update the system created github IDP by invoking the IDP management API with the following cURL command. Make sure to replace `<client_id>`, `<client_secret>`, and `<app_callback_url>` with the values you copied from your GitHub OAuth application.

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers/550e8400-e29b-41d4-a716-446655440001 \
  -d '{
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Github",
      "description": "Login with Github",
      "properties": [
          {
              "name": "client_id",
              "value": "<client_id>",
              "is_secret": false
          },
          {
              "name": "client_secret",
              "value": "<client_secret>",
              "is_secret": true
          },
          {
              "name": "redirect_uri",
              "value": "<app_callback_url>",
              "is_secret": false
          },
          {
              "name": "scopes",
              "value": "user:email,read:user",
              "is_secret": false
          }
      ]
  }'
  ```

  > Note: Refer [Identity Provider Configuration Guide](/docs/content/configure-identity-providers.md) for more details on configuring identity providers.

- Update the system default application to use the Github login template by invoking the application management API with the following cURL command.

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications/550e8400-e29b-41d4-a716-446655440000 \
  --data '{
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Test SPA",
      "description": "Initial testing App",
      "client_id": "client123",
      "client_secret": "***",
      "redirect_uris": [
          "https://localhost:3000"
      ],
      "grant_types": [
          "client_credentials",
          "authorization_code"
      ],
      "auth_flow_graph_id": "auth_flow_config_github"
  }'
  ```

  > Note: Refer [Authentication Flow Customization Guide](/docs/content/customize-auth-flow.md) for more details on customizing authentication flows.

- Open the sample app in your browser and click on the "Continue with GitHub" button. You will be redirected to the GitHub login page. Enter your GitHub credentials and authorize the application.

- If the login is successful, you will be redirected to the home page of the sample app with the access token.

#### 7️⃣ Try SMS OTP Login

SMS One-Time Password (OTP) authentication allows users to authenticate using a one-time code sent to their mobile number. Prior to using SMS OTP, you need to configure a message provider to send SMS messages. Follow the steps below to set up SMS OTP authentication.

##### Step 1: Configure a Message Provider

Configure a message provider to send SMS messages. You can use services like Twilio, Vonage, or a custom service of your choice.

```bash
curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/notification-senders/message \
-d '{
  "name": "Custom SMS Sender",
  "description": "Sender for sending SMS messages",
  "provider": "custom",
  "properties": [
    {
      "name": "url",
      "value": "<custom_sms_provider_url>"
    },
    {
      "name": "http_method",
      "value": "POST"
    },
    {
      "name": "content_type",
      "value": "JSON"
    }
  ]
}'
```

> Note: Refer [Message Provider Configuration Guide](/docs/content/configure-message-providers.md) for more details on configuring message providers.

##### Step 2: Update SMS OTP Flow Configuration

Update the `senderName` property in the `auth_flow_config_sms` flow definition to use your configured message sender.

##### Step 3: Configure Application to Use SMS OTP Flow

Update the system application to use the SMS OTP authentication flow template:

```bash
curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications/550e8400-e29b-41d4-a716-446655440000 \
--data '{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Test SPA",
    "description": "Initial testing App",
    "client_id": "client123",
    "client_secret": "***",
    "redirect_uris": [
        "https://localhost:3000"
    ],
    "grant_types": [
        "client_credentials",
        "authorization_code"
    ],
    "auth_flow_graph_id": "auth_flow_config_sms"
}'
```

  > Note: Refer [Authentication Flow Customization Guide](/docs/content/customize-auth-flow.md) for more details on customizing authentication flows.

##### Step 4: Create a User with Mobile Number

Create a user with a `mobileNumber` attribute to receive SMS OTP:

```bash
curl -kL -H 'Content-Type: application/json' https://localhost:8090/users \
-d '{
    "organizationUnit": "456e8400-e29b-41d4-a716-446655440001",
    "type": "superhuman",
    "attributes": {
        "username": "thor",
        "password": "thor@123",
        "email": "thor@thunder.sky",
        "firstName": "Thor",
        "lastName": "Odinson",
        "age": 1534,
        "abilities": [
            "strength",
            "speed",
            "healing"
        ],
        "address": {
            "city": "Asgard",
            "zip": "00100"
        },
        "mobileNumber": "+94xxxxxxxxx"
    }
}'
```

##### Step 5: Try Out SMS OTP Authentication

- Initiate the authentication flow with your application ID:

  ```bash
  curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "applicationId": "550e8400-e29b-41d4-a716-446655440000",
      "flowType": "AUTHENTICATION"
  }'
  ```

  You'll receive a response with a flow ID and username input request.

- Provide the username to continue the flow:

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "flowId": "<flow_id>",
      "inputs": {
          "username": "thor"
      }
  }'
  ```

  An OTP will be sent to the user's mobile number.

- Complete authentication by providing the OTP:

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "flowId": "<flow_id>",
      "inputs": {
          "otp": "696546"
      }
  }'
  ```

- If the OTP is valid, you will receive a response with the auth assertion.

#### 8️⃣ Try Self Registration

Self-registration allows users to create their own accounts. Open the sample app in your browser and click on the "Sign up" button. You will be presented with the registration options based on the configured registration flow for the application.

When you configure an authentication flow for an application, Thunder automatically assigns the equivalent registration flow. For example, if you set `auth_flow_graph_id` to `auth_flow_config_basic`, the system will automatically use `registration_flow_config_basic` for registration.

> Note: Refer [Registration Flow Customization Guide](/docs/content/customize-registration-flow.md) for more details on configuring registration flows.

---

<details>
<summary><h2>🔍 Feature Walkthrough</h2></summary>

</br><p>This section provides a detailed walkthrough of the authentication flows supported by WSO2 Thunder. You can try out these flows using the sample app provided with the product or by using the cURL commands provided below.</p>

<details>
<summary><h3>🔐 App Native Authentication</h3></summary>

</br><p>WSO2 Thunder supports app native authentication flows, allowing users to execute login flows via REST APIs. This is particularly useful for mobile and desktop applications that require a native login experience.</p>

  > Note: Refer [Authentication Flow Customization Guide](/docs/content/customize-auth-flow.md) for more details on customizing authentication flows.

<details>
<summary><h4>1️⃣ Login with Username and Password</h4></summary>

- Create a user in the system if you haven't already. You can use the following cURL command to create a user with the required attributes.

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/users \
  -d '{
      "organizationUnit": "456e8400-e29b-41d4-a716-446655440001",
      "type": "superhuman",
      "attributes": {
          "username": "thor",
          "password": "thor@123",
          "email": "thor@thunder.sky",
          "firstName": "Thor",
          "lastName": "Odinson",
          "age": 1534,
          "abilities": [
              "strength",
              "speed",
              "healing"
          ],
          "address": {
              "city": "Asgard",
              "zip": "00100"
          }
      }
  }'
  ```

- Create an application or update the existing system default application to use the basic auth login template. You can use the following cURL command to update the default application.

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications/550e8400-e29b-41d4-a716-446655440000 \
  --data '{
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Test SPA",
      "description": "Initial testing App",
      "client_id": "client123",
      "client_secret": "***",
      "redirect_uris": [
          "https://localhost:3000"
      ],
      "grant_types": [
          "client_credentials",
          "authorization_code"
      ],
      "auth_flow_graph_id": "auth_flow_config_basic"
  }'
  ```

- Start login flow for the application with the following cURL command:

  ```bash
  curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "applicationId": "<application_id>",
      "flowType": "AUTHENTICATION"
  }'
  ```

  You'll receive a response similar to the following:

  ```json
  {
      "flowId": "db93a19e-c23f-4cfc-a45f-0e0bc157f6d5",
      "flowStatus": "PROMPT_ONLY",
      "type": "VIEW",
      "data": {
        "inputs": [
          {
              "name": "username",
              "type": "string",
              "required": true
          },
          {
              "name": "password",
              "type": "string",
              "required": true
          }
        ]
      }
  }
  ```

- Make the second cURL request to complete the login flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response. Also, replace the `username` and `password` with the credentials of the user you created in the first step.

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "flowId": "<flow_id>",
      "inputs": {
          "username": "thor",
          "password": "thor@123"
      }
  }'
  ```

- If the login is successful, you will receive a response with the auth assertion.

</details>
<details>
<summary><h4>2️⃣ Login with Google</h4></summary>

- Create an OAuth application in your Google account following the instructions given in the [Google documentation](https://developers.google.com/identity/protocols/oauth2/web-server#creatingcred).
  - Configure the Authorized origin and Redirect URI as per your application.
  - Copy the **Client ID** and **Client Secret**.

- Update the system created Google IDP by invoking the IDP management API with the following cURL command. Make sure to replace `<client_id>`, `<client_secret>`, and `<app_callback_url>` with the values you copied from your Google OAuth application.

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers/550e8400-e29b-41d4-a716-446655440002 \
  -d '{
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Google",
      "description": "Login with Google",
      "properties": [
          {
              "name": "client_id",
              "value": "<client_id>",
              "is_secret": false
          },
          {
              "name": "client_secret",
              "value": "<client_secret>",
              "is_secret": true
          },
          {
              "name": "redirect_uri",
              "value": "<app_callback_url>",
              "is_secret": false
          },
          {
              "name": "scopes",
              "value": "openid,email,profile",
              "is_secret": false
          }
      ]
  }'
  ```

  > Note: Refer [Identity Provider Configuration Guide](/docs/content/configure-identity-providers.md) for more details on configuring identity providers.

- Create an application or update the existing system default application to use the Google login template. You can use the following cURL command to update the default application.

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications/550e8400-e29b-41d4-a716-446655440000 \
  --data '{
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Test SPA",
      "description": "Initial testing App",
      "client_id": "client123",
      "client_secret": "***",
      "redirect_uris": [
          "https://localhost:3000"
      ],
      "grant_types": [
          "client_credentials",
          "authorization_code"
      ],
      "auth_flow_graph_id": "auth_flow_config_google"
  }'
  ```

- Start login flow for the application with the following cURL command:

  ```bash
  curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "applicationId": "<application_id>",
      "flowType": "AUTHENTICATION"
  }'
  ```

  You'll receive a response similar to the following:

  ```json
  {
      "flowId": "80d57e64-8082-4096-bb0e-22b2187f8265",
      "flowStatus": "INCOMPLETE",
      "type": "REDIRECTION",
      "data": {
          "redirectURL": "<google_auth_redirect_url>",
          "inputs": [
            {
                "name": "code",
                "type": "string",
                "required": true
            },
            {
                "name": "nonce",
                "type": "string",
                "required": false
            }
          ],
          "additionalData": {
            "idpName": "Google"
          }
      }
  }
  ```

- Open the `redirect_url` in your browser. You will be redirected to the Google login page. Enter your Google credentials and authorize the application.

- After successful authentication, you will be redirected to the redirect URI with the authorization code, state and other parameters.

  ```bash
  https://localhost:3000/?code=<code>&state=80d57e64-8082-4096-bb0e-22b2187f8265
  ```

- Copy the authorization code and make the second cURL request to complete the login flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response.

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "flowId": "<flow_id>",
      "inputs": {
          "code": "<code>"
      }
  }'
  ```

- If the login is successful, you will receive a response with the auth assertion.

</details>
<details>
<summary><h4>3️⃣ Login with GitHub</h4></summary>

- Create an OAuth application in your Github account following the instructions given in the [Github documentation](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app).
  - Configure home page and callback URLs as per your application.
  - Copy the **Client ID** and **Client Secret**.

- Update the system created github IDP by invoking the IDP management API with the following cURL command. Make sure to replace `<client_id>`, `<client_secret>`, and `<app_callback_url>` with the values you copied from your GitHub OAuth application.

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/identity-providers/550e8400-e29b-41d4-a716-446655440001 \
  -d '{
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Github",
      "description": "Login with Github",
      "properties": [
          {
              "name": "client_id",
              "value": "<client_id>",
              "is_secret": false
          },
          {
              "name": "client_secret",
              "value": "<client_secret>",
              "is_secret": true
          },
          {
              "name": "redirect_uri",
              "value": "<app_callback_url>",
              "is_secret": false
          },
          {
              "name": "scopes",
              "value": "user:email,read:user",
              "is_secret": false
          }
      ]
  }'
  ```

  > Note: Refer [Identity Provider Configuration Guide](/docs/content/configure-identity-providers.md) for more details on configuring identity providers.

- Create an application or update the existing system default application to use the GitHub login template. You can use the following cURL command to update the default application.

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications/550e8400-e29b-41d4-a716-446655440000 \
  --data '{
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Test SPA",
      "description": "Initial testing App",
      "client_id": "client123",
      "client_secret": "***",
      "redirect_uris": [
          "https://localhost:3000"
      ],
      "grant_types": [
          "client_credentials",
          "authorization_code"
      ],
      "auth_flow_graph_id": "auth_flow_config_github"
  }'
  ```

- Start login flow for the application with the following cURL command:

  ```bash
  curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "applicationId": "<application_id>",
      "flowType": "AUTHENTICATION"
  }'
  ```

  You'll receive a response similar to the following:

  ```json
  {
      "flowId": "80d57e64-8082-4096-bb0e-22b2187f8265",
      "flowStatus": "INCOMPLETE",
      "type": "REDIRECTION",
      "data": {
        "redirectURL": "<github_auth_redirect_url>",
        "inputs": [
          {
            "name": "code",
            "type": "string",
            "required": true
          }
        ],
        "additionalData": {
          "idpName": "Github"
        }
      }
  }
  ```

- Open the `redirect_url` in your browser. You will be redirected to the GitHub login page. Enter your GitHub credentials and authorize the application.

- After successful authentication, you will be redirected to the redirect URI with the authorization code and state.

  ```bash
  https://localhost:3000/?code=<code>&state=80d57e64-8082-4096-bb0e-22b2187f8265
  ```

- Copy the authorization code and make the second cURL request to complete the login flow. Make sure to replace `<flow_id>` with the `flowId` received in the previous response.

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "flowId": "<flow_id>",
      "inputs": {
          "code": "<code>"
      }
  }'
  ```

- If the login is successful, you will receive a response with the auth assertion.

</details>
<details>
<summary><h4>4️⃣ Login with SMS OTP</h4></summary>

- Configure a message provider to send SMS messages. You can use services like Twilio, Vonage, or a custom service of your choice.

  ```bash
  curl -kL -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/notification-senders/message \
  -d '{
    "name": "Custom SMS Sender",
    "description": "Sender for sending SMS messages",
    "provider": "custom",
    "properties": [
      {
        "name": "url",
        "value": "<custom_sms_provider_url>"
      },
      {
        "name": "http_method",
        "value": "POST"
      },
      {
        "name": "content_type",
        "value": "JSON"
      }
    ]
  }'
  ```

  > Note: Refer [Message Provider Configuration Guide](/docs/content/configure-message-providers.md) for more details on configuring message providers.

- Update the system application to use the SMS OTP authentication flow:

  ```bash
  curl -kL -X PUT -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications/550e8400-e29b-41d4-a716-446655440000 \
  --data '{
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Test SPA",
      "description": "Initial testing App",
      "client_id": "client123",
      "client_secret": "***",
      "redirect_uris": [
          "https://localhost:3000"
      ],
      "grant_types": [
          "client_credentials",
          "authorization_code"
      ],
      "auth_flow_graph_id": "auth_flow_config_sms"
  }'
  ```

- Create a user with a mobile number attribute to receive SMS OTP:

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/users \
  -d '{
      "organizationUnit": "456e8400-e29b-41d4-a716-446655440001",
      "type": "superhuman",
      "attributes": {
          "username": "thor",
          "password": "thor@123",
          "email": "thor@thunder.sky",
          "firstName": "Thor",
          "lastName": "Odinson",
          "mobileNumber": "+94xxxxxxxxx"
      }
  }'
  ```

- Start the login flow for the application:

  ```bash
  curl -kL -H 'Accept: application/json' -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "applicationId": "<application_id>",
      "flowType": "AUTHENTICATION"
  }'
  ```

  You'll receive a response prompting for username input.

- Provide the username to continue the flow:

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "flowId": "<flow_id>",
      "inputs": {
          "username": "thor"
      }
  }'
  ```

  An OTP will be sent to the user's mobile number.

- Complete authentication by providing the OTP:

  ```bash
  curl -kL -H 'Content-Type: application/json' https://localhost:8090/flow/execute \
  -d '{
      "flowId": "<flow_id>",
      "inputs": {
          "otp": "696546"
      }
  }'
  ```

- If the OTP is valid, you will receive a response with the auth assertion.

</details>
</details>
</details>

---

<details>
<summary><h2>⚡ Build the Product from Source</h2></summary>

### ✅ Prerequisites

- Go 1.24+
- Node.js 20+

---

- Build the product with tests using the following command:

```bash
make all
```

</details>

---

<details>
<summary><h2>🛠️ Development Setup</h2></summary>

### Prerequisites

- Go 1.24+
- Node.js 20+

### Start Thunder in Development Mode

- Clone the repository:

```bash
git clone https://github.com/asgardeo/thunder
cd thunder
```

- Run the following command to start the product in development mode:

```bash
make run
```

- The product will start on `https://localhost:8090`.

### Start the Sample App in Development Mode

- Navigate to the sample app directory:

  ```bash
  cd samples/apps/oauth
  ```

- Create a file `.env` in the path `samples/apps/oauth/` and add below values.

  ```
  VITE_REACT_APP_SERVER_FLOW_ENDPOINT=https://localhost:8090/flow
  VITE_REACT_APPLICATIONS_ENDPOINT=https://localhost:8090/applications
  VITE_REACT_APP_AUTH_APP_ID=550e8400-e29b-41d4-a716-446655440000
  VITE_REACT_APP_REDIRECT_BASED_LOGIN=false
  ```

- Install the dependencies:

  ```bash
  npm install
  ```

- Run the sample app using the following command:

  ```bash
  npm run dev
  ```
  
- Open your browser and navigate to `http://localhost:5173` to see the sample app in action.

### Remote Debugging Setup

Thunder supports remote debugging using Delve debugger, enabling debugging from any IDE that supports the Debug Adapter Protocol (DAP).

#### Install Delve Debugger

Install Delve using Go:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Add Delve to your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Start Thunder in Debug Mode

From the distribution directory:

```bash
./start.sh --debug
```

The debugger will listen on `localhost:2345` by default.

#### Connect from IDE

**VS Code:**
- Use the provided `.vscode/launch.json` configuration
- Press `F5` or go to `Run and Debug`

**GoLand/IntelliJ:**
- Go to `Run → Edit Configurations → + → Go Remote`
- Set Host: `127.0.0.1`, Port: `2345`

**Other IDEs:**
- Configure DAP client to connect to `127.0.0.1:2345`

</details>

---

<details>
<summary><h2>🔧 Advanced Configurations</h2></summary>

<details>
<summary><h3>🗄️ Running with PostgreSQL Database</h3></summary>

#### 🔧 Step 1: Start PostgreSQL

- Create a Docker container for PostgreSQL with `thunderdb` database.

  ```bash
  docker run -d -p 5432:5432 --name postgres \
    -e POSTGRES_USER=asgthunder \
    -e POSTGRES_PASSWORD=asgthunder \
    -e POSTGRES_DB=thunderdb \
    postgres
  ```

- Create the `runtimedb` in the same PostgreSQL container.

  ```bash
  docker exec -it postgres psql -U asgthunder -d thunderdb -c "CREATE DATABASE runtimedb;"
  ```

#### 🗂 Step 2: Initialize the Database

- Populate the `thunderdb` database with the required tables and data.

  ```bash
  docker exec -i postgres psql -U asgthunder -d thunderdb < backend/dbscripts/thunderdb/postgres.sql
  ```

- Populate the `runtimedb` database with the required tables and data.

  ```bash
  docker exec -i postgres psql -U asgthunder -d thunderdb < backend/dbscripts/runtimedb/postgres.sql
  ```

#### 🛠 Step 3: Configure Thunder to Use PostgreSQL

1. Open the `backend/cmd/server/repository/conf/deployment.yaml` file.
2. Update the `database` section to point to the PostgreSQL database:
```yaml
database:
  identity:
    type: "postgres"
    hostname: "localhost"
    port: 5432
    name: "thunderdb"
    username: "asgthunder"
    password: "asgthunder"
    sslmode: "disable"
  runtime:
    type: "postgres"
    hostname: "localhost"
    port: 5432
    name: "runtimedb"
    username: "asgthunder"
    password: "asgthunder"
    sslmode: "disable"
```

#### ▶️ Step 4: Run the Product

   ```bash
   make run
   ```

The product will now use the PostgreSQL database for its operations.

</details>
</details>

---

## License

Licenses this source under the Apache License, Version 2.0 ([LICENSE](LICENSE)), You may not use this file except in compliance with the License.

---------------------------------------------------------------------------
(c) Copyright 2025 WSO2 LLC.
