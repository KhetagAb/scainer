package repository

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// URIFromEnv — для тестов/CLI
func URIFromEnv() (string, error) {
	user := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
	pass := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")
	host := strings.TrimSpace(os.Getenv("MONGODB_HOST"))
	return BuildURI(user, pass, host)
}

func BuildURI(user, pass, host string) (string, error) {
	user = strings.TrimSpace(user)
	host = strings.TrimSpace(host)
	if user == "" || pass == "" || host == "" {
		return "", fmt.Errorf("нужны MONGO_INITDB_ROOT_USERNAME, MONGO_INITDB_ROOT_PASSWORD и MONGODB_HOST")
	}
	if !strings.Contains(host, ":") {
		host += ":27017"
	}
	u := &url.URL{
		Scheme:   "mongodb",
		User:     url.UserPassword(user, pass),
		Host:     host,
		Path:     "/",
		RawQuery: "authSource=admin",
	}
	return u.String(), nil
}
