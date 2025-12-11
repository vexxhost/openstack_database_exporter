package db

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseOsloDBConnectionString converts oslo.db/SQLAlchemy-style MySQL connection
// strings to Go MySQL driver DSN format.
//
// Supported input formats:
//   - mysql://user:password@host:port/database
//   - mysql+pymysql://user:password@host:port/database
//   - mysql+mysqldb://user:password@host:port/database
//   - mysql+mysqlconnector://user:password@host:port/database
//
// Output format (Go MySQL DSN):
//   - user:password@tcp(host:port)/database?parseTime=true
func ParseOsloDBConnectionString(connectionString string) (string, error) {
	if connectionString == "" {
		return "", fmt.Errorf("connection string is empty")
	}

	// Normalize the scheme - oslo.db uses mysql+driver format
	normalized := connectionString
	for _, prefix := range []string{"mysql+pymysql://", "mysql+mysqldb://", "mysql+mysqlconnector://"} {
		if strings.HasPrefix(connectionString, prefix) {
			normalized = "mysql://" + strings.TrimPrefix(connectionString, prefix)
			break
		}
	}

	// Parse as URL
	u, err := url.Parse(normalized)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Validate scheme
	if u.Scheme != "mysql" {
		return "", fmt.Errorf("unsupported database scheme: %s (only mysql is supported)", u.Scheme)
	}

	// Extract user info
	var userPart string
	if u.User != nil {
		username := u.User.Username()
		password, hasPassword := u.User.Password()
		if hasPassword {
			userPart = fmt.Sprintf("%s:%s", username, password)
		} else {
			userPart = username
		}
	}

	// Extract host and port
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "3306"
	}

	// Extract database name (path without leading /)
	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		return "", fmt.Errorf("database name is required")
	}

	// Build Go MySQL DSN
	// Format: [user[:password]@][net[(addr)]]/dbname[?param1=value1&...]
	var dsn string
	if userPart != "" {
		dsn = fmt.Sprintf("%s@tcp(%s:%s)/%s", userPart, host, port, database)
	} else {
		dsn = fmt.Sprintf("tcp(%s:%s)/%s", host, port, database)
	}

	// Preserve query parameters and add parseTime=true if not present
	queryParams := u.Query()
	if queryParams.Get("parseTime") == "" {
		queryParams.Set("parseTime", "true")
	}

	if len(queryParams) > 0 {
		dsn = dsn + "?" + queryParams.Encode()
	}

	return dsn, nil
}
