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
	NvarcharTest nvarchar(10) NULL,
	VarcharTest varchar(10) NULL,
	DateTimeTest datetime NULL,
	DateTime2Test datetime2 NULL,
	DateTest date NULL,
	TimeTest time NULL,
	DecimalTest decimal(38, 10) NULL
);
ALTER TABLE go_TypeTest ADD CONSTRAINT PK_go_TypeTest PRIMARY KEY CLUSTERED (Id);

INSERT INTO go_TypeTest (NvarcharTest, VarcharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, DecimalTest)
VALUES (N'行1', 'Row1', '2021-07-01 15:38:39.583', '2021-07-01 15:38:50.4257813', '2021-07-01', '12:01:01.345', 1.45678999);

INSERT INTO go_TypeTest (NvarcharTest, VarcharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, DecimalTest)
VALUES (N'行2', 'Row2', '2021-07-02 15:38:39.583', '2021-07-02 15:38:50.4257813', '2021-07-02', '12:01:02.345', 2.45678999);

INSERT INTO go_TypeTest (NvarcharTest, VarcharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, DecimalTest)
VALUES (N'行3', 'Row3', '2021-07-03 15:38:39.583', '2021-07-03 15:38:50.4257813', '2021-07-03', '12:01:03.345', 3.45678999);

INSERT INTO go_TypeTest (NvarcharTest, VarcharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, DecimalTest)
VALUES (N'行4', 'Row4', '2021-07-04 15:38:39.583', '2021-07-04 15:38:50.4257813', '2021-07-04', '12:01:04.345', 4.45678999);

INSERT INTO go_TypeTest (NvarcharTest, VarcharTest, DateTimeTest, DateTime2Test, DateTest, TimeTest, DecimalTest)
VALUES (N'行4', null, null, null, null, null,null);
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
