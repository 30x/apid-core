package wrap

import (
	"database/sql/driver"

	"sync/atomic"

	"github.com/30x/apid-core"
	"github.com/mattn/go-sqlite3"
)

type wrapConn struct {
	*sqlite3.SQLiteConn
	log         apid.LogService
	stmtCounter int64
	txCounter   int64
}

func (c *wrapConn) Swap(cc *sqlite3.SQLiteConn) {
	c.SQLiteConn = cc
}

func (c *wrapConn) Prepare(query string) (driver.Stmt, error) {
	stmtID := atomic.AddInt64(&c.stmtCounter, 1)
	log := c.log.WithField("stmt", stmtID)
	log.Debugf("begin prepare stmt: %s", query)

	stmt, err := c.SQLiteConn.Prepare(query)
	if err != nil {
		log.Errorf("prepare stmt failed: %s", err)
		return nil, err
	}

	log.Debug("end prepare stmt")
	s := stmt.(*sqlite3.SQLiteStmt)
	return &wrapStmt{s, log}, nil
}

func (c *wrapConn) Begin() (driver.Tx, error) {
	txID := atomic.AddInt64(&c.txCounter, 1)
	log := c.log.WithField("tx", txID)
	log.Debug("begin trans %d", txID)

	tx, err := c.SQLiteConn.Begin()
	if err != nil {
		log.Errorf("begin trans failed: %s", err)
		return nil, err
	}

	log.Debug("end begin trans")
	t := tx.(*sqlite3.SQLiteTx)
	return &wrapTx{t, log}, nil
}

func (c *wrapConn) Close() (err error) {
	c.log.Debug("begin close conn")

	if err = c.SQLiteConn.Close(); err != nil {
		c.log.Errorf("close conn failed: %s", err)
		return
	}

	c.log.Debug("end close conn")
	return
}

func (c *wrapConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	c.log.Debugf("query: %s args: %#v", query, args)
	return c.SQLiteConn.Query(query, args)
}
