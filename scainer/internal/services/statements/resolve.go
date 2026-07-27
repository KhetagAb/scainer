package statements

// ResolveStatementSource exported for tests.
func ResolveStatementSource(c Contest) string {
	return resolveStatementSource(c)
}

// TODO(statements): заменить эвристику на явный statement source у контеста
// или расширенный резолв (CF lesson на lksh, contest без parallel и т.д.).
func resolveStatementSource(c Contest) string {
	if c.ParallelID != "" {
		return SourceLksh
	}
	return ""
}
