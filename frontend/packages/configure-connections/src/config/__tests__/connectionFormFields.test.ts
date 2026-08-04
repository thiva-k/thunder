// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {describe, expect, it} from 'vitest';
import {ConnectionTypes} from '../../models/connection';
import {fieldsForMode} from '../connectionFormFields';

function fieldNames(type: (typeof ConnectionTypes)[keyof typeof ConnectionTypes], mode: 'create' | 'edit'): string[] {
  return fieldsForMode(type, mode).map((field) => field.name);
}

describe('fieldsForMode', () => {
  it('hides the redirect URI and scopes on create for Google/GitHub, showing them on edit', () => {
    expect(fieldNames(ConnectionTypes.GOOGLE, 'create')).toEqual(['name', 'clientId', 'clientSecret']);
    expect(fieldNames(ConnectionTypes.GOOGLE, 'edit')).toEqual([
      'name',
      'clientId',
      'clientSecret',
      'redirectUri',
      'scopes',
      'prompt',
    ]);
  });

  it('shows only the required fields for OIDC on create, all fields on edit', () => {
    expect(fieldNames(ConnectionTypes.OIDC, 'create')).toEqual([
      'name',
      'clientId',
      'clientSecret',
      'authorizationEndpoint',
      'tokenEndpoint',
    ]);
    expect(fieldNames(ConnectionTypes.OIDC, 'edit')).toEqual([
      'name',
      'clientId',
      'clientSecret',
      'authorizationEndpoint',
      'tokenEndpoint',
      'issuer',
      'userInfoEndpoint',
      'jwksEndpoint',
      'logoutEndpoint',
      'redirectUri',
      'scopes',
      'prompt',
      'tokenExchangeEnabled',
      'trustedTokenAudience',
    ]);
  });

  it('shows only the required fields for OAuth 2.0 on create, all fields on edit', () => {
    expect(fieldNames(ConnectionTypes.OAUTH, 'create')).toEqual([
      'name',
      'clientId',
      'clientSecret',
      'authorizationEndpoint',
      'tokenEndpoint',
      'userInfoEndpoint',
    ]);
    expect(fieldNames(ConnectionTypes.OAUTH, 'edit')).toEqual([
      'name',
      'clientId',
      'clientSecret',
      'authorizationEndpoint',
      'tokenEndpoint',
      'userInfoEndpoint',
      'logoutEndpoint',
      'redirectUri',
      'scopes',
      'prompt',
    ]);
  });

  it('does not hide any SMS vendor fields on create', () => {
    expect(fieldNames(ConnectionTypes.TWILIO, 'create')).toEqual(fieldNames(ConnectionTypes.TWILIO, 'edit'));
    expect(fieldNames(ConnectionTypes.VONAGE, 'create')).toEqual(fieldNames(ConnectionTypes.VONAGE, 'edit'));
    expect(fieldNames(ConnectionTypes.SMS_GATEWAY, 'create')).toEqual([
      'name',
      'url',
      'httpMethod',
      'contentType',
      'httpHeaders',
    ]);
    expect(fieldNames(ConnectionTypes.SMS_GATEWAY, 'create')).toEqual(fieldNames(ConnectionTypes.SMS_GATEWAY, 'edit'));
  });

  it('offers the SMS gateway method and content type as selects with the API-accepted values', () => {
    const fields = fieldsForMode(ConnectionTypes.SMS_GATEWAY, 'create');
    const httpMethod = fields.find((field) => field.name === 'httpMethod');
    const contentType = fields.find((field) => field.name === 'contentType');

    expect(httpMethod).toMatchObject({kind: 'select', required: true, defaultValue: 'POST'});
    expect(httpMethod?.options?.map((option) => option.value)).toEqual(['POST', 'GET']);
    expect(contentType).toMatchObject({kind: 'select', required: true, defaultValue: 'JSON'});
    expect(contentType?.options?.map((option) => option.value)).toEqual(['JSON', 'FORM']);
  });
});
