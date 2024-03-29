package smarthttp

import (
	"context"
	"io"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.com/gitlab-org/gitaly/v16/internal/cache"
	"gitlab.com/gitlab-org/gitaly/v16/internal/log"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
)

type infoRefCache struct {
	streamer cache.Streamer
}

func newInfoRefCache(streamer cache.Streamer) infoRefCache {
	return infoRefCache{
		streamer: streamer,
	}
}

var (
	// prometheus counters
	cacheAttemptTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "gitaly_inforef_cache_attempt_total",
			Help: "Total number of smarthttp info-ref RPCs accessing the cache",
		},
	)
	hitMissTotals = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitaly_inforef_cache_hit_miss_total",
			Help: "Total number of smarthttp info-ref RPC hit/miss/err cache accesses",
		},
		[]string{"type"},
	)

	// counter functions are package vars to enable easy overriding for tests
	countAttempt = func() { cacheAttemptTotal.Inc() }
	countHit     = func() { hitMissTotals.WithLabelValues("hit").Inc() }
	countMiss    = func() { hitMissTotals.WithLabelValues("miss").Inc() }
	countErr     = func() { hitMissTotals.WithLabelValues("err").Inc() }
)

func (c infoRefCache) tryCache(ctx context.Context, in *gitalypb.InfoRefsRequest, w io.Writer, missFn func(io.Writer) error) error {
	if len(in.GetGitConfigOptions()) > 0 ||
		len(in.GetGitProtocol()) > 0 {
		return missFn(w)
	}

	logger := log.FromContext(ctx).WithFields(log.Fields{"service": uploadPackSvc})
	logger.Debug("Attempting to fetch cached response")
	countAttempt()

	stream, err := c.streamer.GetStream(ctx, in.GetRepository(), in)
	switch err {
	case nil:
		defer stream.Close()

		countHit()
		logger.Info("cache hit for UploadPack response")

		if _, err := io.Copy(w, stream); err != nil {
			return structerr.NewInternal("cache copy: %w", err)
		}

		return nil

	case cache.ErrReqNotFound:
		countMiss()
		logger.Info("cache miss for InfoRefsUploadPack response")

		var wg sync.WaitGroup
		defer wg.Wait()

		pr, pw := io.Pipe()

		wg.Add(1)
		go func() {
			defer wg.Done()

			tr := io.TeeReader(pr, w)
			if err := c.streamer.PutStream(ctx, in.Repository, in, tr); err != nil {
				logger.WithError(err).Error("unable to store InfoRefsUploadPack response in cache")

				// discard remaining bytes if caching stream
				// failed so that tee reader is not blocked
				_, err = io.Copy(io.Discard, tr)
				if err != nil {
					logger.WithError(err).
						Error("unable to discard remaining InfoRefsUploadPack cache stream")
				}
			}
		}()

		err = missFn(pw)
		_ = pw.CloseWithError(err) // always returns nil
		return err

	default:
		countErr()
		logger.WithError(err).Info("unable to fetch cached response")

		return missFn(w)
	}
}
