package storagemgr

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/housekeeping"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/localrepo"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/storage"
	"gitlab.com/gitlab-org/gitaly/v16/internal/helper/perm"
	"gitlab.com/gitlab-org/gitaly/v16/internal/safe"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
)

// ErrPartitionManagerStopped is returned when the PartitionManager stops processing transactions.
var ErrPartitionManagerStopped = errors.New("partition manager stopped")

// PartitionManager is responsible for managing the lifecycle of each TransactionManager.
type PartitionManager struct {
	// storages are the storages configured in this Gitaly server. The map is keyed by the storage name.
	storages map[string]*storageManager
	// commandFactory is passed as a dependency to the constructed TransactionManagers.
	commandFactory git.CommandFactory
	// housekeepingManager access to the housekeeping.Manager.
	housekeepingManager housekeeping.Manager
}

// storageManager represents a single storage.
type storageManager struct {
	// mu synchronizes access to the fields of storageManager.
	mu sync.Mutex
	// logger handles all logging for storageManager.
	logger logrus.FieldLogger
	// path is the absolute path to the storage's root.
	path string
	// repoFactory is a factory type that builds localrepo instances for this storage.
	repoFactory localrepo.StorageScopedFactory
	// stagingDirectory is the directory where all of the TransactionManager staging directories
	// should be created.
	stagingDirectory string
	// stopped tracks whether the storageManager has been stopped. If it is stopped,
	// no new transactions are allowed to begin.
	stopped bool
	// db is the handle to the key-value store used for storing the storage's database state.
	database *badger.DB
	// partitions contains all the active partitions. Each repository can have up to one partition.
	partitions map[string]*partition
	// activePartitions keeps track of active partitions.
	activePartitions sync.WaitGroup
}

func (sm *storageManager) stop() {
	sm.mu.Lock()
	// Mark the storage as stopped so no new transactions can begin anymore. This
	// also means no more partitions are spawned.
	sm.stopped = true
	for _, ptn := range sm.partitions {
		// Stop all partitions.
		ptn.stop()
	}
	sm.mu.Unlock()

	// Wait for all partitions to finish.
	sm.activePartitions.Wait()

	if err := sm.database.Close(); err != nil {
		sm.logger.WithError(err).Error("failed closing storage's database")
	}
}

// finalizeTransaction decrements the partition's pending transaction count and stops it if there are no more
// transactions pending.
func (sm *storageManager) finalizeTransaction(ptn *partition) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	ptn.pendingTransactionCount--
	if ptn.pendingTransactionCount == 0 {
		ptn.stop()
	}
}

// finalizableTransaction wraps a transaction to track the number of in-flight transactions for a Partition.
type finalizableTransaction struct {
	// finalize is called when the transaction is either committed or rolled back.
	finalize func()
	// Transaction is the underlying transaction.
	*Transaction
}

// Commit commits the transaction and runs the finalizer.
func (tx *finalizableTransaction) Commit(ctx context.Context) error {
	defer tx.finalize()
	return tx.Transaction.Commit(ctx)
}

// Rollback rolls back the transaction and runs the finalizer.
func (tx *finalizableTransaction) Rollback() error {
	defer tx.finalize()
	return tx.Transaction.Rollback()
}

// newFinalizableTransaction returns a wrapped transaction that executes finalizeTransaction when the transaction
// is committed or rolled back.
func (sm *storageManager) newFinalizableTransaction(ptn *partition, tx *Transaction) *finalizableTransaction {
	finalized := false
	return &finalizableTransaction{
		finalize: func() {
			if finalized {
				return
			}

			finalized = true
			sm.finalizeTransaction(ptn)
		},
		Transaction: tx,
	}
}

// partition contains the transaction manager and tracks the number of in-flight transactions for the partition.
type partition struct {
	// shuttingDown is closed when the partition has no longer any active transactions.
	shuttingDown chan struct{}
	// shutdown is closed to signal when the partition is finished shutting down. Clients stumbling on the
	// partition when it is shutting down wait on this channel to know when the partition has shut down and they
	// should retry.
	shutdown chan struct{}
	// transactionManager manages all transactions for the partition.
	transactionManager *TransactionManager
	// pendingTransactionCount holds the current number of in flight transactions being processed by the manager.
	pendingTransactionCount uint
}

// stop stops the partition's transaction manager.
func (ptn *partition) stop() {
	close(ptn.shuttingDown)
	ptn.transactionManager.Stop()
}

// isStopping returns whether partition shutdown has been initiated.
func (ptn *partition) isStopping() bool {
	select {
	case <-ptn.shuttingDown:
		return true
	default:
		return false
	}
}

