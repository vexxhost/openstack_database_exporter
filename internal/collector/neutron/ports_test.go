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

func TestPortCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection of ports",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id",
					"mac_address",
					"device_owner",
					"status",
					"network_id",
					"admin_state_up",
					"binding_vif_type",
					"fixed_ips",
				}).AddRow(
					"ac9e4be5-a27b-41b5-b4bd-cdbf832b03f1",
					"fa:16:3e:23:e5:0d",
					"network:router_interface",
					"ACTIVE",
					"667a6256-d975-49ed-bfe1-822887202213",
					1,
					"unbound",
					"10.0.0.1",
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetPorts)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_port port
# TYPE openstack_neutron_port gauge
openstack_neutron_port{admin_state_up="true",binding_vif_type="unbound",device_owner="network:router_interface",fixed_ips="10.0.0.1",id="ac9e4be5-a27b-41b5-b4bd-cdbf832b03f1",mac_address="fa:16:3e:23:e5:0d",network_id="667a6256-d975-49ed-bfe1-822887202213",status="ACTIVE"} 1
# HELP openstack_neutron_ports ports
# TYPE openstack_neutron_ports gauge
openstack_neutron_ports 1
`,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		return &testPortCollector{NewPortCollector(db, logger)}
	})
}

type testPortCollector struct {
	*PortCollector
}

func (t *testPortCollector) Collect(ch chan<- prometheus.Metric) {
	if err := t.PortCollector.Collect(ch); err != nil {
		panic("unexpected error: " + err.Error())
	}
}
