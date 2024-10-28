package contrib

import (
	"embed"
)

//go:embed wasm/*
var fs embed.FS

// FS returns the embedded filesystem for the contrib package.
func FS() embed.FS {
	return fs
}
