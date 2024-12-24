package resolver

import (
	"github.com/stretchr/testify/require"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

func currentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return filepath.Dir(filename)
}

func TestParseGitCredentials(t *testing.T) {
	t.Setenv("HOME", path.Join(currentDir(), "testdata"))

	creds, err := ParseGitCredentials()
	require.NoError(t, err)
	require.Len(t, creds, 3)

	def := creds.Get("https://invalid.com")
	require.Equal(t, def, creds["default"])

	githubCred := creds.Get("https://github.com/repo/name")
	require.Equal(t, creds["https://github.com"], githubCred)
	require.Equal(t, githubCred.Helper, []string{"!/opt/homebrew/bin/gh auth git-credential"})
}
