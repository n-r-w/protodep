package version

import (
	"fmt"
	"runtime/debug"
)

const (
	unspecified = "unspecified"
)

var (
	gitCommit     = unspecified
	gitCommitFull = unspecified
	buildDate     = unspecified
	version       = unspecified
)

type Info struct {
	GitCommit     string
	GitCommitFull string
	BuildDate     string
	Version       string
}

func Get() Info {
	return Info{
		GitCommit:     gitCommit,
		GitCommitFull: gitCommitFull,
		BuildDate:     buildDate,
		Version:       version,
	}
}

func (i Info) String() string {
	if i.Version == unspecified {
		info, _ := debug.ReadBuildInfo()
		return fmt.Sprintf(`{"Version": %q}`, info.Main.Version)
	}

	return fmt.Sprintf(
		`{"Version": %q, "GitCommit": %q, "GitCommitFull": %q, "BuildDate": %q}`,
		i.Version,
		i.GitCommit,
		i.GitCommitFull,
		i.BuildDate,
	)
}
