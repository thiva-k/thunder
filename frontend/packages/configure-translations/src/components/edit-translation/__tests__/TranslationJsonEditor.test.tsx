// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {render, screen, act, fireEvent} from '@thunderid/test-utils';
import {describe, expect, it, vi, beforeEach, afterEach} from 'vitest';
import TranslationJsonEditor from '@/components/edit-translation/TranslationJsonEditor';

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next');
  return {
    ...actual,
    useTranslation: () => ({t: (key: string) => key}),
  };
});

// Monaco Editor is not available in jsdom; replace it with a plain textarea
// that mirrors the same value/onChange contract.
vi.mock('@monaco-editor/react', () => ({
  default: ({value, onChange}: {value: string; onChange?: (v: string | undefined) => void}) => (
    <textarea data-testid="monaco-editor" value={value} onChange={(e) => onChange?.(e.target.value)} />
  ),
}));

const sampleValues = {'actions.save': 'Save', 'actions.cancel': 'Cancel'};
const sampleServerKeys = Object.keys(sampleValues);

// Helper: fire a change event on the editor and advance the 400ms debounce.
// userEvent.type() deadlocks under vi.useFakeTimers() because its internal
// per-keystroke delays also use setTimeout; fireEvent.change() is synchronous
// and avoids the issue entirely.
function changeEditor(editor: HTMLElement, value: string) {
  fireEvent.change(editor, {target: {value}});
  act(() => {
    vi.advanceTimersByTime(400);
  });
}

describe('TranslationJsonEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('Rendering', () => {
    it('renders the Monaco editor with the initial JSON value', () => {
      render(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      const editor = screen.getByTestId('monaco-editor');
      const parsed = JSON.parse((editor as HTMLTextAreaElement).value) as Record<string, string>;

      expect(parsed).toEqual(sampleValues);
    });

    it('does not show the invalid-JSON warning on initial render', () => {
      render(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      expect(screen.queryByText('editor.jsonInvalid')).not.toBeInTheDocument();
    });
  });

  describe('Valid JSON changes', () => {
    it('calls onChange with the parsed record after the debounce fires', () => {
      const onChange = vi.fn();

      render(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={onChange}
        />,
      );

      changeEditor(screen.getByTestId('monaco-editor'), JSON.stringify({'actions.save': 'Enregistrer'}));

      expect(onChange).toHaveBeenCalledWith({'actions.save': 'Enregistrer'});
    });

    it('does not show the invalid-JSON warning for valid JSON', () => {
      render(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      changeEditor(screen.getByTestId('monaco-editor'), '{"key": "value"}');

      expect(screen.queryByText('editor.jsonInvalid')).not.toBeInTheDocument();
    });
  });

  describe('Invalid JSON handling', () => {
    it('shows a warning alert when the editor contains invalid JSON', () => {
      render(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      changeEditor(screen.getByTestId('monaco-editor'), '{not valid json');

      expect(screen.getByText('editor.jsonInvalid')).toBeInTheDocument();
    });

    it('does not call onChange while JSON is invalid', () => {
      const onChange = vi.fn();

      render(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={onChange}
        />,
      );

      changeEditor(screen.getByTestId('monaco-editor'), '{invalid');

      expect(onChange).not.toHaveBeenCalled();
    });

    it('does not show the warning alert when the editor is empty', () => {
      render(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      changeEditor(screen.getByTestId('monaco-editor'), '');

      expect(screen.queryByText('editor.jsonInvalid')).not.toBeInTheDocument();
    });
  });

  describe('External value updates', () => {
    it('syncs the editor when values prop changes to a new object reference', () => {
      const {rerender} = render(
        <TranslationJsonEditor
          values={sampleValues}
          serverKeys={sampleServerKeys}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      const newValues = {'page.title': 'My Page'};
      rerender(
        <TranslationJsonEditor
          values={newValues}
          serverKeys={Object.keys(newValues)}
          isCustomNamespace={false}
          colorMode="light"
          onChange={vi.fn()}
        />,
      );

      const editor = screen.getByTestId('monaco-editor');
      const parsed = JSON.parse((editor as HTMLTextAreaElement).value) as Record<string, string>;

      expect(parsed).toEqual(newValues);
    });
  });
});
