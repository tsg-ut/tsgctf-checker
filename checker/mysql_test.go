package checker

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	TEST_MYSQL_PORT = "3306"
)

type mysqlContainer struct {
	testcontainers.Container
}

// Setup MySQL container for testing.
func setupMysql(ctx context.Context) (*mysqlContainer, error) {
	initdbDir, err := filepath.Abs("../tests/assets/initdb")
	if err != nil {
		return nil, err
	}

	req := testcontainers.ContainerRequest{
		Image: "mysql:8.0",
		Env: map[string]string{
			"MYSQL_DATABASE":             "test",
			"MYSQL_USER":                 "user",
			"MYSQL_PASSWORD":             "password",
			"MYSQL_ALLOW_EMPTY_PASSWORD": "yes",
		},
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", TEST_MYSQL_PORT)},
		WaitingFor: wait.ForSQL(TEST_MYSQL_PORT, "mysql", func(host string, port nat.Port) string {
			cfg := mysql.NewConfig()
			cfg.Net = "tcp"
			cfg.Addr = net.JoinHostPort(host, port.Port())
			cfg.DBName = "test"
			cfg.User = "user"
			cfg.Passwd = "password"
			return cfg.FormatDSN()
		}),
		Mounts: testcontainers.ContainerMounts{
			testcontainers.BindMount(initdbDir, "/docker-entrypoint-initdb.d"),
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &mysqlContainer{container}, nil
}

// Open MySQL connection.
func (s *mysqlContainer) OpenDB(ctx context.Context) (*sqlx.DB, error) {
	host, err := s.Container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := s.Container.MappedPort(ctx, TEST_MYSQL_PORT)
	if err != nil {
		return nil, err
	}

	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(host, port.Port())
	cfg.DBName = "test"
	cfg.User = "user"
	cfg.Passwd = "password"
	cfg.ParseTime = true

	return sqlx.ConnectContext(ctx, "mysql", cfg.FormatDSN())
}

func count_records(db *sqlx.DB) (int, error) {
	var count int
	if err := db.Get(&count, "select count(*) from test_result"); err != nil {
		return 0, err
	}
	return count, nil
}

func TestMysql_MysqlOperations(t *testing.T) {
	ctx := context.Background()
	container, err := setupMysql(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer container.Terminate(ctx)

	db, _ := container.OpenDB(ctx)

	// record results
	chall := Challenge{
		Name:    "test",
		Timeout: 5,
	}
	for i := 0; i < 3; i++ {
		if err := RecordResult(db, chall, ResultSuccess); err != nil {
			t.Fatal(err)
		}
	}

	count, err := count_records(db)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}

	// fetch results
	results, err := FetchResult(db, chall.Name, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Result != ResultSuccess {
		t.Errorf("results[0].Result = %v, want %v", results[0].Result, ResultSuccess)
	}
	if results[0].Name != chall.Name {
		t.Errorf("results[0].Name = %v, want %v", results[0].Name, chall.Name)
	}
}
