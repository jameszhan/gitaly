package hook

import (
	"errors"
	"os/exec"
	"sync"

	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v15/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/v15/streamio"
)

func validateUpdateHookRequest(in *gitalypb.UpdateHookRequest) error {
	return service.ValidateRepository(in.GetRepository())
}

func (s *server) UpdateHook(in *gitalypb.UpdateHookRequest, stream gitalypb.HookService_UpdateHookServer) error {
	if err := validateUpdateHookRequest(in); err != nil {
		return structerr.NewInvalidArgument("%w", err)
	}

	var m sync.Mutex
	stdout := streamio.NewSyncWriter(&m, func(p []byte) error {
		return stream.Send(&gitalypb.UpdateHookResponse{Stdout: p})
	})
	stderr := streamio.NewSyncWriter(&m, func(p []byte) error {
		return stream.Send(&gitalypb.UpdateHookResponse{Stderr: p})
	})

	if err := s.manager.UpdateHook(
		stream.Context(),
		in.GetRepository(),
		string(in.GetRef()),
		in.GetOldValue(),
		in.GetNewValue(),
		in.GetEnvironmentVariables(),
		stdout,
		stderr,
	); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return updateHookResponse(stream, int32(exitError.ExitCode()))
		}

		return structerr.NewInternal("%w", err)
	}

	return updateHookResponse(stream, 0)
}

func updateHookResponse(stream gitalypb.HookService_UpdateHookServer, code int32) error {
	if err := stream.Send(&gitalypb.UpdateHookResponse{
		ExitStatus: &gitalypb.ExitStatus{Value: code},
	}); err != nil {
		return structerr.NewInternal("sending response: %w", err)
	}

	return nil
}
