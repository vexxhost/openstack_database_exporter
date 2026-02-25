package testutil

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

// containerName returns a unique name for a test container, keeping the service
// name for easy identification via `docker ps` while appending a short random
// suffix to avoid collisions when go test runs packages in parallel.
func containerName(serviceName string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("osdbe-test-%s-%x", serviceName, b)
}

// prepareInitScripts copies schema files to a temp dir with numbered prefixes
// to guarantee execution order (MariaDB runs init scripts alphabetically).
func prepareInitScripts(t *testing.T, schemaFiles []string) []string {
	t.Helper()

	tmpDir := t.TempDir()
	var initScripts []string
	for i, sf := range schemaFiles {
		absPath, err := filepath.Abs(sf)
		if err != nil {
			t.Fatalf("failed to resolve schema path %s: %v", sf, err)
		}
		content, err := os.ReadFile(absPath)
		if err != nil {
			t.Fatalf("failed to read schema file %s: %v", absPath, err)
		}
		base := filepath.Base(sf)
		orderedName := fmt.Sprintf("%02d_%s", i, base)
		orderedPath := filepath.Join(tmpDir, orderedName)
		if err := os.WriteFile(orderedPath, content, 0644); err != nil {
			t.Fatalf("failed to write ordered schema file %s: %v", orderedPath, err)
		}
		initScripts = append(initScripts, orderedPath)
	}
	return initScripts
}

// NewMySQLContainer starts a named MariaDB container for the given service,
// applies the given schema files, and returns a *sql.DB connection.
//
// The container is named "osdbe-test-<serviceName>" so it's easy to identify
// via `docker ps`. The container is cleaned up automatically when the test ends.
func NewMySQLContainer(t *testing.T, serviceName string, schemaFiles ...string) *sql.DB {
	t.Helper()

	ctx := context.Background()
	initScripts := prepareInitScripts(t, schemaFiles)
	name := containerName(serviceName)

	container, err := mariadb.Run(ctx,
		"mariadb:11",
		mariadb.WithDatabase("testdb"),
		mariadb.WithUsername("test"),
		mariadb.WithPassword("test"),
		mariadb.WithScripts(initScripts...),
		testcontainers.WithName(name),
	)
	if err != nil {
		t.Fatalf("failed to start MariaDB container %s: %v", name, err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate MariaDB container %s: %v", name, err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Wait for the database to be ready
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		if err := db.PingContext(ctx); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatalf("timed out waiting for MariaDB to be ready")
		case <-time.After(500 * time.Millisecond):
		}
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// MySQLContainerResult holds a database connection and oslo.db-format URL
// returned by NewMySQLContainerWithURL.
type MySQLContainerResult struct {
	DB        *sql.DB
	URL       string // oslo.db format, e.g. mysql+pymysql://test:test@host:port/testdb
	container testcontainers.Container
}

// Terminate stops and removes the underlying MariaDB container. This is useful
// for testing resilience when a database becomes unavailable.
func (r MySQLContainerResult) Terminate(ctx context.Context) error {
	return testcontainers.TerminateContainer(r.container)
}

// NewMySQLContainerWithURL is like NewMySQLContainer but also returns an
// oslo.db-format connection URL suitable for passing to collector.Config.
func NewMySQLContainerWithURL(t *testing.T, serviceName string, schemaFiles ...string) MySQLContainerResult {
	t.Helper()

	ctx := context.Background()
	initScripts := prepareInitScripts(t, schemaFiles)
	name := containerName(serviceName)

	container, err := mariadb.Run(ctx,
		"mariadb:11",
		mariadb.WithDatabase("testdb"),
		mariadb.WithUsername("test"),
		mariadb.WithPassword("test"),
		mariadb.WithScripts(initScripts...),
		testcontainers.WithName(name),
	)
	if err != nil {
		t.Fatalf("failed to start MariaDB container %s: %v", name, err)
	}

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate MariaDB container %s: %v", name, err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Wait for the database to be ready
	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		if err := db.PingContext(ctx2); err == nil {
			break
		}
		select {
		case <-ctx2.Done():
			t.Fatalf("timed out waiting for MariaDB to be ready")
		case <-time.After(500 * time.Millisecond):
		}
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	// Build oslo.db-format URL from container
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}
	mappedPort, err := container.MappedPort(ctx, "3306/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}
	osloURL := fmt.Sprintf("mysql+pymysql://test:test@%s:%s/testdb", host, mappedPort.Port())

	return MySQLContainerResult{DB: db, URL: osloURL, container: container}
}

// SeedSQL executes raw SQL statements against the database.
// Multiple statements can be separated by semicolons.
func SeedSQL(t *testing.T, db *sql.DB, statements ...string) {
	t.Helper()

	for _, stmt := range statements {
		for _, s := range splitStatements(stmt) {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			if _, err := db.Exec(s); err != nil {
				t.Fatalf("failed to execute seed SQL %q: %v", truncate(s, 100), err)
			}
		}
	}
}

// SkipIfNoDocker skips the test if Docker is not available.
func SkipIfNoDocker(t *testing.T) {
	t.Helper()

	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("skipping integration test: SKIP_INTEGRATION is set")
	}
}

func splitStatements(sql string) []string {
	return strings.Split(sql, ";")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
