package cinder

import (
	"database/sql"
	"io"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vexxhost/openstack_database_exporter/internal/collector/project"
	cinderdb "github.com/vexxhost/openstack_database_exporter/internal/db/cinder"
)

func TestLimitsCollector(t *testing.T) {
	cols := []string{"project_id", "resource", "hard_limit", "in_use"}
	vtCols := []string{"id", "name"}

	type limitsTestCase struct {
		Name            string
		SetupMock       func(sqlmock.Sqlmock)
		ExpectedMetrics string
		ExpectError     bool
	}

	tests := []limitsTestCase{
		{
			Name: "successful collection with quota limits",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols)

				for _, id := range []string{
					"0c4e939acacf4376bdcd1129f1a054ad",
					"0cbd49cbf76d405d9c86562e1d579bd3",
					"2db68fed84324f29bb73130c6c2094fb",
					"3d594eb0f04741069dbbb521635b21c7",
					"43ebde53fc314b1c9ea2b8c5dc744927",
					"4b1eb781a47440acb8af9850103e537f",
					"5961c443439d4fcebe42643723755e9d",
					"fdb8424c4e4f4c0ba32c52e2de3bd80e",
				} {
					rows.AddRow(id, "gigabytes", 1000, 0)
					rows.AddRow(id, "backup_gigabytes", 1000, 0)
				}

				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetVolumeTypes)).WillReturnRows(sqlmock.NewRows(vtCols))
			},
			ExpectedMetrics: `# HELP openstack_cinder_limits_backup_max_gb limits_backup_max_gb
# TYPE openstack_cinder_limits_backup_max_gb gauge
openstack_cinder_limits_backup_max_gb{tenant="0c4e939acacf4376bdcd1129f1a054ad",tenant_id="0c4e939acacf4376bdcd1129f1a054ad"} 1000
openstack_cinder_limits_backup_max_gb{tenant="0cbd49cbf76d405d9c86562e1d579bd3",tenant_id="0cbd49cbf76d405d9c86562e1d579bd3"} 1000
openstack_cinder_limits_backup_max_gb{tenant="2db68fed84324f29bb73130c6c2094fb",tenant_id="2db68fed84324f29bb73130c6c2094fb"} 1000
openstack_cinder_limits_backup_max_gb{tenant="3d594eb0f04741069dbbb521635b21c7",tenant_id="3d594eb0f04741069dbbb521635b21c7"} 1000
openstack_cinder_limits_backup_max_gb{tenant="43ebde53fc314b1c9ea2b8c5dc744927",tenant_id="43ebde53fc314b1c9ea2b8c5dc744927"} 1000
openstack_cinder_limits_backup_max_gb{tenant="4b1eb781a47440acb8af9850103e537f",tenant_id="4b1eb781a47440acb8af9850103e537f"} 1000
openstack_cinder_limits_backup_max_gb{tenant="5961c443439d4fcebe42643723755e9d",tenant_id="5961c443439d4fcebe42643723755e9d"} 1000
openstack_cinder_limits_backup_max_gb{tenant="fdb8424c4e4f4c0ba32c52e2de3bd80e",tenant_id="fdb8424c4e4f4c0ba32c52e2de3bd80e"} 1000
# HELP openstack_cinder_limits_backup_used_gb limits_backup_used_gb
# TYPE openstack_cinder_limits_backup_used_gb gauge
openstack_cinder_limits_backup_used_gb{tenant="0c4e939acacf4376bdcd1129f1a054ad",tenant_id="0c4e939acacf4376bdcd1129f1a054ad"} 0
openstack_cinder_limits_backup_used_gb{tenant="0cbd49cbf76d405d9c86562e1d579bd3",tenant_id="0cbd49cbf76d405d9c86562e1d579bd3"} 0
openstack_cinder_limits_backup_used_gb{tenant="2db68fed84324f29bb73130c6c2094fb",tenant_id="2db68fed84324f29bb73130c6c2094fb"} 0
openstack_cinder_limits_backup_used_gb{tenant="3d594eb0f04741069dbbb521635b21c7",tenant_id="3d594eb0f04741069dbbb521635b21c7"} 0
openstack_cinder_limits_backup_used_gb{tenant="43ebde53fc314b1c9ea2b8c5dc744927",tenant_id="43ebde53fc314b1c9ea2b8c5dc744927"} 0
openstack_cinder_limits_backup_used_gb{tenant="4b1eb781a47440acb8af9850103e537f",tenant_id="4b1eb781a47440acb8af9850103e537f"} 0
openstack_cinder_limits_backup_used_gb{tenant="5961c443439d4fcebe42643723755e9d",tenant_id="5961c443439d4fcebe42643723755e9d"} 0
openstack_cinder_limits_backup_used_gb{tenant="fdb8424c4e4f4c0ba32c52e2de3bd80e",tenant_id="fdb8424c4e4f4c0ba32c52e2de3bd80e"} 0
# HELP openstack_cinder_limits_volume_max_gb limits_volume_max_gb
# TYPE openstack_cinder_limits_volume_max_gb gauge
openstack_cinder_limits_volume_max_gb{tenant="0c4e939acacf4376bdcd1129f1a054ad",tenant_id="0c4e939acacf4376bdcd1129f1a054ad"} 1000
openstack_cinder_limits_volume_max_gb{tenant="0cbd49cbf76d405d9c86562e1d579bd3",tenant_id="0cbd49cbf76d405d9c86562e1d579bd3"} 1000
openstack_cinder_limits_volume_max_gb{tenant="2db68fed84324f29bb73130c6c2094fb",tenant_id="2db68fed84324f29bb73130c6c2094fb"} 1000
openstack_cinder_limits_volume_max_gb{tenant="3d594eb0f04741069dbbb521635b21c7",tenant_id="3d594eb0f04741069dbbb521635b21c7"} 1000
openstack_cinder_limits_volume_max_gb{tenant="43ebde53fc314b1c9ea2b8c5dc744927",tenant_id="43ebde53fc314b1c9ea2b8c5dc744927"} 1000
openstack_cinder_limits_volume_max_gb{tenant="4b1eb781a47440acb8af9850103e537f",tenant_id="4b1eb781a47440acb8af9850103e537f"} 1000
openstack_cinder_limits_volume_max_gb{tenant="5961c443439d4fcebe42643723755e9d",tenant_id="5961c443439d4fcebe42643723755e9d"} 1000
openstack_cinder_limits_volume_max_gb{tenant="fdb8424c4e4f4c0ba32c52e2de3bd80e",tenant_id="fdb8424c4e4f4c0ba32c52e2de3bd80e"} 1000
# HELP openstack_cinder_limits_volume_used_gb limits_volume_used_gb
# TYPE openstack_cinder_limits_volume_used_gb gauge
openstack_cinder_limits_volume_used_gb{tenant="0c4e939acacf4376bdcd1129f1a054ad",tenant_id="0c4e939acacf4376bdcd1129f1a054ad"} 0
openstack_cinder_limits_volume_used_gb{tenant="0cbd49cbf76d405d9c86562e1d579bd3",tenant_id="0cbd49cbf76d405d9c86562e1d579bd3"} 0
openstack_cinder_limits_volume_used_gb{tenant="2db68fed84324f29bb73130c6c2094fb",tenant_id="2db68fed84324f29bb73130c6c2094fb"} 0
openstack_cinder_limits_volume_used_gb{tenant="3d594eb0f04741069dbbb521635b21c7",tenant_id="3d594eb0f04741069dbbb521635b21c7"} 0
openstack_cinder_limits_volume_used_gb{tenant="43ebde53fc314b1c9ea2b8c5dc744927",tenant_id="43ebde53fc314b1c9ea2b8c5dc744927"} 0
openstack_cinder_limits_volume_used_gb{tenant="4b1eb781a47440acb8af9850103e537f",tenant_id="4b1eb781a47440acb8af9850103e537f"} 0
openstack_cinder_limits_volume_used_gb{tenant="5961c443439d4fcebe42643723755e9d",tenant_id="5961c443439d4fcebe42643723755e9d"} 0
openstack_cinder_limits_volume_used_gb{tenant="fdb8424c4e4f4c0ba32c52e2de3bd80e",tenant_id="fdb8424c4e4f4c0ba32c52e2de3bd80e"} 0
`,
		},
		{
			Name: "empty results",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetVolumeTypes)).WillReturnRows(sqlmock.NewRows(vtCols))
			},
			ExpectedMetrics: "",
		},
		{
			Name: "single project with non-zero usage",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("proj-abc", "gigabytes", 500, 250).
					AddRow("proj-abc", "backup_gigabytes", 200, 75)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetVolumeTypes)).WillReturnRows(sqlmock.NewRows(vtCols))
			},
			ExpectedMetrics: `# HELP openstack_cinder_limits_backup_max_gb limits_backup_max_gb
# TYPE openstack_cinder_limits_backup_max_gb gauge
openstack_cinder_limits_backup_max_gb{tenant="proj-abc",tenant_id="proj-abc"} 200
# HELP openstack_cinder_limits_backup_used_gb limits_backup_used_gb
# TYPE openstack_cinder_limits_backup_used_gb gauge
openstack_cinder_limits_backup_used_gb{tenant="proj-abc",tenant_id="proj-abc"} 75
# HELP openstack_cinder_limits_volume_max_gb limits_volume_max_gb
# TYPE openstack_cinder_limits_volume_max_gb gauge
openstack_cinder_limits_volume_max_gb{tenant="proj-abc",tenant_id="proj-abc"} 500
# HELP openstack_cinder_limits_volume_used_gb limits_volume_used_gb
# TYPE openstack_cinder_limits_volume_used_gb gauge
openstack_cinder_limits_volume_used_gb{tenant="proj-abc",tenant_id="proj-abc"} 250
`,
		},
		{
			Name: "only gigabytes resource (no backup) - defaults applied",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("proj-1", "gigabytes", 1000, 100)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetVolumeTypes)).WillReturnRows(sqlmock.NewRows(vtCols))
			},
			ExpectedMetrics: `# HELP openstack_cinder_limits_backup_max_gb limits_backup_max_gb
# TYPE openstack_cinder_limits_backup_max_gb gauge
openstack_cinder_limits_backup_max_gb{tenant="proj-1",tenant_id="proj-1"} 1000
# HELP openstack_cinder_limits_backup_used_gb limits_backup_used_gb
# TYPE openstack_cinder_limits_backup_used_gb gauge
openstack_cinder_limits_backup_used_gb{tenant="proj-1",tenant_id="proj-1"} 0
# HELP openstack_cinder_limits_volume_max_gb limits_volume_max_gb
# TYPE openstack_cinder_limits_volume_max_gb gauge
openstack_cinder_limits_volume_max_gb{tenant="proj-1",tenant_id="proj-1"} 1000
# HELP openstack_cinder_limits_volume_used_gb limits_volume_used_gb
# TYPE openstack_cinder_limits_volume_used_gb gauge
openstack_cinder_limits_volume_used_gb{tenant="proj-1",tenant_id="proj-1"} 100
`,
		},
		{
			Name: "volume type quotas default to -1 when no per-type quota exists",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("proj-1", "gigabytes", 1000, 50).
					AddRow("proj-1", "backup_gigabytes", 500, 10)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnRows(rows)
				vtRows := sqlmock.NewRows(vtCols).
					AddRow("type-1", "standard").
					AddRow("type-2", "__DEFAULT__")
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetVolumeTypes)).WillReturnRows(vtRows)
			},
			ExpectedMetrics: `# HELP openstack_cinder_limits_backup_max_gb limits_backup_max_gb
# TYPE openstack_cinder_limits_backup_max_gb gauge
openstack_cinder_limits_backup_max_gb{tenant="proj-1",tenant_id="proj-1"} 500
# HELP openstack_cinder_limits_backup_used_gb limits_backup_used_gb
# TYPE openstack_cinder_limits_backup_used_gb gauge
openstack_cinder_limits_backup_used_gb{tenant="proj-1",tenant_id="proj-1"} 10
# HELP openstack_cinder_limits_volume_max_gb limits_volume_max_gb
# TYPE openstack_cinder_limits_volume_max_gb gauge
openstack_cinder_limits_volume_max_gb{tenant="proj-1",tenant_id="proj-1"} 1000
# HELP openstack_cinder_limits_volume_used_gb limits_volume_used_gb
# TYPE openstack_cinder_limits_volume_used_gb gauge
openstack_cinder_limits_volume_used_gb{tenant="proj-1",tenant_id="proj-1"} 50
# HELP openstack_cinder_volume_type_quota_gigabytes volume_type_quota_gigabytes
# TYPE openstack_cinder_volume_type_quota_gigabytes gauge
openstack_cinder_volume_type_quota_gigabytes{tenant="proj-1",tenant_id="proj-1",volume_type="__DEFAULT__"} -1
openstack_cinder_volume_type_quota_gigabytes{tenant="proj-1",tenant_id="proj-1",volume_type="standard"} -1
`,
		},
		{
			Name: "per-volume-type quota picked up from gigabytes_ resource",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow("proj-1", "gigabytes", 1000, 50).
					AddRow("proj-1", "backup_gigabytes", 500, 10).
					AddRow("proj-1", "gigabytes_standard", 300, 0)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnRows(rows)
				vtRows := sqlmock.NewRows(vtCols).
					AddRow("type-1", "standard").
					AddRow("type-2", "__DEFAULT__")
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetVolumeTypes)).WillReturnRows(vtRows)
			},
			ExpectedMetrics: `# HELP openstack_cinder_limits_backup_max_gb limits_backup_max_gb
# TYPE openstack_cinder_limits_backup_max_gb gauge
openstack_cinder_limits_backup_max_gb{tenant="proj-1",tenant_id="proj-1"} 500
# HELP openstack_cinder_limits_backup_used_gb limits_backup_used_gb
# TYPE openstack_cinder_limits_backup_used_gb gauge
openstack_cinder_limits_backup_used_gb{tenant="proj-1",tenant_id="proj-1"} 10
# HELP openstack_cinder_limits_volume_max_gb limits_volume_max_gb
# TYPE openstack_cinder_limits_volume_max_gb gauge
openstack_cinder_limits_volume_max_gb{tenant="proj-1",tenant_id="proj-1"} 1000
# HELP openstack_cinder_limits_volume_used_gb limits_volume_used_gb
# TYPE openstack_cinder_limits_volume_used_gb gauge
openstack_cinder_limits_volume_used_gb{tenant="proj-1",tenant_id="proj-1"} 50
# HELP openstack_cinder_volume_type_quota_gigabytes volume_type_quota_gigabytes
# TYPE openstack_cinder_volume_type_quota_gigabytes gauge
openstack_cinder_volume_type_quota_gigabytes{tenant="proj-1",tenant_id="proj-1",volume_type="__DEFAULT__"} -1
openstack_cinder_volume_type_quota_gigabytes{tenant="proj-1",tenant_id="proj-1",volume_type="standard"} 300
`,
		},
		{
			Name: "null project_id",
			SetupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(cols).
					AddRow(nil, "gigabytes", 1000, 0)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnRows(rows)
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetVolumeTypes)).WillReturnRows(sqlmock.NewRows(vtCols))
			},
			ExpectedMetrics: `# HELP openstack_cinder_limits_backup_max_gb limits_backup_max_gb
# TYPE openstack_cinder_limits_backup_max_gb gauge
openstack_cinder_limits_backup_max_gb{tenant="",tenant_id=""} 1000
# HELP openstack_cinder_limits_backup_used_gb limits_backup_used_gb
# TYPE openstack_cinder_limits_backup_used_gb gauge
openstack_cinder_limits_backup_used_gb{tenant="",tenant_id=""} 0
# HELP openstack_cinder_limits_volume_max_gb limits_volume_max_gb
# TYPE openstack_cinder_limits_volume_max_gb gauge
openstack_cinder_limits_volume_max_gb{tenant="",tenant_id=""} 1000
# HELP openstack_cinder_limits_volume_used_gb limits_volume_used_gb
# TYPE openstack_cinder_limits_volume_used_gb gauge
openstack_cinder_limits_volume_used_gb{tenant="",tenant_id=""} 0
`,
		},
		{
			Name: "query error",
			SetupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(cinderdb.GetProjectQuotaLimits)).WillReturnError(sql.ErrConnDone)
			},
			ExpectedMetrics: "",
			ExpectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			tt.SetupMock(mock)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			// Create a ProjectResolver with no keystone (will fall back to project IDs as names)
			resolver := project.NewResolver(logger, nil, 0)
			collector := NewLimitsCollector(db, logger, resolver)

			if tt.ExpectedMetrics != "" {
				err = testutil.CollectAndCompare(collector, strings.NewReader(tt.ExpectedMetrics))
				if tt.ExpectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			} else {
				problems, err := testutil.CollectAndLint(collector)
				assert.Len(t, problems, 0)
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
