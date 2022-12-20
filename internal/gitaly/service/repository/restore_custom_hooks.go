package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"gitlab.com/gitlab-org/gitaly/v15/internal/command"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/transaction"
	"gitlab.com/gitlab-org/gitaly/v15/internal/metadata/featureflag"
	"gitlab.com/gitlab-org/gitaly/v15/internal/safe"
	"gitlab.com/gitlab-org/gitaly/v15/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v15/internal/transaction/txinfo"
	"gitlab.com/gitlab-org/gitaly/v15/internal/transaction/voting"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/v15/streamio"
)

func (s *server) RestoreCustomHooks(stream gitalypb.RepositoryService_RestoreCustomHooksServer) error {
	if featureflag.TransactionalRestoreCustomHooks.IsEnabled(stream.Context()) {
		return s.restoreCustomHooksWithVoting(stream)
	}

	firstRequest, err := stream.Recv()
	if err != nil {
		return structerr.NewInternal("first request failed %w", err)
	}

	repository := firstRequest.GetRepository()
	if err := service.ValidateRepository(repository); err != nil {
		return structerr.NewInvalidArgument("%w", err)
	}

	reader := streamio.NewReader(func() ([]byte, error) {
		if firstRequest != nil {
			data := firstRequest.GetData()
			firstRequest = nil
			return data, nil
		}

		request, err := stream.Recv()
		return request.GetData(), err
	})

	repoPath, err := s.locator.GetPath(repository)
	if err != nil {
		return structerr.NewInternal("getting repo path failed %w", err)
	}

	cmdArgs := []string{
		"-xf",
		"-",
		"-C",
		repoPath,
		customHooksDir,
	}

	ctx := stream.Context()
	cmd, err := command.New(ctx, append([]string{"tar"}, cmdArgs...), command.WithStdin(reader))
	if err != nil {
		return structerr.NewInternal("Could not untar custom hooks tar %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return structerr.NewInternal("cmd wait failed: %w", err)
	}

	return stream.SendAndClose(&gitalypb.RestoreCustomHooksResponse{})
}

func (s *server) restoreCustomHooksWithVoting(stream gitalypb.RepositoryService_RestoreCustomHooksServer) error {
	firstRequest, err := stream.Recv()
	if err != nil {
		return structerr.NewInternal("first request failed %w", err)
	}

	ctx := stream.Context()

	repository := firstRequest.GetRepository()
	if err := service.ValidateRepository(repository); err != nil {
		return structerr.NewInvalidArgument("%w", err)
	}

	v := voting.NewVoteHash()

	repoPath, err := s.locator.GetRepoPath(repository)
	if err != nil {
		return structerr.NewInternal("RestoreCustomHooks: getting repo path failed %w", err)
	}

	customHooksPath := filepath.Join(repoPath, customHooksDir)

	if err = os.MkdirAll(customHooksPath, os.ModePerm); err != nil {
		return structerr.NewInternal("making custom hooks directory %w", err)
	}

	lockDir, err := safe.NewLockingDirectory(customHooksPath)
	if err != nil {
		return structerr.NewInternal("RestoreCustomHooks: creating locking directory: %w", err)
	}

	if err := lockDir.Lock(); err != nil {
		return structerr.NewInternal("locking directory failed: %w", err)
	}

	defer func() {
		if !lockDir.IsLocked() {
			return
		}

		if err := lockDir.Unlock(); err != nil {
			ctxlogrus.Extract(ctx).WithError(err).Warn("could not unlock directory")
		}
	}()

	if err := voteCustomHooks(ctx, s.txManager, &v, voting.Prepared); err != nil {
		return err
	}

	reader := streamio.NewReader(func() ([]byte, error) {
		var data []byte
		defer func() {
			_, _ = v.Write(data)
		}()

		if firstRequest != nil {
			data = firstRequest.GetData()
			firstRequest = nil
			return data, nil
		}

		request, err := stream.Recv()

		data = request.GetData()
		return data, err
	})

	cmdArgs := []string{
		"-xf",
		"-",
		"-C",
		repoPath,
		customHooksDir,
	}

	cmd, err := command.New(ctx, append([]string{"tar"}, cmdArgs...), command.WithStdin(reader))
	if err != nil {
		return structerr.NewInternal("Could not untar custom hooks tar %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return structerr.NewInternal("cmd wait failed: %w", err)
	}

	if err := voteCustomHooks(ctx, s.txManager, &v, voting.Committed); err != nil {
		return err
	}

	if err := lockDir.Unlock(); err != nil {
		return structerr.NewInternal("committing lock dir %w", err)
	}

	return stream.SendAndClose(&gitalypb.RestoreCustomHooksResponse{})
}

func voteCustomHooks(
	ctx context.Context,
	txManager transaction.Manager,
	v *voting.VoteHash,
	phase voting.Phase,
) error {
	tx, err := txinfo.TransactionFromContext(ctx)
	if errors.Is(err, txinfo.ErrTransactionNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	vote, err := v.Vote()
	if err != nil {
		return err
	}

	if err := txManager.Vote(ctx, tx, vote, phase); err != nil {
		return fmt.Errorf("vote failed: %w", err)
	}

	return nil
}
