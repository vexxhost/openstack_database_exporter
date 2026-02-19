package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnect_EmptyString(t *testing.T) {
	_, err := Connect("")
	require.Error(t, err)
}

func TestConnect_InvalidScheme(t *testing.T) {
	_, err := Connect("postgresql://user:pass@localhost:5432/db")
	require.Error(t, err)
}

func TestConnect_UnreachableHost(t *testing.T) {
	// Valid DSN format but host is unreachable — Connect should fail on Ping
	_, err := Connect("mysql://user:pass@192.0.2.1:3306/testdb")
	require.Error(t, err)
}
