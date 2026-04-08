package qm_namedsql

import (
	"sync"
)

// 是双代缓存每代的最大条目数，内存上限为 2 倍该值。
const _DefaultCacheCapacity = 1024

//	一个基于双代淘汰策略的有界缓存。
//
// 当 hot 代条目数达到容量上限时，将 hot 降级为 cold，并创建新的 hot 代。
// 最大内存占用为 2 * capacity 条目。
type twoGenCache struct {
	mu       sync.RWMutex
	hot      map[string]ParsedResult // 当前活跃代。
	cold     map[string]ParsedResult // 上一代（待淘汰）。
	capacity int
}

func newTwoGenCache(capacity int) *twoGenCache {
	return &twoGenCache{
		hot:      make(map[string]ParsedResult, capacity),
		cold:     make(map[string]ParsedResult),
		capacity: capacity,
	}
}

// load 从缓存中读取。hot 未命中时查 cold，cold 命中则将条目提升至 hot。
func (c *twoGenCache) load(key string) (ParsedResult, bool) {
	c.mu.RLock()
	if v, ok := c.hot[key]; ok {
		c.mu.RUnlock()
		return v, true
	}
	v, ok := c.cold[key]
	c.mu.RUnlock()

	if ok {
		// 将 cold 中命中的条目提升到 hot，使其不因下次轮转而丢失。
		c.store(key, v)
	}
	return v, ok
}

// store 将条目写入 hot 代。hot 满时触发轮转：cold = hot，hot = 新空 map。
func (c *twoGenCache) store(key string, value ParsedResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.hot) >= c.capacity {
		c.cold = c.hot
		c.hot = make(map[string]ParsedResult, c.capacity)
	}
	c.hot[key] = value
}

// parsedSqlCache 是 ParseNamedSql 使用的包级共享缓存实例。
var parsedSqlCache = newTwoGenCache(_DefaultCacheCapacity)

// 用于初始化合法字符集合 map，用于快速筛选合法字符。
var onceInitParamNameMap = sync.Once{}

// 定义参数名允许的字符。
var legalParamNameCharactersMap map[rune]struct{}

// isLegalParamNameCharter 用于快速判断某个字符是否是占位符参数合法字符。
func isLegalParamNameCharter(r rune) bool {
	onceInitParamNameMap.Do(func() {
		const legalParamNameCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
		legalParamNameCharactersMap = make(map[rune]struct{}, len(legalParamNameCharacters))
		for _, r := range legalParamNameCharacters {
			legalParamNameCharactersMap[r] = struct{}{}
		}
	})

	_, ok := legalParamNameCharactersMap[r]
	return ok
}
