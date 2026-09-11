package store

import (
	"fmt"
	"strings"
)

// ── BAR-SOURCE WAVE (2026-09-10) — WHICH FEED WROTE THIS BAR ─────────────────
//
// THE FINDING. The contract-roll wave separated MNQ 09-26 from MNQ 12-26
// correctly and then booted into a chart that was still discontinuous. The boot
// line said 12-26=15992 where ~40 were expected; the research archive said why.
// For the 22:37 CT one-minute bar, same subscription, same label:
//
//	fact 16516009  bar_update       (live)    open 29360     close 29358.25
//	fact 16518205  bars_historical  (replay)  open 29068.75  close 29068.25
//
// NT8's replay serves the December contract ~290 points below Tradovate's live
// feed for the SAME minutes. A contract column cannot separate one contract from
// itself. At every boot or reconnect the replay repainted the ring and the store
// low, the live feed continued high, and the boot minute became a mixed bar —
// the 22:39 bar read open 29068.25 close 29355.25, the exact shape the 21:15
// bar had before it.
//
// THE DAMAGE. Go's upsert was unconditional, so that boot's replay overwrote
// 186 live bars across both symbols and six timeframes. Restored from the 22:15
// backup under owner authorisation (markers ac76b47b, 33fee48e).
//
// THE RULE. Every bar names its feed. Live is the minute as it traded and
// overwrites anything. Historical is a replay and fills only what live never
// wrote — in the store (InsertBars' upsert WHERE) and in the ring
// (BarCache.SeedHistorical keeps an existing live bar over an incoming
// historical one). Readers prefer live. A boot minute whose open came from a
// replay and whose close came from live is marked mixed and no reader accepts
// it.
//
// The NT8 side — whatever merge/back-adjust policy makes the replay sit on
// another scale — is filed for the AddOn wave with the two facts above as the
// evidence. This wave makes Go survive it.

// migrateSourceColumn adds the column (idempotent) and labels pre-existing rows.
//
// PRE-COLUMN ROWS ARE LABELLED LIVE. The first draft labelled them historical
// and called it conservative — "a later live bar may overwrite this". That
// reasoning covered live-over-historical and missed historical-over-historical:
// a future REPLAY overwrites a historical row, which is exactly the damage of
// 2026-09-10 22:39, and the label would have licensed it again at the very next
// boot. The store is the record of what the bot saw as it traded; the whole
// purpose of this wave is that a replay must not repaint that record. So the
// record is live, and only rows this process itself receives as a replay are
// historical.
//
// What this cannot know, stated: rows written by an earlier boot's replay (the
// 22:14–22:39 CT window on 2026-09-10, ~26 per symbol) are on the replay's
// scale and will be labelled live too. Protecting a wrong value is harmless
// here — a future replay would bring the SAME wrong value — and the repair for
// those rows is a reconstruction from the research archive's live facts, which
// writes them as live and is owner-authorised separately.
//
// Rows carrying the roll wave's spans-roll contract label are the mixed
// 21:15–21:17 bars and are marked mixed.
func (s *BarHistoryStore) migrateSourceColumn() error {
	var has int64
	if err := s.db.Raw("SELECT COUNT(*) FROM pragma_table_info('bars') WHERE name='source'").Scan(&has).Error; err != nil {
		return err
	}
	if has == 0 {
		if err := s.db.Exec("ALTER TABLE bars ADD COLUMN source TEXT NOT NULL DEFAULT ''").Error; err != nil {
			return err
		}
	}
	// GORM's AutoMigrate adds the column nullable first; normalise (roll wave lesson).
	if err := s.db.Exec("UPDATE bars SET source = '' WHERE source IS NULL").Error; err != nil {
		return err
	}
	if err := s.db.Exec("UPDATE bars SET source = ? WHERE source = '' AND contract = ?", BarSourceMixed, ContractMixed).Error; err != nil {
		return err
	}
	if err := s.db.Exec("UPDATE bars SET source = ? WHERE source = ''", BarSourceLive).Error; err != nil {
		return err
	}
	return s.db.Exec("CREATE INDEX IF NOT EXISTS idx_bars_source ON bars(symbol, tf, source, open_time_ms)").Error
}

// SourceCensus reports, per symbol, how many rows carry each source label —
// the boot line's figure, read not literal.
func (s *BarHistoryStore) SourceCensus(symbol string) (map[string]int64, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("store required")
	}
	type row struct {
		Source string
		N      int64
	}
	var rows []row
	if err := s.db.Raw("SELECT source, COUNT(*) AS n FROM bars WHERE symbol = ? GROUP BY source", symbol).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := map[string]int64{}
	for _, r := range rows {
		out[r.Source] = r.N
	}
	return out, nil
}

// IsReadableSource reports whether a bar's source may be handed to a reader.
// Mixed never is: it is two price scales in one candle.
func IsReadableSource(src string) bool {
	src = strings.TrimSpace(src)
	return src == BarSourceLive || src == BarSourceHistorical
}
