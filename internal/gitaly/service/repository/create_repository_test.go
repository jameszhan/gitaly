package repository

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v14/internal/git"
	"gitlab.com/gitlab-org/gitaly/v14/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/config/auth"
	"gitlab.com/gitlab-org/gitaly/v14/internal/gitaly/transaction"
	"gitlab.com/gitlab-org/gitaly/v14/internal/metadata"
	"gitlab.com/gitlab-org/gitaly/v14/internal/metadata/featureflag"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testassert"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v14/internal/testhelper/testserver"
	"gitlab.com/gitlab-org/gitaly/v14/internal/transaction/txinfo"
	"gitlab.com/gitlab-org/gitaly/v14/internal/transaction/voting"
	"gitlab.com/gitlab-org/gitaly/v14/proto/go/gitalypb"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateRepository_missingAuth(t *testing.T) {
	t.Parallel()

	testhelper.NewFeatureSets(featureflag.TxAtomicRepositoryCreation).Run(t, testCreateRepositoryMissingAuth)
}

func testCreateRepositoryMissingAuth(t *testing.T, ctx context.Context) {
	cfg, repo, _ := testcfg.BuildWithRepo(t, testcfg.WithBase(config.Cfg{Auth: auth.Config{Token: "some"}}))

	serverSocketPath := runRepositoryServerWithConfig(t, cfg, nil)
	client := newRepositoryClient(t, config.Cfg{Auth: auth.Config{Token: ""}}, serverSocketPath)

	_, err := client.CreateRepository(ctx, &gitalypb.CreateRepositoryRequest{Repository: repo})

	testhelper.RequireGrpcError(t, err, codes.Unauthenticated)
}

func TestCreateRepository_successful(t *testing.T) {
	t.Parallel()

	testhelper.NewFeatureSets(featureflag.TxAtomicRepositoryCreation).Run(t, testCreateRepositorySuccessful)
}

func testCreateRepositorySuccessful(t *testing.T, ctx context.Context) {
	cfg, client := setupRepositoryServiceWithoutRepo(t)

	relativePath := "create-repository-test.git"
	repoDir := filepath.Join(cfg.Storages[0].Path, relativePath)

	repo := &gitalypb.Repository{StorageName: cfg.Storages[0].Name, RelativePath: relativePath}
	req := &gitalypb.CreateRepositoryRequest{Repository: repo}
	_, err := client.CreateRepository(ctx, req)
	require.NoError(t, err)

	require.NoError(t, unix.Access(repoDir, unix.R_OK))
	require.NoError(t, unix.Access(repoDir, unix.W_OK))
	require.NoError(t, unix.Access(repoDir, unix.X_OK))

	for _, dir := range []string{repoDir, filepath.Join(repoDir, "refs")} {
		fi, err := os.Stat(dir)
		require.NoError(t, err)
		require.True(t, fi.IsDir(), "%q must be a directory", fi.Name())

		require.NoError(t, unix.Access(dir, unix.R_OK))
		require.NoError(t, unix.Access(dir, unix.W_OK))
		require.NoError(t, unix.Access(dir, unix.X_OK))
	}

	symRef := testhelper.MustReadFile(t, path.Join(repoDir, "HEAD"))
	require.Equal(t, symRef, []byte(fmt.Sprintf("ref: %s\n", git.DefaultRef)))
}

func TestCreateRepository_failure(t *testing.T) {
	t.Parallel()

	testhelper.NewFeatureSets(featureflag.TxAtomicRepositoryCreation).Run(t, testCreateRepositoryFailure)
}

func testCreateRepositoryFailure(t *testing.T, ctx context.Context) {
	cfg, client := setupRepositoryServiceWithoutRepo(t)

	storagePath := cfg.Storages[0].Path
	fullPath := filepath.Join(storagePath, "foo.git")

	_, err := os.Create(fullPath)
	require.NoError(t, err)

	_, err = client.CreateRepository(ctx, &gitalypb.CreateRepositoryRequest{
		Repository: &gitalypb.Repository{StorageName: cfg.Storages[0].Name, RelativePath: "foo.git"},
	})

	if featureflag.TxAtomicRepositoryCreation.IsEnabled(ctx) {
		testhelper.RequireGrpcError(t, err, codes.AlreadyExists)
	} else {
		testhelper.RequireGrpcError(t, err, codes.Internal)
	}
}

