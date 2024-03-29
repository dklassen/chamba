package gorm

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/dklassen/chamba/Godeps/_workspace/src/github.com/lib/pq/hstore"
)

type postgres struct {
	commonDialect
}

func (postgres) BinVar(i int) string {
	return fmt.Sprintf("$%v", i)
}

func (postgres) SupportLastInsertId() bool {
	return false
}

func (postgres) SqlTag(value reflect.Value, size int, autoIncrease bool) string {
	switch value.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		if autoIncrease {
			return "serial"
		}
		return "integer"
	case reflect.Int64, reflect.Uint64:
		if autoIncrease {
			return "bigserial"
		}
		return "bigint"
	case reflect.Float32, reflect.Float64:
		return "numeric"
	case reflect.String:
		if size > 0 && size < 65532 {
			return fmt.Sprintf("varchar(%d)", size)
		}
		return "text"
	case reflect.Struct:
		if _, ok := value.Interface().(time.Time); ok {
			return "timestamp with time zone"
		}
	case reflect.Map:
		if value.Type() == hstoreType {
			return "hstore"
		}
	default:
		if isByteArrayOrSlice(value) {
			return "bytea"
		} else if isUUID(value) {
			return "uuid"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s) for postgres", value.Type().Name(), value.Kind().String()))
}

var byteType = reflect.TypeOf(uint8(0))

func isByteArrayOrSlice(value reflect.Value) bool {
	return (value.Kind() == reflect.Array || value.Kind() == reflect.Slice) && value.Type().Elem() == byteType
}

func isUUID(value reflect.Value) bool {
	if value.Kind() != reflect.Array || value.Type().Len() != 16 {
		return false
	}
	typename := value.Type().Name()
	lower := strings.ToLower(typename)
	return "uuid" == lower || "guid" == lower
}

func (s postgres) ReturningStr(tableName, key string) string {
	return fmt.Sprintf("RETURNING %v.%v", tableName, key)
}

func (s postgres) HasTable(scope *Scope, tableName string) bool {
	var count int
	s.RawScanInt(scope, &count, "SELECT count(*) FROM INFORMATION_SCHEMA.tables WHERE table_name = ? AND table_type = 'BASE TABLE'", tableName)
	return count > 0
}

func (s postgres) HasColumn(scope *Scope, tableName string, columnName string) bool {
	var count int
	s.RawScanInt(scope, &count, "SELECT count(*) FROM INFORMATION_SCHEMA.columns WHERE table_name = ? AND column_name = ?", tableName, columnName)
	return count > 0
}

func (postgres) RemoveIndex(scope *Scope, indexName string) {
	scope.Err(scope.NewDB().Exec(fmt.Sprintf("DROP INDEX %v", indexName)).Error)
}

func (s postgres) HasIndex(scope *Scope, tableName string, indexName string) bool {
	var count int
	s.RawScanInt(scope, &count, "SELECT count(*) FROM pg_indexes WHERE tablename = ? AND indexname = ?", tableName, indexName)
	return count > 0
}

func (s postgres) CurrentDatabase(scope *Scope) (name string) {
	s.RawScanString(scope, &name, "SELECT CURRENT_DATABASE()")
	return
}

var hstoreType = reflect.TypeOf(Hstore{})

type Hstore map[string]*string

func (h Hstore) Value() (driver.Value, error) {
	hstore := hstore.Hstore{Map: map[string]sql.NullString{}}
	if len(h) == 0 {
		return nil, nil
	}

	for key, value := range h {
		var s sql.NullString
		if value != nil {
			s.String = *value
			s.Valid = true
		}
		hstore.Map[key] = s
	}
	return hstore.Value()
}

func (h *Hstore) Scan(value interface{}) error {
	hstore := hstore.Hstore{}

	if err := hstore.Scan(value); err != nil {
		return err
	}

	if len(hstore.Map) == 0 {
		return nil
	}

	*h = Hstore{}
	for k := range hstore.Map {
		if hstore.Map[k].Valid {
			s := hstore.Map[k].String
			(*h)[k] = &s
		} else {
			(*h)[k] = nil
		}
	}

	return nil
}
