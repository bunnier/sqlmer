package qm_namedsql

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
