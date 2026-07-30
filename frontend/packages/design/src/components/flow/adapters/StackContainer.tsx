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

import {cn} from '@thunderid/utils';
import {Box, Stack} from '@wso2/oxygen-ui';
import type {JSX, ReactNode} from 'react';
import type {FlowComponent} from '../../../models/flow';
import getStackGridSx from '../../../utils/getStackGridSx';

interface StackContainerProps {
  component: FlowComponent;
  children: ReactNode;
}

/**
 * Layout container for STACK flow elements. Renders a CSS grid when the stack sets
 * `items` slots across its main axis, and keeps the flex stack layout otherwise.
 */
export default function StackContainer({component, children}: StackContainerProps): JSX.Element {
  const className = [cn('Flow--stack'), component.classes].filter(Boolean).join(' ');
  const gridSx = getStackGridSx(component);

  if (gridSx) {
    return (
      <Box id={component.id} className={className} sx={gridSx}>
        {children}
      </Box>
    );
  }

  return (
    <Stack
      id={component.id}
      className={className}
      direction={component.direction ?? 'column'}
      spacing={component.gap ?? 2}
      alignItems={component.align ?? 'center'}
      justifyContent={component.justify ?? 'flex-start'}
      // Wrap like the flow builder canvas and the SDK do, instead of squeezing
      // children onto one line. `useFlexGap` keeps the spacing gap based, which is
      // the wrap-safe mode.
      useFlexGap
      sx={{flexWrap: 'wrap'}}
    >
      {children}
    </Stack>
  );
}
