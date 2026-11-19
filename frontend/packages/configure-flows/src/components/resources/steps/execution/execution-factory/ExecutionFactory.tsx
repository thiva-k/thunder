// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {Box, Typography} from '@wso2/oxygen-ui';
import type {ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import {ExecutionTypes, type StepData} from '../../../../../models/steps';
import {getAttributeVerificationTarget} from '../../../../../utils/attributeVerification';
import {getExecutorInputs} from '../../../../../utils/executorInputs';
import ResourceDisplayImage from '../../../../ResourceDisplayImage';
import type {ExecutionMinimalPropsInterface} from '../ExecutionMinimal';

/**
 * Props interface of {@link CommonStepFactory}
 */
export type ExecutionFactoryPropsInterface = ExecutionMinimalPropsInterface;

/**
 * Factory for creating execution types.
 *
 * @param props - Props injected to the component.
 * @returns The ExecutionFactory component.
 */
function ExecutionFactory({resource}: ExecutionFactoryPropsInterface): ReactElement | null {
  const {t} = useTranslation();

  // Display metadata is resolved from the same executor definitions that back the
  // resource panel listing, so the node always mirrors the panel entry.
  const displayImage = resource.display?.image;
  // display.label contains the action/mode (e.g., "Passkey Challenge", "Send SMS OTP")
  const displayLabel = resource.display?.label;

  // A Prepare Verification step says on the canvas what it verifies, so a flow for an agent's owner is told apart
  // from one for a user's own email address or mobile number without opening the step.
  const executor = (resource.data as StepData | undefined)?.action?.executor;
  let summary: string | undefined;
  if (executor?.name === ExecutionTypes.AttributeVerification && executor?.['mode'] === 'prepare') {
    const target = getAttributeVerificationTarget(getExecutorInputs(resource));
    if (target.kind === 'agentOwner') {
      summary = t(
        'flows:core.executions.attributeVerification.summary.agentOwner',
        "Verifies the agent owner's contact",
      );
    } else if (target.kind === 'userAttribute') {
      summary = t(
        'flows:core.executions.attributeVerification.summary.userAttribute',
        "Verifies the user's {{attribute}}",
        {
          attribute: target.attribute,
        },
      );
    } else {
      summary = t(
        'flows:core.executions.attributeVerification.summary.requested',
        "Verifies the user's contact the request names",
      );
    }
  }

  return (
    <Box>
      <Box display="flex" gap={1} alignItems="center">
        <ResourceDisplayImage
          image={displayImage}
          label={displayLabel}
          preserveColor={resource.display?.preserveImageColor}
        />
        <Typography variant="body1">{displayLabel ?? t('flows:core.executions.names.default', 'Executor')}</Typography>
      </Box>
      {summary && (
        <Typography
          variant="caption"
          color="text.secondary"
          sx={{display: 'block', mt: 0.5}}
          data-testid="attribute-verification-summary"
        >
          {summary}
        </Typography>
      )}
    </Box>
  );
}

export default ExecutionFactory;
