package neutron

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	neutrondb "github.com/vexxhost/openstack_database_exporter/internal/db/neutron"
)

var (
	vpnServiceDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "vpn_service"),
		"vpn_service",
		[]string{
			"id",
			"project_id",
			"subnet_id",
			"router_id",
			"admin_state_up",
			"name",
			"external_ipv4",
			"external_ipv6",
			"flavor_id",
		},
		nil,
	)

	vpnServicesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "vpn_services"),
		"vpn_services",
		nil,
		nil,
	)

	vpnSiteConnectionDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "vpn_siteconnection"),
		"vpn_siteconnection",
		[]string{
			"id",
			"project_id",
			"admin_state_up",
			"name",
			"vpn_service_id",
			"ike_policy_id",
			"ipsec_policy_id",
			"peer_id",
			"peer_ep_group_id",
			"local_id",
			"local_ep_group_id",
		},
		nil,
	)

	vpnSiteConnectionsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "vpn_siteconnections"),
		"vpn_siteconnections",
		nil,
		nil,
	)

	vpnEndpointGroupsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "vpn_endpoint_groups"),
		"vpn_endpoint_groups",
		nil,
		nil,
	)

	vpnIkePoliciesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "vpn_ike_policies"),
		"vpn_ike_policies",
		nil,
		nil,
	)

	vpnIpsecPoliciesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "vpn_ipsec_policies"),
		"vpn_ipsec_policies",
		nil,
		nil,
	)
)

type VPNCollector struct {
	db      *sql.DB
	queries *neutrondb.Queries
	logger  *slog.Logger
}

func NewVPNCollector(db *sql.DB, logger *slog.Logger) *VPNCollector {
	return &VPNCollector{
		db:      db,
		queries: neutrondb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "vpn",
		),
	}
}

func (c *VPNCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- vpnServiceDesc
	ch <- vpnServicesDesc
	ch <- vpnSiteConnectionDesc
	ch <- vpnSiteConnectionsDesc
	ch <- vpnEndpointGroupsDesc
	ch <- vpnIkePoliciesDesc
	ch <- vpnIpsecPoliciesDesc
}

func (c *VPNCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	services, err := c.queries.GetVpnServices(ctx)
	if err != nil {
		c.logger.Error("failed to query vpn services", "error", err)
		return
	}

	for _, s := range services {
		ch <- prometheus.MustNewConstMetric(
			vpnServiceDesc,
			prometheus.GaugeValue,
			1,
			s.ID,
			s.ProjectID,
			s.SubnetID,
			s.RouterID,
			strconv.FormatBool(s.AdminStateUp),
			s.Name,
			s.ExternalV4Ip,
			s.ExternalV6Ip,
			s.FlavorID,
		)
	}

	ch <- prometheus.MustNewConstMetric(vpnServicesDesc, prometheus.GaugeValue, float64(len(services)))

	connections, err := c.queries.GetVpnSiteConnections(ctx)
	if err != nil {
		c.logger.Error("failed to query vpn site connections", "error", err)
		return
	}

	for _, c2 := range connections {
		ch <- prometheus.MustNewConstMetric(
			vpnSiteConnectionDesc,
			prometheus.GaugeValue,
			1,
			c2.ID,
			c2.ProjectID,
			strconv.FormatBool(c2.AdminStateUp),
			c2.Name,
			c2.VpnserviceID,
			c2.IkepolicyID,
			c2.IpsecpolicyID,
			c2.PeerID,
			c2.PeerEpGroupID,
			c2.LocalID,
			c2.LocalEpGroupID,
		)
	}

	ch <- prometheus.MustNewConstMetric(vpnSiteConnectionsDesc, prometheus.GaugeValue, float64(len(connections)))

	endpointGroupCount, err := c.queries.GetVpnEndpointGroupCount(ctx)
	if err != nil {
		c.logger.Error("failed to query vpn endpoint groups", "error", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(vpnEndpointGroupsDesc, prometheus.GaugeValue, float64(endpointGroupCount))

	ikePolicyCount, err := c.queries.GetVpnIkePolicyCount(ctx)
	if err != nil {
		c.logger.Error("failed to query vpn ike policies", "error", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(vpnIkePoliciesDesc, prometheus.GaugeValue, float64(ikePolicyCount))

	ipsecPolicyCount, err := c.queries.GetVpnIpsecPolicyCount(ctx)
	if err != nil {
		c.logger.Error("failed to query vpn ipsec policies", "error", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(vpnIpsecPoliciesDesc, prometheus.GaugeValue, float64(ipsecPolicyCount))
}
