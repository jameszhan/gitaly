package localrepo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"gitlab.com/gitlab-org/gitaly/v16/internal/command"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git"
	"gitlab.com/gitlab-org/gitaly/v16/internal/git/catfile"
	"gitlab.com/gitlab-org/gitaly/v16/internal/helper/text"
	"gitlab.com/gitlab-org/gitaly/v16/internal/structerr"
	"gitlab.com/gitlab-org/gitaly/v16/proto/go/gitalypb"
)

// ErrObjectNotFound is returned in case an object could not be found.
var ErrObjectNotFound = errors.New("object not found")

// WriteBlob writes a blob to the repository's object database and
// returns its object ID. Path is used by git to decide which filters to
// run on the content.
func (repo *Repo) WriteBlob(ctx context.Context, path string, content io.Reader) (git.ObjectID, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd, err := repo.Exec(ctx,
		git.Command{
			Name: "hash-object",
			Flags: []git.Option{
				git.ValueFlag{Name: "--path", Value: path},
				git.Flag{Name: "--stdin"},
				git.Flag{Name: "-w"},
			},
		},
		git.WithStdin(content),
		git.WithStdout(stdout),
		git.WithStderr(stderr),
	)
	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", errorWithStderr(err, stderr.Bytes())
	}

	objectHash, err := repo.ObjectHash(ctx)
	if err != nil {
		return "", fmt.Errorf("detecting object hash: %w", err)
	}

	oid, err := objectHash.FromHex(text.ChompBytes(stdout.Bytes()))
	if err != nil {
		return "", err
	}

	return oid, nil
}

// FormatTagError is used by FormatTag() below
type FormatTagError struct {
	expectedLines int
	actualLines   int
}

func (e FormatTagError) Error() string {
	return fmt.Sprintf("should have %d tag header lines, got %d", e.expectedLines, e.actualLines)
}

// FormatTag is used by WriteTag (or for testing) to make the tag
// signature to feed to git-mktag, i.e. the plain-text mktag
// format. This does not create an object, just crafts input for "git
// mktag" to consume.
//
// We are not being paranoid about exhaustive input validation here
// because we're just about to run git's own "fsck" check on this.
//
// However, if someone injected parameters with extra newlines they
// could cause subsequent values to be ignored via a crafted
// message. This someone could also locally craft a tag locally and
// "git push" it. But allowing e.g. someone to provide their own
// timestamp here would at best be annoying, and at worst run up
// against some other assumption (e.g. that some hook check isn't as
// strict on locally generated data).
func FormatTag(
	objectID git.ObjectID,
	objectType string,
	tagName, tagBody []byte,
	committer *gitalypb.User,
	committerDate time.Time,
) (string, error) {
	if committerDate.IsZero() {
		committerDate = time.Now()
	}

	tagHeaderFormat := "object %s\n" +
		"type %s\n" +
		"tag %s\n" +
		"tagger %s <%s> %d %s\n"
	tagBuf := fmt.Sprintf(tagHeaderFormat, objectID.String(), objectType, tagName, committer.GetName(), committer.GetEmail(), committerDate.Unix(), committerDate.Format("-0700"))

	maxHeaderLines := 4
	actualHeaderLines := strings.Count(tagBuf, "\n")
	if actualHeaderLines != maxHeaderLines {
		return "", FormatTagError{expectedLines: maxHeaderLines, actualLines: actualHeaderLines}
	}

	tagBuf += "\n"
	tagBuf += string(tagBody)

	return tagBuf, nil
}

// MktagError is used by WriteTag() below
type MktagError struct {
	tagName []byte
	stderr  string
}

func (e MktagError) Error() string {
	// TODO: Upper-case error message purely for transitory backwards compatibility
	return fmt.Sprintf("Could not update refs/tags/%s. Please refresh and try again.", e.tagName)
}

