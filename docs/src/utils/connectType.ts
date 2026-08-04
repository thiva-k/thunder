// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useSyncExternalStore} from 'react';

export type ConnectType = 'app' | 'agent' | 'mcp';

const DEFAULT_TYPE: ConnectType = 'app';

// Shared in-memory state so the sidebar accordion and the docs-home selector
// stay in sync: changing one updates the other live. It is deliberately NOT
// persisted, so there is no storage access that could throw. The choice is kept
// across in-app navigation (the module stays loaded) and resets to the default
// on a full reload, which always starts the page on "Application".
let current: ConnectType = DEFAULT_TYPE;
const listeners = new Set<() => void>();

export function applyConnectType(type: ConnectType): void {
  current = type;
  listeners.forEach(fn => fn());
}

function subscribe(fn: () => void): () => void {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

export function useConnectType(): ConnectType {
  return useSyncExternalStore(subscribe, () => current, () => DEFAULT_TYPE);
}
