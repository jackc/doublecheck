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
	possiblePaths := []string{
		"/tmp",                // Standard location and homebrew
		"/var/run/postgresql", // Debian / Ubuntu
	}

	for _, path := range possiblePaths {
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

	expectedViews := []string{"syntax error", "with_multiple_errors", "without_errors"}
	if !reflect.DeepEqual(dc.Views(), expectedViews) {
		t.Fatalf("Expected Views() to return %v, but got %v", expectedViews, dc.Views())
	}
}

func TestDoubleCheckCheck(t *testing.T) {
	dc, err := doublecheck.New(&doublecheck.Config{ConnPool: pool})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		testName string
		viewName string
		rows     []map[string]interface{}
	}{
		{
			testName: "Special character view name",
			viewName: "syntax error",
			rows:     []map[string]interface{}{},
		},
		{
			testName: "No errors",
			viewName: "without_errors",
			rows:     []map[string]interface{}{},
		},
		{
			testName: "With multiple errors",
			viewName: "with_multiple_errors",
			rows: []map[string]interface{}{
				map[string]interface{}{"id": float64(7), "error_message": "something went wrong"},
				map[string]interface{}{"id": float64(42), "error_message": "something else went wrong"},
			},
		},
	}

	for i, tt := range tests {
		result, err := dc.Check(tt.viewName)
		if err != nil {
			t.Errorf(`%d. %s: %v`, i, tt.testName, err)
			continue
		}
		if result.ViewName != tt.viewName {
			t.Errorf(`%d. %s: Expected result.ViewName to be "%s", but it was "%s"`, i, tt.testName, tt.viewName, result.ViewName)
		}
		if !reflect.DeepEqual(result.Rows, tt.rows) {
			t.Errorf(`%d. %s: Expected result.Rows to be %#v, but it was %#v`, i, tt.testName, tt.rows, result.Rows)
		}
	}

}
