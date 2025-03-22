package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/doublecheck"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
)

const VERSION = "0.3.0"

type Config struct {
	ConnConfig *pgx.ConnConfig
	Schema     string
	Quiet      bool
	Format     string
}

var cliOptions struct {
	host     string
	port     uint16
	user     string
	password string
	database string
	schema   string

	quiet  bool
	format string
}

func main() {
	cmdVersion := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("doublecheck v%s\n", VERSION)
		},
	}

	cmdList := &cobra.Command{
		Use:   "list",
		Short: "Print views doublecheck will use",
		Run:   List,
	}
	addConfigFlagsToCommand(cmdList)

	cmdCheck := &cobra.Command{
		Use:   "check",
		Short: "Checks doublecheck views for errors",
		Run:   Check,
	}
	addConfigFlagsToCommand(cmdCheck)
	cmdCheck.Flags().StringVarP(&cliOptions.format, "format", "", "", "format (text or json)")
	cmdCheck.Flags().BoolVarP(&cliOptions.quiet, "quiet", "", false, "only print output if error found")

	rootCmd := &cobra.Command{Use: "doublecheck", Short: "doublecheck - data validator"}
	rootCmd.AddCommand(cmdVersion)
	rootCmd.AddCommand(cmdList)
	rootCmd.AddCommand(cmdCheck)
	rootCmd.Execute()
}

func addConfigFlagsToCommand(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cliOptions.host, "host", "", "", "database host")
	cmd.Flags().Uint16VarP(&cliOptions.port, "port", "", 0, "database port")
	cmd.Flags().StringVarP(&cliOptions.user, "user", "", "", "database user")
	cmd.Flags().StringVarP(&cliOptions.password, "password", "", "", "database password")
	cmd.Flags().StringVarP(&cliOptions.database, "database", "", "", "database name")
	cmd.Flags().StringVarP(&cliOptions.schema, "schema", "", "doublecheck", "schema that contains doublecheck views")
}

func List(cmd *cobra.Command, args []string) {
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config:\n  %v\n", err)
		os.Exit(1)
	}

	conn, err := pgx.ConnectConfig(context.Background(), config.ConnConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to PostgreSQL:\n  %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	dc, err := doublecheck.New(&doublecheck.Config{Conn: conn, SchemaName: config.Schema})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize doublecheck:\n  %v\n", err)
		os.Exit(1)
	}

	for _, view := range dc.Views() {
		fmt.Println(view)
	}
}

func Check(cmd *cobra.Command, args []string) {
	config, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config:\n  %v\n", err)
		os.Exit(1)
	}

	conn, err := pgx.ConnectConfig(context.Background(), config.ConnConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to PostgreSQL:\n  %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	dc, err := doublecheck.New(&doublecheck.Config{Conn: conn, SchemaName: config.Schema})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize doublecheck:\n  %v\n", err)
		os.Exit(1)
	}

	result, err := dc.Check(dc.Views())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Check failed:\n  %v\n", err)
		os.Exit(1)
	}

	if config.Quiet && !result.ErrorDetected {
		return
	}

	var formatter doublecheck.Formatter
	switch config.Format {
	case "text":
		formatter = doublecheck.NewTextFormatter(os.Stdout)
	case "json":
		formatter = doublecheck.NewJSONFormatter(os.Stdout)
	}

	err = formatter.Format(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format results:\n  %v\n", err)
		os.Exit(1)
	}
}

func LoadConfig() (*Config, error) {
	config := &Config{Format: "json"}
	if connConfig, err := pgx.ParseConfig(""); err == nil {
		config.ConnConfig = connConfig
	} else {
		return nil, err
	}

	appendConfigFromCLIArgs(config)

	switch config.Format {
	case "text", "json":
	default:
		return nil, errors.New("invalid format")
	}

	return config, nil
}

func appendConfigFromCLIArgs(config *Config) {
	if cliOptions.host != "" {
		config.ConnConfig.Host = cliOptions.host
	}
	if cliOptions.port != 0 {
		config.ConnConfig.Port = cliOptions.port
	}
	if cliOptions.database != "" {
		config.ConnConfig.Database = cliOptions.database
	}
	if cliOptions.user != "" {
		config.ConnConfig.User = cliOptions.user
	}
	if cliOptions.password != "" {
		config.ConnConfig.Password = cliOptions.password
	}

	config.Quiet = cliOptions.quiet

	if cliOptions.format != "" {
		config.Format = cliOptions.format
	}
}
