package sqlen

import (
	"database/sql"
	"fmt"
)

// EnhanceRows 用于在 Enhanced 方法中替换元生的 sql.Rows。
type EnhanceRows struct {
	*sql.Rows
	err error

	columns  []string
	colTypes []*sql.ColumnType
}

// 初始化查询列的元数据。
func (rs *EnhanceRows) initColumns() {
	if rs.columns == nil {
		rs.columns, rs.err = rs.Columns()
		rs.colTypes, rs.err = rs.ColumnTypes()
	}
}

// MapScan 用于把一行数据填充到 map 中。
func (rs *EnhanceRows) Scan(dest ...interface{}) error {
	if rs.err != nil {
		return rs.err
	}

	rs.initColumns()

	// 用原生 row 的 Scan 方法获取数据。
	rs.err = rs.Rows.Scan(dest...)
	if rs.err != nil {
		return rs.err
	}

	// 进行统一的类型处理。
	for i := 0; i < len(rs.colTypes); i++ {
		if destP, ok := dest[i].(*interface{}); ok {
			if rs.err = mapDataType(rs.colTypes[i], destP); rs.err != nil {
				return rs.err
			}
		}
	}

	return nil
}

// noDestScan 自动生成 dest 数组后，通过 Scan 来查询数据。
// (原生 Scan 方法要求和查询的列完全一致，本方法做个兼容。)
func (rs *EnhanceRows) noDestScan() ([]interface{}, error) {
	if rs.err != nil {
		return nil, rs.err
	}

	rs.initColumns()

	// 用来存放 Scan 后返回的数据，db 库要求和查询的列完全一致，所以需要判断 columns 长度。
	dest := make([]interface{}, len(rs.columns))
	for i := range dest {
		dest[i] = new(interface{})
	}

	rs.Scan(dest...)
	return dest, rs.err
}

// MapScan 用于把一行数据填充到 map 中。
func (rs *EnhanceRows) MapScan(dest map[string]interface{}) error {
	values, err := rs.noDestScan()
	if err != nil {
		return err
	}

	for i, column := range rs.columns {
		dest[column] = *(values[i].(*interface{}))
	}

	return nil
}

// SliceScan 用 Slice 的方式返回一行数据。
func (rs *EnhanceRows) SliceScan() ([]interface{}, error) {
	values, err := rs.noDestScan()
	if err != nil {
		return nil, err
	}

	for i := range rs.columns {
		values[i] = *(values[i].(*interface{}))
	}

	return values, nil
}

func (r *EnhanceRows) Err() error {
	if r.err != nil {
		return r.err
	}
	return r.Rows.Err()
}

// mapDataType 用于处理数据库类型到 Go 类型的映射关系。
func mapDataType(colType *sql.ColumnType, dest *interface{}) error {
	switch colType.DatabaseTypeName() {
	// DECIMAL 类型统一使用 string 方式返回。
	case "DECIMAL":
		switch v := (*dest).(type) {
		case []byte:
			if nullable, ok := colType.Nullable(); ok && nullable {
				if v == nil {
					*dest = sql.NullString{String: "", Valid: false}
				} else {
					*dest = sql.NullString{String: string(v), Valid: true}
				}
			} else {
				*dest = string(v)
			}
			return nil
		case string:
			return nil
		default:
			return fmt.Errorf("sqlmer: cannot convert DECIMAL field, colname=%s, value=%v", colType.Name(), v)
		}
	// 字符串在 MySql 等一些驱动中默认是 []byte，这里也做个处理。
	case "NVARCHAR", "VARCHAR":
		switch v := (*dest).(type) {
		case []byte:
			if nullable, ok := colType.Nullable(); ok && nullable {
				if v == nil {
					*dest = sql.NullString{String: "", Valid: false}
				} else {
					*dest = sql.NullString{String: string(v), Valid: true}
				}
			} else {
				*dest = string(v)
			}
			return nil
		case string:
			return nil
		default:
			return fmt.Errorf("sqlmer: cannot convert VARCHAR/NVARCHAR field, colname=%s, value=%v", colType.Name(), v)
		}
	}

	return nil
}
