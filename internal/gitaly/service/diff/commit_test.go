//go:build !gitaly_test_sha256

package diff

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/gittest"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/diff"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/storage"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCommitDiff(t *testing.T) {
	t.Parallel()

	ctx := testhelper.Context(t)
	cfg, client := setupDiffServiceWithoutRepo(t)

	type setupData struct {
		expectedErr  error
		request      *gitalypb.CommitDiffRequest
		expectedDiff []diff.Diff
	}

	for _, tc := range []struct {
		setup func() setupData
		desc  string
	}{
		{
			desc: "diff in single file",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random string of text\n"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("random of string text\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    []byte("@@ -1 +1 @@\n-random string of text\n+random of string text\n"),
						},
					},
				}
			},
		},
		{
			desc: "diff for file deleted",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random string of text\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath)

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     gittest.DefaultObjectHash.ZeroOID.String(),
							OldMode:  0o100644,
							NewMode:  0,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    []byte("@@ -1 +0,0 @@\n-random string of text\n"),
						},
					},
				}
			},
		},
		{
			desc: "file with multiple chunks",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, testhelper.MustReadFile(t, "testdata/file-with-multiple-chunks-before.txt"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, testhelper.MustReadFile(t, "testdata/file-with-multiple-chunks-after.txt"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    testhelper.MustReadFile(t, "testdata/file-with-multiple-chunks-diff.txt"),
						},
					},
				}
			},
		},
		{
			desc: "new file with pluses",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("+\n++\n+++\n++++\n+++++\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath)
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   gittest.DefaultObjectHash.ZeroOID.String(),
							ToID:     blob2.String(),
							OldMode:  0,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    []byte("@@ -0,0 +1,5 @@\n++\n+++\n++++\n+++++\n++++++\n"),
						},
					},
				}
			},
		},
		{
			desc: "no diff in file",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random string of text\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob1.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
						},
					},
				}
			},
		},
		{
			desc: "mode diff in single file",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random string of text\n"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("random of string text\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100755", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100755,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    []byte("@@ -1 +1 @@\n-random string of text\n+random of string text\n"),
						},
					},
				}
			},
		},
		{
			desc: "binary file",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				// Git detects binary files by looking for null characters
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("\x000 hello world"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("\x000 world hello"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100755", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100755,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Binary:   true,
							Patch:    []byte("Binary files a/foo and b/foo differ\n"),
						},
					},
				}
			},
		},
		{
			desc: "single file renamed with mode changes",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random string of text\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "goo", Mode: "100755", OID: blob1},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob1.String(),
							OldMode:  0o100644,
							NewMode:  0o100755,
							FromPath: []byte("foo"),
							ToPath:   []byte("goo"),
						},
					},
				}
			},
		},
		{
			desc: "no newline at the end",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random string of text"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("random of string text"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    []byte("@@ -1 +1 @@\n-random string of text\n\\ No newline at end of file\n+random of string text\n\\ No newline at end of file\n"),
						},
					},
				}
			},
		},
		{
			desc: "filename with tabs, newline",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random string of text\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath)
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo\tbar\ngoo.txt", Mode: "100644", OID: blob1},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   gittest.DefaultObjectHash.ZeroOID.String(),
							ToID:     blob1.String(),
							OldMode:  0,
							NewMode:  0o100644,
							FromPath: []byte("foo\tbar\ngoo.txt"),
							ToPath:   []byte("foo\tbar\ngoo.txt"),
							Patch:    []byte("@@ -0,0 +1 @@\n+random string of text\n"),
						},
					},
				}
			},
		},
		{
			desc: "filename with unicode",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("こんにちは世界\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath)
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "テスト.txt", Mode: "100644", OID: blob1},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   gittest.DefaultObjectHash.ZeroOID.String(),
							ToID:     blob1.String(),
							OldMode:  0,
							NewMode:  0o100644,
							FromPath: []byte("テスト.txt"),
							ToPath:   []byte("テスト.txt"),
							Patch:    []byte("@@ -0,0 +1 @@\n+こんにちは世界\n"),
						},
					},
				}
			},
		},
		{
			desc: "diff.noprefix set to true",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				gittest.Exec(t, cfg, "-C", repoPath, "config", "diff.noprefix", "true")

				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random string of text\n"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("random of string text\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:    repoProto,
						LeftCommitId:  commit1.String(),
						RightCommitId: commit2.String(),
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    []byte("@@ -1 +1 @@\n-random string of text\n+random of string text\n"),
						},
					},
				}
			},
		},
		{
			desc: "whitespace_changes: undefined + ignore_whitespace_change: false",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random text of string\n"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("\trandom text of string\n\n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:             repoProto,
						LeftCommitId:           commit1.String(),
						RightCommitId:          commit2.String(),
						WhitespaceChanges:      gitalypb.CommitDiffRequest_WHITESPACE_CHANGES_UNSPECIFIED,
						IgnoreWhitespaceChange: false,
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    []byte("@@ -1 +1,2 @@\n-random text of string\n+\trandom text of string\n+\n"),
						},
					},
				}
			},
		},
		{
			desc: "whitespace_changes: undefined + ignore_whitespace_change: true",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random text of string\n"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("random text of string \n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:             repoProto,
						LeftCommitId:           commit1.String(),
						RightCommitId:          commit2.String(),
						WhitespaceChanges:      gitalypb.CommitDiffRequest_WHITESPACE_CHANGES_UNSPECIFIED,
						IgnoreWhitespaceChange: true,
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
						},
					},
				}
			},
		},
		{
			desc: "whitespace_changes: dont_ignore",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random text of string\n"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("random text of string \n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:        repoProto,
						LeftCommitId:      commit1.String(),
						RightCommitId:     commit2.String(),
						WhitespaceChanges: gitalypb.CommitDiffRequest_WHITESPACE_CHANGES_UNSPECIFIED,
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    []byte("@@ -1 +1 @@\n-random text of string\n+random text of string \n"),
						},
					},
				}
			},
		},
		{
			desc: "whitespace_changes: ignore",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random text of string\n"))
				// prefix space is not ignored
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte(" random text of string \n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:             repoProto,
						LeftCommitId:           commit1.String(),
						RightCommitId:          commit2.String(),
						WhitespaceChanges:      gitalypb.CommitDiffRequest_WHITESPACE_CHANGES_IGNORE,
						IgnoreWhitespaceChange: true,
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
							Patch:    []byte("@@ -1 +1 @@\n-random text of string\n+ random text of string \n"),
						},
					},
				}
			},
		},
		{
			desc: "whitespace_changes: ignore_all",
			setup: func() setupData {
				repoProto, repoPath := gittest.CreateRepository(t, ctx, cfg)
				blob1 := gittest.WriteBlob(t, cfg, repoPath, []byte("random text of string\n"))
				blob2 := gittest.WriteBlob(t, cfg, repoPath, []byte("random text of string \n"))
				commit1 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob1},
				))
				commit2 := gittest.WriteCommit(t, cfg, repoPath, gittest.WithTreeEntries(
					gittest.TreeEntry{Path: "foo", Mode: "100644", OID: blob2},
				))

				return setupData{
					request: &gitalypb.CommitDiffRequest{
						Repository:             repoProto,
						LeftCommitId:           commit1.String(),
						RightCommitId:          commit2.String(),
						WhitespaceChanges:      gitalypb.CommitDiffRequest_WHITESPACE_CHANGES_IGNORE_ALL,
						IgnoreWhitespaceChange: true,
					},
					expectedDiff: []diff.Diff{
						{
							FromID:   blob1.String(),
							ToID:     blob2.String(),
							OldMode:  0o100644,
							NewMode:  0o100644,
							FromPath: []byte("foo"),
							ToPath:   []byte("foo"),
						},
					},
				}
			},
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			data := tc.setup()
			c, err := client.CommitDiff(ctx, data.request)
			require.Equal(t, err, data.expectedErr)
			if err != nil {
				return
			}

			assertExactReceivedDiffs(t, c, tc.setup().expectedDiff)
		})
	}
}

