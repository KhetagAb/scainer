// Package fs — атомарная запись и санитизация имён для персистентных раскладок на диске
// (store.FS и workspace JPlag).
package fs

import (
	"os"
	"strings"
)

// WriteFileAtomic: tmp рядом с path + rename — падение посреди записи не оставит обрезанный файл.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// SanitizeFileName заменяет ":" и "/" — встречаются в ID вроде "ejudge:50501:12".
func SanitizeFileName(s string) string {
	r := strings.NewReplacer(":", "_", "/", "_")
	return r.Replace(s)
}
