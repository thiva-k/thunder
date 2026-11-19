// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {describe, it, expect} from 'vitest';
import type {FlowNodeInput} from '../../models/responses';
import {getAttributeVerificationTarget, withAttributeVerificationTarget} from '../attributeVerification';

const ownerInput: FlowNodeInput = {
  identifier: 'attribute',
  type: 'TEXT_INPUT',
  required: true,
  defaultValue: 'owner',
  constant: true,
};
const emailInput: FlowNodeInput = {identifier: 'attribute', type: 'TEXT_INPUT', required: true, defaultValue: 'email'};
const otherInput: FlowNodeInput = {identifier: 'other', type: 'TEXT_INPUT', required: false};

describe('getAttributeVerificationTarget', () => {
  it.each([
    {name: 'the agent owner', inputs: [ownerInput], expected: {kind: 'agentOwner'}},
    {name: 'a user attribute', inputs: [otherInput, emailInput], expected: {kind: 'userAttribute', attribute: 'email'}},
    {name: 'the requested attribute with no input', inputs: [], expected: {kind: 'requested'}},
    {
      name: 'the requested attribute with an input that has no default',
      inputs: [{identifier: 'attribute', type: 'TEXT_INPUT', required: true}],
      expected: {kind: 'requested'},
    },
  ])('reads $name', ({inputs, expected}) => {
    expect(getAttributeVerificationTarget(inputs)).toEqual(expected);
  });
});

describe('withAttributeVerificationTarget', () => {
  it('fixes the agent owner as a constant input, keeping other inputs', () => {
    expect(withAttributeVerificationTarget([otherInput, emailInput], {kind: 'agentOwner'})).toEqual([
      otherInput,
      ownerInput,
    ]);
  });

  it('sets a user attribute as a default the request can replace when it replaces the agent owner', () => {
    expect(withAttributeVerificationTarget([ownerInput], {kind: 'userAttribute', attribute: 'mobileNumber'})).toEqual([
      {identifier: 'attribute', type: 'TEXT_INPUT', required: true, defaultValue: 'mobileNumber'},
    ]);
  });

  it("keeps a user attribute input's own settings, so a constant one stays constant", () => {
    const constantEmail: FlowNodeInput = {...emailInput, constant: true};

    expect(withAttributeVerificationTarget([constantEmail], {kind: 'userAttribute', attribute: 'workEmail'})).toEqual([
      {...constantEmail, defaultValue: 'workEmail'},
    ]);
  });

  it('drops the input for the requested attribute, as the shipped self verification flow has none', () => {
    expect(withAttributeVerificationTarget([emailInput, otherInput], {kind: 'requested'})).toEqual([otherInput]);
    expect(withAttributeVerificationTarget([ownerInput], {kind: 'requested'})).toEqual([]);
  });
});
