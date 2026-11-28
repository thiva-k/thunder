// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package cimd

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// documentFetcher retrieves the document at a Client Identifier URL and returns how long it may be kept.
type documentFetcher func(ctx context.Context, clientID string) (*cimdDocument, time.Duration, error)

// documentCacheEntry is a retrieval result kept until expiresAt: the document, or the failure.
type documentCacheEntry struct {
	doc       *cimdDocument
	err       error
	expiresAt time.Time
}

// documentRetrieval is a retrieval in progress, which concurrent requests for the same URL wait for.
type documentRetrieval struct {
	done chan struct{}
	doc  *cimdDocument
	err  error
}

// documentCache keeps the documents of the clients admitted by Client ID domain in memory. Concurrent
// requests for one URL share a single retrieval, and a failure is remembered briefly.
type documentCache struct {
	fetch    documentFetcher
	now      func() time.Time
	mu       sync.Mutex
	entries  map[string]documentCacheEntry
	inflight map[string]*documentRetrieval
}

// newDocumentCache creates a document cache that retrieves through fetch.
func newDocumentCache(fetch documentFetcher) *documentCache {
	return &documentCache{
		fetch:    fetch,
		now:      time.Now,
		entries:  map[string]documentCacheEntry{},
		inflight: map[string]*documentRetrieval{},
	}
}

// get returns the document at clientID, from the cache while it is fresh.
func (c *documentCache) get(ctx context.Context, clientID string) (*cimdDocument, error) {
	c.mu.Lock()
	if entry, ok := c.entries[clientID]; ok && c.now().Before(entry.expiresAt) {
		c.mu.Unlock()
		return entry.doc, entry.err
	}
	if retrieval, ok := c.inflight[clientID]; ok {
		c.mu.Unlock()
		select {
		case <-retrieval.done:
			return retrieval.doc, retrieval.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	retrieval := &documentRetrieval{done: make(chan struct{})}
	c.inflight[clientID] = retrieval
	c.mu.Unlock()

	// The retrieval is shared, so it outlives the request that started it. It is bounded by the fetch
	// timeout.
	doc, ttl, err := c.fetch(context.WithoutCancel(ctx), clientID)
	if err != nil {
		ttl = documentFailureTTL
	}

	c.mu.Lock()
	retrieval.doc, retrieval.err = doc, err
	delete(c.inflight, clientID)
	c.store(clientID, documentCacheEntry{doc: doc, err: err, expiresAt: c.now().Add(ttl)})
	c.mu.Unlock()
	close(retrieval.done)
	return doc, err
}

// store keeps an entry, first dropping expired entries, and an arbitrary one when the cache is full.
// The caller holds the lock.
func (c *documentCache) store(clientID string, entry documentCacheEntry) {
	if len(c.entries) >= maxCachedDocuments {
		now := c.now()
		for key, existing := range c.entries {
			if !now.Before(existing.expiresAt) {
				delete(c.entries, key)
			}
		}
		for key := range c.entries {
			if len(c.entries) < maxCachedDocuments {
				break
			}
			delete(c.entries, key)
		}
	}
	c.entries[clientID] = entry
}

// documentTTL returns how long a document may be kept: its response's max-age within the bounds, the
// shortest time for a response that forbids storing it, and the default when the response says nothing.
func documentTTL(header http.Header) time.Duration {
	for _, directive := range strings.Split(header.Get("Cache-Control"), ",") {
		name, value, _ := strings.Cut(strings.TrimSpace(strings.ToLower(directive)), "=")
		switch name {
		case "no-store", "no-cache":
			return documentMinTTL
		case "max-age":
			seconds, err := strconv.Atoi(strings.Trim(value, `"`))
			if err != nil || seconds < 0 {
				return documentDefaultTTL
			}
			return min(max(time.Duration(seconds)*time.Second, documentMinTTL), documentMaxTTL)
		}
	}
	return documentDefaultTTL
}
