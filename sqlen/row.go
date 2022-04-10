package sqlen

import (
	"database/sql"
)

// EnhanceRow 用于在 Enhanced 方法中替换元生的 sql.Row。
// 注意这里原生的 sql.Row 方法没有开放内部 sql.Rows 结构出来，这里直接通过内嵌 *EnhanceRows 实现。
type EnhanceRow struct {
	rows *EnhanceRows
	err  error
}

// Scan 用于把一行数据填充到 map 中。
func (r *EnhanceRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	if r.rows == nil {
		r.err = sql.ErrNoRows
		return r.err
	}

	defer r.rows.Close()
	if !r.rows.Next() {
		if r.err = r.rows.Err(); r.err != nil {
			return r.err
		}

		r.err = sql.ErrNoRows
		return r.err
	}

	if r.err = r.rows.Scan(dest...); r.err != nil {
		return r.err
	}

	return r.rows.Close()
}

// MapScan 用于把一行数据填充到 map 中。
func (r *EnhanceRow) MapScan() (map[string]any, error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.rows == nil {
		r.err = sql.ErrNoRows
		return nil, r.err
	}

	defer r.rows.Close()
	if !r.rows.Next() {
		if r.err = r.rows.Err(); r.err != nil {
			return nil, r.err
		}

		r.err = sql.ErrNoRows
		return nil, r.err
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
		r.err = sql.ErrNoRows
		return nil, r.err
	}

	defer r.rows.Close()
	if !r.rows.Next() {
		if r.err = r.rows.Err(); r.err != nil {
			return nil, r.err
		}

		r.err = sql.ErrNoRows
		return nil, r.err
	}

	if res, err := r.rows.SliceScan(); err != nil {
		return nil, err
	} else {
		return res, r.rows.Close()
	}
}

func (r *EnhanceRow) Err() error {
	if r.err != nil {
		return r.err
	}

	r.err = r.rows.Err()
	return r.err
}