func TestSuccessfulCommitDiffRequestWithPaths(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	rightCommit := "e4003da16c1c2c3fc4567700121b17bf8e591c6c"
	leftCommit := "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"
	rpcRequest := &gitalypb.CommitDiffRequest{
		Repository:             repo,
		RightCommitId:          rightCommit,
		LeftCommitId:           leftCommit,
		IgnoreWhitespaceChange: false,
		Paths: [][]byte{
			[]byte("CONTRIBUTING.md"),
			[]byte("README.md"),
			[]byte("gitaly/named-file-with-mods"),
			[]byte("gitaly/mode-file-with-mods"),
		},
	}
	c, err := client.CommitDiff(ctx, rpcRequest)
	require.NoError(t, err)

	expectedDiffs := []diff.Diff{
		{
			FromID:   "c1788657b95998a2f177a4f86d68a60f2a80117f",
			ToID:     "b87f61fe2d7b2e208b340a1f3cafea916bd27f75",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("CONTRIBUTING.md"),
			ToPath:   []byte("CONTRIBUTING.md"),
			Binary:   false,
			Patch:    testhelper.MustReadFile(t, "testdata/contributing-md-chunks.txt"),
		},
		{
			FromID:   "faaf198af3a36dbf41961466703cc1d47c61d051",
			ToID:     "877cee6ab11f9094e1bcdb7f1fd9c0001b572185",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("README.md"),
			ToPath:   []byte("README.md"),
			Binary:   false,
			Patch:    testhelper.MustReadFile(t, "testdata/readme-md-chunks.txt"),
		},
		{
			FromID:   "357406f3075a57708d0163752905cc1576fceacc",
			ToID:     "8e5177d718c561d36efde08bad36b43687ee6bf0",
			OldMode:  0o100644,
			NewMode:  0o100755,
			FromPath: []byte("gitaly/mode-file-with-mods"),
			ToPath:   []byte("gitaly/mode-file-with-mods"),
			Binary:   false,
			Patch:    testhelper.MustReadFile(t, "testdata/mode-file-with-mods-chunks.txt"),
		},
		{
			FromID:   "43d24af4e22580f36b1ca52647c1aff75a766a33",
			ToID:     git.ObjectHashSHA1.ZeroOID.String(),
			OldMode:  0o100644,
			NewMode:  0,
			FromPath: []byte("gitaly/named-file-with-mods"),
			ToPath:   []byte("gitaly/named-file-with-mods"),
			Binary:   false,
			Patch:    testhelper.MustReadFile(t, "testdata/named-file-with-mods-chunks.txt"),
		},
	}

	assertExactReceivedDiffs(t, c, expectedDiffs)
}

