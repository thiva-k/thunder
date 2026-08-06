// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen, fireEvent, cleanup} from '@testing-library/react';
import type {Theme} from '@thunderid/design';
import {OxygenUIThemeProvider} from '@wso2/oxygen-ui';
import {useState} from 'react';
import {describe, it, expect, afterEach, vi} from 'vitest';
import TypographyBuilderContent from '../TypographyBuilderContent';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (_key: string, fallback?: string) => fallback ?? _key}),
}));

// useFontStylesheetLink needs a ConfigProvider, which this test tree doesn't otherwise set up.
vi.mock('@thunderid/contexts', () => ({
  useConfig: () => ({config: {brand: {product_name: 'ThunderID'}}}),
}));

afterEach(() => {
  cleanup();
});

interface MutableTypography {
  fontFamily?: string;
  font?: {importURL?: string};
}

function makeDraft(typography: MutableTypography = {fontFamily: ''}): Theme {
  return {typography} as unknown as Theme;
}

function typographyOf(draft: Theme): MutableTypography {
  return draft.typography as unknown as MutableTypography;
}

function renderEditor(draft: Theme) {
  const onUpdate = vi.fn((updater: (d: Theme) => void) => updater(draft));
  render(
    <OxygenUIThemeProvider>
      <TypographyBuilderContent draft={draft} onUpdate={onUpdate} />
    </OxygenUIThemeProvider>,
  );
  return onUpdate;
}

/** Mirrors the real ThemeConfigPanel.updateDraft: clones the draft and re-renders with the new
 *  object, needed for tests that exercise multiple sequential edits (renderEditor's in-place
 *  mutation doesn't give a fresh object reference each time). */
function StatefulEditor({initialDraft, onDraftChange}: {initialDraft: Theme; onDraftChange: (d: Theme) => void}) {
  const [draft, setDraft] = useState(initialDraft);
  const onUpdate = (updater: (d: Theme) => void): void => {
    const next = JSON.parse(JSON.stringify(draft)) as Theme;
    updater(next);
    setDraft(next);
    onDraftChange(next);
  };
  return <TypographyBuilderContent draft={draft} onUpdate={onUpdate} />;
}

function renderStatefulEditor(initialDraft: Theme) {
  let currentDraft = initialDraft;
  const utils = render(
    <OxygenUIThemeProvider>
      <StatefulEditor
        initialDraft={initialDraft}
        onDraftChange={(d) => {
          currentDraft = d;
        }}
      />
    </OxygenUIThemeProvider>,
  );
  return {...utils, getDraft: () => currentDraft};
}

