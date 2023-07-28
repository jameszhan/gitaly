package storagemgr

import (
	"bytes"
	"container/list"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/housekeeping"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/localrepo"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/updateref"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/repoutil"
	"gitlab.com/gitlab-org/gitaly/v16/internal/helper/perm"
	"gitlab.com/gitlab-org/gitaly/v16/internal/safe"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrRepositoryNotFound is returned when the repository doesn't exist.
	ErrRepositoryNotFound = structerr.NewNotFound("repository not found")
	// ErrTransactionProcessingStopped is returned when the TransactionManager stops processing transactions.
	ErrTransactionProcessingStopped = errors.New("transaction processing stopped")
	// ErrTransactionAlreadyCommitted is returned when attempting to rollback or commit a transaction that
	// already had commit called on it.
	ErrTransactionAlreadyCommitted = errors.New("transaction already committed")
	// ErrTransactionAlreadyRollbacked is returned when attempting to rollback or commit a transaction that
	// already had rollback called on it.
	ErrTransactionAlreadyRollbacked = errors.New("transaction already rollbacked")
	// errInitializationFailed is returned when the TransactionManager failed to initialize successfully.
	errInitializationFailed = errors.New("initializing transaction processing failed")
	// errNotDirectory is returned when the repository's path doesn't point to a directory
	errNotDirectory = errors.New("repository's path didn't point to a directory")
)

// InvalidReferenceFormatError is returned when a reference name was invalid.
type InvalidReferenceFormatError struct {
	// ReferenceName is the reference with invalid format.
	ReferenceName git.ReferenceName
}

// Error returns the formatted error string.
func (err InvalidReferenceFormatError) Error() string {
	return fmt.Sprintf("invalid reference format: %q", err.ReferenceName)
}

// ReferenceVerificationError is returned when a reference's old OID did not match the expected.
type ReferenceVerificationError struct {
	// ReferenceName is the name of the reference that failed verification.
	ReferenceName git.ReferenceName
	// ExpectedOID is the OID the reference was expected to point to.
	ExpectedOID git.ObjectID
	// ActualOID is the OID the reference actually pointed to.
	ActualOID git.ObjectID
}

// Error returns the formatted error string.
func (err ReferenceVerificationError) Error() string {
	return fmt.Sprintf("expected %q to point to %q but it pointed to %q", err.ReferenceName, err.ExpectedOID, err.ActualOID)
}

// LogIndex points to a specific position in a repository's write-ahead log.
type LogIndex uint64

// toProto returns the protobuf representation of LogIndex for serialization purposes.
func (index LogIndex) toProto() *gitalypb.LogIndex {
	return &gitalypb.LogIndex{LogIndex: uint64(index)}
}

// String returns a string representation of the LogIndex.
func (index LogIndex) String() string {
	return strconv.FormatUint(uint64(index), 10)
}

// ReferenceUpdate describes the state of a reference's old and new tip in an update.
type ReferenceUpdate struct {
	// Force indicates this is a forced reference update. If set, the reference is pointed
	// to the new value regardless of the old value.
	Force bool
	// OldOID is the old OID the reference is expected to point to prior to updating it.
	// If the reference does not point to the old value, the reference verification fails.
	OldOID git.ObjectID
	// NewOID is the new desired OID to point the reference to.
	NewOID git.ObjectID
}

// DefaultBranchUpdate provides the information to update the default branch of the repo.
type DefaultBranchUpdate struct {
	// Reference is the reference to update the default branch to.
	Reference git.ReferenceName
}

// CustomHooksUpdate models an update to the custom hooks.
type CustomHooksUpdate struct {
	// CustomHooksTAR contains the custom hooks as a TAR. The TAR contains a `custom_hooks`
	// directory which contains the hooks. Setting the update with nil `custom_hooks_tar` clears
	// the hooks from the repository.
	CustomHooksTAR []byte
}

// ReferenceUpdates contains references to update. Reference name is used as the key and the value
// is the expected old tip and the desired new tip.
type ReferenceUpdates map[git.ReferenceName]ReferenceUpdate

// Snapshot contains the read snapshot details of a Transaction.
type Snapshot struct {
	// ReadIndex is the index of the log entry this Transaction is reading the data at.
	ReadIndex LogIndex
	// CustomHookIndex is index of the custom hooks on the disk that are included in this Transactions's
	// snapshot and were the latest on the read index.
	CustomHookIndex LogIndex
	// CustomHookPath is an absolute filesystem path to the custom hooks in this snapshot.
	CustomHookPath string
}

type transactionState int

const (
	// transactionStateOpen indicates the transaction is open, and hasn't been committed or rolled back yet.
	transactionStateOpen = transactionState(iota)
	// transactionStateRollback indicates the transaction has been rolled back.
	transactionStateRollback
	// transactionStateCommit indicates the transaction has already been committed.
	transactionStateCommit
)

// Transaction is a unit-of-work that contains reference changes to perform on the repository.
type Transaction struct {
	// state records whether the transaction is still open. Transaction is open until either Commit()
	// or Rollback() is called on it.
	state transactionState
	// commit commits the Transaction through the TransactionManager.
	commit func(context.Context, *Transaction) error
	// result is where the outcome of the transaction is sent ot by TransactionManager once it
	// has been determined.
	result chan error
	// admitted is closed when the transaction was admitted for processing in the TransactionManager.
	// Transaction queues in admissionQueue to be committed, and is considered admitted once it has
	// been dequeued by TransactionManager.Run(). Once the transaction is admitted, its ownership moves
	// from the client goroutine to the TransactionManager.Run() goroutine, and the client goroutine must
	// not do any modifications to the state of the transcation anymore to avoid races.
	admitted chan struct{}
	// finish cleans up the transaction releasing the resources associated with it. It must be called
	// once the transaction is done with.
	finish func() error
	// finished is closed when the transaction has been finished. This enables waiting on transactions
	// to finish where needed.
	finished chan struct{}

	// stagingDirectory is the directory where the transaction stages its files prior
	// to them being logged. It is cleaned up when the transaction finishes.
	stagingDirectory string
	// quarantineDirectory is the directory within the stagingDirectory where the new objects of the
	// transaction are quarantined.
	quarantineDirectory string
	// packPrefix contains the prefix (`pack-<digest>`) of the transaction's pack if the transaction
	// had objects to log.
	packPrefix string
	// stagingRepository is a repository that is used to stage the transaction. If there are quarantined
	// objects, it has the quarantine applied so the objects are available for verification and packing.
	stagingRepository *localrepo.Repo

	// Snapshot contains the details of the Transaction's read snapshot.
	snapshot Snapshot

	skipVerificationFailures bool
	referenceUpdates         ReferenceUpdates
	defaultBranchUpdate      *DefaultBranchUpdate
	customHooksUpdate        *CustomHooksUpdate
	deleteRepository         bool
	includedObjects          map[git.ObjectID]struct{}
}

