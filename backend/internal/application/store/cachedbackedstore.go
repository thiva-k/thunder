/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package store

import (
	"errors"

	"github.com/asgardeo/thunder/internal/application/constants"
	"github.com/asgardeo/thunder/internal/application/model"
	"github.com/asgardeo/thunder/internal/system/cache"
	"github.com/asgardeo/thunder/internal/system/log"
)

// CachedBackedApplicationStoreInterface defines the interface for the cached backed application store
// which provides methods to manage applications with caching capabilities.
type CachedBackedApplicationStoreInterface interface {
	CreateApplication(app model.ApplicationProcessedDTO) error
	GetTotalApplicationCount() (int, error)
	GetApplicationList() ([]model.BasicApplicationDTO, error)
	GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessed, error)
	GetApplicationByID(id string) (*model.ApplicationProcessedDTO, error)
	GetApplicationByName(name string) (*model.ApplicationProcessedDTO, error)
	UpdateApplication(existingApp, updatedApp *model.ApplicationProcessedDTO) error
	DeleteApplication(id string) error
}

// CachedBackedApplicationStore is the implementation of CachedBackedApplicationStoreInterface.
type CachedBackedApplicationStore struct {
	AppByIDCacheManager   cache.CacheManagerInterface[*model.ApplicationProcessedDTO]
	AppByNameCacheManager cache.CacheManagerInterface[*model.ApplicationProcessedDTO]
	OAuthAppCacheManager  cache.CacheManagerInterface[*model.OAuthAppConfigProcessed]
	Store                 ApplicationStoreInterface
}

// NewCachedBackedApplicationStore creates a new instance of CachedBackedApplicationStore.
func NewCachedBackedApplicationStore() CachedBackedApplicationStoreInterface {
	return &CachedBackedApplicationStore{
		AppByIDCacheManager:   cache.GetCacheManager[*model.ApplicationProcessedDTO]("ApplicationByIDCache"),
		AppByNameCacheManager: cache.GetCacheManager[*model.ApplicationProcessedDTO]("ApplicationByNameCache"),
		OAuthAppCacheManager:  cache.GetCacheManager[*model.OAuthAppConfigProcessed]("OAuthAppCache"),
		Store:                 NewApplicationStore(),
	}
}

// CreateApplication creates a new application and caches it.
func (as *CachedBackedApplicationStore) CreateApplication(app model.ApplicationProcessedDTO) error {
	if err := as.Store.CreateApplication(app); err != nil {
		return err
	}
	as.cacheApplication(&app)
	return nil
}

// GetTotalApplicationCount returns the total count of applications.
func (as *CachedBackedApplicationStore) GetTotalApplicationCount() (int, error) {
	return as.Store.GetTotalApplicationCount()
}

// GetApplicationList returns a list of basic application DTOs.
func (as *CachedBackedApplicationStore) GetApplicationList() ([]model.BasicApplicationDTO, error) {
	return as.Store.GetApplicationList()
}

// GetOAuthApplication retrieves an OAuth application by client ID, using cache if available.
func (as *CachedBackedApplicationStore) GetOAuthApplication(clientID string) (*model.OAuthAppConfigProcessed, error) {
	cacheKey := cache.CacheKey{
		Key: clientID,
	}
	cachedApp, ok := as.OAuthAppCacheManager.Get(cacheKey)
	if ok {
		return cachedApp, nil
	}

	oauthApp, err := as.Store.GetOAuthApplication(clientID)
	if err != nil || oauthApp == nil {
		return oauthApp, err
	}
	as.cacheOAuthApplication(oauthApp)

	return oauthApp, nil
}

// GetApplicationByID retrieves an application by ID, using cache if available.
func (as *CachedBackedApplicationStore) GetApplicationByID(id string) (*model.ApplicationProcessedDTO, error) {
	cacheKey := cache.CacheKey{
		Key: id,
	}
	cachedApp, ok := as.AppByIDCacheManager.Get(cacheKey)
	if ok {
		return cachedApp, nil
	}

	app, err := as.Store.GetApplicationByID(id)
	if err != nil || app == nil {
		return app, err
	}
	as.cacheApplication(app)

	return app, nil
}

// GetApplicationByName retrieves an application by name, using cache if available.
func (as *CachedBackedApplicationStore) GetApplicationByName(name string) (*model.ApplicationProcessedDTO, error) {
	cacheKey := cache.CacheKey{
		Key: name,
	}
	cachedApp, ok := as.AppByNameCacheManager.Get(cacheKey)
	if ok {
		return cachedApp, nil
	}

	app, err := as.Store.GetApplicationByName(name)
	if err != nil || app == nil {
		return app, err
	}
	as.cacheApplication(app)

	return app, nil
}

