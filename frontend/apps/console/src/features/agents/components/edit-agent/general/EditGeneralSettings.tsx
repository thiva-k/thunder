// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Stack} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import OwnerSummarySection from './OwnerSummarySection';
import type {Agent} from '../../../models/agent';
import AgentDeleteDialog from '../../AgentDeleteDialog';
import AttributesSummarySection from '../attributes/AttributesSummarySection';
import DangerZoneSection from '../general-settings/DangerZoneSection';
import OrganizationUnitSection from '../general-settings/OrganizationUnitSection';
import QuickCopySection from '../general-settings/QuickCopySection';

interface EditGeneralSettingsProps {
  agent: Agent;
  copiedField: string | null;
  onCopyToClipboard: (text: string, fieldName: string) => Promise<void>;
  onDeleteSuccess?: () => void;
}

export default function EditGeneralSettings({
  agent,
  copiedField,
  onCopyToClipboard,
  onDeleteSuccess = undefined,
}: EditGeneralSettingsProps): JSX.Element {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  return (
    <>
      <Stack spacing={3}>
        <QuickCopySection agent={agent} copiedField={copiedField} onCopyToClipboard={onCopyToClipboard} />
        <OwnerSummarySection agent={agent} />
        <AttributesSummarySection agent={agent} />
        <OrganizationUnitSection agent={agent} copiedField={copiedField} onCopyToClipboard={onCopyToClipboard} />
        {!agent.isReadOnly && <DangerZoneSection onDeleteClick={() => setDeleteDialogOpen(true)} />}
      </Stack>

      <AgentDeleteDialog
        open={deleteDialogOpen}
        agentId={agent.id}
        onClose={() => setDeleteDialogOpen(false)}
        onSuccess={onDeleteSuccess}
      />
    </>
  );
}
