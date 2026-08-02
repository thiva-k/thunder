// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

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
