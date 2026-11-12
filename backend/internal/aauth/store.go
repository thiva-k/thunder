// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// Runtime store key prefixes.
const (
	keyPending     = "pending:"
	keyCode        = "code:"
	keyLoginState  = "login:"
	keyBinding     = "binding:"
	keyConsent     = "consent:"
	keySubject     = "sub:"
	keySubjectRev  = "subr:"
	keyMission     = "mission:"
	keyIssuedToken = "token:"
	// Bindings, consents, directed identifiers and missions are long-lived records.
	longLivedTTL = int64(365 * 24 * 60 * 60)
)

var errStoreConflict = errors.New("state changed concurrently")

// aauthStore keeps the person server's state in the runtime store.
type aauthStore struct {
	store providers.RuntimeStoreProvider
}

func newAAuthStore(store providers.RuntimeStoreProvider) *aauthStore {
	return &aauthStore{store: store}
}

func (s *aauthStore) putJSON(ctx context.Context, key string, value any, ttl int64) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.store.Put(ctx, providers.NamespaceAAuth, key, data, ttl)
}

// getJSON reads key into out and reports whether it was present.
func (s *aauthStore) getJSON(ctx context.Context, key string, out any) (bool, error) {
	data, err := s.store.Get(ctx, providers.NamespaceAAuth, key)
	if err != nil {
		return false, err
	}
	if data == nil {
		return false, nil
	}
	return true, json.Unmarshal(data, out)
}

func (s *aauthStore) getString(ctx context.Context, key string) (string, error) {
	data, err := s.store.Get(ctx, providers.NamespaceAAuth, key)
	if err != nil || data == nil {
		return "", err
	}
	return string(data), nil
}

// addPending stores a new pending request and its interaction code.
func (s *aauthStore) addPending(ctx context.Context, p *pendingRequest) error {
	ttl := int64(time.Until(p.ExpiresAt).Seconds()) + 1
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if err := s.store.Put(ctx, providers.NamespaceAAuth, keyPending+p.ID, data, ttl); err != nil {
		return err
	}
	if p.Code == "" {
		return nil
	}
	added, err := s.store.PutIfNotExists(ctx, providers.NamespaceAAuth, keyCode+p.Code, []byte(p.ID), ttl)
	if err != nil {
		return err
	}
	if !added {
		return errStoreConflict
	}
	return nil
}

func (s *aauthStore) getPending(ctx context.Context, id string) (*pendingRequest, error) {
	var p pendingRequest
	found, err := s.getJSON(ctx, keyPending+id, &p)
	if err != nil || !found {
		return nil, err
	}
	return &p, nil
}

// swapPending replaces the pending request when its state still equals expected.
func (s *aauthStore) swapPending(ctx context.Context, p *pendingRequest, expected string) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	swapped, err := s.store.CompareFieldAndSwapBeforeExpiry(ctx, providers.NamespaceAAuth,
		keyPending+p.ID, "State", expected, p.ExpiresAt, data)
	if err != nil {
		return err
	}
	if !swapped {
		return errStoreConflict
	}
	return nil
}

func (s *aauthStore) deletePending(ctx context.Context, id string) error {
	return s.store.Delete(ctx, providers.NamespaceAAuth, keyPending+id)
}

// takeCode consumes an interaction code and returns the pending request id it names.
func (s *aauthStore) takeCode(ctx context.Context, code string) (string, error) {
	data, err := s.store.Take(ctx, providers.NamespaceAAuth, keyCode+code)
	if err != nil || data == nil {
		return "", err
	}
	return string(data), nil
}

func (s *aauthStore) putLoginState(ctx context.Context, state, pendingID string, expiresAt time.Time) error {
	return s.store.Put(ctx, providers.NamespaceAAuth, keyLoginState+state, []byte(pendingID),
		int64(time.Until(expiresAt).Seconds())+1)
}

func (s *aauthStore) takeLoginState(ctx context.Context, state string) (string, error) {
	data, err := s.store.Take(ctx, providers.NamespaceAAuth, keyLoginState+state)
	if err != nil || data == nil {
		return "", err
	}
	return string(data), nil
}

func (s *aauthStore) getBinding(ctx context.Context, agentKey string) (string, error) {
	return s.getString(ctx, keyBinding+agentKey)
}

// bindAgent binds an agent to a person. It fails when the agent is already bound to someone else.
func (s *aauthStore) bindAgent(ctx context.Context, agentKey, userID string) error {
	added, err := s.store.PutIfNotExists(ctx, providers.NamespaceAAuth, keyBinding+agentKey,
		[]byte(userID), longLivedTTL)
	if err != nil {
		return err
	}
	if added {
		return nil
	}
	existing, err := s.getBinding(ctx, agentKey)
	if err != nil {
		return err
	}
	if existing != userID {
		return errStoreConflict
	}
	return nil
}

func consentKey(userID, agentKey, resource string) string {
	return keyConsent + userID + "|" + agentKey + "|" + resource
}

func (s *aauthStore) getConsent(ctx context.Context, userID, agentKey, resource string) (*consentRecord, error) {
	var c consentRecord
	found, err := s.getJSON(ctx, consentKey(userID, agentKey, resource), &c)
	if err != nil || !found {
		return nil, err
	}
	return &c, nil
}

func (s *aauthStore) putConsent(ctx context.Context, userID, agentKey, resource string, c *consentRecord) error {
	return s.putJSON(ctx, consentKey(userID, agentKey, resource), c, longLivedTTL)
}

// directedSubject returns the person's directed identifier at the audience, minting one when needed.
func (s *aauthStore) directedSubject(ctx context.Context, userID, audience string, mint func() string) (
	string, error) {
	key := keySubject + userID + "|" + audience
	existing, err := s.getString(ctx, key)
	if err != nil || existing != "" {
		return existing, err
	}
	candidate := mint()
	added, err := s.store.PutIfNotExists(ctx, providers.NamespaceAAuth, key, []byte(candidate), longLivedTTL)
	if err != nil {
		return "", err
	}
	if !added {
		return s.getString(ctx, key)
	}
	if err := s.store.Put(ctx, providers.NamespaceAAuth, keySubjectRev+audience+"|"+candidate,
		[]byte(userID), longLivedTTL); err != nil {
		return "", err
	}
	return candidate, nil
}

// subjectOwner returns the person a directed identifier at the audience belongs to.
func (s *aauthStore) subjectOwner(ctx context.Context, audience, subject string) (string, error) {
	return s.getString(ctx, keySubjectRev+audience+"|"+subject)
}

func (s *aauthStore) putMission(ctx context.Context, m *missionRecord) error {
	return s.putJSON(ctx, keyMission+m.S256, m, longLivedTTL)
}

func (s *aauthStore) getMission(ctx context.Context, s256 string) (*missionRecord, error) {
	var m missionRecord
	found, err := s.getJSON(ctx, keyMission+s256, &m)
	if err != nil || !found {
		return nil, err
	}
	return &m, nil
}

func (s *aauthStore) recordIssuedToken(ctx context.Context, jti string, t *issuedToken) error {
	ttl := t.ExpiresAt - time.Now().Unix() + int64(signatureWindow.Seconds())
	return s.putJSON(ctx, keyIssuedToken+jti, t, ttl)
}
