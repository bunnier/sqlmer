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

// GetExecTimeout 用于获取当前 DbClient 实例的执行超时时间。
func (client *AbstractDbClient) GetExecTimeout() time.Duration {
	return client.config.execTimeout
}

// GetConnTimeout 用于获取当前 DbClient 实例的连接超时时间。
func (client *AbstractDbClient) GetConnTimeout() time.Duration {
	return client.config.connTimeout
}

// Dsn 用于获取当前实例所使用的数据库连接字符串。
func (client *AbstractDbClient) Dsn() string {
	return client.config.Dsn
}

// getExecTimeoutContext 用于获取数据库语句默认超时 context。
func (client *AbstractDbClient) getExecTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), client.GetExecTimeout())
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

// preHandleArgs 用于对 SQL 语句和参数进行预处理。
func preHandleArgs(args ...any) ([]any, error) {
	if !needMergeArgs(args...) {
		return args, nil
	}

	mergedArgs := make(map[string]any, len(args))
	indexParamCount := 0 // 专门用于记录索引参数的计数
	for _, arg := range args {
		if err := handleSingleArg(mergedArgs, arg, &indexParamCount); err != nil {
			return nil, err
		}
	}

	return []any{mergedArgs}, nil
}

// 检查参数类型是否需要合并处理。
func needMergeArgs(args ...any) bool {
	hasMap := false // 判断是否有 map 参数。
	for i, arg := range args {
		argType := reflect.TypeOf(arg)
		if argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		// 如果是 map 类型参数。
		if argType.Kind() == reflect.Map {
			if i > 0 { // 如果不是第一个参数就是 map，需要合并。
				return true
			}
			hasMap = true
			continue
		}

		// 如果是非 time.Time 的结构体，需要合并。
		if argType.Kind() == reflect.Struct && !reflect.TypeOf(time.Time{}).ConvertibleTo(argType) {
			return true
		}

		// 如果已经有 map 且遇到其他类型参数，需要合并。
		if hasMap {
			return true
		}
	}

	return false
}

// 处理单个参数。
func handleSingleArg(paramsMap map[string]any, arg any, indexParamCount *int) error {
	argType := reflect.TypeOf(arg)

	// 处理指针类型。
	if argType.Kind() == reflect.Ptr {
		argType = argType.Elem()
		arg = reflect.ValueOf(arg).Elem().Interface()
	}

	switch {
	// 处理 map 类型参数。
	case argType.Kind() == reflect.Map:
		argMap := arg.(map[string]any)
		for k, v := range argMap {
			paramsMap[k] = v
		}

	// 处理结构体类型参数。
	case argType.Kind() == reflect.Struct && !reflect.TypeOf(time.Time{}).ConvertibleTo(argType):
		argMap, err := dbConv.StructToMap(arg)
		if err != nil {
			return err
		}

		for k, v := range argMap {
			paramsMap[k] = v
		}

	// 索引参数（索引的值 1 开始，所以就是已发现的参数的个数）。
	default:
		*indexParamCount++
		paramsMap["p"+strconv.Itoa(*indexParamCount)] = arg
	}

	return nil
}
