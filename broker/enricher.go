package broker

import "context"

// Enricher resolves per-tool Hints by tool name. Returns ok=false (with err=nil)
// when the tool has no enrichment record — this is expected for most tools and
// callers should treat it as "no hints to surface" rather than an error.
//
// Storage is consumer-owned: implementations may back lookups with SQLite,
// an in-memory map, a remote service, etc. go-toolbroker ships only the
// interface and a no-op default (NopEnricher).
type Enricher interface {
	LookupByToolName(ctx context.Context, toolName string) (Hints, bool, error)
}

// NopEnricher is a no-op Enricher: every lookup returns ok=false, err=nil.
// Useful as a test stand-in and as a safe default when no storage-backed
// enricher has been wired.
type NopEnricher struct{}

// LookupByToolName always returns zero-value Hints, ok=false, nil.
func (NopEnricher) LookupByToolName(context.Context, string) (Hints, bool, error) {
	return Hints{}, false, nil
}
