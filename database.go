package main

import (
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mpetavy/common"
	"github.com/mpetavy/common/sqldb"
)

var (
	dbFile *string
)

func init() {
	common.Events.AddListener(common.EventInit{}, func(event common.Event) {
		dbFile = flag.String("db.file", common.AppFilename(".db"), "Database file")
	})
}

type Database struct {
	Service
	SqlDb *sqldb.SqlDb
}

func NewDatabase() (*Database, error) {
	common.DebugFunc()

	common.StartInfo("Database")

	database := &Database{}

	var err error

	database.SqlDb, err = sqldb.NewSqlDb("sqlite3", *dbFile)
	if common.Error(err) {
		return nil, err
	}

	return database, nil
}

func (database *Database) Reset() error {
	return nil
}

func (database *Database) Close() error {
	common.StopInfo("Database")

	if database.SqlDb != nil {
		err := database.SqlDb.Close()
		if common.Error(err) {
			return err
		}

		database.SqlDb = nil
	}

	return nil
}

func (database *Database) Health() error {
	common.DebugFunc()

	err := database.SqlDb.Health()
	if common.Error(err) {
		return err
	}

	common.Info("Database health ok")

	return nil
}
