// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, renderHook, screen} from '@thunderid/test-utils';
import {useTranslation} from 'react-i18next';
import {describe, it, expect, vi, beforeAll, beforeEach} from 'vitest';
import ConfigureOrganizationUnit, {type ConfigureOrganizationUnitProps} from '../ConfigureOrganizationUnit';

// Mock OrganizationUnitTreePicker to isolate this component's logic
vi.mock('@thunderid/configure-organization-units', () => ({
  OrganizationUnitTreePicker: ({
    rootOuId = undefined,
    value,
    onChange,
    maxHeight = undefined,
  }: {
    rootOuId?: string;
    value: string;
    onChange: (ouId: string) => void;
    maxHeight?: number;
  }) => (
    <div data-testid="ou-tree-picker" data-root-ou-id={rootOuId} data-value={value} data-max-height={maxHeight}>
      <button type="button" data-testid="select-child-ou" onClick={() => onChange('child-ou-1')}>
        Select Child OU
      </button>
    </div>
  ),
}));

describe('ConfigureOrganizationUnit', () => {
  let t: (key: string) => string;

  beforeAll(() => {
    ({t} = renderHook(() => useTranslation()).result.current);
  });

  const mockOnOuIdChange = vi.fn();
  const mockOnReadyChange = vi.fn();

  const defaultProps: ConfigureOrganizationUnitProps = {
    rootOuId: 'root-ou-id',
    selectedOuId: '',
    onOuIdChange: mockOnOuIdChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = (props: Partial<ConfigureOrganizationUnitProps> = {}) =>
    render(<ConfigureOrganizationUnit {...defaultProps} {...props} />);

  it('renders the component with title and subtitle', () => {
    renderComponent();

    expect(
      screen.getByRole('heading', {name: t('users:createWizard.selectOrganizationUnit.title')}),
    ).toBeInTheDocument();
    expect(screen.getByText(t('users:createWizard.selectOrganizationUnit.subtitle'))).toBeInTheDocument();
  });

  it('renders the field label', () => {
    renderComponent();

    expect(
      screen.getByText(t('users:createWizard.selectOrganizationUnit.fieldLabel'), {exact: false, selector: 'label'}),
    ).toBeInTheDocument();
  });

  it('renders the OU tree picker with correct props', () => {
    renderComponent({selectedOuId: 'some-ou'});

    const picker = screen.getByTestId('ou-tree-picker');
    expect(picker).toBeInTheDocument();
    expect(picker).toHaveAttribute('data-root-ou-id', 'root-ou-id');
    expect(picker).toHaveAttribute('data-value', 'some-ou');
    expect(picker).toHaveAttribute('data-max-height', '500');
  });

  it('has the correct data-testid', () => {
    renderComponent();

    expect(screen.getByTestId('configure-organization-unit')).toBeInTheDocument();
  });

  it('auto-selects rootOuId when selectedOuId is empty', () => {
    renderComponent({selectedOuId: ''});

    expect(mockOnOuIdChange).toHaveBeenCalledWith('root-ou-id');
  });

  it('does not auto-select when selectedOuId is already set', () => {
    renderComponent({selectedOuId: 'existing-ou'});

    expect(mockOnOuIdChange).not.toHaveBeenCalled();
  });

  describe('onReadyChange callback', () => {
    it('calls onReadyChange with true when an OU is selected', () => {
      renderComponent({
        selectedOuId: 'some-ou',
        onReadyChange: mockOnReadyChange,
      });

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });

    it('calls onReadyChange with false when no OU is selected', () => {
      renderComponent({
        selectedOuId: '',
        onReadyChange: mockOnReadyChange,
      });

      // auto-select fires first, but onReadyChange is called with selectedOuId.length > 0
      // Since selectedOuId is '' initially, onReadyChange(false) fires
      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
    });

    it('does not crash when onReadyChange is undefined', () => {
      expect(() => {
        renderComponent({selectedOuId: 'some-ou', onReadyChange: undefined});
      }).not.toThrow();
    });

    it('calls onReadyChange when selectedOuId transitions from empty to non-empty', () => {
      const {rerender} = render(
        <ConfigureOrganizationUnit {...defaultProps} selectedOuId="" onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(false);
      mockOnReadyChange.mockClear();

      rerender(
        <ConfigureOrganizationUnit {...defaultProps} selectedOuId="some-ou" onReadyChange={mockOnReadyChange} />,
      );

      expect(mockOnReadyChange).toHaveBeenCalledWith(true);
    });
  });
});