func TestSuccessfulCommitDiffRequestWithTypeChangeDiff(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	rightCommit := "184a47d38677e2e439964859b877ae9bc424ab11"
	leftCommit := "80d56eb72ba5d77fd8af857eced17a7d0640cb82"
	rpcRequest := &gitalypb.CommitDiffRequest{
		Repository:    repo,
		RightCommitId: rightCommit,
		LeftCommitId:  leftCommit,
	}
	c, err := client.CommitDiff(ctx, rpcRequest)
	require.NoError(t, err)

	expectedDiffs := []diff.Diff{
		{
			FromID:   "349cd0f6b1aba8538861d95783cbce6d49d747f8",
			ToID:     git.ObjectHashSHA1.ZeroOID.String(),
			OldMode:  0o120000,
			NewMode:  0,
			FromPath: []byte("gitaly/symlink-to-be-regular"),
			ToPath:   []byte("gitaly/symlink-to-be-regular"),
			Binary:   false,
			Patch:    testhelper.MustReadFile(t, "testdata/symlink-to-be-regular-deleted-chunks.txt"),
		},
		{
			FromID:   git.ObjectHashSHA1.ZeroOID.String(),
			ToID:     "f9e5cc857610185e6feeb494a26bf27551a4f02b",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/symlink-to-be-regular"),
			ToPath:   []byte("gitaly/symlink-to-be-regular"),
			Binary:   false,
			Patch:    testhelper.MustReadFile(t, "testdata/symlink-to-be-regular-added-chunks.txt"),
		},
	}

	assertExactReceivedDiffs(t, c, expectedDiffs)
}

func TestSuccessfulCommitDiffRequestWithIgnoreWhitespaceChange(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	rightCommit := "e4003da16c1c2c3fc4567700121b17bf8e591c6c"
	leftCommit := "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"

	whitespacePaths := [][]byte{
		[]byte("CONTRIBUTING.md"),
		[]byte("MAINTENANCE.md"),
		[]byte("README.md"),
	}
	normalPaths := [][]byte{
		[]byte("gitaly/named-file-with-mods"),
		[]byte("gitaly/mode-file-with-mods"),
	}

	expectedWhitespaceDiffs := []diff.Diff{
		{
			FromID:   "c1788657b95998a2f177a4f86d68a60f2a80117f",
			ToID:     "b87f61fe2d7b2e208b340a1f3cafea916bd27f75",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("CONTRIBUTING.md"),
			ToPath:   []byte("CONTRIBUTING.md"),
			Binary:   false,
		},
		{
			FromID:   "95d9f0a5e7bb054e9dd3975589b8dfc689e20e88",
			ToID:     "5d9c7c0470bf368d61d9b6cd076300dc9d061f14",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("MAINTENANCE.md"),
			ToPath:   []byte("MAINTENANCE.md"),
			Binary:   false,
		},
		{
			FromID:   "faaf198af3a36dbf41961466703cc1d47c61d051",
			ToID:     "877cee6ab11f9094e1bcdb7f1fd9c0001b572185",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("README.md"),
			ToPath:   []byte("README.md"),
			Binary:   false,
		},
	}
	expectedNormalDiffs := []diff.Diff{
		{
			FromID:   "357406f3075a57708d0163752905cc1576fceacc",
			ToID:     "8e5177d718c561d36efde08bad36b43687ee6bf0",
			OldMode:  0o100644,
			NewMode:  0o100755,
			FromPath: []byte("gitaly/mode-file-with-mods"),
			ToPath:   []byte("gitaly/mode-file-with-mods"),
			Binary:   false,
			Patch:    testhelper.MustReadFile(t, "testdata/mode-file-with-mods-chunks.txt"),
		},
		{
			FromID:   "43d24af4e22580f36b1ca52647c1aff75a766a33",
			ToID:     git.ObjectHashSHA1.ZeroOID.String(),
			OldMode:  0o100644,
			NewMode:  0,
			FromPath: []byte("gitaly/named-file-with-mods"),
			ToPath:   []byte("gitaly/named-file-with-mods"),
			Binary:   false,
			Patch:    testhelper.MustReadFile(t, "testdata/named-file-with-mods-chunks.txt"),
		},
	}

	pathsAndDiffs := []struct {
		desc  string
		paths [][]byte
		diffs []diff.Diff
	}{
		{
			desc:  "whitespace paths",
			paths: whitespacePaths,
			diffs: expectedWhitespaceDiffs,
		},
		{
			desc:  "whitespace paths and normal paths",
			paths: append(whitespacePaths, normalPaths...),
			diffs: append(expectedWhitespaceDiffs, expectedNormalDiffs...),
		},
	}

	for _, entry := range pathsAndDiffs {
		t.Run(entry.desc, func(t *testing.T) {
			rpcRequest := &gitalypb.CommitDiffRequest{
				Repository:             repo,
				RightCommitId:          rightCommit,
				LeftCommitId:           leftCommit,
				IgnoreWhitespaceChange: true,
				Paths:                  entry.paths,
			}
			c, err := client.CommitDiff(ctx, rpcRequest)
			require.NoError(t, err)

			assertExactReceivedDiffs(t, c, entry.diffs)
		})
	}
}

