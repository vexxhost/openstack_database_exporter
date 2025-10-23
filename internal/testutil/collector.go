package testutil

import (
	"database/sql"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CollectorTestCase struct {
	Name            string
	SetupMock       func(sqlmock.Sqlmock)
	ExpectedMetrics string
	ExpectError     bool
}

type CollectorFactory[T prometheus.Collector] func(*sql.DB, *slog.Logger) T

func RunCollectorTests[T prometheus.Collector](t *testing.T, tests []CollectorTestCase, newCollector CollectorFactory[T]) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)

			tt.SetupMock(mock)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			collector := newCollector(db, logger)

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
