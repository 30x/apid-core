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
	"strings"

	"sync/atomic"

	"github.com/30x/apid-core"
	"github.com/mattn/go-sqlite3"
)

func NewDriver(d driver.Driver, log apid.LogService) driver.Driver {
	return wrapDriver{d, log, 0}
}

type wrapDriver struct {
	driver.Driver
	log     apid.LogService
	counter int64
}

func (d wrapDriver) Open(dsn string) (driver.Conn, error) {
	connId := atomic.AddInt64(&d.counter, 1)
	log := d.log.WithField("conn", connId)
	log.Debug("begin open conn")

	internalDSN := strings.TrimPrefix(dsn, "dd:")
	internalCon, err := d.Driver.Open(internalDSN)
	if err != nil {
		log.Errorf("open conn failed: %v", err)
		return nil, err
	}

	c := internalCon.(*sqlite3.SQLiteConn)
	return &wrapConn{c, log, 0, 0}, nil
}
