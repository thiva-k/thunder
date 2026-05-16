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

/**
 * Styles for the Alert primitive component.
 *
 * BEM block: `.thunder-alert`
 *
 * Modifiers:
 *   Severity: --info | --success | --warning | --error
 *
 * Elements:
 *   __content | __dismiss
 */
const ALERT_CSS = `
/* ============================================================
   Alert
   ============================================================ */

.thunder-alert {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: calc(var(--thunder-spacing-unit) * 1);
  padding: var(--thunder-alert-paddingY) var(--thunder-alert-paddingX);
  border-radius: var(--thunder-alert-borderRadius);
  border: 1px solid transparent;
  font-family: var(--thunder-typography-fontFamily);
  font-size: var(--thunder-typography-fontSize-sm);
  box-sizing: border-box;
  width: 100%;
  line-height: var(--thunder-typography-lineHeight-normal);
}

.thunder-alert__content {
  flex: 1;
}

.thunder-alert--info {
  background-color: var(--thunder-color-info-light);
  border-color: var(--thunder-color-info-main);
  color: var(--thunder-color-info-contrastText);
}

.thunder-alert--success {
  background-color: var(--thunder-color-success-light);
  border-color: var(--thunder-color-success-main);
  color: var(--thunder-color-success-contrastText);
}

.thunder-alert--warning {
  background-color: var(--thunder-color-warning-light);
  border-color: var(--thunder-color-warning-main);
  color: var(--thunder-color-warning-contrastText);
}

.thunder-alert--error {
  background-color: var(--thunder-color-error-light);
  border-color: var(--thunder-color-error-main);
  color: var(--thunder-color-error-contrastText);
}

.thunder-alert__dismiss {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1em;
  line-height: 0;
  padding: calc(var(--thunder-spacing-unit) * 0.25);
  border-radius: var(--thunder-border-radius-xs);
  color: inherit;
  opacity: 0.6;
  flex-shrink: 0;
  transition: opacity var(--thunder-transition-fast), background-color var(--thunder-transition-fast);
}
.thunder-alert__dismiss:hover {
  opacity: 1;
  background-color: var(--thunder-color-action-hover);
}
`;

export default ALERT_CSS;
