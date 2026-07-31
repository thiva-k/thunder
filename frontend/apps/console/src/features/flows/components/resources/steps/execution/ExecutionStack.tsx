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

import {Box, Typography} from '@wso2/oxygen-ui';
import {Handle, Position, useNodeId, type NodeProps} from '@xyflow/react';
import {useContext, useMemo, type CSSProperties, type KeyboardEvent, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import ResourceDisplayImage from '@/features/flows/components/ResourceDisplayImage';
import ValidationErrorBoundary from '@/features/flows/components/validation-panel/ValidationErrorBoundary';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import CompactStacksContext from '@/features/flows/context/CompactStacksContext';
import useValidationStatus from '@/features/flows/hooks/useValidationStatus';
import type {Resource} from '@/features/flows/models/resources';
import type {StepData} from '@/features/flows/models/steps';
import {
  EXECUTION_STACK_MAX_LAYERS,
  type ExecutionStackData,
  type ExecutionStackMember,
} from '@/features/flows/utils/compactGraphTransforms';
import './ExecutionStack.scss';

interface MemberDisplay {
  description?: string;
  image?: string;
  label: string;
  preserveImageColor?: boolean;
}

const resolveMemberDisplay = (member: ExecutionStackMember): MemberDisplay => {
  const data = member.data as StepData | undefined;
  const display = data?.display as MemberDisplay | undefined;
  const executorName = (data?.action as {executor?: {name?: string}} | undefined)?.executor?.name;
  return {
    description: display?.description,
    image: display?.image,
    label: display?.label ?? executorName ?? 'Executor',
    preserveImageColor: display?.preserveImageColor,
  };
};

/**
 * Execution (Stack) Node component. Rendered in compact (non-verbose) mode in
 * place of a run of consecutive non-branching executors: one chip showing the
 * first executor's icon with a "+N" count inside, a card-deck effect behind,
 * and a hover preview fanning the stacked members out to the right. Clicking
 * expands the stack into its individual member chips. A single in/out edge
 * anchor pair is kept.
 *
 * @param props - Props injected to the component.
 * @returns Execution (Stack) node component.
 */
function ExecutionStack({data}: NodeProps): ReactElement {
  const {expandStack} = useContext(CompactStacksContext);
  const {notifications} = useValidationStatus();
  const stackId: string | null = useNodeId();
  const {t} = useTranslation();

  const stackData = data as ExecutionStackData | undefined;
  const members = useMemo(() => stackData?.members ?? [], [stackData]);

  // Collapsing hides the members' own error boundaries, so the stack surfaces
  // the first member that has a notification. Expanding it then shows exactly
  // which executor is at fault.
  const notifiedMemberId = useMemo(
    () => members.find((member) => notifications.some((notification) => notification.hasResource(member.id)))?.id,
    [members, notifications],
  );
  const headMember = members.at(0);
  const headDisplay = headMember ? resolveMemberDisplay(headMember) : undefined;
  const stackedCount = Math.max(members.length - 1, 0);
  const layerCount = Math.min(stackedCount, EXECUTION_STACK_MAX_LAYERS);

  const handleExpand = (): void => {
    if (stackId) {
      expandStack(
        stackId,
        members.map((member) => member.id),
      );
    }
  };

  return (
    <ValidationErrorBoundary resource={{id: notifiedMemberId ?? stackId ?? ''} as Resource}>
      <Box className="execution-stack-step">
        <Box
          className="execution-stack-step-content"
          role="button"
          tabIndex={0}
          aria-label={members.map((member) => resolveMemberDisplay(member).label).join(', ')}
          aria-expanded={false}
          title={t('flows:core.executions.stack.expandHint', 'Click to expand')}
          onClick={handleExpand}
          // A div with role="button" gets no native Enter/Space activation, so
          // the stack would be focusable but impossible to expand by keyboard.
          onKeyDown={(event: KeyboardEvent<HTMLElement>) => {
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault();
              handleExpand();
            }
          }}
        >
          {Array.from({length: layerCount}, (_, index) => (
            <Box key={index} className={`execution-stack-layer execution-stack-layer-${index + 1}`} />
          ))}
          {/* Hover preview: the stacked members slide out of the deck to the
            right. The main chip itself stays put; only its color changes. */}
          {members.slice(1).map((member, index) => {
            const display = resolveMemberDisplay(member);
            return (
              <Box key={member.id} className="execution-stack-fan-chip" style={{'--fan-index': index} as CSSProperties}>
                {display.image ? (
                  <ResourceDisplayImage
                    image={display.image}
                    label={display.label}
                    size={16}
                    preserveColor={display.preserveImageColor}
                  />
                ) : (
                  <Typography variant="caption" className="execution-stack-chip-fallback">
                    {display.label.charAt(0).toUpperCase()}
                  </Typography>
                )}
              </Box>
            );
          })}
          <Box className="execution-stack-chip">
            {headDisplay?.image ? (
              <ResourceDisplayImage
                image={headDisplay.image}
                label={headDisplay.label}
                size={stackedCount > 0 ? 18 : 24}
                preserveColor={headDisplay.preserveImageColor}
              />
            ) : (
              <Typography variant="subtitle1" className="execution-stack-chip-fallback">
                {(headDisplay?.label ?? 'Executor').charAt(0).toUpperCase()}
              </Typography>
            )}
            {stackedCount > 0 && (
              <Typography variant="caption" className="execution-stack-badge" data-testid="execution-stack-badge">
                +{stackedCount}
              </Typography>
            )}
          </Box>
        </Box>
        <Handle type="target" position={Position.Left} />
        <Handle
          type="source"
          position={Position.Right}
          id={`${stackId ?? ''}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
        />
      </Box>
    </ValidationErrorBoundary>
  );
}

export default ExecutionStack;
