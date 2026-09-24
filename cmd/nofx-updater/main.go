// Command nofx-updater is the M4 updater worker (wave 3b-B) and its attended
// operator verbs. It is the ONLY place a resume request is built
// (internal/updaterwire/resume_census_test.go admits exactly this directory):
//
//	nofx-updater [--install-dir d] serve              run the worker (one job at a time)
//	nofx-updater [--install-dir d] fetch <release_id> materialize + verify a local release (attended)
//	nofx-updater [--install-dir d] status [<job>]     the worker's state, or one job's file
//	nofx-updater [--install-dir d] resume <job>       attended resume of a job parked at nt8_updated
//	nofx-updater [--install-dir d] recovery <job>     the manual steps for a recovery_needed job
//
// It never runs as root, never touches the hold (only the worker's census-
// admitted hold.go does), never mints a token (serve reads the operator's
// NOFX_CUTOVER_TOKEN from its environment) and never runs a step itself.
//
// L4: serve and fetch REFUSE until their production adapters land — the
// activation library adapter after #201 merges to dev, the release re-proof
// and fetch at the U3 fold — so today nothing here can change the
// installation (TestServeAndFetchRefuseUntilTheAdaptersLand).
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"nofx/internal/updaterjob"
	"nofx/internal/updaterwire"
	"nofx/internal/updaterworker"
)

// Seams (tests only).
var (
	geteuid      = os.Geteuid
	isTerminal   = stdinIsTerminal
	checkProcess = updaterworker.CheckProcess
	getwd        = os.Getwd
)

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

const usage = "usage: nofx-updater [--install-dir d] serve | fetch <release_id> | status [<job>] | resume <job> | recovery <job>"

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	top := flag.NewFlagSet("nofx-updater", flag.ContinueOnError)
	top.SetOutput(stderr)
	wd, _ := getwd()
	installDir := top.String("install-dir", wd, "the bot's WorkingDirectory (its .env and DB_PATH decide the data dir)")
	if err := top.Parse(args); err != nil {
		return 2
	}
	rest := top.Args()
	if len(rest) == 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	if geteuid() == 0 {
		fmt.Fprintln(stderr, "nofx-updater: refusing to run as root: run it as the bot's own user (a root-owned data/updater would lock the bot into a hold it cannot read)")
		return 2
	}
	verb, operands := rest[0], rest[1:]
	want := map[string][2]int{"serve": {0, 0}, "fetch": {1, 1}, "status": {0, 1}, "resume": {1, 1}, "recovery": {1, 1}}
	n, ok := want[verb]
	if !ok {
		fmt.Fprintf(stderr, "nofx-updater: unknown subcommand %q\n%s\n", verb, usage)
		return 2
	}
	if len(operands) < n[0] || len(operands) > n[1] {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	t, err := updaterworker.ResolveTarget(*installDir)
	if err != nil {
		fmt.Fprintln(stderr, "nofx-updater:", err)
		return 2
	}
	switch verb {
	case "serve":
		return serve(t, stderr)
	case "fetch":
		fmt.Fprintf(stderr, "nofx-updater fetch %s: not wired yet — the release fetch (U3's FetchRelease) lands at the U3 fold; refusing, nothing was written\n", operands[0])
		return 2
	case "status":
		if len(operands) == 0 {
			return statusWorker(t, stdout, stderr)
		}
		return statusJob(t, operands[0], stdout, stderr)
	case "resume":
		return resume(t, operands[0], stdin, stdout, stderr)
	case "recovery":
		return recovery(t, operands[0], stdout, stderr)
	}
	return 2
}

// serve refuses, in order: the process checks (root, the bot's cgroup, TZ),
// a missing cutover token, and — until they land — the two production
// adapters. It writes nothing before the last check passes.
func serve(t updaterworker.Target, stderr io.Writer) int {
	if err := checkProcess(); err != nil {
		fmt.Fprintln(stderr, "nofx-updater serve:", err)
		return 2
	}
	if _, err := updaterworker.NewHTTPApp(t.BaseURL()); err != nil {
		fmt.Fprintln(stderr, "nofx-updater serve:", err)
		return 2
	}
	fmt.Fprintf(stderr, "nofx-updater serve: %v — the activation library adapter lands after #201 merges to dev and the release re-proof adapter at the U3 fold; refusing to start (install dir %s, data dir %s); nothing was written\n",
		updaterworker.ErrNotWired, t.InstallDir, t.DataDir)
	return 2
}

