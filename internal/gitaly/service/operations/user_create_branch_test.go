//go:build !gitaly_test_sha256

package operations

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/localrepo"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/service"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/service/hook"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/storage"
	"gitlab.com/gitlab-org/gitaly/v16/internal/grpc/backchannel"
	"gitlab.com/gitlab-org/gitaly/v16/internal/grpc/metadata"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper/testserver"
	"gitlab.com/gitlab-org/gitaly/v16/internal/transaction/txinfo"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserCreateBranch_successful(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	ctx, cfg, repoProto, repoPath, client := setupOperationsService(t, ctx)

	repo := localrepo.NewTestRepo(t, cfg, repoProto)

	startPoint := "c7fbe50c7c7419d9701eebe64b1fdacc3df5b9dd"
	startPointCommit, err := repo.ReadCommit(ctx, git.Revision(startPoint))
	require.NoError(t, err)

	testCases := []struct {
		desc           string
		branchName     string
		startPoint     string
		expectedBranch *gitalypb.Branch
	}{
		{
			desc:       "valid branch",
			branchName: "new-branch",
			startPoint: startPoint,
			expectedBranch: &gitalypb.Branch{
				Name:         []byte("new-branch"),
				TargetCommit: startPointCommit,
			},
		},
		// On input like heads/foo and refs/heads/foo we don't
		// DWYM and map it to refs/heads/foo and
		// refs/heads/foo, respectively. Instead we always
		// prepend refs/heads/*, so you get
		// refs/heads/heads/foo and refs/heads/refs/heads/foo
		{
			desc:       "valid branch",
			branchName: "heads/new-branch",
			startPoint: startPoint,
			expectedBranch: &gitalypb.Branch{
				Name:         []byte("heads/new-branch"),
				TargetCommit: startPointCommit,
			},
		},
		{
			desc:       "valid branch",
			branchName: "refs/heads/new-branch",
			startPoint: startPoint,
			expectedBranch: &gitalypb.Branch{
				Name:         []byte("refs/heads/new-branch"),
				TargetCommit: startPointCommit,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			branchName := testCase.branchName
			request := &gitalypb.UserCreateBranchRequest{
				Repository: repoProto,
				BranchName: []byte(branchName),
				StartPoint: []byte(testCase.startPoint),
				User:       gittest.TestUser,
			}

			response, err := client.UserCreateBranch(ctx, request)
			if testCase.expectedBranch != nil {
				defer gittest.Exec(t, cfg, "-C", repoPath, "branch", "-D", branchName)
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expectedBranch, response.Branch)

			branches := gittest.Exec(t, cfg, "-C", repoPath, "for-each-ref", "--", "refs/heads/"+branchName)
			require.Contains(t, string(branches), "refs/heads/"+branchName)
		})
	}
}

func TestUserCreateBranch_Transactions(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	cfg := testcfg.Build(t)

	repo, repoPath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
		SkipCreationViaService: true,
		Seed:                   gittest.SeedGitLabTest,
	})

	transactionServer := &testTransactionServer{}

	cfg.ListenAddr = "127.0.0.1:0" // runs gitaly on the TCP address
	addr := testserver.RunGitalyServer(t, cfg, func(srv *grpc.Server, deps *service.Dependencies) {
		gitalypb.RegisterOperationServiceServer(srv, NewServer(
			deps.GetHookManager(),
			deps.GetTxManager(),
			deps.GetLocator(),
			deps.GetConnsPool(),
			deps.GetGit2goExecutor(),
			deps.GetGitCmdFactory(),
			deps.GetCatfileCache(),
			deps.GetUpdaterWithHooks(),
			deps.GetCfg().Git.SigningKey,
		))
		gitalypb.RegisterHookServiceServer(srv, hook.NewServer(
			deps.GetHookManager(),
			deps.GetLocator(),
			deps.GetGitCmdFactory(),
			deps.GetPackObjectsCache(),
			deps.GetPackObjectsLimiter(),
		))
		// Praefect proxy execution disabled as praefect runs only on the UNIX socket, but
		// the test requires a TCP listening address.
	}, testserver.WithDisablePraefect())

	testcases := []struct {
		desc    string
		address string
	}{
		{
			desc:    "TCP address",
			address: addr,
		},
		{
			desc:    "Unix socket",
			address: "unix://" + cfg.InternalSocketPath(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			defer gittest.Exec(t, cfg, "-C", repoPath, "branch", "-D", "new-branch")

			ctx, err := txinfo.InjectTransaction(ctx, 1, "node", true)
			require.NoError(t, err)
			ctx = metadata.IncomingToOutgoing(ctx)

			client := newMuxedOperationClient(t, ctx, tc.address, cfg.Auth.Token,
				backchannel.NewClientHandshaker(
					testhelper.NewDiscardingLogEntry(t),
					func() backchannel.Server {
						srv := grpc.NewServer()
						gitalypb.RegisterRefTransactionServer(srv, transactionServer)
						return srv
					},
					backchannel.DefaultConfiguration(),
				),
			)

			request := &gitalypb.UserCreateBranchRequest{
				Repository: repo,
				BranchName: []byte("new-branch"),
				StartPoint: []byte("c7fbe50c7c7419d9701eebe64b1fdacc3df5b9dd"),
				User:       gittest.TestUser,
			}

			transactionServer.called = 0
			_, err = client.UserCreateBranch(ctx, request)
			require.NoError(t, err)
			require.Equal(t, 5, transactionServer.called)
		})
	}
}

