package jobdbsync

import (
	"context"

	"backend/core-server/internal/model/event"
)

func (c *MessageQueueConsumer) handleUserLike(ctx context.Context, msg *event.EventUserThumbUp) error {
	deps := c.NewLikeHandlerDependencies()
	return HandleUserLike(ctx, msg, deps)
}

func (c *MessageQueueConsumer) handleUserCancelThumbUp(ctx context.Context, msg *event.EventUserCancelThumbUp) error {
	deps := c.NewLikeHandlerDependencies()
	return HandleUserCancelLike(ctx, msg, deps)
}
