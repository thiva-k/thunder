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

import {Box, IconButton, Stack, Tooltip, Typography} from '@wso2/oxygen-ui';
import {Layers} from '@wso2/oxygen-ui-icons-react';
import {Handle, Position, useNodeId} from '@xyflow/react';
import {useContext, type KeyboardEvent, type MouseEvent, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import ResourceDisplayImage from '@/features/flows/components/ResourceDisplayImage';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import CompactStacksContext from '@/features/flows/context/CompactStacksContext';
import useInteractionState from '@/features/flows/hooks/useInteractionState';
import useUIPanelState from '@/features/flows/hooks/useUIPanelState';
import type {Step, StepData} from '@/features/flows/models/steps';
import './ExecutionCompact.scss';

/**
 * Props interface of {@link ExecutionCompact}
 */
export interface ExecutionCompactPropsInterface {
  /**
   * Resource object of the execution step.
   */
  resource: Step;
}

/**
 * Execution (Compact) Node component. Rendered in place of the full executor
 * card when the canvas is in compact (non-verbose) mode: an icon-only chip
 * that keeps the executor's branching handles on the canvas.
 *
 * @param props - Props injected to the component.
 * @returns Execution (Compact) node component.
 */
function ExecutionCompact({resource}: ExecutionCompactPropsInterface): ReactElement {
  const {setLastInteractedResource, setLastInteractedStepId} = useInteractionState();
  const {setIsOpenResourcePropertiesPanel} = useUIPanelState();
  const {collapseStack, expandedHeadIdToStackId} = useContext(CompactStacksContext);
  const stepId: string | null = useNodeId();

  const {t} = useTranslation();

  // The head chip of an expanded group offers restacking it.
  const restackableStackId = stepId ? expandedHeadIdToStackId.get(stepId) : undefined;

  const displayLabel = resource.display?.label ?? resource.data?.action?.executor?.name ?? 'Executor';
  const displayDescription = resource.display?.description;

  // Same branching detection as ExecutionMinimal: the presence of the
  // onFailure/onIncomplete fields (even empty) marks handle support.
  const stepData = resource.data as StepData | undefined;
  const hasBranchingSupport = stepData?.action && 'onFailure' in stepData.action;
  const hasIncompleteSupport = stepData?.action && 'onIncomplete' in stepData.action;

  const outcomeLabels = resource.display?.outcomes;
  const successLabel = outcomeLabels?.success ?? t('flows:core.executions.handles.success', 'Success');
  const failureLabel = outcomeLabels?.failure ?? t('flows:core.executions.handles.failure', 'Failure');
  const incompleteLabel = outcomeLabels?.incomplete ?? t('flows:core.executions.handles.incomplete', 'Incomplete');

  const handleOpenProperties = (): void => {
    if (stepId !== null) {
      setLastInteractedStepId(stepId);
    }
    setLastInteractedResource({
      ...resource,
      config: {
        ...(resource?.config || {}),
        ...(typeof resource.data?.config === 'object' && resource.data?.config !== null ? resource.data.config : {}),
      },
    });
    setIsOpenResourcePropertiesPanel(true);
  };

  return (
    <Box className="execution-compact-step">
      <Tooltip
        placement="top"
        title={
          <Stack>
            <Typography variant="subtitle2">{displayLabel}</Typography>
            {displayDescription && <Typography variant="caption">{displayDescription}</Typography>}
          </Stack>
        }
      >
        <Box
          className="execution-compact-step-content"
          role="button"
          tabIndex={0}
          aria-label={displayLabel}
          onClick={handleOpenProperties}
          // A div with role="button" gets no native Enter/Space activation, and
          // in compact mode this chip is the only way into the executor.
          onKeyDown={(event: KeyboardEvent<HTMLElement>) => {
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault();
              handleOpenProperties();
            }
          }}
        >
          {resource.display?.image ? (
            <ResourceDisplayImage
              image={resource.display.image}
              label={displayLabel}
              size={24}
              preserveColor={resource.display?.preserveImageColor}
            />
          ) : (
            <Typography variant="subtitle1" className="execution-compact-step-fallback">
              {displayLabel.charAt(0).toUpperCase()}
            </Typography>
          )}
        </Box>
      </Tooltip>
      {restackableStackId && (
        <Tooltip title={t('flows:core.executions.stack.restack', 'Restack executors')} placement="top">
          <IconButton
            size="small"
            className="execution-compact-step-restack"
            aria-label={t('flows:core.executions.stack.restack', 'Restack executors')}
            onClick={(event: MouseEvent<HTMLElement>) => {
              event.stopPropagation();
              collapseStack(restackableStackId);
            }}
          >
            <Layers size={12} />
          </IconButton>
        </Tooltip>
      )}
      <Handle type="target" position={Position.Left} />
      {/* Success handle - always shown on the right */}
      {hasBranchingSupport ? (
        <Tooltip title={successLabel} placement="right">
          <Box className="handle-wrapper success-wrapper">
            <Handle
              type="source"
              position={Position.Right}
              id={`${resource.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
              className="execution-handle-success"
            />
          </Box>
        </Tooltip>
      ) : (
        <Handle
          type="source"
          position={Position.Right}
          id={`${resource.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
        />
      )}
      {/* Failure handle - shown at the bottom when the action supports branching (has onFailure property) */}
      {hasBranchingSupport && (
        <Tooltip title={failureLabel} placement="bottom">
          <Box className="handle-wrapper failure-wrapper">
            <Handle type="source" position={Position.Bottom} id="failure" className="execution-handle-failure" />
          </Box>
        </Tooltip>
      )}
      {/* Incomplete handle - shown at the top when the action supports incomplete (has onIncomplete property) */}
      {hasIncompleteSupport && (
        <Tooltip title={incompleteLabel} placement="top">
          <Box className="handle-wrapper incomplete-wrapper">
            <Handle
              type="source"
              position={Position.Top}
              id={`${resource.id}${VisualFlowConstants.FLOW_BUILDER_INCOMPLETE_HANDLE_SUFFIX}`}
              className="execution-handle-incomplete"
            />
          </Box>
        </Tooltip>
      )}
    </Box>
  );
}

export default ExecutionCompact;
