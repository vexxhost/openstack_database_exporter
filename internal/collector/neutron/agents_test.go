package neutron

import (
	"io"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestAgentsCollector(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("successful collection with agents", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer func() { _ = db.Close() }()

		rows := sqlmock.NewRows([]string{
			"id", "agent_type", "service", "hostname", "admin_state", "zone", "alive",
		}).
			AddRow("agent-001", "L3 agent", "neutron-l3-agent", "ctrl-01", "enabled", "nova", 1).
			AddRow("agent-002", "DHCP agent", "neutron-dhcp-agent", "ctrl-02", "enabled", "nova", 0)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, a.agent_type, a.`binary` as service")).WillReturnRows(rows)

		collector := NewAgentsCollector(db, logger)

		expected := `# HELP openstack_neutron_agent_state agent_state
# TYPE openstack_neutron_agent_state gauge
openstack_neutron_agent_state{adminState="enabled",hostname="ctrl-01",id="agent-001",service="neutron-l3-agent",zone="nova"} 1
openstack_neutron_agent_state{adminState="enabled",hostname="ctrl-02",id="agent-002",service="neutron-dhcp-agent",zone="nova"} 0
`
		err = testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("disabled agent", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer func() { _ = db.Close() }()

		rows := sqlmock.NewRows([]string{
			"id", "agent_type", "service", "hostname", "admin_state", "zone", "alive",
		}).
			AddRow("agent-001", "L3 agent", "neutron-l3-agent", "ctrl-01", "disabled", "nova", 1)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, a.agent_type, a.`binary` as service")).WillReturnRows(rows)

		collector := NewAgentsCollector(db, logger)

		expected := `# HELP openstack_neutron_agent_state agent_state
# TYPE openstack_neutron_agent_state gauge
openstack_neutron_agent_state{adminState="disabled",hostname="ctrl-01",id="agent-001",service="neutron-l3-agent",zone="nova"} 1
`
		err = testutil.CollectAndCompare(collector, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty results", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer func() { _ = db.Close() }()

		rows := sqlmock.NewRows([]string{
			"id", "agent_type", "service", "hostname", "admin_state", "zone", "alive",
		})
		mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, a.agent_type, a.`binary` as service")).WillReturnRows(rows)

		collector := NewAgentsCollector(db, logger)
		count := testutil.CollectAndCount(collector)
		if count != 0 {
			t.Fatalf("expected 0 metrics for empty agents, got %d", count)
		}
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create sqlmock: %v", err)
		}
		defer func() { _ = db.Close() }()

		mock.ExpectQuery(regexp.QuoteMeta("SELECT a.id, a.agent_type, a.`binary` as service")).
			WillReturnError(sqlmock.ErrCancelled)

		collector := NewAgentsCollector(db, logger)
		count := testutil.CollectAndCount(collector)
		if count != 0 {
			t.Fatalf("expected 0 metrics on error, got %d", count)
		}
	})
}