// UpdateApplication updates an existing application and caches the updated version.
func (as *CachedBackedApplicationStore) UpdateApplication(existingApp,
	updatedApp *model.ApplicationProcessedDTO) error {
	if err := as.Store.UpdateApplication(updatedApp); err != nil {
		return err
	}

	existingAppClientID := ""
	if len(existingApp.InboundAuthConfig) > 0 && existingApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		existingAppClientID = existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}
	as.invalidateApplicationCache(existingApp.ID, existingApp.Name, existingAppClientID)

	as.cacheApplication(updatedApp)

	if len(updatedApp.InboundAuthConfig) > 0 && updatedApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		as.cacheOAuthApplication(updatedApp.InboundAuthConfig[0].OAuthAppConfig)
	}

	return nil
}

// DeleteApplication deletes an application by ID and invalidates the caches.
func (as *CachedBackedApplicationStore) DeleteApplication(id string) error {
	cacheKey := cache.CacheKey{
		Key: id,
	}
	existingApp, ok := as.AppByIDCacheManager.Get(cacheKey)
	if !ok {
		var err error
		existingApp, err = as.Store.GetApplicationByID(id)
		if err != nil {
			if errors.Is(err, constants.ApplicationNotFoundError) {
				return nil
			}
			return err
		}
	}
	if existingApp == nil {
		return nil
	}

	if err := as.Store.DeleteApplication(id); err != nil {
		return err
	}

	existingAppClientID := ""
	if len(existingApp.InboundAuthConfig) > 0 && existingApp.InboundAuthConfig[0].OAuthAppConfig != nil {
		existingAppClientID = existingApp.InboundAuthConfig[0].OAuthAppConfig.ClientID
	}
	as.invalidateApplicationCache(id, existingApp.Name, existingAppClientID)

	return nil
}

// cacheApplication caches the application and OAuth application configuration if it exists.
func (as *CachedBackedApplicationStore) cacheApplication(app *model.ApplicationProcessedDTO) {
	if app == nil {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))

	if app.ID != "" {
		appByIDCacheKey := cache.CacheKey{
			Key: app.ID,
		}
		if err := as.AppByIDCacheManager.Set(appByIDCacheKey, app); err != nil {
			logger.Error("Failed to cache application by ID", log.String("appID", app.ID), log.Error(err))
		}
	}
	if app.Name != "" {
		appByNameCacheKey := cache.CacheKey{
			Key: app.Name,
		}
		if err := as.AppByNameCacheManager.Set(appByNameCacheKey, app); err != nil {
			logger.Error("Failed to cache application by name", log.String("appName", app.Name),
				log.Error(err))
		}
	}

	// Cache the OAuth application configuration if it exists.
	if len(app.InboundAuthConfig) > 0 && app.InboundAuthConfig[0].OAuthAppConfig != nil {
		as.cacheOAuthApplication(app.InboundAuthConfig[0].OAuthAppConfig)
	}
}

// cacheOAuthApplication caches the OAuth application configuration if it exists.
func (as *CachedBackedApplicationStore) cacheOAuthApplication(oAuthAppConfig *model.OAuthAppConfigProcessed) {
	if oAuthAppConfig == nil || oAuthAppConfig.ClientID == "" {
		return
	}
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))

	oauthCacheKey := cache.CacheKey{
		Key: oAuthAppConfig.ClientID,
	}
	if err := as.OAuthAppCacheManager.Set(oauthCacheKey, oAuthAppConfig); err != nil {
		logger.Error("Failed to cache OAuth application", log.String("clientID", oAuthAppConfig.ClientID),
			log.Error(err))
	}
}

// InvalidateApplicationCache invalidates all application caches.
func (as *CachedBackedApplicationStore) invalidateApplicationCache(appID, appName, clientID string) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CachedBackedApplicationStore"))

	if appID != "" {
		cacheKey := cache.CacheKey{
			Key: appID,
		}
		err := as.AppByIDCacheManager.Delete(cacheKey)
		if err != nil {
			logger.Error("Failed to delete application cache by ID", log.String("appID", appID), log.Error(err))
		}
	}
	if appName != "" {
		cacheKey := cache.CacheKey{
			Key: appName,
		}
		err := as.AppByNameCacheManager.Delete(cacheKey)
		if err != nil {
			logger.Error("Failed to delete application cache by name", log.String("appName", appName), log.Error(err))
		}
	}
	if clientID != "" {
		oauthCacheKey := cache.CacheKey{
			Key: clientID,
		}
		err := as.OAuthAppCacheManager.Delete(oauthCacheKey)
		if err != nil {
			logger.Error("Failed to delete OAuth application cache", log.String("clientID", clientID), log.Error(err))
		}
	}
}