// WriteTag writes a tag to the repository's object database with
// git-mktag and returns its object ID.
//
// It's important that this be git-mktag and not git-hash-object due
// to its fsck sanity checking semantics.
func (repo *Repo) WriteTag(
	ctx context.Context,
	objectID git.ObjectID,
	objectType string,
	tagName, tagBody []byte,
	committer *gitalypb.User,
	committerDate time.Time,
) (git.ObjectID, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	tagBuf, err := FormatTag(objectID, objectType, tagName, tagBody, committer, committerDate)
	if err != nil {
		return "", err
	}

	content := strings.NewReader(tagBuf)

	cmd, err := repo.Exec(ctx,
		git.Command{
			Name: "mktag",
		},
		git.WithStdin(content),
		git.WithStdout(stdout),
		git.WithStderr(stderr),
	)
	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", MktagError{tagName: tagName, stderr: stderr.String()}
	}

	objectHash, err := repo.ObjectHash(ctx)
	if err != nil {
		return "", fmt.Errorf("detecting object hash: %w", err)
	}

	tagID, err := objectHash.FromHex(text.ChompBytes(stdout.Bytes()))
	if err != nil {
		return "", fmt.Errorf("could not parse tag ID: %w", err)
	}

	return tagID, nil
}

// InvalidObjectError is returned when trying to get an object id that is invalid or does not exist.
type InvalidObjectError string

func (err InvalidObjectError) Error() string { return fmt.Sprintf("invalid object %q", string(err)) }

// ReadObjectInfo attempts to read the object info based on a revision.
func (repo *Repo) ReadObjectInfo(ctx context.Context, rev git.Revision) (*catfile.ObjectInfo, error) {
	infoReader, cleanup, err := repo.catfileCache.ObjectInfoReader(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("getting object info reader: %w", err)
	}
	defer cleanup()

	objectInfo, err := infoReader.Info(ctx, rev)
	if err != nil {
		if errors.As(err, &catfile.NotFoundError{}) {
			return nil, InvalidObjectError(rev)
		}
		return nil, fmt.Errorf("getting object info: %w", err)
	}

	return objectInfo, nil
}

// ReadObject reads an object from the repository's object database. InvalidObjectError
// is returned if the oid does not refer to a valid object.
func (repo *Repo) ReadObject(ctx context.Context, oid git.ObjectID) ([]byte, error) {
	objectReader, cancel, err := repo.catfileCache.ObjectReader(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("create object reader: %w", err)
	}
	defer cancel()

	object, err := objectReader.Object(ctx, oid.Revision())
	if err != nil {
		if errors.As(err, &catfile.NotFoundError{}) {
			return nil, InvalidObjectError(oid.String())
		}
		return nil, fmt.Errorf("get object from reader: %w", err)
	}

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("read object from reader: %w", err)
	}

	return data, nil
}

type readCommitConfig struct {
	withTrailers bool
}

// ReadCommitOpt is an option for ReadCommit.
type ReadCommitOpt func(*readCommitConfig)

// WithTrailers will cause ReadCommit to parse commit trailers.
func WithTrailers() ReadCommitOpt {
	return func(cfg *readCommitConfig) {
		cfg.withTrailers = true
	}
}

// ReadCommit reads the commit specified by the given revision. If no such
// revision exists, it will return an ErrObjectNotFound error.
func (repo *Repo) ReadCommit(ctx context.Context, revision git.Revision, opts ...ReadCommitOpt) (*gitalypb.GitCommit, error) {
	var cfg readCommitConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	objectReader, cancel, err := repo.catfileCache.ObjectReader(ctx, repo)
	if err != nil {
		return nil, err
	}
	defer cancel()

	var commit *gitalypb.GitCommit
	if cfg.withTrailers {
		commit, err = catfile.GetCommitWithTrailers(ctx, repo.gitCmdFactory, repo, objectReader, revision)
	} else {
		commit, err = catfile.GetCommit(ctx, objectReader, revision)
	}

	if err != nil {
		if errors.As(err, &catfile.NotFoundError{}) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}

	return commit, nil
}

// InvalidCommitError is returned when the revision does not point to a valid commit object.
type InvalidCommitError git.Revision

func (err InvalidCommitError) Error() string {
	return fmt.Sprintf("invalid commit: %q", string(err))
}

