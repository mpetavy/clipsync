package main

import (
	"github.com/mpetavy/common"
	"github.com/mpetavy/common/orm"
	"github.com/mpetavy/common/orm/sqlite"
)

type Database struct {
	Service
	ORM *orm.ORM
}

const (
	SchemaVersion = 1
)

func NewDatabase() (*Database, error) {
	common.DebugFunc()

	common.StartInfo("Database")

	database := &Database{}

	driver, err := sqlite.NewDriver()
	if common.Error(err) {
		return nil, err
	}

	modelsAndSchemas := []orm.ORMSchemaModel{
		{
			Model:       &orm.DBInfo{},
			Schema:      &orm.DBInfoSchema,
			CanTruncate: false,
		},
		{
			Model:       &orm.Log{},
			Schema:      &orm.LogSchema,
			CanTruncate: true,
		},
		{
			Model:       &Bookmark{},
			Schema:      &BookmarkSchema,
			CanTruncate: true,
		},
	}

	database.ORM, err = orm.NewORM(driver, SchemaVersion, modelsAndSchemas)
	if common.Error(err) {
		return nil, err
	}

	err = database.ORM.Prepare(nil, nil)
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

	return nil
}

func (database *Database) Health() error {
	common.DebugFunc()

	db, err := database.ORM.Gorm.DB()
	if common.Error(err) {
		return err
	}

	err = db.Ping()
	if common.Error(err) {
		return err
	}

	return nil
}
