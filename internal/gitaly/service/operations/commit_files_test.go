	"strings"
func TestUserCommitFiles(t *testing.T) {
	ctx, cancel := testhelper.Context()
	ctx, cfg, _, _, client := setupOperationsService(t, ctx)

		targetRelativePath = "target-repository"
	startRepo, startRepoPath, cleanup := gittest.InitBareRepoAt(t, cfg.Storages[0])
	t.Cleanup(cleanup)
	pathToStorage := strings.TrimSuffix(startRepoPath, startRepo.RelativePath)
	repoPath := filepath.Join(pathToStorage, targetRelativePath)
						StorageName:  startRepo.GetStorageName(),
			repo := &gitalypb.Repository{
				StorageName:   startRepo.GetStorageName(),
				RelativePath:  targetRelativePath,
				GlRepository:  gittest.GlRepository,
				GlProjectPath: gittest.GlProjectPath,
			}

					repo,
func TestUserCommitFilesStableCommitID(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, cfg, _, _, client := setupOperationsService(t, ctx)
	repoProto, repoPath, cleanup := gittest.InitBareRepoAt(t, cfg.Storages[0])
	repo := localrepo.New(git.NewExecCommandFactory(cfg), repoProto, cfg)
	for key, values := range testhelper.GitalyServersMetadata(t, cfg.SocketPath) {
	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, cfg, repo, repoPath, client := setupOperationsService(t, ctx)
	newRepo, newRepoPath, newRepoCleanupFn := gittest.InitBareRepoAt(t, cfg.Storages[0])
			repo:          repo,
			repoPath:      repoPath,
			repo:          repo,
			repoPath:      repoPath,
			repo:            repo,
			repoPath:        repoPath,
			headCommit, err := localrepo.New(git.NewExecCommandFactory(cfg), tc.repo, cfg).ReadCommit(ctx, git.Revision(tc.branchName))
	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, cfg, _, _, client := setupOperationsService(t, ctx)
			testRepo, testRepoPath, cleanupFn := gittest.CloneRepoAtStorage(t, cfg.Storages[0], t.Name())
	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, cfg, repoProto, repoPath, client := setupOperationsService(t, ctx)
	repo := localrepo.New(git.NewExecCommandFactory(cfg), repoProto, cfg)
	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, cfg, repoProto, _, client := setupOperationsService(t, ctx)
	repo := localrepo.New(git.NewExecCommandFactory(cfg), repoProto, cfg)
	testSuccessfulUserCommitFilesRemoteRepositoryRequest(func(header *gitalypb.UserCommitFilesRequest) {
	})
	testSuccessfulUserCommitFilesRemoteRepositoryRequest(func(header *gitalypb.UserCommitFilesRequest) {
	})
func testSuccessfulUserCommitFilesRemoteRepositoryRequest(setHeader func(header *gitalypb.UserCommitFilesRequest)) func(*testing.T) {
	return func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()
		ctx, cfg, repoProto, _, client := setupOperationsService(t, ctx)
		gitCmdFactory := git.NewExecCommandFactory(cfg)
		repo := localrepo.New(gitCmdFactory, repoProto, cfg)
		newRepoProto, _, newRepoCleanupFn := gittest.InitBareRepoAt(t, cfg.Storages[0])
		newRepo := localrepo.New(gitCmdFactory, newRepoProto, cfg)
	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, cfg, _, _, client := setupOperationsService(t, ctx)
	repoProto, _, cleanup := gittest.InitBareRepoAt(t, cfg.Storages[0])
	defer cleanup()
	repo := localrepo.New(git.NewExecCommandFactory(cfg), repoProto, cfg)
	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, _, repoProto, repoPath, client := setupOperationsService(t, ctx)
	headerRequest := headerRequest(repoProto, testhelper.TestUser, branchName, commitFilesMessage)
			gittest.WriteCustomHook(t, repoPath, hookName, hookContent)
	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, _, repo, _, client := setupOperationsService(t, ctx)
				headerRequest(repo, testhelper.TestUser, "feature", commitFilesMessage),
				headerRequest(repo, testhelper.TestUser, "feature", commitFilesMessage),
				headerRequest(repo, testhelper.TestUser, "utf-dir", commitFilesMessage),
	ctx, cancel := testhelper.Context()
	defer cancel()
	ctx, _, repo, _, client := setupOperationsService(t, ctx)
			req:  headerRequest(repo, nil, branchName, commitFilesMessage),
			req:  headerRequest(repo, testhelper.TestUser, "", commitFilesMessage),
			req:  headerRequest(repo, testhelper.TestUser, branchName, nil),
			req:  setStartSha(headerRequest(repo, testhelper.TestUser, branchName, commitFilesMessage), "foobar"),
			req:  headerRequest(repo, &gitalypb.User{}, branchName, commitFilesMessage),
			req:  headerRequest(repo, &gitalypb.User{Name: []byte(""), Email: []byte("")}, branchName, commitFilesMessage),
			req:  headerRequest(repo, &gitalypb.User{Name: []byte(" "), Email: []byte(" ")}, branchName, commitFilesMessage),
			req:  headerRequest(repo, &gitalypb.User{Name: []byte("Jane Doe"), Email: []byte("")}, branchName, commitFilesMessage),
			req:  headerRequest(repo, &gitalypb.User{Name: []byte(""), Email: []byte("janedoe@gitlab.com")}, branchName, commitFilesMessage),