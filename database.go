package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/fatih/structtag"
	"github.com/mpetavy/common"
	"github.com/mpetavy/common/sqldb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"reflect"
	"strings"
)

var (
	dbFile          *string
	dbTruncate      = flag.Bool("db.truncate", false, "Truncate database tables")
	dbSlowThreshold = flag.Int("db.slow.threshold", 5000, "Slow threshold database query")
	dbLogLevel      = flag.String("pg.loglevel", "INFO", "Log level to start logging to database")
)

var (
	ErrNotFound       = fmt.Errorf("record not found")
	ErrDuplicateFound = fmt.Errorf("duplicate record found")
)

type Database struct {
	Service
	Gorm       *gorm.DB
	ServerConn *sql.DB
	gormConfig *gorm.Config
	mu         common.ReentrantMutex
}

const (
	SchemaVersion = 1
)

type gormLogger struct {
	io.Writer
}

func init() {
	common.Events.AddListener(common.EventInit{}, func(event common.Event) {
		dbFile = flag.String("db.file", common.AppFilename(".db"), "Database file")
	})
}

func (lw gormLogger) Printf(s string, v ...any) {
	if len(v) > 0 {
		s = fmt.Sprintf(strings.ReplaceAll(s, "\n", ""), v...)
	}

	common.Debug("GORM: %s", s)
}

func UpdateSchema(db *gorm.DB, st *common.StringTable, model any, schema any) (string, error) {
	common.DebugFunc()

	modelStruct, ok := model.(reflect.Value)
	if !ok {
		modelStruct = reflect.Indirect(reflect.ValueOf(model))
	}

	modelType := modelStruct.Type()
	if modelType.Kind() != reflect.Struct {
		return "", fmt.Errorf("model type should be a struct")
	}

	tableName := db.Config.NamingStrategy.TableName(modelType.Name())

	schemaStruct, ok := schema.(reflect.Value)
	if !ok {
		schemaStruct = reflect.Indirect(reflect.ValueOf(schema))
	}

	schemaType := modelStruct.Type()
	if schemaType.Kind() != reflect.Struct {
		return "", fmt.Errorf("model type should be a struct")
	}

	fieldTableName := "TableName"
	field := schemaStruct.FieldByName(fieldTableName)
	if !field.CanSet() {
		return "", fmt.Errorf("field does not exist in schema struct: %s", fieldTableName)
	}

	field.Set(reflect.ValueOf(tableName))

	if modelType.NumField() != schemaType.NumField() {
		return "", fmt.Errorf("model does not have the same number of fields")
	}

	st.AddCols("", "", "", "")
	st.AddCols(fmt.Sprintf("**%s**", tableName), "", "", "")
	st.AddCols("", "", "", "")

	for i := 0; i < modelType.NumField(); i++ {
		fieldName := modelType.Field(i).Name
		dbFieldName := db.Config.NamingStrategy.ColumnName(modelStruct.String(), fieldName)

		fieldTags, err := structtag.Parse(string(modelType.Field(i).Tag))
		if common.Error(err) {
			return "", err
		}

		descTag, err := fieldTags.Get("desc")
		if descTag != nil {
			st.AddCols(tableName, dbFieldName, modelType.Field(i).Type.String(), descTag.Value())
		}

		gormTag, err := fieldTags.Get("gorm")
		if gormTag != nil {
			if strings.HasPrefix(gormTag.Value(), "foreignKey") {
				continue
			}
		}

		_, ok := schemaType.FieldByName(fieldName)
		if !ok {
			return "", fmt.Errorf("field does not exist in struct: %s", fieldName)
		}

		field := schemaStruct.FieldByName(fieldName)

		if !field.CanSet() {
			return "", fmt.Errorf("field does not exist in schema struct: %s", fieldName)
		}

		field.Set(reflect.ValueOf(dbFieldName))
	}

	return tableName, nil
}

func NewDatabase() (*Database, error) {
	common.DebugFunc()

	common.StartInfo("Database")

	database := &Database{}

	dialector := sqlite.Open(*dbFile)

	logLevel := logger.Error
	if common.IsLogVerboseEnabled() {
		logLevel = logger.Silent
	}

	logLevel = logger.Info

	newLogger := logger.New(&gormLogger{},
		logger.Config{
			SlowThreshold:             common.MillisecondToDuration(*dbSlowThreshold), // Slow SQL threshold
			LogLevel:                  logLevel,                                       // Log level
			IgnoreRecordNotFoundError: false,                                          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,                                          // Don't include params in the SQL log
			Colorful:                  false,                                          // Disable color
		},
	)
	newLogger.LogMode(logger.Info)

	database.gormConfig = &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true,
	}

	var err error

	database.Gorm, err = gorm.Open(dialector, database.gormConfig)
	if common.Error(err) {
		return nil, err
	}

	database.ServerConn, err = database.Gorm.DB()
	if common.Error(err) {
		return nil, err
	}

	err = database.ServerConn.Ping()
	if common.Error(err) {
		return nil, err
	}

	err = database.RunSynchronized(func() error {
		err := database.Prepare()
		if common.Error(err) {
			return err
		}

		return nil
	})
	if common.Error(err) {
		return nil, err
	}

	return database, nil
}

