package sqlen

import (
	"database/sql"
	"fmt"
)

// EnhanceRows 用于在 Enhanced 方法中替换元生的 sql.Rows。
type EnhanceRows struct {
	*sql.Rows
	err      error
	columns  []string
	colTypes []*sql.ColumnType
}

// 用 Scan 来查询数据，原生 scan 方法要求和查询的列完全一致，本方法做个兼容。
func (rs *EnhanceRows) scan() ([]interface{}, error) {
	// 如果列的元数据为空，需要初始化。
	if rs.columns == nil {
		columns, err := rs.Columns()
		if err != nil {
			rs.err = err
			return nil, err
		}
		rs.columns = columns

		colTypes, err := rs.ColumnTypes()
		if err != nil {
			rs.err = err
			return nil, err
		}
		rs.colTypes = colTypes
	}

	// 用来存放 Scan 后返回的数据，db 库要求和查询的列完全一致，所以需要判断 columns 长度。
	values := make([]interface{}, len(rs.columns))
	for i := range values {
		values[i] = new(interface{})
	}

	// 用原生 row 的 Scan 方法获取数据。
	err := rs.Scan(values...)
	if err != nil {
		return nil, err
	}

	return values, nil
}

// MapScan 用于把一行数据填充到 map 中。
func (rs *EnhanceRows) MapScan(dest map[string]interface{}) error {
	values, err := rs.scan()
	if err != nil {
		return err
	}

	for i, column := range rs.columns {
		val := *(values[i].(*interface{}))
		if dest[column], err = mapDataType(rs.colTypes[i], val); err != nil {
			rs.err = err
			return err
		}
	}

	return rs.Err()
}

// SliceScan 用 Slice 的方式返回一行数据。
func (rs *EnhanceRows) SliceScan() ([]interface{}, error) {
	values, err := rs.scan()
	if err != nil {
		return nil, err
	}

	for i := range rs.columns {
		val := *(values[i].(*interface{}))
		if values[i], err = mapDataType(rs.colTypes[i], val); err != nil {
			rs.err = err
			return nil, err
		}
	}

	return values, rs.Err()
}

func (r *EnhanceRows) Err() error {
	if r.err != nil {
		return r.err
	}
	return r.Rows.Err()
}

// mapDataType 用于处理数据库类型到 Go 类型的映射关系。
func mapDataType(colType *sql.ColumnType, value interface{}) (interface{}, error) {
	switch colType.DatabaseTypeName() {
	default:
		return value, nil // 非需要特殊处理的类型，直接返回。

	// DECIMAL 类型统一使用 string 方式返回。
	case "DECIMAL":
		switch v := value.(type) {
		case []byte:
			if nullable, ok := colType.Nullable(); ok && nullable {
				if v == nil {
					return sql.NullString{String: "", Valid: false}, nil
				} else {
					return sql.NullString{String: string(v), Valid: true}, nil
				}
			} else {
				return string(v), nil
			}
		case string:
			return v, nil
		default:
			return nil, fmt.Errorf("data: cannot convert DECIMAL field, colname=%s, value=%v", colType.Name(), v)
		}

	// 字符串在 MySql 中默认是 byte 数组，这里也做个处理。
	case "NVARCHAR", "VARCHAR":
		switch v := value.(type) {
		case []byte:
			if nullable, ok := colType.Nullable(); ok && nullable {
				if v == nil {
					return sql.NullString{String: "", Valid: false}, nil
				} else {
					return sql.NullString{String: string(v), Valid: true}, nil
				}
			} else {
				return string(v), nil
			}
		case string:
			return v, nil
		default:
			return nil, fmt.Errorf("data: cannot convert VARCHAR/NVARCHAR field, colname=%s, value=%v", colType.Name(), v)
		}
	}
}

// EnhanceRow 用于在 Enhanced 方法中替换元生的 sql.Row。
// 注意这里原生的 sql.Row 方法没有开放内部 sql.Rows 结构出来，所以直接通过 sql.Rows 实现。
type EnhanceRow struct {
	rows *EnhanceRows
	err  error
}

// Scan 用于把一行数据填充到 map 中。
func (r *EnhanceRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}

	if r.rows == nil {
		r.err = sql.ErrNoRows
		return r.err
	}
	defer r.rows.Close()

	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		r.err = sql.ErrNoRows
		return sql.ErrNoRows
	}

	return r.rows.Scan(dest...)
}

// MapScan 用于把一行数据填充到 map 中。
func (r *EnhanceRow) MapScan(dest map[string]interface{}) error {
	if r.err != nil {
		return r.err
	}

	if r.rows == nil {
		r.err = sql.ErrNoRows
		return r.err
	}
	defer r.rows.Close()

	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		r.err = sql.ErrNoRows
		return sql.ErrNoRows
	}

	return r.rows.MapScan(dest)
}

// SliceScan 用 Slice 返回一行数据。
func (r *EnhanceRow) SliceScan() ([]interface{}, error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.rows == nil {
		r.err = sql.ErrNoRows
		return nil, r.err
	}

	defer r.rows.Close()
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return nil, err
		}
		r.err = sql.ErrNoRows
		return nil, sql.ErrNoRows
	}

	return r.rows.SliceScan()
}

func (r *EnhanceRow) Err() error {
	if r.err != nil {
		return r.err
	}
	return r.rows.Err()
}
