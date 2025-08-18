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

const base64UrlDecode = (base64UrlString: string): string => {
    // Convert Base64URL → Base64
    let base64 = base64UrlString.replace(/-/g, '+').replace(/_/g, '/');
  
    // Pad with `=` if necessary
    while (base64.length % 4 !== 0) {
      base64 += '=';
    }
  
    return atob(base64);
}

/**
 * Decodes a JWT token string into its header, payload, and signature components.
 * 
 * @param token JWT token string to decode.
 * @returns An object containing the decoded header, payload, and signature.
 */
export const decodeJwt = (token: string) => {
    try {
        const [header, payload, signature] = token.split('.');

        const decodedHeader = JSON.parse(base64UrlDecode(header));
        const decodedPayload = JSON.parse(base64UrlDecode(payload));
        return {
            header: decodedHeader,
            payload: decodedPayload,
            // Signature is not decoded as it's not base64-encoded JSON.
            signature,
        };
    } catch (error) {
        console.error('Failed to decode token:', error);
        return null;
    }
};