// Begin opens a new transaction. The caller must call either Commit or Rollback to release
// the resources tied to the transaction. The returned Transaction is not safe for concurrent use.
//
// The returned Transaction's read snapshot includes all writes that were committed prior to the
// Begin call. Begin blocks until the committed writes have been applied to the repository.
func (mgr *TransactionManager) Begin(ctx context.Context) (_ *Transaction, returnedErr error) {
	// Wait until the manager has been initialized so the notification channels
	// and the log indexes are loaded.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mgr.initialized:
		if !mgr.initializationSuccessful {
			return nil, errInitializationFailed
		}
	}

	mgr.mutex.Lock()

	txn := &Transaction{
		commit: mgr.commit,
		snapshot: Snapshot{
			ReadIndex:       mgr.appendedLogIndex,
			CustomHookIndex: mgr.customHookIndex,
			CustomHookPath:  customHookPathForLogIndex(mgr.repositoryPath, mgr.customHookIndex),
		},
		admitted: make(chan struct{}),
		finished: make(chan struct{}),
	}

	// If there are no custom hooks stored through the WAL yet, then default to the custom hooks
	// that may already exist in the repository for backwards compatibility.
	if txn.snapshot.CustomHookIndex == 0 {
		txn.snapshot.CustomHookPath = filepath.Join(mgr.repositoryPath, repoutil.CustomHooksDir)
	}

	openTransactionElement := mgr.openTransactions.PushBack(txn)

	readReady := mgr.applyNotifications[txn.snapshot.ReadIndex]
	repositoryExists := mgr.repositoryExists
	mgr.mutex.Unlock()
	if readReady == nil {
		// The snapshot log entry is already applied if there is no notification channel for it.
		// If so, the transaction is ready to begin immediately.
		readReady = make(chan struct{})
		close(readReady)
	}

	txn.finish = func() error {
		defer close(txn.finished)

		mgr.mutex.Lock()
		mgr.openTransactions.Remove(openTransactionElement)
		mgr.mutex.Unlock()

		if txn.stagingDirectory != "" {
			if err := os.RemoveAll(txn.stagingDirectory); err != nil {
				return fmt.Errorf("remove staging directory: %w", err)
			}
		}

		return nil
	}

	defer func() {
		if returnedErr != nil {
			if err := txn.finish(); err != nil {
				ctxlogrus.Extract(ctx).WithError(err).Error("failed finishing unsuccessful transaction begin")
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-mgr.ctx.Done():
		return nil, ErrTransactionProcessingStopped
	case <-readReady:
		if !repositoryExists {
			return nil, ErrRepositoryNotFound
		}

		var err error
		txn.stagingDirectory, err = os.MkdirTemp(mgr.stagingDirectory, "")
		if err != nil {
			return nil, fmt.Errorf("mkdir temp: %w", err)
		}

		txn.quarantineDirectory = filepath.Join(txn.stagingDirectory, "quarantine")
		if err := os.MkdirAll(filepath.Join(txn.quarantineDirectory, "pack"), perm.PrivateDir); err != nil {
			return nil, fmt.Errorf("create quarantine directory: %w", err)
		}

		if err := mgr.setupStagingRepository(ctx, txn); err != nil {
			return nil, fmt.Errorf("setup staging repository: %w", err)
		}

		return txn, nil
	}
}

// RewriteRepository returns a copy of the repository that has been set up to correctly access the
// transaction's repository.
func (txn *Transaction) RewriteRepository(repo *gitalypb.Repository) *gitalypb.Repository {
	rewritten := proto.Clone(repo).(*gitalypb.Repository)
	rewritten.RelativePath = txn.stagingRepository.GetRelativePath()
	rewritten.GitObjectDirectory = txn.stagingRepository.GetGitObjectDirectory()
	rewritten.GitAlternateObjectDirectories = txn.stagingRepository.GetGitAlternateObjectDirectories()
	return rewritten
}

func (txn *Transaction) updateState(newState transactionState) error {
	switch txn.state {
	case transactionStateOpen:
		txn.state = newState
		return nil
	case transactionStateRollback:
		return ErrTransactionAlreadyRollbacked
	case transactionStateCommit:
		return ErrTransactionAlreadyCommitted
	default:
		return fmt.Errorf("unknown transaction state: %q", txn.state)
	}
}

// Commit performs the changes. If no error is returned, the transaction was successful and the changes
// have been performed. If an error was returned, the transaction may or may not be persisted.
func (txn *Transaction) Commit(ctx context.Context) (returnedErr error) {
	if err := txn.updateState(transactionStateCommit); err != nil {
		return err
	}

	defer func() {
		if err := txn.finishUnadmitted(); err != nil && returnedErr == nil {
			returnedErr = err
		}
	}()

	return txn.commit(ctx, txn)
}

// Rollback releases resources associated with the transaction without performing any changes.
func (txn *Transaction) Rollback() error {
	if err := txn.updateState(transactionStateRollback); err != nil {
		return err
	}

	return txn.finishUnadmitted()
}

// finishUnadmitted cleans up after the transaction if it wasn't yet admitted. If the transaction was admitted,
// the Transaction is being processed by TransactionManager. The clean up responsibility moves there as well
// to avoid races.
func (txn *Transaction) finishUnadmitted() error {
	select {
	case <-txn.admitted:
		return nil
	default:
		return txn.finish()
	}
}

// Snapshot returns the details of the Transaction's read snapshot.
func (txn *Transaction) Snapshot() Snapshot {
	return txn.snapshot
}

// SkipVerificationFailures configures the transaction to skip reference updates that fail verification.
// If a reference update fails verification with this set, the update is dropped from the transaction but
// other successful reference updates will be made. By default, the entire transaction is aborted if a
// reference fails verification.
//
// The default behavior models `git push --atomic`. Toggling this option models the behavior without
// the `--atomic` flag.
func (txn *Transaction) SkipVerificationFailures() {
	txn.skipVerificationFailures = true
}

// UpdateReferences updates the given references as part of the transaction. If UpdateReferences is called
// multiple times, only the changes from the latest invocation take place.
func (txn *Transaction) UpdateReferences(updates ReferenceUpdates) {
	txn.referenceUpdates = updates
}

// DeleteRepository deletes the repository when the transaction is committed.
func (txn *Transaction) DeleteRepository() {
	txn.deleteRepository = true
}

// SetDefaultBranch sets the default branch as part of the transaction. If SetDefaultBranch is called
// multiple times, only the changes from the latest invocation take place. The reference is validated
// to exist.
func (txn *Transaction) SetDefaultBranch(new git.ReferenceName) {
	txn.defaultBranchUpdate = &DefaultBranchUpdate{Reference: new}
}

// SetCustomHooks sets the custom hooks as part of the transaction. If SetCustomHooks is called multiple
// times, only the changes from the latest invocation take place. The custom hooks are extracted as is and
// are not validated. Setting a nil hooksTAR removes the hooks from the repository.
func (txn *Transaction) SetCustomHooks(customHooksTAR []byte) {
	txn.customHooksUpdate = &CustomHooksUpdate{CustomHooksTAR: customHooksTAR}
}

// IncludeObject includes the given object and its dependencies in the transaction's logged pack file even
// if the object is unreachable from the references.
func (txn *Transaction) IncludeObject(oid git.ObjectID) {
	if txn.includedObjects == nil {
		txn.includedObjects = map[git.ObjectID]struct{}{}
	}

	txn.includedObjects[oid] = struct{}{}
}

// walFilesPath returns the path to the directory where this transaction is staging the files that will
// be logged alongside the transaction's log entry.
func (txn *Transaction) walFilesPath() string {
	return filepath.Join(txn.stagingDirectory, "wal-files")
}

// TransactionManager is responsible for transaction management of a single repository. Each repository has
// a single TransactionManager; it is the repository's single-writer. It accepts writes one at a time from
// the admissionQueue. Each admitted write is processed in three steps:
//
//  1. The references being updated are verified by ensuring the expected old tips match what the references
//     actually point to prior to update. The entire transaction is by default aborted if a single reference
//     fails the verification step. The reference verification behavior can be controlled on a per-transaction
//     level by setting:
//     - The reference verification failures can be ignored instead of aborting the entire transaction.
//     If done, the references that failed verification are dropped from the transaction but the updates
//     that passed verification are still performed.
//     - The reference verification may also be skipped if the write is force updating references. If
//     done, the current state of the references is ignored and they are directly updated to point
//     to the new tips.
//  2. The transaction is appended to the write-ahead log. Once the write has been logged, it is effectively
//     committed and will be applied to the repository even after restarting.
//  3. The transaction is applied from the write-ahead log to the repository by actually performing the reference
//     changes.
//
// The goroutine that issued the transaction is waiting for the result while these steps are being performed. As
// there is no transaction control for readers yet, the issuer is only notified of a successful write after the
// write has been applied to the repository.
//
// TransactionManager recovers transactions after interruptions by applying the write-ahead logged transactions to
// the repository on start up.
//
// TransactionManager maintains the write-ahead log in a key-value store. It maintains the following key spaces:
// - `repository/<repository_id:string>/log/index/applied`
//   - This key stores the index of the log entry that has been applied to the repository. This allows for
//     determining how far a repository is in processing the log and which log entries need to be applied
//     after starting up. Repository starts from log index 0 if there are no log entries recorded to have
//     been applied.
//
// - `repository/<repository_id:string>/log/entry/<log_index:uint64>`
//   - These keys hold the actual write-ahead log entries. A repository's first log entry starts at index 1
//     and the log index keeps monotonically increasing from there on without gaps. The write-ahead log
//     entries are processed in ascending order.
//
// The values in the database are marshaled protocol buffer messages. Numbers in the keys are encoded as big
// endian to preserve the sort order of the numbers also in lexicographical order.
type TransactionManager struct {
	// ctx is the context used for all operations.
	ctx context.Context
	// close cancels ctx and stops the transaction processing.
	close context.CancelFunc

	// closing is closed when close is called. It unblock transactions that are waiting to be admitted.
	closing <-chan struct{}
	// closed is closed when Run returns. It unblocks transactions that are waiting for a result after
	// being admitted. This is differentiated from ctx.Done in order to enable testing that Run correctly
	// releases awaiters when the transactions processing is stopped.
	closed chan struct{}
	// stagingDirectory is a path to a directory where this TransactionManager should stage the files of the transactions
	// before it logs them. The TransactionManager cleans up the files during runtime but stale files may be
	// left around after crashes. The files are temporary and any leftover files are expected to be cleaned up when
	// Gitaly starts.
	stagingDirectory string
	// commandFactory is used to spawn git commands without a repository.
	commandFactory git.CommandFactory

	// repositoryExists marks whether the repository exists or not. The repository may not exist if it has
	// never been created, or if it has been deleted.
	repositoryExists bool
	// repository is the repository this TransactionManager is acting on.
	repository *localrepo.Repo
	// repositoryPath is the path to the repository this TransactionManager is acting on.
	repositoryPath string
	// relativePath is the repository's relative path inside the storage.
	relativePath string
	// db is the handle to the key-value store used for storing the write-ahead log related state.
	db database
	// admissionQueue is where the incoming writes are waiting to be admitted to the transaction
	// manager.
	admissionQueue chan *Transaction
	// openTransactions contains all transactions that have been begun but not yet committed or rolled back.
	// The transactions are ordered from the oldest to the newest.
	openTransactions *list.List

	// initialized is closed when the manager has been initialized. It's used to block new transactions
	// from beginning prior to the manager having initialized its runtime state on start up.
	initialized chan struct{}
	// initializationSuccessful is set if the TransactionManager initialized successfully. If it didn't,
	// transactions will fail to begin.
	initializationSuccessful bool
	// mutex guards access to applyNotifications and appendedLogIndex. These fields are accessed by both
	// Run and Begin which are ran in different goroutines.
	mutex sync.Mutex
	// applyNotifications stores channels that are closed when a log entry is applied. These
	// are used to block transactions from beginning before their snapshot is ready.
	applyNotifications map[LogIndex]chan struct{}
	// appendedLogIndex holds the index of the last log entry appended to the log.
	appendedLogIndex LogIndex
	// appliedLogIndex holds the index of the last log entry applied to the repository
	appliedLogIndex LogIndex
	// customHookIndex stores the log index of the latest committed custom custom hooks in the repository.
	customHookIndex LogIndex
	// housekeepingManager access to the housekeeping.Manager.
	housekeepingManager housekeeping.Manager

	// awaitingTransactions contains transactions waiting for their log entry to be applied to
	// the repository. It's keyed by the log index the transaction is waiting to be applied and the
	// value is the resultChannel that is waiting the result.
	awaitingTransactions map[LogIndex]resultChannel
}

// NewTransactionManager returns a new TransactionManager for the given repository.
func NewTransactionManager(
	db *badger.DB,
	storagePath,
	relativePath,
	stagingDir string,
	cmdFactory git.CommandFactory,
	housekeepingManager housekeeping.Manager,
	repositoryFactory localrepo.StorageScopedFactory,
) *TransactionManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TransactionManager{
		ctx:                  ctx,
		close:                cancel,
		closing:              ctx.Done(),
		closed:               make(chan struct{}),
		commandFactory:       cmdFactory,
		repository:           repositoryFactory.Build(relativePath),
		repositoryPath:       filepath.Join(storagePath, relativePath),
		relativePath:         relativePath,
		db:                   newDatabaseAdapter(db),
		admissionQueue:       make(chan *Transaction),
		openTransactions:     list.New(),
		initialized:          make(chan struct{}),
		applyNotifications:   make(map[LogIndex]chan struct{}),
		stagingDirectory:     stagingDir,
		housekeepingManager:  housekeepingManager,
		awaitingTransactions: make(map[LogIndex]resultChannel),
	}
}

