/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {renderHook} from '@testing-library/react';
import {describe, it, expect, vi} from 'vitest';
import useFlowRoutes, {defaultFlowRoutePaths} from '../useFlowRoutes';

const mockUseRoutes = vi.fn();

vi.mock('@thunderid/contexts', () => ({
  useRoutes: () => mockUseRoutes() as unknown,
}));

describe('useFlowRoutes', () => {
  describe('defaults', () => {
    it('should fall back to the default paths when the host supplies none', () => {
      mockUseRoutes.mockReturnValue({});

      const {result} = renderHook(() => useFlowRoutes());

      expect(result.current.flows.list()).toBe('/flows');
      expect(result.current.flows.create()).toBe('/flows/create');
      expect(result.current.flows.detail('flow-1')).toBe('/flows/flow-1');
    });

    it('should expose the same defaults through defaultFlowRoutePaths', () => {
      expect(defaultFlowRoutePaths.flows.list()).toBe('/flows');
      expect(defaultFlowRoutePaths.flows.create()).toBe('/flows/create');
      expect(defaultFlowRoutePaths.flows.detail('flow-1')).toBe('/flows/flow-1');
    });
  });

  describe('host overrides', () => {
    it('should prefer the host supplied paths when present', () => {
      mockUseRoutes.mockReturnValue({
        flows: {
          list: () => '/custom/flows',
          create: () => '/custom/flows/new',
          detail: (flowId: string) => `/custom/flows/${flowId}`,
        },
      });

      const {result} = renderHook(() => useFlowRoutes());

      expect(result.current.flows.list()).toBe('/custom/flows');
      expect(result.current.flows.create()).toBe('/custom/flows/new');
      expect(result.current.flows.detail('flow-2')).toBe('/custom/flows/flow-2');
    });
  });
});