describe('TypographyBuilderContent font modes', () => {
  it('defaults to the web-safe font picker when no import URL is set', () => {
    renderEditor(makeDraft());
    expect(screen.getByRole('combobox')).toBeTruthy();
    expect(screen.queryByLabelText('Font Import URL')).toBeNull();
  });

  it('shows the bare name selected when the theme stores the default font stack', () => {
    renderEditor(makeDraft({fontFamily: "'Inter Variable', sans-serif"}));
    expect(screen.getByRole<HTMLInputElement>('combobox').value).toBe('Inter Variable');
  });

  it('starts in import mode showing the configured import URL', () => {
    const draft = makeDraft({
      fontFamily: 'Poppins',
      font: {importURL: 'https://fonts.googleapis.com/css2?family=Poppins'},
    });
    renderEditor(draft);
    const urlField = screen.getByLabelText<HTMLInputElement>('Font Import URL');
    expect(urlField.value).toBe('https://fonts.googleapis.com/css2?family=Poppins');
  });

  it('reveals the import fields when switching to import mode', () => {
    renderEditor(makeDraft());
    fireEvent.click(screen.getByRole('button', {name: 'Use a Custom Font'}));
    expect(screen.getByLabelText('Font Import URL')).toBeTruthy();
    expect(screen.getAllByLabelText('Font Family').length).toBeGreaterThan(0);
  });

  it('clears the inherited family when switching to import so the field shows its placeholder', () => {
    const draft = makeDraft({fontFamily: "'Inter Variable', sans-serif"});
    renderEditor(draft);
    fireEvent.click(screen.getByRole('button', {name: 'Use a Custom Font'}));
    expect(typographyOf(draft).fontFamily).toBe('');
    expect(screen.getByLabelText<HTMLInputElement>('Font Family').value).toBe('');
  });

  it('typing an import URL writes typography.font.importURL', () => {
    const draft = makeDraft();
    renderEditor(draft);
    fireEvent.click(screen.getByRole('button', {name: 'Use a Custom Font'}));
    fireEvent.change(screen.getByLabelText('Font Import URL'), {
      target: {value: 'https://fonts.bunny.net/css?family=inter'},
    });
    expect(typographyOf(draft).font?.importURL).toBe('https://fonts.bunny.net/css?family=inter');
  });

  it('clears the applied import URL when switching to a web-safe font', () => {
    const draft = makeDraft({
      fontFamily: 'Poppins',
      font: {importURL: 'https://fonts.googleapis.com/css2?family=Poppins'},
    });
    renderEditor(draft);
    fireEvent.click(screen.getByRole('button', {name: 'Use a web-safe font'}));
    expect(typographyOf(draft).font).toBeUndefined();
  });

  it('remembers the import URL and family when toggling to web-safe and back', () => {
    const {getDraft} = renderStatefulEditor(
      makeDraft({fontFamily: 'Poppins', font: {importURL: 'https://fonts.googleapis.com/css2?family=Poppins'}}),
    );
    // Toggling to web-safe drops the import from the applied theme, but remembers it.
    fireEvent.click(screen.getByRole('button', {name: 'Use a web-safe font'}));
    expect(typographyOf(getDraft()).font).toBeUndefined();
    // Toggling back to import restores the remembered URL and family.
    fireEvent.click(screen.getByRole('button', {name: 'Use a Custom Font'}));
    expect(typographyOf(getDraft()).font?.importURL).toBe('https://fonts.googleapis.com/css2?family=Poppins');
    expect(typographyOf(getDraft()).fontFamily).toBe('Poppins');
  });

  it('promotes the toggle to import mode when a draft change restores an import URL (e.g. Revert)', () => {
    const webSafeDraft = makeDraft({fontFamily: 'Arial'});
    const {rerender} = render(
      <OxygenUIThemeProvider>
        <TypographyBuilderContent draft={webSafeDraft} onUpdate={vi.fn()} />
      </OxygenUIThemeProvider>,
    );
    expect(screen.queryByLabelText('Font Import URL')).toBeNull();

    // A Revert replaces the draft in place with one that carries an import URL again.
    const revertedDraft = makeDraft({
      fontFamily: 'Poppins',
      font: {importURL: 'https://fonts.googleapis.com/css2?family=Poppins'},
    });
    rerender(
      <OxygenUIThemeProvider>
        <TypographyBuilderContent draft={revertedDraft} onUpdate={vi.fn()} />
      </OxygenUIThemeProvider>,
    );
    const urlField = screen.getByLabelText<HTMLInputElement>('Font Import URL');
    expect(urlField.value).toBe('https://fonts.googleapis.com/css2?family=Poppins');
  });

  it('demotes the toggle back to web-safe when a draft change restores a plain family with no import URL (e.g. Revert)', () => {
    const importDraft = makeDraft({
      fontFamily: 'Poppins',
      font: {importURL: 'https://fonts.googleapis.com/css2?family=Poppins'},
    });
    const {rerender} = render(
      <OxygenUIThemeProvider>
        <TypographyBuilderContent draft={importDraft} onUpdate={vi.fn()} />
      </OxygenUIThemeProvider>,
    );
    expect(screen.getByLabelText('Font Import URL')).toBeTruthy();

    // A Revert replaces the draft in place with the original web-safe theme (no import URL).
    const revertedDraft = makeDraft({fontFamily: 'Arial'});
    rerender(
      <OxygenUIThemeProvider>
        <TypographyBuilderContent draft={revertedDraft} onUpdate={vi.fn()} />
      </OxygenUIThemeProvider>,
    );
    expect(screen.queryByLabelText('Font Import URL')).toBeNull();
    expect(screen.getByRole('combobox')).toBeTruthy();
  });

  it('demotes the toggle back to web-safe when a draft change restores an empty family with no import URL (e.g. Revert)', () => {
    const importDraft = makeDraft({
      fontFamily: 'Poppins',
      font: {importURL: 'https://fonts.googleapis.com/css2?family=Poppins'},
    });
    const {rerender} = render(
      <OxygenUIThemeProvider>
        <TypographyBuilderContent draft={importDraft} onUpdate={vi.fn()} />
      </OxygenUIThemeProvider>,
    );
    expect(screen.getByLabelText('Font Import URL')).toBeTruthy();

    // A Revert can restore a draft with neither a family nor an import URL set.
    const revertedDraft = makeDraft();
    rerender(
      <OxygenUIThemeProvider>
        <TypographyBuilderContent draft={revertedDraft} onUpdate={vi.fn()} />
      </OxygenUIThemeProvider>,
    );
    expect(screen.queryByLabelText('Font Import URL')).toBeNull();
    expect(screen.getByRole('combobox')).toBeTruthy();
  });

  it('falls the live preview back to the product default (not a generic serif) for a partial font name', () => {
    renderStatefulEditor(makeDraft());
    fireEvent.click(screen.getByRole('button', {name: 'Use a Custom Font'}));
    fireEvent.change(screen.getByLabelText('Font Family'), {target: {value: 'P'}});

    // sx compiles to a generated CSS class via emotion, so assert against the injected stylesheet.
    screen.getByText('The quick brown fox jumps over the lazy dog.');
    const styles = Array.from(document.querySelectorAll('style'))
      .map((style) => style.textContent)
      .join('\n');
    expect(styles).toContain('font-family:P,');
    expect(styles).toMatch(/Inter Variable/);
  });

  it('loads the imported font stylesheet into the main document, not just the GatePreview iframe', () => {
    renderEditor(
      makeDraft({fontFamily: 'Poppins', font: {importURL: 'https://fonts.googleapis.com/css2?family=Poppins'}}),
    );
    const link = document.querySelector(
      'link[rel="stylesheet"][href="https://fonts.googleapis.com/css2?family=Poppins"]',
    );
    expect(link).not.toBeNull();
  });

  it('clearing the import URL also clears the family, reverting the whole selection to default', () => {
    const draft = makeDraft({
      fontFamily: 'Poppins',
      font: {importURL: 'https://fonts.googleapis.com/css2?family=Poppins'},
    });
    renderEditor(draft);
    fireEvent.change(screen.getByLabelText('Font Import URL'), {target: {value: ''}});
    expect(typographyOf(draft).fontFamily).toBe('');
    expect(typographyOf(draft).font).toBeUndefined();
  });
});
