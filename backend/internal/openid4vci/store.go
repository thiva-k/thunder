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

package openid4vci

import (
	"context"
)

// offerStore persists short-lived issuer-initiated credential offers keyed by id.
type offerStore interface {
	Save(ctx context.Context, rec *offerRecord) error
	Get(ctx context.Context, id string) (*offerRecord, bool)
}

// nonceStore persists short-lived c_nonce values keyed by the nonce string.
type nonceStore interface {
	Save(ctx context.Context, rec *nonceRecord) error
	Get(ctx context.Context, nonce string) (*nonceRecord, bool)
	Delete(ctx context.Context, nonce string) error
}

// memoryNonceStore is a process-local nonceStore for test wiring. Not safe for concurrent use.
type memoryNonceStore struct {
	entries map[string]*nonceRecord
}

func newInMemoryNonceStore() nonceStore {
	return &memoryNonceStore{entries: map[string]*nonceRecord{}}
}

func (m *memoryNonceStore) Save(_ context.Context, rec *nonceRecord) error {
	m.entries[rec.Nonce] = rec
	return nil
}

func (m *memoryNonceStore) Get(_ context.Context, nonce string) (*nonceRecord, bool) {
	rec, ok := m.entries[nonce]
	return rec, ok
}

func (m *memoryNonceStore) Delete(_ context.Context, nonce string) error {
	delete(m.entries, nonce)
	return nil
}
