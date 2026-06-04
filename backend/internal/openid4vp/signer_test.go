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

package openid4vp

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/thunder-id/thunderid/internal/system/cryptolib"
	kmprovider "github.com/thunder-id/thunderid/internal/system/kmprovider/common"
	"github.com/thunder-id/thunderid/tests/mocks/crypto/cryptomock"
)

func newSignerMock(
	t *testing.T, key *ecdsa.PrivateKey, info kmprovider.PublicKeyInfo,
) *cryptomock.RuntimeCryptoProviderMock {
	t.Helper()
	m := cryptomock.NewRuntimeCryptoProviderMock(t)
	m.EXPECT().GetPublicKeys(mock.Anything, mock.Anything).Return([]kmprovider.PublicKeyInfo{info}, nil).Maybe()
	m.EXPECT().Sign(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(
			_ context.Context, _ kmprovider.KeyRef, _ cryptolib.SignAlgorithm, content []byte,
		) ([]byte, error) {
			return cryptolib.Generate(content, cryptolib.ECDSASHA256, key)
		}).Maybe()
	return m
}

func TestRequestSignerSignsVerifiableJAR(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	info := kmprovider.PublicKeyInfo{
		KeyID:          "vp-signing",
		Algorithm:      cryptolib.AlgorithmES256,
		PublicKey:      &key.PublicKey,
		Thumbprint:     "thumb-1",
		CertificateDER: []byte{0x30, 0x82, 0x01, 0x02, 0x03},
	}
	m := newSignerMock(t, key, info)

	signer, err := newRequestSigner(context.Background(), m, "vp-signing")
	require.NoError(t, err)

	jar, err := signer.signRequestObject(context.Background(), map[string]interface{}{
		"response_type": "vp_token",
		"client_id":     "x509_hash:abc",
	})
	require.NoError(t, err)

	parts := strings.Split(jar, ".")
	require.Len(t, parts, 3)

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)
	var header map[string]interface{}
	require.NoError(t, json.Unmarshal(headerJSON, &header))
	assert.Equal(t, "ES256", header["alg"])
	assert.Equal(t, requestObjectType, header["typ"])
	assert.Equal(t, "thumb-1", header["kid"])
	x5c := header["x5c"].([]interface{})
	require.Len(t, x5c, 1)
	assert.Equal(t, base64.StdEncoding.EncodeToString(info.CertificateDER), x5c[0])

	// Signature is in JWS P1363 format (r||s, 32 bytes each for P-256).
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	require.NoError(t, err)
	require.Len(t, sig, 64)
	signingInput := parts[0] + "." + parts[1]
	hashed := sha256.Sum256([]byte(signingInput))
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	assert.True(t, ecdsa.Verify(&key.PublicKey, hashed[:], r, s))
}

func TestNewRequestSignerErrors(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	t.Run("nil provider", func(t *testing.T) {
		_, err := newRequestSigner(context.Background(), nil, "k")
		assert.ErrorIs(t, err, ErrPolicy)
	})

	t.Run("no key found", func(t *testing.T) {
		m := cryptomock.NewRuntimeCryptoProviderMock(t)
		m.EXPECT().GetPublicKeys(mock.Anything, mock.Anything).Return(nil, nil)
		_, err := newRequestSigner(context.Background(), m, "missing")
		assert.ErrorIs(t, err, ErrPolicy)
	})

	t.Run("provider returns error", func(t *testing.T) {
		m := cryptomock.NewRuntimeCryptoProviderMock(t)
		m.EXPECT().GetPublicKeys(mock.Anything, mock.Anything).Return(nil, errors.New("provider unavailable"))
		_, err := newRequestSigner(context.Background(), m, "k")
		assert.Error(t, err)
	})

	t.Run("unsupported algorithm", func(t *testing.T) {
		info := kmprovider.PublicKeyInfo{
			KeyID:          "vp-signing",
			Algorithm:      cryptolib.Algorithm("UNSUPPORTED-ALG"),
			PublicKey:      &key.PublicKey,
			CertificateDER: []byte{0x30, 0x82},
		}
		m := cryptomock.NewRuntimeCryptoProviderMock(t)
		m.EXPECT().GetPublicKeys(mock.Anything, mock.Anything).Return([]kmprovider.PublicKeyInfo{info}, nil)
		_, err := newRequestSigner(context.Background(), m, "vp-signing")
		assert.ErrorIs(t, err, ErrPolicy)
	})

	t.Run("missing certificate", func(t *testing.T) {
		info := kmprovider.PublicKeyInfo{
			KeyID:     "vp-signing",
			Algorithm: cryptolib.AlgorithmES256,
			PublicKey: &key.PublicKey,
		}
		m := cryptomock.NewRuntimeCryptoProviderMock(t)
		m.EXPECT().GetPublicKeys(mock.Anything, mock.Anything).Return([]kmprovider.PublicKeyInfo{info}, nil)
		_, err := newRequestSigner(context.Background(), m, "vp-signing")
		assert.ErrorIs(t, err, ErrPolicy)
	})
}

