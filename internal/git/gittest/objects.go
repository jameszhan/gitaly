package gittest

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v16/internal/helper/text"
)

// ObjectHashIsSHA256 returns if the current default object hash is SHA256.
func ObjectHashIsSHA256() bool {
	return DefaultObjectHash.EmptyTreeOID == git.ObjectHashSHA256.EmptyTreeOID
}

// ObjectHashDependent returns the value from the given map that is associated with the default
// object hash (e.g. "sha1", "sha256"). Fails in case the map doesn't contain the current object
// hash.
func ObjectHashDependent[T any](tb testing.TB, valuesByObjectHash map[string]T) T {
	tb.Helper()
	require.Contains(tb, valuesByObjectHash, DefaultObjectHash.Format)
	return valuesByObjectHash[DefaultObjectHash.Format]
}

// ListObjects returns a list of all object IDs in the repository.
func ListObjects(tb testing.TB, cfg config.Cfg, repoPath string) []git.ObjectID {
	tb.Helper()

	rawOutput := bytes.Split(
		bytes.TrimSpace(
			Exec(tb, cfg, "-C", repoPath, "cat-file", "--batch-check=%(objectname)", "--batch-all-objects"),
		),
		[]byte{'\n'},
	)

	objects := []git.ObjectID{}
	if len(rawOutput[0]) > 0 {
		for _, oid := range rawOutput {
			objects = append(objects, git.ObjectID(oid))
		}
	}

	return objects
}

// ObjectSize returns the size of the object identified by the given ID.
func ObjectSize(tb testing.TB, cfg config.Cfg, repoPath string, objectID git.ObjectID) int64 {
	output := Exec(tb, cfg, "-C", repoPath, "cat-file", "-s", objectID.String())
	size, err := strconv.ParseInt(text.ChompBytes(output), 10, 64)
	require.NoError(tb, err)
	return size
}

// RequireObjectExists asserts that the given repository does contain an object with the specified
// object ID.
func RequireObjectExists(tb testing.TB, cfg config.Cfg, repoPath string, objectID git.ObjectID) {
	requireObjectExists(tb, cfg, repoPath, objectID, true)
}

// RequireObjectNotExists asserts that the given repository does not contain an object with the
// specified object ID.
func RequireObjectNotExists(tb testing.TB, cfg config.Cfg, repoPath string, objectID git.ObjectID) {
	requireObjectExists(tb, cfg, repoPath, objectID, false)
}

func requireObjectExists(tb testing.TB, cfg config.Cfg, repoPath string, objectID git.ObjectID, exists bool) {
	cmd := NewCommand(tb, cfg, "-C", repoPath, "cat-file", "-e", objectID.String())
	cmd.Env = []string{
		"GIT_ALLOW_PROTOCOL=", // To prevent partial clone reaching remote repo over SSH
	}

	if exists {
		require.NoError(tb, cmd.Run(), "checking for object should succeed")
		return
	}
	require.Error(tb, cmd.Run(), "checking for object should fail")
}

// GetGitPackfileDirSize gets the number of 1k blocks of a git object directory
func GetGitPackfileDirSize(tb testing.TB, repoPath string) int64 {
	return getGitDirSize(tb, repoPath, "objects", "pack")
}

func getGitDirSize(tb testing.TB, repoPath string, subdirs ...string) int64 {
	cmd := exec.Command("du", "-s", "-k", filepath.Join(append([]string{repoPath}, subdirs...)...))
	output, err := cmd.Output()
	require.NoError(tb, err)
	if len(output) < 2 {
		tb.Error("invalid output of du -s -k")
	}

	outputSplit := strings.SplitN(string(output), "\t", 2)
	blocks, err := strconv.ParseInt(outputSplit[0], 10, 64)
	require.NoError(tb, err)

	return blocks
}

// WriteBlobs writes n distinct blobs into the git repository's object
// database. Each object has the current time in nanoseconds as contents.
func WriteBlobs(tb testing.TB, cfg config.Cfg, testRepoPath string, n int) []string {
	var blobIDs []string
	for i := 0; i < n; i++ {
		contents := []byte(strconv.Itoa(time.Now().Nanosecond()))
		blobIDs = append(blobIDs, WriteBlob(tb, cfg, testRepoPath, contents).String())
	}

	return blobIDs
}

// WriteBlob writes the given contents as a blob into the repository and returns its OID.
func WriteBlob(tb testing.TB, cfg config.Cfg, testRepoPath string, contents []byte) git.ObjectID {
	hex := text.ChompBytes(ExecOpts(tb, cfg, ExecConfig{Stdin: bytes.NewReader(contents)},
		"-C", testRepoPath, "hash-object", "-w", "--stdin",
	))
	oid, err := DefaultObjectHash.FromHex(hex)
	require.NoError(tb, err)
	return oid
}
