package core

import (
	"github.com/fbiville/headache/internal/pkg/fs"
	"github.com/fbiville/headache/internal/pkg/helper"
	"github.com/fbiville/headache/internal/pkg/vcs"
)

func DefaultEnvironment() *Environment {
	return &Environment{
		VersioningClient: &vcs.Client{
			Vcs: &vcs.Git{},
		},
		FileSystem:     fs.DefaultFileSystem(),
		Clock:          helper.SystemClock{},
		SchemaLocation: "https://fbiville.github.io/headache/schema.json",
	}
}

type Environment struct {
	VersioningClient vcs.VersioningClient
	FileSystem       *fs.FileSystem
	Clock            helper.Clock
	SchemaLocation   string
}
