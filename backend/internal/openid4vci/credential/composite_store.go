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

package credential

import (
	"context"
	"errors"

	serverconst "github.com/thunder-id/thunderid/internal/system/constants"
	declarativeresource "github.com/thunder-id/thunderid/internal/system/declarative_resource"
)

// compositeConfigurationStore combines a file-based (immutable) and a database
// (mutable) store: reads merge both, writes route to the database, and declarative
// configurations cannot be modified or deleted.
type compositeConfigurationStore struct {
	fileStore configurationStoreInterface
	dbStore   configurationStoreInterface
}

func newCompositeConfigurationStore(fileStore, dbStore configurationStoreInterface) *compositeConfigurationStore {
	return &compositeConfigurationStore{fileStore: fileStore, dbStore: dbStore}
}

// Create creates a new credential configuration in the database store only.
func (c *compositeConfigurationStore) Create(ctx context.Context, dto CredentialConfigurationDTO) error {
	return c.dbStore.Create(ctx, dto)
}

// GetByID retrieves a credential configuration, checking the database then the file store.
func (c *compositeConfigurationStore) GetByID(ctx context.Context, id string) (*CredentialConfigurationDTO, error) {
	return declarativeresource.CompositeGetHelper(
		func() (*CredentialConfigurationDTO, error) { return c.dbStore.GetByID(ctx, id) },
		func() (*CredentialConfigurationDTO, error) { return c.fileStore.GetByID(ctx, id) },
		ErrNotFound,
	)
}

// GetByHandle retrieves a credential configuration, checking the database then the file store.
func (c *compositeConfigurationStore) GetByHandle(
	ctx context.Context, handle string,
) (*CredentialConfigurationDTO, error) {
	return declarativeresource.CompositeGetHelper(
		func() (*CredentialConfigurationDTO, error) { return c.dbStore.GetByHandle(ctx, handle) },
		func() (*CredentialConfigurationDTO, error) { return c.fileStore.GetByHandle(ctx, handle) },
		ErrNotFound,
	)
}

// ListSummaries retrieves minimal listing data from both stores and merges the results.
func (c *compositeConfigurationStore) ListSummaries(ctx context.Context) ([]CredentialConfigurationSummary, error) {
	dbSummaries, err := c.dbStore.ListSummaries(ctx)
	if err != nil {
		return nil, err
	}
	fileSummaries, err := c.fileStore.ListSummaries(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(dbSummaries))
	result := make([]CredentialConfigurationSummary, 0, len(dbSummaries)+len(fileSummaries))
	for _, s := range dbSummaries {
		if !seen[s.ID] {
			seen[s.ID] = true
			result = append(result, s)
		}
	}
	for _, s := range fileSummaries {
		if !seen[s.ID] {
			seen[s.ID] = true
			result = append(result, s)
		}
	}
	if len(result) > serverconst.MaxCompositeStoreRecords {
		return nil, ErrResultLimitExceededInCompositeMode
	}
	return result, nil
}

// List retrieves credential configurations from both stores and merges the results.
func (c *compositeConfigurationStore) List(ctx context.Context) ([]CredentialConfigurationDTO, error) {
	dbConfigs, err := c.dbStore.List(ctx)
	if err != nil {
		return nil, err
	}
	fileConfigs, err := c.fileStore.List(ctx)
	if err != nil {
		return nil, err
	}

	configs, limitExceeded, err := declarativeresource.CompositeMergeListHelperWithLimit(
		func() (int, error) { return len(dbConfigs), nil },
		func() (int, error) { return len(fileConfigs), nil },
		func(int) ([]CredentialConfigurationDTO, error) { return dbConfigs, nil },
		func(int) ([]CredentialConfigurationDTO, error) { return fileConfigs, nil },
		mergeAndDeduplicate,
		len(dbConfigs)+len(fileConfigs),
		0,
		serverconst.MaxCompositeStoreRecords,
	)
	if err != nil {
		return nil, err
	}
	if limitExceeded {
		return nil, ErrResultLimitExceededInCompositeMode
	}
	return configs, nil
}

// Update updates a credential configuration in the database store only, returning
// ErrConfigurationIsImmutable when the configuration is declarative (file-based).
func (c *compositeConfigurationStore) Update(ctx context.Context, dto CredentialConfigurationDTO) error {
	return declarativeresource.CompositeUpdateHelper(
		dto,
		func(d CredentialConfigurationDTO) string { return d.ID },
		func(id string) (bool, error) { return c.existsInFileStore(ctx, id) },
		func(d CredentialConfigurationDTO) error { return c.dbStore.Update(ctx, d) },
		ErrConfigurationIsImmutable,
	)
}

// Delete deletes a credential configuration from the database store only, returning
// ErrConfigurationIsImmutable when the configuration is declarative (file-based).
func (c *compositeConfigurationStore) Delete(ctx context.Context, id string) error {
	return declarativeresource.CompositeDeleteHelper(
		id,
		func(id string) (bool, error) { return c.existsInFileStore(ctx, id) },
		func(id string) error { return c.dbStore.Delete(ctx, id) },
		ErrConfigurationIsImmutable,
	)
}

// IsDeclarative reports whether the credential configuration is file-based (immutable).
func (c *compositeConfigurationStore) IsDeclarative(ctx context.Context, id string) (bool, error) {
	return c.existsInFileStore(ctx, id)
}

// existsInFileStore reports whether the id exists in the immutable file store.
func (c *compositeConfigurationStore) existsInFileStore(ctx context.Context, id string) (bool, error) {
	_, err := c.fileStore.GetByID(ctx, id)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return false, err
}

// mergeAndDeduplicate merges configurations from both stores, removing duplicate IDs
// with database entries taking precedence.
func mergeAndDeduplicate(dbConfigs, fileConfigs []CredentialConfigurationDTO) []CredentialConfigurationDTO {
	seen := make(map[string]bool, len(dbConfigs))
	result := make([]CredentialConfigurationDTO, 0, len(dbConfigs)+len(fileConfigs))
	for i := range dbConfigs {
		if !seen[dbConfigs[i].ID] {
			seen[dbConfigs[i].ID] = true
			result = append(result, dbConfigs[i])
		}
	}
	for i := range fileConfigs {
		if !seen[fileConfigs[i].ID] {
			seen[fileConfigs[i].ID] = true
			result = append(result, fileConfigs[i])
		}
	}
	return result
}
