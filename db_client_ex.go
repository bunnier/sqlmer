package sqlmer

import (
	"io"
	"reflect"

	"github.com/cmstar/go-conv"
)

// DbClientEx 加强 DbClient ，提供强类型的转化方法。
type DbClientEx struct {
	DbClient  // 原始的 DbClient 实例。
	conv.Conv // 为当前实例提供类型转换的 conv.Conv 实例。
}

// NewDbClientEx 创建 EnhancedDbClient 的新实例。
// 其加强 DbClient ，提供强类型的转化方法。
func NewDbClientEx(raw DbClient) *DbClientEx {
	// 提供 mysql 的 snake_case 名称的字段到 Go 的 CamelCase 字段的匹配。
	dbConv := conv.Conv{
		Conf: conv.Config{
			FieldMatcherCreator: &conv.SimpleMatcherCreator{
				Conf: conv.SimpleMatcherConfig{
					Tag:            "conv",
					CamelSnakeCase: true,
				},
			},
		},
	}
	return &DbClientEx{raw, dbConv}
}

// GetStruct 获取一行的查询结果，转化并填充到 ptr 。 ptr 必须是 struct 类型的指针。
func (c *DbClientEx) GetStruct(ptr any, query string, args ...any) error {
	m, err := c.Get(query, args...)
	if err != nil {
		return err
	}

	err = c.Conv.Convert(m, ptr)
	if err != nil {
		return err
	}

	return nil
}

// MustGetStruct 类似 GetStruct ，但出现错误时不返回 error ，而是 panic 。
func (c *DbClientEx) MustGetStruct(ptr any, query string, args ...any) {
	err := c.GetStruct(ptr, query, args...)
	if err != nil {
		panic(err)
	}
}

// ScalarString 查询第一行第一列，并返回目标类型的值。
// 若查询没有命中行，返回空指针和 ok=false ；若有结果但值是 null ，则返回空指针和 ok=true 。
// 若值不是目标类型，则尝试转换类型。
func (c *DbClientEx) ScalarString(query string, args ...any) (value *string, ok bool, err error) {
	v, ok, err := c.Scalar(query, args...)
	if !ok || err != nil {
		return
	}

	var res string
	err = c.Conv.Convert(v, &res)
	if err != nil {
		return
	}

	value = &res
	return
}

// MustScalarString 类似 ScalarString ，但出现错误时不返回 error ，而是 panic 。
func (c *DbClientEx) MustScalarString(query string, args ...any) (value *string, ok bool) {
	value, ok, err := c.ScalarString(query, args...)
	if err != nil {
		panic(err)
	}
	return
}

// ScalarInt 查询第一行第一列，并返回目标类型的值。
// 若查询没有命中行，返回空指针和 ok=false ；若有结果但值是 null ，则返回空指针和 ok=true 。
// 若值不是目标类型，则尝试转换类型。
func (c *DbClientEx) ScalarInt(query string, args ...any) (value *int, ok bool, err error) {
	v, ok, err := c.Scalar(query, args...)
	if !ok || err != nil {
		return
	}

	var res int
	err = c.Conv.Convert(v, &res)
	if err != nil {
		return
	}

	value = &res
	return
}

// MustScalarInt 类似 ScalarInt ，但出现错误时不返回 error ，而是 panic 。
func (c *DbClientEx) MustScalarInt(query string, args ...any) (value *int, ok bool) {
	value, ok, err := c.ScalarInt(query, args...)
	if err != nil {
		panic(err)
	}
	return
}

// ScalarInt64 查询第一行第一列，并返回目标类型的值。
// 若查询没有命中行，返回空指针和 ok=false ；若有结果但值是 null ，则返回空指针和 ok=true 。
// 若值不是目标类型，则尝试转换类型。
func (c *DbClientEx) ScalarInt64(query string, args ...any) (value *int64, ok bool, err error) {
	v, ok, err := c.Scalar(query, args...)
	if !ok || err != nil {
		return
	}

	var res int64
	err = c.Conv.Convert(v, &res)
	if err != nil {
		return
	}

	value = &res
	return
}