// NewPartitionManager returns a new PartitionManager.
func NewPartitionManager(
	configuredStorages []config.Storage,
	cmdFactory git.CommandFactory,
	housekeepingManager housekeeping.Manager,
	localRepoFactory localrepo.Factory,
	logger logrus.FieldLogger,
) (*PartitionManager, error) {
	storages := make(map[string]*storageManager, len(configuredStorages))
	for _, storage := range configuredStorages {
		repoFactory, err := localRepoFactory.ScopeByStorage(storage.Name)
		if err != nil {
			return nil, fmt.Errorf("scope by storage: %w", err)
		}

		stagingDir := stagingDirectoryPath(storage.Path)
		// Remove a possible already existing staging directory as it may contain stale files
		// if the previous process didn't shutdown gracefully.
		if err := os.RemoveAll(stagingDir); err != nil {
			return nil, fmt.Errorf("failed clearing storage's staging directory: %w", err)
		}

		if err := os.Mkdir(stagingDir, perm.PrivateDir); err != nil {
			return nil, fmt.Errorf("create storage's staging directory: %w", err)
		}

		databaseDir := filepath.Join(storage.Path, "database")
		if err := os.Mkdir(databaseDir, perm.PrivateDir); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("create storage's database directory: %w", err)
		}

		if err := safe.NewSyncer().SyncHierarchy(storage.Path, "database"); err != nil {
			return nil, fmt.Errorf("sync database directory: %w", err)
		}

		db, err := OpenDatabase(databaseDir)
		if err != nil {
			return nil, fmt.Errorf("create storage's database directory: %w", err)
		}

		storages[storage.Name] = &storageManager{
			logger:           logger.WithField("storage", storage.Name),
			path:             storage.Path,
			repoFactory:      repoFactory,
			stagingDirectory: stagingDir,
			database:         db,
			partitions:       map[string]*partition{},
		}
	}

	return &PartitionManager{storages: storages, commandFactory: cmdFactory, housekeepingManager: housekeepingManager}, nil
}

func stagingDirectoryPath(storagePath string) string {
	return filepath.Join(storagePath, "staging")
}

// Begin gets the TransactionManager for the specified repository and starts a transaction. If a
// TransactionManager is not already running, a new one is created and used. The partition tracks
// the number of pending transactions and this counter gets incremented when Begin is invoked.
func (pm *PartitionManager) Begin(ctx context.Context, repo storage.Repository) (*finalizableTransaction, error) {
	storageMgr, ok := pm.storages[repo.GetStorageName()]
	if !ok {
		return nil, structerr.NewNotFound("unknown storage: %q", repo.GetStorageName())
	}

	relativePath, err := storage.ValidateRelativePath(storageMgr.path, repo.GetRelativePath())
	if err != nil {
		return nil, structerr.NewInvalidArgument("validate relative path: %w", err)
	}

	for {
		storageMgr.mu.Lock()
		if storageMgr.stopped {
			storageMgr.mu.Unlock()
			return nil, ErrPartitionManagerStopped
		}

		ptn, ok := storageMgr.partitions[relativePath]
		if !ok {
			ptn = &partition{
				shuttingDown: make(chan struct{}),
				shutdown:     make(chan struct{}),
			}

			stagingDir, err := os.MkdirTemp(storageMgr.stagingDirectory, "")
			if err != nil {
				storageMgr.mu.Unlock()
				return nil, fmt.Errorf("create staging directory: %w", err)
			}

			mgr := NewTransactionManager(storageMgr.database, storageMgr.path, relativePath, stagingDir, pm.commandFactory, pm.housekeepingManager, storageMgr.repoFactory)

			ptn.transactionManager = mgr

			storageMgr.partitions[relativePath] = ptn

			storageMgr.activePartitions.Add(1)
			go func() {
				logger := storageMgr.logger.WithField("partition", relativePath)

				if err := mgr.Run(); err != nil {
					logger.WithError(err).Error("partition failed")
				}

				// In the event that TransactionManager stops running, a new TransactionManager will
				// need to be started in order to continue processing transactions. The partition is
				// deleted allowing the next transaction for the repository to create a new partition
				// and TransactionManager.
				storageMgr.mu.Lock()
				delete(storageMgr.partitions, relativePath)
				storageMgr.mu.Unlock()

				close(ptn.shutdown)

				// If the TransactionManager returned due to an error, it could be that there are still
				// in-flight transactions operating on their staged state. Removing the staging directory
				// while they are active can lead to unexpected errors. Wait with the removal until they've
				// all finished, and only then remove the staging directory.
				//
				// All transactions must eventually finish, so we don't wait on a context cancellation here.
				<-ptn.shuttingDown

				if err := os.RemoveAll(stagingDir); err != nil {
					logger.WithError(err).Error("failed removing partition's staging directory")
				}

				storageMgr.activePartitions.Done()
			}()
		}

		if ptn.isStopping() {
			// If the partition is in the process of shutting down, the partition should not be
			// used. The lock is released while waiting for the partition to complete shutdown as to
			// not block other partitions from processing transactions. Once shutdown is complete, a
			// new attempt is made to get a valid partition.
			storageMgr.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-ptn.shutdown:
			}

			continue
		}

		ptn.pendingTransactionCount++
		storageMgr.mu.Unlock()

		transaction, err := ptn.transactionManager.Begin(ctx)
		if err != nil {
			// The pending transaction count needs to be decremented since the transaction is no longer
			// inflight. A transaction failing does not necessarily mean the transaction manager has
			// stopped running. Consequently, if there are no other pending transactions the partition
			// should be stopped.
			storageMgr.finalizeTransaction(ptn)

			return nil, err
		}

		return storageMgr.newFinalizableTransaction(ptn, transaction), nil
	}
}

// Stop stops transaction processing for all storages and waits for shutdown completion.
func (pm *PartitionManager) Stop() {
	var activeStorages sync.WaitGroup
	for _, storageMgr := range pm.storages {
		activeStorages.Add(1)
		storageMgr := storageMgr
		go func() {
			storageMgr.stop()
			activeStorages.Done()
		}()
	}

	activeStorages.Wait()
}