func TestSuccessfulCommitDiffRequestWithWordDiff(t *testing.T) {
	ctx := testhelper.Context(t)
	cfg, repo, repoPath, client := setupDiffService(t, ctx)

	rightCommit := "ab2c9622c02288a2bbaaf35d96088cfdff31d9d9"
	leftCommit := "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"

	var diffPatches [][]byte
	output := gittest.Exec(t, cfg, "-C", repoPath, "diff", "--word-diff=porcelain", leftCommit, rightCommit)
	diffPerFile := bytes.Split(output, []byte("diff --git"))

	for _, s := range diffPerFile {
		if idx := bytes.Index(s, []byte("@@")); idx != -1 {
			diffPatches = append(diffPatches, s[idx:])
		}
	}

	expectedDiffs := []diff.Diff{
		{
			FromID:   "faaf198af3a36dbf41961466703cc1d47c61d051",
			ToID:     "877cee6ab11f9094e1bcdb7f1fd9c0001b572185",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("README.md"),
			ToPath:   []byte("README.md"),
			Binary:   false,
			Patch:    diffPatches[0],
		},
		{
			FromID:   "bdea48ee65c869eb0b86b1283069d76cce0a7254",
			ToID:     "0000000000000000000000000000000000000000",
			OldMode:  0o100644,
			NewMode:  0,
			FromPath: []byte("gitaly/deleted-file"),
			ToPath:   []byte("gitaly/deleted-file"),
			Binary:   false,
			Patch:    diffPatches[1],
		},
		{
			FromID:   "aa408b4556e594f7974390ad6b86210617fbda6e",
			ToID:     "1c69c4d2a65ad05c24ac3b6780b5748b97ffd3aa",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/file-with-multiple-chunks"),
			ToPath:   []byte("gitaly/file-with-multiple-chunks"),
			Binary:   false,
			Patch:    diffPatches[2],
		},
		{
			FromID:   "0000000000000000000000000000000000000000",
			ToID:     "389c7a36a6e133268b0d36b00e7ffc0f3a5b6651",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/file-with-pluses.txt"),
			ToPath:   []byte("gitaly/file-with-pluses.txt"),
			Binary:   false,
			Patch:    diffPatches[3],
		},
		{
			FromID:   "0000000000000000000000000000000000000000",
			ToID:     "bc2ef601a538d69ef99d5bdafa605e63f902e8e4",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/logo-white.png"),
			ToPath:   []byte("gitaly/logo-white.png"),
			Binary:   true,
			Patch:    []byte("Binary files /dev/null and b/gitaly/logo-white.png differ\n"),
		},
		{
			FromID:   "ead5a0eee1391308803cfebd8a2a8530495645eb",
			ToID:     "ead5a0eee1391308803cfebd8a2a8530495645eb",
			OldMode:  0o100644,
			NewMode:  0o100755,
			FromPath: []byte("gitaly/mode-file"),
			ToPath:   []byte("gitaly/mode-file"),
			Binary:   false,
		},
		{
			FromID:   "357406f3075a57708d0163752905cc1576fceacc",
			ToID:     "8e5177d718c561d36efde08bad36b43687ee6bf0",
			OldMode:  0o100644,
			NewMode:  0o100755,
			FromPath: []byte("gitaly/mode-file-with-mods"),
			ToPath:   []byte("gitaly/mode-file-with-mods"),
			Binary:   false,
			Patch:    diffPatches[4],
		},
		{
			FromID:   "43d24af4e22580f36b1ca52647c1aff75a766a33",
			ToID:     "0000000000000000000000000000000000000000",
			OldMode:  0o100644,
			NewMode:  0,
			FromPath: []byte("gitaly/named-file-with-mods"),
			ToPath:   []byte("gitaly/named-file-with-mods"),
			Binary:   false,
			Patch:    diffPatches[5],
		},
		{
			FromID:   "0000000000000000000000000000000000000000",
			ToID:     "b464dff7a75ccc92fbd920fd9ae66a84b9d2bf94",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/no-newline-at-the-end"),
			ToPath:   []byte("gitaly/no-newline-at-the-end"),
			Binary:   false,
			Patch:    diffPatches[6],
		},
		{
			FromID:   "4e76e90b3c7e52390de9311a23c0a77575aed8a8",
			ToID:     "4e76e90b3c7e52390de9311a23c0a77575aed8a8",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/named-file"),
			ToPath:   []byte("gitaly/renamed-file"),
			Binary:   false,
		},
		{
			FromID:   "0000000000000000000000000000000000000000",
			ToID:     "3856c00e9450a51a62096327167fc43d3be62eef",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/renamed-file-with-mods"),
			ToPath:   []byte("gitaly/renamed-file-with-mods"),
			Binary:   false,
			Patch:    diffPatches[7],
		},
		{
			FromID:   "0000000000000000000000000000000000000000",
			ToID:     "a135e3e0d4af177a902ca57dcc4c7fc6f30858b1",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/tab\tnewline\n file"),
			ToPath:   []byte("gitaly/tab\tnewline\n file"),
			Binary:   false,
			Patch:    diffPatches[8],
		},
		{
			FromID:   "0000000000000000000000000000000000000000",
			ToID:     "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
			OldMode:  0,
			NewMode:  0o100755,
			FromPath: []byte("gitaly/テスト.txt"),
			ToPath:   []byte("gitaly/テスト.txt"),
			Binary:   false,
		},
		{
			FromID:   "0000000000000000000000000000000000000000",
			ToID:     "b1e67221afe8461efd244b487afca22d46b95eb8",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("z-short-diff"),
			ToPath:   []byte("z-short-diff"),
			Binary:   false,
			Patch:    diffPatches[9],
		},
	}

	testCases := []struct {
		noPrefixConfig string
		desc           string
	}{
		{noPrefixConfig: "false", desc: "Git config diff.noprefix set to false"},
		{noPrefixConfig: "true", desc: "Git config diff.noprefix set to true"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			gittest.Exec(t, cfg, "-C", repoPath, "config", "diff.noprefix", testCase.noPrefixConfig)
			rpcRequest := &gitalypb.CommitDiffRequest{
				Repository:             repo,
				RightCommitId:          rightCommit,
				LeftCommitId:           leftCommit,
				IgnoreWhitespaceChange: false,
				DiffMode:               gitalypb.CommitDiffRequest_WORDDIFF,
			}
			c, err := client.CommitDiff(ctx, rpcRequest)
			if err != nil {
				t.Fatal(err)
			}

			assertExactReceivedDiffs(t, c, expectedDiffs)
		})
	}
}

