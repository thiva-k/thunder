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

import {describe, it, expect} from 'vitest';
import {prismjsInjectCore} from '../prismjs-inject-core';

function transform(id: string, code = 'Prism.languages.python = {};') {
  const plugin = prismjsInjectCore();
  const hook = plugin.transform as (code: string, id: string) => {code: string; map: null} | null;
  return hook.call(plugin, code, id);
}

describe('prismjsInjectCore', () => {
  it('has the expected plugin name', () => {
    expect(prismjsInjectCore().name).toBe('prismjs-inject-core');
  });

  it('injects a Prism import for a POSIX language file path', () => {
    const result = transform('/repo/node_modules/prismjs/components/prism-python.js');
    expect(result?.code).toBe("import Prism from 'prismjs';\nPrism.languages.python = {};");
    expect(result?.map).toBeNull();
  });

  it('injects a Prism import for a Windows-style language file path', () => {
    const result = transform('C:\\repo\\node_modules\\prismjs\\components\\prism-python.js');
    expect(result?.code).toBe("import Prism from 'prismjs';\nPrism.languages.python = {};");
  });

  it('does not inject an import for the core module itself', () => {
    expect(transform('/repo/node_modules/prismjs/components/prism-core.js')).toBeNull();
  });

  it('does not inject an import for files outside prismjs/components', () => {
    expect(transform('/repo/src/components/prism-python.js')).toBeNull();
  });

  it('does not inject an import for unrelated modules', () => {
    expect(transform('/repo/src/index.ts')).toBeNull();
  });
});
