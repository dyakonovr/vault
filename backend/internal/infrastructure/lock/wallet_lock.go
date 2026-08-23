package lock

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrLockTimeout = errors.New("wallet lock timeout error")

type walletLockEntry struct {
	mu   *sync.Mutex
	refs int64 // сколько горутин ждут мьютекс. Нужна для очистки мапы
}

type WalletLockManager struct {
	mu         *sync.Mutex
	walletsMap map[int64]*walletLockEntry
}

func NewWalletLockManager() *WalletLockManager {
	return &WalletLockManager{
		mu:         &sync.Mutex{},
		walletsMap: make(map[int64]*walletLockEntry),
	}
}

func (m *WalletLockManager) Lock(ctx context.Context, walletID int64, timeout time.Duration) (func(), error) {
	m.mu.Lock()

	walletEntry, ok := m.walletsMap[walletID]
	if !ok {
		walletEntry = &walletLockEntry{
			mu: &sync.Mutex{},
			refs: 0,
		}
		m.walletsMap[walletID] = walletEntry
	}
	walletEntry.refs += 1
	m.mu.Unlock()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			m.releaseRef(walletEntry, walletID)
			return nil, ctx.Err()
		}

		if walletEntry.mu.TryLock() {
			once := sync.Once{}

			return func() {
				once.Do(func() {
					walletEntry.mu.Unlock()
					m.releaseRef(walletEntry, walletID)
				})
			}, nil
		}

		time.Sleep(10 * time.Millisecond)
	}

	m.releaseRef(walletEntry, walletID)
	return nil, ErrLockTimeout
}

// func (m *WalletLockManager) LockCouple(ctx context.Context, walletID1, walletID2 int64, timeout time.Duration) (func(), error) {
// 	wallet1First := walletID1 > walletID2
	

// }

func (m *WalletLockManager) releaseRef(entry *walletLockEntry, walletID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry.refs--
	if entry.refs <= 0 {
		delete(m.walletsMap, walletID)
	}
}

func (m *WalletLockManager) Mutex() *sync.Mutex         { return m.mu }
func (m *WalletLockManager) WalletsMap() map[int64]*walletLockEntry { return m.walletsMap }