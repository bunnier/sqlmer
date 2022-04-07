CREATE TABLE go_TypeTest (
	id int(11) NOT NULL AUTO_INCREMENT,
	intTest int NOT NULL,
	tinyintTest tinyint NOT NULL,
	smallIntTest smallint NOT NULL,
	bigIntTest bigInt NOT NULL,
	unsignedTest int unsigned NOT NULL,
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
	
	nullIntTest int NULL,
	nullTinyintTest tinyint NULL,
	nullSmallIntTest smallint NULL,
	nullBigIntTest bigInt NULL,
	nullUnsignedTest int unsigned NULL,
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

INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
VALUES (1, 1, 1, 1, 1, N'行1', '行1char', '行1text','2021-07-01','2021-07-01 15:38:50.425','2021-07-01 15:38:50.425', 1.456, 1.15678, 1.45678999, 1);

INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest)
VALUES (2, 2, 2, 2, 2, N'行2', '行2char', '行2text', '2021-07-02', '2021-07-02 15:38:50.425','2021-07-02 15:38:50.425', 2.456, 2.15678, 2.45678999, 1);

INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
	nullIntTest, nullTinyintTest, nullSmallIntTest, nullBigIntTest, nullUnsignedTest, nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest)
VALUES (3, 3, 3, 3, 3, N'行3', '行3char', '行3text', '2021-07-03', '2021-07-03 15:38:50.425','2021-07-03 15:38:50.425', 3.456, 3.15678, 3.45678999, 1,
	3, 3, 3, 3, 3, N'行3', '行3char', '行3text', '2021-07-03', '2021-07-03 15:38:50.425','2021-07-03 15:38:50.425', 3.456, 3.15678, 3.45678999, 1);

INSERT INTO go_TypeTest(intTest, tinyintTest, smallIntTest, bigIntTest, unsignedTest, varcharTest, charTest, charTextTest, dateTest, dateTimeTest, timestampTest, floatTest, doubleTest, decimalTest, bitTest,
	nullIntTest, nullTinyintTest, nullSmallIntTest, nullBigIntTest, nullUnsignedTest, nullVarcharTest, nullCharTest, nullTextTest, nullDateTest, nullDateTimeTest, nullTimestampTest, nullFloatTest, nullDoubleTest, nullDecimalTest, nullBitTest)
VALUES (4, 4, 4, 4, 4, N'行4', '行4char', '行4text', '2021-07-04', '2021-07-04 15:38:50.425','2021-07-04 15:38:50.425', 4.456, 4.15678, 4.45678999, 0,
	4, 4, 4, 4, 4, N'行4', '行4char', '行4text', '2021-07-04', '2021-07-04 15:38:50.425','2021-07-04 15:38:50.425', 4.456, 4.15678, 4.45678999, 0);