package trader

// UPDATER-NT8-CLOSED — the nt8_absent gate verdict (owner ruling 22:1x CT
// 09-27: NT8 closed = the safest moment to update). Production call site:
// InstallationGateStatus over the real leg closures. Each case breaks exactly
// one thing and names the absent leg that must fail. The mutations the CTO
// will run — drop the 60s window, drop each ledger leg, drop the revoke, drop
// the SIM check — must each fail one of these.

import (
	"strings"
	"testing"
	"time"

	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
	ntTrader "nofx/trader/ninjatrader"
)

// newAbsentFixture: NT8 link down 90 s, hold "job-g", barrier engaged, one
// bound Sim101 account, flat ledger, no planner reads, no queued signals.
func newAbsentFixture(t *testing.T) (*gateFixture, *ntwire.TCPServer) {
	t.Helper()
	dir := withMaintenanceDir(t)
	at, st := resetTrader(t, store.StrategyConfig{})
	at.id = "gate-t1"
	s := ntwire.NewTCPServer(nil)
	s.SetAccountsList([]ntwire.AccountInfo{{Name: "Sim101", IsSim: true}}, "Sim101")
	at.trader = ntTrader.NewTCPTrader(s, "MNQ", "Sim101")
	setHold(t, dir, "job-g")
	f := &gateFixture{dir: dir, st: st, loaded: map[string]*AutoTrader{at.id: at},
		wire: installationWire{
			Connected:          false,
			Rec:                ntwire.ConnectionRecord{AcceptSeq: 3, RemotePort: 50123},
			HasAck:             false,
			DisconnectedAt:     time.Now().Add(-90 * time.Second),
			HaveDisconnectedAt: true,
		}}
	prevW, prevC := installationWireView, installationTraderCutover
	installationWireView = func([]*AutoTrader) (installationWire, bool) { return f.wire, true }
	installationTraderCutover = func(*AutoTrader) []CutoverLeg {
		return []CutoverLeg{{N: 1, Name: "db_open_positions", Pass: true}, {N: 2, Name: "api_positions", Pass: true}, {N: 4, Name: "working_orders", Pass: true}}
	}
	t.Cleanup(func() { installationWireView, installationTraderCutover = prevW, prevC })
	return f, s
}

func absentLegOf(a *NT8AbsentView, name string) (InstallationGateLeg, bool) {
	if a == nil {
		return InstallationGateLeg{}, false
	}
	for _, l := range a.Legs {
		if l.Name == name {
			return l, true
		}
	}
	return InstallationGateLeg{}, false
}

func mustFailAbsent(t *testing.T, a *NT8AbsentView, name, want string) {
	t.Helper()
	if a == nil {
		t.Fatalf("nt8_absent missing entirely")
	}
	if a.Ready {
		t.Fatalf("nt8_absent READY although leg %s must fail: %+v", name, a.Legs)
	}
	l, ok := absentLegOf(a, name)
	if !ok {
		t.Fatalf("absent leg %s missing: %+v", name, a.Legs)
	}
	if l.Pass {
		t.Fatalf("absent leg %s passed; want fail (%s)", name, l.Detail)
	}
	if want != "" && !strings.Contains(l.Detail, want) {
		t.Fatalf("absent leg %s detail %q does not say %q", name, l.Detail, want)
	}
}

func TestInstallationGateNt8AbsentReadyWhenEveryLegPasses(t *testing.T) {
	f, _ := newAbsentFixture(t)
	g := f.run()
	if g.NT8Absent == nil {
		t.Fatal("nt8_absent missing when an NT8 trader exists")
	}
	a := g.NT8Absent
	if !a.Eligible || !a.Ready {
		t.Fatalf("absent view must be eligible and ready: %+v", a)
	}
	if a.LinkDownSince == "" {
		t.Fatal("link_down_since must be recorded")
	}
	for _, name := range []string{"hold", "go_drained", "in_flight_sends", "queued_signals", "planner_in_flight", "ledger_exposure", "sim_accounts"} {
		l, ok := absentLegOf(a, name)
		if !ok {
			t.Errorf("absent leg %s missing", name)
			continue
		}
		if l.Source == "" {
			t.Errorf("absent leg %s quotes no source", name)
		}
		if !l.Pass {
			t.Errorf("absent leg %s failed: %s", name, l.Detail)
		}
	}
	// the NORMAL verdict stays not-ready: with the AddOn down there is no ack
	if g.Ready {
		t.Fatalf("the normal gate must not be READY with the AddOn disconnected")
	}
}

