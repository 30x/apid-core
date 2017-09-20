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

package wrap

import (
	"database/sql/driver"
	"github.com/apid/apid-core"
	"github.com/mattn/go-sqlite3"
)

type wrapStmt struct {
	*sqlite3.SQLiteStmt
	log apid.LogService
}

func (s *wrapStmt) Close() (err error) {
	s.log.Debug("begin close stmt")

	if err = s.SQLiteStmt.Close(); err != nil {
		s.log.Debugf("close stmt failed: %s", err)
		return
	}

	s.log.Debug("end close stmt")
	return
}

func (s *wrapStmt) NumInput() int {
	return s.SQLiteStmt.NumInput()
}

func (s *wrapStmt) Exec(args []driver.Value) (result driver.Result, err error) {
	s.log.Debugf("begin exec: %#v", args)

	result, err = s.SQLiteStmt.Exec(args)
	if err != nil {
		s.log.Errorf("failed exec: %s", err)
		return
	}

	s.log.Debugf("end exec: %#v", result)
	return
}

func (s *wrapStmt) Query(args []driver.Value) (rows driver.Rows, err error) {
	s.log.Debugf("begin query: %#v", args)

	rows, err = s.SQLiteStmt.Query(args)
	if err != nil {
		s.log.Errorf("failed query: %s", err)
		return
	}

	s.log.Debugf("end query: %#v", rows)
	return
}
