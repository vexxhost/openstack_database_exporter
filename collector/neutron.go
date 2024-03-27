package collector

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type NeutronAgentBinary string

const (
	NEUTRON_L3_AGENT_BINARY NeutronAgentBinary = "neutron-l3-agent"
)

type NeutronAgent struct {
	ID                 uuid.UUID
	AgentType          string
	Binary             NeutronAgentBinary
	Topic              string
	Host               string
	AdminStateUp       bool
	CreatedAt          time.Time
	StartedAt          time.Time
	HeartbeatTimestamp time.Time
	Description        string
	Configurations     datatypes.JSON
	Load               int
	AvailabilityZone   string
	ResourceVersions   datatypes.JSON
	ResourcesSynced    bool
}

func (n *NeutronAgent) Alive() bool {
	return time.Since(n.HeartbeatTimestamp) < 75*time.Second
}

func (NeutronAgent) TableName() string {
	return "agents"
}

type NeutronHaRouterAgentPortBindingsState string

const (
	ACTIVE NeutronHaRouterAgentPortBindingsState = "active"
	BACKUP NeutronHaRouterAgentPortBindingsState = "backup"
)

type NeutronHaRouterAgentPortBindings struct {
	RouterID  uuid.UUID
	L3AgentId uuid.UUID
	L3Agent   NeutronAgent
	State     NeutronHaRouterAgentPortBindingsState
}

func (NeutronHaRouterAgentPortBindings) TableName() string {
	return "ha_router_agent_port_bindings"
}

type NeutronDatabaseCollector struct {
	db     *gorm.DB
	logger log.Logger

	l3AgentOfRouter *prometheus.Desc
}

func NewNeutronDatabaseCollector(logger log.Logger, dsn string) prometheus.Collector {
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

	return &NeutronDatabaseCollector{
		db: db,

		l3AgentOfRouter: prometheus.NewDesc(
			"openstack_neutron_l3_agent_of_router",
			"l3_agent_of_router",
			[]string{"router_id", "l3_agent_id", "ha_state", "agent_alive", "agent_admin_up", "agent_host"},
			nil,
		),
	}
}

func (c *NeutronDatabaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.l3AgentOfRouter
}

func (c *NeutronDatabaseCollector) Collect(ch chan<- prometheus.Metric) {
	rows, err := c.db.Model(&NeutronHaRouterAgentPortBindings{}).Joins(
		"L3Agent", c.db.Select([]string{"id", "heartbeat_timestamp", "admin_state_up", "host"}).Model(&NeutronAgent{}),
	).Select([]string{"router_id", "l3_agent_id", "state"}).Rows()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query database", "err", err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var binding NeutronHaRouterAgentPortBindings
		c.db.ScanRows(rows, &binding)

		var state int
		if binding.L3Agent.Alive() {
			state = 1
		}

		ch <- prometheus.MustNewConstMetric(c.l3AgentOfRouter,
			prometheus.GaugeValue, float64(state), binding.RouterID.String(), binding.L3Agent.ID.String(),
			string(binding.State), strconv.FormatBool(binding.L3Agent.Alive()), strconv.FormatBool(binding.L3Agent.AdminStateUp), binding.L3Agent.Host)
	}

}
