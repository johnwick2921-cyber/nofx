package api

import "nofx/store"

func latestOverlayRevision(rows []*store.PlanOverlayDB) int {
	latest := 0
	for _, row := range rows {
		if row.OverlayVersion > latest {
			latest = row.OverlayVersion
		}
	}
	return latest
}
