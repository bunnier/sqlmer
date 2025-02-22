package sqlmer

import (
	"reflect"
	"testing"
	"time"
)

type StrA struct {
	A string
}

type StrB struct {
	B string
}

type StrAB struct {
	A string
	B string
}

type StrABC struct {
	StrAB
	C string
}

func Test_preHandleArgs(t *testing.T) {
	type TypeA struct {
		A string
	}

	type TypeB struct {
		B string
	}

	type TypeAB struct {
		TypeA
		TypeB
	}

	type TypeABC struct {
		TypeAB
		C string
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

	t.Run("mixed_types", func(t *testing.T) {
		args := []any{
			TypeAB{TypeA: TypeA{A: "ab_a"}, TypeB: TypeB{B: "ab_b"}},
			1,
			TypeABC{TypeAB: TypeAB{TypeA: TypeA{A: "abc_a"}, TypeB: TypeB{B: "abc_b"}}, C: "abc_c"},
			[]byte{1, 2},
			map[string]any{"C": "map_c", "D": "map_d"},
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
