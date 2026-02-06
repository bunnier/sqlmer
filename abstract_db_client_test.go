package sqlmer

import (
	"reflect"
	"testing"
	"time"
)

func Test_preHandleArgs(t *testing.T) {
	type TypeA struct {
		A string
	}

	type TypeAptr struct {
		Ap *string
	}

	type TypeAptr2 struct {
		TypeAptr
	}

	type TypeAptr3 struct {
		TypeAptr2
	}

	type TypeB struct {
		B string
	}

	type TypeIgnore struct {
		Ignore string
	}

	type TypeAB struct {
		TypeA
		TypeB
		*TypeIgnore // 不赋值此字段，它在作为参数时会被忽略掉。
	}

	type TypeABC struct {
		TypeAB
		C string
	}

	type TypeTimeA struct {
		TypeA
		Time time.Time
	}

	type TypePtrTimeA struct {
		TypeA
		TimePtr *time.Time
	}

	testTime := time.Date(2021, 7, 3, 0, 0, 0, 0, time.UTC)

	t.Run("empty_args", func(t *testing.T) {
		got, err := preHandleArgs()
		if err != nil {
			t.Errorf("preHandleArgs() error = %v, want no error", err)
			return
		}
		if len(got) != 0 {
			t.Errorf("preHandleArgs() = %v, want empty slice", got)
		}
	})

	t.Run("basic_type_slice", func(t *testing.T) {
		got, err := preHandleArgs([]int{1, 2, 3})
		if err != nil {
			t.Errorf("preHandleArgs() error = %v, want no error", err)
			return
		}
		if !reflect.DeepEqual(got, []any{[]int{1, 2, 3}}) {
			t.Errorf("preHandleArgs() = %v, want %v", got, []any{[]int{1, 2, 3}})
		}
	})

	t.Run("single_map", func(t *testing.T) {
		got, err := preHandleArgs(map[string]any{"T": "t"})
		if err != nil {
			t.Errorf("preHandleArgs() error = %v, want no error", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"T": "t"}}) {
			t.Errorf("preHandleArgs() = %v, want %v", got, []any{map[string]any{"T": "t"}})
		}
	})

	t.Run("multiple_maps", func(t *testing.T) {
		got, err := preHandleArgs(map[string]any{"T": "t"}, map[string]any{"T": "t2"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"T": "t2"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"T": "t2"}})
		}
	})

	t.Run("single_struct", func(t *testing.T) {
		got, err := preHandleArgs(TypeA{A: "a"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "a"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "a"}})
		}
	})

	t.Run("single_ptr_struct", func(t *testing.T) {
		pStr := "a"
		got, err := preHandleArgs(TypeAptr3{TypeAptr2: TypeAptr2{TypeAptr: TypeAptr{Ap: &pStr}}})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"Ap": pStr}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"Ap": "a"}})
		}

		checkEmbeddedField := func(v any) {
			got, err = preHandleArgs(v)
			if err != nil {
				t.Errorf("mergeArgs() error = %v, wantErr false", err)
				return
			}
			if !reflect.DeepEqual(got, []any{map[string]any{"Ap": nil}}) {
				t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"Ap": nil}})
			}
		}
		checkEmbeddedField(TypeAptr{})
		checkEmbeddedField(TypeAptr2{})
		checkEmbeddedField(TypeAptr3{})
	})

	t.Run("single_struct_pointer", func(t *testing.T) {
		got, err := preHandleArgs(&TypeA{A: "a"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "a"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "a"}})
		}
	})

	t.Run("multiple_structs", func(t *testing.T) {
		got, err := preHandleArgs(TypeA{A: "a"}, TypeB{B: "b"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "a", "B": "b"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "a", "B": "b"}})
		}
	})

	t.Run("multiple_struct_pointers", func(t *testing.T) {
		got, err := preHandleArgs(&TypeA{A: "a"}, &TypeB{B: "b"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "a", "B": "b"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "a", "B": "b"}})
		}
	})

	t.Run("nested_struct", func(t *testing.T) {
		got, err := preHandleArgs(TypeAB{TypeA: TypeA{A: "ab_a"}, TypeB: TypeB{B: "ab_b"}})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "ab_a", "B": "ab_b"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "ab_a", "B": "ab_b"}})
		}
	})

	t.Run("deep_nested_struct", func(t *testing.T) {
		got, err := preHandleArgs(TypeABC{TypeAB: TypeAB{TypeA: TypeA{A: "abc_a"}, TypeB: TypeB{B: "abc_b"}}, C: "abc_c"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "abc_a", "B": "abc_b", "C": "abc_c"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "abc_a", "B": "abc_b", "C": "abc_c"}})
		}
	})

	t.Run("time_convert", func(t *testing.T) {
		testTime := time.Now()
		paramTimeA := TypeTimeA{
			TypeA: TypeA{A: "abc_a"},
			Time:  testTime,
		}
		got, err := preHandleArgs(paramTimeA)
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "abc_a", "Time": testTime}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "abc_a", "Time": testTime}})
		}
	})

	t.Run("time_ptr_convert", func(t *testing.T) {
		testTime := time.Now()
		paramTimeA := TypePtrTimeA{
			TypeA:   TypeA{A: "abc_a"},
			TimePtr: &testTime,
		}
		got, err := preHandleArgs(paramTimeA)
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "abc_a", "TimePtr": testTime}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "abc_a", "TimePtr": testTime}})
		}

		paramTimeA = TypePtrTimeA{
			TypeA:   TypeA{A: "abc_a"},
			TimePtr: nil,
		}

		got, err = preHandleArgs(paramTimeA)
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "abc_a", "TimePtr": nil}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "abc_a", "TimePtr": nil}})
		}
	})

	t.Run("mixed_types", func(t *testing.T) {
		ab := TypeAB{}
		ab.A = "ab_a"
		ab.B = "ab_b"

		abc := TypeABC{}
		abc.A = "abc_a"
		abc.B = "abc_b" // override ab.A
		abc.C = "abc_c"

		m := map[string]any{
			"C": "map_c", // override abc.C
			"D": "map_d",
		}

		args := []any{
			ab,
			1,
			abc,
			[]byte{1, 2},
			m,
			[]int{1, 2, 3},
			testTime,
		}
		expected := []any{map[string]any{
			"A":  "abc_a",
			"B":  "abc_b",
			"C":  "map_c",
			"D":  "map_d",
			"p1": 1,
			"p2": []byte{1, 2},
			"p3": []int{1, 2, 3},
			"p4": testTime,
		}}
		got, err := preHandleArgs(args...)
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("mergeArgs() = %v, want %v", got, expected)
		}
	})
}
