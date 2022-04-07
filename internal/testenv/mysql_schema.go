package testenv

import "database/sql"

var mysqlSchemas []Schema = []Schema{
	{Create: `
CREATE TABLE go_TypeTest (
	id int(11) NOT NULL AUTO_INCREMENT,
	varcharTest varchar(10) NOT NULL,
	dateTest date NULL,
	dateTimeTest datetime NULL,
	timestampTest timestamp NULL,
	decimalTest DECIMAL(15, 10) NULL,
	PRIMARY KEY (id),
	KEY idx_go_TypeTest (id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
`, Drop: `drop table go_TypeTest`},
	{Create: `
INSERT INTO go_TypeTest(VarcharTest, dateTest, dateTimeTest, timestampTest, decimalTest)
VALUES (N'行1', '2021-07-01', '2021-07-01 15:38:50.425','2021-07-01 15:38:50.425', 1.45678999);
`, Drop: ``},
	{Create: `
INSERT INTO go_TypeTest(VarcharTest, dateTest, dateTimeTest, timestampTest, decimalTest)
VALUES (N'行2', '2021-07-02', '2021-07-02 15:38:50.425','2021-07-02 15:38:50.425', 2.45678999);
`, Drop: ``},
	{Create: `
INSERT INTO go_TypeTest(VarcharTest, dateTest, dateTimeTest, timestampTest, decimalTest)
VALUES (N'行3', '2021-07-03', '2021-07-03 15:38:50.425','2021-07-03 15:38:50.425', 3.45678999);
`, Drop: ``},
	{Create: `
INSERT INTO go_TypeTest(VarcharTest, dateTest, dateTimeTest, timestampTest, decimalTest)
VALUES (N'行4', '2021-07-04', '2021-07-04 15:38:50.425','2021-07-04 15:38:50.425', 4.45678999);
`, Drop: ``},
	{Create: `
INSERT INTO go_TypeTest(VarcharTest, dateTest, dateTimeTest, timestampTest, decimalTest)
VALUES (N'行5', null,null,null,null);
`, Drop: ``},
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
