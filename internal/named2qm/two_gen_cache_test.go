package named2qm

import (
	"testing"
)

func Test_twoGenCache_hit_in_hot(t *testing.T) {
	c := newTwoGenCache(10)
	v := ParsedResult{Sql: "SELECT 1", Names: []string{"id"}}
	c.store("key1", v)

	got, ok := c.load("key1")
	if !ok {
		t.Fatal("expected cache hit in hot, got miss")
	}
	if got.Sql != v.Sql {
		t.Errorf("expected sql=%s, got=%s", v.Sql, got.Sql)
	}
}

func Test_twoGenCache_hit_in_cold(t *testing.T) {
	c := newTwoGenCache(2)
	v := ParsedResult{Sql: "SELECT cold", Names: []string{"x"}}
	c.store("cold_key", v)

	c.store("fill1", ParsedResult{Sql: "s1", Names: []string{"a"}})
	c.store("fill2", ParsedResult{Sql: "s2", Names: []string{"b"}})

	got, ok := c.load("cold_key")
	if !ok {
		t.Fatal("expected cache hit in cold, got miss")
	}
	if got.Sql != v.Sql {
		t.Errorf("expected sql=%s, got=%s", v.Sql, got.Sql)
	}

	c.mu.RLock()
	_, promotedToHot := c.hot["cold_key"]
	c.mu.RUnlock()
	if !promotedToHot {
		t.Error("expected cold_key to be promoted to hot after cold hit")
	}
}

func Test_twoGenCache_eviction_on_capacity(t *testing.T) {
	c := newTwoGenCache(2)

	c.store("k1", ParsedResult{Sql: "s1", Names: []string{"a"}})
	c.store("k2", ParsedResult{Sql: "s2", Names: []string{"b"}})

	c.store("k3", ParsedResult{Sql: "s3", Names: []string{"c"}})

	c.mu.RLock()
	hotLen := len(c.hot)
	coldLen := len(c.cold)
	c.mu.RUnlock()

	if hotLen != 1 {
		t.Errorf("expected hot len=1 after eviction, got=%d", hotLen)
	}
	if coldLen != 2 {
		t.Errorf("expected cold len=2 after eviction, got=%d", coldLen)
	}
}

func Test_twoGenCache_miss(t *testing.T) {
	c := newTwoGenCache(10)
	_, ok := c.load("nonexistent")
	if ok {
		t.Error("expected cache miss, got hit")
	}
}

func Test_NewQuestionMarkBinder_invalid_capacity(t *testing.T) {
	_, err := NewQuestionMarkSqlBinder(0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_QuestionMarkBinder_has_independent_cache(t *testing.T) {
	binder1, err := NewQuestionMarkSqlBinder(1)
	if err != nil {
		t.Fatal(err)
	}

	binder2, err := NewQuestionMarkSqlBinder(1)
	if err != nil {
		t.Fatal(err)
	}

	sqlText := "SELECT * FROM t WHERE id=@id"
	result1 := binder1.ParseNamedSqlToQuestionMark(sqlText)
	if result1.Sql != "SELECT * FROM t WHERE id=?" {
		t.Errorf("expected parsed sql, got %s", result1.Sql)
	}

	_, ok := binder1.cache.load(sqlText)
	if !ok {
		t.Fatal("expected binder1 cache hit, got miss")
	}

	_, ok = binder2.cache.load(sqlText)
	if ok {
		t.Fatal("expected binder2 cache miss, got hit")
	}
}
