package nova

import (
	"database/sql"
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	novadb "github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	novaapidb "github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestQuotasCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with quotas data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "project_id", "resource", "hard_limit",
				}).AddRow(
					1, "project-1", "instances", 10,
				).AddRow(
					2, "project-1", "cores", 20,
				).AddRow(
					3, "project-1", "ram", 40960,
				)

				mock.ExpectQuery("SELECT (.+) FROM quotas").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty quotas",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "project_id", "resource", "hard_limit",
				})
				mock.ExpectQuery("SELECT (.+) FROM quotas").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM quotas").WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		collector := NewQuotasCollector(logger, novadb.New(db), novaapidb.New(db), nil, project.NewResolver(logger, nil, 0))
		return &quotasCollectorWrapper{collector}
	})
}

type quotasCollectorWrapper struct {
	*QuotasCollector
}

func (w *quotasCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	_ = w.QuotasCollector.Collect(ch)
}
