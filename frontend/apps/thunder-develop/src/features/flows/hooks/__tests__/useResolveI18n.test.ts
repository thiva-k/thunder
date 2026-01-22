/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {renderHook} from '@testing-library/react';
import useResolveI18n from '../useResolveI18n';

// Mock react-i18next
const mockT = vi.fn((key: string) => `translated:${key}`);

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: mockT,
  }),
}));

describe('useResolveI18n', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Hook Interface', () => {
    it('should return a string', () => {
      const {result} = renderHook(() => useResolveI18n('test'));

      expect(typeof result.current).toBe('string');
    });

    it('should be memoized and return same value for same inputs', () => {
      const {result, rerender} = renderHook(({value}) => useResolveI18n(value), {
        initialProps: {value: 'test'},
      });

      const firstResult = result.current;
      rerender({value: 'test'});
      const secondResult = result.current;

      expect(firstResult).toBe(secondResult);
    });
  });

  describe('Empty and undefined values', () => {
    it('should return empty string for undefined value', () => {
      const {result} = renderHook(() => useResolveI18n(undefined));

      expect(result.current).toBe('');
    });

    it('should return empty string for empty string value', () => {
      const {result} = renderHook(() => useResolveI18n(''));

      expect(result.current).toBe('');
    });
  });

  describe('Non-i18n pattern values', () => {
    it('should return original value when not matching i18n pattern', () => {
      const {result} = renderHook(() => useResolveI18n('Hello World'));

      expect(result.current).toBe('Hello World');
      expect(mockT).not.toHaveBeenCalled();
    });

    it('should return original value for partial pattern match', () => {
      const {result} = renderHook(() => useResolveI18n('{{t(key)'));

      expect(result.current).toBe('{{t(key)');
      expect(mockT).not.toHaveBeenCalled();
    });

    it('should return original value for malformed pattern', () => {
      const {result} = renderHook(() => useResolveI18n('{t(key)}'));

      expect(result.current).toBe('{t(key)}');
      expect(mockT).not.toHaveBeenCalled();
    });
  });

  describe('i18n pattern resolution', () => {
    it('should resolve i18n pattern and return translated value', () => {
      const {result} = renderHook(() => useResolveI18n('{{t(common.button.submit)}}'));

      expect(result.current).toBe('translated:common.button.submit');
      expect(mockT).toHaveBeenCalledWith('common.button.submit');
    });

    it('should resolve i18n pattern with simple key', () => {
      const {result} = renderHook(() => useResolveI18n('{{t(title)}}'));

      expect(result.current).toBe('translated:title');
      expect(mockT).toHaveBeenCalledWith('title');
    });

    it('should resolve i18n pattern with nested namespace key', () => {
      const {result} = renderHook(() => useResolveI18n('{{t(flows:steps.view.label)}}'));

      expect(result.current).toBe('translated:flows:steps.view.label');
      expect(mockT).toHaveBeenCalledWith('flows:steps.view.label');
    });
  });

  describe('stripHtml option', () => {
    it('should strip HTML tags when stripHtml is true', () => {
      const {result} = renderHook(() => useResolveI18n('<p>{{t(common.title)}}</p>', true));

      expect(result.current).toBe('translated:common.title');
      expect(mockT).toHaveBeenCalledWith('common.title');
    });

    it('should strip nested HTML tags when stripHtml is true', () => {
      const {result} = renderHook(() => useResolveI18n('<div><span>{{t(nested.key)}}</span></div>', true));

      expect(result.current).toBe('translated:nested.key');
      expect(mockT).toHaveBeenCalledWith('nested.key');
    });

    it('should not strip HTML tags when stripHtml is false (default)', () => {
      const {result} = renderHook(() => useResolveI18n('<p>{{t(common.title)}}</p>'));

      // With HTML tags, it won't match the i18n pattern, so returns original
      expect(result.current).toBe('<p>{{t(common.title)}}</p>');
      expect(mockT).not.toHaveBeenCalled();
    });

    it('should return original value when stripHtml is true but no i18n pattern', () => {
      const {result} = renderHook(() => useResolveI18n('<p>Hello World</p>', true));

      expect(result.current).toBe('<p>Hello World</p>');
      expect(mockT).not.toHaveBeenCalled();
    });
  });

  describe('Value changes', () => {
    it('should update when value changes', () => {
      const {result, rerender} = renderHook(({value}) => useResolveI18n(value), {
        initialProps: {value: '{{t(key1)}}'},
      });

      expect(result.current).toBe('translated:key1');

      rerender({value: '{{t(key2)}}'});

      expect(result.current).toBe('translated:key2');
    });

    it('should update when stripHtml option changes', () => {
      const {result, rerender} = renderHook(({value, stripHtml}) => useResolveI18n(value, stripHtml), {
        initialProps: {value: '<p>{{t(key)}}</p>', stripHtml: false},
      });

      // Without stripHtml, returns original (doesn't match pattern)
      expect(result.current).toBe('<p>{{t(key)}}</p>');

      rerender({value: '<p>{{t(key)}}</p>', stripHtml: true});

      // With stripHtml, resolves the pattern
      expect(result.current).toBe('translated:key');
    });
  });
});
