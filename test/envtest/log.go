package envtest

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	ctrl "sigs.k8s.io/controller-runtime"
)

// LogsObserver is an interface for observing logs.
type LogsObserver interface {
	// All returns all the logs that have been observed so far.
	All() []observer.LoggedEntry
}

// CreateTestLogger creates a logger for use in tests.
// It returns the logger - which is also added to the context - and the observer
// which can be used to dump logs if the test fails.
func CreateTestLogger(ctx context.Context) (context.Context, logr.Logger, LogsObserver) {
	core, logs := observer.New(zap.DebugLevel)
	logger := zapr.NewLogger(zap.New(core))
	ctx = ctrl.LoggerInto(ctx, logger)
	// NOTE: do not use ctrl.SetLogger() because that would prevent the tests to
	// pass when run with flag -count N with N > 1 because SetLogger() will only
	// set the logger once.
	return ctx, logger, logs
}

// DumpLogsIfTestFailed dumps the provided logs the if the test failed.
func DumpLogsIfTestFailed(t *testing.T, logs LogsObserver) {
	t.Helper()

	if !t.Failed() {
		return
	}

	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	t.Logf("Test %s failed: dumping controller logs\n", t.Name())
	for _, entry := range logs.All() {
		b, err := encoder.EncodeEntry(entry.Entry, entry.Context)
		assert.NoError(t, err)
		t.Logf("%s", b.String())
	}
}
