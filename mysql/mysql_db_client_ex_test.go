package mysql_test

import (
	"reflect"
	"testing"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
)

func getClientEx(t *testing.T) *sqlmer.DbClientEx {
	c, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatalf("cannot get the client: %v", err)
	}
	ex := sqlmer.Extend(c)
	return ex
}

func TestDbClientEx_ScalarString(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		const expect = "a"

		v, ok, err := c.ScalarString("SELECT @p1", expect)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		if v == nil {
			t.Fatalf("expect value, got nil")
		}

		if *v != expect {
			t.Fatalf("expect value=%v, got %v", expect, *v)
		}
	})

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarString("SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("conv", func(t *testing.T) {
		v, _, _ := c.ScalarString("SELECT @v", map[string]any{"v": 123456})
		if *v != "123456" {
			t.Fatalf("expect value=%v, got %v", "123456", *v)
		}
	})

	t.Run("err", func(t *testing.T) {
		v, ok, err := c.ScalarString("sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})
}
func TestDbClientEx_MustScalarString(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarString("sql-error")
}

func TestDbClientEx_ScalarInt(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		const expect = 123

		v, ok, err := c.ScalarInt("SELECT @p1", expect)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		if v == nil {
			t.Fatalf("expect value, got nil")
		}

		if *v != expect {
			t.Fatalf("expect value=%v, got %v", expect, *v)
		}
	})

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarInt("SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("conv", func(t *testing.T) {
		v, _, _ := c.ScalarInt("SELECT @v", map[string]any{"v": "1122"})
		if *v != 1122 {
			t.Fatalf("expect value=%v, got %v", 1122, *v)
		}
	})

	t.Run("err", func(t *testing.T) {
		v, ok, err := c.ScalarInt("sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})
}

func TestDbClientEx_MustScalarInt(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarInt("sql-error")
}
func TestDbClientEx_ScalarInt64(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		const expect = int64(123)

		v, ok, err := c.ScalarInt64("SELECT @p1", expect)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		if v == nil {
			t.Fatalf("expect value, got nil")
		}

		if *v != expect {
			t.Fatalf("expect value=%v, got %v", expect, *v)
		}
	})

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarInt64("SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("conv", func(t *testing.T) {
		v, _, _ := c.ScalarInt64("SELECT @v", map[string]any{"v": "1122"})
		if *v != 1122 {
			t.Fatalf("expect value=%v, got %v", 1122, *v)
		}
	})

	t.Run("err", func(t *testing.T) {
		v, ok, err := c.ScalarInt64("sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})
}

func TestDbClientEx_MustScalarInt64(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarInt64("sql-error")
}
func TestDbClientEx_ScalarInt32(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		const expect = int32(123)

		v, ok, err := c.ScalarInt32("SELECT @p1", expect)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		if v == nil {
			t.Fatalf("expect value, got nil")
		}

		if *v != expect {
			t.Fatalf("expect value=%v, got %v", expect, *v)
		}
	})

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarInt32("SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("conv", func(t *testing.T) {
		v, _, _ := c.ScalarInt32("SELECT @v", map[string]any{"v": "1122"})
		if *v != 1122 {
			t.Fatalf("expect value=%v, got %v", 1122, *v)
		}
	})

	t.Run("err", func(t *testing.T) {
		v, ok, err := c.ScalarInt32("sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})
}

func TestDbClientEx_MustScalarInt32(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarInt32("sql-error")
}
func TestDbClientEx_ScalarInt16(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		const expect = int16(123)

		v, ok, err := c.ScalarInt16("SELECT @p1", expect)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		if v == nil {
			t.Fatalf("expect value, got nil")
		}

		if *v != expect {
			t.Fatalf("expect value=%v, got %v", expect, *v)
		}
	})

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarInt16("SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("conv", func(t *testing.T) {
		v, _, _ := c.ScalarInt16("SELECT @v", map[string]any{"v": "1122"})
		if *v != 1122 {
			t.Fatalf("expect value=%v, got %v", 1122, *v)
		}
	})

	t.Run("err", func(t *testing.T) {
		v, ok, err := c.ScalarInt16("sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})
}

func TestDbClientEx_MustScalarInt16(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarInt16("sql-error")
}
func TestDbClientEx_ScalarInt8(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		const expect = int8(123)

		v, ok, err := c.ScalarInt8("SELECT @p1", expect)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		if v == nil {
			t.Fatalf("expect value, got nil")
		}

		if *v != expect {
			t.Fatalf("expect value=%v, got %v", expect, *v)
		}
	})

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarInt8("SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("conv", func(t *testing.T) {
		v, _, _ := c.ScalarInt8("SELECT @v", map[string]any{"v": "33"})
		if *v != 33 {
			t.Fatalf("expect value=%v, got %v", 33, *v)
		}
	})

	t.Run("err", func(t *testing.T) {
		v, ok, err := c.ScalarInt8("sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})
}

func TestDbClientEx_MustScalarInt8(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarInt8("sql-error")
}

func TestDbClientEx_ScalarBool(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		const expect = true

		v, ok, err := c.ScalarBool("SELECT @p1", expect)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		if v == nil {
			t.Fatalf("expect value, got nil")
		}

		if *v != expect {
			t.Fatalf("expect value=%v, got %v", expect, *v)
		}
	})

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarBool("SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("conv", func(t *testing.T) {
		v, _, _ := c.ScalarBool("SELECT @v", map[string]any{"v": 100})
		if !*v {
			t.Fatalf("expect value=%v, got %v", true, *v)
		}
	})

	t.Run("err", func(t *testing.T) {
		v, ok, err := c.ScalarBool("sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})
}

func TestDbClientEx_MustScalarBool(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarBool("sql-error")
}

func TestDbClientEx_ScalarFloat32(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		const expect = float32(0.5)

		v, ok, err := c.ScalarFloat32("SELECT @p1", expect)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		if v == nil {
			t.Fatalf("expect value, got nil")
		}

		if *v != expect {
			t.Fatalf("expect value=%v, got %v", expect, *v)
		}
	})

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarFloat32("SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("conv", func(t *testing.T) {
		v, _, _ := c.ScalarFloat32("SELECT @v", map[string]any{"v": "-0.5"})
		if *v != float32(-0.5) {
			t.Fatalf("expect value=%v, got %v", -0.5, *v)
		}
	})

	t.Run("err", func(t *testing.T) {
		v, ok, err := c.ScalarFloat32("sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})
}

func TestDbClientEx_MustScalarFloat32(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarFloat32("sql-error")
}

func TestDbClientEx_ScalarFloat64(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		const expect = float64(0.5)

		v, ok, err := c.ScalarFloat64("SELECT @p1", expect)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		if v == nil {
			t.Fatalf("expect value, got nil")
		}

		if *v != expect {
			t.Fatalf("expect value=%v, got %v", expect, *v)
		}
	})

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarFloat64("SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("conv", func(t *testing.T) {
		v, _, _ := c.ScalarFloat64("SELECT @v", map[string]any{"v": "-0.5"})
		if *v != float64(-0.5) {
			t.Fatalf("expect value=%v, got %v", -0.5, *v)
		}
	})

	t.Run("err", func(t *testing.T) {
		v, ok, err := c.ScalarFloat64("sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})
}

func TestDbClientEx_MustScalarFloat64(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarFloat64("sql-error")
}

func TestDbClientEx_GetStruct(t *testing.T) {
	c := getClientEx(t)

	type rowType struct {
		IntTest       int
		VarcharTest   string
		NullFloatTest *float64
	}

	t.Run("hit", func(t *testing.T) {
		var got rowType
		ok, err := c.GetStruct(&got, "SELECT intTest, varcharTest, nullFloatTest FROM go_TypeTest WHERE id=@p1", 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		expect := rowType{
			IntTest:     1,
			VarcharTest: "行1",
		}
		if !reflect.DeepEqual(expect, got) {
			t.Fatalf("expect %#v, got %#v", expect, got)
		}
	})

	t.Run("miss", func(t *testing.T) {
		var got rowType
		ok, err := c.GetStruct(&got, "SELECT intTest, varcharTest, nullFloatTest FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}
	})

	t.Run("err", func(t *testing.T) {
		var got rowType
		ok, err := c.GetStruct(&got, "sql-error")
		if err == nil {
			t.Fatalf("expect error, got nil")
		}

		if ok {
			t.Fatalf("expect ok=false")
		}
	})
}

func TestDbClientEx_MustGetStruct(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	var v struct{}
	c.MustGetStruct(&v, "sql-error")
}

func TestDbClientEx_ScalarType(t *testing.T) {
	c := getClientEx(t)

	t.Run("miss", func(t *testing.T) {
		v, ok, err := c.ScalarType(reflect.TypeOf(0), "SELECT 'a' FROM go_TypeTest WHERE 1=0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ok {
			t.Fatalf("expect ok=false")
		}

		if v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("null", func(t *testing.T) {
		typ := reflect.TypeOf((*string)(nil))
		v, ok, err := c.ScalarType(typ, "SELECT null")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		s, ok := v.(*string)
		if !ok {
			t.Fatalf("expect *string")
		}

		if s != nil {
			t.Fatalf("expect nil, got %v", s)
		}
	})

	t.Run("string", func(t *testing.T) {
		v, ok, err := c.ScalarType(reflect.TypeOf(""), "SELECT @v", map[string]any{"v": 1122})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		s, ok := v.(string)
		if !ok {
			t.Fatalf("expect string")
		}

		if s != "1122" {
			t.Fatalf("expect 1122, got %v", s)
		}
	})

	t.Run("float32", func(t *testing.T) {
		v, ok, err := c.ScalarType(reflect.TypeOf(float32(0)), "SELECT @v", map[string]any{"v": 1122})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		s, ok := v.(float32)
		if !ok {
			t.Fatalf("expect string")
		}

		if s != 1122 {
			t.Fatalf("expect 1122, got %v", s)
		}
	})
}

func TestDbClientEx_MustScalarType(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarType(reflect.TypeOf(0), "error-sql")
}

func TestDbClientEx_ScalarOf(t *testing.T) {
	c := getClientEx(t)

	t.Run("string", func(t *testing.T) {
		v, ok, err := c.ScalarOf("", "SELECT @v", map[string]any{"v": 1122})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatalf("expect ok=true")
		}

		s, ok := v.(string)
		if !ok {
			t.Fatalf("expect string")
		}

		if s != "1122" {
			t.Fatalf("expect 1122, got %v", s)
		}
	})
}

func TestDbClientEx_MustScalarOf(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustScalarOf(0, "error-sql")
}

func TestDbClientEx_ListType(t *testing.T) {
	c := getClientEx(t)

	type rowType struct {
		Id           string
		IntTest      *int
		V            string `conv:"varcharTest"`
		CharTest     *string
		NullTextTest *string
	}

	t.Run("hit", func(t *testing.T) {
		query := `SELECT id, intTest, varcharTest, charTest, nullTextTest FROM go_TypeTest WHERE id IN (@p1)`
		res, err := c.ListType(reflect.TypeOf(new(rowType)), query, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		list, ok := res.([]*rowType)
		if !ok {
			t.Fatalf("expect []*rowType, got %t", res)
		}

		if len(list) != 1 {
			t.Fatalf("expect length=1, got %v", len(list))
		}

		v := *list[0]
		if v.Id != "1" {
			t.Fatalf("expect ID=1, got %v", v.Id)
		}

		if *v.IntTest != 1 {
			t.Fatalf("expect IntTest=1, got %v", v.IntTest)
		}

		if v.V != "行1" {
			t.Fatalf("expect V=行1, got %v", v.V)
		}

		if *v.CharTest != "行1char" {
			t.Fatalf("expect CharTest=行1char, got %v", v.CharTest)
		}

		if v.NullTextTest != nil {
			t.Fatalf("expect NullTextTest=nil, got %v", v.NullTextTest)
		}
	})

	t.Run("miss", func(t *testing.T) {
		query := `SELECT id, intTest, varcharTest, charTest, nullTextTest FROM go_TypeTest WHERE id IN (-1)`
		res, err := c.ListType(reflect.TypeOf(new(rowType)), query)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		list, ok := res.([]*rowType)
		if !ok {
			t.Fatalf("expect []*rowType, got %t", res)
		}

		if len(list) != 0 {
			t.Fatalf("expect length=0, got %v", len(list))
		}
	})
}

func TestDbClientEx_MustListType(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustListType(reflect.TypeOf(0), "error-sql")
}

func TestDbClientEx_ListOf(t *testing.T) {
	c := getClientEx(t)

	t.Run("hit", func(t *testing.T) {
		query := `SELECT id, intTest, varcharTest, charTest, nullTextTest FROM go_TypeTest WHERE id IN (@p1, @p2, @p3, @p4)`
		res, err := c.ListOf(0, query, 1, 2, 3, 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		list, ok := res.([]int)
		if !ok {
			t.Fatalf("expect []*rowType, got %t", res)
		}

		expect := []int{1, 2, 3, 4}
		if !reflect.DeepEqual(expect, list) {
			t.Fatalf("expect expect %v, got %v", expect, list)
		}
	})
}

func TestDbClientEx_MustListOf(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("should panic")
		}
	}()

	c := getClientEx(t)
	c.MustListOf(0, "error-sql")
}
