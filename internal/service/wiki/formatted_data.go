package wiki

import (
	"gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/internal/rubyserver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) WikiGetFormattedData(request *gitalypb.WikiGetFormattedDataRequest, stream gitalypb.WikiService_WikiGetFormattedDataServer) error {
	ctx := stream.Context()

	if len(request.GetTitle()) == 0 {
		return status.Errorf(codes.InvalidArgument, "WikiGetFormattedData: Empty Title")
	}

	client, err := s.WikiServiceClient(ctx)
	if err != nil {
		return err
	}

	clientCtx, err := rubyserver.SetHeaders(ctx, request.GetRepository())
	if err != nil {
		return err
	}

	rubyStream, err := client.WikiGetFormattedData(clientCtx, request)
	if err != nil {
		return err
	}

	return rubyserver.Proxy(func() error {
		resp, err := rubyStream.Recv()
		if err != nil {
			md := rubyStream.Trailer()
			stream.SetTrailer(md)
			return err
		}
		return stream.Send(resp)
	})
}
