package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/doublecheck"
	"github.com/jackc/pgx"
	"github.com/spf13/cobra"
)

const VERSION = "0.0.1"

type Config struct {
	ConnPoolConfig pgx.ConnPoolConfig
	Schema         string
}

var cliOptions struct {
	host     string
	port     uint16
	user     string
	password string
	database string
	schema   string
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

	rootCmd := &cobra.Command{Use: "doublecheck", Short: "doublecheck - data validator"}
	rootCmd.AddCommand(cmdVersion)
	rootCmd.AddCommand(cmdList)
	rootCmd.AddCommand(cmdCheck)
	rootCmd.Execute()
}

func extractConfig() (pgx.ConnConfig, error) {
	config, err := pgx.ParseEnvLibpq()
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
		config.Database = config.User
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

	pool, err := pgx.NewConnPool(config.ConnPoolConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to PostgreSQL:\n  %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	dc, err := doublecheck.New(&doublecheck.Config{ConnPool: pool, SchemaName: config.Schema})
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

	pool, err := pgx.NewConnPool(config.ConnPoolConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to PostgreSQL:\n  %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	dc, err := doublecheck.New(&doublecheck.Config{ConnPool: pool, SchemaName: config.Schema})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to initialize doublecheck:\n  %v\n", err)
		os.Exit(1)
	}

	result, err := dc.Check(dc.Views())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Check failed:\n  %v\n", err)
		os.Exit(1)
	}

	formatter := doublecheck.NewJSONFormatter(os.Stdout)
	err = formatter.Format(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format results:\n  %v\n", err)
		os.Exit(1)
	}
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	if connConfig, err := extractConfig(); err == nil {
		config.ConnPoolConfig.ConnConfig = connConfig
	} else {
		return nil, err
	}

	appendConfigFromCLIArgs(config)

	return config, nil
}

func appendConfigFromCLIArgs(config *Config) {
	if cliOptions.host != "" {
		config.ConnPoolConfig.Host = cliOptions.host
	}
	if cliOptions.port != 0 {
		config.ConnPoolConfig.Port = cliOptions.port
	}
	if cliOptions.database != "" {
		config.ConnPoolConfig.Database = cliOptions.database
	}
	if cliOptions.user != "" {
		config.ConnPoolConfig.User = cliOptions.user
	}
	if cliOptions.password != "" {
		config.ConnPoolConfig.Password = cliOptions.password
	}
}
