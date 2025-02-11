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

func TestArgsMerge(t *testing.T) {
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

	t.Run("空参数列表", func(t *testing.T) {
		got, err := mergeArgs()
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if len(got) != 0 {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{})
		}
	})

	t.Run("基础类型切片参数", func(t *testing.T) {
		got, err := mergeArgs([]int{1, 2, 3})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{[]int{1, 2, 3}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{[]int{1, 2, 3}})
		}
	})

	t.Run("单个map参数", func(t *testing.T) {
		got, err := mergeArgs(map[string]any{"T": "t"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"T": "t"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"T": "t"}})
		}
	})

	t.Run("多个map参数", func(t *testing.T) {
		got, err := mergeArgs(map[string]any{"T": "t"}, map[string]any{"T": "t2"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"T": "t2"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"T": "t2"}})
		}
	})

	t.Run("单个结构体参数", func(t *testing.T) {
		got, err := mergeArgs(TypeA{A: "a"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "a"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "a"}})
		}
	})

	t.Run("单个指针结构体参数", func(t *testing.T) {
		got, err := mergeArgs(&TypeA{A: "a"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "a"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "a"}})
		}
	})

	t.Run("多个结构体参数", func(t *testing.T) {
		got, err := mergeArgs(TypeA{A: "a"}, TypeB{B: "b"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "a", "B": "b"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "a", "B": "b"}})
		}
	})

	t.Run("多个结构体指针参数", func(t *testing.T) {
		got, err := mergeArgs(&TypeA{A: "a"}, &TypeB{B: "b"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "a", "B": "b"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "a", "B": "b"}})
		}
	})

	t.Run("嵌套结构体参数", func(t *testing.T) {
		got, err := mergeArgs(TypeAB{TypeA: TypeA{A: "ab_a"}, TypeB: TypeB{B: "ab_b"}})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "ab_a", "B": "ab_b"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "ab_a", "B": "ab_b"}})
		}
	})

	t.Run("多层嵌套结构体参数", func(t *testing.T) {
		got, err := mergeArgs(TypeABC{TypeAB: TypeAB{TypeA: TypeA{A: "abc_a"}, TypeB: TypeB{B: "abc_b"}}, C: "abc_c"})
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, []any{map[string]any{"A": "abc_a", "B": "abc_b", "C": "abc_c"}}) {
			t.Errorf("mergeArgs() = %v, want %v", got, []any{map[string]any{"A": "abc_a", "B": "abc_b", "C": "abc_c"}})
		}
	})

	t.Run("混合类型参数", func(t *testing.T) {
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
		got, err := mergeArgs(args...)
		if err != nil {
			t.Errorf("mergeArgs() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("mergeArgs() = %v, want %v", got, expected)
		}
	})
}
