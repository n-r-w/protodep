package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Dependency interface {
	Load() (*ProtoDep, error)
}

type DependencyImpl struct {
	targetDir string
	tomlPath  string
}

func NewDependency(targetDir string) Dependency {
	return &DependencyImpl{
		targetDir: targetDir,
		tomlPath:  filepath.Join(targetDir, "protodep.toml"),
	}
}

func (d *DependencyImpl) Load() (*ProtoDep, error) {
	content, err := os.ReadFile(filepath.Clean(d.tomlPath))
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", d.tomlPath, err)
	}

	var conf ProtoDep
	if _, err := toml.Decode(string(content), &conf); err != nil {
		return nil, fmt.Errorf("decode toml: %w", err)
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("found invalid configuration: %w", err)
	}

	return &conf, nil
}
