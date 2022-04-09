package sqlen

import (
	"database/sql"
	"reflect"
)

// 用于统一不同驱动在 Go 中的映射类型。
type UnifyDataTypeFn func(columnType *sql.ColumnType, dest *any)

// GetScanTypeFunc 用于根据列的类型信息获取一个能用于 Scan 的 Go 类型。
type GetScanTypeFunc func(columnType *sql.ColumnType) reflect.Type

// EnhanceRows 用于在 Enhanced 方法中替换元生的 sql.Rows。
type EnhanceRows struct {
	*sql.Rows
	getScanTypeFn GetScanTypeFunc // 用于获取用于 Scan 的数据类型。
	unifyDataType UnifyDataTypeFn

	columnMetaSlice []*columnMeta // 用于对列的元数据做缓存。

	err error
}

type columnMeta struct {
	colType  *sql.ColumnType // 列类型。
	scanType reflect.Type    // 用于 Scan 的类型。
}

// 初始化查询列的元数据。
func (rs *EnhanceRows) initColumns() error {
	if rs.columnMetaSlice != nil {
		return nil
	}

	if colTypes, err := rs.ColumnTypes(); err != nil {
		return err
	} else {
		rs.columnMetaSlice = make([]*columnMeta, 0, len(colTypes))
		for _, cType := range colTypes {
			rs.columnMetaSlice = append(rs.columnMetaSlice, &columnMeta{cType, nil})
		}
	}
	return nil
}

// MapScan 用于把一行数据填充到 map 中。
func (rs *EnhanceRows) Scan(dest ...any) error {
	if rs.err != nil {
		return rs.err
	}

	// 直接用原生 row 的 Scan 方法获取数据。
	return rs.Rows.Scan(dest...)
}

// MapScan 用于把一行数据填充到 map 中。
func (rs *EnhanceRows) MapScan(dest map[string]any) error {
	values, err := rs.SliceScan()
	if err != nil {
		return err
	}

	for i, columnMeta := range rs.columnMetaSlice {
		dest[columnMeta.colType.Name()] = values[i]
	}

	return nil
}

// SliceScan 用 Slice 的方式返回一行数据。
func (rs *EnhanceRows) SliceScan() ([]any, error) {
	if rs.err != nil {
		return nil, rs.err
	}

	rs.initColumns()

	// 用来存放 Scan 后返回的数据，db 库要求和查询的列完全一致，所以需要判断 columns 长度。
	dest := make([]any, len(rs.columnMetaSlice))
	destRefVal := make([]reflect.Value, len(rs.columnMetaSlice))

	for i, colMeta := range rs.columnMetaSlice {
		if colMeta.scanType == nil {
			// 第一次查询，需要获取 scan 的类型，获取后缓存到列元数据里。
			colMeta.scanType = unifyScanType(rs.getScanTypeFn(colMeta.colType), colMeta.colType)
		}

		refVal := reflect.New(colMeta.scanType) // 使用数据库驱动标记的类型来接收数据。
		dest[i] = refVal.Interface()            // 注意，这里传入的是指定值的指针。
		destRefVal[i] = refVal                  // 保存这个 Reflect.value 在后面用于解引用。
	}

	rs.Scan(dest...)

	for i := 0; i < len(rs.columnMetaSlice); i++ {
		// 之前为了能让 Scan 修改数据，保存的是指针，而返回上层时候只需要实际数据，因此进行解引用。
		dest[i] = destRefVal[i].Elem().Interface()

		extractNullableColumnValue(rs.columnMetaSlice[i].colType, &dest[i]) // 进行统一的空值处理逻辑。
		if dest[i] != nil {
			rs.unifyDataType(rs.columnMetaSlice[i].colType, &dest[i]) // 进行数据库定制的类型处理。
		}
	}

	return dest, rs.err
}

func (r *EnhanceRows) Err() error {
	if r.err != nil {
		return r.err
	}
	return r.Rows.Err()
}
