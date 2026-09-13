package trader

import (
	"errors"
	"nofx/kernel"
	"nofx/store"
	"strings"
	"testing"
)

type unknownPositionEntryTrader struct {
	MockTrader
	reads, opens, closes int
	firstHeld            bool
}

func (m *unknownPositionEntryTrader) GetPositions() ([]map[string]interface{}, error) {
	m.reads++
	if m.firstHeld && m.reads == 1 {
		return []map[string]interface{}{{"symbol": "MNQ", "side": "LONG", "positionAmt": 1.0}}, nil
	}
	return nil, errors.New("post-exit account snapshot unknown")
}
func (m *unknownPositionEntryTrader) OpenLong(string, float64, int) (map[string]interface{}, error) {
	m.opens++
	return nil, nil
}
func (m *unknownPositionEntryTrader) CloseLong(string, float64) (map[string]interface{}, error) {
	m.closes++
	return nil, nil
}
func TestProductionOpenRefusesUnknownAccountPositions(t *testing.T) {
	m := &unknownPositionEntryTrader{}
	at := &AutoTrader{id: "unknown-position", exchange: "ninjatrader", trader: m}
	err := at.executeOpenLongWithRecord(&kernel.Decision{Symbol: "MNQ"}, &store.DecisionAction{})
	if err == nil || !strings.Contains(err.Error(), "reconcile-before-open") || m.opens != 0 || m.closes != 0 || m.reads != 1 {
		t.Fatalf("entry failed to refuse unknown: %v reads=%d opens=%d closes=%d", err, m.reads, m.opens, m.closes)
	}
}
func TestFlattenWaitDoesNotTreatUnreadableAsFlat(t *testing.T) {
	m := &unknownPositionEntryTrader{firstHeld: true}
	at := &AutoTrader{id: "unknown-after-flatten", exchange: "ninjatrader", trader: m}
	if err := at.reconcileBeforeOpenNT("MNQ", "long"); err == nil || m.closes != 1 || m.opens != 0 {
		t.Fatalf("unreadable became flat: err=%v closes=%d opens=%d", err, m.closes, m.opens)
	}
}