func TestSuccessfulCommitDiffRequestWithLimits(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	rightCommit := "899d3d27b04690ac1cd9ef4d8a74fde0667c57f1"
	leftCommit := "184a47d38677e2e439964859b877ae9bc424ab11"

	type diffAttributes struct {
		path                                string
		collapsed, overflowMarker, tooLarge bool
	}

	requestsAndResults := []struct {
		desc    string
		request *gitalypb.CommitDiffRequest
		result  []diffAttributes
	}{
		{
			desc: "no enforcement",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits: false,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md"},
				{path: "LICENSE"},
				{path: "PROCESS.md"},
				{path: "VERSION"},
			},
		},
		{
			desc: "max file count enforcement",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits: true,
				MaxFiles:      3,
				MaxLines:      1000,
				MaxBytes:      3 * 5 * 1024,
				MaxPatchBytes: 100000,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md"},
				{path: "LICENSE"},
				{overflowMarker: true},
			},
		},
		{
			desc: "max file count enforcement with collect all paths",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits:   true,
				MaxFiles:        3,
				MaxLines:        1000,
				MaxBytes:        3 * 5 * 1024,
				MaxPatchBytes:   100000,
				CollectAllPaths: true,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md"},
				{path: "LICENSE"},
				{path: "PROCESS.md", overflowMarker: true},
				{path: "VERSION", overflowMarker: true},
			},
		},
		{
			desc: "max line count enforcement",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits: true,
				MaxFiles:      5,
				MaxLines:      90,
				MaxBytes:      5 * 5 * 1024,
				MaxPatchBytes: 100000,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md"},
				{overflowMarker: true},
			},
		},
		{
			desc: "max line count enforcement with collect all paths",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits:   true,
				MaxFiles:        5,
				MaxLines:        90,
				MaxBytes:        5 * 5 * 1024,
				MaxPatchBytes:   100000,
				CollectAllPaths: true,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md"},
				{path: "LICENSE", overflowMarker: true},
				{path: "PROCESS.md", overflowMarker: true},
				{path: "VERSION", overflowMarker: true},
			},
		},
		{
			desc: "max byte count enforcement",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits: true,
				MaxFiles:      5,
				MaxLines:      1000,
				MaxBytes:      6900,
				MaxPatchBytes: 100000,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md"},
				{path: "LICENSE"},
				{path: "PROCESS.md"},
				{overflowMarker: true},
			},
		},
		{
			desc: "max byte count enforcement with collect all paths",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits:   true,
				MaxFiles:        5,
				MaxLines:        1000,
				MaxBytes:        6900,
				MaxPatchBytes:   100000,
				CollectAllPaths: true,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md"},
				{path: "LICENSE"},
				{path: "PROCESS.md"},
				{path: "VERSION", overflowMarker: true},
			},
		},
		{
			desc: "no collapse",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits: true,
				CollapseDiffs: false,
				MaxFiles:      3,
				MaxLines:      1000,
				MaxBytes:      3 * 5 * 1024,
				SafeMaxFiles:  1,
				SafeMaxLines:  1000,
				SafeMaxBytes:  1 * 5 * 1024,
				MaxPatchBytes: 100000,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md"},
				{path: "LICENSE"},
				{overflowMarker: true},
			},
		},
		{
			desc: "set as too large when exceeding single patch limit",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits: true,
				CollapseDiffs: false,
				MaxFiles:      5,
				MaxLines:      1000,
				MaxBytes:      3 * 5 * 1024,
				SafeMaxFiles:  3,
				SafeMaxLines:  1000,
				SafeMaxBytes:  1 * 5 * 1024,
				MaxPatchBytes: 1200,
			},
			result: []diffAttributes{
				{path: "CHANGELOG", tooLarge: true},
				{path: "CONTRIBUTING.md", tooLarge: true},
				{path: "LICENSE", tooLarge: false},
				{path: "PROCESS.md", tooLarge: true},
				{path: "VERSION", tooLarge: false},
			},
		},
		{
			desc: "max patch bytes for file type",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits:                 true,
				CollapseDiffs:                 false,
				MaxFiles:                      5,
				MaxLines:                      1000,
				MaxBytes:                      3 * 5 * 1024,
				SafeMaxFiles:                  3,
				SafeMaxLines:                  1000,
				SafeMaxBytes:                  1 * 5 * 1024,
				MaxPatchBytes:                 1200,
				MaxPatchBytesForFileExtension: map[string]int32{".md": 10000},
			},
			result: []diffAttributes{
				{path: "CHANGELOG", tooLarge: true},
				{path: "CONTRIBUTING.md", tooLarge: false},
				{path: "LICENSE", tooLarge: false},
				{path: "PROCESS.md", tooLarge: false},
				{path: "VERSION", tooLarge: false},
			},
		},
		{
			desc: "collapse after safe max file count is exceeded",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits: true,
				CollapseDiffs: true,
				MaxFiles:      3,
				MaxLines:      1000,
				MaxBytes:      3 * 5 * 1024,
				SafeMaxFiles:  1,
				SafeMaxLines:  1000,
				SafeMaxBytes:  1 * 5 * 1024,
				MaxPatchBytes: 100000,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md", collapsed: true},
				{path: "LICENSE", collapsed: true},
				{overflowMarker: true},
			},
		},
		{
			desc: "collapse after safe max line count is exceeded",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits: true,
				CollapseDiffs: true,
				MaxFiles:      5,
				MaxLines:      100,
				MaxBytes:      5 * 5 * 1024,
				SafeMaxFiles:  5,
				SafeMaxLines:  40,
				SafeMaxBytes:  5 * 5 * 1024,
				MaxPatchBytes: 100000,
			},
			result: []diffAttributes{
				{path: "CHANGELOG", collapsed: true},
				{path: "CONTRIBUTING.md", collapsed: true},
				{path: "LICENSE"},
				{path: "PROCESS.md", collapsed: true},
				{path: "VERSION"},
			},
		},
		{
			desc: "collapse after safe max byte count is exceeded",
			request: &gitalypb.CommitDiffRequest{
				EnforceLimits: true,
				CollapseDiffs: true,
				MaxFiles:      4,
				MaxLines:      1000,
				MaxBytes:      4 * 5 * 1024,
				SafeMaxFiles:  4,
				SafeMaxLines:  1000,
				SafeMaxBytes:  4830,
				MaxPatchBytes: 100000,
			},
			result: []diffAttributes{
				{path: "CHANGELOG"},
				{path: "CONTRIBUTING.md"},
				{path: "LICENSE"},
				{path: "PROCESS.md", collapsed: true},
				{overflowMarker: true},
			},
		},
	}

	for _, requestAndResult := range requestsAndResults {
		t.Run(requestAndResult.desc, func(t *testing.T) {
			request := requestAndResult.request
			request.Repository = repo
			request.LeftCommitId = leftCommit
			request.RightCommitId = rightCommit
			c, err := client.CommitDiff(ctx, request)
			require.NoError(t, err)

			receivedDiffs := getDiffsFromCommitDiffClient(t, c)

			require.Equal(t, len(requestAndResult.result), len(receivedDiffs), "number of diffs received")
			for i, diff := range receivedDiffs {
				expectedDiff := requestAndResult.result[i]

				require.Equal(t, expectedDiff.overflowMarker, diff.OverflowMarker, "%s overflow marker", diff.FromPath)
				require.Equal(t, expectedDiff.tooLarge, diff.TooLarge, "%s too large", diff.FromPath)
				require.Equal(t, expectedDiff.path, string(diff.FromPath), "%s path", diff.FromPath)
				require.Equal(t, expectedDiff.collapsed, diff.Collapsed, "%s collapsed", diff.FromPath)

				if expectedDiff.collapsed {
					require.Empty(t, diff.Patch, "patch")
				}
			}
		})
	}
}

