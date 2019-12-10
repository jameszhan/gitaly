package repository_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/config"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/service/repository"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func TestReplicateRepository(t *testing.T) {
	tmpPath, cleanup := testhelper.TempDir(t, testhelper.GitlabTestStoragePath(), t.Name())
	defer cleanup()

	replicaPath := filepath.Join(tmpPath, "replica")
	require.NoError(t, os.MkdirAll(replicaPath, 0755))

	defer func(storages []config.Storage) {
		config.Config.Storages = storages
	}(config.Config.Storages)

	config.Config.Storages = []config.Storage{
		config.Storage{
			Name: "default",
			Path: testhelper.GitlabTestStoragePath(),
		},
		config.Storage{
			Name: "replica",
			Path: replicaPath,
		},
	}

	server, serverSocketPath := runFullServer(t)
	defer server.Stop()

	testRepo, testRepoPath, cleanupRepo := testhelper.NewTestRepo(t)
	defer cleanupRepo()

	config.Config.SocketPath = serverSocketPath

	repoClient, conn := repository.NewRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	targetRepo := *testRepo
	targetRepo.StorageName = "replica"

	ctx, cancel := testhelper.Context()
	defer cancel()
	md := testhelper.GitalyServersMetadata(t, serverSocketPath)
	injectedCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := repoClient.ReplicateRepository(injectedCtx, &gitalypb.ReplicateRepositoryRequest{
		Repository: &targetRepo,
		Source:     testRepo,
	})
	require.NoError(t, err)

	targetRepoPath, err := helper.GetRepoPath(&targetRepo)
	require.NoError(t, err)

	testhelper.MustRunCommand(t, nil, "git", "-C", targetRepoPath, "fsck")

	sourceRefs := testhelper.GetRepositoryRefs(t, testRepoPath)
	targetRefs := testhelper.GetRepositoryRefs(t, targetRepoPath)
	require.Equal(t, sourceRefs, targetRefs)

	// create another branch
	_, anotherNewBranch := testhelper.CreateCommitOnNewBranch(t, testRepoPath)
	_, err = repoClient.ReplicateRepository(injectedCtx, &gitalypb.ReplicateRepositoryRequest{
		Repository: &targetRepo,
		Source:     testRepo,
	})
	require.NoError(t, err)
	require.Equal(t,
		testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "show-ref", "--hash", "--verify", fmt.Sprintf("refs/heads/%s", anotherNewBranch)),
		testhelper.MustRunCommand(t, nil, "git", "-C", targetRepoPath, "show-ref", "--hash", "--verify", fmt.Sprintf("refs/heads/%s", anotherNewBranch)),
	)
}

func TestReplicateRepositoryInvalidArguments(t *testing.T) {
	testCases := []struct {
		description   string
		input         *gitalypb.ReplicateRepositoryRequest
		expectedError string
	}{
		{
			description: "everything ✅",
			input: &gitalypb.ReplicateRepositoryRequest{
				Repository: &gitalypb.Repository{
					StorageName:  "praefect-internal-0",
					RelativePath: "/ab/cd/abcdef1234",
				},
				Source: &gitalypb.Repository{
					StorageName:  "praefect-internal-1",
					RelativePath: "/ab/cd/abcdef1234",
				},
			},
			expectedError: "",
		},
		{
			description: "empty repository",
			input: &gitalypb.ReplicateRepositoryRequest{
				Repository: nil,
				Source: &gitalypb.Repository{
					StorageName:  "praefect-internal-1",
					RelativePath: "/ab/cd/abcdef1234",
				},
			},
			expectedError: "repository cannot be empty",
		},
		{
			description: "empty source",
			input: &gitalypb.ReplicateRepositoryRequest{
				Repository: &gitalypb.Repository{
					StorageName:  "praefect-internal-0",
					RelativePath: "/ab/cd/abcdef1234",
				},
				Source: nil,
			},
			expectedError: "repository cannot be empty",
		},
		{
			description: "source and repository have different relative paths",
			input: &gitalypb.ReplicateRepositoryRequest{
				Repository: &gitalypb.Repository{
					StorageName:  "praefect-internal-0",
					RelativePath: "/ab/cd/abcdef1234",
				},
				Source: &gitalypb.Repository{
					StorageName:  "praefect-internal-1",
					RelativePath: "/ab/cd/abcdef4321",
				},
			},
			expectedError: "both source and repository should have the same relative path",
		},
		{
			description: "source and repository have the same storage",
			input: &gitalypb.ReplicateRepositoryRequest{
				Repository: &gitalypb.Repository{
					StorageName:  "praefect-internal-0",
					RelativePath: "/ab/cd/abcdef1234",
				},
				Source: &gitalypb.Repository{
					StorageName:  "praefect-internal-0",
					RelativePath: "/ab/cd/abcdef1234",
				},
			},
			expectedError: "repository and source have the same storage",
		},
	}

	server, serverSocketPath := repository.RunRepoServer(t)
	defer server.Stop()

	client, conn := repository.NewRepositoryClient(t, serverSocketPath)
	defer conn.Close()

	ctx, cancel := testhelper.Context()
	defer cancel()

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := client.ReplicateRepository(ctx, tc.input)
			testhelper.RequireGrpcError(t, err, codes.InvalidArgument)
		})
	}
}
