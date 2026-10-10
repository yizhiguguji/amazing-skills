package amazing

import "embed"

// Bundle contains portable instructions, pinned upstream data and templates.
//
//go:embed promax skills catalog.json providers.lock.json qa-sources.lock.json docs/licenses examples/qa-continuous-change/sources examples/qa-continuous-change/fixtures examples/qa-security-scope/sources examples/qa-security-scope/fixtures
var Bundle embed.FS