// MustScalarInt64 类似 ScalarInt64 ，但出现错误时不返回 error ，而是 panic 。
func (c *DbClientEx) MustScalarInt64(query string, args ...any) (value *int64, ok bool) {
	value, ok, err := c.ScalarInt64(query, args...)
	if err != nil {
		panic(err)
	}
	return
}

// ScalarInt32 查询第一行第一列，并返回目标类型的值。
// 若查询没有命中行，返回空指针和 ok=false ；若有结果但值是 null ，则返回空指针和 ok=true 。
// 若值不是目标类型，则尝试转换类型。
func (c *DbClientEx) ScalarInt32(query string, args ...any) (value *int32, ok bool, err error) {
	v, ok, err := c.Scalar(query, args...)
	if !ok || err != nil {
		return
	}

	var res int32
	err = c.Conv.Convert(v, &res)
	if err != nil {
		return
	}

	value = &res
	return
}

// MustScalarInt32 类似 ScalarInt32 ，但出现错误时不返回 error ，而是 panic 。
func (c *DbClientEx) MustScalarInt32(query string, args ...any) (value *int32, ok bool) {
	value, ok, err := c.ScalarInt32(query, args...)
	if err != nil {
		panic(err)
	}
	return
}

// ScalarInt16 查询第一行第一列，并返回目标类型的值。
// 若查询没有命中行，返回空指针和 ok=false ；若有结果但值是 null ，则返回空指针和 ok=true 。
// 若值不是目标类型，则尝试转换类型。
func (c *DbClientEx) ScalarInt16(query string, args ...any) (value *int16, ok bool, err error) {
	v, ok, err := c.Scalar(query, args...)
	if !ok || err != nil {
		return
	}

	var res int16
	err = c.Conv.Convert(v, &res)
	if err != nil {
		return
	}

	value = &res
	return
}

// MustScalarInt16 类似 ScalarInt16 ，但出现错误时不返回 error ，而是 panic 。
func (c *DbClientEx) MustScalarInt16(query string, args ...any) (value *int16, ok bool) {
	value, ok, err := c.ScalarInt16(query, args...)
	if err != nil {
		panic(err)
	}
	return
}

// ScalarInt8 查询第一行第一列，并返回目标类型的值。
// 若查询没有命中行，返回空指针和 ok=false ；若有结果但值是 null ，则返回空指针和 ok=true 。
// 若值不是目标类型，则尝试转换类型。
func (c *DbClientEx) ScalarInt8(query string, args ...any) (value *int8, ok bool, err error) {
	v, ok, err := c.Scalar(query, args...)
	if !ok || err != nil {
		return
	}

	var res int8
	err = c.Conv.Convert(v, &res)
	if err != nil {
		return
	}

	value = &res
	return
}

// MustScalarInt8 类似 ScalarInt8 ，但出现错误时不返回 error ，而是 panic 。
func (c *DbClientEx) MustScalarInt8(query string, args ...any) (value *int8, ok bool) {
	value, ok, err := c.ScalarInt8(query, args...)
	if err != nil {
		panic(err)
	}
	return
}

// ScalarBool 查询第一行第一列，并返回目标类型的值。
// 若查询没有命中行，返回空指针和 ok=false ；若有结果但值是 null ，则返回空指针和 ok=true 。
// 若值不是目标类型，则尝试转换类型。
func (c *DbClientEx) ScalarBool(query string, args ...any) (value *bool, ok bool, err error) {
	v, ok, err := c.Scalar(query, args...)
	if !ok || err != nil {
		return
	}

	var res bool
	err = c.Conv.Convert(v, &res)
	if err != nil {
		return
	}

	value = &res
	return
}

// MustScalarBool 类似 ScalarBool ，但出现错误时不返回 error ，而是 panic 。
func (c *DbClientEx) MustScalarBool(query string, args ...any) (value *bool, ok bool) {
	value, ok, err := c.ScalarBool(query, args...)
	if err != nil {
		panic(err)
	}
	return
}

// ScalarType 查询第一行第一列，并返回目标类型的值。
// 若查询没有命中行，返回空指针和 ok=false ；若有结果但值是 null ，则返回空指针和 ok=true 。
// 若值不是目标类型，则尝试转换类型。
func (c *DbClientEx) ScalarType(typ reflect.Type, query string, args ...any) (value any, ok bool, err error) {
	v, ok, err := c.Scalar(query, args...)
	if !ok && err != nil {
		return
	}

	value, err = c.Conv.ConvertType(v, typ)
	return
}

