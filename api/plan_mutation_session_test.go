package api

import (
	"encoding/json"
	"nofx/kernel"
	"testing"
	"time"
)

func TestPlanMutationSessionGapAndOvernightChain(t *testing.T) {
	s, st := askTestServer(t)
	reg := kernel.DefaultSessionRegistry()
	for i := range reg.Sessions {
		if reg.Sessions[i].Name == "ASIA" {
			reg.Sessions[i].Enabled = true
			reg.Sessions[i].WindowStartCT = "17:00"
			reg.Sessions[i].WindowEndCT = "02:00"
			reg.Sessions[i].ReadCT = "16:30"
			reg.Sessions[i].FlatCT = "02:00"
		}
	}
	raw, err := json.Marshal(reg)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetSystemConfig(kernel.SessionRegistryConfigKey, string(raw)); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 14, 0, 30, 0, 0, kernel.CTLocation())
	session, date, ok := s.planMutationSessionAt("t", now)
	if !ok || session.Name != "ASIA" || date != "2026-09-13" {
		t.Fatalf("overnight chain: session=%+v date=%q ok=%v", session, date, ok)
	}
	session, date, ok = s.planMutationSessionAt("t", time.Date(2026, 9, 14, 16, 0, 0, 0, kernel.CTLocation()))
	if ok || session != nil || date != "" {
		t.Fatalf("session gap must refuse: session=%+v date=%q ok=%v", session, date, ok)
	}
}
