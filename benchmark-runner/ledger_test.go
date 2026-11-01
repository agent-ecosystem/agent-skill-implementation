package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/agent-ecosystem/agent-skill-implementation/benchmark-runner/checks"
)

func entry(h, id, verdict, vehicle, conf, version, model string) LedgerEntry {
	return LedgerEntry{Harness: h, CheckID: id, Status: "observed", Verdict: verdict, Vehicle: vehicle, Confidence: conf, HarnessVersion: version, Model: model}
}

func side(entries ...LedgerEntry) map[string]map[string]LedgerEntry {
	out := map[string]map[string]LedgerEntry{}
	for _, e := range entries {
		if out[e.Harness] == nil {
			out[e.Harness] = map[string]LedgerEntry{}
		}
		out[e.Harness][e.CheckID] = e
	}
	return out
}

func TestDiffLedgersClassifies(t *testing.T) {
	base := side(
		entry("claude-code", "same", "a", "harness-push", "transcript-direct", "2.1.274", "m1"),
		entry("claude-code", "moved", "a", "harness-push", "transcript-direct", "2.1.274", "m1"),
		entry("claude-code", "graded", "a", "", "behavioral-inference", "2.1.274", "m1"),
		entry("claude-code", "dropped", "a", "harness-push", "transcript-direct", "2.1.274", "m1"),
		entry("codex", "untouched-harness", "a", "model-pull", "transcript-direct", "0.157.0", "m2"),
	)
	newer := side(
		entry("claude-code", "same", "a", "harness-push", "transcript-direct", "2.1.309", "m3"),
		entry("claude-code", "moved", "b", "harness-push", "transcript-direct", "2.1.309", "m3"),
		entry("claude-code", "graded", "a", "harness-push", "transcript-direct", "2.1.309", "m3"),
		entry("claude-code", "fresh", "a", "harness-push", "transcript-direct", "2.1.309", "m3"),
		LedgerEntry{Harness: "claude-code", CheckID: "errored", Status: "error", Verdict: "runner-error"},
	)
	deltas, unchanged, fresh, notRerun := diffLedgers(base, newer)
	if len(deltas) != 2 || deltas[0].CheckID != "graded" || deltas[1].CheckID != "moved" {
		t.Fatalf("deltas = %+v", deltas)
	}
	if deltas[0].verdictChanged() || !deltas[1].verdictChanged() {
		t.Fatalf("verdictChanged misclassified: %+v", deltas)
	}
	if len(unchanged) != 1 || unchanged[0].CheckID != "same" {
		t.Fatalf("unchanged = %+v", unchanged)
	}
	if len(fresh) != 1 || fresh[0].CheckID != "fresh" {
		t.Fatalf("fresh = %+v", fresh)
	}
	// "dropped" was not rerun; codex was not in the newer run at all and
	// must not be reported as not rerun.
	if len(notRerun) != 1 || notRerun[0].CheckID != "dropped" {
		t.Fatalf("notRerun = %+v", notRerun)
	}
}

func TestRenderLedgerDiffDraft(t *testing.T) {
	base := side(entry("claude-code", "name-directory-mismatch", "directory-name-identity", "harness-push", "transcript-direct", "2.1.274", "claude-sonnet-5"))
	newer := side(
		entry("claude-code", "name-directory-mismatch", "listed-under-both", "harness-push", "transcript-direct", "2.1.309", "claude-sonnet-5-5"),
		entry("copilot", "discovery-reading-depth", "metadata-only", "harness-push", "transcript-direct", "1.0.105", "claude-sonnet-5.5"),
	)
	out := renderLedgerDiff(base, newer)
	for _, want := range []string{
		"- Claude Code (headless): 2.1.309 (claude-sonnet-5-5), previously 2.1.274 (claude-sonnet-5)",
		"first run",
		"### Verdicts that moved (1 of 1 compared)",
		"- **Claude Code** `name-directory-mismatch`: `directory-name-identity` → `listed-under-both` (harness-push, transcript-direct)",
		"### First observed in this run (1)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("draft missing %q:\n%s", want, out)
		}
	}
}

func TestLoadLedgersLatestWins(t *testing.T) {
	dir := t.TempDir()
	write := func(name, ended, verdict string) string {
		path := dir + "/" + name
		data := `{"source":"` + name + `","generated":"2026-11-01","entries":[{"harness":"codex","check_id":"x","status":"observed","verdict":"` + verdict + `","ended":"` + ended + `"}]}`
		if err := writeFile(path, data); err != nil {
			t.Fatal(err)
		}
		return path
	}
	older := write("older.json", "2026-09-25T10:00:00Z", "old")
	newer := write("newer.json", "2026-10-31T10:00:00Z", "new")
	// Order of arguments must not matter; the later end time wins.
	got, err := loadLedgers([]string{newer, older})
	if err != nil {
		t.Fatal(err)
	}
	if got["codex"]["x"].Verdict != "new" {
		t.Fatalf("latest did not win: %+v", got["codex"]["x"])
	}
}

