package sqlen

import (
	"database/sql"
	"errors"
)

// EnhanceRow 用于在 Enhanced 方法中替换元生的 sql.Row。
// 注意这里原生的 sql.Row 方法没有开放内部 sql.Rows 结构出来，这里直接通过内嵌 *EnhanceRows 实现。
type EnhanceRow struct {
	rows *EnhanceRows
	err  error // 仅保存创建查询时的错误，与 sql.Row.Err 保持一致。
}

// Scan 将单行查询结果填充到目标变量中，调用后关闭游标。
func (r *EnhanceRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if r.rows == nil {
		return sql.ErrNoRows
	}

	defer r.rows.Close()

	// 与 sql.Row.Scan 一致，关闭游标后 RawBytes 的底层内存不再有效。
	for _, value := range dest {
		if _, ok := value.(*sql.RawBytes); ok {
			return errors.New("sql: RawBytes isn't allowed on Row.Scan")
		}
	}

	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}

		return sql.ErrNoRows
	}

	if err := r.rows.Scan(dest...); err != nil {
		return err
	}

	return r.rows.Close()
}

// MapScan 用于把一行数据填充到 map 中。
func (r *EnhanceRow) MapScan() (map[string]any, error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.rows == nil {
		return nil, sql.ErrNoRows
	}

	defer r.rows.Close()
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return nil, err
		}

		return nil, sql.ErrNoRows
	}

	if res, err := r.rows.MapScan(); err != nil {
		return nil, err
	} else {
		return res, r.rows.Close()
	}
}

// SliceScan 用 Slice 返回一行数据。
func (r *EnhanceRow) SliceScan() ([]any, error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.rows == nil {
		return nil, sql.ErrNoRows
	}

	defer r.rows.Close()
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return nil, err
		}

		return nil, sql.ErrNoRows
	}

	if res, err := r.rows.SliceScan(); err != nil {
		return nil, err
	} else {
		return res, r.rows.Close()
	}
}

// Err 返回创建查询时的错误；无匹配行或扫描失败通过 Scan 系列方法返回。
func (r *EnhanceRow) Err() error {
	return r.err
}
