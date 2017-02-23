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