// resultChannel represents a future that will yield the result of a transaction once its
// outcome has been decided.
type resultChannel chan error

// commit queues the transaction for processing and returns once the result has been determined.
func (mgr *TransactionManager) commit(ctx context.Context, transaction *Transaction) error {
	transaction.result = make(resultChannel, 1)

	if err := mgr.stageHooks(ctx, transaction); err != nil {
		return fmt.Errorf("stage hooks: %w", err)
	}

	if err := mgr.packObjects(ctx, transaction); err != nil {
		return fmt.Errorf("pack objects: %w", err)
	}

	select {
	case mgr.admissionQueue <- transaction:
		close(transaction.admitted)

		select {
		case err := <-transaction.result:
			return unwrapExpectedError(err)
		case <-ctx.Done():
			return ctx.Err()
		case <-mgr.closed:
			return ErrTransactionProcessingStopped
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-mgr.closing:
		return ErrTransactionProcessingStopped
	}
}

// stageHooks extracts the new hooks, if any, into <stagingDirectory>/custom_hooks. This is ensures the TAR
// is valid prior to committing the transaction. The hooks files on the disk are also used to compute a vote
// for Praefect.
func (mgr *TransactionManager) stageHooks(ctx context.Context, transaction *Transaction) error {
	if transaction.customHooksUpdate == nil || len(transaction.customHooksUpdate.CustomHooksTAR) == 0 {
		return nil
	}

	if err := repoutil.ExtractHooks(
		ctx,
		bytes.NewReader(transaction.customHooksUpdate.CustomHooksTAR),
		transaction.stagingDirectory,
		false,
	); err != nil {
		return fmt.Errorf("extract hooks: %w", err)
	}

	return nil
}

// setupStagingRepository sets a repository that is used to stage the transaction. The staging repository
// has the quarantine applied so the objects are available for packing and verifying the references.
func (mgr *TransactionManager) setupStagingRepository(ctx context.Context, transaction *Transaction) error {
	quarantinedRepo, err := mgr.repository.Quarantine(transaction.quarantineDirectory)
	if err != nil {
		return fmt.Errorf("quarantine: %w", err)
	}

	transaction.stagingRepository = quarantinedRepo

	return nil
}

// packPrefixRegexp matches the output of `git index-pack` where it
// prints the packs prefix in the format `pack <digest>`.
var packPrefixRegexp = regexp.MustCompile(`^pack\t([0-9a-f]+)\n$`)

// shouldPackObjects checks whether the quarantine directory has any non-default content in it.
// If so, this signifies objects were written into it and we should pack them.
func shouldPackObjects(quarantineDirectory string) (bool, error) {
	errHasNewContent := errors.New("new content found")

	// The quarantine directory itself and the pack directory within it are created when the transaction
	// begins. These don't signify new content so we ignore them.
	preExistingDirs := map[string]struct{}{
		quarantineDirectory:                        {},
		filepath.Join(quarantineDirectory, "pack"): {},
	}
	if err := filepath.Walk(quarantineDirectory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if _, ok := preExistingDirs[path]; ok {
			// The pre-existing directories don't signal any new content to pack.
			return nil
		}

		// Use an error sentinel to cancel the walk as soon as new content has been found.
		return errHasNewContent
	}); err != nil {
		if errors.Is(err, errHasNewContent) {
			return true, nil
		}

		return false, fmt.Errorf("check for objects: %w", err)
	}

	return false, nil
}

