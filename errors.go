package sqlmer

import "errors"

// ErrTran 是数据库执行事务操作遇到的错误。
var ErrTran = errors.New("dbTrans")

// ErrConnect 是 DbClient 获取数据库连接时遇到的错误。
var ErrConnect = errors.New("dbConnect")

// ErrGetEffectedRows 当数据库不支持获取影响行数时候，会返回改类型的错误。
var ErrGetEffectedRows = errors.New("dbClient: the db driver do not support getting effected rows")

// ErrSqlParamParse 解析 SQL 语句中的参数遇到错误时候，会返回该类型错误。
var ErrParseParamFailed = errors.New("dbClient: failed to parse named params")

// ErrExpectedSizeWrong 当执行语句时候，没有影响到预期行数，返回该类型错误。
var ErrExpectedSizeWrong = errors.New("dbClient: effected rows was wrong")

// ErrExecutingSql 当执行 SQL 语句执行时遇到错误，返回该类型错误。
var ErrExecutingSql = errors.New("dbClient: failed to execute sql")
