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

import dbmodel "github.com/asgardeo/thunder/internal/system/database/model"

var (
	// queryCreateApplication creates an application with gateway config only.
	// Identity fields (name, description, clientId, credentials) are stored in the ENTITY table.
	queryCreateApplication = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-01",
		Query: "INSERT INTO APPLICATION (ID, AUTH_FLOW_ID, " +
			"REGISTRATION_FLOW_ID, IS_REGISTRATION_FLOW_ENABLED, THEME_ID, LAYOUT_ID, APP_JSON, DEPLOYMENT_ID) " +
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
	}
	// queryCreateOAuthApplication creates an OAuth config entry.
	// CLIENT_ID and CLIENT_SECRET are stored in the ENTITY table.
	queryCreateOAuthApplication = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-02",
		Query: "INSERT INTO APP_OAUTH_INBOUND_CONFIG (ENTITY_ID, OAUTH_CONFIG_JSON, DEPLOYMENT_ID) VALUES ($1, $2, $3)",
	}
	// queryGetApplicationByAppID retrieves application gateway config by app/entity ID.
	// Identity fields come from the ENTITY table via the entity provider.
	queryGetApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-03",
		Query: "SELECT app.ID, app.AUTH_FLOW_ID, " +
			"app.REGISTRATION_FLOW_ID, app.IS_REGISTRATION_FLOW_ENABLED, app.THEME_ID, app.LAYOUT_ID, app.APP_JSON " +
			"FROM APPLICATION app " +
			"WHERE app.ID = $1 AND app.DEPLOYMENT_ID = $2",
	}
	// queryGetOAuthApplicationByEntityID retrieves OAuth config by entity/app ID.
	queryGetOAuthApplicationByEntityID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-05",
		Query: "SELECT ENTITY_ID, OAUTH_CONFIG_JSON FROM APP_OAUTH_INBOUND_CONFIG WHERE ENTITY_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// queryGetApplicationList lists all application gateway configs.
	// Identity fields come from the ENTITY table via the entity provider.
	queryGetApplicationList = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-06",
		Query: "SELECT app.ID, app.AUTH_FLOW_ID, " +
			"app.REGISTRATION_FLOW_ID, app.IS_REGISTRATION_FLOW_ENABLED, app.THEME_ID, app.LAYOUT_ID, app.APP_JSON " +
			"FROM APPLICATION app WHERE app.DEPLOYMENT_ID = $1",
	}
	// queryUpdateApplicationByAppID updates application gateway config.
	// Identity fields are updated via the ENTITY table.
	queryUpdateApplicationByAppID = dbmodel.DBQuery{
		ID: "ASQ-APP_MGT-07",
		Query: "UPDATE APPLICATION SET AUTH_FLOW_ID=$2, " +
			"REGISTRATION_FLOW_ID=$3, IS_REGISTRATION_FLOW_ENABLED=$4, THEME_ID=$5, LAYOUT_ID=$6, APP_JSON=$7 " +
			"WHERE ID = $1 AND DEPLOYMENT_ID = $8",
	}
	// queryUpdateOAuthApplicationByAppID updates OAuth config.
	queryUpdateOAuthApplicationByAppID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-08",
		Query: "UPDATE APP_OAUTH_INBOUND_CONFIG SET OAUTH_CONFIG_JSON=$2 WHERE ENTITY_ID=$1 AND DEPLOYMENT_ID=$3",
	}
	// queryDeleteApplicationByAppID deletes an application by app/entity ID.
	queryDeleteApplicationByAppID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-09",
		Query: "DELETE FROM APPLICATION WHERE ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// queryGetApplicationCount gets the total count of applications.
	queryGetApplicationCount = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-10",
		Query: "SELECT COUNT(*) as total FROM APPLICATION WHERE DEPLOYMENT_ID = $1",
	}
	// queryDeleteOAuthApplicationByAppID deletes an OAuth config by app/entity ID.
	queryDeleteOAuthApplicationByAppID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-11",
		Query: "DELETE FROM APP_OAUTH_INBOUND_CONFIG WHERE ENTITY_ID = $1 AND DEPLOYMENT_ID = $2",
	}
	// queryCheckApplicationExistsByID checks if an application exists by app/entity ID.
	queryCheckApplicationExistsByID = dbmodel.DBQuery{
		ID:    "ASQ-APP_MGT-12",
		Query: "SELECT COUNT(*) as count FROM APPLICATION WHERE ID = $1 AND DEPLOYMENT_ID = $2",
	}
)
