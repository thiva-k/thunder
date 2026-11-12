// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

// signedRequest is an HTTP request whose HTTP Message Signature (RFC 9421) has been parsed but not
// yet verified: the label, the signature, the signature base, and the Signature-Key member that
// carries its key.
type signedRequest struct {
	label     string
	signature []byte
	base      []byte
	created   int64
	keyMember sfMember
	body      []byte
}

// parseSignedRequest reads the Signature, Signature-Input, and Signature-Key headers, checks the
// covered components and freshness the AAuth profile requires of a request to the person server,
// content-digest and content-type included on a body, and builds the signature base. It returns a
// signature error code on failure. The request body is read and kept so that handlers can decode it
// afterwards.
func parseSignedRequest(r *http.Request, now time.Time) (*signedRequest, *sigError) {
	sigHeader := r.Header.Get(headerSignature)
	inputHeader := r.Header.Get(headerSignatureInput)
	keyHeader := r.Header.Get(headerSignatureKey)
	if sigHeader == "" || inputHeader == "" || keyHeader == "" {
		return nil, &sigError{code: sigErrInvalidSignature, detail: "missing signature headers"}
	}
	inputs, err := parseSFDictionary(inputHeader)
	if err != nil {
		return nil, &sigError{code: sigErrInvalidSignature, detail: "malformed Signature-Input"}
	}
	sigs, err := parseSFDictionary(sigHeader)
	if err != nil {
		return nil, &sigError{code: sigErrInvalidSignature, detail: "malformed Signature"}
	}
	keys, err := parseSFDictionary(keyHeader)
	if err != nil {
		return nil, &sigError{code: sigErrInvalidKey, detail: "malformed Signature-Key"}
	}

	// Verify the first label that has all three members.
	for _, input := range inputs {
		if input.list == nil {
			continue
		}
		sigMember, ok := sigs.get(input.key)
		if !ok || sigMember.item == nil {
			continue
		}
		keyMember, ok := keys.get(input.key)
		if !ok {
			continue
		}
		signature, ok := sigMember.item.value.([]byte)
		if !ok {
			return nil, &sigError{code: sigErrInvalidSignature, detail: "signature is not a byte sequence"}
		}
		return buildSignedRequest(r, input, signature, keyMember, now)
	}
	return nil, &sigError{code: sigErrInvalidSignature, detail: "no signature label with a Signature-Key member"}
}

func buildSignedRequest(r *http.Request, input sfMember, signature []byte, keyMember sfMember,
	now time.Time) (*signedRequest, *sigError) {
	covered := make([]string, 0, len(input.list.items))
	for _, item := range input.list.items {
		name, ok := item.value.(string)
		if !ok || len(item.params) != 0 {
			return nil, &sigError{code: sigErrInvalidInput, detail: "unsupported covered component"}
		}
		covered = append(covered, strings.ToLower(name))
	}
	required := []string{"@method", "@authority", "@path", "signature-key"}
	body, readErr := readBody(r)
	if readErr != nil {
		return nil, &sigError{code: sigErrInvalidSignature, detail: "unreadable body"}
	}
	if len(body) > 0 {
		required = append(required, "content-digest", "content-type")
	}
	for _, c := range required {
		if !slices.Contains(covered, c) {
			return nil, &sigError{code: sigErrInvalidInput, detail: "signature must cover " + c,
				requiredInput: required}
		}
	}

	createdValue, ok := param(input.list.params, "created")
	created, isInt := createdValue.(int64)
	if !ok || !isInt {
		return nil, &sigError{code: sigErrInvalidSignature, detail: "created is required"}
	}
	nowUnix := now.Unix()
	window := int64(signatureWindow.Seconds())
	if created < nowUnix-window {
		return nil, &sigError{code: sigErrInvalidSignature, detail: "signature is too old"}
	}
	if created > nowUnix+window {
		return nil, &sigError{code: sigErrClockSkew, detail: "signature created in the future"}
	}
	if expiresValue, ok := param(input.list.params, "expires"); ok {
		if expires, isInt := expiresValue.(int64); !isInt || expires < nowUnix {
			return nil, &sigError{code: sigErrInvalidSignature, detail: "signature expired"}
		}
	}

	if slices.Contains(covered, "content-digest") {
		if err := verifyContentDigest(r.Header.Get(headerContentDigest), body); err != nil {
			return nil, err
		}
	}

	var base strings.Builder
	for _, c := range covered {
		value, ok := componentValue(r, c)
		if !ok {
			return nil, &sigError{code: sigErrInvalidInput, detail: "covered component " + c + " is absent"}
		}
		base.WriteString(sfString(c) + ": " + value + "\n")
	}
	base.WriteString(sfString("@signature-params") + ": " + serializeInnerList(input.list))

	return &signedRequest{
		label:     input.key,
		signature: signature,
		base:      []byte(base.String()),
		created:   created,
		keyMember: keyMember,
		body:      body,
	}, nil
}

// componentValue returns the canonical value of a covered component (RFC 9421 Section 2).
func componentValue(r *http.Request, name string) (string, bool) {
	switch name {
	case "@method":
		return strings.ToUpper(r.Method), true
	case "@authority":
		return requestAuthority(r), true
	case "@path":
		path := r.URL.EscapedPath()
		if path == "" {
			path = "/"
		}
		return path, true
	case "@query":
		return "?" + r.URL.RawQuery, true
	case "@scheme":
		return "https", true
	case "@target-uri":
		uri := "https://" + requestAuthority(r) + r.URL.EscapedPath()
		if r.URL.RawQuery != "" {
			uri += "?" + r.URL.RawQuery
		}
		return uri, true
	}
	if strings.HasPrefix(name, "@") {
		return "", false
	}
	values := r.Header.Values(name)
	if len(values) == 0 {
		return "", false
	}
	trimmed := make([]string, 0, len(values))
	for _, v := range values {
		trimmed = append(trimmed, strings.TrimSpace(v))
	}
	return strings.Join(trimmed, ", "), true
}

// requestAuthority returns the request's authority, lowercased, without the default HTTPS port.
func requestAuthority(r *http.Request) string {
	host := r.Host
	if host == "" {
		host = r.URL.Host
	}
	return strings.TrimSuffix(strings.ToLower(host), ":443")
}

// verifyContentDigest checks the Content-Digest header (RFC 9530) against the body. At least one
// supported digest must be present and every supported digest present must match.
func verifyContentDigest(header string, body []byte) *sigError {
	dict, err := parseSFDictionary(header)
	if err != nil || len(dict) == 0 {
		return &sigError{code: sigErrInvalidSignature, detail: "malformed Content-Digest"}
	}
	matched := false
	for _, member := range dict {
		if member.item == nil {
			continue
		}
		digest, ok := member.item.value.([]byte)
		if !ok {
			continue
		}
		var sum []byte
		switch member.key {
		case "sha-256":
			s := sha256.Sum256(body)
			sum = s[:]
		case "sha-512":
			s := sha512.Sum512(body)
			sum = s[:]
		default:
			continue
		}
		if !bytes.Equal(sum, digest) {
			return &sigError{code: sigErrInvalidSignature, detail: "Content-Digest does not match the body"}
		}
		matched = true
	}
	if !matched {
		return &sigError{code: sigErrInvalidSignature, detail: "no supported Content-Digest algorithm"}
	}
	return nil
}

// readBody reads the request body and replaces it so later readers see the same bytes.
func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodyBytes))
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

// encodeB64URL encodes b as unpadded base64url.
func encodeB64URL(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
