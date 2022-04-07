package testenv

import "database/sql"

var mysqlSchemas []Schema = []Schema{
	{
		Create: `
CREATE TABLE go_TypeTest (
	id int(11) NOT NULL AUTO_INCREMENT,
	varcharTest varchar(10) NOT NULL,
	charTest char(10) NOT NULL,
	charTextTest text NOT NULL,
	dateTest date NOT NULL,
	dateTimeTest datetime NOT NULL,
	timestampTest timestamp NOT NULL,
	floatTest float(5, 3) NOT NULL,
	doubleTest double(10, 5) NOT NULL,
	decimalTest DECIMAL(15, 10) NOT NULL,
	bitTest bit(1) NOT NULL,
	
	nullVarcharTest varchar(10) NULL,
	nullCharTest char(10) NULL,
	nullTextTest text NULL,
	nullDateTest date NULL,
	nullDateTimeTest datetime NULL,
	nullTimestampTest timestamp NULL,
	nullFloatTest float(5, 3) NULL,
	nullDoubleTest double(10, 5) NULL,
	nullDecimalTest DECIMAL(15, 10) NULL,
	nullBitTest bit(1) NULL,

	PRIMARY KEY (id),
	KEY idx_go_TypeTest (id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
`,
		Drop: `DROP TABLE go_TypeTest`,
	},
	{
		Create: `
INSERT INTO go_TypeTest(varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
VALUES (N'行1', '行1char', '行1text','2021-07-01','2021-07-01 15:38:50.425','2021-07-01 15:38:50.425', 1.456, 1.15678, 1.45678999, 1);
`, Drop: ``,
	},
	{
		Create: `
INSERT INTO go_TypeTest(varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
VALUES (N'行2', '行2char', '行2text', '2021-07-02', '2021-07-02 15:38:50.425','2021-07-02 15:38:50.425', 2.456, 2.15678, 2.45678999, 1);
`,
		Drop: ``,
	},
	{
		Create: `
INSERT INTO go_TypeTest(varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest)
VALUES (N'行3', '行3char', '行3text', '2021-07-03', '2021-07-03 15:38:50.425','2021-07-03 15:38:50.425', 3.456, 3.15678, 3.45678999, 1,
N'行3', '行3char', '行3text', '2021-07-03', '2021-07-03 15:38:50.425','2021-07-03 15:38:50.425', 3.456, 3.15678, 3.45678999, 1);
`,
		Drop: ``,
	},
	{
		Create: `
INSERT INTO go_TypeTest(varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
	nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest)
VALUES (N'行4', '行4char', '行4text', '2021-07-04', '2021-07-04 15:38:50.425','2021-07-04 15:38:50.425', 4.456, 4.15678, 4.45678999, 0,
N'行4', '行4char', '行4text', '2021-07-04', '2021-07-04 15:38:50.425','2021-07-04 15:38:50.425', 4.456, 4.15678, 4.45678999, 0);
`,
		Drop: ``,
	},
}

// CreateMysqlSchema 用于初始化 MySql 的测试用表结构。
func CreateMysqlSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, schema := range mysqlSchemas {
		_, err := db.Exec(schema.Create)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// DropMysqlSchema 用于初始化MySql的测试用表结构。
func DropMysqlSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, schema := range mysqlSchemas {
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