package tests

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"backend/core-server/internal/rpc/likepb"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	likeGRPCAddr   = "127.0.0.1:8081"
	likeObjectType = "article"
	likeObjectID   = "test_concurrent_like_object"
	likeOwnerID    = "owner_test_001"
	concurrentN    = 100
)

func newLikeClient(t *testing.T) (likepb.LikeServiceClient, func()) {
	t.Helper()

	conn, err := grpc.NewClient(
		likeGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err, "dial %s failed，请先启动 core-server", likeGRPCAddr)

	return likepb.NewLikeServiceClient(conn), func() { _ = conn.Close() }
}

// TestConcurrentThumbUp_100Users 100 个不同用户并发点赞同一对象。
func TestConcurrentThumbUp_100Users(t *testing.T) {
	client, cleanup := newLikeClient(t)
	defer cleanup()

	objectID := fmt.Sprintf("%s_thumbup_%d", likeObjectID, time.Now().UnixNano())

	var (
		success int64
		failed  int64
		wg      sync.WaitGroup
	)

	wg.Add(concurrentN)
	for i := 0; i < concurrentN; i++ {
		i := i
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			userID := fmt.Sprintf("concurrent_thumbup_user_%03d", i)
			resp, err := client.ThumbUp(ctx, &likepb.ThumbUpRequest{
				UserID:        userID,
				ObjectType:    likeObjectType,
				ObjectID:      objectID,
				ObjectOwnerID: likeOwnerID,
			})
			if err != nil {
				atomic.AddInt64(&failed, 1)
				t.Logf("ThumbUp user=%s err=%v", userID, err)
				return
			}
			if !resp.GetSuccess() {
				atomic.AddInt64(&failed, 1)
				t.Logf("ThumbUp user=%s success=false", userID)
				return
			}
			atomic.AddInt64(&success, 1)
		}()
	}
	wg.Wait()

	t.Logf("ThumbUp done: success=%d failed=%d objectID=%s", success, failed, objectID)
	require.Equal(t, int64(concurrentN), success, "expected all %d concurrent likes to succeed", concurrentN)
	require.Zero(t, failed)
}

// TestConcurrentCancelThumbUp_100Users 先为 100 个用户点赞，再并发取消。
func TestConcurrentCancelThumbUp_100Users(t *testing.T) {
	client, cleanup := newLikeClient(t)
	defer cleanup()

	objectID := fmt.Sprintf("%s_cancel_%d", likeObjectID, time.Now().UnixNano())

	// 准备：串行点赞，保证取消前状态确定
	for i := 0; i < concurrentN; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userID := fmt.Sprintf("concurrent_cancel_user_%03d", i)
		_, err := client.ThumbUp(ctx, &likepb.ThumbUpRequest{
			UserID:        userID,
			ObjectType:    likeObjectType,
			ObjectID:      objectID,
			ObjectOwnerID: likeOwnerID,
		})
		cancel()
		require.NoError(t, err, "prepare ThumbUp user=%s", userID)
	}

	var (
		success int64
		failed  int64
		wg      sync.WaitGroup
	)

	wg.Add(concurrentN)
	for i := 0; i < concurrentN; i++ {
		i := i
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			userID := fmt.Sprintf("concurrent_cancel_user_%03d", i)
			resp, err := client.CancelThumbUp(ctx, &likepb.CancelThumbUpRequest{
				UserID:        userID,
				ObjectType:    likeObjectType,
				ObjectID:      objectID,
				ObjectOwnerID: likeOwnerID,
			})
			if err != nil {
				atomic.AddInt64(&failed, 1)
				t.Logf("CancelThumbUp user=%s err=%v", userID, err)
				return
			}
			if !resp.GetSuccess() {
				atomic.AddInt64(&failed, 1)
				t.Logf("CancelThumbUp user=%s success=false", userID)
				return
			}
			atomic.AddInt64(&success, 1)
		}()
	}
	wg.Wait()

	t.Logf("CancelThumbUp done: success=%d failed=%d objectID=%s", success, failed, objectID)
	require.Equal(t, int64(concurrentN), success, "expected all %d concurrent cancels to succeed", concurrentN)
	require.Zero(t, failed)
}

// TestConcurrentThumbUp_SameUser 同一用户并发 100 次点赞：仅 1 次成功，其余应失败（已点赞）。
func TestConcurrentThumbUp_SameUser(t *testing.T) {
	client, cleanup := newLikeClient(t)
	defer cleanup()

	userID := fmt.Sprintf("same_user_%d", time.Now().UnixNano())
	objectID := fmt.Sprintf("%s_same_user", likeObjectID)

	var (
		success int64
		failed  int64
		wg      sync.WaitGroup
	)

	wg.Add(concurrentN)
	for i := 0; i < concurrentN; i++ {
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := client.ThumbUp(ctx, &likepb.ThumbUpRequest{
				UserID:        userID,
				ObjectType:    likeObjectType,
				ObjectID:      objectID,
				ObjectOwnerID: likeOwnerID,
			})
			if err != nil {
				atomic.AddInt64(&failed, 1)
				return
			}
			atomic.AddInt64(&success, 1)
		}()
	}
	wg.Wait()

	t.Logf("same-user ThumbUp: success=%d failed=%d", success, failed)
	require.Equal(t, int64(1), success, "only one like should succeed for the same user/object")
	require.Equal(t, int64(concurrentN-1), failed)

	// 清理：取消点赞
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.CancelThumbUp(ctx, &likepb.CancelThumbUpRequest{
		UserID:        userID,
		ObjectType:    likeObjectType,
		ObjectID:      objectID,
		ObjectOwnerID: likeOwnerID,
	})
	require.NoError(t, err)
}

// TestThumbUpThenCancel 单用户：点赞 → 再点赞失败 → 取消 → 再取消失败。
func TestThumbUpThenCancel(t *testing.T) {
	client, cleanup := newLikeClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userID := fmt.Sprintf("single_user_%d", time.Now().UnixNano())
	objectID := fmt.Sprintf("%s_single", likeObjectID)

	resp, err := client.ThumbUp(ctx, &likepb.ThumbUpRequest{
		UserID:        userID,
		ObjectType:    likeObjectType,
		ObjectID:      objectID,
		ObjectOwnerID: likeOwnerID,
	})
	require.NoError(t, err)
	require.True(t, resp.GetSuccess())

	_, err = client.ThumbUp(ctx, &likepb.ThumbUpRequest{
		UserID:        userID,
		ObjectType:    likeObjectType,
		ObjectID:      objectID,
		ObjectOwnerID: likeOwnerID,
	})
	require.Error(t, err, "second ThumbUp should fail with already liked")

	cancelResp, err := client.CancelThumbUp(ctx, &likepb.CancelThumbUpRequest{
		UserID:        userID,
		ObjectType:    likeObjectType,
		ObjectID:      objectID,
		ObjectOwnerID: likeOwnerID,
	})
	require.NoError(t, err)
	require.True(t, cancelResp.GetSuccess())

	_, err = client.CancelThumbUp(ctx, &likepb.CancelThumbUpRequest{
		UserID:        userID,
		ObjectType:    likeObjectType,
		ObjectID:      objectID,
		ObjectOwnerID: likeOwnerID,
	})
	require.Error(t, err, "second CancelThumbUp should fail with like not exists")
}
