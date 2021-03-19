package panel

import (
	"embed"
)

//go:embed assets
var Assets embed.FS

//go:embed index.html
var Index string
