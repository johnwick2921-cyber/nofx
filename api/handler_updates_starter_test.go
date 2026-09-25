package api

// W-ONE-BUTTON M4 3b-B U5b (c) — knob ON, install hands off over the worker
// socket, at the production call site: the gin install route through the
// real gate, the production verdict (updaterworker.FetchRelease), and a REAL
// wireserver listening on <data>/updater/worker.sock in the temp data dir.
// Accepted: OK with state "requested", or — a re-send — OK with the job's
// own state as its job file (updaterjob's production writer) records it for
// THIS release. Anything else, or no worker, is M3's 503.

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"

	"nofx/internal/updaterwire"
)

func TestInstallHandsOffOverTheSocket(t *testing.T) {
	const installUnavailable = `{"error":"installer unavailable"}`
	cases := []struct {
		name      string
		worker    bool
		answer    func(g0 string) (bool, string) // g0: the grant's job id
		jobFile   string                         // "" none; else the release id the job file names
		wantCode  int
		wantCalls int
	}{
		{"worker accepts: requested", true, func(string) (bool, string) { return true, "requested" }, "", http.StatusAccepted, 1},
		{"worker refuses", true, func(string) (bool, string) { return false, "busy" }, "", http.StatusServiceUnavailable, 1},
		{"worker says another state, no job file", true, func(string) (bool, string) { return true, "downloaded" }, "", http.StatusServiceUnavailable, 1},
		{"re-send: the job's own state, same release", true, func(string) (bool, string) { return true, "downloaded" }, updRelease, http.StatusAccepted, 1},
		{"re-send: a state the job file does not hold", true, func(string) (bool, string) { return true, "gate_ok" }, updRelease, http.StatusServiceUnavailable, 1},
		{"re-send: the job file names another release", true, func(string) (bool, string) { return true, "downloaded" }, "v2026.09.24-7", http.StatusServiceUnavailable, 1},
		{"no worker", false, nil, "", http.StatusServiceUnavailable, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(updaterKnobEnv, "1")
			e := newUpdEnv(t)
			fetchTestRelease(t, e.dataDir, updRelease)
			g := e.grant(updRelease)
			if c.jobFile != "" {
				writeTestJob(t, e.dataDir, g.JobID, c.jobFile)
			}
			var w *testWorker
			if c.worker {
				w = startTestWorker(t, e.dataDir, func(string, string) (bool, string) { return c.answer(g.JobID) })
			}
			resp := e.do("POST", "/api/updates/install", grantBody(g))
			if resp.Code != c.wantCode {
				t.Fatalf("install = %d %s, want %d", resp.Code, resp.Body.String(), c.wantCode)
			}
			switch c.wantCode {
			case http.StatusAccepted:
				var body map[string]any
				if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil || len(body) != 1 || body["job_id"] != g.JobID {
					t.Fatalf("202 body = %s, want exactly {\"job_id\":%q}", resp.Body.String(), g.JobID)
				}
			case http.StatusServiceUnavailable:
				if resp.Body.String() != installUnavailable {
					t.Fatalf("503 body = %s, want M3's %s", resp.Body.String(), installUnavailable)
				}
			}
			if w != nil {
				installs, other := w.calls()
				if len(installs) != c.wantCalls || len(other) != 0 {
					t.Fatalf("the worker saw installs=%v other verbs=%v, want %d install frame(s) and nothing else", installs, other, c.wantCalls)
				}
				if installs[0] != (updaterwire.InstallPayload{ReleaseID: updRelease, JobID: g.JobID}) {
					t.Fatalf("the worker got %+v, want exactly the grant's {release_id %s, job_id %s}", installs[0], updRelease, g.JobID)
				}
			}
		})
	}
}

// With the knob ON the status reports what the server now holds — the same
// three keys (the M3 key-set pin), install enabled because a real verifier
// and a starter are wired.
func TestUpdatesStatusWithTheKnobOn(t *testing.T) {
	t.Setenv(updaterKnobEnv, "1")
	e := newUpdEnv(t)
	w := e.do("GET", "/api/updates", "")
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if strings.Join(keys, ",") != "enrolled,install_enabled,manifest_verifier" {
		t.Fatalf("knob ON: GET /api/updates keys = %v", keys)
	}
	if w.Body.String() != `{"enrolled":true,"install_enabled":true,"manifest_verifier":"configured"}` {
		t.Fatalf("knob ON: GET /api/updates = %s", w.Body.String())
	}
}
