package stats

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/localrepo"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v15/internal/testhelper/testcfg"
	"gitlab.com/gitlab-org/gitaly/v15/proto/go/gitalypb"
)

func TestRepositoryProfile(t *testing.T) {
	ctx := testhelper.Context(t)
	cfg := testcfg.Build(t)

	repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
		SkipCreationViaService: true,
	})
	repo := localrepo.NewTestRepo(t, cfg, repoProto)

	hasBitmap, err := HasBitmap(repoPath)
	require.NoError(t, err)
	require.False(t, hasBitmap, "repository should not have a bitmap initially")
	packfiles, err := GetPackfiles(repoPath)
	require.NoError(t, err)
	require.Empty(t, packfiles)
	packfilesCount, err := PackfilesCount(repoPath)
	require.NoError(t, err)
	require.Zero(t, packfilesCount)

	blobs := 10
	blobIDs := gittest.WriteBlobs(t, cfg, repoPath, blobs)

	looseObjects, err := LooseObjects(ctx, repo)
	require.NoError(t, err)
	require.Equal(t, uint64(blobs), looseObjects)

	for _, blobID := range blobIDs {
		commitID := gittest.WriteCommit(t, cfg, repoPath,
			gittest.WithTreeEntries(gittest.TreeEntry{
				Mode: "100644", Path: "blob", OID: git.ObjectID(blobID),
			}),
		)
		gittest.Exec(t, cfg, "-C", repoPath, "update-ref", "refs/heads/"+blobID, commitID.String())
	}

	// write a loose object
	gittest.WriteBlobs(t, cfg, repoPath, 1)

	gittest.Exec(t, cfg, "-C", repoPath, "repack", "-A", "-b", "-d")

	looseObjects, err = LooseObjects(ctx, repo)
	require.NoError(t, err)
	require.Equal(t, uint64(1), looseObjects)

	// write another loose object
	blobID := gittest.WriteBlobs(t, cfg, repoPath, 1)[0]

	// due to OS semantics, ensure that the blob has a timestamp that is after the packfile
	theFuture := time.Now().Add(10 * time.Minute)
	require.NoError(t, os.Chtimes(filepath.Join(repoPath, "objects", blobID[0:2], blobID[2:]), theFuture, theFuture))

	looseObjects, err = LooseObjects(ctx, repo)
	require.NoError(t, err)
	require.Equal(t, uint64(2), looseObjects)
}

func TestLogObjectInfo(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	cfg := testcfg.Build(t)

	locator := config.NewLocator(cfg)
	storagePath, err := locator.GetStorageByName(cfg.Storages[0].Name)
	require.NoError(t, err)

	requireObjectsInfo := func(entries []*logrus.Entry) ObjectsInfo {
		for _, entry := range entries {
			if entry.Message == "repository objects info" {
				objectsInfo, ok := entry.Data["objects_info"]
				require.True(t, ok)
				require.IsType(t, ObjectsInfo{}, objectsInfo)
				return objectsInfo.(ObjectsInfo)
			}
		}

		require.FailNow(t, "no objects info log entry found")
		return ObjectsInfo{}
	}

	t.Run("shared repo with multiple alternates", func(t *testing.T) {
		t.Parallel()

		logger, hook := test.NewNullLogger()
		ctx := ctxlogrus.ToContext(ctx, logger.WithField("test", "logging"))

		_, repoPath1 := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
			SkipCreationViaService: true,
		})
		gittest.WriteCommit(t, cfg, repoPath1, gittest.WithMessage("repo1"), gittest.WithBranch("main"))

		_, repoPath2 := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
			SkipCreationViaService: true,
		})
		gittest.WriteCommit(t, cfg, repoPath2, gittest.WithMessage("repo2"), gittest.WithBranch("main"))

		// clone existing local repo with two alternates
		targetRepoName := gittest.NewRepositoryName(t)
		targetRepoPath := filepath.Join(storagePath, targetRepoName)
		gittest.Exec(t, cfg, "clone", "--bare", "--shared", repoPath1, "--reference", repoPath1, "--reference", repoPath2, targetRepoPath)

		LogObjectsInfo(ctx, localrepo.NewTestRepo(t, cfg, &gitalypb.Repository{
			StorageName:  cfg.Storages[0].Name,
			RelativePath: targetRepoName,
		}))

		objectsInfo := requireObjectsInfo(hook.AllEntries())
		require.Equal(t, ObjectsInfo{
			Alternates: []string{
				filepath.Join(repoPath1, "/objects"),
				filepath.Join(repoPath2, "/objects"),
			},
		}, objectsInfo)
	})

	t.Run("repo without alternates", func(t *testing.T) {
		t.Parallel()

		logger, hook := test.NewNullLogger()
		ctx := ctxlogrus.ToContext(ctx, logger.WithField("test", "logging"))

		repo, repoPath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
			SkipCreationViaService: true,
		})
		gittest.WriteCommit(t, cfg, repoPath, gittest.WithBranch("main"))

		LogObjectsInfo(ctx, localrepo.NewTestRepo(t, cfg, repo))

		objectsInfo := requireObjectsInfo(hook.AllEntries())
		require.Equal(t, ObjectsInfo{
			LooseObjects:     2,
			LooseObjectsSize: 8,
		}, objectsInfo)
	})
}

