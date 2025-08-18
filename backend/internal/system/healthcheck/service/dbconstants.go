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

// Package service provides health check-related business logic and operations.
package service

import "github.com/asgardeo/thunder/internal/system/database/model"

var queryConfigDBTable = model.DBQuery{
	ID:    "HLC-00001",
	Query: "SELECT ALLOWED_ORIGINS FROM IDN_OAUTH_ALLOWED_ORIGINS",
}

var queryRuntimeDBTable = model.DBQuery{
	ID:    "HLC-00002",
	Query: "SELECT CODE_ID FROM IDN_OAUTH2_AUTHZ_CODE",
}
