package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/agent-ecosystem/agent-skill-implementation/benchmark-runner/checks"
)

// A ledger is the committable record of one run of the checks: every finding's
// verdict with the harness release and model that produced it, and
// nothing else. Results directories hold the archived transcripts and
// stay out of version control; ledgers are small enough to commit, so
// the history of what each platform did, release by release, is public
// and the changelog can be checked against it.
type Ledger struct {
	// Source names the results directory the ledger was taken from.
	Source string `json:"source"`
	// Generated is the date the ledger file was written.
	Generated string        `json:"generated"`
	Entries   []LedgerEntry `json:"entries"`
}

// LedgerEntry is one finding reduced to its verdict and provenance.
type LedgerEntry struct {
	Harness        string         `json:"harness"`
	HarnessVersion string         `json:"harness_version,omitempty"`
	Model          string         `json:"model,omitempty"`
	CheckID        string         `json:"check_id"`
	Status         string         `json:"status"`
	Verdict        string         `json:"verdict"`
	Vehicle        string         `json:"vehicle,omitempty"`
	Confidence     string         `json:"confidence,omitempty"`
	Runs           int            `json:"runs,omitempty"`
	VerdictCounts  map[string]int `json:"verdict_counts,omitempty"`
	// Ended is the finding's end time (RFC 3339, nanosecond precision so
	// it round-trips the finding exactly); when several ledgers cover the
	// same harness and check, the latest entry wins.
	Ended string `json:"ended,omitempty"`
}

func entryFromFinding(f checks.Finding) LedgerEntry {
	e := LedgerEntry{
		Harness:        f.Harness,
		HarnessVersion: f.HarnessVersion,
		Model:          f.Model,
		CheckID:        f.CheckID,
		Status:         f.Status,
		Verdict:        f.Verdict,
		Vehicle:        f.Vehicle,
		Confidence:     f.Confidence,
		Runs:           f.Runs,
		VerdictCounts:  f.VerdictCounts,
	}
	if !f.Ended.IsZero() {
		e.Ended = f.Ended.Format(time.RFC3339Nano)
	}
	return e
}

// writeLedgers writes one ledger per findings directory into outDir,
// named after the directory (results/resweep-2026-10-31 becomes
// resweep-2026-10-31.json).
func writeLedgers(findingsDirs []string, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for _, dir := range findingsDirs {
		all, err := loadFindings([]string{dir})
		if err != nil {
			return err
		}
		l := Ledger{
			Source:    filepath.Base(filepath.Clean(dir)),
			Generated: time.Now().Format("2006-01-02"),
		}
		for _, fs := range all {
			for _, f := range fs {
				l.Entries = append(l.Entries, entryFromFinding(f))
			}
		}
		if len(l.Entries) == 0 {
			return fmt.Errorf("ledger: no findings under %s", dir)
		}
		sort.Slice(l.Entries, func(i, j int) bool {
			a, b := l.Entries[i], l.Entries[j]
			if a.Harness != b.Harness {
				return a.Harness < b.Harness
			}
			return a.CheckID < b.CheckID
		})
		data, err := json.MarshalIndent(l, "", "  ")
		if err != nil {
			return err
		}
		path := filepath.Join(outDir, l.Source+".json")
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s (%d entries)\n", path, len(l.Entries))
	}
	return nil
}

// loadLedgers reads the given ledger files and returns the latest entry
// per (harness, check), so a side of a diff can be assembled from
// several partial runs the way the reports are.
func loadLedgers(paths []string) (map[string]map[string]LedgerEntry, error) {
	out := map[string]map[string]LedgerEntry{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("ledger: %w", err)
		}
		var l Ledger
		if err := json.Unmarshal(data, &l); err != nil {
			return nil, fmt.Errorf("ledger: %s: %w", path, err)
		}
		for _, e := range l.Entries {
			if out[e.Harness] == nil {
				out[e.Harness] = map[string]LedgerEntry{}
			}
			if prev, ok := out[e.Harness][e.CheckID]; !ok || e.Ended > prev.Ended {
				out[e.Harness][e.CheckID] = e
			}
		}
	}
	return out, nil
}

// ledgerDelta is one (harness, check) whose verdict, vehicle, or
// confidence differs between a baseline and a newer run.
type ledgerDelta struct {
	Harness string
	CheckID string
	Old     LedgerEntry
	New     LedgerEntry
}

