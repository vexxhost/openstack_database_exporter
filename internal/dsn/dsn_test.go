package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOsloDBConnectionString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "basic mysql URL",
			input:    "mysql://user:password@localhost:3306/database",
			expected: "user:password@tcp(localhost:3306)/database?parseTime=true",
		},
		{
			name:     "mysql+pymysql URL",
			input:    "mysql+pymysql://user:password@localhost:3306/database",
			expected: "user:password@tcp(localhost:3306)/database?parseTime=true",
		},
		{
			name:     "mysql+mysqldb URL",
			input:    "mysql+mysqldb://user:password@localhost:3306/database",
			expected: "user:password@tcp(localhost:3306)/database?parseTime=true",
		},
		{
			name:     "mysql+mysqlconnector URL",
			input:    "mysql+mysqlconnector://user:password@localhost:3306/database",
			expected: "user:password@tcp(localhost:3306)/database?parseTime=true",
		},
		{
			name:     "URL without port defaults to 3306",
			input:    "mysql://user:password@localhost/database",
			expected: "user:password@tcp(localhost:3306)/database?parseTime=true",
		},
		{
			name:     "URL with special characters in password",
			input:    "mysql://user:p%40ssw0rd@localhost:3306/database",
			expected: "user:p@ssw0rd@tcp(localhost:3306)/database?parseTime=true",
		},
		{
			name:     "URL without password",
			input:    "mysql://user@localhost:3306/database",
			expected: "user@tcp(localhost:3306)/database?parseTime=true",
		},
		{
			name:     "URL with query parameters",
			input:    "mysql://user:password@localhost:3306/database?charset=utf8mb4",
			expected: "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=true",
		},
		{
			name:     "URL with parseTime already set",
			input:    "mysql://user:password@localhost:3306/database?parseTime=false",
			expected: "user:password@tcp(localhost:3306)/database?parseTime=false",
		},
		{
			name:     "URL with remote host",
			input:    "mysql://nova:secret@db.example.com:3306/nova",
			expected: "nova:secret@tcp(db.example.com:3306)/nova?parseTime=true",
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "unsupported scheme",
			input:       "postgresql://user:password@localhost:5432/database",
			expectError: true,
		},
		{
			name:        "missing database name",
			input:       "mysql://user:password@localhost:3306/",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseOsloDBConnectionString(tt.input)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
