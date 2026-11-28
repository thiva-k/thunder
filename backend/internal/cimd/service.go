// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// Package cimd supports clients registered from an OAuth Client ID Metadata Document: it previews a
// document and applies the document rules to the values submitted for such a client.
package cimd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/thunder-id/thunderid/internal/cert"
	"github.com/thunder-id/thunderid/internal/serverconfig"
	"github.com/thunder-id/thunderid/internal/system/config"
	syshttp "github.com/thunder-id/thunderid/internal/system/http"
	"github.com/thunder-id/thunderid/internal/system/jose/jws"
	"github.com/thunder-id/thunderid/internal/system/log"
	sysutils "github.com/thunder-id/thunderid/internal/system/utils"
	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// CIMDServiceInterface previews Client ID Metadata Documents, applies the document rules to the values
// submitted for a CIMD application or agent, and resolves the clients its Client ID domains admit.
type CIMDServiceInterface interface {
	Preview(ctx context.Context, clientID string) (*PreviewResponse, *tidcommon.ServiceError)
	ValidateOAuthProfile(clientID string, profile *providers.OAuthProfile, hasClientSecret bool,
		existingClientID string, existing *providers.OAuthProfile) *tidcommon.ServiceError
	// ClientIDDomainCandidates returns the Client ID domains that would admit the client with the given Client
	// Identifier URL, most specific first, or nil when clients cannot sign in by domain or the URL is blocked.
	ClientIDDomainCandidates(ctx context.Context, clientID string) ([]string, *tidcommon.ServiceError)
	// ResolveDocumentClient retrieves the document of a client a Client ID domain admits and checks it against
	// the document rules. It returns nil when the document is unavailable or breaks a rule.
	ResolveDocumentClient(ctx context.Context, clientID string) (*DocumentClient, *tidcommon.ServiceError)
}

// cimdService is the default implementation of CIMDServiceInterface.
type cimdService struct {
	httpClient   syshttp.HTTPClientInterface
	configReader serverconfig.ServerConfigService
	documents    *documentCache
	logger       *log.Logger
}

// newCIMDService creates a new instance of cimdService. Documents are retrieved by a client that
// refuses private and loopback addresses on the connected address and follows no redirects.
func newCIMDService(configReader serverconfig.ServerConfigService) CIMDServiceInterface {
	service := &cimdService{
		httpClient: syshttp.NewHTTPClientWithCheckRedirect(func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}),
		configReader: configReader,
		logger:       log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CIMDService")),
	}
	service.documents = newDocumentCache(func(ctx context.Context, clientID string) (
		*cimdDocument, time.Duration, error) {
		doc, ttl, err := fetchCIMDDocument(ctx, service.httpClient, clientID)
		if err != nil {
			return nil, 0, err
		}
		if err := resolveDocumentKeys(ctx, service.httpClient, doc); err != nil {
			return nil, 0, err
		}
		return doc, ttl, nil
	})
	return service
}

// Preview retrieves and validates the Client ID Metadata Document at clientID and returns the values
// to store. It stores nothing.
func (s *cimdService) Preview(ctx context.Context, clientID string) (
	*PreviewResponse, *tidcommon.ServiceError) {
	clientIDURL, err := parseClientIdentifierURL(clientID)
	if err != nil {
		return nil, &ErrorInvalidClientID
	}

	doc, _, err := fetchCIMDDocument(ctx, s.httpClient, clientID)
	if err != nil {
		s.logger.Debug(ctx, "Failed to retrieve Client ID Metadata Document",
			log.String("clientID", clientID), log.Error(err))
		return nil, &ErrorDocumentUnavailable
	}
	oauthConfig, svcErr := documentOAuthConfig(clientID, doc)
	if svcErr != nil {
		return nil, svcErr
	}

	// The name and links are editable defaults, so a value the application would reject is shortened or
	// left out rather than failing the registration.
	name := doc.ClientName
	if name == "" {
		name = clientIDURL.Hostname()
	}
	if runes := []rune(name); len(runes) > maxApplicationNameLength {
		name = string(runes[:maxApplicationNameLength])
	}
	return &PreviewResponse{
		Name:      name,
		URL:       absoluteHTTPURL(doc.ClientURI),
		TosURI:    absoluteHTTPURL(doc.TosURI),
		PolicyURI: absoluteHTTPURL(doc.PolicyURI),
		Contacts:  doc.Contacts,
		InboundAuthConfig: []providers.InboundAuthConfigWithSecret{
			{Type: providers.OAuthInboundAuthType, OAuthConfig: oauthConfig},
		},
	}, nil
}

