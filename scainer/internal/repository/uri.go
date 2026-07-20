package repository

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func URIFromEnv() (string, error) {
	user := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
	pass := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")
	host := strings.TrimSpace(os.Getenv("MONGODB_HOST"))
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
