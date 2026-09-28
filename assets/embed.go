package assets

import "embed"

//go:embed themes/*.json ranks.json logo.json digits.json messages.json
var Files embed.FS
