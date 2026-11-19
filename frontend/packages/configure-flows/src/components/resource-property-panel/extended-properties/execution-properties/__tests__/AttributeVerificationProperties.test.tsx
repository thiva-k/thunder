// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type {ReactElement} from 'react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import FlowTypeContext from '../../../../../context/FlowTypeContext';
import {FlowType} from '../../../../../models/flows';
import type {Resource} from '../../../../../models/resources';
import type {FlowNodeInput} from '../../../../../models/responses';
import AttributeVerificationProperties from '../AttributeVerificationProperties';

vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({
    t: (_key: string, fallback?: string) => fallback ?? _key,
  }),
}));

vi.mock('@thunderid/configure-user-types', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@thunderid/configure-user-types')>()),
  useGetUserTypeAttributes: () => ({
    attributes: [
      {attribute: 'email', credential: false, unique: true, userTypes: ['Person']},
      {attribute: 'mobileNumber', credential: false, unique: false, userTypes: ['Person']},
      {attribute: 'password', credential: true, unique: false, userTypes: ['Person']},
    ],
    isLoading: false,
  }),
}));

const ownerInput: FlowNodeInput = {
  identifier: 'attribute',
  type: 'TEXT_INPUT',
  required: true,
  defaultValue: 'owner',
  constant: true,
};

const prepareStep = (inputs?: FlowNodeInput[]): Resource =>
  ({
    id: 'prepare_verification',
    data: {action: {executor: {name: 'AttributeVerificationExecutor', mode: 'prepare', inputs}}},
  }) as unknown as Resource;

const renderIn = (ui: ReactElement, flowType?: FlowType) =>
  render(<FlowTypeContext.Provider value={flowType}>{ui}</FlowTypeContext.Provider>);

describe('AttributeVerificationProperties', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    mockOnChange.mockClear();
  });

  it("shows a user's attribute and its default for the shipped user verification flow", () => {
    renderIn(
      <AttributeVerificationProperties
        resource={prepareStep([{identifier: 'attribute', type: 'TEXT_INPUT', required: true, defaultValue: 'email'}])}
        onChange={mockOnChange}
      />,
      FlowType.ADMIN_INITIATED,
    );

    expect(screen.getByRole('radio', {name: 'User'})).toBeChecked();
    expect(screen.getByRole('combobox', {name: 'Attribute to verify'})).toHaveValue('email');
  });

  it('leaves the attribute empty when the request names it, as in the shipped self flow', () => {
    renderIn(<AttributeVerificationProperties resource={prepareStep()} onChange={mockOnChange} />, FlowType.SELF);

    expect(screen.getByRole('radio', {name: 'User'})).toBeChecked();
    expect(screen.getByRole('combobox', {name: 'Attribute to verify'})).toHaveValue('');
  });

  it("shows an agent's owner for the shipped owner verification flow, with no attribute to pick", () => {
    renderIn(
      <AttributeVerificationProperties resource={prepareStep([ownerInput])} onChange={mockOnChange} />,
      FlowType.ADMIN_INITIATED,
    );

    expect(screen.getByRole('radio', {name: 'Agent owner'})).toBeChecked();
    expect(screen.queryByRole('combobox', {name: 'Attribute to verify'})).not.toBeInTheDocument();
    expect(screen.getByText(/The owner needs no account/)).toBeInTheDocument();
    expect(screen.queryByTestId('attribute-verification-owner-flow-type')).not.toBeInTheDocument();
  });

  it.each([
    {name: 'the shipped self flow', inputs: undefined, flowType: FlowType.SELF},
    {
      name: 'the shipped admin user flow',
      inputs: [{identifier: 'attribute', type: 'TEXT_INPUT', required: true, defaultValue: 'email'}],
      flowType: FlowType.ADMIN_INITIATED,
    },
    {name: 'the shipped agent owner flow', inputs: [ownerInput], flowType: FlowType.ADMIN_INITIATED},
  ])('changes nothing in $name just by being opened and closed', ({inputs, flowType}) => {
    const {unmount} = renderIn(
      <AttributeVerificationProperties resource={prepareStep(inputs)} onChange={mockOnChange} />,
      flowType,
    );
    unmount();

    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it("stores an agent's owner as a constant attribute input", async () => {
    const user = userEvent.setup();
    renderIn(
      <AttributeVerificationProperties resource={prepareStep()} onChange={mockOnChange} />,
      FlowType.ADMIN_INITIATED,
    );

    await user.click(screen.getByRole('radio', {name: 'Agent owner'}));

    expect(mockOnChange).toHaveBeenCalledWith('data.action.executor.inputs', [ownerInput], expect.anything());
  });

  it('drops the owner input when switched back to a user attribute', async () => {
    const user = userEvent.setup();
    renderIn(
      <AttributeVerificationProperties resource={prepareStep([ownerInput])} onChange={mockOnChange} />,
      FlowType.ADMIN_INITIATED,
    );

    await user.click(screen.getByRole('radio', {name: 'User'}));

    expect(mockOnChange).toHaveBeenCalledWith('data.action.executor.inputs', undefined, expect.anything());
  });

  it('stores a typed attribute as a default the request can replace', async () => {
    const user = userEvent.setup();
    renderIn(<AttributeVerificationProperties resource={prepareStep()} onChange={mockOnChange} />, FlowType.SELF);

    await user.type(screen.getByRole('combobox', {name: 'Attribute to verify'}), 'mobileNumber');
    await user.tab();

    expect(mockOnChange).toHaveBeenCalledWith(
      'data.action.executor.inputs',
      [{identifier: 'attribute', type: 'TEXT_INPUT', required: true, defaultValue: 'mobileNumber'}],
      expect.anything(),
    );
  });

  it.each([FlowType.SELF, FlowType.REGISTRATION, FlowType.RECOVERY])(
    "warns that a %s flow cannot verify an agent's owner",
    (flowType) => {
      renderIn(
        <AttributeVerificationProperties resource={prepareStep([ownerInput])} onChange={mockOnChange} />,
        flowType,
      );

      expect(screen.getByTestId('attribute-verification-owner-flow-type')).toHaveTextContent(
        'Only an administrator-initiated flow can verify an agent owner',
      );
    },
  );
});
