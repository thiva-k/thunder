// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {useGetUserTypeAttributes} from '@thunderid/configure-user-types';
import {
  Alert,
  FormControlLabel,
  FormHelperText,
  FormLabel,
  Radio,
  RadioGroup,
  Stack,
  Typography,
} from '@wso2/oxygen-ui';
import {useContext, useEffect, useMemo, useRef, useState, type ReactNode} from 'react';
import {useTranslation} from 'react-i18next';
import type {CommonResourcePropertiesPropsInterface} from './types';
import FlowTypeContext from '../../../../context/FlowTypeContext';
import {FlowType} from '../../../../models/flows';
import type {FlowNodeInput} from '../../../../models/responses';
import {
  getAttributeVerificationTarget,
  withAttributeVerificationTarget,
  type AttributeVerificationTarget,
} from '../../../../utils/attributeVerification';
import {getExecutorInputs} from '../../../../utils/executorInputs';
import UserAttributeAutocomplete from '../UserAttributeAutocomplete';

/**
 * The choices of what a prepare step verifies.
 */
const TargetChoice = {
  USER_ATTRIBUTE: 'userAttribute',
  AGENT_OWNER: 'agentOwner',
} as const;

/**
 * Properties of a Prepare Verification step: whether it verifies a user's email address or mobile number, and
 * which attribute, or an agent's owner. The choice is stored as the executor's `attribute` input, so a flow saved
 * here runs as one written by hand does.
 *
 * @param props - Props injected to the component.
 * @returns The AttributeVerificationProperties component.
 */
function AttributeVerificationProperties({resource, onChange}: CommonResourcePropertiesPropsInterface): ReactNode {
  const {t} = useTranslation();
  const flowType = useContext(FlowTypeContext);

  const inputs: FlowNodeInput[] = useMemo(() => (resource ? getExecutorInputs(resource) : []), [resource]);
  const target = getAttributeVerificationTarget(inputs);
  const isAgentOwner = target.kind === 'agentOwner';
  const attribute = target.kind === 'userAttribute' ? target.attribute : '';

  // The user type is not known while the flow is authored, so attributes from every user type are suggested.
  const {attributes: allAttributes, isLoading: isLoadingAttributes} = useGetUserTypeAttributes({
    enabled: !isAgentOwner,
  });
  const attributeOptions = useMemo(() => allAttributes.filter((option) => !option.credential), [allAttributes]);

  const commit = (next: AttributeVerificationTarget): void => {
    const nextInputs = withAttributeVerificationTarget(inputs, next);
    onChange('data.action.executor.inputs', nextInputs.length > 0 ? nextInputs : undefined, resource);
  };

  // Typed text stays local until blur or Enter, like DraftTextField.
  const [attributeInput, setAttributeInput] = useState<string>(attribute);
  const [syncedAttribute, setSyncedAttribute] = useState<string>(attribute);
  if (attribute !== syncedAttribute) {
    setSyncedAttribute(attribute);
    setAttributeInput(attribute);
  }

  const commitAttribute = (value: string): void => {
    const trimmed = value.trim();
    setAttributeInput(trimmed);
    if (trimmed === attribute) {
      return;
    }
    commit(trimmed ? {kind: 'userAttribute', attribute: trimmed} : {kind: 'requested'});
  };

  // Removing a focused input does not reliably fire blur, so a typed attribute is committed when the panel
  // unmounts, as DraftTextField does.
  const pendingCommitRef = useRef<(() => void) | null>(null);
  useEffect(() => {
    pendingCommitRef.current = (): void => {
      const trimmed = attributeInput.trim();
      if (!isAgentOwner && trimmed !== attribute) {
        commit(trimmed ? {kind: 'userAttribute', attribute: trimmed} : {kind: 'requested'});
      }
    };
  });
  useEffect(
    () => () => {
      pendingCommitRef.current?.();
    },
    [],
  );

  const handleTargetChange = (value: string): void => {
    if (value === TargetChoice.AGENT_OWNER) {
      commit({kind: 'agentOwner'});
      return;
    }
    commit({kind: 'requested'});
  };

  return (
    <Stack gap={2}>
      <Typography variant="body2" color="text.secondary">
        {t(
          'flows:core.executions.attributeVerification.description',
          'Choose whose email address or mobile number this step verifies. The flow then sends it a code or a link, and a Record Verification step marks it verified.',
        )}
      </Typography>

      <div>
        <FormLabel id="attribute-verification-target-label">
          {t('flows:core.executions.attributeVerification.target.label', 'Whose contact to verify')}
        </FormLabel>
        <RadioGroup
          aria-labelledby="attribute-verification-target-label"
          value={isAgentOwner ? TargetChoice.AGENT_OWNER : TargetChoice.USER_ATTRIBUTE}
          onChange={(event) => handleTargetChange(event.target.value)}
        >
          <FormControlLabel
            value={TargetChoice.USER_ATTRIBUTE}
            control={<Radio size="small" />}
            label={t('flows:core.executions.attributeVerification.target.userAttribute', 'User')}
          />
          <FormControlLabel
            value={TargetChoice.AGENT_OWNER}
            control={<Radio size="small" />}
            label={t('flows:core.executions.attributeVerification.target.agentOwner', 'Agent owner')}
          />
        </RadioGroup>
      </div>

      {isAgentOwner ? (
        <>
          <Typography variant="body2" color="text.secondary">
            {t(
              'flows:core.executions.attributeVerification.agentOwner.hint',
              'The email address or mobile number that names the external owner of the agent the request names. The owner needs no account.',
            )}
          </Typography>
          {flowType && flowType !== FlowType.ADMIN_INITIATED && (
            <Alert severity="warning" data-testid="attribute-verification-owner-flow-type">
              {t(
                'flows:core.executions.attributeVerification.agentOwner.flowTypeWarning',
                "Only an administrator-initiated flow can verify an agent owner's contact, so this flow cannot be saved with this choice.",
              )}
            </Alert>
          )}
        </>
      ) : (
        <div>
          <UserAttributeAutocomplete
            id="attribute-verification-attribute"
            label={t('flows:core.executions.attributeVerification.attribute.label', 'Attribute to verify')}
            placeholder={t('flows:core.executions.attributeVerification.attribute.placeholder', 'email')}
            size="small"
            options={attributeOptions}
            loading={isLoadingAttributes}
            inputValue={attributeInput}
            onChange={commitAttribute}
            onInputChange={(value: string, reason: string) => {
              if (reason === 'input') {
                setAttributeInput(value);
              }
              if (reason === 'clear') {
                commitAttribute('');
              }
            }}
            onBlur={() => commitAttribute(attributeInput)}
          />
          <FormHelperText>
            {t(
              'flows:core.executions.attributeVerification.attribute.hint',
              'An email address or mobile number attribute of the user the request names. The request can name another. Leave it empty to have the request always name one.',
            )}
          </FormHelperText>
        </div>
      )}
    </Stack>
  );
}

export default AttributeVerificationProperties;
