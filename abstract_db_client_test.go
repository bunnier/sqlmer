package sqlmer_test

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/mssql"
	"github.com/bunnier/sqlmer/mysql"
	"github.com/pkg/errors"
)

var testConf testenv.TestConf = testenv.MustLoadTestConfig("test_conf.yml")

// 用于获取一个 SqlServer 测试库的 DbClient 对象。
func getMsSqlDbClient() (sqlmer.DbClient, error) {
	return mssql.NewMsSqlDbClient(
		testConf.SqlServer,
		sqlmer.WithConnTimeout(time.Second*15),
		sqlmer.WithExecTimeout(time.Second*15),
	)
}

// 用于获取一个 MySql 测试库的 DbClient 对象。
func getMySqlDbClient() (sqlmer.DbClient, error) {
	return mysql.NewMySqlDbClient(
		testConf.MySql,
		sqlmer.WithConnTimeout(time.Second*15),
		sqlmer.WithExecTimeout(time.Second*15),
	)
}

func Test_internalDbClient_Scalar(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	mysqlClient, err := getMySqlDbClient()
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
		{
			"mysql",
			mysqlClient,
			args{
				"SELECT Id FROM go_TypeTest WHERE id=@p1",
				[]interface{}{1},
			},
			int64(1),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.Scalar(tt.args.sqlText, tt.args.args...)
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
	now := time.Now()
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	mysqlClient, err := getMySqlDbClient()
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
				"INSERT INTO go_TypeTest (NvarcharTest, VarcharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, DecimalTest) VALUES (N'行5', 'Row5', @p1, @p1, @p1, @p1, 1.45678999);",
				[]interface{}{now},
			},
			false,
		},
		{
			"mysql",
			mysqlClient,
			args{
				"INSERT INTO go_TypeTest(VarcharTest, dateTest, dateTimeTest, timestampTest, decimalTest) VALUES (N'行5', @p1, @p1, @p1, 1.45678999)",
				[]interface{}{now},
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
			if !errors.Is(err, sqlmer.ErrSql) {
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
	mysqlClient, err := getMySqlDbClient()
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
		{
			"mysql_exist",
			mysqlClient,
			args{
				"SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id=1",
				[]interface{}{},
			},
			true,
			false,
		},
		{
			"mysql_notexist",
			mysqlClient,
			args{
				"SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id=10000",
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
	mysqlClient, err := getMySqlDbClient()
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
			"mssql",
			mssqlClient,
			args{
				"SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest FROM go_TypeTest WHERE Id=1",
				[]interface{}{},
			},
			map[string]interface{}{
				"NvarcharTest":  "行1",
				"VarcharTest":   "Row1",
				"DateTimeTest":  time.Date(2021, 7, 1, 15, 38, 39, 583000000, time.UTC),
				"DateTime2Test": time.Date(2021, 7, 1, 15, 38, 50, 425781300, time.UTC),
				"DateTest":      time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
				"TimeTest":      time.Date(1, 1, 1, 12, 1, 1, 345000000, time.UTC),
				"DecimalTest":   sql.NullString{String: "1.4567899900", Valid: true},
			},
			false,
		},
		{
			"mysql",
			mysqlClient,
			args{
				"SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id=1",
				[]interface{}{},
			},
			map[string]interface{}{
				"varcharTest":   "行1",
				"dateTest":      time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
				"dateTimeTest":  time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
				"timestampTest": time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
				"decimalTest":   sql.NullString{String: "1.4567899900", Valid: true},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("internalDbClient.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_internalDbClient_SliceGet(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	mysqlClient, err := getMySqlDbClient()
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
					"DecimalTest":   sql.NullString{String: "1.4567899900", Valid: true},
				},
				{
					"NvarcharTest":  "行2",
					"VarcharTest":   "Row2",
					"DateTimeTest":  time.Date(2021, 7, 2, 15, 38, 39, 583000000, time.UTC),
					"DateTime2Test": time.Date(2021, 7, 2, 15, 38, 50, 425781300, time.UTC),
					"DateTest":      time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC),
					"TimeTest":      time.Date(1, 1, 1, 12, 1, 2, 345000000, time.UTC),
					"DecimalTest":   sql.NullString{String: "2.4567899900", Valid: true},
				},
			},
			false,
		},
		{
			"mysql",
			mysqlClient,
			args{
				"SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (1,2)",
				[]interface{}{},
			},
			[]map[string]interface{}{
				{
					"varcharTest":   "行1",
					"dateTest":      time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
					"dateTimeTest":  time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
					"timestampTest": time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
					"decimalTest":   sql.NullString{String: "1.4567899900", Valid: true},
				},
				{
					"varcharTest":   "行2",
					"dateTest":      time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC),
					"dateTimeTest":  time.Date(2021, 7, 2, 15, 38, 50, 0, time.UTC),
					"timestampTest": time.Date(2021, 7, 2, 15, 38, 50, 0, time.UTC),
					"decimalTest":   sql.NullString{String: "2.4567899900", Valid: true},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.SliceGet(tt.args.sqlText, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("internalDbClient.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("internalDbClient.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_internalDbClient_Rows(t *testing.T) {
	mssqlClient, err := getMsSqlDbClient()
	if err != nil {
		t.Fatal(err)
	}
	mysqlClient, err := getMySqlDbClient()
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
					"DecimalTest":   sql.NullString{String: "1.4567899900", Valid: true},
				},
				{
					"NvarcharTest":  "行2",
					"VarcharTest":   "Row2",
					"DateTimeTest":  time.Date(2021, 7, 2, 15, 38, 39, 583000000, time.UTC),
					"DateTime2Test": time.Date(2021, 7, 2, 15, 38, 50, 425781300, time.UTC),
					"DateTest":      time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC),
					"TimeTest":      time.Date(1, 1, 1, 12, 1, 2, 345000000, time.UTC),
					"DecimalTest":   sql.NullString{String: "2.4567899900", Valid: true},
				},
			},
			false,
		},
		{
			"mysql",
			mysqlClient,
			args{
				"SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (1,2)",
				[]interface{}{},
			},
			[]map[string]interface{}{
				{
					"varcharTest":   "行1",
					"dateTest":      time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
					"dateTimeTest":  time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
					"timestampTest": time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
					"decimalTest":   sql.NullString{String: "1.4567899900", Valid: true},
				},
				{
					"varcharTest":   "行2",
					"dateTest":      time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC),
					"dateTimeTest":  time.Date(2021, 7, 2, 15, 38, 50, 0, time.UTC),
					"timestampTest": time.Date(2021, 7, 2, 15, 38, 50, 0, time.UTC),
					"decimalTest":   sql.NullString{String: "2.4567899900", Valid: true},
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
