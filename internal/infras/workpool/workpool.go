package workerpool

import (
	"context"
	"core-server/internal/infras/clog"
	"core-server/internal/model/event"
	"sync"
	"time"

	gerror "github.com/pkg/errors"
	"go.uber.org/zap"
)

type MessageHandler func(context.Context, *event.Message) error

// TaskWorkerPool
// 主要用于管理任务提交，保证在主线程shutdown时能够排空所有待执行任务
type TaskWorkerPool struct {
	// config
	workerCount int

	// log
	log *clog.Log

	// events channel (with buffers)
	eventMessage chan *event.Message

	// lifecycle controller
	wg     *sync.WaitGroup // 等待所有worker完成
	ctx    context.Context
	cancel context.CancelFunc

	handler MessageHandler

	// shutdown flag
	closed     bool
	closedLock sync.Mutex
}

func NewTaskWorkerPool(
	queueSize int,
	workerCount int,
	log *clog.Log,
	handler MessageHandler) *TaskWorkerPool {

	ctx, cancel := context.WithCancel(context.Background())

	return &TaskWorkerPool{
		workerCount:  workerCount,
		log:          log,
		eventMessage: make(chan *event.Message, queueSize),
		wg:           &sync.WaitGroup{},
		ctx:          ctx,
		cancel:       cancel,
		handler:      handler,
	}
}

func (pool *TaskWorkerPool) SetHandler(handler MessageHandler) {
	pool.handler = handler
}

func (pool *TaskWorkerPool) Start() {
	for i := 0; i < pool.workerCount; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}
}

func (pool *TaskWorkerPool) Submit(msg *event.Message) error {
	pool.closedLock.Lock()
	if pool.closed {
		pool.closedLock.Unlock()
		return nil // Pool is closed, silently ignore
	}
	pool.closedLock.Unlock()

	// 使用 select 添加关闭检测
	select {
	case pool.eventMessage <- msg:
		return nil
	case <-pool.ctx.Done():
		// Pool is shutting down, silently ignore
		return nil
	default:
		return gerror.Errorf("task queue is full")
	}
}

func (pool *TaskWorkerPool) Shutdown(timeout time.Duration) error {
	// 先设置 closed 标志，这样 Submit 会立即返回
	pool.closedLock.Lock()
	if pool.closed {
		pool.closedLock.Unlock()
		return nil // Already closed
	}
	pool.closed = true
	pool.closedLock.Unlock()

	// 1. 关闭队列，表示不再接收新任务
	close(pool.eventMessage)

	// 2. 等待所有 worker 处理完队列中的剩余任务
	done := make(chan struct{})
	go func() {
		pool.wg.Wait()
		close(done)
	}()

	// 3. 设置超时，防止无限等待
	select {
	case <-done:
		pool.log.Info("worker pool shut down gracefully")
		return nil
	case <-time.After(timeout):
		pool.log.Warn("worker pool shutdown timeout, force cancel")
		// 强制取消正在处理的任务上下文
		pool.cancel()
		return gerror.Errorf("shutdown timeout")
	}
}

func (pool *TaskWorkerPool) worker() {
	defer pool.wg.Done()

	for {
		select {
		case msg, ok := <-pool.eventMessage:
			if !ok {
				return
			}

			// process msg
			// 这里使用不会cancel的context，防止因为pool的ctx关闭导致不能继续处理任务
			if err := pool.handler(context.Background(), msg); err != nil {
				pool.log.Error("fail to handler message", zap.Any("message", msg), zap.Error(err))
			}

		case <-pool.ctx.Done():
			// 收到取消信号，但需要继续处理 channel 中剩余任务
			// 这里不立即退出，而是继续从 channel 取任务直到 channel 关闭并排空
			pool.log.Info("pool worker receive ctx done signal...")
		}
	}
}
