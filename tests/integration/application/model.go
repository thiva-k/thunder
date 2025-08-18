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

package application

type Application struct {
	ID                      string   `json:"id"`
	Name                    string   `json:"name"`
	Description             string   `json:"description"`
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	TokenEndpointAuthMethod []string `json:"token_endpoint_auth_method,omitempty"`
}

type ApplicationList struct {
	TotalResults int           `json:"totalResults"`
	Count        int           `json:"count"`
	Applications []Application `json:"applications"`
}

func compareStringSlices(a, b []string) bool {

	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (app *Application) equals(expectedApp Application) bool {

	return app.ID == expectedApp.ID && app.Name == expectedApp.Name && app.Description == expectedApp.Description &&
		app.ClientID == expectedApp.ClientID && compareStringSlices(app.RedirectURIs, expectedApp.RedirectURIs) &&
		compareStringSlices(app.GrantTypes, expectedApp.GrantTypes) &&
		compareStringSlices(app.ResponseTypes, expectedApp.ResponseTypes) &&
		compareStringSlices(app.TokenEndpointAuthMethod, expectedApp.TokenEndpointAuthMethod)
}