func TestFailedCommitDiffRequestDueToValidationError(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	const rightCommit = "d42783470dc29fde2cf459eb3199ee1d7e3f3a72"
	const leftCommit = rightCommit + "~" // Parent of rightCommit

	for _, tc := range []struct {
		desc   string
		req    *gitalypb.CommitDiffRequest
		exrErr error
	}{
		{
			desc: "Repository doesn't exist",
			req:  &gitalypb.CommitDiffRequest{Repository: &gitalypb.Repository{StorageName: "fake", RelativePath: "path"}, RightCommitId: rightCommit, LeftCommitId: leftCommit},
			exrErr: testhelper.GitalyOrPraefect(
				testhelper.ToInterceptedMetadata(structerr.NewInvalidArgument(
					"%w", storage.NewStorageNotFoundError("fake"),
				)),
				testhelper.ToInterceptedMetadata(structerr.NewInvalidArgument(
					"repo scoped: %w", storage.NewStorageNotFoundError("fake"),
				)),
			),
		},
		{
			desc: "Repository is nil",
			req:  &gitalypb.CommitDiffRequest{Repository: nil, RightCommitId: rightCommit, LeftCommitId: leftCommit},
			exrErr: testhelper.GitalyOrPraefect(
				structerr.NewInvalidArgument("%w", storage.ErrRepositoryNotSet),
				structerr.NewInvalidArgument("repo scoped: %w", storage.ErrRepositoryNotSet),
			),
		},
		{
			desc:   "RightCommitId is empty",
			req:    &gitalypb.CommitDiffRequest{Repository: repo, RightCommitId: "", LeftCommitId: leftCommit},
			exrErr: status.Error(codes.InvalidArgument, "empty RightCommitId"),
		},
		{
			desc:   "LeftCommitId is empty",
			req:    &gitalypb.CommitDiffRequest{Repository: repo, RightCommitId: rightCommit, LeftCommitId: ""},
			exrErr: status.Error(codes.InvalidArgument, "empty LeftCommitId"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := client.CommitDiff(ctx, tc.req)
			require.NoError(t, err)

			err = drainCommitDiffResponse(c)
			testhelper.RequireGrpcError(t, tc.exrErr, err)
		})
	}
}

func TestFailedCommitDiffRequestWithNonExistentCommit(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	nonExistentCommitID := "deadfacedeadfacedeadfacedeadfacedeadface"
	leftCommit := nonExistentCommitID + "~" // Parent of rightCommit
	rpcRequest := &gitalypb.CommitDiffRequest{Repository: repo, RightCommitId: nonExistentCommitID, LeftCommitId: leftCommit}
	c, err := client.CommitDiff(ctx, rpcRequest)
	require.NoError(t, err)

	err = drainCommitDiffResponse(c)
	testhelper.RequireGrpcCode(t, err, codes.Unavailable)
}

