// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apid

import (
	"database/sql"
)

type DataService interface {
	DB() (DB, error)
	DBForID(id string) (db DB, err error)

	DBVersion(version string) (db DB, err error)
	DBVersionForID(id, version string) (db DB, err error)

	// will set DB to close and delete when no more references
	ReleaseDB(version string)
	ReleaseCommonDB()
	ReleaseDBForID(id, version string)
}

type DB interface {
	Ping() error
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Begin() (*sql.Tx, error)

	//Close() error
	//Stats() sql.DBStats
	//Driver() driver.Driver
}