// ClientIDDomainCandidates returns the Client ID domains that would admit the client with the given Client
// Identifier URL, most specific first: its host, then a wildcard for each parent domain of at least two labels.
// It returns nil when clients cannot sign in by domain: Client ID Metadata Documents are off for the
// deployment, unregistered clients are turned off, the URL is blocked, or clientID is not a Client Identifier
// URL.
func (s *cimdService) ClientIDDomainCandidates(ctx context.Context, clientID string) (
	[]string, *tidcommon.ServiceError) {
	if !config.GetServerRuntime().Config.OAuth.CIMD.IsEnabled() {
		return nil, nil
	}
	clientIDURL, err := parseClientIdentifierURL(clientID)
	if err != nil {
		return nil, nil
	}
	settings, svcErr := s.loadSettings(ctx)
	if svcErr != nil {
		return nil, svcErr
	}
	if !settings.Enabled || slices.Contains(settings.BlockedClientIDs, clientID) {
		return nil, nil
	}
	return domainCandidates(clientIDURL.Hostname()), nil
}

// ResolveDocumentClient retrieves the document of a client a Client ID domain admits and checks it against
// the document rules. It returns nil when clientID is not a Client Identifier URL, or the document is
// unavailable or breaks a rule.
func (s *cimdService) ResolveDocumentClient(ctx context.Context, clientID string) (
	*DocumentClient, *tidcommon.ServiceError) {
	clientIDURL, err := parseClientIdentifierURL(clientID)
	if err != nil {
		return nil, nil
	}
	doc, err := s.documents.get(ctx, clientID)
	if err != nil {
		s.logger.Debug(ctx, "Client ID Metadata Document of a client a domain admits is unavailable",
			log.String("clientID", clientID), log.Error(err))
		return nil, nil
	}
	oauthConfig, ruleErr := documentOAuthConfig(clientID, doc)
	if ruleErr != nil {
		s.logger.Debug(ctx, "Client ID Metadata Document of a client a domain admits breaks a document rule",
			log.String("clientID", clientID), log.String("errorCode", ruleErr.Code))
		return nil, nil
	}
	// The client authenticates with the keys retrieved with its document, never by having the token endpoint
	// retrieve a URL the document names.
	if oauthConfig.Certificate != nil && oauthConfig.Certificate.Type == cert.CertificateTypeJWKSURI {
		if doc.resolvedJWKS == "" {
			return nil, nil
		}
		oauthConfig.Certificate = &providers.Certificate{Type: cert.CertificateTypeJWKS, Value: doc.resolvedJWKS}
	}
	return newDocumentClient(clientIDURL, doc, oauthConfig), nil
}

// loadSettings reads the effective cimd server configuration section.
func (s *cimdService) loadSettings(ctx context.Context) (unregisteredClientSettings, *tidcommon.ServiceError) {
	if s.configReader == nil {
		return unregisteredClientSettings{}, nil
	}
	merged, svcErr := s.configReader.GetMergedConfig(ctx, string(serverconfig.ConfigNameCIMD))
	if svcErr != nil {
		s.logger.Error(ctx, "Failed to read the cimd server configuration", log.String("errorCode", svcErr.Code))
		return unregisteredClientSettings{}, &tidcommon.InternalServerError
	}
	settings, ok := merged.(unregisteredClientSettings)
	if !ok {
		s.logger.Error(ctx, "Unexpected type of the cimd server configuration",
			log.String("type", fmt.Sprintf("%T", merged)))
		return unregisteredClientSettings{}, &tidcommon.InternalServerError
	}
	return settings, nil
}

