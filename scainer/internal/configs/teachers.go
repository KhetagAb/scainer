package configs

import (
	"fmt"
	"strings"
)

// ParseTeachersLogins разбирает TEACHERS_LOGINS: "alice;bob".
func ParseTeachersLogins(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var out []string
	seen := make(map[string]bool)
	for _, part := range strings.Split(raw, ";") {
		login := strings.TrimSpace(part)
		if login == "" {
			continue
		}
		if seen[login] {
			return nil, fmt.Errorf("teachers logins: дублирующийся login %q", login)
		}
		seen[login] = true
		out = append(out, login)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("teachers logins: пустой список")
	}
	return out, nil
}
