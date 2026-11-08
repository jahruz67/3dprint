// Package profiles bundles starter materials and developer-only reference fixtures.
package profiles

import "embed"

//go:embed *.json
var Files embed.FS