// newDocumentClient builds the client a document describes: its checked OAuth configuration, the grants it
// declares, the post-logout redirect URIs that pass the redirect URI rules, and its name, logo and links.
func newDocumentClient(clientIDURL *url.URL, doc *cimdDocument,
	oauthConfig *providers.OAuthConfigWithSecret) *DocumentClient {
	client := &DocumentClient{
		OAuthConfig:        oauthConfig,
		DeclaredGrantTypes: doc.GrantTypes,
		DeclaredScopes:     strings.Fields(doc.Scope),
		Name:               strings.TrimSpace(doc.ClientName),
		Properties:         map[string]interface{}{},
	}
	if len(client.DeclaredGrantTypes) == 0 {
		client.DeclaredGrantTypes = []string{string(providers.GrantTypeAuthorizationCode)}
	}
	for _, redirectURI := range doc.PostLogoutRedirectURIs {
		if isCIMDRedirectURI(redirectURI, clientIDURL) &&
			(doc.allowsLoopbackRedirects() || !isLoopbackRedirectURI(redirectURI)) {
			client.PostLogoutRedirectURIs = append(client.PostLogoutRedirectURIs, redirectURI)
		}
	}
	if client.Name == "" {
		client.Name = clientIDURL.Hostname()
	}
	if runes := []rune(client.Name); len(runes) > maxApplicationNameLength {
		client.Name = string(runes[:maxApplicationNameLength])
	}
	for key, value := range map[string]string{
		propertyLogoURL:   doc.LogoURI,
		propertyURL:       doc.ClientURI,
		propertyTosURI:    doc.TosURI,
		propertyPolicyURI: doc.PolicyURI,
	} {
		if link := absoluteHTTPURL(value); link != "" {
			client.Properties[key] = link
		}
	}
	return client
}

// Apply makes the OAuth client of the application or agent whose Client ID domains admit this client into
// this client's own: the client ID, redirect URIs, authentication method and keys of its document, and only the
// grants that both the application or agent holds and the document declares.
func (d *DocumentClient) Apply(client *providers.OAuthClient) {
	method := d.OAuthConfig.TokenEndpointAuthMethod
	grantTypes := make([]providers.GrantType, 0, len(client.GrantTypes))
	for _, grantType := range client.GrantTypes {
		if slices.Contains(d.DeclaredGrantTypes, string(grantType)) && isCIMDGrantAllowed(grantType, method) {
			grantTypes = append(grantTypes, grantType)
		}
	}
	client.ClientID = d.OAuthConfig.ClientID
	client.RedirectURIs = slices.Clone(d.OAuthConfig.RedirectURIs)
	client.Scopes = slices.Clone(d.DeclaredScopes)
	client.PostLogoutRedirectURIs = slices.Clone(d.PostLogoutRedirectURIs)
	client.GrantTypes = grantTypes
	client.TokenEndpointAuthMethod = method
	client.PublicClient = d.OAuthConfig.PublicClient
	client.Certificate = d.OAuthConfig.Certificate
	client.PKCERequired = true
	client.ClientIDMetadataDocument = true
	client.ResolvedByClientIDDomain = true
}

// domainCandidates returns the Client ID domains that match host, most specific first: host itself, then
// "*." and each parent domain of at least two labels, longest first.
func domainCandidates(host string) []string {
	host = strings.ToLower(host)
	candidates := []string{host}
	labels := strings.Split(host, ".")
	for i := 1; i <= len(labels)-2; i++ {
		candidates = append(candidates, wildcardDomainPrefix+strings.Join(labels[i:], "."))
	}
	return candidates
}

