package mysql_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/mysql"
)

func init() {
	testenv.TryInitConfig("..")
}

func Test_NewMySqlDbClient(t *testing.T) {
	dbClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(dbClient.Dsn(), testenv.TestConf.Mysql) {
		t.Errorf("mysqlDbClient.Dsn() connString = %v, want contains  %v", dbClient.Dsn(), testenv.TestConf.Mysql)
	}

	_, err = mysql.NewMySqlDbClient("test",
		sqlmer.WithConnTimeout(testenv.DefaultTimeout),
		sqlmer.WithExecTimeout(testenv.DefaultTimeout),
		sqlmer.WithPingCheck(true))
	if err == nil {
		t.Errorf("mysqlDbClient.NewMsSqlDbClient() err = nil, want has a err")
	}
}

func Test_internalDbClient_Scalar(t *testing.T) {
	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []any
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		want    any
		wantErr bool
	}{
		{
			"mysql",
			mysqlClient,
			args{
				"SELECT Id FROM go_TypeTest WHERE id=@p1",
				[]any{1},
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
	now := time.Now()
	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []any
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		wantErr bool
	}{
		{
			"mysql",
			mysqlClient,
			args{
				`INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
				VALUES (5, 5, 5, 5, 5, N'行5', '行5char', '行5text','2021-07-05','2021-07-05 15:38:50.425','2021-07-05 15:38:50.425', 5.456, 5.15678, 5.45678999, 1);`,
				[]any{now},
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
	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []any
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		want    bool
		wantErr bool
	}{
		{
			"mysql_exist",
			mysqlClient,
			args{
				"SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id=1",
				[]any{},
			},
			true,
			false,
		},
		{
			"mysql_notexist",
			mysqlClient,
			args{
				"SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id=10000",
				[]any{},
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
	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []any
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			"mysql_nullable_null",
			mysqlClient,
			args{
				`SELECT intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
				nullIntTest, nullTinyintTest, nullSmallIntTest, nullBigIntTest, nullUnsignedTest, nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest,
				null colNull 
				FROM go_TypeTest WHERE id=1`,
				[]any{},
			},
			map[string]any{
				"intTest":           int64(1),
				"tinyintTest":       int64(1),
				"smallIntTest":      int64(1),
				"bigIntTest":        int64(1),
				"unsignedTest":      int64(1),
				"varcharTest":       "行1",
				"charTest":          "行1char",
				"charTextTest":      "行1text",
				"dateTest":          time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
				"dateTimeTest":      time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
				"timestampTest":     time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
				"floatTest":         float64(1.456),
				"doubleTest":        float64(1.15678),
				"decimalTest":       "1.4567899900",
				"bitTest":           []byte{1},
				"nullIntTest":       nil,
				"nullTinyintTest":   nil,
				"nullSmallIntTest":  nil,
				"nullBigIntTest":    nil,
				"nullUnsignedTest":  nil,
				"nullVarcharTest":   nil,
				"nullCharTest":      nil,
				"nullTextTest":      nil,
				"nullDateTest":      nil,
				"nullDateTimeTest":  nil,
				"nullTimestampTest": nil,
				"nullFloatTest":     nil,
				"nullDoubleTest":    nil,
				"nullDecimalTest":   nil,
				"nullBitTest":       nil,
				"colNull":           nil,
			},
			false,
		},
		{
			"mysql_nullable_hasValue",
			mysqlClient,
			args{
				`SELECT intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
				nullIntTest, nullTinyintTest, nullSmallIntTest, nullBigIntTest, nullUnsignedTest, nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest,
				null colNull 
				FROM go_TypeTest WHERE id=3`,
				[]any{},
			},
			map[string]any{
				"intTest":           int64(3),
				"tinyintTest":       int64(3),
				"smallIntTest":      int64(3),
				"bigIntTest":        int64(3),
				"unsignedTest":      int64(3),
				"varcharTest":       "行3",
				"charTest":          "行3char",
				"charTextTest":      "行3text",
				"dateTest":          time.Date(2021, 7, 3, 0, 0, 0, 0, time.UTC),
				"dateTimeTest":      time.Date(2021, 7, 3, 15, 38, 50, 0, time.UTC),
				"timestampTest":     time.Date(2021, 7, 3, 15, 38, 50, 0, time.UTC),
				"floatTest":         float64(3.456),
				"doubleTest":        float64(3.15678),
				"decimalTest":       "3.4567899900",
				"bitTest":           []byte{1},
				"nullIntTest":       int64(3),
				"nullTinyintTest":   int64(3),
				"nullSmallIntTest":  int64(3),
				"nullBigIntTest":    int64(3),
				"nullUnsignedTest":  int64(3),
				"nullVarcharTest":   "行3",
				"nullCharTest":      "行3char",
				"nullTextTest":      "行3text",
				"nullDateTest":      time.Date(2021, 7, 3, 0, 0, 0, 0, time.UTC),
				"nullDateTimeTest":  time.Date(2021, 7, 3, 15, 38, 50, 0, time.UTC),
				"nullTimestampTest": time.Date(2021, 7, 3, 15, 38, 50, 0, time.UTC),
				"nullFloatTest":     float64(3.456),
				"nullDoubleTest":    float64(3.15678),
				"nullDecimalTest":   "3.4567899900",
				"nullBitTest":       []byte{1},
				"colNull":           nil,
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
					}
					t.Errorf("fieldname = %s, internalDbClient.Get() = %v, want %v", k, v, wantV)
				}
			}
		})
	}
}

func Test_internalDbClient_Rows(t *testing.T) {
	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		sqlText string
		args    []any
	}
	tests := []struct {
		name    string
		client  sqlmer.DbClient
		args    args
		want    []map[string]any
		wantErr bool
	}{
		{
			"mysql",
			mysqlClient,
			args{
				"SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (1,2)",
				[]any{},
			},
			[]map[string]any{
				{
					"varcharTest":   "行1",
					"dateTest":      time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
					"dateTimeTest":  time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
					"timestampTest": time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
					"decimalTest":   "1.4567899900",
				},
				{
					"varcharTest":   "行2",
					"dateTest":      time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC),
					"dateTimeTest":  time.Date(2021, 7, 2, 15, 38, 50, 0, time.UTC),
					"timestampTest": time.Date(2021, 7, 2, 15, 38, 50, 0, time.UTC),
					"decimalTest":   "2.4567899900",
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
				got, err := rows.MapScan()
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