// statusWorker asks the running worker (status with no job id).
func statusWorker(t updaterworker.Target, stdout, stderr io.Writer) int {
	c, err := updaterwire.DialWorker(t.DataDir)
	if err != nil {
		fmt.Fprintf(stderr, "nofx-updater status: no worker answers (%v)\n", err)
		return 1
	}
	defer c.Close()
	resp, err := c.Do(updaterwire.NewStatus(""))
	if err != nil {
		fmt.Fprintln(stderr, "nofx-updater status:", err)
		return 1
	}
	return printResponse(resp, stdout, stderr)
}

// statusJob reads one job's file (no worker needed): the API view plus the
// fields an operator needs, READ; absent ones print n/a.
func statusJob(t updaterworker.Target, jobID string, stdout, stderr io.Writer) int {
	j, err := updaterjob.Read(t.DataDir, jobID)
	if err != nil {
		fmt.Fprintf(stderr, "nofx-updater status %s: %v\n", jobID, err)
		return 1
	}
	v := updaterjob.View(j)
	out := map[string]any{"job_id": v.JobID, "release_id": j.ReleaseID, "state": v.State, "phase": j.Phase, "attempts": j.Attempts,
		"receipts": len(j.Receipts), "blocker": na(j.Blocker), "error": na(j.Error), "recovery_reason": na(j.RecoveryReason)}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Fprintln(stdout, string(b))
	return 0
}

// resume is the ATTENDED resume from nt8_updated: a terminal, the job id
// typed back, then the one resume frame the module ever builds.
func resume(t updaterworker.Target, jobID string, stdin io.Reader, stdout, stderr io.Writer) int {
	if !updaterwire.ValidJobID(jobID) {
		fmt.Fprintln(stderr, "nofx-updater resume: invalid job id")
		return 2
	}
	if !isTerminal(stdin) {
		fmt.Fprintln(stderr, "nofx-updater resume: refusing — a resume is attended: run it from a terminal and type the job id")
		return 2
	}
	j, err := updaterjob.Read(t.DataDir, jobID)
	if err != nil {
		fmt.Fprintf(stderr, "nofx-updater resume %s: %v\n", jobID, err)
		return 1
	}
	fmt.Fprintf(stdout, "job %s (release %s) is %s/%s.\nblocker: %s\n", j.JobID, j.ReleaseID, j.State, j.Phase, na(j.Blocker))
	fmt.Fprint(stdout, "Only after the AddOn is compiled (copy → F5 → full NT8 restart): type the job id to resume: ")
	line, err := bufio.NewReader(io.LimitReader(stdin, 256)).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintln(stderr, "nofx-updater resume:", err)
		return 2
	}
	if strings.TrimSpace(line) != jobID {
		fmt.Fprintln(stderr, "nofx-updater resume: the typed id does not match; nothing was sent")
		return 2
	}
	c, err := updaterwire.DialWorker(t.DataDir)
	if err != nil {
		fmt.Fprintf(stderr, "nofx-updater resume: no worker answers (%v)\n", err)
		return 1
	}
	defer c.Close()
	resp, err := c.Do(updaterwire.NewResume(jobID))
	if err != nil {
		fmt.Fprintln(stderr, "nofx-updater resume:", err)
		return 1
	}
	return printResponse(resp, stdout, stderr)
}

// recovery prints the job-specific manual steps (C16).
func recovery(t updaterworker.Target, jobID string, stdout, stderr io.Writer) int {
	j, err := updaterjob.Read(t.DataDir, jobID)
	if err != nil {
		fmt.Fprintf(stderr, "nofx-updater recovery %s: %v\n", jobID, err)
		return 1
	}
	fmt.Fprint(stdout, updaterworker.RecoveryText(j, t))
	return 0
}

func printResponse(resp updaterwire.Response, stdout, stderr io.Writer) int {
	if resp.OK {
		fmt.Fprintln(stdout, resp.State)
		return 0
	}
	fmt.Fprintln(stderr, "refused:", resp.Error)
	return 1
}

func na(s string) string {
	if s == "" {
		return "n/a"
	}
	return s
}