// validateDomainProfile checks an OAuth profile that admits clients by the domain of their Client ID
// Metadata Document URL: it has no values the admitted clients take from their documents, and its domains
// are valid and distinct.
func validateDomainProfile(clientID string, p *providers.OAuthProfile, hasClientSecret bool) *tidcommon.ServiceError {
	if !p.ClientIDMetadataDocument {
		return &ErrorClientIDDomainsWithoutDocument
	}
	if clientID != "" || hasClientSecret || len(p.RedirectURIs) > 0 || len(p.PostLogoutRedirectURIs) > 0 ||
		p.Certificate != nil || p.TokenEndpointAuthMethod != string(providers.TokenEndpointAuthMethodNone) ||
		!p.PublicClient || !p.PKCERequired {
		return &ErrorDomainClientValues
	}
	if len(p.ClientIDDomains) > maxClientIDDomains {
		return &ErrorInvalidClientIDDomain
	}
	seen := make(map[string]bool, len(p.ClientIDDomains))
	for _, domain := range p.ClientIDDomains {
		if !isValidClientIDDomain(domain) {
			return ErrorInvalidClientIDDomainName.WithParams(map[string]string{"domain": domain})
		}
		if seen[domain] {
			return &ErrorInvalidClientIDDomain
		}
		seen[domain] = true
	}
	if !slices.Contains(p.GrantTypes, string(providers.GrantTypeAuthorizationCode)) {
		return &ErrorInvalidGrantTypes
	}
	for _, grantType := range p.GrantTypes {
		if !isCIMDGrantAllowed(providers.GrantType(grantType), providers.TokenEndpointAuthMethodPrivateKeyJWT) {
			return &ErrorInvalidGrantTypes
		}
	}
	if slices.Contains(p.GrantTypes, string(providers.GrantTypeTokenExchange)) &&
		(p.Token == nil || p.Token.IDJAG == nil || !p.Token.IDJAG.Enabled) {
		return &ErrorTokenExchangeRequiresIDJAG
	}
	return nil
}

// isValidClientIDDomain reports whether domain is a lowercase DNS name of at least two labels, optionally
// starting with "*." to match its subdomains. IP addresses are refused, since a Client Identifier URL never has
// one.
func isValidClientIDDomain(domain string) bool {
	name := strings.TrimPrefix(domain, wildcardDomainPrefix)
	if name == "" || len(name) > maxDomainLength || net.ParseIP(name) != nil {
		return false
	}
	labels := strings.Split(name, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if !isValidDomainLabel(label) {
			return false
		}
	}
	return true
}

// isValidDomainLabel reports whether label is a lowercase DNS label: letters, digits and inner hyphens.
func isValidDomainLabel(label string) bool {
	if label == "" || len(label) > maxDomainLabelLength || label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for _, r := range label {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return true
}

// documentOAuthConfig checks a retrieved document against the document rules and maps it to the OAuth
// configuration of its client.
func documentOAuthConfig(clientID string, doc *cimdDocument) (
	*providers.OAuthConfigWithSecret, *tidcommon.ServiceError) {
	if doc.ClientID != clientID {
		return nil, &ErrorClientIDMismatch
	}
	if doc.ClientSecret != "" {
		return nil, &ErrorClientSecretNotAllowed
	}
	// RFC 7591 defaults response_types to code, so only a document that lists them must include it.
	if len(doc.ResponseTypes) > 0 && !slices.Contains(doc.ResponseTypes, string(providers.ResponseTypeCode)) {
		return nil, &ErrorInvalidResponseTypes
	}
	oauthConfig, svcErr := buildCIMDOAuthConfig(clientID, doc)
	if svcErr != nil {
		return nil, svcErr
	}
	if svcErr := validateCIMDRules(clientID, ruleProfile(oauthConfig), false); svcErr != nil {
		return nil, svcErr
	}
	if !doc.allowsLoopbackRedirects() && slices.ContainsFunc(oauthConfig.RedirectURIs, isLoopbackRedirectURI) {
		return nil, &ErrorInvalidRedirectURI
	}
	return oauthConfig, nil
}

// fetchCIMDDocument retrieves the document at clientID as JSON within the time and size limits, and
// returns how long its response lets it be kept.
func fetchCIMDDocument(ctx context.Context, client syshttp.HTTPClientInterface, clientID string) (
	*cimdDocument, time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, cimdFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, clientID, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return nil, 0, errors.New("document is not served as application/json")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, cimdMaxDocumentSize+1))
	if err != nil {
		return nil, 0, err
	}
	if len(body) > cimdMaxDocumentSize {
		return nil, 0, errors.New("document exceeds the size limit")
	}

	var doc cimdDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, 0, err
	}
	return &doc, documentTTL(resp.Header), nil
}

