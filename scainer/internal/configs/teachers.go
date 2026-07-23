package configs

import (
	"fmt"
	"strings"
)

// ParseTeachersPasswords разбирает TEACHERS_PASSWORDS: "login:password,login2:password2".
// Пароль может содержать ":" — разделитель только первый в паре.
func ParseTeachersPasswords(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	out := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		login, password, ok := strings.Cut(pair, ":")
		if !ok {
			return nil, fmt.Errorf("teachers passwords: ожидается login:password, получено %q", pair)
		}
		login = strings.TrimSpace(login)
		password = strings.TrimSpace(password)
		if login == "" {
			return nil, fmt.Errorf("teachers passwords: пустой login в %q", pair)
		}
		if password == "" {
			return nil, fmt.Errorf("teachers passwords: пустой пароль для %q", login)
		}
		if _, exists := out[login]; exists {
			return nil, fmt.Errorf("teachers passwords: дублирующийся login %q", login)
		}
		out[login] = password
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("teachers passwords: пустой список")
	}
	return out, nil
}
