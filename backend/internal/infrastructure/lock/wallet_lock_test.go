package lock_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"vault/internal/infrastructure/lock"

	"github.com/stretchr/testify/require"
)

func TestLock_Success(t *testing.T) {
	mgr := lock.NewWalletLockManager()
	ctx := context.Background()

	unlock, err := mgr.Lock(ctx, 1, time.Second)
	require.NoError(t, err)
	require.NotNil(t, unlock)

	unlock()
}

func TestLock_BlocksSecondGoroutine(t *testing.T) {
	mgr := lock.NewWalletLockManager()
	ctx := context.Background()

	unlock, err := mgr.Lock(ctx, 42, time.Second)
	require.NoError(t, err)

	var secondAcquired atomic.Bool
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		unlock2, err := mgr.Lock(ctx, 42, time.Second)
		if err == nil {
			secondAcquired.Store(true)
			unlock2()
		}
	}()

	time.Sleep(50 * time.Millisecond)
	require.False(t, secondAcquired.Load(), "second goroutine must not acquire lock while first holds it")

	unlock()
	wg.Wait()
	require.True(t, secondAcquired.Load(), "second goroutine should acquire lock after first releases")
}

func TestLock_Timeout(t *testing.T) {
	mgr := lock.NewWalletLockManager()
	ctx := context.Background()

	unlock, err := mgr.Lock(ctx, 1, time.Second)
	require.NoError(t, err)
	defer unlock()

	start := time.Now()
	_, err = mgr.Lock(ctx, 1, 100*time.Millisecond)
	elapsed := time.Since(start)

	require.ErrorIs(t, err, lock.ErrLockTimeout)
	require.InDelta(t, 100, elapsed.Milliseconds(), 50)
}

func TestLock_ContextCanceled(t *testing.T) {
	mgr := lock.NewWalletLockManager()
	ctx := context.Background()

	unlock, err := mgr.Lock(ctx, 1, time.Second)
	require.NoError(t, err)
	defer unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = mgr.Lock(ctx, 1, time.Second)
	require.ErrorIs(t, err, context.Canceled)
}

func TestLock_MapCleanup(t *testing.T) {
	mgr := lock.NewWalletLockManager()
	ctx := context.Background()

	unlock, err := mgr.Lock(ctx, 1, time.Second)
	require.NoError(t, err)
	unlock()

	mgrMu := mgr.Mutex()
	mgrMu.Lock()
	_, exists := mgr.WalletsMap()[1]
	mgrMu.Unlock()
	require.False(t, exists, "entry must be removed after unlock")

	unlock2, err := mgr.Lock(ctx, 1, time.Second)
	require.NoError(t, err)
	defer unlock2()

	mgrMu.Lock()
	_, exists = mgr.WalletsMap()[1]
	mgrMu.Unlock()
	require.True(t, exists, "new entry must be created on re-lock")
}
