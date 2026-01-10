// package ui just embed the static files
package ui

import "embed"

//go:embed "static"
var Files embed.FS
