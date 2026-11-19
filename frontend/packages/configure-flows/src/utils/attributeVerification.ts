// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {FlowNodeInput} from '../models/responses';

/**
 * The AttributeVerificationExecutor input that names what a prepare step verifies.
 */
export const ATTRIBUTE_VERIFICATION_INPUT = 'attribute';

/**
 * The value of {@link ATTRIBUTE_VERIFICATION_INPUT} that verifies an agent's external owner instead of an
 * attribute of a user. Only an administrator-initiated flow may use it, since only that flow type binds the
 * subject its caller names.
 */
export const AGENT_OWNER_ATTRIBUTE = 'owner';

/**
 * What a prepare step verifies: an agent's owner, a user attribute the step names (which the request may
 * replace unless the input is constant), or the attribute the request that starts the flow names.
 */
export type AttributeVerificationTarget =
  | {kind: 'agentOwner'}
  | {kind: 'userAttribute'; attribute: string}
  | {kind: 'requested'};

/**
 * Reads what a prepare step verifies from its executor inputs.
 *
 * @param inputs - The step's executor inputs.
 * @returns The verification target.
 */
export const getAttributeVerificationTarget = (inputs: FlowNodeInput[]): AttributeVerificationTarget => {
  const attribute = inputs.find((input) => input.identifier === ATTRIBUTE_VERIFICATION_INPUT)?.defaultValue ?? '';

  if (attribute === AGENT_OWNER_ATTRIBUTE) {
    return {kind: 'agentOwner'};
  }

  return attribute ? {kind: 'userAttribute', attribute} : {kind: 'requested'};
};

/**
 * Returns the executor inputs that make a prepare step verify the given target, keeping every other input.
 * An agent's owner is a constant input, so the request cannot turn the step to a user's attribute. A user
 * attribute keeps the input's own settings, so a constant one stays constant; one replacing the agent's owner
 * starts as a default the request can replace, as the shipped user verification flows do. The requested
 * target drops the input, leaving the executor's own required input for the request to fill.
 *
 * @param inputs - The step's executor inputs.
 * @param target - What the step should verify.
 * @returns The new executor inputs.
 */
export const withAttributeVerificationTarget = (
  inputs: FlowNodeInput[],
  target: AttributeVerificationTarget,
): FlowNodeInput[] => {
  const current = inputs.find((input) => input.identifier === ATTRIBUTE_VERIFICATION_INPUT);
  const others = inputs.filter((input) => input !== current);

  if (target.kind === 'requested') {
    return others;
  }

  if (target.kind === 'agentOwner') {
    return [
      ...others,
      {
        identifier: ATTRIBUTE_VERIFICATION_INPUT,
        type: 'TEXT_INPUT',
        required: true,
        defaultValue: AGENT_OWNER_ATTRIBUTE,
        constant: true,
      },
    ];
  }

  const base: FlowNodeInput =
    current && current.defaultValue !== AGENT_OWNER_ATTRIBUTE
      ? current
      : {identifier: ATTRIBUTE_VERIFICATION_INPUT, type: 'TEXT_INPUT', required: true};

  return [...others, {...base, defaultValue: target.attribute}];
};