// IsAncestor returns whether the parent is an ancestor of the child. InvalidCommitError is returned
// if either revision does not point to a commit in the repository.
func (repo *Repo) IsAncestor(ctx context.Context, parent, child git.Revision) (bool, error) {
	const notValidCommitName = "fatal: Not a valid commit name"

	stderr := &bytes.Buffer{}
	if err := repo.ExecAndWait(ctx,
		git.Command{
			Name:  "merge-base",
			Flags: []git.Option{git.Flag{Name: "--is-ancestor"}},
			Args:  []string{parent.String(), child.String()},
		},
		git.WithStderr(stderr),
	); err != nil {
		status, ok := command.ExitStatus(err)
		if ok && status == 1 {
			return false, nil
		} else if ok && strings.HasPrefix(stderr.String(), notValidCommitName) {
			commitOID := strings.TrimSpace(strings.TrimPrefix(stderr.String(), notValidCommitName))
			return false, InvalidCommitError(commitOID)
		}

		return false, fmt.Errorf("determine ancestry: %w, stderr: %q", err, stderr)
	}

	return true, nil
}

// BadObjectError is returned when attempting to walk a bad object.
type BadObjectError struct {
	// ObjectID is the object id of the object that was bad.
	ObjectID git.ObjectID
}

// Error returns the error message.
func (err BadObjectError) Error() string {
	return fmt.Sprintf("bad object %q", err.ObjectID)
}

// ObjectReadError is returned when reading an object fails.
type ObjectReadError struct {
	// ObjectID is the object id of the object that git failed to read
	ObjectID git.ObjectID
}

// Error returns the error message.
func (err ObjectReadError) Error() string {
	return fmt.Sprintf("failed reading object %q", err.ObjectID)
}

var (
	regexpBadObjectError  = regexp.MustCompile(`^fatal: bad object ([[:xdigit:]]*)\n$`)
	regexpObjectReadError = regexp.MustCompile(`^error: Could not read ([[:xdigit:]]*)\n`)
)

// WalkUnreachableObjects walks the object graph starting from heads and writes to the output object IDs
// that are included in the walk but unreachable from any of the repository's references. Heads should
// return object IDs separated with a newline. Output is object IDs separated by newlines.
func (repo *Repo) WalkUnreachableObjects(ctx context.Context, heads io.Reader, output io.Writer) error {
	var stderr bytes.Buffer
	if err := repo.ExecAndWait(ctx,
		git.Command{
			Name: "rev-list",
			Flags: []git.Option{
				git.Flag{Name: "--objects"},
				git.Flag{Name: "--not"},
				git.Flag{Name: "--all"},
				git.Flag{Name: "--stdin"},
			},
		},
		git.WithStdin(heads),
		git.WithStdout(output),
		git.WithStderr(&stderr),
	); err != nil {
		if matches := regexpBadObjectError.FindSubmatch(stderr.Bytes()); len(matches) > 1 {
			return BadObjectError{ObjectID: git.ObjectID(matches[1])}
		}

		if matches := regexpObjectReadError.FindSubmatch(stderr.Bytes()); len(matches) > 1 {
			return ObjectReadError{ObjectID: git.ObjectID(matches[1])}
		}

		return structerr.New("rev-list: %w", err).WithMetadata("stderr", stderr.String())
	}

	return nil
}

// PackObjects takes in object IDs separated by newlines. It packs the objects into a pack file and
// writes it into the output.
func (repo *Repo) PackObjects(ctx context.Context, objectIDs io.Reader, output io.Writer) error {
	var stderr bytes.Buffer
	if err := repo.ExecAndWait(ctx,
		git.Command{
			Name: "pack-objects",
			Flags: []git.Option{
				git.Flag{Name: "-q"},
				git.Flag{Name: "--stdout"},
			},
		},
		git.WithStdin(objectIDs),
		git.WithStderr(&stderr),
		git.WithStdout(output),
	); err != nil {
		return structerr.New("pack objects: %w", err).WithMetadata("stderr", stderr.String())
	}

	return nil
}

// UnpackObjects unpacks the objects from the pack file to the repository's object database.
func (repo *Repo) UnpackObjects(ctx context.Context, packFile io.Reader) error {
	stderr := &bytes.Buffer{}
	if err := repo.ExecAndWait(ctx,
		git.Command{
			Name: "unpack-objects",
			Flags: []git.Option{
				git.Flag{Name: "-q"},
			},
		},
		git.WithStdin(packFile),
		git.WithStderr(stderr),
	); err != nil {
		return structerr.New("unpack objects: %w", err).WithMetadata("stderr", stderr.String())
	}

	return nil
}
