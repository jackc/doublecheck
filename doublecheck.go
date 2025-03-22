package doublecheck

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	Conn       *pgx.Conn
	SchemaName string
}

type DoubleCheck struct {
	schemaName string
	views      []string
	conn       *pgx.Conn
}

type CheckResult struct {
	Database      string        `json:"database"`
	Schema        string        `json:"schema"`
	User          string        `json:"user"`
	StartTime     time.Time     `json:"start_time"`
	Duration      time.Duration `json:"duration"`
	ErrorDetected bool          `json:"error_detected"`
	ViewResults   []ViewResult  `json:"view_results"`
}

type ViewResult struct {
	Name      string                   `json:"name"`
	StartTime time.Time                `json:"start_time"`
	Duration  time.Duration            `json:"duration"`
	Rows      []map[string]interface{} `json:"rows"`
}

func New(config *Config) (*DoubleCheck, error) {
	var dc DoubleCheck

	if config.Conn == nil {
		return nil, errors.New("config.Conn cannot be null")
	}
	dc.conn = config.Conn

	dc.schemaName = config.SchemaName
	if dc.schemaName == "" {
		dc.schemaName = "doublecheck"
	}

	views, err := dc.getViews()
	if err != nil {
		return nil, err
	}
	dc.views = views

	return &dc, nil
}

const getViewsSQL = `select table_name from information_schema.views where table_schema=$1 order by 1`

func (dc *DoubleCheck) getViews() ([]string, error) {
	rows, err := dc.conn.Query(context.Background(), getViewsSQL, dc.schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var viewName string
		err = rows.Scan(&viewName)
		if err != nil {
			return nil, err
		}
		views = append(views, viewName)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return views, nil
}

func (dc *DoubleCheck) SchemaName() string {
	return dc.schemaName
}

func (dc *DoubleCheck) Views() []string {
	return dc.views
}

func (dc *DoubleCheck) Check(viewNames []string) (*CheckResult, error) {
	var database, user string
	if err := dc.conn.QueryRow(context.Background(), "select current_database(), current_user").Scan(&database, &user); err != nil {
		return nil, err
	}

	cr := &CheckResult{
		Database:  database,
		Schema:    dc.SchemaName(),
		User:      user,
		StartTime: time.Now(),
	}

	cr.ViewResults = make([]ViewResult, 0, len(viewNames))
	for _, view := range viewNames {
		vr, err := dc.checkView(view)
		if err != nil {
			return nil, err
		}
		cr.ViewResults = append(cr.ViewResults, *vr)
	}

	for _, vr := range cr.ViewResults {
		if len(vr.Rows) > 0 {
			cr.ErrorDetected = true
		}
	}

	cr.Duration = time.Now().Sub(cr.StartTime)

	return cr, nil
}

func (dc *DoubleCheck) checkView(viewName string) (*ViewResult, error) {
	vr := &ViewResult{
		Name:      viewName,
		StartTime: time.Now(),
		Rows:      []map[string]interface{}{},
	}

	sql := fmt.Sprintf(
		`select row_to_json(t) from %s.%s t`,
		quoteIdentifier(dc.SchemaName()),
		quoteIdentifier(viewName),
	)

	rows, err := dc.conn.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rowJSON map[string]interface{}
		err = rows.Scan(&rowJSON)
		if err != nil {
			return nil, err
		}
		vr.Rows = append(vr.Rows, rowJSON)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	vr.Duration = time.Now().Sub(vr.StartTime)

	return vr, nil
}

func quoteIdentifier(input string) string {
	return `"` + strings.Replace(input, `"`, `""`, -1) + `"`
}
