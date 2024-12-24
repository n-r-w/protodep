package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRepositoryURLWithSSH(t *testing.T) {
	target := &AuthProviderWithSSH{}
	actual := target.GetRepositoryURL("github.com/torqio/protodep")

	require.Equal(t, "ssh://github.com/torqio/protodep.git", actual)
}

func TestGetRepositoryURLWithSSHAgent(t *testing.T) {
	target := &AuthProviderWithSSHAgent{}
	actual := target.GetRepositoryURL("github.com/torqio/protodep")

	require.Equal(t, "ssh://github.com/torqio/protodep.git", actual)
}

func TestGetRepositoryURLHTTPS(t *testing.T) {
	target := &AuthProviderHTTPS{}
	actual := target.GetRepositoryURL("github.com/torqio/protodep")

	require.Equal(t, "https://github.com/torqio/protodep.git", actual)
}
