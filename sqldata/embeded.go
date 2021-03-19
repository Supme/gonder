package sqldata

import "embed"

//go:embed update dump.sql
var Dump embed.FS
