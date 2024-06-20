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

type NovaApiBuildRequest struct {
	ProjectID    uuid.UUID
	InstanceUUID uuid.UUID
}

func (NovaApiBuildRequest) TableName() string {
	return "build_requests"
}

type NovaApiDatabaseCollector struct {
	db     *gorm.DB
	logger log.Logger

	buildRequest *prometheus.Desc
}

func newNovaApiCollector(logger log.Logger, db *gorm.DB) prometheus.Collector {
	return &NovaApiDatabaseCollector{
		db:     db,
		logger: logger,

		buildRequest: prometheus.NewDesc(
			"openstack_nova_api_build_request",
			"build_request",
			[]string{"project_id", "instance_uuid"},
			nil,
		),
	}
}

func NewNovaApiDatabaseCollector(logger log.Logger, dsn string) prometheus.Collector {
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

	return newNovaApiCollector(logger, db)
}

func (c *NovaApiDatabaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.buildRequest
}

func (c *NovaApiDatabaseCollector) Collect(ch chan<- prometheus.Metric) {
	rows, err := c.db.Model(&NovaApiBuildRequest{}).Select([]string{"project_id", "instance_uuid"}).Rows()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query database", "err", err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var buildRequest NovaApiBuildRequest
		c.db.ScanRows(rows, &buildRequest)

		fmt.Println(buildRequest.ProjectID.String(), buildRequest.InstanceUUID.String())

		ch <- prometheus.MustNewConstMetric(c.buildRequest,
			prometheus.GaugeValue, float64(1), buildRequest.ProjectID.String(), buildRequest.InstanceUUID.String())
	}
}
