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
	"github.com/bunnier/sqlmer/sqlen"
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

	t.Run("mysql1", func(t *testing.T) {
		got, _, err := mysqlClient.Scalar("SELECT Id FROM go_TypeTest WHERE id=@p1", 1)
		if err != nil {
			t.Errorf("internalDbClient.Scalar() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, int64(1)) {
			t.Errorf("internalDbClient.Scalar() = %v, want %v", got, int64(1))
		}
	})

	t.Run("mysql2", func(t *testing.T) {
		got, _, err := mysqlClient.Scalar("SELECT Id FROM go_TypeTest WHERE id in (@p1)", []int{1})
		if err != nil {
			t.Errorf("internalDbClient.Scalar() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, int64(1)) {
			t.Errorf("internalDbClient.Scalar() = %v, want %v", got, int64(1))
		}
	})

	t.Run("mysql3", func(t *testing.T) {
		got, _, err := mysqlClient.Scalar("SELECT COUNT(1) FROM go_TypeTest WHERE id in (@p1)", []int{1, 2, 3})
		if err != nil {
			t.Errorf("internalDbClient.Scalar() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, int64(3)) {
			t.Errorf("internalDbClient.Scalar() = %v, want %v", got, int64(3))
		}
	})

	t.Run("mysql4", func(t *testing.T) {
		got, _, err := mysqlClient.Scalar("SELECT COUNT(1) FROM go_TypeTest WHERE id=@p1 OR id=@p2 OR id=@name1 OR id=@name2",
			map[string]any{
				"p1":    1,
				"p2":    2,
				"name1": 3,
				"name2": 4,
			})

		if err != nil {
			t.Errorf("internalDbClient.Scalar() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, int64(4)) {
			t.Errorf("internalDbClient.Scalar() = %v, want %v", got, int64(4))
		}
	})

	t.Run("mysql5", func(t *testing.T) {
		got, _, err := mysqlClient.Scalar("SELECT COUNT(1) FROM go_TypeTest WHERE varcharTest=@varcharTest OR id=@p1 OR id=@p2",
			3,
			map[string]any{"varcharTest": "行1"},
			2)

		if err != nil {
			t.Errorf("internalDbClient.Scalar() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, int64(3)) {
			t.Errorf("internalDbClient.Scalar() = %v, want %v", got, int64(3))
		}
	})
}

func Test_internalDbClient_Execute(t *testing.T) {
	now := time.Now()
	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("mysql", func(t *testing.T) {
		sqlText := `INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
				VALUES (5, 5, 5, 5, 5, N'行5', '行5char', '行5text','2021-07-05','2021-07-05 15:38:50.425','2021-07-05 15:38:50.425', 5.456, 5.15678, 5.45678999, 1);`

		effectRow, err := mysqlClient.Execute(sqlText, now)
		if err != nil {
			t.Errorf("internalDbClient.Execute() error = %v, wantErr %v", err, false)
		}

		if effectRow != int64(1) {
			t.Errorf("internalDbClient.Execute() effectRow = %v, want %v", effectRow, int64(1))
		}

		err = mysqlClient.SizedExecute(1, sqlText, now)
		if err != nil {
			t.Errorf("internalDbClient.SizedExecute() error = %v, wantErr %v", err, false)
		}

		err = mysqlClient.SizedExecute(2, sqlText, now)
		if !errors.Is(err, sqlmer.ErrExpectedSizeWrong) {
			t.Errorf("internalDbClient.SizedExecute() error = %v, wantErr DbSqlError", err)
		}
	})
}

func Test_internalDbClient_Exists(t *testing.T) {
	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("mysql_exist", func(t *testing.T) {
		got, err := mysqlClient.Exists("SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id=1")
		if err != nil {
			t.Errorf("internalDbClient.Exists() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, true) {
			t.Errorf("internalDbClient.Exists() = %v, want %v", got, true)
		}
	})

	t.Run("mysql_notexist", func(t *testing.T) {
		got, err := mysqlClient.Exists("SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id=10000")
		if err != nil {
			t.Errorf("internalDbClient.Exists() error = %v, wantErr %v", err, false)
			return
		}
		if !reflect.DeepEqual(got, false) {
			t.Errorf("internalDbClient.Exists() = %v, want %v", got, false)
		}
	})
}

func Test_internalDbClient_Get(t *testing.T) {
	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("mysql_nullable_null", func(t *testing.T) {
		got, err := mysqlClient.Get(`SELECT intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
				nullIntTest, nullTinyintTest, nullSmallIntTest, nullBigIntTest, nullUnsignedTest, nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest,
				null colNull 
				FROM go_TypeTest WHERE id=1`)
		if err != nil {
			t.Errorf("internalDbClient.Get() error = %v, wantErr %v", err, false)
			return
		}

		want := map[string]any{
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
		}

		for k, v := range got {
			wantV := want[k]
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

	t.Run("mysql_nullable_hasValue", func(t *testing.T) {
		got, err := mysqlClient.Get(`SELECT intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
				nullIntTest, nullTinyintTest, nullSmallIntTest, nullBigIntTest, nullUnsignedTest, nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest,
				null colNull 
				FROM go_TypeTest WHERE id=3`)
		if err != nil {
			t.Errorf("internalDbClient.Get() error = %v, wantErr %v", err, false)
			return
		}

		want := map[string]any{
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
		}

		for k, v := range got {
			wantV := want[k]
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

func Test_internalDbClient_Rows(t *testing.T) {
	row1 := map[string]any{
		"varcharTest":   "行1",
		"dateTest":      time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		"dateTimeTest":  time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
		"timestampTest": time.Date(2021, 7, 1, 15, 38, 50, 0, time.UTC),
		"decimalTest":   "1.4567899900",
	}
	row2 := map[string]any{
		"varcharTest":   "行2",
		"dateTest":      time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC),
		"dateTimeTest":  time.Date(2021, 7, 2, 15, 38, 50, 0, time.UTC),
		"timestampTest": time.Date(2021, 7, 2, 15, 38, 50, 0, time.UTC),
		"decimalTest":   "2.4567899900",
	}
	row3 := map[string]any{
		"varcharTest":   "行3",
		"dateTest":      time.Date(2021, 7, 3, 0, 0, 0, 0, time.UTC),
		"dateTimeTest":  time.Date(2021, 7, 3, 15, 38, 50, 0, time.UTC),
		"timestampTest": time.Date(2021, 7, 3, 15, 38, 50, 0, time.UTC),
		"decimalTest":   "3.4567899900",
	}
	row4 := map[string]any{
		"varcharTest":   "行4",
		"dateTest":      time.Date(2021, 7, 4, 0, 0, 0, 0, time.UTC),
		"dateTimeTest":  time.Date(2021, 7, 4, 15, 38, 50, 0, time.UTC),
		"timestampTest": time.Date(2021, 7, 4, 15, 38, 50, 0, time.UTC),
		"decimalTest":   "4.4567899900",
	}

	mysqlClient, err := testenv.NewMysqlClient()
	if err != nil {
		t.Fatal(err)
	}

	rowAssert := func(rows *sqlen.EnhanceRows, want []map[string]any) {
		index := 0
		for rows.Next() {
			got, err := rows.MapScan()
			if err != nil {
				t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
				return
			}
			if !reflect.DeepEqual(got, want[index]) {
				t.Errorf("internalDbClient.Get() = %v, want %v", got, want[index])
				return
			}
			index++
		}
	}

	t.Run("noParams", func(t *testing.T) {
		rows, err := mysqlClient.Rows("SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (1,2)")
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row1, row2})
	})

	t.Run("arrayParams", func(t *testing.T) {
		rows, err := mysqlClient.Rows("SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (@p1)", []int{1, 2})
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row1, row2})
	})

	t.Run("arrayParams2", func(t *testing.T) {
		rows, err := mysqlClient.Rows("SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (@ids)",
			map[string]any{
				"ids": []int{2},
			})

		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row2})
	})

	t.Run("structParams", func(t *testing.T) {
		rows, err := mysqlClient.Rows("SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (@Ids)",
			struct {
				Ids []int
			}{
				Ids: []int{2},
			})
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row2})
	})

	t.Run("structParamsMerge1", func(t *testing.T) {
		rows, err := mysqlClient.Rows("SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (@Ids) AND id!=@p1",
			struct{ Ids []int }{
				Ids: []int{1, 2},
			},
			1,
		)
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row2})
	})

	t.Run("structParamsMerge2", func(t *testing.T) {
		rows, err := mysqlClient.Rows("SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (@Ids) AND id!=@p1",
			struct{ Ids []int }{Ids: []int{1, 2}},
			1,
			map[string]any{"Ids": []int{1, 2, 3}},
		)
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row2, row3})
	})

	t.Run("structParamsMerge3", func(t *testing.T) {
		rows, err := mysqlClient.Rows("SELECT varcharTest,dateTest,dateTimeTest,timestampTest,decimalTest FROM go_TypeTest WHERE id IN (@Ids) OR id IN (@p1, @p2)",
			map[string]any{"Ids": []int{1}},
			2,
			4,
			struct{ Ids []int }{Ids: []int{3}},
		)
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row2, row3, row4})
	})
}