// resolveDocumentKeys retrieves the key set at the jwks_uri of a document that authenticates with private_key_jwt,
// within the time and size limits and with the protections of the document retrieval, and keeps it when it is a
// JSON key set of public keys.
func resolveDocumentKeys(ctx context.Context, client syshttp.HTTPClientInterface, doc *cimdDocument) error {
	if doc.TokenEndpointAuthMethod != string(providers.TokenEndpointAuthMethodPrivateKeyJWT) || doc.JWKSURI == "" {
		return nil
	}
	if err := syshttp.IsSSRFSafeURL(doc.JWKSURI); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, cimdFetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, doc.JWKSURI, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d retrieving the key set", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, cimdMaxJWKSSize+1))
	if err != nil {
		return err
	}
	if len(body) > cimdMaxJWKSSize {
		return errors.New("key set exceeds the size limit")
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, body); err != nil {
		return err
	}
	if !isPublicKeySet(compact.String()) {
		return errors.New("key set is not a set of public keys")
	}
	doc.resolvedJWKS = compact.String()
	return nil
}

// buildCIMDOAuthConfig maps a document to the OAuth configuration of a CIMD application. Grants are
// narrowed, never widened: the declared authorization_code and refresh_token grants are kept when the
// deployment allows them. Any other grant the document rules permit is added by an administrator.
func buildCIMDOAuthConfig(clientID string, doc *cimdDocument) (
	*providers.OAuthConfigWithSecret, *tidcommon.ServiceError) {
	method := providers.TokenEndpointAuthMethod(doc.TokenEndpointAuthMethod)
	if method == "" {
		// RFC 7591 default, which a CIMD client cannot use.
		method = providers.TokenEndpointAuthMethodClientSecretBasic
	}

	var certificate *providers.Certificate
	if method == providers.TokenEndpointAuthMethodPrivateKeyJWT {
		hasJWKS := len(doc.JWKS) > 0 && string(doc.JWKS) != "null"
		switch {
		case (doc.JWKSURI != "") == hasJWKS:
			return nil, &ErrorInvalidAuthMethod
		case doc.JWKSURI != "":
			certificate = &providers.Certificate{Type: cert.CertificateTypeJWKSURI, Value: doc.JWKSURI}
		default:
			var jwks bytes.Buffer
			if err := json.Compact(&jwks, doc.JWKS); err != nil {
				return nil, &ErrorInvalidAuthMethod
			}
			certificate = &providers.Certificate{Type: cert.CertificateTypeJWKS, Value: jwks.String()}
		}
	}

	declaredGrants := doc.GrantTypes
	if len(declaredGrants) == 0 {
		declaredGrants = []string{string(providers.GrantTypeAuthorizationCode)}
	}
	if !slices.Contains(declaredGrants, string(providers.GrantTypeAuthorizationCode)) {
		return nil, &ErrorInvalidGrantTypes
	}
	deploymentGrants := config.GetServerRuntime().Config.OAuth.AllowedGrantTypes
	var grantTypes []providers.GrantType
	for _, grantType := range documentGrantTypes {
		if slices.Contains(declaredGrants, string(grantType)) &&
			(len(deploymentGrants) == 0 || slices.Contains(deploymentGrants, string(grantType))) {
			grantTypes = append(grantTypes, grantType)
		}
	}

	return &providers.OAuthConfigWithSecret{
		ClientID:                 clientID,
		ClientIDMetadataDocument: true,
		RedirectURIs:             doc.RedirectURIs,
		GrantTypes:               grantTypes,
		ResponseTypes:            []providers.ResponseType{providers.ResponseTypeCode},
		TokenEndpointAuthMethod:  method,
		PublicClient:             method == providers.TokenEndpointAuthMethodNone,
		PKCERequired:             true,
		Certificate:              certificate,
	}, nil
}

