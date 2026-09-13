package store

import "testing"

func TestBrokerTerminalArmCannotReauthorizeSameVersion(t *testing.T) {
	for _, state := range ArmStateNames() {
		if !IsTerminalArmState(state) {
			continue
		}
		t.Run(state, func(t *testing.T) {
			db := newArmedTestDB(t)
			st := NewArmedOrderStore(db)
			row := &ArmedOrderDB{TraderID: "t1", PlanID: "P", Version: 2, Session: "NY", Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 90, TargetPx: 120}
			if err := st.UpsertArm(row); err != nil {
				t.Fatal(err)
			}
			if err := st.SetSignal(row.ID, "broker-entry"); err != nil {
				t.Fatal(err)
			}
			if err := st.SetState(row.ID, state, "owner/completed"); err != nil {
				t.Fatal(err)
			}
			retry := &ArmedOrderDB{TraderID: "t1", PlanID: "P", Version: 2, Session: "NY", Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 90, TargetPx: 120}
			if err := st.UpsertArm(retry); err != nil {
				t.Fatal(err)
			}
			live, err := st.ListNonTerminal("t1")
			if err != nil {
				t.Fatal(err)
			}
			if len(live) != 0 {
				t.Fatalf("same version revived broker terminal order: %+v", live)
			}
			var count int64
			db.Model(&ArmedOrderDB{}).Count(&count)
			if count != 1 {
				t.Fatalf("minted successor without new authorization: %d rows", count)
			}
			retry.Version = 3
			if err := st.UpsertArm(retry); err != nil {
				t.Fatal(err)
			}
			live, err = st.ListNonTerminal("t1")
			if err != nil || len(live) != 1 || live[0].SignalID != "" || live[0].Version != 3 {
				t.Fatalf("new version must authorize fresh placement: %+v err=%v", live, err)
			}
			if live[0].BootID != ProcessBootID() || live[0].ArmedUnderVersion != 3 {
				t.Fatalf("successor lost authorization provenance: boot=%q armed_under=%d", live[0].BootID, live[0].ArmedUnderVersion)
			}
		})
	}
}
