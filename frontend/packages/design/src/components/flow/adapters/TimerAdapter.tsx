// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {FlowTimer, type FlowTimerRenderProps} from '@thunderid/react';
import {cn} from '@thunderid/utils';
import {Alert, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';

/**
 * Props for the TimerAdapter component.
 */
interface TimerAdapterProps {
  /** Duration in seconds until the step expires */
  expiresIn: number;
  /** Text template with {time} placeholder, resolved from the component label */
  textTemplate?: string;
}

/**
 * Oxygen-UI styled timer adapter.
 *
 * Uses the SDK's `FlowTimer` render-prop component to manage
 * the countdown, then renders oxygen-ui styled text.
 */
export default function TimerAdapter({
  expiresIn,
  textTemplate = 'Time remaining: {time}',
}: TimerAdapterProps): JSX.Element {
  return (
    <FlowTimer expiresIn={expiresIn}>
      {({isExpired, formattedTime}: FlowTimerRenderProps) =>
        isExpired ? (
          <Alert className={cn('Flow--timer', 'Alert--root')} severity="warning" sx={{mt: 1}}>
            <Typography className={cn('Text--body2')} variant="body2">
              {formattedTime}
            </Typography>
          </Alert>
        ) : (
          <Typography className={cn('Flow--timer', 'Text--body2')} variant="body2" color="warning.main" sx={{mt: 1}}>
            {textTemplate.replace('{time}', formattedTime)}
          </Typography>
        )
      }
    </FlowTimer>
  );
}
