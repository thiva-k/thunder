// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Stack, Button, Alert} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import {useState, useCallback, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import AddAssignmentDialog from './AddAssignmentDialog';
import ManageAssignmentsSection from './ManageAssignmentsSection';
import useAddRoleAssignments from '../../../api/useAddRoleAssignments';
import useRemoveRoleAssignments from '../../../api/useRemoveRoleAssignments';
import type {RoleAssignment} from '../../../models/role';

interface EditAssignmentsSettingsProps {
  roleId: string;
  isReadOnly?: boolean;
}

/**
 * Assignments tab content for the Role edit page.
 * Provides assignment listing, add, and remove functionality.
 */
export default function EditAssignmentsSettings({
  roleId,
  isReadOnly = false,
}: EditAssignmentsSettingsProps): JSX.Element {
  const {t} = useTranslation();
  const addRoleAssignments = useAddRoleAssignments();
  const removeRoleAssignments = useRemoveRoleAssignments();
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeAssignmentTab, setActiveAssignmentTab] = useState(0);

  const handleAddAssignments = useCallback(
    (newAssignments: RoleAssignment[]) => {
      if (newAssignments.length === 0) return;
      addRoleAssignments.mutate(
        {roleId, assignments: newAssignments},
        {
          onSuccess: () => {
            setAddDialogOpen(false);
            setError(null);
          },
          onError: (err: Error) => {
            setError(err.message ?? t('roles:assignments.add.error'));
          },
        },
      );
    },
    [roleId, addRoleAssignments, t],
  );

  const handleRemoveAssignment = useCallback(
    (assignmentToRemove: RoleAssignment) => {
      removeRoleAssignments.mutate(
        {roleId, assignments: [{id: assignmentToRemove.id, type: assignmentToRemove.type}]},
        {
          onSuccess: () => {
            setError(null);
          },
          onError: (err: Error) => {
            setError(err.message ?? t('roles:assignments.remove.error'));
          },
        },
      );
    },
    [roleId, removeRoleAssignments, t],
  );

  return (
    <Stack spacing={3}>
      {error && (
        <Alert severity="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      <ManageAssignmentsSection
        roleId={roleId}
        onRemoveAssignment={handleRemoveAssignment}
        activeAssignmentTab={activeAssignmentTab}
        onAssignmentTabChange={setActiveAssignmentTab}
        isReadOnly={isReadOnly}
        headerAction={
          !isReadOnly ? (
            <Button
              variant="contained"
              size="small"
              startIcon={<Plus size={16} />}
              onClick={() => setAddDialogOpen(true)}
            >
              {t('roles:edit.assignments.sections.manage.addAssignment')}
            </Button>
          ) : undefined
        }
      />

      {addDialogOpen && !isReadOnly && (
        <AddAssignmentDialog
          open={addDialogOpen}
          roleId={roleId}
          onClose={() => setAddDialogOpen(false)}
          onAdd={handleAddAssignments}
          initialTab={activeAssignmentTab}
        />
      )}
    </Stack>
  );
}
