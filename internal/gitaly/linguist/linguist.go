package linguist

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/go-git/go-git/v5/plumbing/format/gitattributes"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"gitlab.com/gitlab-org/gitaly/v15/internal/command"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/catfile"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/gitpipe"
	"gitlab.com/gitlab-org/gitaly/v15/internal/git/localrepo"
	"gitlab.com/gitlab-org/gitaly/v15/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/v15/internal/helper/env"
	"gitlab.com/gitlab-org/gitaly/v15/internal/metadata/featureflag"
)

// ByteCountPerLanguage represents a counter value (bytes) per language.
type ByteCountPerLanguage map[string]uint64

// Instance is a holder of the defined in the system language settings.
type Instance struct {
	cfg          config.Cfg
	catfileCache catfile.Cache
	repo         *localrepo.Repo
}

// New creates a new instance that can be used to calculate language stats for
// the given repo.
func New(cfg config.Cfg, catfileCache catfile.Cache, repo *localrepo.Repo) *Instance {
	return &Instance{
		cfg:          cfg,
		catfileCache: catfileCache,
		repo:         repo,
	}
}

// Stats returns the repository's language stats as reported by 'git-linguist'.
func (inst *Instance) Stats(ctx context.Context, commitID string) (ByteCountPerLanguage, error) {
	if featureflag.GoLanguageStats.IsEnabled(ctx) {
		return inst.enryStats(ctx, commitID)
	}

	cmd, err := inst.startGitLinguist(ctx, commitID)
	if err != nil {
		return nil, fmt.Errorf("starting linguist: %w", err)
	}

	data, err := io.ReadAll(cmd)
	if err != nil {
		return nil, fmt.Errorf("reading linguist output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("waiting for linguist: %w", err)
	}

	stats := make(ByteCountPerLanguage)
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("unmarshaling stats: %w", err)
	}

	return stats, nil
}

// Color returns the color Linguist has assigned to language.
func Color(language string) string {
	if color := enry.GetColor(language); color != "#cccccc" {
		return color
	}

	colorSha := sha256.Sum256([]byte(language))
	return fmt.Sprintf("#%x", colorSha[0:3])
}

