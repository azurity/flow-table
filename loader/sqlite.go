package loader

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

type SqliteLoader struct{}

func (loader *SqliteLoader) Simple() bool {
	return false
}

func (loader *SqliteLoader) Load(val string) (map[string]any, error) {
	db, err := sql.Open("sqlite", val)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT name FROM sqlite_schema WHERE type = ?", "table")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tableNames := []string{}
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		tableNames = append(tableNames, name)
	}

	ret := map[string]any{}

	for _, name := range tableNames {
		rows, err := db.Query(fmt.Sprintf("SELECT * from %s", name))
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		typeInstance, err := makeTableType(rows)
		if err != nil {
			return nil, err
		}

		data := []any{}
		for rows.Next() {
			value := reflect.New(typeInstance)
			fields := []interface{}{}
			for i := 0; i < value.Elem().NumField(); i++ {
				fields = append(fields, value.Elem().Field(i).Addr().Interface())
			}
			rows.Scan(fields...)
			data = append(data, value.Interface())
		}
		ret[name] = data
	}
	return ret, nil
}

func makeFieldType(colType *sql.ColumnType) reflect.Type {
	if colType.ScanType() != nil {
		return colType.ScanType()
	} else {
		typeName := colType.DatabaseTypeName()
		if strings.Contains(typeName, "INT") {
			return reflect.TypeOf(int64(0))
		}
		if strings.Contains(typeName, "CHAR") || strings.Contains(typeName, "CLOB") || strings.Contains(typeName, "TEXT") {
			return reflect.TypeOf("")
		}
		if strings.Contains(typeName, "BLOB") {
			return reflect.TypeOf([]byte{})
		}
		if strings.Contains(typeName, "REAL") || strings.Contains(typeName, "FLOA") || strings.Contains(typeName, "DOUB") {
			return reflect.TypeOf(float64(0))
		}
		return reflect.TypeOf("")
	}
}

func makeTableType(rows *sql.Rows) (reflect.Type, error) {
	props, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	fields := []reflect.StructField{}
	for i, name := range props {
		fields = append(fields, reflect.StructField{
			Name: "Field_" + name,
			Type: makeFieldType(types[i]),
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, name)),
		})
	}
	return reflect.StructOf(fields), nil
}
