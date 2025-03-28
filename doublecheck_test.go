package doublecheck_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/doublecheck"
	"github.com/jackc/pgx/v5"
)

var (
	conn *pgx.Conn
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
	config, err := pgx.ParseConfig("dbname=doublecheck_test")
	if err != nil {
		return err
	}

	conn, err = pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		return err
	}

	return nil
}

func TestDoubleCheckViews(t *testing.T) {
	dc, err := doublecheck.New(&doublecheck.Config{Conn: conn})
	if err != nil {
		t.Fatal(err)
	}

	expectedViews := []string{"syntax error", "with_multiple_errors", "without_errors"}
	if !reflect.DeepEqual(dc.Views(), expectedViews) {
		t.Fatalf("Expected Views() to return %v, but got %v", expectedViews, dc.Views())
	}
}

func TestDoubleCheckCheck(t *testing.T) {
	dc, err := doublecheck.New(&doublecheck.Config{Conn: conn})
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
		result, err := dc.Check([]string{tt.viewName})
		if err != nil {
			t.Errorf(`%d. %s: %v`, i, tt.testName, err)
			continue
		}

		vr := result.ViewResults[0]
		if vr.Name != tt.viewName {
			t.Errorf(`%d. %s: Expected result.ViewName to be "%s", but it was "%s"`, i, tt.testName, tt.viewName, vr.Name)
		}
		if !reflect.DeepEqual(vr.Rows, tt.rows) {
			t.Errorf(`%d. %s: Expected result.Rows to be %#v, but it was %#v`, i, tt.testName, tt.rows, vr.Rows)
		}
	}
}
