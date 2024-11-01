package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/n-r-w/protodep/internal/auth"
	"github.com/n-r-w/protodep/internal/config"
	"github.com/n-r-w/protodep/internal/logger"
)

const masterBranch = "master"

type Git struct {
	protodepDir  string
	dep          config.ProtoDepDependency
	authProvider auth.AuthProvider
}

func NewGit(protodepDir string, dep config.ProtoDepDependency, authProvider auth.AuthProvider) *Git {
	return &Git{
		protodepDir:  protodepDir,
		dep:          dep,
		authProvider: authProvider,
	}
}

type OpenedRepository struct {
	Repository *git.Repository
	Dep        config.ProtoDepDependency
	Hash       string
}

func (r *Git) Open() (*OpenedRepository, error) {
	branch := masterBranch
	if r.dep.Branch != "" {
		branch = r.dep.Branch
	}

	revision := r.dep.Revision

	reponame := r.dep.Repository()
	repopath := filepath.Join(r.protodepDir, reponame)

	authMethod, err := r.authProvider.AuthMethod()
	if err != nil {
		return nil, err
	}

	var (
		rep  *git.Repository
		stat os.FileInfo
	)

	if stat, err = os.Stat(repopath); err == nil && stat.IsDir() {
		spinner := logger.InfoWithSpinner("Getting %s ", reponame)

		rep, err = git.PlainOpen(repopath)
		if err != nil {
			return nil, fmt.Errorf("open repository: %w", err)
		}
		spinner.Stop()

		fetchOpts := &git.FetchOptions{
			Auth: authMethod,
		}

		// TODO: Validate remote setting.
		// TODO: If .protodep cache remains with SSH, change remote target to HTTPS.

		if err = rep.Fetch(fetchOpts); err != nil {
			if err != git.NoErrAlreadyUpToDate {
				return nil, fmt.Errorf("fetch repository %s: %w", repopath, err)
			}
		}
		spinner.Finish()

	} else {
		spinner := logger.InfoWithSpinner("Getting %s ", reponame)
		url := r.authProvider.GetRepositoryURL(reponame)
		// IDEA: Is it better to register both ssh and HTTP?
		rep, err = git.PlainClone(repopath, false, &git.CloneOptions{
			Auth: authMethod,
			URL:  url,
		})
		if err != nil {
			return nil, fmt.Errorf("clone repository %s: %w", url, err)
		}
		spinner.Finish()
	}

	wt, err := rep.Worktree()
	if err != nil {
		return nil, fmt.Errorf("get worktree: %w", err)
	}

	if revision == "" {
		var target *plumbing.Reference
		target, err = r.resolveReference(rep, branch)
		if err != nil {
			return nil, fmt.Errorf("change branch to %s: %w", branch, err)
		}

		if err = wt.Checkout(&git.CheckoutOptions{Hash: target.Hash()}); err != nil {
			return nil, fmt.Errorf("checkout revision to %s: %w", revision, err)
		}

		head := plumbing.NewHashReference(plumbing.HEAD, target.Hash())
		if err = rep.Storer.SetReference(head); err != nil {
			return nil, fmt.Errorf("set head to %s: %w", branch, err)
		}
	} else {
		var opts git.CheckoutOptions

		tag := plumbing.NewTagReferenceName(revision)
		_, err = rep.Reference(tag, false)
		if err != nil && err != plumbing.ErrReferenceNotFound {
			return nil, fmt.Errorf("tag '%s' reference: %w", tag, err)
		} else {
			if err != nil {
				// Tag not found, revision must be a hash
				logger.Info("%s is not a tag, checking out by hash", revision)
				hash := plumbing.NewHash(revision)
				opts = git.CheckoutOptions{Hash: hash}
			} else {
				logger.Info("%s is a tag, checking out by tag", revision)
				opts = git.CheckoutOptions{Branch: tag}
			}
		}

		if err = wt.Checkout(&opts); err != nil {
			return nil, fmt.Errorf("checkout to %s: %w", revision, err)
		}
	}

	commiter, err := rep.Log(&git.LogOptions{})
	if err != nil {
		return nil, fmt.Errorf("get commit: %w", err)
	}

	current, err := commiter.Next()
	if err != nil {
		return nil, fmt.Errorf("get current commit: %w", err)
	}

	return &OpenedRepository{
		Repository: rep,
		Dep:        r.dep,
		Hash:       current.Hash.String(),
	}, nil
}

func (r *Git) ProtoRootDir() string {
	return filepath.Join(r.protodepDir, r.dep.Target)
}

func (r *Git) resolveReference(rep *git.Repository, branch string) (*plumbing.Reference, error) {
	if branch != masterBranch {
		return r.getReference(rep, branch)
	}
	// If master branch is failed, try main branch.
	target, err := r.getReference(rep, branch)
	if err == plumbing.ErrReferenceNotFound {
		return r.getReference(rep, "main")
	}
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (r *Git) getReference(rep *git.Repository, branch string) (*plumbing.Reference, error) {
	return rep.Storer.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", branch)))
}
