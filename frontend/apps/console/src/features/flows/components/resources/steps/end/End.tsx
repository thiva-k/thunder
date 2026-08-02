// Copyright 2023-2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Fab} from '@wso2/oxygen-ui';
import {Handle, Position} from '@xyflow/react';
import type {ReactElement} from 'react';
import type {CommonStepFactoryPropsInterface} from '../CommonStepFactory';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import './End.scss';

/**
 * Props interface of {@link End}
 */
export type EndPropsInterface = CommonStepFactoryPropsInterface;

/**
 * End Node component.
 * This is a custom node supported by react flow renderer library.
 * See {@link https://reactflow.dev/docs/api/node-types/} for its documentation
 * and {@link https://reactflow.dev/examples/custom-node/} for an example
 *
 * @param _props - Props injected to the component (unused).
 * @returns End node component.
 */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function End(_props: EndPropsInterface): ReactElement {
  return (
    <div>
      <Handle
        className="hidden-handle"
        id={`end${VisualFlowConstants.FLOW_BUILDER_PREVIOUS_HANDLE_SUFFIX}`}
        type="target"
        position={Position.Left}
      />
      <Fab aria-label="end" className="end" variant="extended" size="small">
        End
      </Fab>
    </div>
  );
}

export default End;