// packObjects packs the objects included in the transaction into a single pack file that is ready
// for logging. The pack file includes all unreachable objects that are about to be made reachable and
// unreachable objects that have been explicitly included in the transaction.
func (mgr *TransactionManager) packObjects(ctx context.Context, transaction *Transaction) error {
	if shouldPack, err := shouldPackObjects(transaction.quarantineDirectory); err != nil {
		return fmt.Errorf("should pack objects: %w", err)
	} else if !shouldPack {
		return nil
	}

	objectHash, err := transaction.stagingRepository.ObjectHash(ctx)
	if err != nil {
		return fmt.Errorf("object hash: %w", err)
	}

	heads := make([]string, 0, len(transaction.referenceUpdates)+len(transaction.includedObjects))
	for _, update := range transaction.referenceUpdates {
		if update.NewOID == objectHash.ZeroOID {
			// Reference deletions can't introduce new objects so ignore them.
			continue
		}

		heads = append(heads, update.NewOID.String())
	}

	for objectID := range transaction.includedObjects {
		heads = append(heads, objectID.String())
	}

	if len(heads) == 0 {
		// No need to pack objects if there are no changes that can introduce new objects.
		return nil
	}

	objectsReader, objectsWriter := io.Pipe()

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() (returnedErr error) {
		defer func() { objectsWriter.CloseWithError(returnedErr) }()

		if err := transaction.stagingRepository.WalkUnreachableObjects(ctx,
			strings.NewReader(strings.Join(heads, "\n")),
			objectsWriter,
		); err != nil {
			return fmt.Errorf("walk objects: %w", err)
		}

		return nil
	})

	packReader, packWriter := io.Pipe()
	group.Go(func() (returnedErr error) {
		defer func() {
			objectsReader.CloseWithError(returnedErr)
			packWriter.CloseWithError(returnedErr)
		}()

		if err := transaction.stagingRepository.PackObjects(ctx, objectsReader, packWriter); err != nil {
			return fmt.Errorf("pack objects: %w", err)
		}

		return nil
	})

	group.Go(func() (returnedErr error) {
		defer packReader.CloseWithError(returnedErr)

		if err := os.Mkdir(transaction.walFilesPath(), perm.PrivateDir); err != nil {
			return fmt.Errorf("create wal files directory: %w", err)
		}

		// index-pack places the pack, index, and reverse index into the repository's object directory.
		// The staging repository is configured with a quarantine so we execute it there.
		var stdout, stderr bytes.Buffer
		if err := transaction.stagingRepository.ExecAndWait(ctx, git.Command{
			Name:  "index-pack",
			Flags: []git.Option{git.Flag{Name: "--stdin"}, git.Flag{Name: "--rev-index"}},
			Args:  []string{filepath.Join(transaction.walFilesPath(), "objects.pack")},
		}, git.WithStdin(packReader), git.WithStdout(&stdout), git.WithStderr(&stderr)); err != nil {
			return structerr.New("index pack: %w", err).WithMetadata("stderr", stderr.String())
		}

		matches := packPrefixRegexp.FindStringSubmatch(stdout.String())
		if len(matches) != 2 {
			return structerr.New("unexpected index-pack output").WithMetadata("stdout", stdout.String())
		}

		// Sync the files and the directory entries so everything is flushed to the disk prior
		// to moving on to committing the log entry. This way we only have to flush the directory
		// move when we move the staged files into the log.
		if err := safe.NewSyncer().SyncRecursive(transaction.walFilesPath()); err != nil {
			return fmt.Errorf("sync recursive: %w", err)
		}

		transaction.packPrefix = fmt.Sprintf("pack-%s", matches[1])

		return nil
	})

	return group.Wait()
}

// unwrapExpectedError unwraps expected errors that may occur and returns them directly to the caller.
func unwrapExpectedError(err error) error {
	// The manager controls its own execution context and it is canceled only when Stop is called.
	// Any context.Canceled errors returned are thus from shutting down so we report that here.
	if errors.Is(err, context.Canceled) {
		return ErrTransactionProcessingStopped
	}

	return err
}

// Run starts the transaction processing. On start up Run loads the indexes of the last appended and applied
// log entries from the database. It will then apply any transactions that have been logged but not applied
// to the repository. Once the recovery is completed, Run starts processing new transactions by verifying the
// references, logging the transaction and finally applying it to the repository. The transactions are acknowledged
// once they've been applied to the repository.
//
// Run keeps running until Stop is called or it encounters a fatal error. All transactions will error with
// ErrTransactionProcessingStopped when Run returns.
func (mgr *TransactionManager) Run() (returnedErr error) {
	defer func() {
		// On-going operations may fail with a context canceled error if the manager is stopped. This is
		// not a real error though given the manager will recover from this on restart. Swallow the error.
		if errors.Is(returnedErr, context.Canceled) {
			returnedErr = nil
		}
	}()

	// Defer the Stop in order to release all on-going Commit calls in case of error.
	defer close(mgr.closed)
	defer mgr.Close()

	if err := mgr.initialize(mgr.ctx); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	for {
		if mgr.appliedLogIndex < mgr.appendedLogIndex {
			logIndex := mgr.appliedLogIndex + 1

			if err := mgr.applyLogEntry(mgr.ctx, logIndex); err != nil {
				return fmt.Errorf("apply log entry: %w", err)
			}

			continue
		}

		if err := mgr.processTransaction(); err != nil {
			return fmt.Errorf("process transaction: %w", err)
		}
	}
}

