package neutron

import (
	"database/sql"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

func TestQuotaCollector(t *testing.T) {
	type quotaTestCase struct {
		Name            string
		SetupMock       func(sqlmock.Sqlmock)
		ExpectedMetrics string
	}

	tests := []quotaTestCase{
		{
			Name: "single project with explicit quotas and usage",
			SetupMock: func(mock sqlmock.Sqlmock) {
				quotaRows := sqlmock.NewRows([]string{"project_id", "resource", "limit"}).
					AddRow("proj-1", "security_group", -1).
					AddRow("proj-1", "security_group_rule", -1)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetQuotas)).WillReturnRows(quotaRows)

				usageRows := sqlmock.NewRows([]string{"project_id", "resource", "cnt"}).
					AddRow("proj-1", "floatingip", 2).
					AddRow("proj-1", "network", 1).
					AddRow("proj-1", "port", 5).
					AddRow("proj-1", "router", 1).
					AddRow("proj-1", "security_group", 4).
					AddRow("proj-1", "security_group_rule", 26).
					AddRow("proj-1", "subnet", 3)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetResourceCountsByProject)).WillReturnRows(usageRows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_quota_floatingip quota_floatingip
# TYPE openstack_neutron_quota_floatingip gauge
openstack_neutron_quota_floatingip{tenant="proj-1",type="limit"} 50
openstack_neutron_quota_floatingip{tenant="proj-1",type="reserved"} 0
openstack_neutron_quota_floatingip{tenant="proj-1",type="used"} 2
# HELP openstack_neutron_quota_network quota_network
# TYPE openstack_neutron_quota_network gauge
openstack_neutron_quota_network{tenant="proj-1",type="limit"} 100
openstack_neutron_quota_network{tenant="proj-1",type="reserved"} 0
openstack_neutron_quota_network{tenant="proj-1",type="used"} 1
# HELP openstack_neutron_quota_port quota_port
# TYPE openstack_neutron_quota_port gauge
openstack_neutron_quota_port{tenant="proj-1",type="limit"} 500
openstack_neutron_quota_port{tenant="proj-1",type="reserved"} 0
openstack_neutron_quota_port{tenant="proj-1",type="used"} 5
# HELP openstack_neutron_quota_rbac_policy quota_rbac_policy
# TYPE openstack_neutron_quota_rbac_policy gauge
openstack_neutron_quota_rbac_policy{tenant="proj-1",type="limit"} 10
openstack_neutron_quota_rbac_policy{tenant="proj-1",type="reserved"} 0
openstack_neutron_quota_rbac_policy{tenant="proj-1",type="used"} 0
# HELP openstack_neutron_quota_router quota_router
# TYPE openstack_neutron_quota_router gauge
openstack_neutron_quota_router{tenant="proj-1",type="limit"} 10
openstack_neutron_quota_router{tenant="proj-1",type="reserved"} 0
openstack_neutron_quota_router{tenant="proj-1",type="used"} 1
# HELP openstack_neutron_quota_security_group quota_security_group
# TYPE openstack_neutron_quota_security_group gauge
openstack_neutron_quota_security_group{tenant="proj-1",type="limit"} -1
openstack_neutron_quota_security_group{tenant="proj-1",type="reserved"} 0
openstack_neutron_quota_security_group{tenant="proj-1",type="used"} 4
# HELP openstack_neutron_quota_security_group_rule quota_security_group_rule
# TYPE openstack_neutron_quota_security_group_rule gauge
openstack_neutron_quota_security_group_rule{tenant="proj-1",type="limit"} -1
openstack_neutron_quota_security_group_rule{tenant="proj-1",type="reserved"} 0
openstack_neutron_quota_security_group_rule{tenant="proj-1",type="used"} 26
# HELP openstack_neutron_quota_subnet quota_subnet
# TYPE openstack_neutron_quota_subnet gauge
openstack_neutron_quota_subnet{tenant="proj-1",type="limit"} 100
openstack_neutron_quota_subnet{tenant="proj-1",type="reserved"} 0
openstack_neutron_quota_subnet{tenant="proj-1",type="used"} 3
# HELP openstack_neutron_quota_subnetpool quota_subnetpool
# TYPE openstack_neutron_quota_subnetpool gauge
openstack_neutron_quota_subnetpool{tenant="proj-1",type="limit"} -1
openstack_neutron_quota_subnetpool{tenant="proj-1",type="reserved"} 0
openstack_neutron_quota_subnetpool{tenant="proj-1",type="used"} 0
`,
		},
		{
			Name: "empty results - no projects",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetQuotas)).WillReturnRows(
					sqlmock.NewRows([]string{"project_id", "resource", "limit"}))
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetResourceCountsByProject)).WillReturnRows(
					sqlmock.NewRows([]string{"project_id", "resource", "cnt"}))
			},
			ExpectedMetrics: "",
		},
		{
			Name: "query error on quotas",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetQuotas)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			tt.SetupMock(mock)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			resolver := project.NewResolver(logger, nil, 0)
			collector := NewQuotaCollector(db, logger, resolver)

			if tt.ExpectedMetrics != "" {
				err = testutil.CollectAndCompare(collector, strings.NewReader(tt.ExpectedMetrics))
				assert.NoError(t, err)
			} else {
				problems, err := testutil.CollectAndLint(collector)
				assert.Len(t, problems, 0)
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
