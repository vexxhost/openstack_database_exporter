package nova

import (
	"database/sql"
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
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
					"project_id", "resource", "hard_limit", "allocated", "reserved",
				}).AddRow(
					"project-1", "instances", 10, 2, 0,
				).AddRow(
					"project-1", "cores", 20, 4, 0,
				).AddRow(
					"project-1", "ram", 40960, 8192, 0,
				)

				mock.ExpectQuery("SELECT (.+) FROM quotas").WillReturnRows(rows)
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty quotas",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"project_id", "resource", "hard_limit", "allocated", "reserved",
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
		// Note: QuotasCollector expects a placement DB as third parameter
		collector := NewQuotasCollector(logger, novadb.New(db), novaapidb.New(db), nil)
		return &quotasCollectorWrapper{collector}
	})
}

type quotasCollectorWrapper struct {
	*QuotasCollector
}

func (w *quotasCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	_ = w.QuotasCollector.Collect(ch)
}