// verdictChanged reports whether the delta moved the verdict itself, as
// opposed to only its vehicle or confidence.
func (d ledgerDelta) verdictChanged() bool { return d.Old.Verdict != d.New.Verdict }

// diffLedgers compares a newer side against a baseline. Only entries the
// newer side observed are judged; a baseline entry the new run did not
// rerun is listed, not counted as a change.
func diffLedgers(baseline, newer map[string]map[string]LedgerEntry) (deltas []ledgerDelta, unchanged, fresh, notRerun []ledgerDelta) {
	for h, fs := range newer {
		for id, n := range fs {
			if n.Status != checks.StatusObserved {
				continue
			}
			o, ok := baseline[h][id]
			d := ledgerDelta{Harness: h, CheckID: id, Old: o, New: n}
			switch {
			case !ok:
				fresh = append(fresh, d)
			case o.Verdict != n.Verdict || o.Vehicle != n.Vehicle || o.Confidence != n.Confidence:
				deltas = append(deltas, d)
			default:
				unchanged = append(unchanged, d)
			}
		}
	}
	for h, fs := range baseline {
		for id, o := range fs {
			if _, ok := newer[h]; !ok {
				continue // harness not in the newer run at all
			}
			if _, ok := newer[h][id]; !ok {
				notRerun = append(notRerun, ledgerDelta{Harness: h, CheckID: id, Old: o})
			}
		}
	}
	for _, s := range [][]ledgerDelta{deltas, unchanged, fresh, notRerun} {
		sort.Slice(s, func(i, j int) bool {
			if s[i].Harness != s[j].Harness {
				return s[i].Harness < s[j].Harness
			}
			return s[i].CheckID < s[j].CheckID
		})
	}
	return deltas, unchanged, fresh, notRerun
}

// renderLedgerDiff writes the diff as a markdown draft for a changelog
// entry. It sorts the facts; deciding which verdict moves are
// platform-level and which are the model's choice stays with the
// maintainer, who has the transcripts.
func renderLedgerDiff(baseline, newer map[string]map[string]LedgerEntry) string {
	deltas, unchanged, fresh, notRerun := diffLedgers(baseline, newer)
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }

	var harnesses []string
	for h := range newer {
		harnesses = append(harnesses, h)
	}
	sort.Strings(harnesses)

	w("### Releases tested")
	w("")
	for _, h := range harnesses {
		name := platformNames[h]
		if name == "" {
			name = h
		}
		nv, nm := releaseSummary(newer[h])
		if old, ok := baseline[h]; ok {
			ov, om := releaseSummary(old)
			w("- %s: %s (%s), previously %s (%s)", name, nv, nm, ov, om)
		} else {
			w("- %s: %s (%s), first run", name, nv, nm)
		}
	}
	w("")

	var moved, evidenceOnly []ledgerDelta
	for _, d := range deltas {
		if d.verdictChanged() {
			moved = append(moved, d)
		} else {
			evidenceOnly = append(evidenceOnly, d)
		}
	}
	w("### Verdicts that moved (%d of %d compared)", len(moved), len(moved)+len(evidenceOnly)+len(unchanged))
	w("")
	if len(moved) == 0 {
		w("None.")
		w("")
	}
	for _, d := range moved {
		runs := ""
		if d.New.Runs > 1 {
			runs = fmt.Sprintf(", %d/%d runs", d.New.VerdictCounts[d.New.Verdict], d.New.Runs)
		}
		w("- **%s** `%s`: `%s` → `%s` (%s, %s%s)", shortName(d.Harness), d.CheckID, d.Old.Verdict, d.New.Verdict, orDash(d.New.Vehicle), orDash(d.New.Confidence), runs)
	}
	if len(moved) > 0 {
		w("")
	}
	if len(evidenceOnly) > 0 {
		w("### Same verdict, different evidence grade (%d)", len(evidenceOnly))
		w("")
		for _, d := range evidenceOnly {
			w("- **%s** `%s`: `%s` stays; vehicle %s → %s, confidence %s → %s", shortName(d.Harness), d.CheckID, d.New.Verdict, orDash(d.Old.Vehicle), orDash(d.New.Vehicle), orDash(d.Old.Confidence), orDash(d.New.Confidence))
		}
		w("")
	}
	if len(fresh) > 0 {
		w("### First observed in this run (%d)", len(fresh))
		w("")
		for _, d := range fresh {
			w("- **%s** `%s`: `%s`", shortName(d.Harness), d.CheckID, d.New.Verdict)
		}
		w("")
	}
	if len(notRerun) > 0 {
		w("### In the baseline but not rerun (%d)", len(notRerun))
		w("")
		// A partial run leaves most of the baseline untouched; past a
		// handful, a per-harness count says it better than a list.
		if len(notRerun) > 20 {
			perHarness := map[string]int{}
			for _, d := range notRerun {
				perHarness[d.Harness]++
			}
			for _, h := range harnesses {
				if n := perHarness[h]; n > 0 {
					w("- **%s**: %d earlier verdicts stand", shortName(h), n)
				}
			}
		} else {
			for _, d := range notRerun {
				w("- **%s** `%s`: still `%s` from %s", shortName(d.Harness), d.CheckID, d.Old.Verdict, orDash(d.Old.HarnessVersion))
			}
		}
		w("")
	}
	return b.String()
}

