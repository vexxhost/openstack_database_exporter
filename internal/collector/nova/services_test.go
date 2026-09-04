package nova

import (
	"database/sql"
	"log/slog"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	novadb "github.com/vexxhost/openstack_database_exporter/internal/db/nova"
	novaapidb "github.com/vexxhost/openstack_database_exporter/internal/db/nova_api"
	"github.com/vexxhost/openstack_database_exporter/internal/testutil"
)

func TestServicesCollector(t *testing.T) {
	now := time.Now()

	tests := []testutil.CollectorTestCase{
		{
			Name: "successful collection with services data",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "host", "binary", "topic", "disabled", "disabled_reason",
					"last_seen_up", "forced_down", "version", "report_count", "deleted",
				}).AddRow(
					1, "uuid-scheduler-1", "controller-1", "nova-scheduler", "scheduler", 0, "",
					now, 0, 29, 150, 0,
				).AddRow(
					2, "uuid-compute-1", "compute-1", "nova-compute", "compute", 0, "",
					now, 0, 29, 200, 0,
				).AddRow(
					3, "uuid-compute-2", "compute-2", "nova-compute", "compute", 1, "maintenance",
					now, 0, 29, 180, 0,
				)

				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetServices)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetHostAvailabilityZones)).WillReturnRows(
					sqlmock.NewRows([]string{"host", "availability_zone"}).
						AddRow("compute-1", "az1").
						AddRow("compute-2", "az2"),
				)
			},
			ExpectedMetrics: `# HELP openstack_nova_agent_state agent_state
# TYPE openstack_nova_agent_state gauge
openstack_nova_agent_state{adminState="disabled",disabledReason="maintenance",hostname="compute-2",id="uuid-compute-2",service="nova-compute",zone="az2"} 0
openstack_nova_agent_state{adminState="enabled",disabledReason="",hostname="compute-1",id="uuid-compute-1",service="nova-compute",zone="az1"} 1
openstack_nova_agent_state{adminState="enabled",disabledReason="",hostname="controller-1",id="uuid-scheduler-1",service="nova-scheduler",zone="internal"} 1
# HELP openstack_nova_availability_zones availability_zones
# TYPE openstack_nova_availability_zones gauge
openstack_nova_availability_zones 2
`,
		},
		{
			Name: "services with mixed states",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "host", "binary", "topic", "disabled", "disabled_reason",
					"last_seen_up", "forced_down", "version", "report_count", "deleted",
				}).AddRow(
					1, "uuid-scheduler-1", "controller-1", "nova-scheduler", "scheduler", 0, "",
					now, 0, 29, 150, 0,
				).AddRow(
					2, "uuid-compute-1", "compute-1", "nova-compute", "compute", 1, "down for maintenance",
					now, 1, 29, 100, 0,
				).AddRow(
					3, "uuid-conductor-1", "controller-1", "nova-conductor", "conductor", 0, "",
					now, 0, 29, 175, 0,
				)

				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetServices)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetHostAvailabilityZones)).WillReturnRows(
					sqlmock.NewRows([]string{"host", "availability_zone"}).
						AddRow("compute-1", "az1"),
				)
			},
			ExpectedMetrics: `# HELP openstack_nova_agent_state agent_state
# TYPE openstack_nova_agent_state gauge
openstack_nova_agent_state{adminState="disabled",disabledReason="down for maintenance",hostname="compute-1",id="uuid-compute-1",service="nova-compute",zone="az1"} 0
openstack_nova_agent_state{adminState="enabled",disabledReason="",hostname="controller-1",id="uuid-conductor-1",service="nova-conductor",zone="internal"} 1
openstack_nova_agent_state{adminState="enabled",disabledReason="",hostname="controller-1",id="uuid-scheduler-1",service="nova-scheduler",zone="internal"} 1
# HELP openstack_nova_availability_zones availability_zones
# TYPE openstack_nova_availability_zones gauge
openstack_nova_availability_zones 1
`,
		},
		{
			Name: "empty services",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "uuid", "host", "binary", "topic", "disabled", "disabled_reason",
					"last_seen_up", "forced_down", "version", "report_count", "deleted",
				})
				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetServices)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(novaapidb.GetHostAvailabilityZones)).WillReturnRows(
					sqlmock.NewRows([]string{"host", "availability_zone"}),
				)
			},
			ExpectedMetrics: `# HELP openstack_nova_agent_state agent_state
# TYPE openstack_nova_agent_state gauge
# HELP openstack_nova_availability_zones availability_zones
# TYPE openstack_nova_availability_zones gauge
openstack_nova_availability_zones 0
`,
		},
		{
			Name: "database query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(novadb.GetServices)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: ``,
		},
	}

	testutil.RunCollectorTests(t, tests, func(db *sql.DB, logger *slog.Logger) prometheus.Collector {
		collector := NewServicesCollector(logger, novadb.New(db), novaapidb.New(db))
		return &servicesCollectorWrapper{collector}
	})
}

type servicesCollectorWrapper struct {
	*ServicesCollector
}

func (w *servicesCollectorWrapper) Collect(ch chan<- prometheus.Metric) {
	_ = w.ServicesCollector.Collect(ch)
}