func (inst *Instance) startGitLinguist(ctx context.Context, commitID string) (*command.Command, error) {
	repoPath, err := inst.repo.Path()
	if err != nil {
		return nil, fmt.Errorf("get repo path: %w", err)
	}

	bundle, err := exec.LookPath("bundle")
	if err != nil {
		return nil, fmt.Errorf("finding bundle executable: %w", err)
	}

	cmd := []string{bundle, "exec", "bin/gitaly-linguist", "--repository=" + repoPath, "--commit=" + commitID}

	internalCmd, err := command.New(ctx, cmd,
		command.WithDir(inst.cfg.Ruby.Dir),
		command.WithEnvironment(env.AllowedRubyEnvironment(os.Environ())),
		command.WithCommandName("git-linguist", "stats"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating command: %w", err)
	}

	return internalCmd, nil
}

func (inst *Instance) enryStats(ctx context.Context, commitID string) (ByteCountPerLanguage, error) {
	stats, err := initLanguageStats(inst.repo)
	if err != nil {
		ctxlogrus.Extract(ctx).WithError(err).Info("linguist load from cache")
	}
	if stats.CommitID == commitID {
		return stats.Totals, nil
	}

	objectReader, cancel, err := inst.catfileCache.ObjectReader(ctx, inst.repo)
	if err != nil {
		return nil, fmt.Errorf("linguist create object reader: %w", err)
	}
	defer cancel()

	attrMatcher, err := inst.newAttrMatcher(ctx, objectReader, commitID)
	if err != nil {
		return nil, fmt.Errorf("linguist new attribute matcher: %w", err)
	}

	var revlistIt gitpipe.RevisionIterator

	full, err := inst.needsFullRecalculation(ctx, stats.CommitID, commitID)
	if err != nil {
		return nil, fmt.Errorf("linguist cannot determine full recalculation: %w", err)
	}

	if full {
		stats = newLanguageStats()

		skipFunc := func(result *gitpipe.RevisionResult) (bool, error) {
			// Skip files that are an excluded filetype based on filename.
			return newFileInstance(string(result.ObjectName), attrMatcher).IsExcluded(), nil
		}

		// Full recalculation is needed, so get all the files for the
		// commit using git-ls-tree(1).
		revlistIt = gitpipe.LsTree(ctx, inst.repo,
			commitID,
			gitpipe.LsTreeWithRecursive(),
			gitpipe.LsTreeWithBlobFilter(),
			gitpipe.LsTreeWithSkip(skipFunc),
		)
	} else {
		// Stats are cached for one commit, so get the git-diff-tree(1)
		// between that commit and the one we're calculating stats for.

		skipFunc := func(result *gitpipe.RevisionResult) (bool, error) {
			// Skip files that are deleted, or
			// an excluded filetype based on filename.
			if git.ObjectHashSHA1.IsZeroOID(result.OID) ||
				newFileInstance(string(result.ObjectName), attrMatcher).IsExcluded() {
				// It's a little bit of a hack to use this skip
				// function, but for every file that's deleted,
				// remove the stats.
				stats.drop(string(result.ObjectName))
				return true, nil
			}
			return false, nil
		}

		revlistIt = gitpipe.DiffTree(ctx, inst.repo,
			stats.CommitID, commitID,
			gitpipe.DiffTreeWithRecursive(),
			gitpipe.DiffTreeWithIgnoreSubmodules(),
			gitpipe.DiffTreeWithSkip(skipFunc),
		)
	}

	objectIt, err := gitpipe.CatfileObject(ctx, objectReader, revlistIt)
	if err != nil {
		return nil, fmt.Errorf("linguist gitpipe: %w", err)
	}

	for objectIt.Next() {
		object := objectIt.Result()
		filename := string(object.ObjectName)

		lang, size, err := newFileInstance(filename, attrMatcher).DetermineStats(object)
		if err != nil {
			return nil, fmt.Errorf("linguist determine stats: %w", err)
		}

		// Ensure object content is completely consumed
		if _, err := io.Copy(io.Discard, object); err != nil {
			return nil, fmt.Errorf("linguist discard excess blob: %w", err)
		}

		if len(lang) == 0 {
			stats.drop(filename)

			continue
		}

		stats.add(filename, lang, size)
	}

	if err := objectIt.Err(); err != nil {
		return nil, fmt.Errorf("linguist object iterator: %w", err)
	}

	if err := stats.save(inst.repo, commitID); err != nil {
		return nil, fmt.Errorf("linguist language stats save: %w", err)
	}

	return stats.Totals, nil
}

func (inst *Instance) newAttrMatcher(ctx context.Context, objectReader catfile.ObjectReader, commitID string) (gitattributes.Matcher, error) {
	var gitattrObject io.Reader
	var err error

	gitattrObject, err = objectReader.Object(ctx, git.Revision(commitID+":.gitattributes"))
	if catfile.IsNotFound(err) {
		gitattrObject = strings.NewReader("")
	} else if err != nil {
		return nil, fmt.Errorf("read .gitattributes: %w", err)
	}

	attrs, err := gitattributes.ReadAttributes(gitattrObject, nil, true)
	if err != nil {
		return nil, fmt.Errorf("read attr: %w", err)
	}

	// Reverse the slice because of a bug in go-git, see
	// https://github.com/go-git/go-git/pull/585
	attrsLen := len(attrs)
	attrsMid := attrsLen / 2
	for i := 0; i < attrsMid; i++ {
		j := attrsLen - i - 1
		attrs[i], attrs[j] = attrs[j], attrs[i]
	}

	return gitattributes.NewMatcher(attrs), nil
}

func (inst *Instance) needsFullRecalculation(ctx context.Context, cachedID, commitID string) (bool, error) {
	if cachedID == "" {
		return true, nil
	}

	err := inst.repo.ExecAndWait(ctx, git.SubCmd{
		Name:        "diff",
		Flags:       []git.Option{git.Flag{Name: "--quiet"}},
		Args:        []string{fmt.Sprintf("%v..%v", cachedID, commitID)},
		PostSepArgs: []string{".gitattributes"},
	})
	if err == nil {
		return false, nil
	}
	if code, ok := command.ExitStatus(err); ok && code == 1 {
		return true, nil
	}

	return true, fmt.Errorf("git diff .gitattributes: %w", err)
}