// processTransaction waits for a transaction and processes it by verifying and
// logging it.
func (mgr *TransactionManager) processTransaction() (returnedErr error) {
	var cleanUps []func() error
	defer func() {
		for _, cleanUp := range cleanUps {
			if err := cleanUp(); err != nil && returnedErr == nil {
				returnedErr = fmt.Errorf("clean up: %w", err)
			}
		}
	}()

	var transaction *Transaction
	select {
	case transaction = <-mgr.admissionQueue:
		// The Transaction does not finish itself anymore once it has been admitted for
		// processing. This avoids the Transaction concurrently removing the staged state
		// while the manager is still operating on it. We thus need to defer its finishing.
		cleanUps = append(cleanUps, transaction.finish)
	case <-mgr.ctx.Done():
	}

	// Return if the manager was stopped. The select is indeterministic so this guarantees
	// the manager stops the processing even if there are transactions in the queue.
	if err := mgr.ctx.Err(); err != nil {
		return err
	}

	if err := func() (commitErr error) {
		if !mgr.repositoryExists {
			return ErrRepositoryNotFound
		}

		logEntry := &gitalypb.LogEntry{}

		var err error
		logEntry.ReferenceUpdates, err = mgr.verifyReferences(mgr.ctx, transaction)
		if err != nil {
			return fmt.Errorf("verify references: %w", err)
		}

		if transaction.defaultBranchUpdate != nil {
			if err := mgr.verifyDefaultBranchUpdate(mgr.ctx, transaction); err != nil {
				return fmt.Errorf("verify default branch update: %w", err)
			}

			logEntry.DefaultBranchUpdate = &gitalypb.LogEntry_DefaultBranchUpdate{
				ReferenceName: []byte(transaction.defaultBranchUpdate.Reference),
			}
		}

		if transaction.customHooksUpdate != nil {
			logEntry.CustomHooksUpdate = &gitalypb.LogEntry_CustomHooksUpdate{
				CustomHooksTar: transaction.customHooksUpdate.CustomHooksTAR,
			}
		}

		nextLogIndex := mgr.appendedLogIndex + 1
		if transaction.packPrefix != "" {
			logEntry.PackPrefix = transaction.packPrefix

			removeFiles, err := mgr.storeWALFiles(mgr.ctx, nextLogIndex, transaction)
			cleanUps = append(cleanUps, func() error {
				// The transaction's files might have been moved successfully in to the log.
				// If anything fails before the transaction is committed, the files must be removed as otherwise
				// they would occupy the slot of the next log entry. If this can't be done, the TransactionManager
				// will exit with an error. The files will be cleaned up on restart and no further processing is
				// allowed until that happens.
				if commitErr != nil {
					return removeFiles()
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("store wal files: %w", err)
			}
		}

		if transaction.deleteRepository {
			logEntry.RepositoryDeletion = &gitalypb.LogEntry_RepositoryDeletion{}
		}

		return mgr.appendLogEntry(nextLogIndex, logEntry)
	}(); err != nil {
		transaction.result <- err
		return nil
	}

	mgr.awaitingTransactions[mgr.appendedLogIndex] = transaction.result

	return nil
}

// Close stops the transaction processing causing Run to return.
func (mgr *TransactionManager) Close() { mgr.close() }

// isClosing returns whether closing of the manager was initiated.
func (mgr *TransactionManager) isClosing() bool {
	select {
	case <-mgr.closing:
		return true
	default:
		return false
	}
}

// initialize initializes the TransactionManager's state from the database. It loads the appendend and the applied
// indexes and initializes the notification channels that synchronize transaction beginning with log entry applying.
func (mgr *TransactionManager) initialize(ctx context.Context) error {
	defer close(mgr.initialized)

	var appliedLogIndex gitalypb.LogIndex
	if err := mgr.readKey(keyAppliedLogIndex(mgr.relativePath), &appliedLogIndex); err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return fmt.Errorf("read applied log index: %w", err)
	}

	mgr.appliedLogIndex = LogIndex(appliedLogIndex.LogIndex)

	// The index of the last appended log entry is determined from the indexes of the latest entry in the log and
	// the latest applied log entry. If there is a log entry, it is the latest appended log entry. If there are no
	// log entries, the latest log entry must have been applied to the repository and pruned away, meaning the index
	// of the last appended log entry is the same as the index if the last applied log entry.
	//
	// As the log indexes in the keys are encoded in big endian, the latest log entry can be found by taking
	// the first key when iterating the log entry key space in reverse.
	if err := mgr.db.View(func(txn databaseTransaction) error {
		logPrefix := keyPrefixLogEntries(mgr.relativePath)

		iterator := txn.NewIterator(badger.IteratorOptions{Reverse: true, Prefix: logPrefix})
		defer iterator.Close()

		mgr.appendedLogIndex = mgr.appliedLogIndex

		// The iterator seeks to a key that is greater than or equal than seeked key. Since we are doing a reverse
		// seek, we need to add 0xff to the prefix so the first iterated key is the latest log entry.
		if iterator.Seek(append(logPrefix, 0xff)); iterator.Valid() {
			mgr.appendedLogIndex = LogIndex(binary.BigEndian.Uint64(bytes.TrimPrefix(iterator.Item().Key(), logPrefix)))
		}

		return nil
	}); err != nil {
		return fmt.Errorf("determine appended log index: %w", err)
	}

	if err := mgr.determineRepositoryExistence(); err != nil {
		return fmt.Errorf("determine repository existence: %w", err)
	}

	if mgr.repositoryExists {
		if err := mgr.createDirectories(); err != nil {
			return fmt.Errorf("create directories: %w", err)
		}
	}

	var err error
	mgr.customHookIndex, err = mgr.determineCustomHookIndex(ctx, mgr.appendedLogIndex, mgr.appliedLogIndex)
	if err != nil {
		return fmt.Errorf("determine hook index: %w", err)
	}

	// Each unapplied log entry should have a notification channel that gets closed when it is applied.
	// Create these channels here for the log entries.
	for i := mgr.appliedLogIndex + 1; i <= mgr.appendedLogIndex; i++ {
		mgr.applyNotifications[i] = make(chan struct{})
	}

	if err := mgr.removeStaleWALFiles(mgr.ctx, mgr.appendedLogIndex); err != nil {
		return fmt.Errorf("remove stale packs: %w", err)
	}

	mgr.initializationSuccessful = true

	return nil
}

// determineRepositoryExistence determines whether the repository exists or not by looking
// at whether the directory exists and whether there is a deletion request logged.
func (mgr *TransactionManager) determineRepositoryExistence() error {
	stat, err := os.Stat(mgr.repositoryPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("stat repository directory: %w", err)
		}
	}

	if stat != nil {
		if !stat.IsDir() {
			return errNotDirectory
		}

		mgr.repositoryExists = true
	}

	// Check whether the last log entry is a repository deletion. If so,
	// the repository has been deleted but the deletion wasn't yet applied.
	// The deletion is the last entry always as no further writes are
	// accepted if the repository doesn't exist.
	if mgr.appliedLogIndex < mgr.appendedLogIndex {
		logEntry, err := mgr.readLogEntry(mgr.appendedLogIndex)
		if err != nil {
			return fmt.Errorf("read log entry: %w", err)
		}

		if logEntry.RepositoryDeletion != nil {
			mgr.repositoryExists = false
		}
	}

	return nil
}

// determineCustomHookIndex determines the latest custom hooks in the repository.
//
//  1. First we iterate through the unapplied log in reverse order. The first log entry that
//     contains custom hooks must have the latest custom hooks since it is the latest log entry.
//  2. If we don't find any custom hooks in the log, the latest hooks could have been applied
//     to the repository already and the log entry pruned away. Look at the custom hooks on the
//     disk to see which are the latest.
//  3. If we found no custom hooks in the log nor in the repository, there are no custom hooks
//     configured.
func (mgr *TransactionManager) determineCustomHookIndex(ctx context.Context, appendedIndex, appliedIndex LogIndex) (LogIndex, error) {
	if !mgr.repositoryExists {
		// If the repository doesn't exist, then there are no hooks either.
		return 0, nil
	}

	for i := appendedIndex; appliedIndex < i; i-- {
		logEntry, err := mgr.readLogEntry(i)
		if err != nil {
			return 0, fmt.Errorf("read log entry: %w", err)
		}

		if logEntry.CustomHooksUpdate != nil {
			return i, nil
		}
	}

	hookDirs, err := os.ReadDir(filepath.Join(mgr.repositoryPath, "wal", "hooks"))
	if err != nil {
		return 0, fmt.Errorf("read hook directories: %w", err)
	}

	var hookIndex LogIndex
	for _, dir := range hookDirs {
		rawIndex, err := strconv.ParseUint(dir.Name(), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse hook index: %w", err)
		}

		if index := LogIndex(rawIndex); hookIndex < index {
			hookIndex = index
		}
	}

	return hookIndex, err
}

// createDirectories creates the directories that are expected to exist
// in the repository for storing the state. Initializing them simplifies
// rest of the code as it doesn't need handling for when they don't.
func (mgr *TransactionManager) createDirectories() error {
	for _, relativePath := range []string{
		"wal/hooks",
		"wal/packs",
	} {
		directory := filepath.Join(mgr.repositoryPath, relativePath)
		if _, err := os.Stat(directory); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("stat directory: %w", err)
			}

			if err := os.MkdirAll(directory, fs.ModePerm); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}

			if err := safe.NewSyncer().SyncHierarchy(mgr.repositoryPath, relativePath); err != nil {
				return fmt.Errorf("sync: %w", err)
			}
		}
	}

	return nil
}

