// Package ui just embed the static files
package ui

import "embed"

//go:embed "html" "static"
var Files embed.FS
