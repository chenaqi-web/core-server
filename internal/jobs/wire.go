package jobs

import (
	jobdbsync "core-server/internal/jobs/job-dbsync"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	jobdbsync.NewMessageQueueConsumer,
)
