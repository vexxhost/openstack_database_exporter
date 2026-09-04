package designate

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	designatedb "github.com/vexxhost/openstack_database_exporter/internal/db/designate"
)

var (
	designateUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "up"),
		"up",
		nil,
		nil,
	)

	zonesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "zones"),
		"zones",
		nil,
		nil,
	)

	zoneStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "zone_status"),
		"zone_status",
		[]string{"id", "name", "status", "tenant_id", "type"},
		nil,
	)

	recordsetsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "recordsets"),
		"recordsets",
		[]string{"zone_id", "zone_name", "tenant_id"},
		nil,
	)

	recordsetsStatusDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, Subsystem, "recordsets_status"),
		"recordsets_status",
		[]string{"id", "name", "status", "zone_id", "zone_name", "type"},
		nil,
	)
)

var zoneStatusOrder = []string{"pending", "active", "deleted", "error"}
var recordsetStatusOrder = []string{"pending", "active", "deleted", "error"}

func formatUUID(s string) string {
	if len(s) == 32 {
		return uuid.MustParse(s).String()
	}
	return s
}

func mapStatus(status string, order []string) int {
	s := strings.ToLower(status)
	for idx, v := range order {
		if v == s {
			return idx
		}
	}
	return -1
}

type ZonesCollector struct {
	db      *sql.DB
	queries *designatedb.Queries
	logger  *slog.Logger
}

func NewZonesCollector(db *sql.DB, logger *slog.Logger) *ZonesCollector {
	return &ZonesCollector{
		db:      db,
		queries: designatedb.New(db),
		logger: logger.With(
			"namespace", Namespace,
			"subsystem", Subsystem,
			"collector", "zones",
		),
	}
}

func (c *ZonesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- designateUpDesc
	ch <- zonesDesc
	ch <- zoneStatusDesc
	ch <- recordsetsDesc
	ch <- recordsetsStatusDesc
}

func (c *ZonesCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	zones, err := c.queries.GetZones(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(designateUpDesc, prometheus.GaugeValue, 0)
		c.logger.Error("failed to query zones", "error", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(zonesDesc, prometheus.GaugeValue, float64(len(zones)))

	for _, zone := range zones {
		tenantID := ""
		if zone.TenantID.Valid {
			tenantID = zone.TenantID.String
		}

		ch <- prometheus.MustNewConstMetric(
			zoneStatusDesc,
			prometheus.GaugeValue, float64(mapStatus(string(zone.Status), zoneStatusOrder)),
			formatUUID(zone.ID), zone.Name, string(zone.Status), tenantID, string(zone.Type),
		)
	}

	recordsets, err := c.queries.GetRecordsets(ctx)
	if err != nil {
		c.logger.Error("failed to query recordsets", "error", err)
	} else {
		recordsetsByZone := make(map[string]int)
		for _, rs := range recordsets {
			recordsetsByZone[rs.ZoneID]++

			tenantID := ""
			if rs.TenantID.Valid {
				tenantID = rs.TenantID.String
			}
			_ = tenantID

			statusStr := ""
			switch s := rs.Status.(type) {
			case string:
				statusStr = s
			case []byte:
				statusStr = string(s)
			}

			ch <- prometheus.MustNewConstMetric(
				recordsetsStatusDesc,
				prometheus.GaugeValue, float64(mapStatus(statusStr, recordsetStatusOrder)),
				formatUUID(rs.ID), rs.Name, statusStr, formatUUID(rs.ZoneID), rs.ZoneName, string(rs.Type),
			)
		}

		for _, zone := range zones {
			tenantID := ""
			if zone.TenantID.Valid {
				tenantID = zone.TenantID.String
			}

			ch <- prometheus.MustNewConstMetric(
				recordsetsDesc,
				prometheus.GaugeValue, float64(recordsetsByZone[zone.ID]),
				formatUUID(zone.ID), zone.Name, tenantID,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(designateUpDesc, prometheus.GaugeValue, 1)
}
