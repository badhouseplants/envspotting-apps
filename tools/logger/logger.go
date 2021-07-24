package logger

import (
	"context"

	"github.com/sirupsen/logrus"

	grpcLogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	grpcCtxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

var (
	logrusLogger = logrus.New()
	logger       *logrus.Entry
	// decider      = func(code codes.Code) logrus.Level {
	// if code == codes.OK {
	// return logrus.InfoLevel
	// }
	// return logrus.ErrorLevel
	// }
	GrpcLogrusEntry *logrus.Entry
	GrpcLogrusOpts  []grpcLogrus.Option
)

func init() {
	logrusLogger.Formatter = &logrus.JSONFormatter{}
	GrpcLogrusEntry = logrus.NewEntry(logrusLogger)
	GrpcLogrusOpts = []grpcLogrus.Option{
		// grpcLogrus.WithLevels(decider),
	}
}

func GetGrpcLogger(ctx context.Context) *logrus.Entry {
	grpcCtxtags.Extract(ctx)
	return ctxlogrus.Extract(ctx)
}

func SetServerLogger() {
	logrusLogger.Formatter = &logrus.JSONFormatter{}
	logger = logrus.NewEntry(logrusLogger)
}

func GetServerLogger() *logrus.Entry {
	if logger == nil {
		SetServerLogger()
	}
	return logger
}
