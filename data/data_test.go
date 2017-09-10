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

package data_test

import (
	"fmt"
	"github.com/30x/apid-core"
	"github.com/30x/apid-core/data"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math/rand"
	"strconv"
	"time"
)

const (
	count    = 500
	setupSql = `
		CREATE TABLE test_1 (id INTEGER PRIMARY KEY, counter TEXT);
		CREATE TABLE test_2 (id INTEGER PRIMARY KEY, counter TEXT);`
)

var (
	r *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

var _ = Describe("Data Service", func() {

	It("should not allow reserved id or version", func() {
		_, err := apid.Data().DBForID("common")
		Expect(err).To(HaveOccurred())

		_, err = apid.Data().DBVersion("base")
		Expect(err).To(HaveOccurred())

		_, err = apid.Data().DBVersionForID("common", "base")
		Expect(err).To(HaveOccurred())
	})
	It("should be able to change versions of a datbase", func() {
		var versions []string
		var dbs []apid.DB

		for i := 0; i < 2; i++ {
			version := time.Now().String()
			db, err := apid.Data().DBVersionForID("test", version)
			Expect(err).NotTo(HaveOccurred())
			setup(db)
			versions = append(versions, version)
			dbs = append(dbs, db)
		}

		for _, db := range dbs {
			var numRows int
			err := db.QueryRow(`SELECT count(*) FROM test_2`).Scan(&numRows)
			Expect(err).NotTo(HaveOccurred())
			Expect(numRows).To(Equal(count))
		}
	})

	It("should be able to release a database", func() {
		db, err := apid.Data().DBVersionForID("release", "version")
		Expect(err).NotTo(HaveOccurred())
		setup(db)
		id := data.VersionedDBID("release", "version")
		Expect(db.Stats().OpenConnections).To(Equal(1))
		// run finalizer
		data.Delete(id).(func(db *data.ApidDb))(db.(*data.ApidDb))
		Expect(db.Stats().OpenConnections).To(Equal(0))
		Expect(data.DBPath(id)).ShouldNot(BeAnExistingFile())
	})

	It("should handle concurrent read & serialized write, and throttle", func() {
		db, err := apid.Data().DBForID("test")
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		Expect(err).NotTo(HaveOccurred())
		setup(db)
		finished := make(chan bool, count+1)

		go func() {
			for i := 0; i < count; i++ {
				write(db, i)
			}
			finished <- true
		}()

		for i := 0; i < count; i++ {
			go func() {
				read(db)
				finished <- true
			}()
		}

		for i := 0; i < count+1; i++ {
			<-finished
			// Since conns are shared, will always be <= 10
			Expect(db.Stats().OpenConnections).Should(BeNumerically("<=", 10))
		}
	}, 10)

	It("should handle concurrent write", func() {
		db, err := apid.Data().DBForID("test_write")
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		Expect(err).NotTo(HaveOccurred())
		setup(db)
		finished := make(chan bool, count)

		for i := 0; i < count; i++ {
			go func(j int) {
				write(db, j)
				finished <- true
			}(i)
		}

		for i := 0; i < count; i++ {
			<-finished
			// Only one connection should get opened, as connections are serialized.
			Expect(db.Stats().OpenConnections).To(Equal(1))
		}
	}, 10)

	It("should handle concurrent read & write", func() {
		db, err := apid.Data().DBForID("test_read_write")
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		Expect(err).NotTo(HaveOccurred())
		setup(db)
		finished := make(chan bool, 2*count)

		for i := 0; i < count; i++ {
			go func(j int) {
				write(db, j)
				finished <- true
			}(i)
			go func() {
				read(db)
				finished <- true
			}()
		}

		for i := 0; i < 2*count; i++ {
			<-finished
		}
	}, 10)

	It("should handle concurrent alter table by seralizing them", func() {
		db, err := apid.Data().DBForID("test_alter")
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		Expect(err).NotTo(HaveOccurred())
		setup(db)
		alterCount := 50
		finished := make(chan bool, alterCount)

		for i := 0; i < alterCount; i++ {
			go func(j int) {
				alter(db, j, "test_1")
				finished <- true
			}(i)
		}

		for i := 0; i < alterCount; i++ {
			<-finished
			Expect(db.Stats().OpenConnections).To(Equal(1))
		}
	}, 10)

	It("should handle seralized read & alter table", func() {
		db, err := apid.Data().DBForID("test_read_alter")
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		Expect(err).NotTo(HaveOccurred())
		setup(db)
		alterCount := 50
		for i := 0; i < alterCount; i++ {
			alter(db, i, "test_1")
			read(db)
		}
	}, 10)

	It("concurrent read & alter the same table ", func() {
		db, err := apid.Data().DBForID("test_conn_read_alter")
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		Expect(err).NotTo(HaveOccurred())
		setup(db)
		alterCount := 50
		finished := make(chan bool, count+alterCount)
		for i := 0; i < alterCount; i++ {
			go func(j int) {
				alter(db, j, "test_1")
				finished <- true
			}(i)
		}

		for i := 0; i < count; i++ {
			go func() {
				read(db)
				finished <- true
			}()
		}

		for i := 0; i < count+alterCount; i++ {
			<-finished
		}
	}, 10)

	It("concurrent read & alter different tables", func() {
		db, err := apid.Data().DBForID("test_conn_read_alter_diff_table")
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		Expect(err).NotTo(HaveOccurred())
		setup(db)
		alterCount := 50
		finished := make(chan bool, count+alterCount)
		for i := 0; i < alterCount; i++ {
			go func(j int) {
				alter(db, j, "test_2")
				finished <- true
			}(i)
		}

		for i := 0; i < count; i++ {
			go func() {
				read(db)
				finished <- true
			}()
		}

		for i := 0; i < count+alterCount; i++ {
			<-finished
		}
	}, 10)
})

func setup(db apid.DB) {
	_, err := db.Exec(setupSql)
	Expect(err).Should(Succeed())
	tx, err := db.Begin()
	Expect(err).Should(Succeed())
	for i := 0; i < count; i++ {
		_, err := tx.Exec("INSERT INTO test_2 (counter) VALUES (?);", strconv.Itoa(i))
		Expect(err).Should(Succeed())
	}
	Expect(tx.Commit()).Should(Succeed())
}

func read(db apid.DB) {
	defer GinkgoRecover()
	var counter string
	tx, err := db.BeginReadOnly()
	Expect(err).Should(Succeed())
	defer tx.Rollback()
	rows, err := tx.Query(`SELECT counter FROM test_1 LIMIT 5`)
	Expect(err).Should(Succeed())
	Expect(rows.Close()).Should(Succeed())
	Expect(tx.Commit()).Should(Succeed())
	for rows.Next() {
		rows.Scan(&counter)
	}
	fmt.Print(".")
}

func write(db apid.DB, i int) {
	defer GinkgoRecover()
	// DB INSERT as a txn
	tx, err := db.Begin()
	Expect(err).Should(Succeed())
	defer tx.Rollback()
	prep, err := tx.Prepare("INSERT INTO test_1 (counter) VALUES ($1);")
	Expect(err).Should(Succeed())
	_, err = prep.Exec(strconv.Itoa(i))
	Expect(err).Should(Succeed())
	Expect(prep.Close()).Should(Succeed())
	Expect(tx.Commit()).Should(Succeed())
	// DB INSERT directly, not via a txn
	//_, err = db.Exec("INSERT INTO test_1 (counter) VALUES ($?)", i+10000)
	//Expect(err).Should(Succeed())
	fmt.Print("+")
}

func alter(db apid.DB, i int, tableName string) {
	defer GinkgoRecover()
	// DB INSERT as a txn
	tx, err := db.Begin()
	Expect(err).Should(Succeed())
	defer tx.Rollback()
	prep, err := tx.Prepare("ALTER TABLE " + tableName + " ADD COLUMN colname_" + strconv.Itoa(i) + " text DEFAULT ''")
	Expect(err).Should(Succeed())
	_, err = prep.Exec()
	Expect(err).Should(Succeed())
	Expect(prep.Close()).Should(Succeed())
	Expect(tx.Commit()).Should(Succeed())
	fmt.Print("+")
}