func TestObjectsInfoForRepository(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	cfg := testcfg.Build(t)

	_, alternatePath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
		SkipCreationViaService: true,
	})
	alternatePath = filepath.Join(alternatePath, "objects")

	for _, tc := range []struct {
		desc                string
		setup               func(t *testing.T, repoPath string)
		expectedErr         error
		expectedObjectsInfo ObjectsInfo
	}{
		{
			desc: "empty repository",
			setup: func(*testing.T, string) {
			},
		},
		{
			desc: "single blob",
			setup: func(t *testing.T, repoPath string) {
				gittest.WriteBlob(t, cfg, repoPath, []byte("x"))
			},
			expectedObjectsInfo: ObjectsInfo{
				LooseObjects:     1,
				LooseObjectsSize: 4,
			},
		},
		{
			desc: "single packed blob",
			setup: func(t *testing.T, repoPath string) {
				blobID := gittest.WriteBlob(t, cfg, repoPath, []byte("x"))
				gittest.WriteRef(t, cfg, repoPath, "refs/tags/blob", blobID)
				// We use `-d`, which also prunes objects that have been packed.
				gittest.Exec(t, cfg, "-C", repoPath, "repack", "-Ad")
			},
			expectedObjectsInfo: ObjectsInfo{
				PackedObjects:        1,
				Packfiles:            1,
				PackfilesSize:        1,
				PackfileBitmapExists: true,
			},
		},
		{
			desc: "single pruneable blob",
			setup: func(t *testing.T, repoPath string) {
				blobID := gittest.WriteBlob(t, cfg, repoPath, []byte("x"))
				gittest.WriteRef(t, cfg, repoPath, "refs/tags/blob", blobID)
				// This time we don't use `-d`, so the object will exist both in
				// loose and packed form.
				gittest.Exec(t, cfg, "-C", repoPath, "repack", "-a")
			},
			expectedObjectsInfo: ObjectsInfo{
				LooseObjects:         1,
				LooseObjectsSize:     4,
				PackedObjects:        1,
				Packfiles:            1,
				PackfilesSize:        1,
				PackfileBitmapExists: true,
				PruneableObjects:     1,
			},
		},
		{
			desc: "garbage",
			setup: func(t *testing.T, repoPath string) {
				garbagePath := filepath.Join(repoPath, "objects", "pack", "garbage")
				require.NoError(t, os.WriteFile(garbagePath, []byte("x"), 0o600))
			},
			expectedObjectsInfo: ObjectsInfo{
				Garbage: 1,
				// git-count-objects(1) somehow does not count this file's size,
				// which I've verified manually.
				GarbageSize: 0,
			},
		},
		{
			desc: "alternates",
			setup: func(t *testing.T, repoPath string) {
				infoAlternatesPath := filepath.Join(repoPath, "objects", "info", "alternates")
				require.NoError(t, os.WriteFile(infoAlternatesPath, []byte(alternatePath), 0o600))
			},
			expectedObjectsInfo: ObjectsInfo{
				Alternates: []string{
					alternatePath,
				},
			},
		},
		{
			desc: "non-split commit-graph without bloom filter",
			setup: func(t *testing.T, repoPath string) {
				gittest.WriteCommit(t, cfg, repoPath, gittest.WithBranch("main"))
				gittest.Exec(t, cfg, "-C", repoPath, "commit-graph", "write", "--reachable")
			},
			expectedObjectsInfo: ObjectsInfo{
				LooseObjects:     2,
				LooseObjectsSize: 8,
				CommitGraph: CommitGraphInfo{
					Exists: true,
				},
			},
		},
		{
			desc: "non-split commit-graph with bloom filter",
			setup: func(t *testing.T, repoPath string) {
				gittest.WriteCommit(t, cfg, repoPath, gittest.WithBranch("main"))
				gittest.Exec(t, cfg, "-C", repoPath, "commit-graph", "write", "--reachable", "--changed-paths")
			},
			expectedObjectsInfo: ObjectsInfo{
				LooseObjects:     2,
				LooseObjectsSize: 8,
				CommitGraph: CommitGraphInfo{
					Exists:          true,
					HasBloomFilters: true,
				},
			},
		},
		{
			desc: "all together",
			setup: func(t *testing.T, repoPath string) {
				infoAlternatesPath := filepath.Join(repoPath, "objects", "info", "alternates")
				require.NoError(t, os.WriteFile(infoAlternatesPath, []byte(alternatePath), 0o600))

				// We write a single packed blob.
				blobID := gittest.WriteBlob(t, cfg, repoPath, []byte("x"))
				gittest.WriteRef(t, cfg, repoPath, "refs/tags/blob", blobID)
				gittest.Exec(t, cfg, "-C", repoPath, "repack", "-Ad")

				// And two loose ones.
				gittest.WriteBlob(t, cfg, repoPath, []byte("1"))
				gittest.WriteBlob(t, cfg, repoPath, []byte("2"))

				// And three garbage-files. This is done so we've got unique counts
				// everywhere.
				for _, file := range []string{"garbage1", "garbage2", "garbage3"} {
					garbagePath := filepath.Join(repoPath, "objects", "pack", file)
					require.NoError(t, os.WriteFile(garbagePath, []byte("x"), 0o600))
				}
			},
			expectedObjectsInfo: ObjectsInfo{
				LooseObjects:         2,
				LooseObjectsSize:     8,
				PackedObjects:        1,
				Packfiles:            1,
				PackfilesSize:        1,
				PackfileBitmapExists: true,
				Garbage:              3,
				Alternates: []string{
					alternatePath,
				},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
				SkipCreationViaService: true,
			})
			repo := localrepo.NewTestRepo(t, cfg, repoProto)

			tc.setup(t, repoPath)

			objectsInfo, err := ObjectsInfoForRepository(ctx, repo)
			require.Equal(t, tc.expectedErr, err)
			require.Equal(t, tc.expectedObjectsInfo, objectsInfo)
		})
	}
}

