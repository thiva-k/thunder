// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {
  Box,
  Button,
  Card,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
  Tooltip,
  Typography,
} from '@wso2/oxygen-ui';
import {
  CogIcon,
  ExternalLink as ExternalLinkIcon,
  TrashIcon,
  Workflow as WorkflowIcon,
} from '@wso2/oxygen-ui-icons-react';
import {Handle, Position, useNodeId, useReactFlow} from '@xyflow/react';
import {memo, useMemo, useState, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import useFlowRoutes from '../../../../hooks/useFlowRoutes';
import ValidationErrorBoundary from '../../../validation-panel/ValidationErrorBoundary';
import type {CommonStepFactoryPropsInterface} from '../CommonStepFactory';
import StepTitle from '../StepTitle';
import useGetFlows from '@/features/flows/api/useGetFlows';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import useInteractionState from '@/features/flows/hooks/useInteractionState';
import useUIPanelState from '@/features/flows/hooks/useUIPanelState';
import {ResourceTypes} from '@/features/flows/models/resources';
import type {BasicFlowDefinition} from '@/features/flows/models/responses';
import {StepCategories, StepTypes, type Step, type StepData} from '@/features/flows/models/steps';
import '../execution/ExecutionMinimal.scss';
import './Call.scss';

export type CallPropsInterface = CommonStepFactoryPropsInterface;

type CallStepData = StepData & {flow?: {ref?: string}};

const CALL_NODE_WIDTH = 260;

/**
 * Call Node component for cross-flow invocation. It shares the executor card's header but
 * its body carries a distinct "reference" treatment (dashed primary-tinted outline, primary
 * wash, and a workflow glyph) so it does not read as an executor, and it exposes a flow
 * reference instead of an executor along with both `onSuccess` (right) and `onFailure`
 * (bottom) handles. The node card also carries an "open referenced flow" shortcut that jumps
 * the builder to the callee flow (with an unsaved-changes confirm).
 */
function Call({resources, data}: CallPropsInterface): ReactElement {
  const stepId: string | null = useNodeId();
  const {t} = useTranslation();
  const navigate = useNavigate();
  const flowRoutes = useFlowRoutes();
  const {setLastInteractedResource, setLastInteractedStepId} = useInteractionState();
  const {setIsOpenResourcePropertiesPanel} = useUIPanelState();
  const {deleteElements} = useReactFlow();
  const {data: flowsData} = useGetFlows({limit: 100});
  const [isOpenConfirmDialog, setIsOpenConfirmDialog] = useState<boolean>(false);

  const callData: CallStepData = (data as CallStepData) ?? {};
  const flowRef: string = callData.flow?.ref ?? '';
  const paletteEntry: Step | undefined = resources?.[0];
  const displayLabel: string = paletteEntry?.display?.label ?? t('flows:core.call.unconfiguredLabel', 'Flow');

  const referencedFlow: BasicFlowDefinition | undefined = useMemo<BasicFlowDefinition | undefined>(
    () => (flowsData?.flows ?? []).find((f: BasicFlowDefinition) => f.id === flowRef),
    [flowsData, flowRef],
  );

  const canOpen = Boolean(referencedFlow);

  const resource: Step = {
    ...(paletteEntry ?? ({} as Step)),
    id: stepId ?? '',
    type: StepTypes.Call,
    category: StepCategories.Workflow,
    resourceType: ResourceTypes.Step,
    data: callData,
    display: {
      ...(paletteEntry?.display ?? {}),
      label: displayLabel,
      showOnResourcePanel: false,
    },
  } as Step;

  const handleConfigClick = (): void => {
    if (stepId !== null) {
      setLastInteractedStepId(stepId);
    }
    setLastInteractedResource(resource);
    setIsOpenResourcePropertiesPanel(true);
  };

  const handleCardClick = (): void => {
    if (stepId !== null) {
      setLastInteractedStepId(stepId);
    }
    setLastInteractedResource(resource);
  };

  const handleDelete = (): void => {
    if (stepId) {
      // eslint-disable-next-line @typescript-eslint/no-floating-promises
      deleteElements({nodes: [{id: stepId}]});
    }
  };

  const handleOpenReferencedFlow = (event: React.MouseEvent): void => {
    event.stopPropagation();
    if (!referencedFlow) {
      return;
    }
    setIsOpenConfirmDialog(true);
  };

  const handleConfirmOpenReferencedFlow = (): void => {
    setIsOpenConfirmDialog(false);
    if (!referencedFlow) {
      return;
    }
    // eslint-disable-next-line @typescript-eslint/no-floating-promises
    navigate(flowRoutes.flows.detail(referencedFlow.id));
  };

  const bodyLabel: string = referencedFlow
    ? referencedFlow.name
    : t('flows:core.call.selectFlow', 'Select a flow to invoke');

  return (
    <ValidationErrorBoundary resource={resource}>
      <Box
        className="execution-minimal-step has-branching call-step"
        data-testid="call-node"
        sx={{width: CALL_NODE_WIDTH}}
      >
        <Box
          display="flex"
          justifyContent="space-between"
          alignItems="center"
          className="execution-minimal-step-action-panel"
          sx={{backgroundColor: '#151515', height: 44, px: 2, py: 1.25, gap: 1.5}}
        >
          <StepTitle label={displayLabel} />
          <Box display="flex" alignItems="center" gap={0.5}>
            <Tooltip title={t('flows:core.call.tooltip.configure', 'Configure')}>
              <IconButton size="small" onClick={handleConfigClick} sx={{color: 'common.white'}}>
                <CogIcon size={18} />
              </IconButton>
            </Tooltip>
            <Tooltip title={t('flows:core.call.tooltip.delete', 'Delete')}>
              <IconButton size="small" onClick={handleDelete} sx={{color: 'common.white'}}>
                <TrashIcon size={18} />
              </IconButton>
            </Tooltip>
          </Box>
        </Box>
        <Handle type="target" position={Position.Left} />
        <Card
          className="execution-minimal-step-content"
          onClick={handleCardClick}
          sx={{p: 2}}
          data-testid="call-node-content"
        >
          <Box display="flex" alignItems="center" justifyContent="space-between" gap={1}>
            <Box className="call-step-reference-icon">
              <WorkflowIcon size={18} />
            </Box>
            <Box sx={{minWidth: 0, flex: 1}}>
              <Typography variant="caption" className="call-step-reference-label" sx={{display: 'block'}}>
                {t('flows:core.call.referencedFlow', 'Referenced flow')}
              </Typography>
              <Typography
                variant="body2"
                data-testid="call-node-flow-ref"
                sx={{
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                  color: referencedFlow ? 'text.primary' : 'text.secondary',
                  fontStyle: referencedFlow ? 'normal' : 'italic',
                }}
              >
                {bodyLabel}
              </Typography>
            </Box>
            <Tooltip
              title={
                canOpen
                  ? t('flows:core.call.tooltip.openFlow', 'Open referenced flow')
                  : t('flows:core.call.tooltip.openFlowDisabled', 'Configure a referenced flow to enable')
              }
            >
              <span>
                <IconButton
                  size="small"
                  onClick={handleOpenReferencedFlow}
                  disabled={!canOpen}
                  data-testid="call-open-referenced-flow"
                >
                  <ExternalLinkIcon size={16} />
                </IconButton>
              </span>
            </Tooltip>
          </Box>
        </Card>
        <Tooltip title={t('flows:core.call.handles.success', 'On success')} placement="right">
          <Box className="handle-wrapper success-wrapper">
            <Handle
              type="source"
              position={Position.Right}
              id={`${stepId ?? ''}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
              className="execution-handle-success"
            />
          </Box>
        </Tooltip>
        <Tooltip title={t('flows:core.call.handles.failure', 'On failure')} placement="bottom">
          <Box className="handle-wrapper failure-wrapper">
            <Handle type="source" position={Position.Bottom} id="failure" className="execution-handle-failure" />
          </Box>
        </Tooltip>
      </Box>
      <Dialog
        open={isOpenConfirmDialog}
        onClose={() => setIsOpenConfirmDialog(false)}
        maxWidth="sm"
        fullWidth
        data-testid="call-open-referenced-flow-dialog"
      >
        <DialogTitle>{t('flows:core.call.openFlow.dialog.title', 'Open referenced flow?')}</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {t('flows:core.call.openFlow.dialog.description', 'Any unsaved changes to the current flow will be lost.')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setIsOpenConfirmDialog(false)}>
            {t('flows:core.call.openFlow.dialog.cancel', 'Cancel')}
          </Button>
          <Button onClick={handleConfirmOpenReferencedFlow} color="primary" variant="contained">
            {t('flows:core.call.openFlow.dialog.confirm', 'Continue')}
          </Button>
        </DialogActions>
      </Dialog>
    </ValidationErrorBoundary>
  );
}

export default memo(Call);
