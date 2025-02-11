package sqlmer

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/bunnier/sqlmer/sqlen"
	"github.com/cmstar/go-conv"
)

var _ DbClient = (*AbstractDbClient)(nil)

// AbstractDbClient 是一个 DbClient 的抽象实现。
type AbstractDbClient struct {
	config *DbClientConfig      // 存储数据库连接配置。
	Db     *sqlen.DbEnhance     // 内部依赖的连接池。
	Exer   sqlen.EnhancedDbExer // 获取方法实际使用的执行对象。
}

// NewAbstractDbClient 用于获取一个 internalDbClient 对象。
func NewAbstractDbClient(config *DbClientConfig) (*AbstractDbClient, error) {
	// 控制连接超时的 context。
	ctx, cancelFunc := context.WithTimeout(context.Background(), config.connTimeout)
	defer cancelFunc()

	// db 可能已经由 option 传入了。
	if config.Db == nil {
		if config.Driver == "" || config.Dsn == "" {
			return nil, fmt.Errorf("%w: driver or dsn is empty", ErrConnect)
		}

		var err error
		config.Db, err = getDb(ctx, config.Driver, config.Dsn, config.withPingCheck)
		if err != nil {
			return nil, err
		}
	}

	dbEnhance := sqlen.NewDbEnhance(config.Db, config.getScanTypeFunc, config.unifyDataTypeFunc)
	return &AbstractDbClient{
		config: config,
		Db:     dbEnhance,
		Exer:   dbEnhance,
	}, nil
}

// 用于获取数据库连接池对象。
func getDb(ctx context.Context, driverName string, dsn string, withPingCheck bool) (*sql.DB, error) {
	db, err := sql.Open(driverName, dsn) // 获取连接池。
	if err != nil {
		return nil, err
	}

	if withPingCheck {
		if err = db.PingContext(ctx); err != nil { // Open 操作并不会实际建立链接，需要 ping 一下，确保连接可用。
			return nil, err
		}
	}

	return db, nil
}

// getExecTimeoutContext 用于获取数据库语句默认超时 context。
func (client *AbstractDbClient) getExecTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), client.config.execTimeout)
}

// Dsn 用于获取当前实例所使用的数据库连接字符串。
func (client *AbstractDbClient) Dsn() string {
	return client.config.Dsn
}

var dbConv = conv.Conv{
	Conf: conv.Config{
		FieldMatcherCreator: &conv.SimpleMatcherCreator{
			Conf: conv.SimpleMatcherConfig{
				Tag:            "conv",
				CamelSnakeCase: true,
			},
		},
	},
}

// mergeArgs 用于对 SQL 语句和参数进行预处理，支持多结构体参数的处理。
// 它会尝试从结构体参数中提取字段值，并将其合并到一个 map 中再交由 bindArgs 处理。
func mergeArgs(args ...any) ([]any, error) {
	needMerge := false // 是否需要合并参数。
	hasMap := false    // 是否有map类型的参数。

	// 1. 遍历所有参数，判断是否需要合并参数。
	for i := 0; i < len(args); i++ {
		argType := reflect.TypeOf(args[i])

		// 如果参数是指针，获取其指向的值。
		if argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		hasMap = hasMap || argType.Kind() == reflect.Map

		// 如果有 Map 类型的参数，且多余一个，也需要走合并逻辑。
		if hasMap && i > 0 {
			needMerge = true
			break
		}

		if argType.Kind() == reflect.Struct && !reflect.TypeOf(time.Time{}).ConvertibleTo(argType) {
			needMerge = true
			break
		}
	}

	// 2. 如果没有需要合并的参数，直接返回。
	if !needMerge {
		return args, nil
	}

	// 3. 开始参数处理逻辑。
	paramsMap := make(map[string]any)
	indexParamIdx := 0
	for _, arg := range args {
		argType := reflect.TypeOf(arg)

		switch {
		// 如果参数是 map 类型，直接合并到 paramsMap 中。
		case argType.Kind() == reflect.Map:
			argMap := arg.(map[string]any)
			for k, v := range argMap {
				paramsMap[k] = v
			}

		// 如果参数是结构体类型，转成 map 后合并。
		case argType.Kind() == reflect.Struct && !reflect.TypeOf(time.Time{}).ConvertibleTo(argType):
			argMap, err := dbConv.StructToMap(arg)
			if err != nil {
				return nil, err
			}

			// 将参数合并到 paramsMap 中。
			for k, v := range argMap {
				paramsMap[k] = v
			}

			// 其余类型，作为索引参数传递。
		default:
			indexParamIdx++
			paramsMap["p"+strconv.Itoa(indexParamIdx)] = arg
			continue
		}
	}

	// 使用合并后的参数map调用原有的bindArgs函数
	return []any{paramsMap}, nil
}
