package doublecheck_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jackc/doublecheck"
	"github.com/jackc/pgx"
)

var (
	pool *pgx.ConnPool
)

func TestMain(m *testing.M) {
	flag.Parse()
	err := setup()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to setup test: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func setup() error {
	config, err := extractConfig()
	if err != nil {
		return err
	}

	pool, err = pgx.NewConnPool(config)
	if err != nil {
		return err
	}

	return nil
}

func extractConfig() (config pgx.ConnPoolConfig, err error) {
	config.ConnConfig, err = pgx.ParseEnvLibpq()
	if err != nil {
		return config, err
	}

	if config.Host == "" {
		config.Host = findSocketPath()
	}
	if config.Host == "" {
		config.Host = "localhost"
	}

	if config.User == "" {
		config.User = os.Getenv("USER")
	}

	if config.Database == "" {
		config.Database = "doublecheck_test"
	}

	return config, nil
}

func findSocketPath() string {
	for _, path := range []string{"/var/run/postgresql"} {
		matches, _ := filepath.Glob(fmt.Sprintf("%s/.s.PGSQL*", path))
		if len(matches) > 0 {
			return path
		}
	}

	return ""
}

func TestDoubleCheckViews(t *testing.T) {
	dc, err := doublecheck.New(&doublecheck.Config{ConnPool: pool})
	if err != nil {
		t.Fatal(err)
	}

	expectedViews := []string{"with_multiple_errors", "without_errors"}
	if !reflect.DeepEqual(dc.Views(), expectedViews) {
		t.Fatalf("Expected Views() to return %v, but got %v", expectedViews, dc.Views())
	}
}
