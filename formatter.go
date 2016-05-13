package doublecheck

type Formatter interface {
	Format(result *CheckResult) error
}
