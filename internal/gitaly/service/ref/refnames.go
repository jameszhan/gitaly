package ref

import (
	"bufio"
	"context"

	"gitlab.com/gitlab-org/gitaly/v15/internal/git"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper/chunk"
	"gitlab.com/gitlab-org/gitaly/v15/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// FindAllTagNames creates a stream of ref names for all tags in the given repository
func (s *server) FindAllTagNames(in *gitalypb.FindAllTagNamesRequest, stream gitalypb.RefService_FindAllTagNamesServer) error {
	ctx := stream.Context()

	if err := service.ValidateRepository(in.GetRepository()); err != nil {
		return structerr.NewInvalidArgument("%w", err)
	}

	repo := s.localrepo(in.GetRepository())
	chunker := chunk.New(&findAllTagNamesSender{stream: stream})

	if err := listRefNames(ctx, repo, chunker, "refs/tags", nil); err != nil {
		return structerr.NewInternal("%w", err)
	}

	return nil
}

type findAllTagNamesSender struct {
	stream   gitalypb.RefService_FindAllTagNamesServer
	tagNames [][]byte
}

func (ts *findAllTagNamesSender) Reset() { ts.tagNames = nil }
func (ts *findAllTagNamesSender) Append(m proto.Message) {
	ts.tagNames = append(ts.tagNames, []byte(m.(*wrapperspb.StringValue).Value))
}

func (ts *findAllTagNamesSender) Send() error {
	return ts.stream.Send(&gitalypb.FindAllTagNamesResponse{Names: ts.tagNames})
}

func listRefNames(ctx context.Context, repo git.RepositoryExecutor, chunker *chunk.Chunker, prefix string, extraArgs []string) error {
	flags := []git.Option{
		git.Flag{Name: "--format=%(refname)"},
	}

	for _, arg := range extraArgs {
		flags = append(flags, git.Flag{Name: arg})
	}

	cmd, err := repo.Exec(ctx, git.Command{
		Name:  "for-each-ref",
		Flags: flags,
		Args:  []string{prefix},
	})
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmd)
	for scanner.Scan() {
		// Important: don't use scanner.Bytes() because the slice will become
		// invalid on the next loop iteration. Instead, use scanner.Text() to
		// force a copy.
		if err := chunker.Send(&wrapperspb.StringValue{Value: scanner.Text()}); err != nil {
			return err
		}
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return chunker.Flush()
}
