// Package docs menyediakan spesifikasi OpenAPI yang ditanam (embed) ke dalam binary.
package docs

import _ "embed"

//go:embed api.json
var Spec []byte
