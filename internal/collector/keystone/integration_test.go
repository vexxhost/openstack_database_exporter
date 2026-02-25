//go:build integration

package keystone

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	itest "github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestIntegration_IdentityCollector(t *testing.T) {
	itest.SkipIfNoDocker(t)

	db := itest.NewMySQLContainer(t, "keystone", "../../../sql/keystone/schema.sql")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("empty database", func(t *testing.T) {
		collector := NewIdentityCollector(db, logger)

		expected := `# HELP openstack_identity_domains domains
# TYPE openstack_identity_domains gauge
openstack_identity_domains 0
# HELP openstack_identity_groups groups
# TYPE openstack_identity_groups gauge
openstack_identity_groups 0
# HELP openstack_identity_projects projects
# TYPE openstack_identity_projects gauge
openstack_identity_projects 0
# HELP openstack_identity_regions regions
# TYPE openstack_identity_regions gauge
openstack_identity_regions 0
# HELP openstack_identity_up up
# TYPE openstack_identity_up gauge
openstack_identity_up 1
# HELP openstack_identity_users users
# TYPE openstack_identity_users gauge
openstack_identity_users 0
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("with data", func(t *testing.T) {
		// Insert a domain (is_domain=1) - must come first because of FK
		itest.SeedSQL(t, db,
			`INSERT INTO project (id, name, enabled, domain_id, is_domain) VALUES
			('domain-001', 'TestDomain', 1, 'domain-001', 1),
			('domain-002', 'default', 1, 'domain-002', 1)`,
			// Insert projects (is_domain=0) - references domain via FK
			`INSERT INTO project (id, name, enabled, domain_id, parent_id, is_domain) VALUES
			('proj-001', 'test-project-1', 1, 'domain-001', 'domain-001', 0),
			('proj-002', 'test-project-2', 0, 'domain-002', 'domain-002', 0)`,
			// Insert tags
			`INSERT INTO project_tag (project_id, name) VALUES
			('proj-001', 'env:prod'),
			('proj-001', 'team:infra')`,
			// Insert users
			`INSERT INTO user (id, enabled, domain_id, created_at) VALUES
			('user-001', 1, 'domain-001', NOW()),
			('user-002', 0, 'domain-002', NOW()),
			('user-003', 1, 'domain-001', NOW())`,
			// Insert regions
			`INSERT INTO region (id, description) VALUES
			('RegionOne', 'Primary region'),
			('RegionTwo', 'Secondary region')`,
			// Insert groups
			`INSERT INTO `+"`group`"+` (id, domain_id, name) VALUES
			('grp-001', 'domain-001', 'admins')`,
		)

		collector := NewIdentityCollector(db, logger)

		// Verify counts
		expected := `# HELP openstack_identity_domains domains
# TYPE openstack_identity_domains gauge
openstack_identity_domains 2
# HELP openstack_identity_groups groups
# TYPE openstack_identity_groups gauge
openstack_identity_groups 1
# HELP openstack_identity_projects projects
# TYPE openstack_identity_projects gauge
openstack_identity_projects 2
# HELP openstack_identity_regions regions
# TYPE openstack_identity_regions gauge
openstack_identity_regions 2
# HELP openstack_identity_up up
# TYPE openstack_identity_up gauge
openstack_identity_up 1
# HELP openstack_identity_users users
# TYPE openstack_identity_users gauge
openstack_identity_users 3
`
		err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
			"openstack_identity_domains",
			"openstack_identity_projects",
			"openstack_identity_users",
			"openstack_identity_regions",
			"openstack_identity_groups",
			"openstack_identity_up",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify domain_info and project_info are emitted
		domainInfoCount := testutil.CollectAndCount(collector, "openstack_identity_domain_info")
		if domainInfoCount != 2 {
			t.Fatalf("expected 2 domain_info metrics, got %d", domainInfoCount)
		}

		projectInfoCount := testutil.CollectAndCount(collector, "openstack_identity_project_info")
		if projectInfoCount != 2 {
			t.Fatalf("expected 2 project_info metrics, got %d", projectInfoCount)
		}
	})
}