// MustScalarType 类似 ScalarType ，但出现错误时不返回 error ，而是 panic 。
func (c *DbClientEx) MustScalarType(typ reflect.Type, query string, args ...any) (value any, ok bool) {
	value, ok, err := c.ScalarType(typ, query, args...)
	if err != nil {
		panic(err)
	}
	return
}

// ListOf 将查询结果的每一行转换到指定类型。返回转换后的元素的列表，需给定列表中的元素的类型。若查询没有命中行，返回空集。
//
// 注意，给定的 elemTyp 是元素的类型，返回的则是该元素的 slice ：
//   list, err := ListType(reflect.TypeOf(0), query)
//   if err != nil {
//     return err
//   }
//   infos := list.([]int)
//
func (c *DbClientEx) ListType(elemTyp reflect.Type, query string, args ...any) (any, error) {
	rows, err := c.Rows(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // This error is ignored.

	underTyp := elemTyp
	for underTyp.Kind() == reflect.Ptr {
		underTyp = underTyp.Elem()
	}
	complex := underTyp.Kind() == reflect.Struct || underTyp.Kind() == reflect.Map
	vList := reflect.MakeSlice(reflect.SliceOf(elemTyp), 0, 0)

	for rows.Next() {
		var row any

		// 目标类型是复杂类型的，统一（目前也只能）将行转到 map ，再从 map 转换。
		// 非复杂类型（就是数字、字符串这些），则从每一行的第一列的值转换。
		// 未来， conv 库应该支持类似字段迭代器的功能，就可以不走 map 了，节省一层转换开销。
		if complex {
			m := make(map[string]any)
			err = rows.MapScan(m)
			if err != nil {
				return nil, err
			}
			row = m
		} else {
			// 当前 Scan() 方法会返回驱动相关的类型，而不会自动转换到 go 标准类型，导致 conv 不能。
			// 但 MapScan()/SliceScan() 会做转换。为了能用到这个转换，先取 slice 然后取第一个元素。
			// 以后下层库改进，可以 Scan() 出标准类型后，再直接用 Scan() 取数据。
			vals, err := rows.SliceScan()
			if err != nil {
				return nil, err
			}
			row = vals[0]
		}

		item, err := c.Conv.ConvertType(row, elemTyp)
		if err != nil {
			return nil, err
		}

		vList = reflect.Append(vList, reflect.ValueOf(item))
	}

	err = rows.Err()
	if err != nil && err != io.EOF {
		return nil, err
	}
	return vList.Interface(), nil
}

// MustListType 类似 ListType ，但出现错误时不返回 error ，而是 panic 。
//
// 注意，给定的 elemTyp 是元素的类型，返回的则是该元素的 slice ：
//   list := MustListType(reflect.TypeOf(0), query).([]int)
//
func (c *DbClientEx) MustListType(elemTyp reflect.Type, query string, args ...any) any {
	v, err := c.ListType(elemTyp, query, args...)
	if err != nil {
		panic(err)
	}
	return v
}

// ListOf 将查询结果的每一行转换到指定类型。返回转换后的元素的列表，需给定列表中的元素的类型。若查询没有命中行，返回空集。
//
// 注意，给定的 elemTyp 是元素的类型，返回的则是该元素的 slice ：
//   type Info struct { /* fields */ }
//   list, err := ListOf(new(Info), query)
//   if err != nil {
//     return err
//   }
//   infos := list.([]*Info)
//
func (c *DbClientEx) ListOf(elemExample any, query string, args ...any) (any, error) {
	elemTyp := reflect.TypeOf(elemExample)
	return c.ListType(elemTyp, query, args...)
}

// MustListOf 类似 ListOf ，但出现错误时不返回 error ，而是 panic 。
//   type Info struct { /* fields */ }
//   list := MustListOf(new(Info), query).([]*Info)
//
func (c *DbClientEx) MustListOf(elemExample any, query string, args ...any) any {
	elemTyp := reflect.TypeOf(elemExample)
	v, err := c.ListType(elemTyp, query, args...)
	if err != nil {
		panic(err)
	}
	return v
}
