package collector

import (
	"context"
	"fmt"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	gormlogger.Interface

	config *gormlogger.Config
	logger kitlog.Logger
}

var (
	DefaultConfig = &gormlogger.Config{
		SlowThreshold: 200 * time.Millisecond,
	}
)

func NewGormLogger(logger kitlog.Logger, config *gormlogger.Config) *GormLogger {
	return &GormLogger{
		config: config,
		logger: logger,
	}
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	level.Info(l.logger).Log(msg, data)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	level.Warn(l.logger).Log(msg, data)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	level.Error(l.logger).Log(msg, data)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	var logger kitlog.Logger
	var msg string

	if err != nil {
		logger = level.Error(l.logger)
		msg = "gorm error"
	} else if elapsed > l.config.SlowThreshold {
		logger = level.Warn(l.logger)
		msg = fmt.Sprintf("SLOW SQL >= %v", l.config.SlowThreshold)
	} else {
		logger = level.Debug(l.logger)
		msg = "gorm trace"
	}

	if rows == -1 {
		logger.Log("msg", msg, "err", err, "duration", float64(elapsed.Nanoseconds())/1e6, "sql", sql)
	} else {
		logger.Log("msg", msg, "err", err, "duration", float64(elapsed.Nanoseconds())/1e6, "sql", sql, "rows", rows)
	}
}
