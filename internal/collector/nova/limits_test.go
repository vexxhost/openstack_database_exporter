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

func TestLimitsCollector(t *testing.T) {
	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with limits data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Mock GetQuotas query
				quotasRows := sqlmock.NewRows([]string{
					"id", "project_id", "resource", "hard_limit",
				}).AddRow(
					1, "project1", "instances", 10,
				).AddRow(
					2, "project1", "cores", 20,
				).AddRow(
					3, "project1", "ram", 51200,
				)
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetQuotas)).WillReturnRows(quotasRows)

				// Note: No placement query expected since placementDB is nil in tests
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "empty limits data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				// Mock empty GetQuotas result
				quotasRows := sqlmock.NewRows([]string{
					"id", "project_id", "resource", "hard_limit",
				})
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetQuotas)).WillReturnRows(quotasRows)

				// Note: No placement query expected since placementDB is nil in tests
			},
			ExpectedMetrics: ``,
		},
		{
			Name: "quota query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetQuotas)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		// Note: LimitsCollector expects a placement DB as fourth parameter, but passing nil for tests
		collector := NewLimitsCollector(logger, novadb.New(db), novaapidb.New(db), nil)
		return &limitsCollectorWrapper{collector}
	})
}

type limitsCollectorWrapper struct {
	*LimitsCollector
}

func (w *limitsCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	_ = w.LimitsCollector.Collect(ch)
}
