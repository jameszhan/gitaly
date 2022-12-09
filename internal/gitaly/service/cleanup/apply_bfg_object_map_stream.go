package cleanup

import (
	"io"

	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/service/cleanup/internalrefs"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/service/cleanup/notifier"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper/chunk"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/v15/streamio"
	"google.golang.org/protobuf/proto"
)

type bfgStreamReader struct {
	firstRequest *gitalypb.ApplyBfgObjectMapStreamRequest

	server gitalypb.CleanupService_ApplyBfgObjectMapStreamServer
}

type bfgStreamWriter struct {
	entries []*gitalypb.ApplyBfgObjectMapStreamResponse_Entry

	server gitalypb.CleanupService_ApplyBfgObjectMapStreamServer
}

func (s *server) ApplyBfgObjectMapStream(server gitalypb.CleanupService_ApplyBfgObjectMapStreamServer) error {
	firstRequest, err := server.Recv()
	if err != nil {
		return helper.ErrInternal(err)
	}

	if err := validateFirstRequest(firstRequest); err != nil {
		return helper.ErrInvalidArgumentf("%w", err)
	}

	ctx := server.Context()
	repo := s.localrepo(firstRequest.GetRepository())
	reader := &bfgStreamReader{firstRequest: firstRequest, server: server}
	chunker := chunk.New(&bfgStreamWriter{server: server})

	notifier, cancel, err := notifier.New(ctx, s.catfileCache, repo, chunker)
	if err != nil {
		return helper.ErrInternal(err)
	}
	defer cancel()

	// It doesn't matter if new internal references are added after this RPC
	// starts running - they shouldn't point to the objects removed by the BFG
	cleaner, err := internalrefs.NewCleaner(ctx, repo, notifier.Notify)
	if err != nil {
		return helper.ErrInternal(err)
	}

	if err := cleaner.ApplyObjectMap(ctx, reader.streamReader()); err != nil {
		if invalidErr, ok := err.(internalrefs.ErrInvalidObjectMap); ok {
			return helper.ErrInvalidArgumentf("%w", invalidErr)
		}

		return helper.ErrInternal(err)
	}

	return helper.ErrInternal(chunker.Flush())
}

func validateFirstRequest(req *gitalypb.ApplyBfgObjectMapStreamRequest) error {
	return service.ValidateRepository(req.GetRepository())
}

func (r *bfgStreamReader) readOne() ([]byte, error) {
	if r.firstRequest != nil {
		data := r.firstRequest.GetObjectMap()
		r.firstRequest = nil
		return data, nil
	}

	req, err := r.server.Recv()
	if err != nil {
		return nil, err
	}

	return req.GetObjectMap(), nil
}

func (r *bfgStreamReader) streamReader() io.Reader {
	return streamio.NewReader(r.readOne)
}

func (w *bfgStreamWriter) Append(m proto.Message) {
	w.entries = append(
		w.entries,
		m.(*gitalypb.ApplyBfgObjectMapStreamResponse_Entry),
	)
}

func (w *bfgStreamWriter) Reset() {
	w.entries = nil
}

func (w *bfgStreamWriter) Send() error {
	msg := &gitalypb.ApplyBfgObjectMapStreamResponse{Entries: w.entries}

	return w.server.Send(msg)
}
