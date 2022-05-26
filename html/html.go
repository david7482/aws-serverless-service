package html

import "embed"

// F Embed all static files into compiled binary
//go:embed templates/*
var F embed.FS