// removeStaleWALFiles removes files from the log directory that have no associated log entry.
// Such files can be left around if transaction's files were moved in place successfully
// but the manager was interrupted before successfully persisting the log entry itself.
func (mgr *TransactionManager) removeStaleWALFiles(ctx context.Context, appendedIndex LogIndex) error {
	// Log entries are appended one by one to the log. If a write is interrupted, the only possible stale
	// pack would be for the next log index. Remove the pack if it exists.
	possibleStaleFilesPath := walFilesPathForLogIndex(mgr.repositoryPath, appendedIndex+1)
	if _, err := os.Stat(possibleStaleFilesPath); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove: %w", err)
		}

		return nil
	}

	if err := os.RemoveAll(possibleStaleFilesPath); err != nil {
		return fmt.Errorf("remove all: %w", err)
	}

	// Sync the parent directory to flush the file deletion.
	if err := safe.NewSyncer().SyncParent(possibleStaleFilesPath); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

// storeWALFiles moves the transaction's logged files from the staging directory to their destination in the log.
// It returns a function, even on errors, that must be called to clean up the files if committing the log entry
// fails.
func (mgr *TransactionManager) storeWALFiles(ctx context.Context, index LogIndex, transaction *Transaction) (func() error, error) {
	removeFiles := func() error { return nil }

	destinationPath := walFilesPathForLogIndex(mgr.repositoryPath, index)
	if err := os.Rename(
		transaction.walFilesPath(),
		destinationPath,
	); err != nil {
		return removeFiles, fmt.Errorf("move wal files: %w", err)
	}

	removeFiles = func() error {
		if err := os.Remove(destinationPath); err != nil {
			return fmt.Errorf("remove wal files: %w", err)
		}

		return nil
	}

	// Sync the parent directory. The pack's contents are synced when the pack file is computed.
	if err := safe.NewSyncer().Sync(filepath.Dir(destinationPath)); err != nil {
		return removeFiles, fmt.Errorf("sync: %w", err)
	}

	return removeFiles, nil
}

// walFilesPathForLogIndex returns an absolute path to a given log entry's WAL files.
func walFilesPathForLogIndex(repoPath string, index LogIndex) string {
	return filepath.Join(repoPath, "wal", "packs", index.String())
}

// packFilePath returns a log entry's pack file's absolute path in the wal files directory.
func packFilePath(walFiles string) string {
	return filepath.Join(walFiles, "transaction.pack")
}

// verifyReferences verifies that the references in the transaction apply on top of the already accepted
// reference changes. The old tips in the transaction are verified against the current actual tips.
// It returns the write-ahead log entry for the transaction if it was successfully verified.
func (mgr *TransactionManager) verifyReferences(ctx context.Context, transaction *Transaction) ([]*gitalypb.LogEntry_ReferenceUpdate, error) {
	if len(transaction.referenceUpdates) == 0 {
		return nil, nil
	}

	var referenceUpdates []*gitalypb.LogEntry_ReferenceUpdate
	for referenceName, update := range transaction.referenceUpdates {
		if err := git.ValidateReference(string(referenceName)); err != nil {
			return nil, InvalidReferenceFormatError{ReferenceName: referenceName}
		}

		if !update.Force {
			actualOldTip, err := transaction.stagingRepository.ResolveRevision(ctx, referenceName.Revision())
			if errors.Is(err, git.ErrReferenceNotFound) {
				objectHash, err := transaction.stagingRepository.ObjectHash(ctx)
				if err != nil {
					return nil, fmt.Errorf("object hash: %w", err)
				}

				actualOldTip = objectHash.ZeroOID
			} else if err != nil {
				return nil, fmt.Errorf("resolve revision: %w", err)
			}

			if update.OldOID != actualOldTip {
				if transaction.skipVerificationFailures {
					continue
				}

				return nil, ReferenceVerificationError{
					ReferenceName: referenceName,
					ExpectedOID:   update.OldOID,
					ActualOID:     actualOldTip,
				}
			}
		}

		referenceUpdates = append(referenceUpdates, &gitalypb.LogEntry_ReferenceUpdate{
			ReferenceName: []byte(referenceName),
			NewOid:        []byte(update.NewOID),
		})
	}

	// Sort the reference updates so the reference changes are always logged in a deterministic order.
	sort.Slice(referenceUpdates, func(i, j int) bool {
		return bytes.Compare(
			referenceUpdates[i].ReferenceName,
			referenceUpdates[j].ReferenceName,
		) == -1
	})

	if err := mgr.verifyReferencesWithGit(ctx, referenceUpdates, transaction.stagingRepository); err != nil {
		return nil, fmt.Errorf("verify references with git: %w", err)
	}

	return referenceUpdates, nil
}

// vefifyReferencesWithGit verifies the reference updates with git by preparing reference transaction. This ensures
// the updates will go through when they are being applied in the log. This also catches any invalid reference names
// and file/directory conflicts with Git's loose reference storage which can occur with references like
// 'refs/heads/parent' and 'refs/heads/parent/child'.
func (mgr *TransactionManager) verifyReferencesWithGit(ctx context.Context, referenceUpdates []*gitalypb.LogEntry_ReferenceUpdate, stagingRepository *localrepo.Repo) error {
	updater, err := mgr.prepareReferenceTransaction(ctx, referenceUpdates, stagingRepository)
	if err != nil {
		return fmt.Errorf("prepare reference transaction: %w", err)
	}

	return updater.Close()
}

// verifyDefaultBranchUpdate verifies the default branch referance update. This is done by first checking if it is one of
// the references in the current transaction which is not scheduled to be deleted. If not, we check if its a valid reference
// name in the repository. We don't do reference name validation because any reference going through the transaction manager
// has name validation and we can rely on that.
func (mgr *TransactionManager) verifyDefaultBranchUpdate(ctx context.Context, transaction *Transaction) error {
	referenceName := transaction.defaultBranchUpdate.Reference

	if err := git.ValidateReference(referenceName.String()); err != nil {
		return InvalidReferenceFormatError{ReferenceName: referenceName}
	}

	return nil
}

// applyDefaultBranchUpdate applies the default branch update to the repository from the log entry.
func (mgr *TransactionManager) applyDefaultBranchUpdate(ctx context.Context, defaultBranch *gitalypb.LogEntry_DefaultBranchUpdate) error {
	if defaultBranch == nil {
		return nil
	}

	var stderr bytes.Buffer
	if err := mgr.repository.ExecAndWait(ctx, git.Command{
		Name: "symbolic-ref",
		Args: []string{"HEAD", string(defaultBranch.ReferenceName)},
	}, git.WithStderr(&stderr), git.WithDisabledHooks()); err != nil {
		return structerr.New("exec symbolic-ref: %w", err).WithMetadata("stderr", stderr.String())
	}

	return nil
}

