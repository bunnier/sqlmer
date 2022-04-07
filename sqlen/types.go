package sqlen

import (
	"database/sql"
	"reflect"
)

// extractNullableValue 用于将 nullable 类型提取为实际值。
func extractNullableValue(columnType *sql.ColumnType, srcP *interface{}) {
	switch src := (*srcP).(type) {
	case sql.NullBool:
		if !src.Valid {
			*srcP = nil
			break
		}
		*srcP = src.Bool
	case sql.NullByte:
		if !src.Valid {
			*srcP = nil
			break
		}
		*srcP = src.Byte
	case sql.NullFloat64:
		if !src.Valid {
			*srcP = nil
			break
		}
		*srcP = src.Float64
	case sql.NullInt16:
		if !src.Valid {
			*srcP = nil
			break
		}
		*srcP = src.Int16
	case sql.NullInt32:
		if !src.Valid {
			*srcP = nil
			break
		}
		*srcP = src.Int32
	case sql.NullInt64:
		if !src.Valid {
			*srcP = nil
			break
		}
		*srcP = src.Int64
	case sql.NullString:
		if !src.Valid {
			*srcP = nil
			break
		}
		*srcP = src.String
	case sql.NullTime:
		if !src.Valid {
			*srcP = nil
			break
		}
		*srcP = src.Time
	case sql.RawBytes:
		if src == nil {
			*srcP = nil
		}
	}
}

var (
	scanTypeFloat32 = reflect.TypeOf(float32(0))
	scanTypeFloat64 = reflect.TypeOf(float64(0))

	scanTypeInt8  = reflect.TypeOf(int8(0))
	scanTypeInt16 = reflect.TypeOf(int16(0))
	scanTypeInt32 = reflect.TypeOf(int32(0))
	scanTypeInt64 = reflect.TypeOf(int64(0))

	scanTypeUint8  = reflect.TypeOf(uint8(0))
	scanTypeUint16 = reflect.TypeOf(uint16(0))
	scanTypeUint32 = reflect.TypeOf(uint32(0))
	scanTypeUint64 = reflect.TypeOf(uint64(0))
)

// 为了统一可空字段和非可空字段的返回值，这里统一将数字类型提升到 nullable 支持的类型。
func unifyNumber(val reflect.Type) reflect.Type {
	switch val {
	// nullable 的 int 只支持 int64。
	case scanTypeInt8:
		return scanTypeInt64
	case scanTypeInt16:
		return scanTypeInt64
	case scanTypeInt32:
		return scanTypeInt64

	// nullable 的 uint 只支持 int64。
	case scanTypeUint8:
		return scanTypeInt64
	case scanTypeUint16:
		return scanTypeInt64
	case scanTypeUint64:
		return scanTypeInt64

	// nullable 的 int 只支持 float64。
	case scanTypeFloat32:
		return scanTypeFloat64

	default:
		return val
	}
}
