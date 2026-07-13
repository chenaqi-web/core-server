package jobaggregator

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"backend/core-server/internal/domain"
	"backend/core-server/internal/model/entity"
	"backend/core-server/internal/model/enum"

	"github.com/hashicorp/go-multierror"
)

type ObjectCountAggregator struct {
	logger        *slog.Logger
	buffer        map[string]*ObjectCountDelta
	mu            sync.Mutex
	flushInterval time.Duration
	bufferSize    int
	dbTimeout     time.Duration
	countRepo     domain.CountDomain
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewObjectCountAggregator(
	logger *slog.Logger,
	countRepo domain.CountDomain,
	flushInterval time.Duration,
	bufferSize int,
	dbTimeout time.Duration,
) *ObjectCountAggregator {
	ctx, cancel := context.WithCancel(context.Background())
	return &ObjectCountAggregator{
		logger:        logger,
		buffer:        make(map[string]*ObjectCountDelta),
		flushInterval: flushInterval,
		bufferSize:    bufferSize,
		dbTimeout:     dbTimeout,
		countRepo:     countRepo,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (a *ObjectCountAggregator) Start() {
	go a.flushLoop()
}

func (a *ObjectCountAggregator) Stop() {
	a.cancel()
	a.Flush()
}

func (a *ObjectCountAggregator) Push(_ context.Context, interactionType, objectType, objectID string) {
	a.add(interactionType, objectType, objectID, 1)
}

func (a *ObjectCountAggregator) Pop(_ context.Context, interactionType, objectType, objectID string) {
	a.add(interactionType, objectType, objectID, -1)
}

func (a *ObjectCountAggregator) add(interactionType, objectType, objectID string, delta int64) {
	key := getObjectCountDeltaKey(interactionType, objectType, objectID)

	a.mu.Lock()
	if existing, ok := a.buffer[key]; ok {
		existing.Delta += delta
	} else {
		a.buffer[key] = &ObjectCountDelta{
			InteractionType: interactionType,
			ObjectType:      objectType,
			ObjectID:        objectID,
			Delta:           delta,
		}
	}
	shouldFlush := len(a.buffer) >= a.bufferSize
	a.mu.Unlock()

	if shouldFlush {
		a.Flush()
	}
}

func (a *ObjectCountAggregator) Flush() {
	a.mu.Lock()
	if len(a.buffer) == 0 {
		a.mu.Unlock()
		return
	}
	toFlush := a.buffer
	a.buffer = make(map[string]*ObjectCountDelta)
	a.mu.Unlock()

	if err := a.batchUpdateMySQL(toFlush); err != nil {
		a.logger.Error("batch update interaction_count failed", "err", err)
		a.mu.Lock()
		for k, v := range toFlush {
			if existing, ok := a.buffer[k]; ok {
				existing.Delta += v.Delta
			} else {
				a.buffer[k] = v
			}
		}
		a.mu.Unlock()
	}
}

func (a *ObjectCountAggregator) batchUpdateMySQL(data map[string]*ObjectCountDelta) error {
	if len(data) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.dbTimeout)
	defer cancel()

	var multiErr *multierror.Error
	for _, delta := range data {
		ent := &entity.InteractionCount{
			InteractionType: enum.ParseInteractionType(delta.InteractionType),
			ObjectType:      enum.ParseObjectType(delta.ObjectType),
			ObjectID:        delta.ObjectID,
		}
		if err := a.countRepo.Upsert(ctx, ent, delta.Delta); err != nil {
			a.logger.Error("upsert interaction_count failed", "object_id", ent.ObjectID, "err", err)
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr.ErrorOrNil()
}

func (a *ObjectCountAggregator) flushLoop() {
	ticker := time.NewTicker(a.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.Flush()
		case <-a.ctx.Done():
			a.Flush()
			return
		}
	}
}
