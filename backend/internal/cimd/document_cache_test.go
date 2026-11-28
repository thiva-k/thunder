// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package cimd

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type DocumentCacheTestSuite struct {
	suite.Suite
	now time.Time
}

func TestDocumentCacheTestSuite(t *testing.T) {
	suite.Run(t, new(DocumentCacheTestSuite))
}

func (s *DocumentCacheTestSuite) SetupTest() {
	s.now = time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)
}

func (s *DocumentCacheTestSuite) newCache(fetch documentFetcher) *documentCache {
	cache := newDocumentCache(fetch)
	cache.now = func() time.Time { return s.now }
	return cache
}

func (s *DocumentCacheTestSuite) TestGet_KeepsDocumentUntilItExpires() {
	var calls atomic.Int32
	cache := s.newCache(func(context.Context, string) (*cimdDocument, time.Duration, error) {
		calls.Add(1)
		return &cimdDocument{ClientID: testCIMDClientID}, 2 * time.Minute, nil
	})

	doc, err := cache.get(context.Background(), testCIMDClientID)
	s.Require().NoError(err)
	s.Equal(testCIMDClientID, doc.ClientID)
	_, err = cache.get(context.Background(), testCIMDClientID)
	s.Require().NoError(err)
	s.Equal(int32(1), calls.Load())

	s.now = s.now.Add(2 * time.Minute)
	_, err = cache.get(context.Background(), testCIMDClientID)
	s.Require().NoError(err)
	s.Equal(int32(2), calls.Load())
}

func (s *DocumentCacheTestSuite) TestGet_RemembersFailureBriefly() {
	var calls atomic.Int32
	cache := s.newCache(func(context.Context, string) (*cimdDocument, time.Duration, error) {
		calls.Add(1)
		return nil, 0, errors.New("unavailable")
	})

	_, err := cache.get(context.Background(), testCIMDClientID)
	s.Error(err)
	_, err = cache.get(context.Background(), testCIMDClientID)
	s.Error(err)
	s.Equal(int32(1), calls.Load())

	s.now = s.now.Add(documentFailureTTL)
	_, err = cache.get(context.Background(), testCIMDClientID)
	s.Error(err)
	s.Equal(int32(2), calls.Load())
}

func (s *DocumentCacheTestSuite) TestGet_SharesConcurrentRetrieval() {
	release := make(chan struct{})
	var calls atomic.Int32
	cache := s.newCache(func(context.Context, string) (*cimdDocument, time.Duration, error) {
		calls.Add(1)
		<-release
		return &cimdDocument{ClientID: testCIMDClientID}, time.Minute, nil
	})

	var wg sync.WaitGroup
	results := make(chan *cimdDocument, 5)
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			doc, err := cache.get(context.Background(), testCIMDClientID)
			if err == nil {
				results <- doc
			}
		}()
	}
	s.Eventually(func() bool {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		return len(cache.inflight) == 1
	}, time.Second, time.Millisecond)
	close(release)
	wg.Wait()
	close(results)

	count := 0
	for doc := range results {
		s.Equal(testCIMDClientID, doc.ClientID)
		count++
	}
	s.Equal(5, count)
	s.Equal(int32(1), calls.Load())
}

func (s *DocumentCacheTestSuite) TestGet_WaiterStopsWithItsRequest() {
	release := make(chan struct{})
	defer close(release)
	cache := s.newCache(func(context.Context, string) (*cimdDocument, time.Duration, error) {
		<-release
		return &cimdDocument{}, time.Minute, nil
	})
	go func() { _, _ = cache.get(context.Background(), testCIMDClientID) }()
	s.Eventually(func() bool {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		return len(cache.inflight) == 1
	}, time.Second, time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := cache.get(ctx, testCIMDClientID)

	s.ErrorIs(err, context.Canceled)
}

func (s *DocumentCacheTestSuite) TestStore_EvictsWhenFull() {
	cache := s.newCache(nil)
	for i := range maxCachedDocuments {
		cache.entries[string(rune('a'+i%26))+time.Duration(i).String()] = documentCacheEntry{
			expiresAt: s.now.Add(time.Minute),
		}
	}
	cache.entries["expired"] = documentCacheEntry{expiresAt: s.now}

	cache.store(testCIMDClientID, documentCacheEntry{expiresAt: s.now.Add(time.Minute)})

	s.LessOrEqual(len(cache.entries), maxCachedDocuments)
	s.Contains(cache.entries, testCIMDClientID)
	s.NotContains(cache.entries, "expired")
}

func (s *DocumentCacheTestSuite) TestDocumentTTL() {
	cases := []struct {
		cacheControl string
		want         time.Duration
	}{
		{"", documentDefaultTTL},
		{"public, max-age=300", 5 * time.Minute},
		{"max-age=10", documentMinTTL},
		{"max-age=86400", documentMaxTTL},
		{`max-age="120"`, 2 * time.Minute},
		{"max-age=abc", documentDefaultTTL},
		{"no-store,no-cache,max-age=0", documentMinTTL},
		{"private", documentDefaultTTL},
	}
	for _, tc := range cases {
		s.Run(tc.cacheControl, func() {
			s.Equal(tc.want, documentTTL(http.Header{"Cache-Control": []string{tc.cacheControl}}))
		})
	}
}
