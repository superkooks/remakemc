package assets

import "embed"

// Embed all assets in this directory
//
//go:embed *
var Files embed.FS
