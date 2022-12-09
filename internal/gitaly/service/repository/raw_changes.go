package repository

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"gitlab.com/gitlab-org/gitaly/v15/internal/git"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/catfile"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/rawdiff"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper/chunk"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
	"google.golang.org/protobuf/proto"
)

func (s *server) GetRawChanges(req *gitalypb.GetRawChangesRequest, stream gitalypb.RepositoryService_GetRawChangesServer) error {
	ctx := stream.Context()
	repository := req.GetRepository()
	if err := service.ValidateRepository(repository); err != nil {
		return helper.ErrInvalidArgumentf("%w", err)
	}
	repo := s.localrepo(repository)

	objectInfoReader, cancel, err := s.catfileCache.ObjectInfoReader(stream.Context(), repo)
	if err != nil {
		return helper.ErrInternal(err)
	}
	defer cancel()

	if err := validateRawChangesRequest(ctx, req, objectInfoReader); err != nil {
		return helper.ErrInvalidArgumentf("%w", err)
	}

	if err := s.getRawChanges(stream, repo, objectInfoReader, req.GetFromRevision(), req.GetToRevision()); err != nil {
		return helper.ErrInternal(err)
	}

	return nil
}

func validateRawChangesRequest(ctx context.Context, req *gitalypb.GetRawChangesRequest, objectInfoReader catfile.ObjectInfoReader) error {
	if from := req.FromRevision; !git.ObjectHashSHA1.IsZeroOID(git.ObjectID(from)) {
		if _, err := objectInfoReader.Info(ctx, git.Revision(from)); err != nil {
			return fmt.Errorf("invalid 'from' revision: %q", from)
		}
	}

	if to := req.ToRevision; !git.ObjectHashSHA1.IsZeroOID(git.ObjectID(to)) {
		if _, err := objectInfoReader.Info(ctx, git.Revision(to)); err != nil {
			return fmt.Errorf("invalid 'to' revision: %q", to)
		}
	}

	return nil
}

func (s *server) getRawChanges(stream gitalypb.RepositoryService_GetRawChangesServer, repo git.RepositoryExecutor, objectInfoReader catfile.ObjectInfoReader, from, to string) error {
	if git.ObjectHashSHA1.IsZeroOID(git.ObjectID(to)) {
		return nil
	}

	if git.ObjectHashSHA1.IsZeroOID(git.ObjectID(from)) {
		from = git.ObjectHashSHA1.EmptyTreeOID.String()
	}

	ctx := stream.Context()

	diffCmd, err := repo.Exec(ctx, git.SubCmd{
		Name:  "diff",
		Flags: []git.Option{git.Flag{Name: "--raw"}, git.Flag{Name: "-z"}},
		Args:  []string{from, to},
	})
	if err != nil {
		return fmt.Errorf("start git diff: %w", err)
	}

	p := rawdiff.NewParser(diffCmd)
	chunker := chunk.New(&rawChangesSender{stream: stream})

	for {
		d, err := p.NextDiff()
		if err == io.EOF {
			break // happy path
		}
		if err != nil {
			return fmt.Errorf("read diff: %w", err)
		}

		change, err := changeFromDiff(ctx, objectInfoReader, d)
		if err != nil {
			return fmt.Errorf("build change from diff line: %w", err)
		}

		if err := chunker.Send(change); err != nil {
			return fmt.Errorf("send response: %w", err)
		}
	}

	if err := diffCmd.Wait(); err != nil {
		return fmt.Errorf("wait git diff: %w", err)
	}

	return chunker.Flush()
}

type rawChangesSender struct {
	stream  gitalypb.RepositoryService_GetRawChangesServer
	changes []*gitalypb.GetRawChangesResponse_RawChange
}

func (s *rawChangesSender) Reset() { s.changes = nil }
func (s *rawChangesSender) Append(m proto.Message) {
	s.changes = append(s.changes, m.(*gitalypb.GetRawChangesResponse_RawChange))
}

func (s *rawChangesSender) Send() error {
	response := &gitalypb.GetRawChangesResponse{RawChanges: s.changes}
	return s.stream.Send(response)
}

// Ordinarily, Git uses 0000000000000000000000000000000000000000, the
// "null SHA", to represent a non-existing object. In the output of `git
// diff --raw` however there are only abbreviated SHA's, i.e. with less
// than 40 characters. Within this context the null SHA is a string that
// consists of 1 to 40 zeroes.
var zeroRegexp = regexp.MustCompile(`\A0+\z`)

const submoduleTreeEntryMode = "160000"

func changeFromDiff(ctx context.Context, objectInfoReader catfile.ObjectInfoReader, d *rawdiff.Diff) (*gitalypb.GetRawChangesResponse_RawChange, error) {
	resp := &gitalypb.GetRawChangesResponse_RawChange{}

	newMode64, err := strconv.ParseInt(d.DstMode, 8, 32)
	if err != nil {
		return nil, err
	}
	resp.NewMode = int32(newMode64)

	oldMode64, err := strconv.ParseInt(d.SrcMode, 8, 32)
	if err != nil {
		return nil, err
	}
	resp.OldMode = int32(oldMode64)

	if err := setOperationAndPaths(d, resp); err != nil {
		return nil, err
	}

	shortBlobID := d.DstSHA
	blobMode := d.DstMode
	if zeroRegexp.MatchString(shortBlobID) {
		shortBlobID = d.SrcSHA
		blobMode = d.SrcMode
	}

	if blobMode != submoduleTreeEntryMode {
		info, err := objectInfoReader.Info(ctx, git.Revision(shortBlobID))
		if err != nil {
			return nil, fmt.Errorf("find %q: %w", shortBlobID, err)
		}

		resp.BlobId = info.Oid.String()
		resp.Size = info.Size
	}

	return resp, nil
}

func setOperationAndPaths(d *rawdiff.Diff, resp *gitalypb.GetRawChangesResponse_RawChange) error {
	if len(d.Status) == 0 {
		return fmt.Errorf("empty diff status")
	}

	resp.NewPathBytes = []byte(d.SrcPath)
	resp.OldPathBytes = []byte(d.SrcPath)

	switch d.Status[0] {
	case 'A':
		resp.Operation = gitalypb.GetRawChangesResponse_RawChange_ADDED
		resp.OldPathBytes = nil
	case 'C':
		resp.Operation = gitalypb.GetRawChangesResponse_RawChange_COPIED
		resp.NewPathBytes = []byte(d.DstPath)
	case 'D':
		resp.Operation = gitalypb.GetRawChangesResponse_RawChange_DELETED
		resp.NewPathBytes = nil
	case 'M':
		resp.Operation = gitalypb.GetRawChangesResponse_RawChange_MODIFIED
	case 'R':
		resp.Operation = gitalypb.GetRawChangesResponse_RawChange_RENAMED
		resp.NewPathBytes = []byte(d.DstPath)
	case 'T':
		resp.Operation = gitalypb.GetRawChangesResponse_RawChange_TYPE_CHANGED
	default:
		resp.Operation = gitalypb.GetRawChangesResponse_RawChange_UNKNOWN
	}

	return nil
}
