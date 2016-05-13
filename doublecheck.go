package doublecheck

import (
	"errors"
	"github.com/jackc/pgx"
)

type Config struct {
	ConnPool   *pgx.ConnPool
	SchemaName string
}

type DoubleCheck struct {
	schemaName string
	views      []string
	pool       *pgx.ConnPool
}

func New(config *Config) (*DoubleCheck, error) {
	var dc DoubleCheck

	if config.ConnPool == nil {
		return nil, errors.New("config.ConnPool cannot be null")
	}
	dc.pool = config.ConnPool

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
	rows, err := dc.pool.Query(getViewsSQL, dc.schemaName)
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
