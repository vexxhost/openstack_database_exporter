package collector

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type NovaInstance struct {
	UUID      uuid.UUID
	TaskState string
	Deleted   int
}

func (NovaInstance) TableName() string {
	return "instances"
}

type NovaDatabaseCollector struct {
	db     *gorm.DB
	logger log.Logger

	serverTaskState *prometheus.Desc
}

func newNovaCollector(logger log.Logger, db *gorm.DB) prometheus.Collector {
	return &NovaDatabaseCollector{
		db:     db,
		logger: logger,

		serverTaskState: prometheus.NewDesc(
			"openstack_nova_server_task_state",
			"server_task_state",
			[]string{"id", "task_state"},
			nil,
		),
	}
}

func NewNovaDatabaseCollector(logger log.Logger, dsn string) prometheus.Collector {
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

	return newNovaCollector(logger, db)
}

func (c *NovaDatabaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.serverTaskState
}

func (c *NovaDatabaseCollector) Collect(ch chan<- prometheus.Metric) {
	rows, err := c.db.Model(&NovaInstance{}).Select([]string{"uuid", "task_state"}).Where(map[string]interface{}{"deleted": 0}).Rows()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query database", "err", err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var instance NovaInstance
		c.db.ScanRows(rows, &instance)

		taskState := 0
		if instance.TaskState != "" {
			taskState = 1
		}

		ch <- prometheus.MustNewConstMetric(c.serverTaskState,
			prometheus.GaugeValue, float64(taskState), instance.UUID.String(), instance.TaskState)
	}
}