func TestCountLooseObjects(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	cfg := testcfg.Build(t)

	createRepo := func(t *testing.T) (*localrepo.Repo, string) {
		repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
			SkipCreationViaService: true,
		})
		return localrepo.NewTestRepo(t, cfg, repoProto), repoPath
	}

	requireLooseObjectsInfo := func(t *testing.T, repo *localrepo.Repo, cutoff time.Time, expectedInfo LooseObjectsInfo) {
		info, err := LooseObjectsInfoForRepository(repo, cutoff)
		require.NoError(t, err)
		require.Equal(t, expectedInfo, info)
	}

	t.Run("empty repository", func(t *testing.T) {
		repo, _ := createRepo(t)
		requireLooseObjectsInfo(t, repo, time.Now(), LooseObjectsInfo{})
	})

	t.Run("object in random shard", func(t *testing.T) {
		repo, repoPath := createRepo(t)

		differentShard := filepath.Join(repoPath, "objects", "a0")
		require.NoError(t, os.MkdirAll(differentShard, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(differentShard, "123456"), []byte("foobar"), 0o644))

		requireLooseObjectsInfo(t, repo, time.Now(), LooseObjectsInfo{
			Count:      1,
			Size:       6,
			StaleCount: 1,
			StaleSize:  6,
		})
	})

	t.Run("objects in multiple shards", func(t *testing.T) {
		repo, repoPath := createRepo(t)

		for i, shard := range []string{"00", "17", "32", "ff"} {
			shardPath := filepath.Join(repoPath, "objects", shard)
			require.NoError(t, os.MkdirAll(shardPath, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(shardPath, "123456"), make([]byte, i), 0o644))
		}

		requireLooseObjectsInfo(t, repo, time.Now(), LooseObjectsInfo{
			Count:      4,
			Size:       6,
			StaleCount: 4,
			StaleSize:  6,
		})
	})

	t.Run("object in shard with grace period", func(t *testing.T) {
		repo, repoPath := createRepo(t)

		shard := filepath.Join(repoPath, "objects", "17")
		require.NoError(t, os.MkdirAll(shard, 0o755))

		objectPaths := []string{
			filepath.Join(shard, "123456"),
			filepath.Join(shard, "654321"),
		}

		cutoffDate := time.Now()
		afterCutoffDate := cutoffDate.Add(1 * time.Minute)
		beforeCutoffDate := cutoffDate.Add(-1 * time.Minute)

		for _, objectPath := range objectPaths {
			require.NoError(t, os.WriteFile(objectPath, []byte("1"), 0o644))
			require.NoError(t, os.Chtimes(objectPath, afterCutoffDate, afterCutoffDate))
		}

		// Objects are recent, so with the cutoff-date they shouldn't be counted.
		requireLooseObjectsInfo(t, repo, time.Now(), LooseObjectsInfo{
			Count: 2,
			Size:  2,
		})

		for i, objectPath := range objectPaths {
			// Modify the object's mtime should cause it to be counted.
			require.NoError(t, os.Chtimes(objectPath, beforeCutoffDate, beforeCutoffDate))

			requireLooseObjectsInfo(t, repo, time.Now(), LooseObjectsInfo{
				Count:      2,
				Size:       2,
				StaleCount: uint64(i) + 1,
				StaleSize:  uint64(i) + 1,
			})
		}
	})

	t.Run("shard with garbage", func(t *testing.T) {
		repo, repoPath := createRepo(t)

		shard := filepath.Join(repoPath, "objects", "17")
		require.NoError(t, os.MkdirAll(shard, 0o755))

		for _, objectName := range []string{"garbage", "012345"} {
			require.NoError(t, os.WriteFile(filepath.Join(shard, objectName), nil, 0o644))
		}

		requireLooseObjectsInfo(t, repo, time.Now(), LooseObjectsInfo{
			Count:      1,
			StaleCount: 1,
		})
	})
}

