// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import {getErrorMessage} from '@thunderid/utils';

/**
 * The subset of the APP-1039 (flow mismatch) API error response this util reads.
 */
interface FlowMismatchApiError {
  code: string;
  description?: {
    params?: {
      sourceFlowType?: string;
      flowType?: string;
    };
  };
}

const FLOW_MISMATCH_ERROR_CODE = 'APP-1039';

/**
 * Maps the raw flow-type values the backend sends in APP-1039's params
 * to the i18n keys already used for the same flows on the application's Flows tab.
 */
const FLOW_TYPE_LABEL_KEYS: Record<string, string> = {
  authentication: 'edit.flows.labels.authFlow',
  registration: 'edit.flows.labels.registrationFlow',
  recovery: 'edit.flows.labels.recoveryFlow',
  signout: 'edit.flows.labels.signOutFlow',
};

/**
 * Extracts a localized error message from an application API error response, with a dedicated,
 * actionable message for APP-1039 (conflicting flow references).
 *
 * For APP-1039, builds the message from the error's `sourceFlowType`/`flowType` params,
 * translated into the same flow labels shown on the application's Flows tab (e.g. "Sign-up Flow",
 * "Recovery Flow"), instead of using the backend's generic description text. Falls back to
 * {@link getErrorMessage} for every other error, and for APP-1039 responses without usable params.
 *
 * @param error - The error thrown by the mutation
 * @param t - The i18next translation function scoped to the relevant namespace
 * @param fallbackKey - i18n key to use when no specific message is found (e.g. `'create.error'`)
 * @returns Localized error message string
 *
 * @public
 */
export default function getApplicationErrorMessage(
  error: Error,
  t: (key: string, options?: Record<string, unknown>) => string,
  fallbackKey: string,
): string {
  const apiError = (error as {response?: {data?: FlowMismatchApiError}}).response?.data;

  if (apiError?.code === FLOW_MISMATCH_ERROR_CODE) {
    const sourceFlowType = apiError.description?.params?.sourceFlowType?.toLowerCase();
    const flowType = apiError.description?.params?.flowType?.toLowerCase();
    const sourceLabelKey = sourceFlowType ? FLOW_TYPE_LABEL_KEYS[sourceFlowType] : undefined;
    const flowLabelKey = flowType ? FLOW_TYPE_LABEL_KEYS[flowType] : undefined;

    if (sourceLabelKey && flowLabelKey) {
      return t(`errors.${FLOW_MISMATCH_ERROR_CODE}`, {
        sourceFlowType: t(sourceLabelKey),
        flowType: t(flowLabelKey),
      });
    }

    // Skip the generic code-based lookup: it would resolve the same errors.APP-1039 key and
    // return its {{sourceFlowType}}/{{flowType}} placeholders unresolved.
    return t(fallbackKey);
  }

  return getErrorMessage(error, t, fallbackKey);
}
