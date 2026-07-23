/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package model

// IsValidApplicationType reports whether t is a recognized application type.
func IsValidApplicationType(t ApplicationType) bool {
	switch t {
	case ApplicationTypeBrowser, ApplicationTypeFullStack, ApplicationTypeMobile,
		ApplicationTypeM2M, ApplicationTypeCustom:
		return true
	default:
		return false
	}
}

// ResolveApplicationType converts a raw stored type value into the canonical ApplicationType,
// defaulting to ApplicationTypeCustom when raw is empty or unrecognized. This covers applications
// created before the type attribute existed, or holding a legacy/corrupted value, so callers treat
// them as unconstrained rather than failing.
func ResolveApplicationType(raw string) ApplicationType {
	if t := ApplicationType(raw); IsValidApplicationType(t) {
		return t
	}
	return ApplicationTypeCustom
}
