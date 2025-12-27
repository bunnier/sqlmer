package testenv

import "database/sql"

var sqliteSchemas []schema = []schema{
	{
		Create: `
CREATE TABLE go_TypeTest (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	intTest INTEGER NOT NULL,
	tinyintTest INTEGER NOT NULL,
	smallIntTest INTEGER NOT NULL,
	bigIntTest INTEGER NOT NULL,
	unsignedTest INTEGER NOT NULL,
	varcharTest VARCHAR(10) NOT NULL,
	charTest CHAR(10) NOT NULL,
	charTextTest TEXT NOT NULL,
	dateTest DATE NOT NULL,
	dateTimeTest DATETIME NOT NULL,
	timestampTest TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	floatTest FLOAT NOT NULL,
	doubleTest DOUBLE NOT NULL,
	decimalTest DECIMAL(15, 10) NOT NULL,
	bitTest INTEGER NOT NULL,

	nullIntTest INTEGER NULL,
	nullTinyintTest INTEGER NULL,
	nullSmallIntTest INTEGER NULL,
	nullBigIntTest INTEGER NULL,
	nullUnsignedTest INTEGER NULL,
	nullVarcharTest VARCHAR(10) NULL,
	nullCharTest CHAR(10) NULL,
	nullTextTest TEXT NULL,
	nullDateTest DATE NULL,
	nullDateTimeTest DATETIME NULL,
	nullTimestampTest TIMESTAMP NULL,
	nullFloatTest FLOAT NULL,
	nullDoubleTest DOUBLE NULL,

	nullDecimalTest DECIMAL(15, 10) NULL,
	nullBitTest INTEGER NULL
);
`,
		Drop: `DROP TABLE go_TypeTest`,
	},
	{
		Create: `
INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
VALUES (1, 1, 1, 1, 1, '行1', '行1char', '行1text','2021-07-01','2021-07-01 15:38:50.425','2021-07-01 15:38:50.425', 1.456, 1.15678, '1.4567899900', 1);
`, Drop: ``,
	},
	{
		Create: `
INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
VALUES (2, 2, 2, 2, 2, '行2', '行2char', '行2text', '2021-07-02', '2021-07-02 15:38:50.425','2021-07-02 15:38:50.425', 2.456, 2.15678, '2.4567899900', 1);
`,
		Drop: ``,
	},
	{
		Create: `
INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
	nullIntTest, nullTinyintTest, nullSmallIntTest, nullBigIntTest, nullUnsignedTest, nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest)
VALUES (3, 3, 3, 3, 3, '行3', '行3char', '行3text', '2021-07-03', '2021-07-03 15:38:50.425','2021-07-03 15:38:50.425', 3.456, 3.15678, '3.4567899900', 1,
	3, 3, 3, 3, 3, '行3', '行3char', '行3text', '2021-07-03', '2021-07-03 15:38:50.425','2021-07-03 15:38:50.425', 3.456, 3.15678, '3.4567899900', 1);
`,
		Drop: ``,
	},
	{
		Create: `
INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
	nullIntTest, nullTinyintTest, nullSmallIntTest, nullBigIntTest, nullUnsignedTest, nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest)
VALUES (4, 4, 4, 4, 4, '行4', '行4char', '行4text', '2021-07-04', '2021-07-04 15:38:50.425','2021-07-04 15:38:50.425', 4.456, 4.15678, '4.4567899900', 0,
	4, 4, 4, 4, 4, '行4', '行4char', '行4text', '2021-07-04', '2021-07-04 15:38:50.425','2021-07-04 15:38:50.425', 4.456, 4.15678, '4.4567899900', 0);
`,
		Drop: ``,
	},
}

// CreateSqliteSchema 用于初始化 Sqlite 的测试用表结构。
func CreateSqliteSchema(db *sql.DB) error {
	for _, schema := range sqliteSchemas {
		_, err := db.Exec(schema.Create)
		if err != nil {
			return err
		}
	}
	return nil
}

// DropSqliteSchema 用于初始化 Sqlite 的测试用表结构。
func DropSqliteSchema(db *sql.DB) error {
	for _, schema := range sqliteSchemas {
		if schema.Drop == "" {
			continue
		}
		_, err := db.Exec(schema.Drop)
		if err != nil {
			return err
		}
	}
	return nil
}
