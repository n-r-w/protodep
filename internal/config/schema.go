package config

import (
	"errors"
	"strings"
)

type ProtoDep struct {
	ProtoOutdir  string               `toml:"proto_outdir"`
	Dependencies []ProtoDepDependency `toml:"dependencies"`
}

func (d *ProtoDep) Validate() error {
	if strings.TrimSpace(d.ProtoOutdir) == "" {
		return errors.New("required 'proto_outdir'")
	}
	return nil
}

type ProtoDepDependency struct {
	Target      string   `toml:"target"`
	LocalFolder string   `toml:"local_folder"`
	Subgroup    string   `toml:"subgroup"`
	Revision    string   `toml:"revision"`
	Branch      string   `toml:"branch"`
	Path        string   `toml:"path"`
	Ignores     []string `toml:"ignores"`
	Includes    []string `toml:"includes"`
	Protocol    string   `toml:"protocol"`
	UsernameEnv string   `toml:"username_env"`
	PasswordEnv string   `toml:"password_env"`
}

func (d *ProtoDepDependency) Repository() string {
	tokens := strings.Split(d.Target, "/")
	subgroupTokens := make([]string, 0)
	if d.Subgroup != "" {
		subgroupTokens = strings.Split(d.Subgroup, "/")
	}
	repoTokens := 3 + len(subgroupTokens) //nolint:gomnd
	if len(tokens) > repoTokens {
		return strings.Join(tokens[0:repoTokens], "/")
	} else {
		return d.Target
	}
}

func (d *ProtoDepDependency) Directory() string {
	r := d.Repository()

	if d.Target == r {
		return "."
	} else {
		return "." + strings.Replace(d.Target, r, "", 1)
	}
}

func (d *ProtoDepDependency) Machine() string {
	tokens := strings.Split(d.Target, "/")
	if len(tokens) < 1 {
		return ""
	}
	return tokens[0]
}
