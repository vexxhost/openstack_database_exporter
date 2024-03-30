package collector

import (
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var loadbalancer_status = []string{
	// Octavia API v2 entities have two status codes present in the response body.
	// The provisioning_status describes the lifecycle status of the entity while the operating_status provides the observed status of the entity.
	// Here we put operating_status in metrics value and provisioning_status in metrics label
	"ONLINE",     // Entity is operating normally. All pool members are healthy
	"DRAINING",   // The member is not accepting new connections
	"OFFLINE",    // Entity is administratively disabled
	"ERROR",      // The entity has failed. The member is failing it's health monitoring checks. All of the pool members are in ERROR
	"NO_MONITOR", // No health monitor is configured for this entity and it's status is unknown
}

func mapLoadbalancerStatus(current string) int {
	for idx, status := range loadbalancer_status {
		if current == status {
			return idx
		}
	}
	return -1
}

type OctaviaLoadBalancer struct {
	ProjectID          uuid.UUID
	ID                 uuid.UUID
	Name               string
	ProvisioningStatus string
	OperatingStatus    string
	Provider           string
	VirtualIP          OctaviaVirtualIP `gorm:"foreignKey:LoadBalancerID"`
}

func (OctaviaLoadBalancer) TableName() string {
	return "load_balancer"
}

type OctaviaVirtualIP struct {
	LoadBalancerID uuid.UUID
	IPAddress      string
}

func (OctaviaVirtualIP) TableName() string {
	return "vip"
}

type OctaviaDatabaseCollector struct {
	db     *gorm.DB
	logger log.Logger

	loadbalancerStatus *prometheus.Desc
}

func newOctaviaCollector(logger log.Logger, db *gorm.DB) prometheus.Collector {
	return &OctaviaDatabaseCollector{
		db:     db,
		logger: logger,

		loadbalancerStatus: prometheus.NewDesc(
			"openstack_loadbalancer_loadbalancer_status",
			"loadbalancer_status",
			[]string{"id", "name", "project_id", "operating_status", "provisioning_status", "provider", "vip_address"},
			nil,
		),
	}
}

func NewOctaviaDatabaseCollector(logger log.Logger, dsn string) prometheus.Collector {
	db, err := gorm.Open(
		mysql.Open(
			fmt.Sprintf("%s?parseTime=True", dsn),
		),
		&gorm.Config{
			Logger: NewGormLogger(logger, DefaultConfig),
		},
	)
	if err != nil {
		panic("failed to connect database")
	}

	return newOctaviaCollector(logger, db)
}

func (c *OctaviaDatabaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.loadbalancerStatus
}

func (c *OctaviaDatabaseCollector) Collect(ch chan<- prometheus.Metric) {
	rows, err := c.db.Model(&OctaviaLoadBalancer{}).Joins(
		"VirtualIP", c.db.Select([]string{"ip_address"}).Model(&OctaviaVirtualIP{}),
	).Select([]string{"id", "name", "project_id", "operating_status", "provisioning_status", "provider"}).Rows()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query database", "err", err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var loadBalancer OctaviaLoadBalancer
		c.db.ScanRows(rows, &loadBalancer)

		ch <- prometheus.MustNewConstMetric(c.loadbalancerStatus,
			prometheus.GaugeValue, float64(mapLoadbalancerStatus(loadBalancer.OperatingStatus)), loadBalancer.ID.String(), loadBalancer.Name, strings.Replace(loadBalancer.ProjectID.String(), "-", "", -1),
			loadBalancer.OperatingStatus, loadBalancer.ProvisioningStatus, loadBalancer.Provider, loadBalancer.VirtualIP.IPAddress)
	}
}
