package sqlen

import (
	"database/sql"
	"reflect"
)

// 用于统一不同驱动在 Go 中的映射类型。
type UnifyDataTypeFn func(colDbTypeName string, dest *interface{})

// EnhanceRows 用于在 Enhanced 方法中替换元生的 sql.Rows。
type EnhanceRows struct {
	*sql.Rows
	unifyDataType UnifyDataTypeFn

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
	return rs.Rows.Scan(dest...)
}

// noDestScan 自动生成 dest 数组后，通过 Scan 来查询数据。
// (原生 Scan 方法要求和查询的列完全一致，本方法做个兼容。)
func (rs *EnhanceRows) noDestScan() ([]interface{}, error) {
	if rs.err != nil {
		return nil, rs.err
	}

	rs.initColumns()

	// 用来存放 Scan 后返回的数据，db 库要求和查询的列完全一致，所以需要判断 columns 长度。
	dest := make([]interface{}, len(rs.colTypes))
	destRefVal := make([]reflect.Value, len(rs.colTypes))
	for i, cType := range rs.colTypes {
		refVal := reflect.New(unifyNumber(cType.ScanType())) // 使用数据库驱动标记的类型来接收数据。
		dest[i] = refVal.Interface()                         // 注意，这里传入的是指定值的指针。
		destRefVal[i] = refVal                               // 保存这个 Reflect.value 在后面用于解引用。
	}

	rs.Scan(dest...)

	for i := 0; i < len(rs.colTypes); i++ {
		dest[i] = destRefVal[i].Elem().Interface()

		// 进行统一类型处理。
		rs.unifyDataType(rs.colTypes[i].DatabaseTypeName(), &dest[i])
		extractNullableValue(rs.colTypes[i], &dest[i])
	}

	return dest, rs.err
}

// MapScan 用于把一行数据填充到 map 中。
func (rs *EnhanceRows) MapScan(dest map[string]interface{}) error {
	values, err := rs.noDestScan()
	if err != nil {
		return err
	}

	for i, column := range rs.columns {
		dest[column] = values[i]
	}

	return nil
}

// SliceScan 用 Slice 的方式返回一行数据。
func (rs *EnhanceRows) SliceScan() ([]interface{}, error) {
	return rs.noDestScan()
}

func (r *EnhanceRows) Err() error {
	if r.err != nil {
		return r.err
	}
	return r.Rows.Err()
}