func (database *Database) Reset() error {
	return nil
}

func (database *Database) MigrateDatabase() error {
	common.DebugFunc()

	modelsAndSchemas := []struct {
		Model       any
		Schema      any
		CanTruncate bool
	}{
		{
			Model:       &DBInfo{},
			Schema:      &DBInfoSchema,
			CanTruncate: false,
		},
		{
			Model:       &Log{},
			Schema:      &LogSchema,
			CanTruncate: true,
		},
		{
			Model:       &Bookmark{},
			Schema:      &BookmarkSchema,
			CanTruncate: true,
		},
	}

	dbInfo := &DBInfo{}

	database.Gorm.First(dbInfo)

	common.Debug("Database schema version: %d", dbInfo.SchemaVersion)

	if dbInfo.SchemaVersion > SchemaVersion {
		return fmt.Errorf("Invalid database schema version, found %d but software wants %d", dbInfo.SchemaVersion, SchemaVersion)
	}

	doMigrate := dbInfo.SchemaVersion != SchemaVersion

	if doMigrate {
		common.Info("Migrate database: %d -> %d", dbInfo.SchemaVersion, SchemaVersion)
	}

	if *dbTruncate {
		common.Info("Truncate database")
	}

	st := common.NewStringTable()
	st.AddCols("Table", "fieldname", "type", "description")

	for i := 0; i < len(modelsAndSchemas); i++ {
		if i > 0 {
			st.AddCols("", "", "", "")
		}

		tableName, err := UpdateSchema(database.Gorm, st, modelsAndSchemas[i].Model, modelsAndSchemas[i].Schema)
		if common.Error(err) {
			return err
		}

		if doMigrate {
			err = database.Gorm.AutoMigrate(modelsAndSchemas[i].Model)
			if common.Error(err) {
				return err
			}
		}

		if *dbTruncate && modelsAndSchemas[i].CanTruncate {
			tx := database.Gorm.Exec(fmt.Sprintf("delete from %s", tableName))
			if common.Error(tx.Error) {
				return tx.Error
			}

			tx = database.Gorm.Exec(fmt.Sprintf("alter sequence %s_id_seq restart with 1", tableName))
			if common.Error(tx.Error) {
				return tx.Error
			}
		}
	}

	common.Debug("Schema:\n" + st.Markdown())

	dbInfo = &DBInfo{}

	tx := database.Gorm.FirstOrCreate(dbInfo)
	if common.Error(tx.Error) {
		return tx.Error
	}

	if doMigrate {
		err := database.Gorm.Transaction(func(txTransaction *gorm.DB) error {
			for schemaVersion := dbInfo.SchemaVersion; schemaVersion < SchemaVersion; schemaVersion++ {
				switch schemaVersion {
				}
			}

			dbInfo.SchemaVersion = SchemaVersion

			tx = txTransaction.Save(dbInfo)
			if common.Error(tx.Error) {
				return tx.Error
			}

			return nil
		})
		if common.Error(err) {
			return err
		}
	}

	return nil
}

func (database *Database) InitLogger() error {
	if *dbLogLevel != "" {
		common.Events.AddListener(common.EventLog{}, func(event common.Event) {
			eventLog := event.(common.EventLog)

			level := common.LevelToIndex(eventLog.Entry.Level)

			if level != -1 && level < common.LevelToIndex(*dbLogLevel) {
				return
			}

			db, err := database.Gorm.DB()
			if err != nil {
				return
			}

			err = db.Ping()
			if err != nil {
				return
			}

			msg := eventLog.Entry.Msg
			if common.LevelToIndex(eventLog.Entry.Level) >= common.LevelToIndex(common.LevelError) {
				msg = eventLog.Entry.StacktraceMsg
			}

			logging := &Log{
				CreatedAt: sqldb.NewFieldTime(eventLog.Entry.Time),
				UpdatedAt: sqldb.NewFieldTime(eventLog.Entry.Time),
				Level:     sqldb.NewFieldString(eventLog.Entry.Level),
				Source:    sqldb.NewFieldString(eventLog.Entry.Source),
				Msg:       sqldb.NewFieldString(msg),
			}

			tx := database.Gorm.Create(logging)
			if common.Error(tx.Error) {
				common.Error(tx.Error)
			}
		})
	}

	return nil
}