// TestSignRequestObjectNoKid verifies that when the signing key has no thumbprint
// the kid header field is omitted entirely from the signed JAR.
func TestSignRequestObjectNoKid(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	info := kmprovider.PublicKeyInfo{
		KeyID:          "vp-signing",
		Algorithm:      cryptolib.AlgorithmES256,
		PublicKey:      &key.PublicKey,
		Thumbprint:     "", // no kid
		CertificateDER: []byte{0x30, 0x82, 0x01, 0x02, 0x03},
	}
	m := newSignerMock(t, key, info)

	signer, err := newRequestSigner(context.Background(), m, "vp-signing")
	require.NoError(t, err)

	jar, err := signer.signRequestObject(context.Background(), map[string]interface{}{"a": "b"})
	require.NoError(t, err)

	parts := strings.Split(jar, ".")
	require.Len(t, parts, 3)
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)
	var header map[string]interface{}
	require.NoError(t, json.Unmarshal(headerJSON, &header))
	_, hasKid := header["kid"]
	assert.False(t, hasKid, "kid must be absent when Thumbprint is empty")
}

// TestSignRequestObjectSignError verifies that a crypto-provider signing failure
// is surfaced as an error without panicking.
func TestSignRequestObjectSignError(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	info := kmprovider.PublicKeyInfo{
		KeyID:          "vp-signing",
		Algorithm:      cryptolib.AlgorithmES256,
		PublicKey:      &key.PublicKey,
		CertificateDER: []byte{0x30, 0x82},
	}
	m := cryptomock.NewRuntimeCryptoProviderMock(t)
	m.EXPECT().GetPublicKeys(mock.Anything, mock.Anything).Return([]kmprovider.PublicKeyInfo{info}, nil)
	m.EXPECT().Sign(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("HSM unreachable"))

	signer, err := newRequestSigner(context.Background(), m, "vp-signing")
	require.NoError(t, err)

	_, err = signer.signRequestObject(context.Background(), map[string]interface{}{"a": "b"})
	assert.Error(t, err)
}

// TestEcdsaDERToJWSVariousAlgorithms covers the P-384 and P-521 coordinate-length
// branches, the non-DER passthrough, and the unknown-algorithm passthrough.
func TestEcdsaDERToJWSVariousAlgorithms(t *testing.T) {
	t.Run("P-384 produces 96-byte JWS signature", func(t *testing.T) {
		key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		require.NoError(t, err)
		derSig, err := cryptolib.Generate([]byte("input"), cryptolib.ECDSASHA384, key)
		require.NoError(t, err)

		result := ecdsaDERToJWS(derSig, cryptolib.ECDSASHA384)
		assert.Len(t, result, 96) // 48 bytes × 2

		r := new(big.Int).SetBytes(result[:48])
		s := new(big.Int).SetBytes(result[48:])
		hashed := sha256.Sum256([]byte("input")) // hash used only to confirm r/s are large ints
		_ = hashed
		assert.True(t, r.Sign() > 0 && s.Sign() > 0)
	})

	t.Run("P-521 produces 132-byte JWS signature", func(t *testing.T) {
		key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		require.NoError(t, err)
		derSig, err := cryptolib.Generate([]byte("input"), cryptolib.ECDSASHA512, key)
		require.NoError(t, err)

		result := ecdsaDERToJWS(derSig, cryptolib.ECDSASHA512)
		assert.Len(t, result, 132) // 66 bytes × 2
	})

	t.Run("non-DER input is returned unchanged", func(t *testing.T) {
		raw := []byte("not-asn1-der-data")
		result := ecdsaDERToJWS(raw, cryptolib.ECDSASHA256)
		assert.Equal(t, raw, result)
	})

	t.Run("unknown algorithm returns DER input unchanged", func(t *testing.T) {
		// Build a valid ASN.1 DER signature so the unmarshal succeeds, then
		// provide an unknown algorithm so the switch hits default.
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		derSig, err := cryptolib.Generate([]byte("input"), cryptolib.ECDSASHA256, key)
		require.NoError(t, err)

		var parsed struct{ R, S *big.Int }
		_, err = asn1.Unmarshal(derSig, &parsed)
		require.NoError(t, err)

		result := ecdsaDERToJWS(derSig, cryptolib.SignAlgorithm("UNKNOWN-ALG"))
		assert.Equal(t, derSig, result)
	})
}