// prepareReferenceTransaction prepares a reference transaction with `git update-ref`. It leaves committing
// or aborting up to the caller. Either should be called to clean up the process. The process is cleaned up
// if an error is returned.
func (mgr *TransactionManager) prepareReferenceTransaction(ctx context.Context, referenceUpdates []*gitalypb.LogEntry_ReferenceUpdate, repository *localrepo.Repo) (*updateref.Updater, error) {
	// This section runs git-update-ref(1), but could fail due to existing
	// reference locks. So we create a function which can be called again
	// post cleanup of stale reference locks.
	updateFunc := func() (*updateref.Updater, error) {
		updater, err := updateref.New(ctx, repository, updateref.WithDisabledTransactions(), updateref.WithNoDeref())
		if err != nil {
			return nil, fmt.Errorf("new: %w", err)
		}

		if err := updater.Start(); err != nil {
			return nil, fmt.Errorf("start: %w", err)
		}

		for _, referenceUpdate := range referenceUpdates {
			if err := updater.Update(git.ReferenceName(referenceUpdate.ReferenceName), git.ObjectID(referenceUpdate.NewOid), ""); err != nil {
				return nil, fmt.Errorf("update %q: %w", referenceUpdate.ReferenceName, err)
			}
		}

		if err := updater.Prepare(); err != nil {
			return nil, fmt.Errorf("prepare: %w", err)
		}

		return updater, nil
	}

	// If git-update-ref(1) runs without issues, our work here is done.
	updater, err := updateFunc()
	if err == nil {
		return updater, nil
	}

	// If we get an error due to existing stale reference locks, we should clear it up
	// and retry running git-update-ref(1).
	var updateRefError updateref.AlreadyLockedError
	if errors.As(err, &updateRefError) {
		// Before clearing stale reference locks, we add should ensure that housekeeping doesn't
		// run git-pack-refs(1), which could create new reference locks. So we add an inhibitor.
		success, cleanup, err := mgr.housekeepingManager.AddPackRefsInhibitor(ctx, mgr.repositoryPath)
		if !success {
			return nil, fmt.Errorf("add pack-refs inhibitor: %w", err)
		}
		defer cleanup()

		// We ask housekeeping to cleanup stale reference locks. We don't add a grace period, because
		// transaction manager is the only process which writes into the repository, so it is safe
		// to delete these locks.
		if err := mgr.housekeepingManager.CleanStaleData(ctx, mgr.repository, housekeeping.OnlyStaleReferenceLockCleanup(0)); err != nil {
			return nil, fmt.Errorf("running reflock cleanup: %w", err)
		}

		// We try a second time, this should succeed. If not, there is something wrong and
		// we return the error.
		//
		// Do note, that we've already added an inhibitor above, so git-pack-refs(1) won't run
		// again until we return from this function so ideally this should work, but in case it
		// doesn't we return the error.
		return updateFunc()
	}

	return nil, err
}

// appendLogEntry appends the transaction to the write-ahead log. References that failed verification are skipped and thus not
// logged nor applied later.
func (mgr *TransactionManager) appendLogEntry(nextLogIndex LogIndex, logEntry *gitalypb.LogEntry) error {
	if err := mgr.storeLogEntry(nextLogIndex, logEntry); err != nil {
		return fmt.Errorf("set log entry: %w", err)
	}

	mgr.mutex.Lock()
	mgr.appendedLogIndex = nextLogIndex
	if logEntry.CustomHooksUpdate != nil {
		mgr.customHookIndex = nextLogIndex
	}
	mgr.applyNotifications[nextLogIndex] = make(chan struct{})
	if logEntry.RepositoryDeletion != nil {
		mgr.repositoryExists = false
		mgr.customHookIndex = 0
	}
	mgr.mutex.Unlock()

	return nil
}

// applyLogEntry reads a log entry at the given index and applies it to the repository.
func (mgr *TransactionManager) applyLogEntry(ctx context.Context, logIndex LogIndex) error {
	logEntry, err := mgr.readLogEntry(logIndex)
	if err != nil {
		return fmt.Errorf("read log entry: %w", err)
	}

	if logEntry.RepositoryDeletion != nil {
		// If the repository is being deleted, just delete it without any other changes given
		// they'd all be removed anyway. Reapplying the other changes after a crash would also
		// not work if the repository was successfully deleted before the crash.
		if err := mgr.applyRepositoryDeletion(ctx, logIndex); err != nil {
			return fmt.Errorf("apply repository deletion: %w", err)
		}
	} else {
		if logEntry.PackPrefix != "" {
			if err := mgr.applyPackFile(ctx, logEntry.PackPrefix, logIndex); err != nil {
				return fmt.Errorf("apply pack file: %w", err)
			}
		}

		if err := mgr.applyReferenceUpdates(ctx, logEntry.ReferenceUpdates); err != nil {
			return fmt.Errorf("apply reference updates: %w", err)
		}

		if err := mgr.applyDefaultBranchUpdate(ctx, logEntry.DefaultBranchUpdate); err != nil {
			return fmt.Errorf("writing default branch: %w", err)
		}

		if err := mgr.applyCustomHooks(ctx, logIndex, logEntry.CustomHooksUpdate); err != nil {
			return fmt.Errorf("apply custom hooks: %w", err)
		}
	}

	if err := mgr.storeAppliedLogIndex(logIndex); err != nil {
		return fmt.Errorf("set applied log index: %w", err)
	}

	if err := mgr.deleteLogEntry(logIndex); err != nil {
		return fmt.Errorf("deleting log entry: %w", err)
	}

	mgr.appliedLogIndex = logIndex

	// Notify the transactions waiting for this log entry to be applied prior to beginning.
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	notificationCh, ok := mgr.applyNotifications[logIndex]
	if !ok {
		// This should never happen and is a programming error if it does.
		return fmt.Errorf("no notification channel for LSN %d", logIndex)
	}
	delete(mgr.applyNotifications, logIndex)
	close(notificationCh)

	// There is no awaiter for a transaction if the transaction manager is recovering
	// transactions from the log after starting up.
	if resultChan, ok := mgr.awaitingTransactions[logIndex]; ok {
		resultChan <- nil
		delete(mgr.awaitingTransactions, logIndex)
	}

	return nil
}

