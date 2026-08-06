package logger

import (
	"context"
	"vault/pkg/ctxkeys"

	"github.com/sirupsen/logrus"
)

var instance *logrus.Logger

func init() {
	instance = logrus.New()
	instance.SetFormatter(&logrus.JSONFormatter{})
	instance.SetLevel(logrus.InfoLevel)
}

// FromContext возвращает Entry с полями из контекста (request_id).
func FromContext(ctx context.Context) *logrus.Entry {
	entry := logrus.NewEntry(instance)
	if requestID, ok := ctx.Value(ctxkeys.RequestIDKey).(string); ok && requestID != "" {
		entry = entry.WithField("request_id", requestID)
	}
	return entry
}

func Print(args ...any) {
	instance.Print(args...)
}

func Printf(format string, args ...any) {
	instance.Printf(format, args...)
}

func Error(args ...any) {
	instance.Error(args...)
}

func Errorf(format string, args ...any) {
	instance.Errorf(format, args...)
}

func Fatal(args ...any) {
	instance.Fatal(args...)
}

func Fatalf(format string, args ...any) {
	instance.Fatalf(format, args...)
}

func Warn(args ...any) {
	instance.Warn(args...)
}

func Warnf(format string, args ...any) {
	instance.Warnf(format, args...)
}

func Info(args ...any) {
	instance.Info(args...)
}