func writeFile(path, data string) error { return os.WriteFile(path, []byte(data), 0o644) }

func TestPriorVerdictsPicksLatestEarlierObservation(t *testing.T) {
	now := time.Date(2026, 10, 31, 23, 0, 0, 0, time.UTC)
	all := map[string]map[string]checks.Finding{
		"claude-code": {
			"flipped":   {Harness: "claude-code", CheckID: "flipped", Status: "observed", Verdict: "b", Ended: now},
			"flipflop":  {Harness: "claude-code", CheckID: "flipflop", Status: "observed", Verdict: "a", Ended: now},
			"steady":    {Harness: "claude-code", CheckID: "steady", Status: "observed", Verdict: "a", Ended: now},
			"first":     {Harness: "claude-code", CheckID: "first", Status: "observed", Verdict: "a", Ended: now},
			"errored":   {Harness: "claude-code", CheckID: "errored", Status: "error", Verdict: "runner-error"},
			"untimed":   {Harness: "claude-code", CheckID: "untimed", Status: "observed", Verdict: "b"},
			"offsetted": {Harness: "claude-code", CheckID: "offsetted", Status: "observed", Verdict: "b", Ended: now},
		},
	}
	at := func(s string) string { return s }
	entries := []LedgerEntry{
		// The current finding itself is in the newest ledger; it must not
		// be taken as its own predecessor.
		{Harness: "claude-code", CheckID: "flipped", Status: "observed", Verdict: "b", HarnessVersion: "2.1.309", Ended: at("2026-10-31T23:00:00Z")},
		{Harness: "claude-code", CheckID: "flipped", Status: "observed", Verdict: "a", HarnessVersion: "2.1.274", Ended: at("2026-09-26T10:00:00Z")},
		{Harness: "claude-code", CheckID: "flipped", Status: "observed", Verdict: "z", HarnessVersion: "2.1.212", Ended: at("2026-08-01T10:00:00Z")},
		// a -> b -> a: the previous observation is b, so this counts as changed.
		{Harness: "claude-code", CheckID: "flipflop", Status: "observed", Verdict: "b", HarnessVersion: "2.1.274", Ended: at("2026-09-26T10:00:00Z")},
		{Harness: "claude-code", CheckID: "flipflop", Status: "observed", Verdict: "a", HarnessVersion: "2.1.212", Ended: at("2026-08-01T10:00:00Z")},
		{Harness: "claude-code", CheckID: "steady", Status: "observed", Verdict: "a", HarnessVersion: "2.1.274", Ended: at("2026-09-26T10:00:00Z")},
		// An earlier error is not an observation.
		{Harness: "claude-code", CheckID: "steady", Status: "error", Verdict: "runner-error", Ended: at("2026-09-27T10:00:00Z")},
		{Harness: "claude-code", CheckID: "errored", Status: "observed", Verdict: "a", Ended: at("2026-09-26T10:00:00Z")},
		{Harness: "claude-code", CheckID: "untimed", Status: "observed", Verdict: "a", Ended: at("2026-09-26T10:00:00Z")},
		// Offsets must compare as instants: 19:30-04:00 is 23:30Z, after now.
		{Harness: "claude-code", CheckID: "offsetted", Status: "observed", Verdict: "a", HarnessVersion: "later", Ended: at("2026-10-31T19:30:00-04:00")},
		{Harness: "claude-code", CheckID: "offsetted", Status: "observed", Verdict: "c", HarnessVersion: "earlier", Ended: at("2026-10-31T18:30:00-04:00")},
		// Another harness never leaks in.
		{Harness: "codex", CheckID: "steady", Status: "observed", Verdict: "q", Ended: at("2026-09-26T10:00:00Z")},
	}
	got := priorVerdicts(all, entries)
	cc := got["claude-code"]
	if p := cc["flipped"]; p.Verdict != "a" || p.HarnessVersion != "2.1.274" {
		t.Errorf("flipped prior = %+v, want verdict a on 2.1.274", p)
	}
	if p := cc["flipflop"]; p.Verdict != "b" {
		t.Errorf("flipflop prior = %+v, want b", p)
	}
	for _, id := range []string{"steady", "first", "errored", "untimed"} {
		if _, ok := cc[id]; ok {
			t.Errorf("%s should carry no prior, got %+v", id, cc[id])
		}
	}
	if p := cc["offsetted"]; p.Verdict != "c" || p.HarnessVersion != "earlier" {
		t.Errorf("offsetted prior = %+v, want c on earlier (the -04:00 entry after now must be excluded)", p)
	}
	if shortDate(cc["flipped"].Ended) != "2026-09-26" {
		t.Errorf("shortDate = %q", shortDate(cc["flipped"].Ended))
	}
}
