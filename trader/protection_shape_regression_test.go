package trader

import (
	nt "nofx/provider/ninjatrader"
	"testing"
)

func TestProtectionRequiresOrderShapeDespiteStopName(t *testing.T) {
	for _, tc := range []struct {
		name, action, typ string
		want              protectionAction
	}{
		{"wrong side", "BuyToCover", "StopMarket", protectionPlace},
		{"limit named stop", "Sell", "Limit", protectionPlace},
		{"unknown action", "", "StopMarket", protectionUnknown},
		{"unknown type", "Sell", "", protectionUnknown},
		{"correct closing stop", "Sell", "StopMarket", protectionOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := nt.NT8Order{Symbol: "MNQ", Name: "fixture-sl", Action: tc.action, Type: tc.typ, State: "Accepted", Quantity: 1, StopPrice: 29500}
			v := adjudicateProtection("MNQ", "LONG", 1, []nt.NT8Order{o}, true, nil, 29500)
			if v.Action != tc.want {
				t.Fatalf("action=%s want=%s: %s", v.Action, tc.want, v.Why)
			}
		})
	}
}
