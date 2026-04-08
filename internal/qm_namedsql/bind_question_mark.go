// Package qm_namedsql 提供了对 SQL 命名参数解析为 ? 占位符 SQL 的支持（用于：MySQL、SQLite 等使用 ? 占位符的驱动）。
package qm_namedsql

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bunnier/sqlmer"
)

// 是 SQL 命名参数解析后的结果。
type ParsedResult struct {
	Sql   string   // 处理后的 SQL 语句。
	Names []string // 原语句中用到的命名参数集合（没有去重逻辑，如果有同名参数，会有多项）。
}

// 分析 SQL 语句，提取用到的命名参数名称（按顺序），并将 @ 占位参数转换为驱动支持的 ? 形式。
// 结果会被缓存以提升重复调用性能；无命名参数的 SQL 不写入缓存，以防止动态拼接语句导致内存无限增长。
func ParseNamedSqlToQuestionMark(sqlText string) ParsedResult {
	// 如果缓存中有数据，直接返回。
	if cacheResult, ok := parsedSqlCache.load(sqlText); ok {
		return cacheResult
	}

	names := make([]string, 0, 10) // 存放 SQL 中所有的参数名称。

	fixedSqlTextBuilder := strings.Builder{}
	paramNameBuilder := strings.Builder{}

	inName := false   // 标示当前字符是否正处于参数名称之中。
	inString := false // 标示当前字符是否正处于字符串之中。

	lastIndex := utf8.RuneCountInString(sqlText) - 1 // SQL 语句的最后一个字符索引位置。

	for i, currentRune := range sqlText {
		switch {
		// 遇到字符串开始/结束符号 ' 需要切换字符串状态。
		case currentRune == '\'':
			inString = !inString
			fixedSqlTextBuilder.WriteRune(currentRune)

		// 如果当前状态正处于字符串中，字符无须特殊处理。
		case inString:
			fixedSqlTextBuilder.WriteRune(currentRune)

		// @ 符号标示参数名称部分开始。
		case currentRune == '@':
			// 连续 2 个 @ 可以用来转义，需要跳出参数名作用域。
			if inName && i > 0 && sqlText[i-1] == '@' {
				fixedSqlTextBuilder.WriteRune(currentRune)
				inName = false
				continue
			}
			inName = true
			paramNameBuilder.Reset()

		// 当前状态正处于参数名中，且当前字符合法，应作为参数名的一部分。
		case inName && isLegalParamNameCharter(currentRune):
			paramNameBuilder.WriteRune(currentRune)
			if lastIndex == i { // 如果是最后一个字符，直接写入并结束。
				fixedSqlTextBuilder.WriteString("?")
				names = append(names, paramNameBuilder.String())
			}

		// 当前状态正处于参数名中，且当前字符不是合法参数字符，标示着参数名范围结束。
		case inName && !isLegalParamNameCharter(currentRune):
			inName = false
			fixedSqlTextBuilder.WriteString("?")
			fixedSqlTextBuilder.WriteRune(currentRune)
			names = append(names, paramNameBuilder.String())

		// 非上述情况即为普通 SQL 字符部分，无须特殊处理。
		default:
			fixedSqlTextBuilder.WriteRune(currentRune)
		}
	}

	parsedResult := ParsedResult{fixedSqlTextBuilder.String(), names}
	// 无命名参数的 SQL 通常是动态拼接的语句，跳过缓存以防止内存无限增长。
	if len(names) > 0 {
		parsedSqlCache.store(sqlText, parsedResult)
	}
	return parsedResult
}

