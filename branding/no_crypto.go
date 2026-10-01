package branding

// C8 + C13 guard (crypto removal, plan v10 FINAL).
//
// This file is the SINGLE home of the sweep/host literals (plan v10 C13:
// "ONE regex, exported as a literal from the guard and used by the tables
// and the gate — never re-typed"). The builders extract these literals
// programmatically from THIS file (never hand-typed), and
// scripts/crypto-union-gate.sh asserts every disposition table's `regex:`
// header equals SweepRegexLiteral byte-for-byte.
//
// GO item 2 was DECLINED 2026-10-01 (CTO ruling): stocks/forex stay. The
// literals below are the BASE form — do NOT append alpaca|twelvedata|sina.

// SweepRegexLiteral is THE assembled C13 sweep regex, byte-for-byte the
// literal in plan v10 line 108 (the `aster\b`/`lighter\b` bounds carry BOTH
// word boundaries; `quant\b` keeps its trailing boundary; `"mixed"` matches
// only the DOUBLE-quoted literal). Run case-insensitive with Go regexp or
// GNU grep -E (`\b` is not POSIX ERE).
const SweepRegexLiteral = `binance|bybit|okx|bitget|kucoin|gate\.io|gateio|indodax|hyperliquid|\baster\b|asterdex|\blighter\b|coinank|usdc|usdt|x402|claw402|blockrun|wallet|ai500|hyper_all|hyper_main|oi_top|oi_low|netflow|quant\b|price ranking|"mixed"|币安|欧易|火币|U本位|永续`

// HostCensusLiteral is the C8 host census list (plan v10 C8), byte-for-byte:
// non-test Go files and web/src hold ZERO of these. GO item 2 declined, so
// no alpaca.markets|twelvedata.com|sinajs… additions.
const HostCensusLiteral = `binance\.com|binance\.vision|bybit\.com|okx\.com|bitget\.com|kucoin\.com|gateio|gate\.io|indodax\.com|hyperliquid\.xyz|asterdex|lighter\.xyz|coinank|claw402\.ai|blockrun`

// DeletedImportPrefixes is DERIVED from the plan C4 directory list (one
// import-prefix entry per deleted top-level dir), relative to the module
// root. The census fails on any Go import whose path contains one of these
// as a path segment prefix.
var DeletedImportPrefixes = []string{
	"trader/aster", "trader/binance", "trader/bitget", "trader/bybit",
	"trader/gate", "trader/hyperliquid", "trader/indodax", "trader/kucoin",
	"trader/lighter", "trader/okx",
	"provider/coinank", "provider/hyperliquid",
	"mcp/payment", "wallet",
}

// DeletedSDKModules is the C4 go.mod drop list. CR-A's go mod tidy diff
// appends every additional module tidy removes (quoted in the PR) before
// the integrated head.
var DeletedSDKModules = []string{
	"go-binance", "bybit.go.api", "gateapi-go", "antihax/optional",
	"lighter-go", "poseidon", "go-hyperliquid", "go-ethereum",
}

// ContentAssertSites are the single-quoted sites the sweep regex CANNOT see
// (a trailing-boundary regex with a double-quoted "mixed" never matches
// 'mixed'). CTO ruling 2026-10-01: assert them by CONTENT in the gate,
// separately from the regex sweep. disposition is what the CR-C table must
// carry for the site.
var ContentAssertSites = []struct{ File, Needle, Disposition string }{
	{"web/src/components/plan/ExecutorVerdict.tsx", `arm.state === 'mixed'`, "KEEP"},
	{"web/src/components/trader/TraderConfigModal.tsx", `source_type === 'mixed'`, "CUT"},
}
