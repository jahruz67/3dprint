// Package schemas exposes the exact portable schemas bundled with the application.
package schemas

import "embed"

//go:embed *.schema.json
var Files embed.FS
