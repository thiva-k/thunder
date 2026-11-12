// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	syshttp "github.com/thunder-id/thunderid/internal/system/http"
	"github.com/thunder-id/thunderid/internal/system/jose/jws"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

var (
	errMalformedJWT      = errors.New("malformed JWT")
	errUnsupportedAlg    = errors.New("unsupported algorithm")
	errKeyAlgMismatch    = errors.New("key does not match its algorithm")
	errUnknownKey        = errors.New("key not found")
	errBadSignature      = errors.New("signature verification failed")
	errIssuerMissing     = errors.New("metadata has no issuer")
	errIssuerMismatch    = errors.New("metadata issuer does not match")
	errMetadataFetch     = errors.New("metadata could not be fetched")
	errInvalidIdentifier = errors.New("not a valid server identifier")
)

// cryptoAlgorithm maps a fully-specified JWS algorithm to the identifier the crypto provider uses.
// Polymorphic, symmetric and unsecured algorithms are refused.
func cryptoAlgorithm(alg string) (string, error) {
	switch alg {
	case algEd25519:
		return "EdDSA", nil
	case algES256, "ES384", "ES512", algRS256, "RS512", "PS256", "ML-DSA-44", "ML-DSA-65", "ML-DSA-87":
		return alg, nil
	default:
		return "", errUnsupportedAlg
	}
}

// checkJWKForAlg verifies a public JWK carries a fully-specified alg whose key type and curve agree.
func checkJWKForAlg(jwk map[string]any) (string, error) {
	alg, _ := jwk["alg"].(string)
	if alg == "" {
		return "", errUnsupportedAlg
	}
	if _, err := cryptoAlgorithm(alg); err != nil {
		return "", err
	}
	if _, private := jws.ContainsPrivateMember(jwk); private {
		return "", errKeyAlgMismatch
	}
	kty, _ := jwk["kty"].(string)
	crv, _ := jwk["crv"].(string)
	ok := false
	switch alg {
	case algEd25519:
		ok = kty == "OKP" && crv == "Ed25519" && jwk["x"] != nil
	case algES256:
		ok = kty == "EC" && crv == "P-256" && jwk["x"] != nil && jwk["y"] != nil
	case "ES384":
		ok = kty == "EC" && crv == "P-384" && jwk["x"] != nil && jwk["y"] != nil
	case "ES512":
		ok = kty == "EC" && crv == "P-521" && jwk["x"] != nil && jwk["y"] != nil
	case algRS256, "RS512", "PS256":
		ok = kty == "RSA" && jwk["n"] != nil && jwk["e"] != nil
	case "ML-DSA-44", "ML-DSA-65", "ML-DSA-87":
		ok = kty == "AKP"
	}
	if !ok {
		return "", errKeyAlgMismatch
	}
	return alg, nil
}

// parseJWT splits a compact JWS and decodes its header and claims.
func parseJWT(token string) (*parsedJWT, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errMalformedJWT
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errMalformedJWT
	}
	claimBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errMalformedJWT
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errMalformedJWT
	}
	p := &parsedJWT{signingInput: []byte(parts[0] + "." + parts[1]), signature: signature}
	if json.Unmarshal(headerBytes, &p.header) != nil || json.Unmarshal(claimBytes, &p.claims) != nil {
		return nil, errMalformedJWT
	}
	return p, nil
}

func (p *parsedJWT) headerString(name string) string {
	v, _ := p.header[name].(string)
	return v
}

func (p *parsedJWT) claimString(name string) string {
	v, _ := p.claims[name].(string)
	return v
}

// exp returns the exp claim, or zero when it is absent.
func (p *parsedJWT) exp() int64 {
	switch v := p.claims["exp"].(type) {
	case float64:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	}
	return 0
}

// cnfJWK returns the cnf.jwk claim.
func (p *parsedJWT) cnfJWK() map[string]any {
	cnf, _ := p.claims["cnf"].(map[string]any)
	if cnf == nil {
		return nil
	}
	jwk, _ := cnf["jwk"].(map[string]any)
	return jwk
}

// isServerIdentifier reports whether v is an AAuth server identifier: https, lowercase host, and no
// port, path, query, fragment, or trailing slash.
func isServerIdentifier(v string) bool {
	u, err := url.Parse(v)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.Port() != "" || u.User != nil {
		return false
	}
	return u.Path == "" && u.RawQuery == "" && u.Fragment == "" && !strings.HasSuffix(v, "/") &&
		v == strings.ToLower(v) && v == "https://"+u.Host
}

// isAgentIdentifier reports whether v is an agent identifier of the form aauth:local@domain.
func isAgentIdentifier(v string) bool {
	rest, ok := strings.CutPrefix(v, "aauth:")
	if !ok {
		return false
	}
	local, domain, ok := strings.Cut(rest, "@")
	if !ok || local == "" || len(local) > 255 || domain == "" {
		return false
	}
	for _, c := range local {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') &&
			c != '-' && c != '_' && c != '+' && c != '.' {
			return false
		}
	}
	return isServerIdentifier("https://" + domain)
}

// issuerMetadata is the subset of an AAuth metadata document the person server reads.
type issuerMetadata struct {
	Issuer      string         `json:"issuer"`
	JWKSURI     string         `json:"jwks_uri"`
	JWKS        map[string]any `json:"jwks"`
	Name        string         `json:"name"`
	ClientName  string         `json:"client_name"`
	Description string         `json:"description"`
	LogoURI     string         `json:"logo_uri"`
	AccessMode  string         `json:"access_mode"`
}

type cachedIssuer struct {
	metadata  *issuerMetadata
	keys      []map[string]any
	fetchedAt time.Time
}

