package neutron

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestFloatingIPCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection of floating ips",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id",
					"floating_ip_address",
					"floating_network_id",
					"project_id",
					"router_id",
					"status",
					"fixed_ip_address",
				}).AddRow(
					"8110ce2e-a287-4ea3-8273-65459bde329f",
					"172.24.4.100",
					"23fa64ed-bab5-4180-bc9f-7928c4562d73",
					"02e67ccf829b44438e1c0397b2444a0f",
					"bd221f59-33c9-4d67-8f9c-ccaea76b05c5",
					"ACTIVE",
					"",
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetFloatingIPs)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_floating_ip floating_ip
# TYPE openstack_neutron_floating_ip gauge
openstack_neutron_floating_ip{floating_ip_address="172.24.4.100",floating_network_id="23fa64ed-bab5-4180-bc9f-7928c4562d73",id="8110ce2e-a287-4ea3-8273-65459bde329f",project_id="02e67ccf829b44438e1c0397b2444a0f",router_id="bd221f59-33c9-4d67-8f9c-ccaea76b05c5",status="ACTIVE"} 1
# HELP openstack_neutron_floating_ips floating_ips
# TYPE openstack_neutron_floating_ips gauge
openstack_neutron_floating_ips 1
# HELP openstack_neutron_floating_ips_associated_not_active floating_ips_associated_not_active
# TYPE openstack_neutron_floating_ips_associated_not_active gauge
openstack_neutron_floating_ips_associated_not_active 0
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testFloatingIPCollector{NewFloatingIPCollector(db, logger)}
	})
}

type testFloatingIPCollector struct {
	*FloatingIPCollector
}

func (t *testFloatingIPCollector) Collect(ch chan<- prometheus.Metric) {
	if err := t.FloatingIPCollector.Collect(ch); err != nil {
		panic("unexpected error: " + err.Error())
	}
}
