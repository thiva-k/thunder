// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {loader} from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
// eslint-disable-next-line import-x/default
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
// eslint-disable-next-line import-x/default
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
// eslint-disable-next-line import-x/default
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';

self.MonacoEnvironment = {
  getWorker(_, label) {
    if (label === 'json') return new jsonWorker();
    if (label === 'css' || label === 'scss' || label === 'less') return new cssWorker();
    return new editorWorker();
  },
};

loader.config({monaco});
