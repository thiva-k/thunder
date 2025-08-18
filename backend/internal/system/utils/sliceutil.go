/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

package utils

import "fmt"

// DeepCopyMapOfStrings creates a deep copy of a map with strings.
func DeepCopyMapOfStrings(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// DeepCopyMapOfStringSlices creates a deep copy of a map of string slices.
func DeepCopyMapOfStringSlices(src map[string][]string) map[string][]string {
	if src == nil {
		return nil
	}
	dst := make(map[string][]string, len(src))
	for k, v := range src {
		copied := append([]string(nil), v...)
		dst[k] = copied
	}
	return dst
}

// ClonableInterface defines an interface for clonable types.
type ClonableInterface interface {
	Clone() (ClonableInterface, error)
}

// DeepCopyMapOfClonables creates a deep copy of a map with clonable values.
func DeepCopyMapOfClonables[T ClonableInterface](src map[string]T) (map[string]T, error) {
	if src == nil {
		return nil, nil
	}
	dst := make(map[string]T, len(src))
	for k, v := range src {
		cloned, err := v.Clone()
		if err != nil {
			return nil, fmt.Errorf("failed to clone value for key %s: %w", k, err)
		}
		if _, ok := cloned.(T); !ok {
			return nil, fmt.Errorf("cloned value for key %s is not of type: %T", k, cloned)
		}
		dst[k] = cloned.(T)
	}
	return dst, nil
}
