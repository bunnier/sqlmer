package mssql

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	"errors"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
)

// 初始化测试配置。
var testConf testenv.TestConf = testenv.MustLoadTestConfig("../test_conf.yml")

// 用于获取一个 SqlServer 数据库的 DbClient 对象。
func getMsSqlDbClient() (sqlmer.DbClient, error) {
	return NewMsSqlDbClient(
		testConf.SqlServer,
		sqlmer.WithConnTimeout(time.Second*15),
		sqlmer.WithExecTimeout(time.Second*15),
	)
}

func Test_NewMsSqlDbClient(t *testing.T) {
	dbClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}

	if dbClient.Dsn() != testConf.SqlServer {
		t.Errorf("mssqlDbClient.Dsn() connString = %v, wantConnString %v", dbClient.Dsn(), testConf.SqlServer)
	}

	if dbClient, err = NewMsSqlDbClient("test",
		sqlmer.WithConnTimeout(time.Second*15),
		sqlmer.WithExecTimeout(time.Second*15),
		sqlmer.WithPingCheck(true)); err == nil {
		t.Errorf("mssqlDbClient.NewMsSqlDbClient() err = nil, want has a err")
	}
}

func Test_bindMsSqlArgs(t *testing.T) {
	testCases := []struct {
		name      string
		oriSql    string
		args      []interface{}
		wantSql   string
		wantParam []interface{}
		wantErr   error
	}{
		{
			"map",
			"SELECT * FROM go_TypeTest WHERE Id=@id",
			[]interface{}{
				map[string]interface{}{
					"id": 1,
				},
			},
			"SELECT * FROM go_TypeTest WHERE Id=@id",
			[]interface{}{sql.Named("id", 1)},
			nil,
		},
		{
			"index",
			"SELECT * FROM go_TypeTest WHERE Id=@p1 OR Id=@p2",
			[]interface{}{1, 2},
			"SELECT * FROM go_TypeTest WHERE Id=@p1 OR Id=@p2",
			[]interface{}{1, 2},
			nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fixedSql, args, err := bindMsSqlArgs(tt.oriSql, tt.args...)

			if tt.wantErr != nil {
				if !errors.As(err, &tt.wantErr) {
					t.Errorf("mssqlDbClient.bindMsSqlArgs() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else {
				if err != nil {
					t.Error(err)
					return
				}
				if fixedSql != tt.wantSql {
					t.Errorf("mssqlDbClient.bindMsSqlArgs() sql = %v, wantSql %v", fixedSql, tt.wantSql)
				}

				if !reflect.DeepEqual(args, tt.wantParam) {
					t.Errorf("mssqlDbClient.bindMsSqlArgs() args = %v, wantParam %v", args, tt.wantParam)
				}
			}
		})
	}
}

func Test_internalDbClient_Scalar(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []interface{}
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			"mssql",
			mssqlClient,
			args{
				"SELECT Id FROM go_TypeTest WHERE Id=@p1",
				[]interface{}{1},
			},
			int64(1),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := tt.client.Scalar(tt.args.sqlText, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("internalDbClient.Scalar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("internalDbClient.Scalar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_internalDbClient_Execute(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []interface{}
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		wantErr bool
	}{
		{
			"mssql",
			mssqlClient,
			args{
				`INSERT INTO go_TypeTest (TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest, BinaryTest)
				VALUES (5, 5, 5, 5, N'行5', 'Row5', N'行5', 'Row5', '2021-07-05 15:38:39.583', '2021-07-05 15:38:50.4257813', '2021-07-05', '12:05:01.345', 5.123, 5.12345, 5.45678999, 1);`,
				[]interface{}{},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effectRow, err := tt.client.Execute(tt.args.sqlText, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("internalDbClient.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if effectRow != int64(1) {
				if (err != nil) != tt.wantErr {
					t.Errorf("internalDbClient.Execute() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			err = tt.client.SizedExecute(1, tt.args.sqlText, tt.args.args...)
			if err != nil {
				t.Errorf("internalDbClient.SizedExecute() error = %v, wantErr %v", err, tt.wantErr)
			}

			err = tt.client.SizedExecute(2, tt.args.sqlText, tt.args.args...)
			if !errors.Is(err, sqlmer.ErrExpectedSizeWrong) {
				t.Errorf("internalDbClient.SizedExecute() error = %v, wantErr DbSqlError", err)
			}
		})
	}
}

func Test_internalDbClient_Exists(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []interface{}
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		want    bool
		wantErr bool
	}{
		{
			"mssql_exist",
			mssqlClient,
			args{
				"SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest FROM go_TypeTest WHERE Id=1",
				[]interface{}{},
			},
			true,
			false,
		},
		{
			"mssql_notexist",
			mssqlClient,
			args{
				"SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest FROM go_TypeTest WHERE Id=10000",
				[]interface{}{},
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.Exists(tt.args.sqlText, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("internalDbClient.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("internalDbClient.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_internalDbClient_Get(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []interface{}
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			"mssql_nullable_null",
			mssqlClient,
			args{
				`SELECT TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest, BinaryTest,
				NullableTinyIntTest, NullableSmallIntTest, NullableIntTest, NullableBitTest, NullableNvarcharTest, NullableVarcharTest, NullableNcharTest, NullableCharTest, NullableDateTimeTest, NullableDateTime2Test, NullableDateTest, NullableTimeTest, NullableMoneyTest, NullableFloatTest, NullableDecimalTest, NullableBinaryTest
				FROM go_TypeTest WHERE Id=1`,
				[]interface{}{},
			},
			map[string]interface{}{
				"TinyIntTest":           int64(1),
				"SmallIntTest":          int64(1),
				"IntTest":               int64(1),
				"BitTest":               true,
				"NvarcharTest":          "行1",
				"VarcharTest":           "Row1",
				"NcharTest":             "行1",
				"CharTest":              "Row1",
				"DateTimeTest":          time.Date(2021, 7, 1, 15, 38, 39, 583000000, time.UTC),
				"DateTime2Test":         time.Date(2021, 7, 1, 15, 38, 50, 425781300, time.UTC),
				"DateTest":              time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
				"TimeTest":              time.Date(1, 1, 1, 12, 1, 1, 345000000, time.UTC),
				"MoneyTest":             "1.1230",
				"FloatTest":             float64(1.12345),
				"DecimalTest":           "1.4567899900",
				"BinaryTest":            []byte{1},
				"NullableTinyIntTest":   nil,
				"NullableSmallIntTest":  nil,
				"NullableIntTest":       nil,
				"NullableBitTest":       nil,
				"NullableNvarcharTest":  nil,
				"NullableVarcharTest":   nil,
				"NullableNcharTest":     nil,
				"NullableCharTest":      nil,
				"NullableDateTimeTest":  nil,
				"NullableDateTime2Test": nil,
				"NullableDateTest":      nil,
				"NullableTimeTest":      nil,
				"NullableMoneyTest":     nil,
				"NullableFloatTest":     nil,
				"NullableDecimalTest":   nil,
				"NullableBinaryTest":    nil,
			},
			false,
		},
		{
			"mssql_nullable_hasValue",
			mssqlClient,
			args{
				`SELECT TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest, BinaryTest,
				NullableTinyIntTest, NullableSmallIntTest, NullableIntTest, NullableBitTest, NullableNvarcharTest, NullableVarcharTest, NullableNcharTest, NullableCharTest, NullableDateTimeTest, NullableDateTime2Test, NullableDateTest, NullableTimeTest, NullableMoneyTest, NullableFloatTest, NullableDecimalTest, NullableBinaryTest
				FROM go_TypeTest WHERE Id=3`,
				[]interface{}{},
			},
			map[string]interface{}{
				"TinyIntTest":           int64(3),
				"SmallIntTest":          int64(3),
				"IntTest":               int64(3),
				"BitTest":               true,
				"NvarcharTest":          "行3",
				"VarcharTest":           "Row3",
				"NcharTest":             "行3",
				"CharTest":              "Row3",
				"DateTimeTest":          time.Date(2021, 7, 3, 15, 38, 39, 583000000, time.UTC),
				"DateTime2Test":         time.Date(2021, 7, 3, 15, 38, 50, 425781300, time.UTC),
				"DateTest":              time.Date(2021, 7, 3, 0, 0, 0, 0, time.UTC),
				"TimeTest":              time.Date(1, 1, 1, 12, 3, 1, 345000000, time.UTC),
				"MoneyTest":             "3.1230",
				"FloatTest":             float64(3.12345),
				"DecimalTest":           "3.4567899900",
				"BinaryTest":            []byte{1},
				"NullableTinyIntTest":   int64(3),
				"NullableSmallIntTest":  int64(3),
				"NullableIntTest":       int64(3),
				"NullableBitTest":       true,
				"NullableNvarcharTest":  "行3",
				"NullableVarcharTest":   "Row3",
				"NullableNcharTest":     "行3",
				"NullableCharTest":      "Row3",
				"NullableDateTimeTest":  time.Date(2021, 7, 3, 15, 38, 39, 583000000, time.UTC),
				"NullableDateTime2Test": time.Date(2021, 7, 3, 15, 38, 50, 425781300, time.UTC),
				"NullableDateTest":      time.Date(2021, 7, 3, 0, 0, 0, 0, time.UTC),
				"NullableTimeTest":      time.Date(1, 1, 1, 12, 3, 1, 345000000, time.UTC),
				"NullableMoneyTest":     "3.1230",
				"NullableFloatTest":     float64(3.12345),
				"NullableDecimalTest":   "3.4567899900",
				"NullableBinaryTest":    []byte{1},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.Get(tt.args.sqlText, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("internalDbClient.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for k, v := range got {
				wantV := tt.want[k]
				if !reflect.DeepEqual(v, wantV) {
					if wantFloat, ok := wantV.(float64); ok {
						if wantFloat-v.(float64) < 0.00001 {
							continue
						}
					} else if wantString, ok := wantV.(string); ok {
						if wantString == strings.Trim(v.(string), " ") {
							continue
						}
					}
					t.Errorf("fieldname = %s, internalDbClient.Get() = %v, want %v", k, v, wantV)
				}
			}
		})
	}
}

func Test_internalDbClient_Rows(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []interface{}
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		want    []map[string]interface{}
		wantErr bool
	}{
		{
			"mssql",
			mssqlClient,
			args{
				"SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest FROM go_TypeTest WHERE Id IN (1,2)",
				[]interface{}{},
			},
			[]map[string]interface{}{
				{
					"NvarcharTest":  "行1",
					"VarcharTest":   "Row1",
					"DateTimeTest":  time.Date(2021, 7, 1, 15, 38, 39, 583000000, time.UTC),
					"DateTime2Test": time.Date(2021, 7, 1, 15, 38, 50, 425781300, time.UTC),
					"DateTest":      time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
					"TimeTest":      time.Date(1, 1, 1, 12, 1, 1, 345000000, time.UTC),
					"DecimalTest":   "1.4567899900",
				},
				{
					"NvarcharTest":  "行2",
					"VarcharTest":   "Row2",
					"DateTimeTest":  time.Date(2021, 7, 2, 15, 38, 39, 583000000, time.UTC),
					"DateTime2Test": time.Date(2021, 7, 2, 15, 38, 50, 425781300, time.UTC),
					"DateTest":      time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC),
					"TimeTest":      time.Date(1, 1, 1, 12, 2, 1, 345000000, time.UTC),
					"DecimalTest":   "2.4567899900",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := tt.client.Rows(tt.args.sqlText, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer rows.Close()
			index := 0
			for rows.Next() {
				got := make(map[string]interface{})
				err := rows.MapScan(got)
				if (err != nil) != tt.wantErr {
					t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want[index]) {
					t.Errorf("internalDbClient.Get() = %v, want %v", got, tt.want)
					return
				}
				index++
			}
		})
	}
}
