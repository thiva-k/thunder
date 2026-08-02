// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen, fireEvent} from '@thunderid/test-utils';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import SelectConnectionType from '../SelectConnectionType';

describe('SelectConnectionType', () => {
  const onSelect = vi.fn();
  beforeEach(() => vi.clearAllMocks());

  it('renders the OIDC, OAuth2, trusted-idp, and SMS gateway options', () => {
    render(<SelectConnectionType selectedType={null} onSelect={onSelect} />);
    expect(screen.getByText('What kind of connection do you want to add?')).toBeInTheDocument();
    expect(screen.queryByText('Connection type')).not.toBeInTheDocument();
    expect(screen.getByTestId('connection-type-option-oidc')).toBeInTheDocument();
    expect(screen.getByTestId('connection-type-option-oauth')).toBeInTheDocument();
    expect(screen.getByTestId('connection-type-option-trusted-idp')).toBeInTheDocument();
    expect(screen.getByTestId('connection-type-option-custom-sms')).toBeInTheDocument();
  });

  it('selects the OIDC type when clicked', () => {
    render(<SelectConnectionType selectedType={null} onSelect={onSelect} />);
    fireEvent.click(screen.getByTestId('connection-type-option-oidc'));
    expect(onSelect).toHaveBeenCalledWith('oidc');
  });

  it('selects the OAuth2 type when clicked', () => {
    render(<SelectConnectionType selectedType={null} onSelect={onSelect} />);
    fireEvent.click(screen.getByTestId('connection-type-option-oauth'));
    expect(onSelect).toHaveBeenCalledWith('oauth');
  });

  it('does not select the disabled Custom SMS gateway option', () => {
    render(<SelectConnectionType selectedType={null} onSelect={onSelect} />);
    fireEvent.click(screen.getByTestId('connection-type-option-custom-sms'));
    expect(onSelect).not.toHaveBeenCalled();
  });

  it('selects the Trusted Token Issuer type when clicked', () => {
    render(<SelectConnectionType selectedType={null} onSelect={onSelect} />);
    fireEvent.click(screen.getByTestId('connection-type-option-trusted-idp'));
    expect(onSelect).toHaveBeenCalledWith('trusted-idp');
  });

  it('renders the Trusted Token Issuer option before the coming-soon SMS gateway option', () => {
    render(<SelectConnectionType selectedType={null} onSelect={onSelect} />);

    const optionIds = screen.getAllByTestId(/^connection-type-option-/).map((el) => el.getAttribute('data-testid'));
    expect(optionIds.indexOf('connection-type-option-trusted-idp')).toBeLessThan(
      optionIds.indexOf('connection-type-option-custom-sms'),
    );
  });
});