// ValidateOAuthProfile applies the document rules to an OAuth profile marked as CIMD, and keeps the
// marker, the client ID, and whether the profile admits clients by domain fixed after creation. existing is
// the stored profile on an update, and nil on a create.
func (s *cimdService) ValidateOAuthProfile(clientID string, profile *providers.OAuthProfile,
	hasClientSecret bool, existingClientID string, existing *providers.OAuthProfile) *tidcommon.ServiceError {
	if existing != nil {
		existingByDomain, byDomain := len(existing.ClientIDDomains) > 0, len(profile.ClientIDDomains) > 0
		if existing.ClientIDMetadataDocument != profile.ClientIDMetadataDocument || existingByDomain != byDomain ||
			(existing.ClientIDMetadataDocument && !existingByDomain && existingClientID != clientID) {
			return &ErrorImmutable
		}
	}
	if len(profile.ClientIDDomains) > 0 {
		return validateDomainProfile(clientID, profile, hasClientSecret)
	}
	if !profile.ClientIDMetadataDocument {
		return nil
	}
	return validateCIMDRules(clientID, profile, hasClientSecret)
}

// validateCIMDRules checks the document rules that apply to both a previewed document and the values
// submitted for a CIMD client.
func validateCIMDRules(clientID string, p *providers.OAuthProfile, hasClientSecret bool) *tidcommon.ServiceError {
	clientIDURL, err := parseClientIdentifierURL(clientID)
	if err != nil {
		return &ErrorInvalidClientID
	}
	if len(p.RedirectURIs) == 0 {
		return &ErrorInvalidRedirectURI
	}
	for _, redirectURI := range p.RedirectURIs {
		if !isCIMDRedirectURI(redirectURI, clientIDURL) {
			return &ErrorInvalidRedirectURI
		}
	}
	method := providers.TokenEndpointAuthMethod(p.TokenEndpointAuthMethod)
	switch method {
	case providers.TokenEndpointAuthMethodNone:
	case providers.TokenEndpointAuthMethodPrivateKeyJWT:
		if p.Certificate == nil ||
			(p.Certificate.Type != cert.CertificateTypeJWKS && p.Certificate.Type != cert.CertificateTypeJWKSURI) {
			return &ErrorInvalidAuthMethod
		}
		// The same checks the certificate store applies when the client is created.
		switch {
		case p.Certificate.Type == cert.CertificateTypeJWKSURI && (syshttp.IsSSRFSafeURL(p.Certificate.Value) != nil ||
			!isValidCertificateValue(p.Certificate.Value)):
			return &ErrorInvalidJWKSURI
		case p.Certificate.Type == cert.CertificateTypeJWKS && !isValidCertificateValue(p.Certificate.Value):
			return &ErrorInvalidKeySet
		case p.Certificate.Type == cert.CertificateTypeJWKS && !isPublicKeySet(p.Certificate.Value):
			return &ErrorPrivateKeySet
		}
	default:
		return &ErrorInvalidAuthMethod
	}
	if hasClientSecret {
		return &ErrorClientSecretNotAllowed
	}
	if !slices.Contains(p.GrantTypes, string(providers.GrantTypeAuthorizationCode)) {
		return &ErrorInvalidGrantTypes
	}
	for _, grantType := range p.GrantTypes {
		if !isCIMDGrantAllowed(providers.GrantType(grantType), method) {
			return &ErrorInvalidGrantTypes
		}
	}
	if slices.Contains(p.GrantTypes, string(providers.GrantTypeTokenExchange)) &&
		(p.Token == nil || p.Token.IDJAG == nil || !p.Token.IDJAG.Enabled) {
		return &ErrorTokenExchangeRequiresIDJAG
	}
	// Every CIMD client uses the authorization code grant, so PKCE is required for confidential clients too.
	if !p.PKCERequired {
		return &ErrorPKCERequired
	}
	return nil
}

