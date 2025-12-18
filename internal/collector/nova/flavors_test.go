package nova

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	novadb "github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	novaapidb "github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestFlavorsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with flavors data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "flavorid", "name", "vcpus", "memory_mb", "root_gb", "ephemeral_gb", "swap", "rxtx_factor", "disabled", "is_public",
				}).AddRow(
					1, "m1.small", "small", 1, 2048, 20, 0, 0, 1.0, false, true,
				).AddRow(
					2, "m1.medium", "medium", 2, 4096, 40, 0, 0, 1.0, false, true,
				)

				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetFlavors)).WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty flavors",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "flavorid", "name", "vcpus", "memory_mb", "root_gb", "ephemeral_gb", "swap", "rxtx_factor", "disabled", "is_public",
				})
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetFlavors)).WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetFlavors)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		collector := NewFlavorsCollector(logger, novadb.New(db), novaapidb.New(db))
		return &flavorsCollectorWrapper{collector}
	})
}

type flavorsCollectorWrapper struct {
	*FlavorsCollector
}

func (w *flavorsCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	w.FlavorsCollector.Collect(ch)
}
