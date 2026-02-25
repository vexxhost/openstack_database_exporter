package neutron

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestPortCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with ports",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "mac_address", "device_owner", "status",
					"network_id", "admin_state_up", "ip_allocation",
					"binding_vif_type", "fixed_ips",
				}).AddRow(
					"883f060a-60a2-48af-aba8-88c45a4b0b58",
					"fa:16:3e:2d:97:08",
					"compute:nova",
					"ACTIVE",
					"74917853-7529-46fc-8545-ed70fe691f03",
					true,
					nil,
					"ovs",
					[]byte("10.13.18.143"),
				).AddRow(
					"10e61c4b-cefc-4a38-a374-bf241d9411b5",
					"fa:16:3e:9d:fa:55",
					"Octavia",
					"DOWN",
					"74917853-7529-46fc-8545-ed70fe691f03",
					false,
					nil,
					"unbound",
					[]byte("10.16.0.90"),
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetPorts)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_port port
# TYPE openstack_neutron_port gauge
openstack_neutron_port{admin_state_up="true",binding_vif_type="ovs",device_owner="compute:nova",fixed_ips="10.13.18.143",mac_address="fa:16:3e:2d:97:08",network_id="74917853-7529-46fc-8545-ed70fe691f03",status="ACTIVE",uuid="883f060a-60a2-48af-aba8-88c45a4b0b58"} 1
openstack_neutron_port{admin_state_up="false",binding_vif_type="unbound",device_owner="Octavia",fixed_ips="10.16.0.90",mac_address="fa:16:3e:9d:fa:55",network_id="74917853-7529-46fc-8545-ed70fe691f03",status="DOWN",uuid="10e61c4b-cefc-4a38-a374-bf241d9411b5"} 1
# HELP openstack_neutron_ports ports
# TYPE openstack_neutron_ports gauge
openstack_neutron_ports 2
# HELP openstack_neutron_ports_lb_not_active ports_lb_not_active
# TYPE openstack_neutron_ports_lb_not_active gauge
openstack_neutron_ports_lb_not_active 0
# HELP openstack_neutron_ports_no_ips ports_no_ips
# TYPE openstack_neutron_ports_no_ips gauge
openstack_neutron_ports_no_ips 0
`,
		},
		{
			Name: "LB port not active and port with no IPs",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "mac_address", "device_owner", "status",
					"network_id", "admin_state_up", "ip_allocation",
					"binding_vif_type", "fixed_ips",
				}).AddRow(
					"port-1", "aa:bb:cc:dd:ee:ff",
					"neutron:LOADBALANCERV2", "DOWN",
					"net-1", true, nil, "ovs", []byte("10.0.0.1"),
				).AddRow(
					"port-2", "11:22:33:44:55:66",
					"", "DOWN",
					"net-1", true, nil, "unbound", []byte(""),
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetPorts)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_port port
# TYPE openstack_neutron_port gauge
openstack_neutron_port{admin_state_up="true",binding_vif_type="ovs",device_owner="neutron:LOADBALANCERV2",fixed_ips="10.0.0.1",mac_address="aa:bb:cc:dd:ee:ff",network_id="net-1",status="DOWN",uuid="port-1"} 1
openstack_neutron_port{admin_state_up="true",binding_vif_type="unbound",device_owner="",fixed_ips="",mac_address="11:22:33:44:55:66",network_id="net-1",status="DOWN",uuid="port-2"} 1
# HELP openstack_neutron_ports ports
# TYPE openstack_neutron_ports gauge
openstack_neutron_ports 2
# HELP openstack_neutron_ports_lb_not_active ports_lb_not_active
# TYPE openstack_neutron_ports_lb_not_active gauge
openstack_neutron_ports_lb_not_active 1
# HELP openstack_neutron_ports_no_ips ports_no_ips
# TYPE openstack_neutron_ports_no_ips gauge
openstack_neutron_ports_no_ips 1
`,
		},
		{
			Name: "port with ip_allocation none excluded from no_ips count",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "mac_address", "device_owner", "status",
					"network_id", "admin_state_up", "ip_allocation",
					"binding_vif_type", "fixed_ips",
				}).AddRow(
					"port-1", "aa:bb:cc:dd:ee:ff",
					"network:distributed", "DOWN",
					"net-1", true, "none", "unbound", []byte(""),
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetPorts)).WillReturnRows(rows)
			},
			ExpectedMetrics: `# HELP openstack_neutron_port port
# TYPE openstack_neutron_port gauge
openstack_neutron_port{admin_state_up="true",binding_vif_type="unbound",device_owner="network:distributed",fixed_ips="",mac_address="aa:bb:cc:dd:ee:ff",network_id="net-1",status="DOWN",uuid="port-1"} 1
# HELP openstack_neutron_ports ports
# TYPE openstack_neutron_ports gauge
openstack_neutron_ports 1
# HELP openstack_neutron_ports_lb_not_active ports_lb_not_active
# TYPE openstack_neutron_ports_lb_not_active gauge
openstack_neutron_ports_lb_not_active 0
# HELP openstack_neutron_ports_no_ips ports_no_ips
# TYPE openstack_neutron_ports_no_ips gauge
openstack_neutron_ports_no_ips 0
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetPorts)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
		},
	}

	testutil.RunCollectorTests(t, tests, NewPortCollector)
}