// 对使用 ? 占位符的驱动（如 MySQL、SQLite）做参数预处理。
// 若唯一参数为 map，则按 SQL 中的 @name 顺序绑定；否则按 @p1、@p2 形式从 args 切片取位置参数。
func BindQuestionMarkArgs(sqlText string, args ...any) (string, []any, error) {
	namedParsedResult := ParseNamedSqlToQuestionMark(sqlText)
	parsedNames := len(namedParsedResult.Names)
	argsCount := len(args)
	resultArgs := make([]any, 0, parsedNames)

	// map 按返回的 paramNames 顺序整理一个 slice 返回。
	if argsCount == 1 && reflect.ValueOf(args[0]).Kind() == reflect.Map {
		mapArgs := args[0].(map[string]any)
		for _, paramName := range namedParsedResult.Names { // 这里通过从 SQL 解析出来的参数名去取参数，刚好也解决了参数重复使用的问题。
			if value, ok := mapArgs[paramName]; ok {
				resultArgs = append(resultArgs, value)
			} else {
				return "", nil, fmt.Errorf("%w:\nlack of parameter\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql)
			}
		}
		namedParsedResult.Sql, resultArgs = extendInParams(namedParsedResult.Sql, resultArgs)
		return namedParsedResult.Sql, resultArgs, nil
	}

	// slice 语句中使用的顺序未必是递增的，且可能重复引用同一个索引，所以这里也需要整理顺序。
	for _, paramName := range namedParsedResult.Names {
		if paramName[0] != 'p' { // 要求数字参数的格式为 @p1....@pN。
			return "", nil, fmt.Errorf("%w: parsing parameter failed\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql)
		}

		index, err := strconv.Atoi(paramName[1:]) // 取出索引值。
		if err != nil {
			return "", nil, fmt.Errorf("%w: parsing parameter failed\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql)
		}

		index-- // 占位符从 1 开始，转成索引，需要 -1。

		if index < 0 { // 数字参数的数值从 1 开始，而不是 0，因此不会是负数。
			return "", nil, fmt.Errorf("%w: index must start at 1\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql)
		}

		if index >= argsCount { // 索引值超过参数数量了。
			return "", nil, fmt.Errorf("%w: parameter index out of args range\nsql = %s", sqlmer.ErrParseParamFailed, namedParsedResult.Sql)
		}

		resultArgs = append(resultArgs, args[index])
	}

	namedParsedResult.Sql, resultArgs = extendInParams(namedParsedResult.Sql, resultArgs)
	return namedParsedResult.Sql, resultArgs, nil
}

// extendInParams 处理 SQL IN 子句的参数展开，将切片类型的参数展开为多个问号占位符。
func extendInParams(sqlText string, params []any) (string, []any) {
	newParams := make([]any, 0, len(params))
	var newSqlBuilder strings.Builder

	paramIndex := 0
	inString := false
	for _, r := range sqlText {
		if r == '\'' {
			inString = !inString
			newSqlBuilder.WriteRune(r)
			continue
		}

		if inString {
			newSqlBuilder.WriteRune(r)
			continue
		}

		if r != '?' {
			newSqlBuilder.WriteRune(r)
			continue
		}

		if paramIndex >= len(params) { // 后面没参数了，无需判断了。
			newSqlBuilder.WriteRune(r)
			continue
		}

		param := params[paramIndex]
		paramValue := reflect.ValueOf(param)
		paramIndex++

		// 处理切片类型。
		// 排除 []byte，因为虽然 []byte 也是切片类型，但它是二进制数据，不应该被展开。
		if (paramValue.Kind() == reflect.Slice || paramValue.Kind() == reflect.Array) && !paramValue.Type().ConvertibleTo(reflect.TypeOf([]byte{})) {
			paramLen := paramValue.Len()
			if paramLen == 0 {
				// 空切片替换为 SQL 不可能条件。
				newSqlBuilder.WriteString("NULL")
				continue
			}

			// 生成占位符。
			placeholders := strings.Repeat(",?", paramLen)
			newSqlBuilder.WriteString(placeholders[1:])

			// 展开参数。
			for i := 0; i < paramLen; i++ {
				newParams = append(newParams, paramValue.Index(i).Interface())
			}
		} else {
			newSqlBuilder.WriteRune('?')
			newParams = append(newParams, param)
		}
	}

	return newSqlBuilder.String(), newParams
}