func (database *Database) Prepare() error {
	common.DebugFunc()

	common.Info("Prepare database")

	err := database.MigrateDatabase()
	if common.Error(err) {
		return err
	}

	err = database.InitLogger()
	if common.Error(err) {
		return err
	}

	return nil
}

func (database *Database) Close() error {
	common.StopInfo("Database")

	return nil
}

func (database *Database) Health() error {
	common.DebugFunc()

	db, err := database.Gorm.DB()
	if common.Error(err) {
		return err
	}

	err = db.Ping()
	if common.Error(err) {
		return err
	}

	return nil
}

func (database *Database) RunSynchronized(fn func() error) error {
	database.mu.Lock()
	defer func() {
		database.mu.Unlock()
	}()

	err := fn()
	if common.Error(err) {
		return err
	}

	return nil
}

func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
		return true
	}

	return false
}

func SelectOp(fieldname ...string) string {
	return strings.Join(fieldname, ",")
}

const (
	IsIn    = "_IN_"
	BeginOp = "_BEGIN_"
	EndOp   = "_END"
	AndOp   = "_AND_"
	OrOp    = "_OR_"
	IsNull  = "_IS_NULL_"
)

type WhereTerm struct {
	list []WhereItem
}

func NewWhereTerm() *WhereTerm {
	return &WhereTerm{}
}

func (whereTerm *WhereTerm) Where(whereItem WhereItem) *WhereTerm {
	whereTerm.list = append(whereTerm.list, whereItem)

	return whereTerm
}

func (whereTerm *WhereTerm) Begin() *WhereTerm {
	whereTerm.list = append(whereTerm.list, WhereItem{"", BeginOp, nil})

	return whereTerm
}

func (whereTerm *WhereTerm) End() *WhereTerm {
	whereTerm.list = append(whereTerm.list, WhereItem{"", EndOp, nil})

	return whereTerm
}

func (whereTerm *WhereTerm) Or() *WhereTerm {
	whereTerm.list = append(whereTerm.list, WhereItem{"", OrOp, nil})

	return whereTerm
}

func (whereTerm *WhereTerm) And() *WhereTerm {
	whereTerm.list = append(whereTerm.list, WhereItem{"", AndOp, nil})

	return whereTerm
}

func ToAny[T any](arrOrSingle T) []any {
	anys := []any{}

	common.IgnoreError(common.Catch(func() error {
		valueOf := reflect.ValueOf(arrOrSingle)

		switch valueOf.Kind() {
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			for i := 0; i < valueOf.Len(); i++ {
				anys = append(anys, valueOf.Index(i).Interface())
			}
		default:
			anys = append(anys, arrOrSingle)
		}

		return nil
	}))

	return anys
}

type WhereItem struct {
	Fieldname string
	Operator  string
	Value     any
}

func (whereItem WhereItem) Build() (string, []any) {
	var values []any
	sb := strings.Builder{}

	switch whereItem.Operator {
	case IsIn:
		sb.WriteString(fmt.Sprintf("%s in (?)", whereItem.Fieldname))
		values = append(values, whereItem.Value)
	case BeginOp:
		sb.WriteString("(")
	case EndOp:
		sb.WriteString(")")
	case AndOp:
		sb.WriteString(" AND ")
	case OrOp:
		sb.WriteString(" OR ")
	case IsNull:
		sb.WriteString(fmt.Sprintf("%s IS NULL", whereItem.Fieldname))
	default:
		if whereItem.Value != nil {
			values = append(values, whereItem.Value)
		}
		sb.WriteString(fmt.Sprintf("%s %s ?", whereItem.Fieldname, whereItem.Operator))
	}

	return sb.String(), ToAny(whereItem.Value)
}

func (whereTerm WhereTerm) Build() (string, []any) {
	var values []any
	sb := strings.Builder{}

	for _, whereItem := range whereTerm.list {
		w, _ := whereItem.Build()

		sb.WriteString(w)

		switch whereItem.Operator {
		case IsIn:
			values = append(values, whereItem.Value)
		default:
			if whereItem.Value != nil {
				values = append(values, whereItem.Value)
			}
		}
	}

	return sb.String(), values
}

func OrderByOp(fieldname string, ascending bool) string {
	return fmt.Sprintf("%s %s", fieldname, common.Eval(ascending, "asc", "desc"))
}
