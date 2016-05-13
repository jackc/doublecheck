package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx"
)

func main() {
	config, err := extractConfig()
	if err != nil {
		log.Fatalln(err)
	}

	pool, err := pgx.NewConnPool(config)
	if err != nil {
		log.Fatalln(err)
	}

	rows, err := pool.Query("select * from test_pgx() as t(a);")

	if err != nil {
		fmt.Printf(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		fmt.Printf("doesn't get here")

		var n *int32
		err = rows.Scan(&n)
		if err != nil {
			fmt.Printf(err.Error())
		} else {
			fmt.Printf("%d\n", *n)
		}
	}

	if rows.Err() != nil {
		fmt.Println(rows.Err())
	}

}

func extractConfig() (config pgx.ConnPoolConfig, err error) {
	config.ConnConfig, err = pgx.ParseEnvLibpq()
	if err != nil {
		return config, err
	}

	if config.Host == "" {
		config.Host = "localhost"
	}

	if config.User == "" {
		config.User = os.Getenv("USER")
	}

	if config.Database == "" {
		config.Database = config.User
	}

	config.MaxConnections = 10

	return config, nil
}

type Validator struct {
	schemaName      string
	validationViews []string
	pool            *pgx.ConnPool
}

func NewValidator(schemaName string, pool *pgx.ConnPool) (*Validator, error) {

}

func (v *Validator) SchemaName() string {
	return schemaName
}
