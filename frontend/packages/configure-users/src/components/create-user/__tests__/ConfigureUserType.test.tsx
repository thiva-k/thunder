// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import userEvent from '@testing-library/user-event';
import {render, renderHook, screen, waitFor, within} from '@thunderid/test-utils';
import {useTranslation} from 'react-i18next';
import {describe, it, expect, vi, beforeAll, beforeEach} from 'vitest';
import type {SchemaInterface} from '../../../models/users';
import ConfigureUserType, {type ConfigureUserTypeProps} from '../ConfigureUserType';

const mockSchemas: SchemaInterface[] = [
  {id: 'schema-1', name: 'Employee', ouId: 'ou-1'},
  {id: 'schema-2', name: 'Contractor', ouId: 'ou-2'},
  {id: 'schema-3', name: 'Vendor', ouId: 'ou-3'},
];

describe('ConfigureUserType', () => {
  let t: (key: string) => string;

  beforeAll(() => {
    ({t} = renderHook(() => useTranslation()).result.current);
  });

  const mockOnSchemaChange = vi.fn();
  const mockOnReadyChange = vi.fn();

  const defaultProps: ConfigureUserTypeProps = {
    schemas: mockSchemas,
    selectedSchema: null,
    onSchemaChange: mockOnSchemaChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = (props: Partial<ConfigureUserTypeProps> = {}) =>
    render(<ConfigureUserType {...defaultProps} {...props} />);

  it('renders the component with title and subtitle', () => {
    renderComponent();

    expect(screen.getByRole('heading', {name: t('users:createWizard.selectUserType.title')})).toBeInTheDocument();
    expect(screen.getByText(t('users:createWizard.selectUserType.subtitle'))).toBeInTheDocument();
  });

  it('renders the user type select field', () => {
    renderComponent();

    expect(
      screen.getByText(t('users:createWizard.selectUserType.fieldLabel'), {exact: false, selector: 'label'}),
    ).toBeInTheDocument();
    expect(screen.getByTestId('configure-user-type')).toBeInTheDocument();
  });

  it('renders placeholder when no schema is selected', () => {
    renderComponent();

    expect(screen.getByText(t('users:createWizard.selectUserType.placeholder'), {selector: 'em'})).toBeInTheDocument();
  });

  it('renders all schema options in the select', async () => {
    const user = userEvent.setup();
    renderComponent();

    const select = screen.getByRole('combobox');
    await user.click(select);

    const listbox = await screen.findByRole('listbox');
    await waitFor(() => {
      expect(within(listbox).getByText('Employee')).toBeInTheDocument();
      expect(within(listbox).getByText('Contractor')).toBeInTheDocument();
      expect(within(listbox).getByText('Vendor')).toBeInTheDocument();
    });
  });

  it('calls onSchemaChange when a schema is selected', async () => {
    const user = userEvent.setup();
    renderComponent();

    const select = screen.getByRole('combobox');
    await user.click(select);

    const listbox = await screen.findByRole('listbox');
    await user.click(within(listbox).getByText('Employee'));

    expect(mockOnSchemaChange).toHaveBeenCalledWith(mockSchemas[0]);
  });

  it('calls onSchemaChange when selecting a different schema', async () => {
    const user = userEvent.setup();
    renderComponent({selectedSchema: mockSchemas[0]});

    const select = screen.getByRole('combobox');
    await user.click(select);

    const listbox = await screen.findByRole('listbox');
    await user.click(within(listbox).getByText('Contractor'));

    expect(mockOnSchemaChange).toHaveBeenCalledWith(mockSchemas[1]);
  });

  it('displays the selected schema name', () => {
    renderComponent({selectedSchema: mockSchemas[0]});

    expect(screen.getByText('Employee')).toBeInTheDocument();
  });

  describe('onReadyChange callback', () => {
    it('calls onReadyChange with true when a schema is selected', () => {
      renderComponent({
        selectedSchema: mockSchemas[0],
        onReadyChange: mockOnReadyChange,
      });

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('calls onReadyChange with false when no schema is selected', () => {
      renderComponent({
        selectedSchema: null,
        onReadyChange: mockOnReadyChange,
      });

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('does not crash when onReadyChange is undefined', () => {
      expect(() => {
        renderComponent({selectedSchema: mockSchemas[0], onReadyChange: undefined});
      }).not.toThrow();
    });

    it('calls onReadyChange when selectedSchema transitions from null to non-null', () => {
      const {rerender} = render(
        <ConfigureUserType {...defaultProps} selectedSchema={null} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      mockOnReadyChange.mockClear();

      rerender(
        <ConfigureUserType {...defaultProps} selectedSchema={mockSchemas[0]} onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });
  });

  it('handles empty schemas list', () => {
    renderComponent({schemas: []});

    expect(screen.getByTestId('configure-user-type')).toBeInTheDocument();
  });
});
