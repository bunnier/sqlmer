package testenv

import (
	"database/sql"
)

var mssqlSchemas []Schema = []Schema{
	{
		Create: `
CREATE TABLE go_TypeTest
(
	Id int NOT NULL IDENTITY(1, 1),

	TinyIntTest TINYINT NOT NULL,
	SmallIntTest SMALLINT NOT NULL,
	IntTest INT NOT NULL,
	BitTest BIT NOT NULL,
	NvarcharTest NVARCHAR(10) NOT NULL,
	VarcharTest VARCHAR(10) NOT NULL,
	NcharTest NVARCHAR(10) NOT NULL,
	CharTest VARCHAR(10) NOT NULL,
	DateTimeTest DATETIME NOT NULL,
	DateTime2Test DATETIME2 NOT NULL,
	DateTest DATE NOT NULL,
	TimeTest TIME NOT NULL,
	MoneyTest MONEY NOT NULL,
	FloatTest FLOAT(10) NOT NULL,
	DecimalTest DECIMAL(38, 10) NOT NULL,
	
	NullableTinyIntTest TINYINT NULL,
	NullableSmallIntTest SMALLINT NULL,
	NullableIntTest INT NULL,
	NullableBitTest BIT NULL,
	NullableNvarcharTest NVARCHAR(10) NULL,
	NullableVarcharTest VARCHAR(10) NULL,
	NullableNcharTest NVARCHAR(10) NULL,
	NullableCharTest VARCHAR(10) NULL,
	NullableDateTimeTest DATETIME NULL,
	NullableDateTime2Test DATETIME2 NULL,
	NullableDateTest DATE NULL,
	NullableTimeTest TIME NULL,
	NullableMoneyTest MONEY NULL,
	NullableFloatTest FLOAT(10) NULL,
	NullableDecimalTest DECIMAL(38, 10) NULL
);

ALTER TABLE go_TypeTest ADD CONSTRAINT PK_go_TypeTest PRIMARY KEY CLUSTERED (Id);

INSERT INTO go_TypeTest (TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest)
VALUES (1, 1, 1, 1, N'行1', 'Row1', N'行1', 'Row1', '2021-07-01 15:38:39.583', '2021-07-01 15:38:50.4257813', '2021-07-01', '12:01:01.345', 1.123, 1.12345, 1.45678999);

INSERT INTO go_TypeTest (TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest)
VALUES (2, 2, 2, 2, N'行2', 'Row2', N'行2', 'Row2', '2021-07-02 15:38:39.583', '2021-07-02 15:38:50.4257813', '2021-07-02', '12:02:01.345', 2.123, 2.12345, 2.45678999);

INSERT INTO go_TypeTest (TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest,
NullableTinyIntTest, NullableSmallIntTest, NullableIntTest, NullableBitTest, NullableNvarcharTest, NullableVarcharTest, NullableNcharTest, NullableCharTest, NullableDateTimeTest, NullableDateTime2Test, NullableDateTest, NullableTimeTest, NullableMoneyTest, NullableFloatTest, NullableDecimalTest)
VALUES (3, 3, 3, 3, N'行3', 'Row3', N'行3', 'Row3', '2021-07-03 15:38:39.583', '2021-07-03 15:38:50.4257813', '2021-07-03', '12:03:01.345', 3.123, 3.12345, 3.45678999,
3, 3, 3, 3, N'行3', 'Row3', N'行3', 'Row3', '2021-07-03 15:38:39.583', '2021-07-03 15:38:50.4257813', '2021-07-03', '12:03:01.345', 3.123, 3.12345, 3.45678999);

INSERT INTO go_TypeTest (TinyIntTest, SmallIntTest, IntTest, BitTest, NvarcharTest, VarcharTest, NcharTest, CharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, MoneyTest, FloatTest, DecimalTest,
NullableTinyIntTest, NullableSmallIntTest, NullableIntTest, NullableBitTest, NullableNvarcharTest, NullableVarcharTest, NullableNcharTest, NullableCharTest, NullableDateTimeTest, NullableDateTime2Test, NullableDateTest, NullableTimeTest, NullableMoneyTest, NullableFloatTest, NullableDecimalTest)
VALUES (4, 4, 4, 4, N'行4', 'Row4', N'行4', 'Row4', '2021-07-04 15:38:39.583', '2021-07-04 15:38:50.4257813', '2021-07-04', '12:04:01.345', 4.123, 4.12345, 4.45678999,
4, 4, 4, 4, N'行4', 'Row4', N'行4', 'Row4', '2021-07-04 15:38:39.583', '2021-07-04 15:38:50.4257813', '2021-07-04', '12:04:01.345', 4.123, 4.12345, 4.45678999);
`,
		Drop: `DROP TABLE go_TypeTest`,
	},
}

// CreateMssqlSchema 用于初始化 MsSql 的测试用表结构。
func CreateMssqlSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, schema := range mssqlSchemas {
		_, err := db.Exec(schema.Create)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// DropMssqlSchema 用于初始化 MsSql 的测试用表结构。
func DropMssqlSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, schema := range mssqlSchemas {
		if schema.Drop == "" {
			continue
		}
		_, err := db.Exec(schema.Drop)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
