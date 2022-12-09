package ref

import (
	"context"
	"strings"

	"gitlab.com/gitlab-org/gitaly/v15/internal/command"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
)

// RefExists returns true if the given reference exists. The ref must start with the string `ref/`
func (s *server) RefExists(ctx context.Context, in *gitalypb.RefExistsRequest) (*gitalypb.RefExistsResponse, error) {
	if err := service.ValidateRepository(in.GetRepository()); err != nil {
		return nil, helper.ErrInvalidArgumentf("%w", err)
	}

	ref := string(in.Ref)

	if !isValidRefName(ref) {
		return nil, helper.ErrInvalidArgumentf("invalid refname")
	}

	exists, err := s.refExists(ctx, in.Repository, ref)
	if err != nil {
		return nil, helper.ErrInternal(err)
	}

	return &gitalypb.RefExistsResponse{Value: exists}, nil
}

func (s *server) refExists(ctx context.Context, repo *gitalypb.Repository, ref string) (bool, error) {
	cmd, err := s.gitCmdFactory.New(ctx, repo, git.SubCmd{
		Name:  "show-ref",
		Flags: []git.Option{git.Flag{Name: "--verify"}, git.Flag{Name: "--quiet"}},
		Args:  []string{ref},
	})
	if err != nil {
		return false, err
	}

	err = cmd.Wait()
	if err == nil {
		// Exit code 0: the ref exists
		return true, nil
	}

	if code, ok := command.ExitStatus(err); ok && code == 1 {
		// Exit code 1: the ref does not exist
		return false, nil
	}

	// This will normally occur when exit code > 1
	return false, err
}

func isValidRefName(refName string) bool {
	return strings.HasPrefix(refName, "refs/")
}
