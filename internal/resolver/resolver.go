package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/n-r-w/protodep/internal/auth"
	"github.com/n-r-w/protodep/internal/config"
	"github.com/n-r-w/protodep/internal/logger"
	"github.com/n-r-w/protodep/internal/repository"
)

type protoResource struct {
	source       string
	relativeDest string
}

type Resolver struct {
	conf *Config

	httpsProvider auth.AuthProvider
	sshProvider   auth.AuthProvider

	netrcInfo []netrcLine
}

func New(conf *Config, httpsProvider, sshProvider auth.AuthProvider) (*Resolver, error) {
	s := &Resolver{
		conf:          conf,
		httpsProvider: httpsProvider,
		sshProvider:   sshProvider,
	}

	netrcInfo, err := readNetrc()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read netrc: %w", err)
	}

	s.netrcInfo = netrcInfo

	return s, nil
}

func (s *Resolver) Resolve(cleanupCache bool) error { //nolint:gocognit
	dep := config.NewDependency(s.conf.TargetDir)
	protodep, err := dep.Load()
	if err != nil {
		return err
	}

	protodepDir := filepath.Join(s.conf.HomeDir, ".protodep")

	_, err = os.Stat(protodepDir)
	if cleanupCache && err == nil {
		files, err := os.ReadDir(protodepDir)
		if err != nil {
			return err
		}
		for _, file := range files {
			if file.IsDir() {
				dirpath := filepath.Join(protodepDir, file.Name())
				if err := os.RemoveAll(dirpath); err != nil {
					return err
				}
			}
		}
	}

	outdir := filepath.Join(s.conf.OutputDir, protodep.ProtoOutdir)
	if err := os.RemoveAll(outdir); err != nil {
		return err
	}

	for _, dep := range protodep.Dependencies {
		var sources []protoResource

		if dep.Target != "" && dep.LocalFolder != "" {
			return fmt.Errorf("target and local_folder cannot be set together")
		}

		if dep.LocalFolder != "" {
			if dep.Subgroup != "" || dep.Revision != "" || dep.Branch != "" ||
				dep.Protocol != "" || dep.UsernameEnv != "" || dep.PasswordEnv != "" {
				return fmt.Errorf("subgroup, revision, branch, path, protocol and username_env cannot be set together with local_folder")
			}

			localFolder, err := filepath.Abs(dep.LocalFolder)
			if err != nil {
				return fmt.Errorf("invalid local_folder: %w", err)
			}

			sources, err = s.getSources(dep, localFolder)
			if err != nil {
				return err
			}
		} else if dep.Target != "" {
			gitrepo, err := s.getRepository(dep, protodepDir)
			if err != nil {
				return err
			}
			if _, err = gitrepo.Open(); err != nil {
				return err
			}
			sources, err = s.getSources(dep, gitrepo.ProtoRootDir())
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("target or local_folder must be set")
		}

		for _, s := range sources {
			outpath := filepath.Join(outdir, dep.Path, s.relativeDest)

			content, err := os.ReadFile(s.source)
			if err != nil {
				return err
			}

			if err := writeFileWithDirectory(outpath, content, 0o644); err != nil { //nolint:gomnd
				return err
			}
		}

	}

	return nil
}

func (s *Resolver) getRepository(dep config.ProtoDepDependency, protodepDir string) (*repository.Git, error) {
	var (
		authProvider           auth.AuthProvider
		userName, userPassword string
	)

	if dep.PasswordEnv != "" || dep.UsernameEnv != "" {
		if dep.UsernameEnv == "" || dep.PasswordEnv == "" {
			return nil, fmt.Errorf("auth_username_env and auth_password_env must be set together")
		}

		userName = os.Getenv(dep.UsernameEnv)
		userPassword = os.Getenv(dep.PasswordEnv)

		if userName == "" {
			return nil, fmt.Errorf("auth_username_env %s is empty", dep.UsernameEnv)
		}

		if userPassword == "" {
			return nil, fmt.Errorf("auth_password_env %s is empty", dep.PasswordEnv)
		}
	} else if s.conf.UseNetrc {
		machine := dep.Machine()

		for _, netrc := range s.netrcInfo {
			if netrc.machine == machine && netrc.login != "" && netrc.password != "" {
				userName = netrc.login
				userPassword = netrc.password
				break
			}
		}
	}

	if s.conf.UseHttps || dep.Protocol == "https" || (dep.Protocol == "" && userName != "") {
		if userName != "" {
			authProvider = auth.NewAuthProvider(auth.WithHTTPS(userName, userPassword))
		} else {
			authProvider = s.httpsProvider
		}
	} else {
		if dep.Protocol == "ssh" {
			if dep.UsernameEnv != "" {
				return nil, fmt.Errorf("auth_username_env and auth_password_env are not supported for ssh protocol")
			}
			authProvider = s.sshProvider
		}
	}

	return repository.NewGit(protodepDir, dep, authProvider), nil
}

func (s *Resolver) getSources(dep config.ProtoDepDependency, protoRootDir string) ([]protoResource, error) {
	sources := make([]protoResource, 0)

	compiledIgnores := compileIgnoreToGlob(dep.Ignores)
	compiledIncludes := compileIgnoreToGlob(dep.Includes)

	hasIncludes := len(dep.Includes) > 0

	err := filepath.Walk(protoRootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".proto") {
			isIncludePath := s.isMatchPath(protoRootDir, path, dep.Includes, compiledIncludes)
			isIgnorePath := s.isMatchPath(protoRootDir, path, dep.Ignores, compiledIgnores)

			if hasIncludes && !isIncludePath {
				logger.Info("skipped %s due to include setting", path)
			} else if isIgnorePath {
				logger.Info("skipped %s due to ignore setting", path)
			} else {
				sources = append(sources, protoResource{
					source:       path,
					relativeDest: strings.Replace(path, protoRootDir, "", -1),
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func compileIgnoreToGlob(ignores []string) []glob.Glob {
	globIgnores := make([]glob.Glob, len(ignores))

	for idx, ignore := range ignores {
		globIgnores[idx] = glob.MustCompile(ignore)
	}

	return globIgnores
}

func (s *Resolver) isMatchPath(protoRootDir, target string, paths []string, globMatch []glob.Glob) bool {
	// convert slashes otherwise doesnt work on windows same was as on linux
	target = filepath.ToSlash(target)

	// keeping old logic for backward compatibility
	for _, pathToMatch := range paths {
		// support windows paths correctly
		pathPrefix := filepath.ToSlash(filepath.Join(protoRootDir, pathToMatch))
		if strings.HasPrefix(target, pathPrefix) {
			return true
		}
	}

	for _, pathToMatch := range globMatch {
		if pathToMatch.Match(target) {
			return true
		}
	}

	return false
}

func writeFileWithDirectory(path string, data []byte, perm os.FileMode) error {
	path = filepath.ToSlash(path)
	s := strings.Split(path, "/")

	var dir string
	if len(s) > 1 {
		dir = strings.Join(s[0:len(s)-1], "/")
	} else {
		dir = path
	}

	dir = filepath.FromSlash(dir)
	path = filepath.FromSlash(path)

	if err := os.MkdirAll(dir, 0o750); err != nil { //nolint:gomnd
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("write data to %s: %w", path, err)
	}

	return nil
}

// isAvailableSSH is Check whether this machine can use git protocol
func isAvailableSSH(identifyPath string) (bool, error) {
	if _, err := os.Stat(identifyPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// TODO: validate ssh key
	return true, nil
}