func TestUserCreateBranch_hook(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	ctx, cfg, repo, repoPath, client := setupOperationsService(t, ctx)

	branchName := "new-branch"
	request := &gitalypb.UserCreateBranchRequest{
		Repository: repo,
		BranchName: []byte(branchName),
		StartPoint: []byte("c7fbe50c7c7419d9701eebe64b1fdacc3df5b9dd"),
		User:       gittest.TestUser,
	}

	for _, hookName := range GitlabHooks {
		t.Run(hookName, func(t *testing.T) {
			defer gittest.Exec(t, cfg, "-C", repoPath, "branch", "-D", branchName)

			hookOutputTempPath := gittest.WriteEnvToCustomHook(t, repoPath, hookName)

			_, err := client.UserCreateBranch(ctx, request)
			require.NoError(t, err)

			output := string(testhelper.MustReadFile(t, hookOutputTempPath))
			require.Contains(t, output, "GL_USERNAME="+gittest.TestUser.GlUsername)
		})
	}
}

func TestUserCreateBranch_startPoint(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	ctx, cfg, repoProto, repoPath, client := setupOperationsService(t, ctx)

	repo := localrepo.NewTestRepo(t, cfg, repoProto)

	testCases := []struct {
		desc             string
		branchName       string
		startPoint       string
		startPointCommit string
		user             *gitalypb.User
	}{
		// Similar to prefixing branchName in
		// TestSuccessfulCreateBranchRequest() above:
		// Unfortunately (and inconsistently), the StartPoint
		// reference does have DWYM semantics. See
		// https://gitlab.com/gitlab-org/gitaly/-/issues/3331
		{
			desc:             "the StartPoint parameter does DWYM references (boo!)",
			branchName:       "topic",
			startPoint:       "heads/master",
			startPointCommit: "9a944d90955aaf45f6d0c88f30e27f8d2c41cec0", // TODO: see below
			user:             gittest.TestUser,
		},
		{
			desc:             "the StartPoint parameter does DWYM references (boo!) 2",
			branchName:       "topic2",
			startPoint:       "refs/heads/master",
			startPointCommit: "c642fe9b8b9f28f9225d7ea953fe14e74748d53b", // TODO: see below
			user:             gittest.TestUser,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			gittest.Exec(t, cfg, "-C", repoPath, "update-ref", "refs/heads/"+testCase.startPoint,
				testCase.startPointCommit,
				git.ObjectHashSHA1.ZeroOID.String(),
			)
			request := &gitalypb.UserCreateBranchRequest{
				Repository: repoProto,
				BranchName: []byte(testCase.branchName),
				StartPoint: []byte(testCase.startPoint),
				User:       testCase.user,
			}

			// BEGIN TODO: Uncomment if StartPoint started behaving sensibly
			// like BranchName. See
			// https://gitlab.com/gitlab-org/gitaly/-/issues/3331
			//
			//targetCommitOK, err := repo.ReadCommit(ctx, testCase.startPointCommit)
			// END TODO
			targetCommitOK, err := repo.ReadCommit(ctx, "1e292f8fedd741b75372e19097c76d327140c312")
			require.NoError(t, err)

			response, err := client.UserCreateBranch(ctx, request)
			require.NoError(t, err)
			responseOk := &gitalypb.UserCreateBranchResponse{
				Branch: &gitalypb.Branch{
					Name:         []byte(testCase.branchName),
					TargetCommit: targetCommitOK,
				},
			}
			testhelper.ProtoEqual(t, responseOk, response)
			branches := gittest.Exec(t, cfg, "-C", repoPath, "for-each-ref", "--", "refs/heads/"+testCase.branchName)
			require.Contains(t, string(branches), "refs/heads/"+testCase.branchName)
		})
	}
}