func TestSuccessfulCommitDeltaRequest(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	rightCommit := "742518b2be68fc750bb4c357c0df821a88113286"
	leftCommit := "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"
	rpcRequest := &gitalypb.CommitDeltaRequest{Repository: repo, RightCommitId: rightCommit, LeftCommitId: leftCommit}
	c, err := client.CommitDelta(ctx, rpcRequest)
	require.NoError(t, err)

	expectedDeltas := []diff.Diff{
		{
			FromID:   "faaf198af3a36dbf41961466703cc1d47c61d051",
			ToID:     "877cee6ab11f9094e1bcdb7f1fd9c0001b572185",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("README.md"),
			ToPath:   []byte("README.md"),
		},
		{
			FromID:   "bdea48ee65c869eb0b86b1283069d76cce0a7254",
			ToID:     git.ObjectHashSHA1.ZeroOID.String(),
			OldMode:  0o100644,
			NewMode:  0,
			FromPath: []byte("gitaly/deleted-file"),
			ToPath:   []byte("gitaly/deleted-file"),
		},
		{
			FromID:   "aa408b4556e594f7974390ad6b86210617fbda6e",
			ToID:     "1c69c4d2a65ad05c24ac3b6780b5748b97ffd3aa",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/file-with-multiple-chunks"),
			ToPath:   []byte("gitaly/file-with-multiple-chunks"),
		},
		{
			FromID:   git.ObjectHashSHA1.ZeroOID.String(),
			ToID:     "bc2ef601a538d69ef99d5bdafa605e63f902e8e4",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/logo-white.png"),
			ToPath:   []byte("gitaly/logo-white.png"),
		},
		{
			FromID:   "ead5a0eee1391308803cfebd8a2a8530495645eb",
			ToID:     "ead5a0eee1391308803cfebd8a2a8530495645eb",
			OldMode:  0o100644,
			NewMode:  0o100755,
			FromPath: []byte("gitaly/mode-file"),
			ToPath:   []byte("gitaly/mode-file"),
		},
		{
			FromID:   "357406f3075a57708d0163752905cc1576fceacc",
			ToID:     "8e5177d718c561d36efde08bad36b43687ee6bf0",
			OldMode:  0o100644,
			NewMode:  0o100755,
			FromPath: []byte("gitaly/mode-file-with-mods"),
			ToPath:   []byte("gitaly/mode-file-with-mods"),
		},
		{
			FromID:   "43d24af4e22580f36b1ca52647c1aff75a766a33",
			ToID:     git.ObjectHashSHA1.ZeroOID.String(),
			OldMode:  0o100644,
			NewMode:  0,
			FromPath: []byte("gitaly/named-file-with-mods"),
			ToPath:   []byte("gitaly/named-file-with-mods"),
		},
		{
			FromID:   git.ObjectHashSHA1.ZeroOID.String(),
			ToID:     "b464dff7a75ccc92fbd920fd9ae66a84b9d2bf94",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/no-newline-at-the-end"),
			ToPath:   []byte("gitaly/no-newline-at-the-end"),
		},
		{
			FromID:   "4e76e90b3c7e52390de9311a23c0a77575aed8a8",
			ToID:     "4e76e90b3c7e52390de9311a23c0a77575aed8a8",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/named-file"),
			ToPath:   []byte("gitaly/renamed-file"),
		},
		{
			FromID:   git.ObjectHashSHA1.ZeroOID.String(),
			ToID:     "3856c00e9450a51a62096327167fc43d3be62eef",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/renamed-file-with-mods"),
			ToPath:   []byte("gitaly/renamed-file-with-mods"),
		},
		{
			FromID:   git.ObjectHashSHA1.ZeroOID.String(),
			ToID:     "a135e3e0d4af177a902ca57dcc4c7fc6f30858b1",
			OldMode:  0,
			NewMode:  0o100644,
			FromPath: []byte("gitaly/tab\tnewline\n file"),
			ToPath:   []byte("gitaly/tab\tnewline\n file"),
		},
		{
			FromID:   git.ObjectHashSHA1.ZeroOID.String(),
			ToID:     "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
			OldMode:  0,
			NewMode:  0o100755,
			FromPath: []byte("gitaly/テスト.txt"),
			ToPath:   []byte("gitaly/テスト.txt"),
		},
	}

	assertExactReceivedDeltas(t, c, expectedDeltas)
}

func TestSuccessfulCommitDeltaRequestWithPaths(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	rightCommit := "e4003da16c1c2c3fc4567700121b17bf8e591c6c"
	leftCommit := "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"
	rpcRequest := &gitalypb.CommitDeltaRequest{
		Repository:    repo,
		RightCommitId: rightCommit,
		LeftCommitId:  leftCommit,
		Paths: [][]byte{
			[]byte("CONTRIBUTING.md"),
			[]byte("README.md"),
			[]byte("gitaly/named-file-with-mods"),
			[]byte("gitaly/mode-file-with-mods"),
		},
	}
	c, err := client.CommitDelta(ctx, rpcRequest)
	require.NoError(t, err)

	expectedDeltas := []diff.Diff{
		{
			FromID:   "c1788657b95998a2f177a4f86d68a60f2a80117f",
			ToID:     "b87f61fe2d7b2e208b340a1f3cafea916bd27f75",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("CONTRIBUTING.md"),
			ToPath:   []byte("CONTRIBUTING.md"),
		},
		{
			FromID:   "faaf198af3a36dbf41961466703cc1d47c61d051",
			ToID:     "877cee6ab11f9094e1bcdb7f1fd9c0001b572185",
			OldMode:  0o100644,
			NewMode:  0o100644,
			FromPath: []byte("README.md"),
			ToPath:   []byte("README.md"),
		},
		{
			FromID:   "357406f3075a57708d0163752905cc1576fceacc",
			ToID:     "8e5177d718c561d36efde08bad36b43687ee6bf0",
			OldMode:  0o100644,
			NewMode:  0o100755,
			FromPath: []byte("gitaly/mode-file-with-mods"),
			ToPath:   []byte("gitaly/mode-file-with-mods"),
		},
		{
			FromID:   "43d24af4e22580f36b1ca52647c1aff75a766a33",
			ToID:     git.ObjectHashSHA1.ZeroOID.String(),
			OldMode:  0o100644,
			NewMode:  0,
			FromPath: []byte("gitaly/named-file-with-mods"),
			ToPath:   []byte("gitaly/named-file-with-mods"),
		},
	}

	assertExactReceivedDeltas(t, c, expectedDeltas)
}

