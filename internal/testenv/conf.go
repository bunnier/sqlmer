package testenv

const (
	// MySqlDsn 是测试用例里使用的 MySQL DSN。
	MySqlDsn = "dev:dev@tcp(127.0.0.1:3306)/test"

	// SqlServerDsn 是测试用例里使用的 SQL Server DSN。
	SqlServerDsn = "server=127.0.0.1; database=test; user id=dev;password=dev;"
)

type schema struct {
	Create string
	Drop   string
}