// isCIMDGrantAllowed reports whether a CIMD client that authenticates with method may hold grantType. The
// token exchange and JWT bearer grants are permitted only to a client that proves its identity with
// private_key_jwt.
func isCIMDGrantAllowed(grantType providers.GrantType, method providers.TokenEndpointAuthMethod) bool {
	switch {
	case slices.Contains(documentGrantTypes, grantType):
		return true
	case grantType == providers.GrantTypeTokenExchange, grantType == providers.GrantTypeJWTBearer:
		return method == providers.TokenEndpointAuthMethodPrivateKeyJWT
	default:
		return false
	}
}

// isPublicKeySet reports whether an inline key set parses as a JWK Set and holds no private or symmetric
// key material, which a published document must never carry.
func isPublicKeySet(value string) bool {
	var keySet struct {
		Keys []map[string]interface{} `json:"keys"`
	}
	if err := json.Unmarshal([]byte(value), &keySet); err != nil || len(keySet.Keys) == 0 {
		return false
	}
	for _, key := range keySet.Keys {
		if _, found := jws.ContainsPrivateMember(key); found {
			return false
		}
	}
	return true
}

// isValidCertificateValue reports whether a certificate value has a length the certificate store accepts.
func isValidCertificateValue(value string) bool {
	return len(value) >= minCertificateValueLength && len(value) <= maxCertificateValueLength
}

// ruleProfile returns the fields of a previewed configuration that the document rules read.
func ruleProfile(c *providers.OAuthConfigWithSecret) *providers.OAuthProfile {
	return &providers.OAuthProfile{
		RedirectURIs:            c.RedirectURIs,
		GrantTypes:              sysutils.ConvertToStringSlice(c.GrantTypes),
		TokenEndpointAuthMethod: string(c.TokenEndpointAuthMethod),
		PKCERequired:            c.PKCERequired,
		Certificate:             c.Certificate,
		Token:                   c.Token,
	}
}

// parseClientIdentifierURL parses a Client Identifier URL: an https URL with a DNS host and a path,
// with no user information, fragment or dot segments.
func parseClientIdentifierURL(raw string) (*url.URL, error) {
	if len(raw) > maxClientIdentifierURLLength {
		return nil, errors.New("client identifier URL is too long")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != schemeHTTPS || u.User != nil || strings.Contains(raw, "#") {
		return nil, errors.New("client identifier URL must be https with no user information or fragment")
	}
	if host := u.Hostname(); host == "" || net.ParseIP(host) != nil {
		return nil, errors.New("client identifier URL must have a DNS host")
	}
	if u.Path == "" || u.Path == "/" {
		return nil, errors.New("client identifier URL must have a path")
	}
	for _, segment := range strings.Split(u.Path, "/") {
		if segment == "." || segment == ".." {
			return nil, errors.New("client identifier URL must not contain dot segments")
		}
	}
	return u, nil
}

// isCIMDRedirectURI reports whether a redirect URI is https on the client ID's origin, or http on a
// loopback address.
func isCIMDRedirectURI(raw string, clientIDURL *url.URL) bool {
	// A document supplies exact redirect URIs only: no fragment, and no wildcard pattern even when the
	// deployment allows them.
	if strings.ContainsAny(raw, "#*") {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	switch u.Scheme {
	case schemeHTTPS:
		return strings.EqualFold(u.Host, clientIDURL.Host)
	case "http":
		return sysutils.IsLoopbackRedirectHost(u.Hostname())
	}
	return false
}

// isLoopbackRedirectURI reports whether a redirect URI is http on a loopback address, the only http redirect URI
// the document rules allow.
func isLoopbackRedirectURI(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == schemeHTTP
}

// allowsLoopbackRedirects reports whether the document's client may take a loopback redirect URI: any client
// but one that declares itself a web application.
func (doc *cimdDocument) allowsLoopbackRedirects() bool {
	return doc.ApplicationType != applicationTypeWeb
}

// absoluteHTTPURL returns raw when it is an absolute http or https URL, and an empty string otherwise.
func absoluteHTTPURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != schemeHTTPS && u.Scheme != "http") || u.Host == "" {
		return ""
	}
	return raw
}
