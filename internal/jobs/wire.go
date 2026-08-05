package jobs

import (
	"log/slog"

	jobdbsync "core-server/internal/jobs/job-dbsync"

	"github.com/google/wire"
)

func NewLogger() *slog.Logger {
	return slog.Default()
}

var ProviderSet = wire.NewSet(
	NewLogger,
	jobdbsync.NewMessageQueueConsumer,
)
