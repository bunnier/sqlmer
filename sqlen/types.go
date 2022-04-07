package sqlen

import (
	"database/sql"
)

// extractNullableValue 用于将 nullable 类型提取为实际值。
func extractNullableValue(srcP *interface{}) {
	switch src := (*srcP).(type) {
	case sql.NullBool:
		if !src.Valid {
			*srcP = nil
		}
		*srcP = src.Bool
	case sql.NullByte:
		if !src.Valid {
			*srcP = nil
		}
		*srcP = src.Byte
	case sql.NullFloat64:
		if !src.Valid {
			*srcP = nil
		}
		*srcP = src.Float64
	case sql.NullInt16:
		if !src.Valid {
			*srcP = nil
		}
		*srcP = src.Int16
	case sql.NullInt32:
		if !src.Valid {
			*srcP = nil
		}
		*srcP = src.Int32
	case sql.NullInt64:
		if !src.Valid {
			*srcP = nil
		}
		*srcP = src.Int64
	case sql.NullString:
		if !src.Valid {
			*srcP = nil
		}
		*srcP = src.String
	case sql.NullTime:
		if !src.Valid {
			*srcP = nil
		}
		*srcP = src.Time
	}
}