func TestUserCreateBranch_hookFailure(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	ctx, _, repo, repoPath, client := setupOperationsService(t, ctx)

	request := &gitalypb.UserCreateBranchRequest{
		Repository: repo,
		BranchName: []byte("new-branch"),
		StartPoint: []byte("c7fbe50c7c7419d9701eebe64b1fdacc3df5b9dd"),
		User:       gittest.TestUser,
	}

	hookContent := []byte("#!/bin/sh\necho GL_ID=$GL_ID\nexit 1")

	expectedObject := "GL_ID=" + gittest.TestUser.GlId

	for _, hookName := range gitlabPreHooks {
		gittest.WriteCustomHook(t, repoPath, hookName, hookContent)

		_, err := client.UserCreateBranch(ctx, request)

		testhelper.RequireGrpcError(t, structerr.NewPermissionDenied("creation denied by custom hooks").WithDetail(
			&gitalypb.UserCreateBranchError{
				Error: &gitalypb.UserCreateBranchError_CustomHook{
					CustomHook: &gitalypb.CustomHookError{
						HookType: gitalypb.CustomHookError_HOOK_TYPE_PRERECEIVE,
						Stdout:   []byte(expectedObject + "\n"),
					},
				},
			},
		), err)

	}
}

func TestUserCreateBranch_Failure(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	ctx, _, repo, _, client := setupOperationsService(t, ctx)

	testCases := []struct {
		desc       string
		repo       *gitalypb.Repository
		branchName string
		startPoint string
		user       *gitalypb.User
		err        error
	}{
		{
			desc:       "repository not provided",
			repo:       nil,
			branchName: "shiny-new-branch",
			startPoint: "",
			user:       gittest.TestUser,
			err:        structerr.NewInvalidArgument("%w", storage.ErrRepositoryNotSet),
		},
		{
			desc:       "empty start_point",
			repo:       repo,
			branchName: "shiny-new-branch",
			startPoint: "",
			user:       gittest.TestUser,
			err:        status.Error(codes.InvalidArgument, "empty start point"),
		},
		{
			desc:       "empty user",
			repo:       repo,
			branchName: "shiny-new-branch",
			startPoint: "master",
			user:       nil,
			err:        status.Error(codes.InvalidArgument, "empty user"),
		},
		{
			desc:       "non-existing starting point",
			repo:       repo,
			branchName: "new-branch",
			startPoint: "i-dont-exist",
			user:       gittest.TestUser,
			err:        status.Errorf(codes.FailedPrecondition, "revspec '%s' not found", "i-dont-exist"),
		},
		{
			desc:       "branch exists",
			repo:       repo,
			branchName: "master",
			startPoint: "master",
			user:       gittest.TestUser,
			err: testhelper.WithInterceptedMetadata(
				structerr.NewFailedPrecondition("reference update: state update to %q failed: %w", "prepare", io.EOF),
				"stderr",
				"fatal: prepare: cannot lock ref 'refs/heads/master': reference already exists\n",
			),
		},
		{
			desc:       "conflicting with refs/heads/improve/awesome",
			repo:       repo,
			branchName: "improve",
			startPoint: "master",
			user:       gittest.TestUser,
			err: testhelper.WithInterceptedMetadataItems(
				structerr.NewFailedPrecondition("reference update: file directory conflict"),
				structerr.MetadataItem{Key: "conflicting_reference", Value: "refs/heads/improve"},
				structerr.MetadataItem{Key: "existing_reference", Value: "refs/heads/improve/awesome"},
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			request := &gitalypb.UserCreateBranchRequest{
				Repository: testCase.repo,
				BranchName: []byte(testCase.branchName),
				StartPoint: []byte(testCase.startPoint),
				User:       testCase.user,
			}

			response, err := client.UserCreateBranch(ctx, request)
			testhelper.RequireGrpcError(t, testCase.err, err)
			require.Empty(t, response)
		})
	}
}
