package resolver

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseCredentials(t *testing.T) {
	r, err := ParseGitCredentials()
	require.NoError(t, err)

	repoURL := "https://github.com/galprz-torq/poc-native-ai.git"

	c := r.Get(repoURL)
	require.NotNil(t, c)

	gc, err := c.Evaluate(repoURL)
	require.NoError(t, err)

	fmt.Print(gc.String())
}
