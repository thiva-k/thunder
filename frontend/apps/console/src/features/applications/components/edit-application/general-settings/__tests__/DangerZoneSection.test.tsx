// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {screen, fireEvent, renderWithProviders} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import DangerZoneSection from '../DangerZoneSection';

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:edit.general.sections.dangerZone.title': 'Danger Zone',
        'applications:edit.general.sections.dangerZone.description':
          'Actions in this section are irreversible. Proceed with caution.',
        'applications:edit.general.sections.dangerZone.regenerateSecret.title': 'Regenerate Client Secret',
        'applications:edit.general.sections.dangerZone.regenerateSecret.description':
          'Regenerating the client secret will immediately invalidate the current client secret and cannot be undone.',
        'applications:edit.general.sections.dangerZone.regenerateSecret.button': 'Regenerate Client Secret',
        'applications:edit.general.sections.dangerZone.deleteApplication.title': 'Delete Application',
        'applications:edit.general.sections.dangerZone.deleteApplication.description':
          'Permanently delete this application and all associated data. This action cannot be undone.',
        'applications:edit.general.sections.dangerZone.deleteApplication.button': 'Delete Application',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('DangerZoneSection', () => {
  const mockOnRegenerateClick = vi.fn();
  const mockOnDeleteClick = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render the danger zone section', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.getByText('Danger Zone')).toBeInTheDocument();
    expect(screen.getByText('Actions in this section are irreversible. Proceed with caution.')).toBeInTheDocument();
  });

  it('should always render delete application section', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.getByRole('heading', {name: 'Delete Application', level: 6})).toBeInTheDocument();
    expect(
      screen.getByText('Permanently delete this application and all associated data. This action cannot be undone.'),
    ).toBeInTheDocument();
  });

  it('should render delete button', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    const deleteButton = screen.getByRole('button', {name: 'Delete Application'});
    expect(deleteButton).toBeInTheDocument();
  });

  it('should call onDeleteClick when delete button is clicked', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    const deleteButton = screen.getByRole('button', {name: 'Delete Application'});
    fireEvent.click(deleteButton);

    expect(mockOnDeleteClick).toHaveBeenCalledTimes(1);
  });

  it('should render delete button with error color', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    const deleteButton = screen.getByRole('button', {name: 'Delete Application'});
    expect(deleteButton).toHaveClass('MuiButton-colorError');
  });

  it('should not render regenerate secret section by default', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.queryByRole('button', {name: 'Regenerate Client Secret'})).not.toBeInTheDocument();
  });

  it('should render regenerate secret section when showRegenerateSecret is true', () => {
    renderWithProviders(
      <DangerZoneSection
        showRegenerateSecret
        onRegenerateClick={mockOnRegenerateClick}
        onDeleteClick={mockOnDeleteClick}
      />,
    );

    expect(screen.getByRole('heading', {name: 'Regenerate Client Secret', level: 6})).toBeInTheDocument();
    expect(
      screen.getByText(
        'Regenerating the client secret will immediately invalidate the current client secret and cannot be undone.',
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Regenerate Client Secret'})).toBeInTheDocument();
  });

  it('should call onRegenerateClick when regenerate button is clicked', () => {
    renderWithProviders(
      <DangerZoneSection
        showRegenerateSecret
        onRegenerateClick={mockOnRegenerateClick}
        onDeleteClick={mockOnDeleteClick}
      />,
    );

    const regenerateButton = screen.getByRole('button', {name: 'Regenerate Client Secret'});
    fireEvent.click(regenerateButton);

    expect(mockOnRegenerateClick).toHaveBeenCalledTimes(1);
  });

  it('should render both sections with a divider when showRegenerateSecret is true', () => {
    renderWithProviders(
      <DangerZoneSection
        showRegenerateSecret
        onRegenerateClick={mockOnRegenerateClick}
        onDeleteClick={mockOnDeleteClick}
      />,
    );

    expect(screen.getByRole('button', {name: 'Regenerate Client Secret'})).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Delete Application'})).toBeInTheDocument();
  });
});
