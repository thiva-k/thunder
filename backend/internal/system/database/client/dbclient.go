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

// Package client provides database client implementations for executing queries and managing transactions.
package client

import (
	"strings"

	"github.com/asgardeo/thunder/internal/system/database/model"
	"github.com/asgardeo/thunder/internal/system/log"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// DBClientInterface defines the interface for database operations.
type DBClientInterface interface {
	// Query executes a sql query that returns rows, typically a SELECT, and returns the result as a slice of maps.
	Query(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error)
	// Execute executes a sql query without returning data in any rows, and returns number of rows affected.
	Execute(query model.DBQuery, args ...interface{}) (int64, error)
	// BeginTx starts a new database transaction.
	BeginTx() (model.TxInterface, error)
	// Close closes the database connection.
	Close() error
}

// DBClient is the implementation of DBClientInterface.
type DBClient struct {
	db     model.DBInterface
	dbType string
}

// NewDBClient creates a new instance of DBClient with the provided database connection.
func NewDBClient(db model.DBInterface, dbType string) DBClientInterface {
	return &DBClient{
		db:     db,
		dbType: dbType,
	}
}

// Query executes a sql query that returns rows, typically a SELECT, and returns the result as a slice of maps.
func (client *DBClient) Query(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DBClient"))
	logger.Info("Executing query", log.String("queryID", query.GetID()))

	sqlQuery := query.GetQuery(client.dbType)
	rows, err := client.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Error("Error closing rows", log.Error(closeErr))
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		row := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))
		for i := range row {
			rowPointers[i] = &row[i]
		}

		if err := rows.Scan(rowPointers...); err != nil {
			return nil, err
		}

		result := map[string]interface{}{}
		for i, col := range columns {
			// Normalize column names to lowercase for consistency.
			result[strings.ToLower(col)] = row[i]
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Execute executes a sql query without returning data in any rows, and returns number of rows affected.
func (client *DBClient) Execute(query model.DBQuery, args ...interface{}) (int64, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "DBClient"))
	logger.Info("Executing query", log.String("queryID", query.GetID()))

	sqlQuery := query.GetQuery(client.dbType)
	res, err := client.db.Exec(sqlQuery, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// BeginTx starts a new database transaction.
func (client *DBClient) BeginTx() (model.TxInterface, error) {
	tx, err := client.db.Begin()
	if err != nil {
		return nil, err
	}
	return model.NewTx(tx), nil
}

// Close closes the database connection.
func (client *DBClient) Close() error {
	return client.db.Close()
}
