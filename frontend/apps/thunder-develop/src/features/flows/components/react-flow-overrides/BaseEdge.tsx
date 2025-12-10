/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import {BaseEdge as XYFlowBaseEdge, EdgeLabelRenderer, useReactFlow, useNodes, type EdgeProps} from '@xyflow/react';
import {useState, type ReactElement, type SyntheticEvent} from 'react';
import {XIcon} from '@wso2/oxygen-ui-icons-react';
import {calculateEdgePath} from '../../utils/calculateEdgePath';
import './BaseEdge.scss';

/**
 * Props interface of {@link BaseEdge}
 */
export type BaseEdgePropsInterface = EdgeProps;

/**
 * Enhanced edge component with custom routing algorithm to avoid nodes.
 * Includes custom delete button and label functionality with hover effects.
 */
function BaseEdge({
  id,
  source,
  target,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  label,
  style,
  deletable,
  ...rest
}: BaseEdgePropsInterface): ReactElement {
  const {deleteElements} = useReactFlow();
  const [isHovered, setIsHovered] = useState<boolean>(false);
  const nodes = useNodes();

  // Calculate smart path that routes around nodes
  const {
    path: edgePath,
    centerX: labelX,
    centerY: labelY,
  } = calculateEdgePath(sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, nodes);

  const handleDelete = (event: SyntheticEvent) => {
    event.stopPropagation();
    deleteElements({edges: [{id}]}).catch(() => {});
  };

  const handleDeleteKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      event.stopPropagation();
      deleteElements({edges: [{id}]}).catch(() => {});
    }
  };

  return (
    <g
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Invisible wider path for hover detection */}
      <path
        d={edgePath}
        fill="none"
        stroke="transparent"
        strokeWidth={20}
        style={{cursor: 'pointer'}}
      />
      <XYFlowBaseEdge
        id={id}
        path={edgePath}
        style={{
          ...style,
          strokeWidth: isHovered ? 3 : 2,
          transition: 'stroke-width 0.2s ease',
        }}
        interactionWidth={20}
        markerEnd={rest.markerEnd}
        markerStart={rest.markerStart}
      />
      <EdgeLabelRenderer>
        {label && (
          <div
            style={{'--label-x': `${labelX}px`, '--label-y': `${labelY}px`} as React.CSSProperties}
            className="edge-label-renderer__deletable-edge nodrag nopan"
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
          >
            {label}
          </div>
        )}
        {isHovered && deletable !== false && (
          <div
            className="edge-delete-button nodrag nopan"
            onClick={handleDelete}
            onKeyDown={handleDeleteKeyDown}
            role="button"
            tabIndex={0}
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
            style={{'--label-x': `${labelX}px`, '--label-y': `${labelY}px`} as React.CSSProperties}
          >
            <XIcon size={16} className="edge-delete-button__icon" />
          </div>
        )}
      </EdgeLabelRenderer>
    </g>
  );
}

export default BaseEdge;
