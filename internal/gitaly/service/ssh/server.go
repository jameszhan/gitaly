package ssh

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/storage"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/transaction"
	"gitlab.com/gitlab-org/gitaly/v16/internal/helper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/log"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
)

var (
	defaultUploadPackRequestTimeout    = 10 * time.Minute
	defaultUploadArchiveRequestTimeout = time.Minute
)

type server struct {
	gitalypb.UnimplementedSSHServiceServer
	logger                                   log.Logger
	locator                                  storage.Locator
	gitCmdFactory                            git.CommandFactory
	txManager                                transaction.Manager
	uploadPackRequestTimeoutTickerFactory    func() helper.Ticker
	uploadArchiveRequestTimeoutTickerFactory func() helper.Ticker
	packfileNegotiationMetrics               *prometheus.CounterVec
}

// NewServer creates a new instance of a grpc SSHServer
func NewServer(deps *service.Dependencies, serverOpts ...ServerOpt) gitalypb.SSHServiceServer {
	s := &server{
		logger:        deps.GetLogger(),
		locator:       deps.GetLocator(),
		gitCmdFactory: deps.GetGitCmdFactory(),
		txManager:     deps.GetTxManager(),
		uploadPackRequestTimeoutTickerFactory: func() helper.Ticker {
			return helper.NewTimerTicker(defaultUploadPackRequestTimeout)
		},
		uploadArchiveRequestTimeoutTickerFactory: func() helper.Ticker {
			return helper.NewTimerTicker(defaultUploadArchiveRequestTimeout)
		},
		packfileNegotiationMetrics: prometheus.NewCounterVec(
			prometheus.CounterOpts{},
			[]string{"git_negotiation_feature"},
		),
	}

	for _, serverOpt := range serverOpts {
		serverOpt(s)
	}

	return s
}

// ServerOpt is a self referential option for server
type ServerOpt func(s *server)

// WithUploadPackRequestTimeoutTickerFactory sets the upload pack request timeout ticker factory.
func WithUploadPackRequestTimeoutTickerFactory(factory func() helper.Ticker) ServerOpt {
	return func(s *server) {
		s.uploadPackRequestTimeoutTickerFactory = factory
	}
}

// WithArchiveRequestTimeoutTickerFactory sets the upload pack request timeout ticker factory.
func WithArchiveRequestTimeoutTickerFactory(factory func() helper.Ticker) ServerOpt {
	return func(s *server) {
		s.uploadArchiveRequestTimeoutTickerFactory = factory
	}
}

//nolint:revive // This is unintentionally missing documentation.
func WithPackfileNegotiationMetrics(c *prometheus.CounterVec) ServerOpt {
	return func(s *server) {
		s.packfileNegotiationMetrics = c
	}
}
