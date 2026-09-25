package migrations

import "embed"

// The pattern is `*` and not `*.sql` so the package still builds with no
// migrations yet. Files that are not migrations are ignored when reading it.
//
//go:embed *
var FS embed.FS
