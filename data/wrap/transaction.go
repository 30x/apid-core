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
	"github.com/apid/apid-core"
	"github.com/mattn/go-sqlite3"
)

type wrapTx struct {
	*sqlite3.SQLiteTx
	log apid.LogService
}

func (tx *wrapTx) Commit() (err error) {
	tx.log.Debug("begin commit")

	if err = tx.SQLiteTx.Commit(); err != nil {
		tx.log.Errorf("failed commit: %s", err)
		return
	}

	tx.log.Debug("end commit")
	return
}

func (tx *wrapTx) Rollback() (err error) {
	tx.log.Debug("begin rollback")

	if err = tx.SQLiteTx.Rollback(); err != nil {
		tx.log.Errorf("failed rollback: %s", err)
	}

	tx.log.Debug("end rollback")
	return
}
