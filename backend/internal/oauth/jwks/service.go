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

// Package jwks provides the implementation for retrieving JSON Web Key Sets (JWKS).
package jwks

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"crypto/x509"

	// Use crypto/sha1 only for JWKS x5t as required by spec for thumbprint.
	"crypto/sha1" //nolint:gosec

	"github.com/asgardeo/thunder/internal/cert"
	"github.com/asgardeo/thunder/internal/oauth/jwks/constants"
	"github.com/asgardeo/thunder/internal/oauth/jwks/model"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

// JWKSServiceInterface defines the interface for JWKS service.
type JWKSServiceInterface interface {
	GetJWKS() (*model.JWKSResponse, *serviceerror.ServiceError)
}

// JWKSService implements the JWKSServiceInterface.
type JWKSService struct {
	SystemCertService cert.SystemCertificateServiceInterface
}

// NewJWKSService creates a new instance of JWKSService.
func NewJWKSService() JWKSServiceInterface {
	return &JWKSService{
		SystemCertService: cert.NewSystemCertificateService(),
	}
}

// GetJWKS retrieves the JSON Web Key Set (JWKS) from the server's TLS certificate.
func (s *JWKSService) GetJWKS() (*model.JWKSResponse, *serviceerror.ServiceError) {
	thunderRuntime := config.GetThunderRuntime()

	// Get the certificate kid using the common utility function
	kid, err := s.SystemCertService.GetCertificateKid()
	if err != nil {
		svcErr := constants.ErrorWhileRetrievingCertificateKid
		svcErr.ErrorDescription = err.Error()
		return nil, svcErr
	}

	tlsConfig, err := s.SystemCertService.GetTLSConfig(&thunderRuntime.Config, thunderRuntime.ThunderHome)
	if err != nil {
		svcErr := constants.ErrorWhileRetrievingTLSConfig
		svcErr.ErrorDescription = err.Error()
		return nil, svcErr
	}

	if len(tlsConfig.Certificates) == 0 || len(tlsConfig.Certificates[0].Certificate) == 0 {
		return nil, constants.ErrorNoCertificateFound
	}

	certData := tlsConfig.Certificates[0].Certificate[0]
	parsedCert, err := x509.ParseCertificate(certData)
	if err != nil {
		svcErr := constants.ErrorWhileParsingCertificate
		svcErr.ErrorDescription = err.Error()
		return nil, svcErr
	}

	// x5c: base64 DER encoding
	x5c := []string{base64.StdEncoding.EncodeToString(parsedCert.Raw)}

	// x5t: SHA-1 thumbprint, x5t#S256: SHA-256 thumbprint
	sha1Sum := sha1.Sum(parsedCert.Raw) //nolint:gosec // x5t (SHA-1 thumbprint) is required by spec
	x5t := base64.StdEncoding.EncodeToString(sha1Sum[:])
	h := sha256.New()
	h.Write(parsedCert.Raw)
	x5tS256 := base64.StdEncoding.EncodeToString(h.Sum(nil))

	var jwks model.JWKS
	switch pub := parsedCert.PublicKey.(type) {
	case *rsa.PublicKey:
		encodeBase64URL := func(b []byte) string {
			return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
		}

		n := encodeBase64URL(pub.N.Bytes())
		// Properly encode the exponent as a big-endian byte slice, trimmed of leading zeros
		eBytes := make([]byte, 0, 8)
		e := pub.E
		for e > 0 {
			eBytes = append([]byte{byte(e & 0xff)}, eBytes...)
			e >>= 8
		}
		if len(eBytes) == 0 {
			eBytes = []byte{0}
		}
		eEnc := encodeBase64URL(eBytes)

		jwks = model.JWKS{
			Kid:     kid,
			Kty:     "RSA",
			Use:     "sig",
			Alg:     "RS256",
			N:       n,
			E:       eEnc,
			X5c:     x5c,
			X5t:     x5t,
			X5tS256: x5tS256,
		}
	case *ecdsa.PublicKey:
		encodeBase64URL := func(b []byte) string {
			return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
		}

		crv := pub.Curve.Params().Name
		x := encodeBase64URL(pub.X.Bytes())
		y := encodeBase64URL(pub.Y.Bytes())

		alg := "ES256"
		switch crv {
		case "P-384":
			alg = "ES384"
		case "P-521":
			alg = "ES512"
		}

		jwks = model.JWKS{
			Kid:     kid,
			Kty:     "EC",
			Use:     "sig",
			Alg:     alg,
			Crv:     crv,
			X:       x,
			Y:       y,
			X5c:     x5c,
			X5t:     x5t,
			X5tS256: x5tS256,
		}
	default:
		return nil, constants.ErrorUnsupportedPublicKeyType
	}

	return &model.JWKSResponse{
		Keys: []model.JWKS{jwks},
	}, nil
}