// issuerKeyFetcher discovers an issuer's keys through {iss}/.well-known/{dwk} and caches them.
type issuerKeyFetcher struct {
	client syshttp.HTTPClientInterface
	mu     sync.Mutex
	cache  map[string]*cachedIssuer
}

func newIssuerKeyFetcher(client syshttp.HTTPClientInterface) *issuerKeyFetcher {
	return &issuerKeyFetcher{client: client, cache: map[string]*cachedIssuer{}}
}

// metadata returns the issuer's metadata document, refetching when stale or when refresh is set.
func (f *issuerKeyFetcher) metadata(ctx context.Context, issuer, dwk string, refresh bool) (*cachedIssuer, error) {
	if !isServerIdentifier(issuer) {
		return nil, errInvalidIdentifier
	}
	cacheKey := issuer + "|" + dwk
	f.mu.Lock()
	cached := f.cache[cacheKey]
	f.mu.Unlock()
	if cached != nil {
		age := time.Since(cached.fetchedAt)
		if age < metadataCacheLifetime && (!refresh || age < time.Minute) {
			return cached, nil
		}
	}

	var meta issuerMetadata
	if err := f.getJSON(ctx, issuer+"/.well-known/"+dwk, &meta); err != nil {
		return nil, err
	}
	if meta.Issuer == "" {
		return nil, errIssuerMissing
	}
	if meta.Issuer != issuer {
		return nil, errIssuerMismatch
	}
	jwks := meta.JWKS
	if jwks == nil {
		if meta.JWKSURI == "" || !strings.HasPrefix(meta.JWKSURI, "https://") {
			return nil, errMetadataFetch
		}
		jwks = map[string]any{}
		if err := f.getJSON(ctx, meta.JWKSURI, &jwks); err != nil {
			return nil, err
		}
	}
	keys := []map[string]any{}
	if list, ok := jwks["keys"].([]any); ok {
		for _, k := range list {
			if m, ok := k.(map[string]any); ok {
				keys = append(keys, m)
			}
		}
	}
	entry := &cachedIssuer{metadata: &meta, keys: keys, fetchedAt: time.Now()}
	f.mu.Lock()
	f.cache[cacheKey] = entry
	f.mu.Unlock()
	return entry, nil
}

func (f *issuerKeyFetcher) getJSON(ctx context.Context, target string, out any) error {
	ctx, cancel := context.WithTimeout(ctx, httpFetchTimeout)
	defer cancel()
	// The URL is an issuer identifier checked by isServerIdentifier, and the client refuses private
	// addresses unless configured for local development.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil) //nolint:gosec // G704: see above
	if err != nil {
		return errMetadataFetch
	}
	req.Header.Set("Accept", contentTypeJSON)
	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", errMetadataFetch, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", errMetadataFetch, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxMetadataBytes))
	if err != nil {
		return errMetadataFetch
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("%w: %w", errMetadataFetch, err)
	}
	return nil
}

// key returns the issuer's key with the given kid, refreshing the cache once on a miss.
func (f *issuerKeyFetcher) key(ctx context.Context, issuer, dwk, kid string) (map[string]any, *cachedIssuer,
	error) {
	for _, refresh := range []bool{false, true} {
		entry, err := f.metadata(ctx, issuer, dwk, refresh)
		if err != nil {
			return nil, nil, err
		}
		for _, k := range entry.keys {
			if k["kid"] == kid {
				return k, entry, nil
			}
		}
		if kid == "" && len(entry.keys) == 1 {
			return entry.keys[0], entry, nil
		}
	}
	return nil, nil, errUnknownKey
}

// verifyWithJWK verifies data against a signature with a public JWK, taking the algorithm from it.
func verifyWithJWK(ctx context.Context, crypto providers.RuntimeCryptoProvider, jwk map[string]any,
	expectedAlg string, data, signature []byte) error {
	alg, err := checkJWKForAlg(jwk)
	if err != nil {
		return err
	}
	if expectedAlg != "" && expectedAlg != alg {
		return errKeyAlgMismatch
	}
	cryptoAlg, _ := cryptoAlgorithm(alg)
	public := make(map[string]any, len(jwk))
	for k, v := range jwk {
		if k != "alg" {
			public[k] = v
		}
	}
	if err := crypto.Verify(ctx, providers.KeyRef{PublicKeyJWK: public}, cryptoAlg, data, signature); err != nil {
		return errBadSignature
	}
	return nil
}

// verifyIssuerJWT verifies a JWT issued by another AAuth server, discovering its key from
// {iss}/.well-known/{dwk}. It returns the issuer's metadata.
func verifyIssuerJWT(ctx context.Context, crypto providers.RuntimeCryptoProvider, fetcher *issuerKeyFetcher,
	p *parsedJWT, dwk string) (*issuerMetadata, error) {
	alg := p.headerString("alg")
	if _, err := cryptoAlgorithm(alg); err != nil {
		return nil, err
	}
	issuer := p.claimString("iss")
	if !isServerIdentifier(issuer) {
		return nil, errInvalidIdentifier
	}
	key, entry, err := fetcher.key(ctx, issuer, dwk, p.headerString("kid"))
	if err != nil {
		return nil, err
	}
	if err := verifyWithJWK(ctx, crypto, key, alg, p.signingInput, p.signature); err != nil {
		// A rotated key: refetch once before failing.
		if refreshed, _, rerr := fetcher.key(ctx, issuer, dwk, p.headerString("kid")); rerr == nil {
			if verifyWithJWK(ctx, crypto, refreshed, alg, p.signingInput, p.signature) == nil {
				return entry.metadata, nil
			}
		}
		return nil, err
	}
	return entry.metadata, nil
}

// jktOf returns the RFC 7638 thumbprint of a public JWK.
func jktOf(jwk map[string]any) (string, error) {
	return jws.ComputeJKT(jwk)
}
