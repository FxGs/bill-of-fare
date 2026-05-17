package assets

import "embed"

//go:embed web/templates/*.html web/static/css/*
var FS embed.FS
