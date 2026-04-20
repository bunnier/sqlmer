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
	wrapErr       ErrWrapper

	columnMetaSlice []*columnMeta // 用于对列的元数据做缓存。

	err error
}

type columnMeta struct {
	colType  *sql.ColumnType // 列类型。
	scanType reflect.Type    // 用于 Scan 的类型。
}

// ErrWrapper 用于对延迟暴露的查询错误进行统一包装。
type ErrWrapper func(error) error

// SetErrWrapper 用于为延迟暴露的错误注入统一包装逻辑。
func (rs *EnhanceRows) SetErrWrapper(wrapper ErrWrapper) {
	rs.wrapErr = wrapper
}

func (rs *EnhanceRows) wrap(err error) error {
	if err == nil {
		return nil
	}
	if rs.wrapErr == nil {
		return err
	}
	return rs.wrapErr(err)
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
	rs.err = rs.wrap(rs.Rows.Scan(dest...))
	return rs.err
}

// MapScan 用于把一行数据填充到 map 中。
func (rs *EnhanceRows) MapScan() (map[string]any, error) {
	if sliceRes, err := rs.SliceScan(); err != nil {
		return nil, err
	} else {
		res := make(map[string]any, len(sliceRes))
		for i, columnMeta := range rs.columnMetaSlice {
			res[columnMeta.colType.Name()] = sliceRes[i]
		}
		return res, nil
	}
}

// SliceScan 用 Slice 的方式返回一行数据。
func (rs *EnhanceRows) SliceScan() ([]any, error) {
	if rs.err != nil {
		return nil, rs.err
	}

	if rs.err = rs.initColumns(); rs.err != nil {
		return nil, rs.err
	}

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

	if rs.err = rs.Scan(dest...); rs.err != nil {
		return nil, rs.err
	}

	for i := 0; i < len(rs.columnMetaSlice); i++ {
		// 之前为了能让 Scan 修改数据，保存的是指针，而返回上层时候只需要实际数据，因此进行解引用。
		dest[i] = destRefVal[i].Elem().Interface()

		// 注意：根据 sql.RawBytes 类型的定义，database/sql 的有效性仅在下一次 Scan 或 Close 之前。
		// 如果将该 sql.RawBytes 直接向上层暴露，会因为之后被覆写，导致数据损坏。
		// 因此这里必须对 sql.RawBytes 做一次 deep copy，将数据拷贝到独立内存中。
		if raw, ok := dest[i].(sql.RawBytes); ok {
			if raw == nil {
				dest[i] = nil
			} else {
				frozen := make(sql.RawBytes, len(raw))
				copy(frozen, raw)
				dest[i] = frozen
			}
		}

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
	r.err = r.wrap(r.Rows.Err())
	return r.err
}

// Close 关闭游标，并对关闭过程中暴露的错误做统一包装。
func (r *EnhanceRows) Close() error {
	if r.Rows == nil {
		return nil
	}
	closeErr := r.wrap(r.Rows.Close())
	if r.err != nil {
		return r.err
	}
	r.err = closeErr
	return r.err
}