func BenchmarkCountLooseObjects(b *testing.B) {
	ctx := testhelper.Context(b)
	cfg := testcfg.Build(b)

	createRepo := func(b *testing.B) (*localrepo.Repo, string) {
		repoProto, repoPath := gittest.CreateRepository(b, ctx, cfg, gittest.CreateRepositoryConfig{
			SkipCreationViaService: true,
		})
		return localrepo.NewTestRepo(b, cfg, repoProto), repoPath
	}

	b.Run("empty repository", func(b *testing.B) {
		repo, _ := createRepo(b)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := LooseObjectsInfoForRepository(repo, time.Now())
			require.NoError(b, err)
		}
	})

	b.Run("repository with single object", func(b *testing.B) {
		repo, repoPath := createRepo(b)

		objectPath := filepath.Join(repoPath, "objects", "17", "12345")
		require.NoError(b, os.Mkdir(filepath.Dir(objectPath), 0o755))
		require.NoError(b, os.WriteFile(objectPath, nil, 0o644))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := LooseObjectsInfoForRepository(repo, time.Now())
			require.NoError(b, err)
		}
	})

	b.Run("repository with single object in each shard", func(b *testing.B) {
		repo, repoPath := createRepo(b)

		for i := 0; i < 256; i++ {
			objectPath := filepath.Join(repoPath, "objects", fmt.Sprintf("%02x", i), "12345")
			require.NoError(b, os.Mkdir(filepath.Dir(objectPath), 0o755))
			require.NoError(b, os.WriteFile(objectPath, nil, 0o644))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := LooseObjectsInfoForRepository(repo, time.Now())
			require.NoError(b, err)
		}
	})

	b.Run("repository hitting loose object limit", func(b *testing.B) {
		repo, repoPath := createRepo(b)

		// Usually we shouldn't have a lot more than `looseObjectCount` objects in the
		// repository because we'd repack as soon as we hit that limit. So this benchmark
		// case tries to estimate the usual upper limit for loose objects we'd typically
		// have.
		//
		// Note that we should ideally just use `housekeeping.looseObjectsLimit` here to
		// derive that value. But due to a cyclic dependency that's not possible, so we
		// just use a hard-coded value instead.
		looseObjectCount := 5

		for i := 0; i < 256; i++ {
			shardPath := filepath.Join(repoPath, "objects", fmt.Sprintf("%02x", i))
			require.NoError(b, os.Mkdir(shardPath, 0o755))

			for j := 0; j < looseObjectCount; j++ {
				objectPath := filepath.Join(shardPath, fmt.Sprintf("%d", j))
				require.NoError(b, os.WriteFile(objectPath, nil, 0o644))
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := LooseObjectsInfoForRepository(repo, time.Now())
			require.NoError(b, err)
		}
	})

	b.Run("repository with lots of objects", func(b *testing.B) {
		repo, repoPath := createRepo(b)

		for i := 0; i < 256; i++ {
			shardPath := filepath.Join(repoPath, "objects", fmt.Sprintf("%02x", i))
			require.NoError(b, os.Mkdir(shardPath, 0o755))

			for j := 0; j < 1000; j++ {
				objectPath := filepath.Join(shardPath, fmt.Sprintf("%d", j))
				require.NoError(b, os.WriteFile(objectPath, nil, 0o644))
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := LooseObjectsInfoForRepository(repo, time.Now())
			require.NoError(b, err)
		}
	})
}

func TestPackfileInfoForRepository(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	cfg := testcfg.Build(t)

	createRepo := func(t *testing.T) (*localrepo.Repo, string) {
		repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg, gittest.CreateRepositoryConfig{
			SkipCreationViaService: true,
		})
		return localrepo.NewTestRepo(t, cfg, repoProto), repoPath
	}

	requirePackfilesInfo := func(t *testing.T, repo *localrepo.Repo, expectedInfo PackfilesInfo) {
		info, err := PackfilesInfoForRepository(repo)
		require.NoError(t, err)
		require.Equal(t, expectedInfo, info)
	}

	t.Run("empty repository", func(t *testing.T) {
		repo, _ := createRepo(t)
		requirePackfilesInfo(t, repo, PackfilesInfo{})
	})

	t.Run("single packfile", func(t *testing.T) {
		repo, repoPath := createRepo(t)

		packfileDir := filepath.Join(repoPath, "objects", "pack")
		require.NoError(t, os.MkdirAll(packfileDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(packfileDir, "pack-foo.pack"), []byte("foobar"), 0o644))

		requirePackfilesInfo(t, repo, PackfilesInfo{
			Count: 1,
			Size:  6,
		})
	})

	t.Run("multiple packfiles", func(t *testing.T) {
		repo, repoPath := createRepo(t)

		packfileDir := filepath.Join(repoPath, "objects", "pack")
		require.NoError(t, os.MkdirAll(packfileDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(packfileDir, "pack-foo.pack"), []byte("foobar"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(packfileDir, "pack-bar.pack"), []byte("123"), 0o644))

		requirePackfilesInfo(t, repo, PackfilesInfo{
			Count: 2,
			Size:  9,
		})
	})

	t.Run("multiple packfiles with other data structures", func(t *testing.T) {
		repo, repoPath := createRepo(t)

		packfileDir := filepath.Join(repoPath, "objects", "pack")
		require.NoError(t, os.MkdirAll(packfileDir, 0o755))
		for _, file := range []string{
			"pack-bar.bar",
			"pack-bar.pack",
			"pack-bar.idx",
			"pack-foo.bar",
			"pack-foo.pack",
			"pack-foo.idx",
			"garbage",
		} {
			require.NoError(t, os.WriteFile(filepath.Join(packfileDir, file), []byte("1"), 0o644))
		}

		requirePackfilesInfo(t, repo, PackfilesInfo{
			Count:        2,
			Size:         2,
			GarbageCount: 1,
			GarbageSize:  1,
		})
	})
}
