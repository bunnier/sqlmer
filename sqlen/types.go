package sqlen

import (
	"database/sql"
	"reflect"
	"time"
)

// extractNullableColumnValue 用于将 nullable 类型提取为实际值。
func extractNullableColumnValue(columnType *sql.ColumnType, srcP *interface{}) {
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
			break
		}
		*srcP = src
	}
}

var (
	scanTypeFloat32   = reflect.TypeOf(float32(0))
	scanTypeFloat64   = reflect.TypeOf(float64(0))
	scanTypeNullFloat = reflect.TypeOf(sql.NullFloat64{})

	scanTypeInt8    = reflect.TypeOf(int8(0))
	scanTypeInt16   = reflect.TypeOf(int16(0))
	scanTypeInt32   = reflect.TypeOf(int32(0))
	scanTypeInt64   = reflect.TypeOf(int64(0))
	scanTypeUint8   = reflect.TypeOf(uint8(0))
	scanTypeUint16  = reflect.TypeOf(uint16(0))
	scanTypeUint32  = reflect.TypeOf(uint32(0))
	scanTypeUint64  = reflect.TypeOf(uint64(0))
	scanTypeNullInt = reflect.TypeOf(sql.NullInt64{})

	scanTypeNullTime = reflect.TypeOf(sql.NullTime{})
	scanTypeTime     = reflect.TypeOf(time.Time{})

	scanTypeString     = reflect.TypeOf("")
	scanTypeNullString = reflect.TypeOf(sql.NullString{})

	scanTypeBool     = reflect.TypeOf(false)
	scanTypeNullBool = reflect.TypeOf(sql.NullBool{})

	scanTypeByte     = reflect.TypeOf(byte(0))
	scanTypeNullByte = reflect.TypeOf(sql.NullByte{})
)

// 为了统一可空字段和非可空字段的返回值，这里统一将数字类型提升到 nullable 支持的类型。
func unifyScanType(columnType *sql.ColumnType) reflect.Type {
	nullable, ok := columnType.Nullable()
	nullable = nullable && ok

	scanType := columnType.ScanType()

	if !nullable { // 为了 nullable 和 not null 类型一致，这里对 not null 类型统一做个类型转换。
		switch scanType {
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
		case scanTypeUint32:
			return scanTypeInt64

		// nullable 的 int 只支持 float64。
		case scanTypeFloat32:
			return scanTypeFloat64

		default:
			return scanType
		}
	} else {
		// 将可空类型转为对应的 Null 类型处理。
		switch scanType {
		case scanTypeInt8:
			return scanTypeNullInt
		case scanTypeInt16:
			return scanTypeNullInt
		case scanTypeInt32:
			return scanTypeNullInt
		case scanTypeInt64:
			return scanTypeNullInt

		case scanTypeUint8:
			return scanTypeNullInt
		case scanTypeUint16:
			return scanTypeNullInt
		case scanTypeUint32:
			return scanTypeNullInt
		case scanTypeUint64:
			return scanTypeNullInt

		case scanTypeFloat32:
			return scanTypeNullFloat
		case scanTypeFloat64:
			return scanTypeNullFloat

		case scanTypeTime:
			return scanTypeNullTime

		case scanTypeBool:
			return scanTypeNullBool

		case scanTypeString:
			return scanTypeNullString

		case scanTypeByte:
			return scanTypeNullByte

		default:
			return scanType
		}
	}

}
