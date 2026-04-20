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

// SetErrWrapper 用于为延迟暴露的错误注入统一包装逻辑。
func (r *EnhanceRow) SetErrWrapper(wrapper ErrWrapper) {
	if r.rows != nil {
		r.rows.SetErrWrapper(wrapper)
	}
}

// Scan 用于把一行数据填充到 map 中。
func (r *EnhanceRow) Scan(dest ...any) (err error) {
	if r.err != nil {
		return r.err
	}

	if r.rows == nil {
		r.err = sql.ErrNoRows
		return r.err
	}

	defer func() {
		closeErr := r.rows.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			r.err = err
		}
	}()

	if !r.rows.Next() {
		if err = r.rows.Err(); err != nil {
			return err
		}

		r.err = sql.ErrNoRows
		return r.err
	}

	if err = r.rows.Scan(dest...); err != nil {
		return err
	}

	return nil
}

// MapScan 用于把一行数据填充到 map 中。
func (r *EnhanceRow) MapScan() (res map[string]any, err error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.rows == nil {
		r.err = sql.ErrNoRows
		return nil, r.err
	}

	defer func() {
		closeErr := r.rows.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			r.err = err
		}
	}()

	if !r.rows.Next() {
		if err = r.rows.Err(); err != nil {
			return nil, err
		}

		r.err = sql.ErrNoRows
		return nil, r.err
	}

	res, err = r.rows.MapScan()
	return res, err
}

// SliceScan 用 Slice 返回一行数据。
func (r *EnhanceRow) SliceScan() (res []any, err error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.rows == nil {
		r.err = sql.ErrNoRows
		return nil, r.err
	}

	defer func() {
		closeErr := r.rows.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			r.err = err
		}
	}()

	if !r.rows.Next() {
		if err = r.rows.Err(); err != nil {
			return nil, err
		}

		r.err = sql.ErrNoRows
		return nil, r.err
	}

	res, err = r.rows.SliceScan()
	return res, err
}

func (r *EnhanceRow) Err() error {
	if r.err != nil {
		return r.err
	}

	if r.rows == nil {
		r.err = sql.ErrNoRows
		return r.err
	}

	r.err = r.rows.Err()
	return r.err
}