func TestFailedCommitDeltaRequestDueToValidationError(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	rightCommit := "d42783470dc29fde2cf459eb3199ee1d7e3f3a72"
	leftCommit := rightCommit + "~" // Parent of rightCommit

	rpcRequests := []*gitalypb.CommitDeltaRequest{
		{Repository: &gitalypb.Repository{StorageName: "fake", RelativePath: "path"}, RightCommitId: rightCommit, LeftCommitId: leftCommit}, // Repository doesn't exist
		{Repository: nil, RightCommitId: rightCommit, LeftCommitId: leftCommit},                                                             // Repository is nil
		{Repository: repo, RightCommitId: "", LeftCommitId: leftCommit},                                                                     // RightCommitId is empty
		{Repository: repo, RightCommitId: rightCommit, LeftCommitId: ""},                                                                    // LeftCommitId is empty
	}

	for _, rpcRequest := range rpcRequests {
		t.Run(fmt.Sprintf("%v", rpcRequest), func(t *testing.T) {
			c, err := client.CommitDelta(ctx, rpcRequest)
			require.NoError(t, err)

			err = drainCommitDeltaResponse(c)
			testhelper.RequireGrpcCode(t, err, codes.InvalidArgument)
		})
	}
}

func TestFailedCommitDeltaRequestWithNonExistentCommit(t *testing.T) {
	ctx := testhelper.Context(t)
	_, repo, _, client := setupDiffService(t, ctx)

	nonExistentCommitID := "deadfacedeadfacedeadfacedeadfacedeadface"
	leftCommit := nonExistentCommitID + "~" // Parent of rightCommit
	rpcRequest := &gitalypb.CommitDeltaRequest{Repository: repo, RightCommitId: nonExistentCommitID, LeftCommitId: leftCommit}
	c, err := client.CommitDelta(ctx, rpcRequest)
	require.NoError(t, err)

	err = drainCommitDeltaResponse(c)
	testhelper.RequireGrpcCode(t, err, codes.Unavailable)
}

func drainCommitDiffResponse(c gitalypb.DiffService_CommitDiffClient) error {
	for {
		_, err := c.Recv()
		if err != nil {
			return err
		}
	}
}

func drainCommitDeltaResponse(c gitalypb.DiffService_CommitDeltaClient) error {
	for {
		_, err := c.Recv()
		if err != nil {
			return err
		}
	}
}

func getDiffsFromCommitDiffClient(t *testing.T, client gitalypb.DiffService_CommitDiffClient) []*diff.Diff {
	var diffs []*diff.Diff
	var currentDiff *diff.Diff

	for {
		fetchedDiff, err := client.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		if currentDiff == nil {
			currentDiff = &diff.Diff{
				FromID:         fetchedDiff.FromId,
				ToID:           fetchedDiff.ToId,
				OldMode:        fetchedDiff.OldMode,
				NewMode:        fetchedDiff.NewMode,
				FromPath:       fetchedDiff.FromPath,
				ToPath:         fetchedDiff.ToPath,
				Binary:         fetchedDiff.Binary,
				Collapsed:      fetchedDiff.Collapsed,
				OverflowMarker: fetchedDiff.OverflowMarker,
				Patch:          fetchedDiff.RawPatchData,
				TooLarge:       fetchedDiff.TooLarge,
			}
		} else {
			currentDiff.Patch = append(currentDiff.Patch, fetchedDiff.RawPatchData...)
		}

		if fetchedDiff.EndOfPatch {
			diffs = append(diffs, currentDiff)
			currentDiff = nil
		}
	}

	return diffs
}

func assertExactReceivedDiffs(t *testing.T, client gitalypb.DiffService_CommitDiffClient, expectedDiffs []diff.Diff) {
	fetchedDiffs := getDiffsFromCommitDiffClient(t, client)

	var i int
	var fetchedDiff *diff.Diff

	for i, fetchedDiff = range fetchedDiffs {
		require.Greater(t, len(expectedDiffs), i, "Unexpected diff #%d received: %v", i, fetchedDiff)

		expectedDiff := expectedDiffs[i]
		require.Equal(t, expectedDiff.FromID, fetchedDiff.FromID, "FromID should match")
		require.Equal(t, expectedDiff.ToID, fetchedDiff.ToID, "ToID should match")
		require.Equal(t, expectedDiff.OldMode, fetchedDiff.OldMode, "OldMode should match")
		require.Equal(t, expectedDiff.NewMode, fetchedDiff.NewMode, "NewMode should match")
		require.Equal(t, expectedDiff.FromPath, fetchedDiff.FromPath, "FromPath should match")
		require.Equal(t, expectedDiff.ToPath, fetchedDiff.ToPath, "ToPath should match")
		require.Equal(t, expectedDiff.Binary, fetchedDiff.Binary, "Binary should match")
		require.Equal(t, expectedDiff.Patch, fetchedDiff.Patch, "Patch should match")
	}

	require.Len(t, expectedDiffs, i+1, "Unexpected number of diffs")
}

func assertExactReceivedDeltas(t *testing.T, client gitalypb.DiffService_CommitDeltaClient, expectedDeltas []diff.Diff) {
	t.Helper()

	counter := 0
	for {
		fetchedDeltas, err := client.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		for _, fetchedDelta := range fetchedDeltas.GetDeltas() {
			require.GreaterOrEqual(t, len(expectedDeltas), counter, "Unexpected delta #%d received: %v", counter, fetchedDelta)

			expectedDelta := expectedDeltas[counter]

			require.Equal(t, expectedDelta.FromID, fetchedDelta.FromId)
			require.Equal(t, expectedDelta.ToID, fetchedDelta.ToId)
			require.Equal(t, expectedDelta.OldMode, fetchedDelta.OldMode)
			require.Equal(t, expectedDelta.NewMode, fetchedDelta.NewMode)
			require.Equal(t, expectedDelta.FromPath, fetchedDelta.FromPath)
			require.Equal(t, expectedDelta.ToPath, fetchedDelta.ToPath)

			counter++
		}
	}

	require.Len(t, expectedDeltas, counter, "Unexpected number of deltas")
}
