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
 * Styles for the Button primitive component.
 *
 * BEM block: `.thunder-button`
 *
 * Modifiers:
 *   Variant:  --solid | --outline | --ghost | --text
 *   Color:    --primary | --secondary | --danger
 *   Size:     --small | --medium | --large
 *   State:    --full-width | --loading
 *
 * Elements:
 *   __start-icon | __end-icon | __content | __spinner
 *
 * Note: The `thunder-spin` keyframe animation is defined in
 * `styles/animations.css.ts` and shared with the Spinner component.
 */
const BUTTON_CSS = `
/* ============================================================
   Button
   ============================================================ */

.thunder-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: calc(var(--thunder-spacing-unit) * 0.75);
  border-radius: var(--thunder-button-borderRadius);
  font-family: var(--thunder-typography-fontFamily);
  font-weight: var(--thunder-button-fontWeight);
  letter-spacing: var(--thunder-typography-letterSpacing-normal);
  cursor: pointer;
  outline: none;
  text-decoration: none;
  white-space: nowrap;
  border-width: 1px;
  border-style: solid;
  box-sizing: border-box;
  transition:
    background-color var(--thunder-transition-fast),
    color var(--thunder-transition-fast),
    border-color var(--thunder-transition-fast),
    box-shadow var(--thunder-transition-fast),
    opacity var(--thunder-transition-fast),
    transform var(--thunder-transition-fast);
  position: relative;
  vertical-align: middle;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  user-select: none;
}

.thunder-button:focus-visible {
  outline: none;
  box-shadow: 0 0 0 var(--thunder-focus-ring-width) var(--thunder-focus-ring-color);
}

/* -- Sizes -- */

.thunder-button--small {
  padding: 0 var(--thunder-button-sm-paddingX);
  font-size: var(--thunder-button-sm-fontSize);
  height: var(--thunder-button-sm-height);
}

.thunder-button--medium {
  padding: 0 var(--thunder-button-md-paddingX);
  font-size: var(--thunder-button-md-fontSize);
  height: var(--thunder-button-md-height);
}

.thunder-button--large {
  padding: 0 var(--thunder-button-lg-paddingX);
  font-size: var(--thunder-button-lg-fontSize);
  height: var(--thunder-button-lg-height);
}

/* -- Modifiers -- */

.thunder-button--full-width {
  width: 100%;
}

.thunder-button--loading,
.thunder-button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
  pointer-events: none;
}

/* -- Solid variants -- */

.thunder-button--solid.thunder-button--primary {
  background-color: var(--thunder-color-primary-main);
  color: var(--thunder-color-primary-contrastText);
  border-color: var(--thunder-color-primary-main);
}
.thunder-button--solid.thunder-button--primary:hover:not(:disabled) {
  background-color: var(--thunder-color-primary-dark);
  border-color: var(--thunder-color-primary-dark);
}
.thunder-button--solid.thunder-button--primary:active:not(:disabled) {
  transform: scale(0.98);
}

.thunder-button--solid.thunder-button--secondary {
  background-color: var(--thunder-color-secondary-light);
  color: var(--thunder-color-secondary-main);
  border-color: var(--thunder-color-border);
}
.thunder-button--solid.thunder-button--secondary:hover:not(:disabled) {
  background-color: var(--thunder-color-border);
  border-color: var(--thunder-color-border);
}
.thunder-button--solid.thunder-button--secondary:active:not(:disabled) {
  transform: scale(0.98);
}

.thunder-button--solid.thunder-button--danger {
  background-color: var(--thunder-color-error-main);
  color: #ffffff;
  border-color: var(--thunder-color-error-main);
}
.thunder-button--solid.thunder-button--danger:hover:not(:disabled) {
  filter: brightness(0.92);
}
.thunder-button--solid.thunder-button--danger:active:not(:disabled) {
  transform: scale(0.98);
}

/* -- Outline variants -- */

.thunder-button--outline.thunder-button--primary {
  background-color: transparent;
  color: var(--thunder-color-primary-main);
  border-color: var(--thunder-color-primary-main);
}
.thunder-button--outline.thunder-button--primary:hover:not(:disabled) {
  background-color: var(--thunder-color-primary-light);
}
.thunder-button--outline.thunder-button--primary:active:not(:disabled) {
  transform: scale(0.98);
}

.thunder-button--outline.thunder-button--secondary {
  background-color: transparent;
  color: var(--thunder-color-secondary-main);
  border-color: var(--thunder-color-border);
}
.thunder-button--outline.thunder-button--secondary:hover:not(:disabled) {
  background-color: var(--thunder-color-secondary-light);
  border-color: var(--thunder-color-secondary-main);
}
.thunder-button--outline.thunder-button--secondary:active:not(:disabled) {
  transform: scale(0.98);
}

.thunder-button--outline.thunder-button--danger {
  background-color: transparent;
  color: var(--thunder-color-error-main);
  border-color: var(--thunder-color-error-main);
}
.thunder-button--outline.thunder-button--danger:hover:not(:disabled) {
  background-color: var(--thunder-color-error-light);
}
.thunder-button--outline.thunder-button--danger:active:not(:disabled) {
  transform: scale(0.98);
}

/* -- Ghost variants -- */

.thunder-button--ghost.thunder-button--primary {
  background-color: transparent;
  color: var(--thunder-color-primary-main);
  border-color: transparent;
}
.thunder-button--ghost.thunder-button--primary:hover:not(:disabled) {
  background-color: var(--thunder-color-primary-light);
  border-color: transparent;
}

.thunder-button--ghost.thunder-button--secondary {
  background-color: transparent;
  color: var(--thunder-color-secondary-main);
  border-color: transparent;
}
.thunder-button--ghost.thunder-button--secondary:hover:not(:disabled) {
  background-color: var(--thunder-color-action-hover);
  border-color: transparent;
}

.thunder-button--ghost.thunder-button--danger {
  background-color: transparent;
  color: var(--thunder-color-error-main);
  border-color: transparent;
}
.thunder-button--ghost.thunder-button--danger:hover:not(:disabled) {
  background-color: var(--thunder-color-error-light);
  border-color: transparent;
}

/* -- Text variants -- */

.thunder-button--text {
  border-color: transparent;
  background-color: transparent;
  padding-left: calc(var(--thunder-spacing-unit) * 0.25);
  padding-right: calc(var(--thunder-spacing-unit) * 0.25);
}

.thunder-button--text.thunder-button--primary {
  color: var(--thunder-color-primary-main);
}
.thunder-button--text.thunder-button--primary:hover:not(:disabled) {
  color: var(--thunder-color-primary-dark);
  text-decoration: underline;
  text-underline-offset: 2px;
}

.thunder-button--text.thunder-button--secondary {
  color: var(--thunder-color-secondary-main);
}
.thunder-button--text.thunder-button--secondary:hover:not(:disabled) {
  color: var(--thunder-color-text-primary);
  text-decoration: underline;
  text-underline-offset: 2px;
}

.thunder-button--text.thunder-button--danger {
  color: var(--thunder-color-error-main);
}
.thunder-button--text.thunder-button--danger:hover:not(:disabled) {
  text-decoration: underline;
  text-underline-offset: 2px;
}

/* -- Inner elements -- */

.thunder-button__start-icon,
.thunder-button__end-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  line-height: 0;
}
.thunder-button--small .thunder-button__start-icon svg,
.thunder-button--small .thunder-button__end-icon svg {
  width: 14px;
  height: 14px;
}

.thunder-button__content {
  display: inline-flex;
  align-items: center;
}

.thunder-button__spinner {
  display: inline-block;
  width: 1em;
  height: 1em;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 50%;
  animation: thunder-spin 0.6s linear infinite;
  margin-left: calc(var(--thunder-spacing-unit) * 0.5);
}
`;

export default BUTTON_CSS;