func TestInstallationGateNt8AbsentLinkDownBelowWindowNotEligible(t *testing.T) {
	f, _ := newAbsentFixture(t)
	f.wire.DisconnectedAt = time.Now().Add(-30 * time.Second)
	g := f.run()
	if g.NT8Absent == nil || g.NT8Absent.Eligible {
		t.Fatalf("link down 30s must not be eligible: %+v", g.NT8Absent)
	}
	if g.NT8Absent.Legs != nil {
		t.Fatalf("an uncomputed leg list must be ABSENT, not []: %+v", g.NT8Absent.Legs)
	}
}

func TestInstallationGateNt8AbsentConnectedIsNotEligible(t *testing.T) {
	f, _ := newAbsentFixture(t)
	// a reconnect REVOKES the verdict (the accept clears the stamp) — never
	// grandfathered
	f.wire.Connected = true
	f.wire.Rec.Ack = goodCensusAck("job-g")
	f.wire.HasAck = true
	f.wire.AckAge = time.Second
	g := f.run()
	if g.NT8Absent == nil || g.NT8Absent.Eligible {
		t.Fatalf("a connected AddOn must not be eligible: %+v", g.NT8Absent)
	}
}

func TestInstallationGateNt8AbsentNoDisconnectStampNotEligible(t *testing.T) {
	f, _ := newAbsentFixture(t)
	f.wire.HaveDisconnectedAt = false // never measured → fail closed
	g := f.run()
	if g.NT8Absent == nil || g.NT8Absent.Eligible {
		t.Fatalf("without the server's disconnect stamp the verdict must not be eligible: %+v", g.NT8Absent)
	}
}

func TestInstallationGateNt8AbsentEachLedgerLegRefuses(t *testing.T) {
	t.Run("no hold", func(t *testing.T) {
		f, _ := newAbsentFixture(t)
		if err := store.ClearMaintenanceHold(f.dir, "job-g"); err != nil {
			t.Fatal(err)
		}
		mustFailAbsent(t, f.run().NT8Absent, "hold", "no hold")
	})
	t.Run("an entry send holds a permit", func(t *testing.T) {
		f, _ := newAbsentFixture(t)
		if err := store.ClearMaintenanceHold(f.dir, "job-g"); err != nil {
			t.Fatal(err)
		}
		release, ok := MaintenanceEntryPermit()
		if !ok {
			t.Fatal("fixture: permit refused before the hold")
		}
		defer release()
		setHold(t, f.dir, "job-g")
		g := f.run()
		mustFailAbsent(t, g.NT8Absent, "go_drained", "permit")
		mustFailAbsent(t, g.NT8Absent, "in_flight_sends", "in_flight_sends=1")
	})
	t.Run("signals queued on the wire", func(t *testing.T) {
		f, _ := newAbsentFixture(t)
		f.wire.Queued = 3
		mustFailAbsent(t, f.run().NT8Absent, "queued_signals", "queued_signals=3")
	})
	t.Run("a planner read is in flight", func(t *testing.T) {
		f, _ := newAbsentFixture(t)
		plannerReadInFlight.Store("gate-t1", true)
		defer plannerReadInFlight.Delete("gate-t1")
		mustFailAbsent(t, f.run().NT8Absent, "planner_in_flight", "plannerReadInFlight[gate-t1]")
	})
	t.Run("a picture send is pending", func(t *testing.T) {
		f, _ := newAbsentFixture(t)
		if _, _, err := f.st.PictureHtfClaim(&store.PictureHtfOpportunityDB{OppKey: "opp-1", TraderID: "someone", Stage: "confirmed"}); err != nil {
			t.Fatal(err)
		}
		if won, err := f.st.PictureHtfClaimSubmission("opp-1", "picture-htf-1"); err != nil || !won {
			t.Fatal(err)
		}
		mustFailAbsent(t, f.run().NT8Absent, "ledger_exposure", "picture")
	})
	t.Run("bound account is not SIM", func(t *testing.T) {
		f, s := newAbsentFixture(t)
		s.SetAccountsList([]ntwire.AccountInfo{{Name: "Sim101", IsSim: false}}, "Sim101")
		mustFailAbsent(t, f.run().NT8Absent, "sim_accounts", "not SIM-tradeable")
	})
	t.Run("the server vouches for no account", func(t *testing.T) {
		f, s := newAbsentFixture(t)
		s.SetAccountsList(nil, "")
		mustFailAbsent(t, f.run().NT8Absent, "sim_accounts", "not SIM-tradeable")
	})
	t.Run("a trader bound to no account", func(t *testing.T) {
		f, s := newAbsentFixture(t)
		// rebind the fixture's trader with an empty bound account
		at := f.loaded["gate-t1"]
		at.trader = ntTrader.NewTCPTrader(s, "MNQ", "")
		mustFailAbsent(t, f.run().NT8Absent, "sim_accounts", "bound to no account")
	})
}
