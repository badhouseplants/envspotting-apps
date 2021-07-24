package logger

import "context"

func EnpointHit(ctx context.Context) {
	log := GetGrpcLogger(ctx)
	log.Info("enpoint hit")
}