func TestCreateRepository_invalidArguments(t *testing.T) {
	t.Parallel()

	testhelper.NewFeatureSets(featureflag.TxAtomicRepositoryCreation).Run(t, testCreateRepositoryInvalidArguments)
}

func testCreateRepositoryInvalidArguments(t *testing.T, ctx context.Context) {
	_, client := setupRepositoryServiceWithoutRepo(t)

	testCases := []struct {
		repo *gitalypb.Repository
		code codes.Code
	}{
		{
			repo: &gitalypb.Repository{StorageName: "does not exist", RelativePath: "foobar.git"},
			code: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%+v", tc.repo), func(t *testing.T) {
			_, err := client.CreateRepository(ctx, &gitalypb.CreateRepositoryRequest{Repository: tc.repo})

			require.Error(t, err)
			testhelper.RequireGrpcError(t, err, tc.code)
		})
	}
}

func TestCreateRepository_transactional(t *testing.T) {
	t.Parallel()

	testhelper.NewFeatureSets(featureflag.TxAtomicRepositoryCreation).Run(t, testCreateRepositoryTransactional)
}

func testCreateRepositoryTransactional(t *testing.T, ctx context.Context) {
	var actualVote voting.Vote
	var called int

	mockTxManager := transaction.MockManager{
		VoteFn: func(ctx context.Context, tx txinfo.Transaction, v voting.Vote) error {
			actualVote = v
			called++
			return nil
		},
	}

	cfg, client := setupRepositoryServiceWithoutRepo(t, testserver.WithTransactionManager(&mockTxManager))

	ctx, err := txinfo.InjectTransaction(ctx, 1, "node", true)
	require.NoError(t, err)
	ctx = metadata.IncomingToOutgoing(ctx)

	t.Run("initial creation without refs", func(t *testing.T) {
		called = 0
		actualVote = voting.Vote{}

		_, err = client.CreateRepository(ctx, &gitalypb.CreateRepositoryRequest{
			Repository: &gitalypb.Repository{
				StorageName:  cfg.Storages[0].Name,
				RelativePath: "repo.git",
			},
		})
		require.NoError(t, err)

		require.DirExists(t, filepath.Join(cfg.Storages[0].Path, "repo.git"))
		if featureflag.TxAtomicRepositoryCreation.IsEnabled(ctx) {
			require.Equal(t, 2, called, "expected transactional vote")
		} else {
			require.Equal(t, 1, called, "expected transactional vote")
			require.Equal(t, voting.VoteFromData([]byte{}), actualVote)
		}
	})

	t.Run("idempotent creation with preexisting refs", func(t *testing.T) {
		called = 0
		actualVote = voting.Vote{}

		repo, repoPath := gittest.CloneRepo(t, cfg, cfg.Storages[0])

		_, err = client.CreateRepository(ctx, &gitalypb.CreateRepositoryRequest{
			Repository: repo,
		})

		if featureflag.TxAtomicRepositoryCreation.IsEnabled(ctx) {
			testassert.ProtoEqual(t, status.Error(codes.AlreadyExists, "creating repository: repository exists already"), err)
			return
		}

		require.NoError(t, err)

		refs := gittest.Exec(t, cfg, "-C", repoPath, "for-each-ref")
		require.NotEmpty(t, refs)

		require.Equal(t, 1, called, "expected transactional vote")
		require.Equal(t, voting.VoteFromData(refs), actualVote)
	})
}

func TestCreateRepository_idempotent(t *testing.T) {
	t.Parallel()

	testhelper.NewFeatureSets(featureflag.TxAtomicRepositoryCreation).Run(t, testCreateRepositoryIdempotent)
}

func testCreateRepositoryIdempotent(t *testing.T, ctx context.Context) {
	cfg, repo, repoPath, client := setupRepositoryService(t)

	refsBefore := strings.Split(string(gittest.Exec(t, cfg, "-C", repoPath, "for-each-ref")), "\n")

	req := &gitalypb.CreateRepositoryRequest{Repository: repo}
	_, err := client.CreateRepository(ctx, req)

	if featureflag.TxAtomicRepositoryCreation.IsEnabled(ctx) {
		testassert.ProtoEqual(t, status.Error(codes.AlreadyExists, "creating repository: repository exists already"), err)
		return
	}

	require.NoError(t, err)

	refsAfter := strings.Split(string(gittest.Exec(t, cfg, "-C", repoPath, "for-each-ref")), "\n")

	assert.Equal(t, refsBefore, refsAfter)
}