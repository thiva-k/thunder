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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/thunder-id/thunderid/internal/system/config"
	"github.com/thunder-id/thunderid/internal/system/database/provider"
)

// dbNonceStore persists c_nonces in the runtime database so nonce validation
// works across replicas (the replica that issues the nonce may differ from the
// one that validates it in the credential request).
type dbNonceStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

func newDBNonceStore() nonceStore {
	return &dbNonceStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetServerRuntime().Config.Server.Identifier,
	}
}

func (s *dbNonceStore) Save(ctx context.Context, rec *nonceRecord) error {
	dbClient, err := s.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return fmt.Errorf("failed to get runtime database client: %w", err)
	}
	if _, err = dbClient.ExecuteContext(ctx, queryInsertNonce,
		rec.Nonce, s.deploymentID, rec.ExpiresAt.UTC()); err != nil {
		return fmt.Errorf("failed to insert nonce: %w", err)
	}
	return nil
}

func (s *dbNonceStore) Get(ctx context.Context, nonce string) (*nonceRecord, bool) {
	dbClient, err := s.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return nil, false
	}
	results, err := dbClient.QueryContext(ctx, queryGetNonce, nonce, s.deploymentID)
	if err != nil || len(results) == 0 {
		return nil, false
	}
	row := results[0]
	expiry, err := parseVCITime(row["expiry_time"])
	if err != nil {
		return nil, false
	}
	return &nonceRecord{
		Nonce:     vciColumnString(row["nonce"]),
		ExpiresAt: expiry,
	}, true
}

func (s *dbNonceStore) Delete(ctx context.Context, nonce string) error {
	dbClient, err := s.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return fmt.Errorf("failed to get runtime database client: %w", err)
	}
	if _, err = dbClient.ExecuteContext(ctx, queryDeleteNonce, nonce, s.deploymentID); err != nil {
		return fmt.Errorf("failed to delete nonce: %w", err)
	}
	return nil
}

// dbOfferStore persists issuer-initiated credential offers in the runtime
// database so wallets can retrieve them from any replica.
type dbOfferStore struct {
	dbProvider   provider.DBProviderInterface
	deploymentID string
}

func newDBOfferStore() offerStore {
	return &dbOfferStore{
		dbProvider:   provider.GetDBProvider(),
		deploymentID: config.GetServerRuntime().Config.Server.Identifier,
	}
}

func (s *dbOfferStore) Save(ctx context.Context, rec *offerRecord) error {
	dbClient, err := s.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return fmt.Errorf("failed to get runtime database client: %w", err)
	}
	offerJSON, err := json.Marshal(rec.Offer)
	if err != nil {
		return fmt.Errorf("failed to marshal credential offer: %w", err)
	}
	if _, err = dbClient.ExecuteContext(ctx, queryInsertOffer,
		rec.ID, s.deploymentID, string(offerJSON), rec.ExpiresAt.UTC()); err != nil {
		return fmt.Errorf("failed to insert credential offer: %w", err)
	}
	return nil
}

func (s *dbOfferStore) Get(ctx context.Context, id string) (*offerRecord, bool) {
	dbClient, err := s.dbProvider.GetRuntimeDBClient()
	if err != nil {
		return nil, false
	}
	results, err := dbClient.QueryContext(ctx, queryGetOffer, id, s.deploymentID)
	if err != nil || len(results) == 0 {
		return nil, false
	}
	row := results[0]
	offerBytes := vciColumnBytes(row["offer"])
	var offerMap map[string]interface{}
	if err = json.Unmarshal(offerBytes, &offerMap); err != nil {
		return nil, false
	}
	expiry, err := parseVCITime(row["expiry_time"])
	if err != nil {
		return nil, false
	}
	return &offerRecord{
		ID:        vciColumnString(row["id"]),
		Offer:     offerMap,
		ExpiresAt: expiry,
	}, true
}

// vciColumnString coerces a result-row value to a string, tolerating string/[]byte.
func vciColumnString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

// vciColumnBytes coerces a result-row value to bytes, tolerating []byte/string.
func vciColumnBytes(v interface{}) []byte {
	switch t := v.(type) {
	case []byte:
		return t
	case string:
		return []byte(t)
	default:
		return nil
	}
}

// parseVCITime parses an EXPIRY_TIME column across Postgres (time.Time) and
// SQLite (datetime string) drivers.
func parseVCITime(field interface{}) (time.Time, error) {
	const layout = "2006-01-02 15:04:05.999999999"
	switch v := field.(type) {
	case time.Time:
		return v, nil
	case []byte:
		return parseVCITime(string(v))
	case string:
		trimmed := v
		if parts := strings.SplitN(v, " ", 3); len(parts) >= 2 {
			trimmed = parts[0] + " " + parts[1]
		}
		if t, err := time.Parse(layout, trimmed); err == nil {
			return t, nil
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, nil
		}
		return time.Time{}, fmt.Errorf("error parsing expiry_time: %q", v)
	default:
		return time.Time{}, fmt.Errorf("unexpected type for expiry_time")
	}
}
