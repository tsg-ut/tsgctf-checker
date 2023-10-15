package checker

import (
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Schema of test result table.
type DbResult struct {
	Name      string     `db:"name"`
	Result    TestResult `db:"result"`
	Timestamp time.Time  `db:"timestamp"`
}

// Converter of `Challenge` into `DBResult`.
func (chall *Challenge) intoDbResult(result TestResult) DbResult {
	return DbResult{
		Name:   chall.Name,
		Result: result,
	}
}

// Connect to mysql server and returns instance.
func Connect(dbuser string, dbpass string, dbhost string, dbname string) (*sqlx.DB, error) {
	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(dbhost, "3306")
	cfg.DBName = dbname
	cfg.User = dbuser
	cfg.Passwd = dbpass
	cfg.ParseTime = true
	db, err := sqlx.Connect("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Write and commit test result.
func RecordResult(db *sqlx.DB, chall Challenge, result TestResult) error {
	tx := db.MustBegin()
	dbresult := chall.intoDbResult(result)
	dbresult.Timestamp = time.Now()
	query := "insert into test_result(name, result, timestamp) values(:name, :result, :timestamp)"
	_, err := tx.NamedExec(query, dbresult)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Query test result from DB by challenge ID.
func FetchResult(db *sqlx.DB, chall_name string, limit int) ([]DbResult, error) {
	var results []DbResult

	query := `select name, result, timestamp from test_result where name = ? order by timestamp desc limit ?`
	tx := db.MustBegin()
	if err := tx.Select(&results, query, chall_name, limit); err != nil {
		return results, err
	}
	if err := tx.Commit(); err != nil {
		return results, err
	}
	return results, nil
}
