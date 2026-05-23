package assets

import "embed"

//go:embed web/templates/*.html web/static/css/* seed/menu.yaml
var FS embed.FS
