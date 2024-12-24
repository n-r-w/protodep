package resolver

import (
	"fmt"
	"path/filepath"

	"github.com/torqio/protodep/internal/auth"
	"github.com/torqio/protodep/internal/logger"
)

type Config struct {
	// UseHttps will force https on each proto dependencies fetch.
	UseHttps bool

	// UseGitCredentialsHelper will use git credentials helper for authentication.
	UseGitCredentialsHelper bool

	// UseNetrc will use netrc file for authentication.
	UseNetrc bool

	// HomeDir is the home directory, used as root to find ssh identity files.
	HomeDir string

	// TargetDir is the dependencies directory where protodep.toml files are located.
	TargetDir string

	// OutputDir is the directory where proto files will be cloned.
	OutputDir string

	// BasicAuthUsername is used if `https` mode  is enable. Optional, only if dependency repository needs authentication.
	BasicAuthUsername string

	// BasicAuthPassword is used if `https` mode is enable. Optional, only if dependency repository needs authentication.
	BasicAuthPassword string

	// IdentityFile is used if `ssh` mode is enable. Optional, it is computed like {home}/.ssh/
	IdentityFile string

	// IdentityPassword is used if `ssh` mode is enable. Optional, only if identity file needs a passphrase.
	IdentityPassword string
}

// GetHttpsAuthProvider returns auth provider for https
func (c *Config) GetHttpsAuthProvider() (auth.AuthProvider, error) {
	return auth.NewAuthProvider(auth.WithHTTPS(c.BasicAuthUsername, c.BasicAuthPassword)), nil
}

// GetSshAuthProvider returns auth provider for ssh
func (c *Config) GetSshAuthProvider() (auth.AuthProvider, error) {
	if c.IdentityFile == "" && c.IdentityPassword == "" {
		return auth.NewAuthProvider(), nil
	}

	identifyPath := filepath.Join(c.HomeDir, ".ssh", c.IdentityFile)
	isSSH, err := isAvailableSSH(identifyPath)
	if err != nil {
		return nil, fmt.Errorf("Config.GetSshAuthProvider: %w", err)
	}

	if isSSH {
		return auth.NewAuthProvider(auth.WithPemFile(identifyPath, c.IdentityPassword)), nil
	}

	logger.Warn("The identity file path has been passed but is not available. Falling back to ssh-agent, the default authentication method.")
	return auth.NewAuthProvider(), nil
}
