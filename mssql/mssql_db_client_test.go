package mssql_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"errors"

	"github.com/bunnier/sqlmer"
	"github.com/bunnier/sqlmer/internal/testenv"
	"github.com/bunnier/sqlmer/mssql"
	"github.com/bunnier/sqlmer/sqlen"
)

func init() {
	testenv.TryInitConfig("..")
}

func Test_NewMsSqlDbClient(t *testing.T) {
	dbClient, err := testenv.NewSqlServerClient()
	if err != nil {
		t.Fatal(err)
	}

	if dbClient.Dsn() != testenv.TestConf.SqlServer {
		t.Errorf("mssqlDbClient.Dsn() connString = %v, wantConnString %v", dbClient.Dsn(), testenv.TestConf.SqlServer)
	}

	if dbClient, err = mssql.NewMsSqlDbClient("test",
		sqlmer.WithConnTimeout(time.Second*15),
		sqlmer.WithExecTimeout(time.Second*15),
		sqlmer.WithPingCheck(true)); err == nil {
		t.Errorf("mssqlDbClient.NewMsSqlDbClient() err = nil, want has a err")
	}
}

func Test_internalDbClient_Scalar(t *testing.T) {
	mssqlClient, err := testenv.NewSqlServerClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("mssql", func(t *testing.T) {
		got, _, err := mssqlClient.Scalar("SELECT Id FROM go_TypeTest WHERE Id=@p1", 1)
		if err != nil {
			t.Errorf("internalDbClient.Scalar() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, int64(1)) {
			t.Errorf("internalDbClient.Scalar() = %v, want %v", got, int64(1))
		}
	})
}

func Test_internalDbClient_Execute(t *testing.T) {
	mssqlClient, err := testenv.NewSqlServerClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("mssql", func(t *testing.T) {
		sqlText := `INSERT INTO go_TypeTest (TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest, BinaryTest)
				VALUES (5, 5, 5, 5, N'行5', 'Row5', N'行5', 'Row5', '2021-07-05 15:38:39.583', '2021-07-05 15:38:50.4257813', '2021-07-05', '12:05:01.345', 5.123, 5.12345, 5.45678999, 1);`

		effectRow, err := mssqlClient.Execute(sqlText)
		if err != nil {
			t.Errorf("internalDbClient.Execute() error = %v, wantErr false", err)
		}

		if effectRow != int64(1) {
			t.Errorf("internalDbClient.Execute() effectRow = %v, want 1", effectRow)
		}

		err = mssqlClient.SizedExecute(1, sqlText)
		if err != nil {
			t.Errorf("internalDbClient.SizedExecute() error = %v, wantErr false", err)
		}

		err = mssqlClient.SizedExecute(2, sqlText)
		if !errors.Is(err, sqlmer.ErrExpectedSizeWrong) {
			t.Errorf("internalDbClient.SizedExecute() error = %v, wantErr DbSqlError", err)
		}
	})
}

func Test_internalDbClient_Exists(t *testing.T) {
	mssqlClient, err := testenv.NewSqlServerClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("mssql_exist", func(t *testing.T) {
		got, err := mssqlClient.Exists("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest FROM go_TypeTest WHERE Id=1")
		if err != nil {
			t.Errorf("internalDbClient.Exists() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, true) {
			t.Errorf("internalDbClient.Exists() = %v, want %v", got, true)
		}
	})

	t.Run("mssql_notexist", func(t *testing.T) {
		got, err := mssqlClient.Exists("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest FROM go_TypeTest WHERE Id=10000")
		if err != nil {
			t.Errorf("internalDbClient.Exists() error = %v, wantErr false", err)
			return
		}
		if !reflect.DeepEqual(got, false) {
			t.Errorf("internalDbClient.Exists() = %v, want %v", got, false)
		}
	})
}

func Test_internalDbClient_Get(t *testing.T) {
	row1 := map[string]any{
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
		"colNull":               nil,
	}

	row3 := map[string]any{
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
		"colNull":               nil,
	}

	mssqlClient, err := testenv.NewSqlServerClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("mssql_nullable_null", func(t *testing.T) {
		got, err := mssqlClient.Get(`SELECT TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest, BinaryTest,
				NullableTinyIntTest, NullableSmallIntTest, NullableIntTest, NullableBitTest, NullableNvarcharTest, NullableVarcharTest, NullableNcharTest, NullableCharTest, NullableDateTimeTest, NullableDateTime2Test, NullableDateTest, NullableTimeTest, NullableMoneyTest, NullableFloatTest, NullableDecimalTest, NullableBinaryTest,
				null colNull
				FROM go_TypeTest WHERE Id=1`)
		if err != nil {
			t.Errorf("internalDbClient.Get() error = %v, wantErr false", err)
			return
		}

		for k, v := range got {
			wantV := row1[k]
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

	t.Run("mssql_nullable_hasValue", func(t *testing.T) {
		got, err := mssqlClient.Get(`SELECT TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest, BinaryTest,
				NullableTinyIntTest, NullableSmallIntTest, NullableIntTest, NullableBitTest, NullableNvarcharTest, NullableVarcharTest, NullableNcharTest, NullableCharTest, NullableDateTimeTest, NullableDateTime2Test, NullableDateTest, NullableTimeTest, NullableMoneyTest, NullableFloatTest, NullableDecimalTest, NullableBinaryTest,
				null colNull
				FROM go_TypeTest WHERE Id=3`)
		if err != nil {
			t.Errorf("internalDbClient.Get() error = %v, wantErr false", err)
			return
		}

		for k, v := range got {
			wantV := row3[k]
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

func Test_internalDbClient_Rows(t *testing.T) {
	row1 := map[string]any{
		"NvarcharTest":  "行1",
		"VarcharTest":   "Row1",
		"DateTimeTest":  time.Date(2021, 7, 1, 15, 38, 39, 583000000, time.UTC),
		"DateTime2Test": time.Date(2021, 7, 1, 15, 38, 50, 425781300, time.UTC),
		"DateTest":      time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		"TimeTest":      time.Date(1, 1, 1, 12, 1, 1, 345000000, time.UTC),
		"DecimalTest":   "1.4567899900",
	}
	row2 := map[string]any{
		"NvarcharTest":  "行2",
		"VarcharTest":   "Row2",
		"DateTimeTest":  time.Date(2021, 7, 2, 15, 38, 39, 583000000, time.UTC),
		"DateTime2Test": time.Date(2021, 7, 2, 15, 38, 50, 425781300, time.UTC),
		"DateTest":      time.Date(2021, 7, 2, 0, 0, 0, 0, time.UTC),
		"TimeTest":      time.Date(1, 1, 1, 12, 2, 1, 345000000, time.UTC),
		"DecimalTest":   "2.4567899900",
	}
	row3 := map[string]any{
		"NvarcharTest":  "行3",
		"VarcharTest":   "Row3",
		"DateTimeTest":  time.Date(2021, 7, 3, 15, 38, 39, 583000000, time.UTC),
		"DateTime2Test": time.Date(2021, 7, 3, 15, 38, 50, 425781300, time.UTC),
		"DateTest":      time.Date(2021, 7, 3, 0, 0, 0, 0, time.UTC),
		"TimeTest":      time.Date(1, 1, 1, 12, 3, 1, 345000000, time.UTC),
		"DecimalTest":   "3.4567899900",
	}
	row4 := map[string]any{
		"NvarcharTest":  "行4",
		"VarcharTest":   "Row4",
		"DateTimeTest":  time.Date(2021, 7, 4, 15, 38, 39, 583000000, time.UTC),
		"DateTime2Test": time.Date(2021, 7, 4, 15, 38, 50, 425781300, time.UTC),
		"DateTest":      time.Date(2021, 7, 4, 0, 0, 0, 0, time.UTC),
		"TimeTest":      time.Date(1, 1, 1, 12, 4, 1, 345000000, time.UTC),
		"DecimalTest":   "4.4567899900",
	}

	type IdsType struct {
		Ids []int
	}

	mssqlClient, err := testenv.NewSqlServerClient()
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
		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (1,2)")
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row1, row2})
	})

	t.Run("arrayParams", func(t *testing.T) {
		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (@p1)", []int{1, 2})
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row1, row2})
	})

	t.Run("arrayParams2", func(t *testing.T) {
		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (@ids)",
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
		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (@Ids)",
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
		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (@Ids) AND id!=@p1",
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
		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (@Ids) AND id!=@p1",
			IdsType{Ids: []int{1, 2}},
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
		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (@Ids) OR id IN (@p1, @p2)",
			map[string]any{"Ids": []int{1}},
			2,
			4,
			IdsType{Ids: []int{3}},
		)
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row2, row3, row4})
	})

	t.Run("structParamsMerge4", func(t *testing.T) {
		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (@Ids) OR dateTest=@p1",
			map[string]any{"Ids": []int{1}},
			time.Date(2021, 7, 3, 0, 0, 0, 0, time.UTC),
			IdsType{Ids: []int{2}},
		)
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row2, row3})
	})

	t.Run("structParamsMerge5", func(t *testing.T) {
		type IdsType struct {
			Ids []int
		}

		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (@Ids)",
			IdsType{Ids: []int{2}},
			IdsType{Ids: []int{1}},
		)
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row1})
	})

	t.Run("structParamsMerge6", func(t *testing.T) {
		type IdsType struct {
			Ids []int
		}

		rows, err := mssqlClient.Rows("SELECT NvarcharTest,VarcharTest,DateTimeTest,DateTime2Test,DateTest,TimeTest,DecimalTest  FROM go_TypeTest WHERE id IN (@Ids)",
			IdsType{Ids: []int{2}},
			struct {
				IdsType
			}{
				IdsType{
					Ids: []int{1}, // 嵌套类型也应该覆盖。
				},
			},
		)
		if err != nil {
			t.Errorf("internalDbClient.Rows() error = %v, wantErr %v", err, false)
			return
		}
		defer rows.Close()

		rowAssert(rows, []map[string]any{row1})
	})
}
