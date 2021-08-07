package sqlmer

import (
	"context"
	"database/sql"
	"sync"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

const (
	SqlServeDriver = "sqlserver" // SqlServer 驱动名称。
	MySqlDriver    = "mysql"     // MsSql 驱动名称。
)

// dbCacheMap 连接池缓存结构。
type dbCacheMap struct {
	init sync.Once          // 用来保证 Map 只会初始化一次。
	m    sync.RWMutex       // 用于确保同一个连接的创建过程不并发（TODO: 后续要考虑锁的粒度，理论上对连接串加锁即可）。
	dbs  map[string]*sql.DB // 存放已经初始化的数据库对象，key 为连接字符串，value 为 sql.DB 对象。
}

var dbCache dbCacheMap // 用于缓存已经初始化的数据库对象。

// 用于获取数据库连接池对象。
func getDb(ctx context.Context, driverName string, connectionString string) (*sql.DB, error) {
	dbCache.init.Do(func() {
		dbCache.dbs = make(map[string]*sql.DB) // 确保只延迟初始化一次 CacheMap。
	})

	dbCache.m.RLock()
	db, exist := dbCache.dbs[connectionString]
	dbCache.m.RUnlock()
	if exist {
		return db, nil
	}

	// 初始化可能失败，由于sync.Once在失败后无法再次运行，所以这里不采用 sync.Once 直接用锁控制。
	dbCache.m.Lock()
	defer dbCache.m.Unlock()
	db, exist = dbCache.dbs[connectionString] // 锁内双校验。
	if exist {
		return db, nil
	}

	db, err := sql.Open(driverName, connectionString) // 获取连接池。
	if err != nil {
		return nil, err
	}
	if err = db.PingContext(ctx); err != nil { // 确保连接可用。
		return nil, err
	}

	dbCache.dbs[connectionString] = db // 放入连接池缓存中。
	return db, nil
}
