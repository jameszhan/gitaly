package namespace

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testserver"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestMain(m *testing.M) {
	testhelper.Run(m)
}

func setupNamespaceService(tb testing.TB, opts ...testserver.GitalyServerOpt) (config.Cfg, gitalypb.NamespaceServiceClient) {
	cfgBuilder := testcfg.NewGitalyCfgBuilder(testcfg.WithStorages("default", "other"))
	cfg := cfgBuilder.Build(tb)

	addr := testserver.RunGitalyServer(tb, cfg, func(srv *grpc.Server, deps *service.Dependencies) {
		gitalypb.RegisterNamespaceServiceServer(srv, NewServer(deps))
	}, opts...)

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(tb, err)
	tb.Cleanup(func() { testhelper.MustClose(tb, conn) })

	return cfg, gitalypb.NewNamespaceServiceClient(conn)
}
