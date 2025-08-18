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

package cache

import (
	"time"
)

// CacheKey represents a key for the cache.
type CacheKey struct {
	Key string
}

// ToString returns the string representation of the CacheKey.
func (key CacheKey) ToString() string {
	return key.Key
}

// CacheEntry represents a cache entry.
type CacheEntry[T any] struct {
	Value      T
	ExpiryTime time.Time
}

// CacheStat represents cache statistics.
type CacheStat struct {
	Enabled    bool
	Size       int
	MaxSize    int
	HitCount   int64
	MissCount  int64
	HitRate    float64
	EvictCount int64
}
