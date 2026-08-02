// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import type {JSX} from 'react';
import GatePreview from '../../../components/GatePreview/GatePreview';
import useThemeBuilder from '../contexts/ThemeBuilder/useThemeBuilder';

// Re-export so existing imports of Viewport from this file keep working.
export type {Viewport} from '../../../components/GatePreview/GatePreview';

interface ThemePreviewPanelProps {
  themeId: string | null;
  toolbarPortal?: HTMLElement | null;
}

export default function ThemePreviewPanel({themeId, toolbarPortal = undefined}: ThemePreviewPanelProps): JSX.Element {
  const {draftTheme, displayName} = useThemeBuilder();

  // undefined → no theme selected yet (show prompt), null → loading spinner
  const theme = themeId === null && draftTheme === null ? undefined : draftTheme;

  return <GatePreview theme={theme} displayName={displayName ?? ''} toolbarPortal={toolbarPortal} />;
}
