package cinder

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestLimitsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with quota limits",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"project_id", "resource", "hard_limit", "in_use",
				})

				for id := range FAKE_TENANTS {
					rows.AddRow(id, "gigabytes", 1000, 0)
					rows.AddRow(id, "backup_gigabytes", 1000, 0)
				}

				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_cinder_limits_backup_max_gb limits_backup_max_gb
# TYPE openstack_cinder_limits_backup_max_gb gauge
openstack_cinder_limits_backup_max_gb{tenant="admin",tenant_id="0c4e939acacf4376bdcd1129f1a054ad"} 1000
openstack_cinder_limits_backup_max_gb{tenant="alt_demo",tenant_id="fdb8424c4e4f4c0ba32c52e2de3bd80e"} 1000
openstack_cinder_limits_backup_max_gb{tenant="demo",tenant_id="0cbd49cbf76d405d9c86562e1d579bd3"} 1000
openstack_cinder_limits_backup_max_gb{tenant="invisible_to_admin",tenant_id="5961c443439d4fcebe42643723755e9d"} 1000
openstack_cinder_limits_backup_max_gb{tenant="service",tenant_id="3d594eb0f04741069dbbb521635b21c7"} 1000
openstack_cinder_limits_backup_max_gb{tenant="swifttenanttest1",tenant_id="43ebde53fc314b1c9ea2b8c5dc744927"} 1000
openstack_cinder_limits_backup_max_gb{tenant="swifttenanttest2",tenant_id="2db68fed84324f29bb73130c6c2094fb"} 1000
openstack_cinder_limits_backup_max_gb{tenant="swifttenanttest4",tenant_id="4b1eb781a47440acb8af9850103e537f"} 1000
# HELP openstack_cinder_limits_backup_used_gb limits_backup_used_gb
# TYPE openstack_cinder_limits_backup_used_gb gauge
openstack_cinder_limits_backup_used_gb{tenant="admin",tenant_id="0c4e939acacf4376bdcd1129f1a054ad"} 0
openstack_cinder_limits_backup_used_gb{tenant="alt_demo",tenant_id="fdb8424c4e4f4c0ba32c52e2de3bd80e"} 0
openstack_cinder_limits_backup_used_gb{tenant="demo",tenant_id="0cbd49cbf76d405d9c86562e1d579bd3"} 0
openstack_cinder_limits_backup_used_gb{tenant="invisible_to_admin",tenant_id="5961c443439d4fcebe42643723755e9d"} 0
openstack_cinder_limits_backup_used_gb{tenant="service",tenant_id="3d594eb0f04741069dbbb521635b21c7"} 0
openstack_cinder_limits_backup_used_gb{tenant="swifttenanttest1",tenant_id="43ebde53fc314b1c9ea2b8c5dc744927"} 0
openstack_cinder_limits_backup_used_gb{tenant="swifttenanttest2",tenant_id="2db68fed84324f29bb73130c6c2094fb"} 0
openstack_cinder_limits_backup_used_gb{tenant="swifttenanttest4",tenant_id="4b1eb781a47440acb8af9850103e537f"} 0
# HELP openstack_cinder_limits_volume_max_gb limits_volume_max_gb
# TYPE openstack_cinder_limits_volume_max_gb gauge
openstack_cinder_limits_volume_max_gb{tenant="admin",tenant_id="0c4e939acacf4376bdcd1129f1a054ad"} 1000
openstack_cinder_limits_volume_max_gb{tenant="alt_demo",tenant_id="fdb8424c4e4f4c0ba32c52e2de3bd80e"} 1000
openstack_cinder_limits_volume_max_gb{tenant="demo",tenant_id="0cbd49cbf76d405d9c86562e1d579bd3"} 1000
openstack_cinder_limits_volume_max_gb{tenant="invisible_to_admin",tenant_id="5961c443439d4fcebe42643723755e9d"} 1000
openstack_cinder_limits_volume_max_gb{tenant="service",tenant_id="3d594eb0f04741069dbbb521635b21c7"} 1000
openstack_cinder_limits_volume_max_gb{tenant="swifttenanttest1",tenant_id="43ebde53fc314b1c9ea2b8c5dc744927"} 1000
openstack_cinder_limits_volume_max_gb{tenant="swifttenanttest2",tenant_id="2db68fed84324f29bb73130c6c2094fb"} 1000
openstack_cinder_limits_volume_max_gb{tenant="swifttenanttest4",tenant_id="4b1eb781a47440acb8af9850103e537f"} 1000
# HELP openstack_cinder_limits_volume_used_gb limits_volume_used_gb
# TYPE openstack_cinder_limits_volume_used_gb gauge
openstack_cinder_limits_volume_used_gb{tenant="admin",tenant_id="0c4e939acacf4376bdcd1129f1a054ad"} 0
openstack_cinder_limits_volume_used_gb{tenant="alt_demo",tenant_id="fdb8424c4e4f4c0ba32c52e2de3bd80e"} 0
openstack_cinder_limits_volume_used_gb{tenant="demo",tenant_id="0cbd49cbf76d405d9c86562e1d579bd3"} 0
openstack_cinder_limits_volume_used_gb{tenant="invisible_to_admin",tenant_id="5961c443439d4fcebe42643723755e9d"} 0
openstack_cinder_limits_volume_used_gb{tenant="service",tenant_id="3d594eb0f04741069dbbb521635b21c7"} 0
openstack_cinder_limits_volume_used_gb{tenant="swifttenanttest1",tenant_id="43ebde53fc314b1c9ea2b8c5dc744927"} 0
openstack_cinder_limits_volume_used_gb{tenant="swifttenanttest2",tenant_id="2db68fed84324f29bb73130c6c2094fb"} 0
openstack_cinder_limits_volume_used_gb{tenant="swifttenanttest4",tenant_id="4b1eb781a47440acb8af9850103e537f"} 0
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
			ExpectError:     true,
		},
	}

	testutil.RunCollectorTests(t, tests, NewLimitsCollector)
}