// applyReferenceUpdates applies the applies the given reference updates to the repository.
func (mgr *TransactionManager) applyReferenceUpdates(ctx context.Context, updates []*gitalypb.LogEntry_ReferenceUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	updater, err := mgr.prepareReferenceTransaction(ctx, updates, mgr.repository)
	if err != nil {
		return fmt.Errorf("prepare reference transaction: %w", err)
	}

	if err := updater.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// applyRepositoryDeletion deletes the repository.
//
// Given how the repositories are laid out in the storage, we currently can't support MVCC for them.
// This is because there is only ever a single instance of a given repository.  We have to wait for all
// of the readers to finish before we can delete the repository as otherwise the readers could fail in
// unexpected ways and it would be an isolation violation. Repository deletions thus block before all
// transaction with an older read snapshot are done with the repository.
func (mgr *TransactionManager) applyRepositoryDeletion(ctx context.Context, index LogIndex) error {
	for {
		mgr.mutex.Lock()
		oldestElement := mgr.openTransactions.Front()
		mgr.mutex.Unlock()
		if oldestElement == nil {
			// If there are no open transactions, the deletion can proceed as there are
			// no readers.
			//
			// Any new transaction would have the deletion in their snapshot, and are waiting
			// for it to be applied prior to beginning.
			break
		}

		oldestTransaction := oldestElement.Value.(*Transaction)
		if oldestTransaction.snapshot.ReadIndex >= index {
			// If the oldest transaction is reading at this or later log index, it already has the deletion
			// in its snapshot, and is waiting for it to be applied. Proceed with the deletion as there
			// are no readers with the pre-deletion state in the snapshot.
			break
		}

		for {
			select {
			case <-oldestTransaction.finished:
				// The oldest transaction finished. Proceed to check the second oldest open transaction.
			case transaction := <-mgr.admissionQueue:
				// The oldest transaction could also be waiting to commit. Since the Run goroutine is
				// blocked here waiting for the transaction to finish, the write would never be admitted
				// for processing, leading to a deadlock. Since the repository was deleted, the only correct
				// outcome for the transaction would be to receive a not found error. Admit the transaction,
				// and finish it with the correct result so we can unblock the deletion.
				transaction.result <- ErrRepositoryNotFound
				if err := transaction.finish(); err != nil {
					return fmt.Errorf("finish transaction: %w", err)
				}

				continue
			case <-ctx.Done():
			}

			if err := ctx.Err(); err != nil {
				return err
			}

			break
		}
	}

	if err := os.RemoveAll(mgr.repositoryPath); err != nil {
		return fmt.Errorf("remove repository: %w", err)
	}

	if err := safe.NewSyncer().Sync(filepath.Dir(mgr.repositoryPath)); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

// applyPackFile unpacks the objects from the pack file into the repository if the log entry
// has an associated pack file. This is done by hard linking the pack and index from the
// log into the repository's object directory.
func (mgr *TransactionManager) applyPackFile(ctx context.Context, packPrefix string, logIndex LogIndex) error {
	packDirectory := filepath.Join(mgr.repositoryPath, "objects", "pack")
	for _, fileExtension := range []string{
		".pack",
		".idx",
		".rev",
	} {
		if err := os.Link(
			filepath.Join(walFilesPathForLogIndex(mgr.repositoryPath, logIndex), "objects"+fileExtension),
			filepath.Join(packDirectory, packPrefix+fileExtension),
		); err != nil {
			if !errors.Is(err, fs.ErrExist) {
				return fmt.Errorf("link file: %w", err)
			}

			// The file already existing means that we've already linked it in place or a repack
			// has resulted in the exact same file. No need to do anything about it.
		}
	}

	// Sync the new directory entries created.
	if err := safe.NewSyncer().Sync(packDirectory); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	return nil
}

// applyCustomHooks applies the custom hooks to the repository from the log entry. The custom hooks are stored
// at `<repo>/wal/hooks/<log_index>`. The custom hooks are fsynced prior to returning so it is safe to delete
// the log entry afterwards.
//
// The hooks are also extracted at `<repo>/custom_hooks`. This is done for backwards compatibility, as we want
// the hooks to be present even if the WAL logic is disabled. This ensures we don't lose data if we have to
// disable the WAL logic after rollout.
func (mgr *TransactionManager) applyCustomHooks(ctx context.Context, logIndex LogIndex, update *gitalypb.LogEntry_CustomHooksUpdate) error {
	if update == nil {
		return nil
	}

	targetDirectory := customHookPathForLogIndex(mgr.repositoryPath, logIndex)
	if err := os.Mkdir(targetDirectory, fs.ModePerm); err != nil {
		// The target directory may exist if we previously tried to extract the
		// custom hooks there. TAR overwrites existing files and the custom hooks
		// files are guaranteed to be the same as this is the same log entry.
		if !errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	syncer := safe.NewSyncer()
	extractHooks := func(destinationDir string) error {
		if err := repoutil.ExtractHooks(ctx, bytes.NewReader(update.CustomHooksTar), destinationDir, true); err != nil {
			return fmt.Errorf("extract hooks: %w", err)
		}

		// TAR doesn't sync the extracted files so do it manually here.
		if err := syncer.SyncRecursive(destinationDir); err != nil {
			return fmt.Errorf("sync hooks: %w", err)
		}

		return nil
	}

	if err := extractHooks(targetDirectory); err != nil {
		return fmt.Errorf("extract hooks: %w", err)
	}

	// Sync the parent directory as well.
	if err := syncer.SyncParent(targetDirectory); err != nil {
		return fmt.Errorf("sync hook directory: %w", err)
	}

	// Extract another copy that we can move to `<repo>/custom_hooks` where the hooks exist without the WAL enabled.
	// We make a second copy as if we disable the WAL, we have to clear all of its state prior to re-enabling it.
	// This would clear the hooks so symbolic linking the first copy is not enough.
	tmpDir, err := os.MkdirTemp(mgr.stagingDirectory, "")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}

	if err := extractHooks(tmpDir); err != nil {
		return fmt.Errorf("extract legacy hooks: %w", err)
	}

	legacyHooksPath := filepath.Join(mgr.repositoryPath, repoutil.CustomHooksDir)
	// The hooks are lost if we perform this removal but fail to perform the remaining operations and the
	// WAL is disabled before succeeding. This is an existing issue already with SetCustomHooks RPC.
	if err := os.RemoveAll(legacyHooksPath); err != nil {
		return fmt.Errorf("remove existing legacy hooks: %w", err)
	}

	if err := os.Rename(tmpDir, legacyHooksPath); err != nil {
		return fmt.Errorf("move legacy hooks in place: %w", err)
	}

	if err := syncer.SyncParent(legacyHooksPath); err != nil {
		return fmt.Errorf("sync legacy hooks directory entry: %w", err)
	}

	return nil
}

// customHookPathForLogIndex returns the filesystem paths where the custom hooks
// for the given log index are stored.
func customHookPathForLogIndex(repositoryPath string, logIndex LogIndex) string {
	return filepath.Join(repositoryPath, "wal", "hooks", logIndex.String())
}

// deleteLogEntry deletes the log entry at the given index from the log.
func (mgr *TransactionManager) deleteLogEntry(index LogIndex) error {
	return mgr.deleteKey(keyLogEntry(mgr.relativePath, index))
}

// readLogEntry returns the log entry from the given position in the log.
func (mgr *TransactionManager) readLogEntry(index LogIndex) (*gitalypb.LogEntry, error) {
	var logEntry gitalypb.LogEntry
	key := keyLogEntry(mgr.relativePath, index)

	if err := mgr.readKey(key, &logEntry); err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}

	return &logEntry, nil
}

// storeLogEntry stores the log entry in the repository's write-ahead log at the given index.
func (mgr *TransactionManager) storeLogEntry(index LogIndex, entry *gitalypb.LogEntry) error {
	return mgr.setKey(keyLogEntry(mgr.relativePath, index), entry)
}

// storeAppliedLogIndex stores the repository's applied log index in the database.
func (mgr *TransactionManager) storeAppliedLogIndex(index LogIndex) error {
	return mgr.setKey(keyAppliedLogIndex(mgr.relativePath), index.toProto())
}

// setKey marshals and stores a given protocol buffer message into the database under the given key.
func (mgr *TransactionManager) setKey(key []byte, value proto.Message) error {
	marshaledValue, err := proto.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	writeBatch := mgr.db.NewWriteBatch()
	defer writeBatch.Cancel()

	if err := writeBatch.Set(key, marshaledValue); err != nil {
		return fmt.Errorf("set: %w", err)
	}

	return writeBatch.Flush()
}

// readKey reads a key from the database and unmarshals its value in to the destination protocol
// buffer message.
func (mgr *TransactionManager) readKey(key []byte, destination proto.Message) error {
	return mgr.db.View(func(txn databaseTransaction) error {
		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		return item.Value(func(value []byte) error { return proto.Unmarshal(value, destination) })
	})
}

// deleteKey deletes a key from the database.
func (mgr *TransactionManager) deleteKey(key []byte) error {
	return mgr.db.Update(func(txn databaseTransaction) error {
		if err := txn.Delete(key); err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		return nil
	})
}

// keyAppliedLogIndex returns the database key storing a repository's last applied log entry's index.
func keyAppliedLogIndex(repositoryID string) []byte {
	return []byte(fmt.Sprintf("repository/%s/log/index/applied", repositoryID))
}

// keyLogEntry returns the database key storing a repository's log entry at a given index.
func keyLogEntry(repositoryID string, index LogIndex) []byte {
	marshaledIndex := make([]byte, binary.Size(index))
	binary.BigEndian.PutUint64(marshaledIndex, uint64(index))
	return []byte(fmt.Sprintf("%s%s", keyPrefixLogEntries(repositoryID), marshaledIndex))
}

// keyPrefixLogEntries returns the key prefix holding repository's write-ahead log entries.
func keyPrefixLogEntries(repositoryID string) []byte {
	return []byte(fmt.Sprintf("repository/%s/log/entry/", repositoryID))
}
