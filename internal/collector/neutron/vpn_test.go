package neutron

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestVPNCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with vpn data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				serviceRows := sqlmock.NewRows([]string{
					"id", "project_id", "subnet_id", "router_id",
					"admin_state_up", "name", "external_v4_ip",
					"external_v6_ip", "flavor_id", "status",
				}).AddRow(
					"vpn-svc-1", "proj-1", "subnet-1", "router-1",
					true, "my-vpn", "1.2.3.4", "", "", "ACTIVE",
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnServices)).WillReturnRows(serviceRows)

				connRows := sqlmock.NewRows([]string{
					"id", "project_id", "admin_state_up", "name",
					"vpnservice_id", "ikepolicy_id", "ipsecpolicy_id",
					"peer_id", "peer_ep_group_id", "local_id", "local_ep_group_id",
					"status",
				}).AddRow(
					"conn-1", "proj-1", true, "site1",
					"vpn-svc-1", "ike-1", "ipsec-1",
					"peer-1", "epg-1", "local-1", "epg-2",
					"ACTIVE",
				)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnSiteConnections)).WillReturnRows(connRows)

				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnEndpointGroupCount)).WillReturnRows(
					sqlmock.NewRows([]string{"cnt"}).AddRow(int64(2)),
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnIkePolicyCount)).WillReturnRows(
					sqlmock.NewRows([]string{"cnt"}).AddRow(int64(3)),
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnIpsecPolicyCount)).WillReturnRows(
					sqlmock.NewRows([]string{"cnt"}).AddRow(int64(1)),
				)
			},
			ExpectedMetrics: `# HELP openstack_neutron_vpn_endpoint_groups vpn_endpoint_groups
# TYPE openstack_neutron_vpn_endpoint_groups gauge
openstack_neutron_vpn_endpoint_groups 2
# HELP openstack_neutron_vpn_ike_policies vpn_ike_policies
# TYPE openstack_neutron_vpn_ike_policies gauge
openstack_neutron_vpn_ike_policies 3
# HELP openstack_neutron_vpn_ipsec_policies vpn_ipsec_policies
# TYPE openstack_neutron_vpn_ipsec_policies gauge
openstack_neutron_vpn_ipsec_policies 1
# HELP openstack_neutron_vpn_service vpn_service
# TYPE openstack_neutron_vpn_service gauge
openstack_neutron_vpn_service{admin_state_up="true",external_ipv4="1.2.3.4",external_ipv6="",flavor_id="",id="vpn-svc-1",name="my-vpn",project_id="proj-1",router_id="router-1",subnet_id="subnet-1"} 1
# HELP openstack_neutron_vpn_services vpn_services
# TYPE openstack_neutron_vpn_services gauge
openstack_neutron_vpn_services 1
# HELP openstack_neutron_vpn_siteconnection vpn_siteconnection
# TYPE openstack_neutron_vpn_siteconnection gauge
openstack_neutron_vpn_siteconnection{admin_state_up="true",id="conn-1",ike_policy_id="ike-1",ipsec_policy_id="ipsec-1",local_ep_group_id="epg-2",local_id="local-1",name="site1",peer_ep_group_id="epg-1",peer_id="peer-1",project_id="proj-1",vpn_service_id="vpn-svc-1"} 1
# HELP openstack_neutron_vpn_siteconnections vpn_siteconnections
# TYPE openstack_neutron_vpn_siteconnections gauge
openstack_neutron_vpn_siteconnections 1
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnServices)).WillReturnRows(
					sqlmock.NewRows([]string{
						"id", "project_id", "subnet_id", "router_id",
						"admin_state_up", "name", "external_v4_ip",
						"external_v6_ip", "flavor_id", "status",
					}),
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnSiteConnections)).WillReturnRows(
					sqlmock.NewRows([]string{
						"id", "project_id", "admin_state_up", "name",
						"vpnservice_id", "ikepolicy_id", "ipsecpolicy_id",
						"peer_id", "peer_ep_group_id", "local_id", "local_ep_group_id",
						"status",
					}),
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnEndpointGroupCount)).WillReturnRows(
					sqlmock.NewRows([]string{"cnt"}).AddRow(int64(0)),
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnIkePolicyCount)).WillReturnRows(
					sqlmock.NewRows([]string{"cnt"}).AddRow(int64(0)),
				)
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnIpsecPolicyCount)).WillReturnRows(
					sqlmock.NewRows([]string{"cnt"}).AddRow(int64(0)),
				)
			},
			ExpectedMetrics: `# HELP openstack_neutron_vpn_endpoint_groups vpn_endpoint_groups
# TYPE openstack_neutron_vpn_endpoint_groups gauge
openstack_neutron_vpn_endpoint_groups 0
# HELP openstack_neutron_vpn_ike_policies vpn_ike_policies
# TYPE openstack_neutron_vpn_ike_policies gauge
openstack_neutron_vpn_ike_policies 0
# HELP openstack_neutron_vpn_ipsec_policies vpn_ipsec_policies
# TYPE openstack_neutron_vpn_ipsec_policies gauge
openstack_neutron_vpn_ipsec_policies 0
# HELP openstack_neutron_vpn_services vpn_services
# TYPE openstack_neutron_vpn_services gauge
openstack_neutron_vpn_services 0
# HELP openstack_neutron_vpn_siteconnections vpn_siteconnections
# TYPE openstack_neutron_vpn_siteconnections gauge
openstack_neutron_vpn_siteconnections 0
`,
		},
		{
			Name: "query error on vpn services",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(neutrondb.GetVpnServices)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
		},
	}

	testutil.RunCollectorTests(t, tests, NewVPNCollector)
}