// releaseSummary lists the harness versions and models a side observed.
func releaseSummary(fs map[string]LedgerEntry) (versions, models string) {
	vs, ms := map[string]bool{}, map[string]bool{}
	for _, e := range fs {
		if e.HarnessVersion != "" {
			vs[e.HarnessVersion] = true
		}
		if e.Model != "" {
			ms[e.Model] = true
		}
	}
	return joinSorted(vs), joinSorted(ms)
}

func joinSorted(set map[string]bool) string {
	var out []string
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return "-"
	}
	return strings.Join(out, ", ")
}

func shortName(h string) string {
	if n := shortNames[h]; n != "" {
		return n
	}
	return h
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// loadLedgerDir reads every ledger file under dir and returns their
// entries, so the report generator can see each platform's history
// without the results directories.
func loadLedgerDir(dir string) ([]LedgerEntry, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("ledger: no ledger files under %s", dir)
	}
	var out []LedgerEntry
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("ledger: %w", err)
		}
		var l Ledger
		if err := json.Unmarshal(data, &l); err != nil {
			return nil, fmt.Errorf("ledger: %s: %w", path, err)
		}
		out = append(out, l.Entries...)
	}
	return out, nil
}

// priorVerdicts finds, for each finding being reported, the most recent
// earlier observation of the same harness and check, and keeps it when
// its verdict differs from the current one. The result is what the
// reports mark as "changed since": the previous observation, whatever
// release it came from, not the whole history.
func priorVerdicts(all map[string]map[string]checks.Finding, entries []LedgerEntry) map[string]map[string]LedgerEntry {
	out := map[string]map[string]LedgerEntry{}
	for h, fs := range all {
		for id, f := range fs {
			if f.Status != checks.StatusObserved || f.Ended.IsZero() {
				continue
			}
			// The current finding is itself in the newest ledger. Cut off
			// at whole seconds so a ledger written at second precision
			// still excludes it rather than reporting it as its own
			// unchanged predecessor.
			cutoff := f.Ended.Truncate(time.Second)
			var prev LedgerEntry
			var prevAt time.Time
			found := false
			for _, e := range entries {
				if e.Harness != h || e.CheckID != id || e.Status != checks.StatusObserved || e.Ended == "" {
					continue
				}
				// RFC 3339 strings with offsets do not order as
				// instants; parse before comparing.
				et, err := time.Parse(time.RFC3339Nano, e.Ended)
				if err != nil || !et.Before(cutoff) {
					continue
				}
				if !found || et.After(prevAt) {
					prev, prevAt, found = e, et, true
				}
			}
			if found && prev.Verdict != f.Verdict {
				if out[h] == nil {
					out[h] = map[string]LedgerEntry{}
				}
				out[h][id] = prev
			}
		}
	}
	return out
}

// shortDate renders a ledger entry's end time as a date.
func shortDate(rfc3339 string) string {
	if t, err := time.Parse(time.RFC3339Nano, rfc3339); err == nil {
		return t.Format("2006-01-02")
	}
	return rfc3339
}
