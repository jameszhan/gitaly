package namespace

import (
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/storage"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
)

type server struct {
	gitalypb.UnimplementedNamespaceServiceServer
	locator storage.Locator
}

// NewServer creates a new instance of a gRPC namespace server
func NewServer(deps *service.Dependencies) gitalypb.NamespaceServiceServer {
	return &server{locator: deps.GetLocator()}
}
