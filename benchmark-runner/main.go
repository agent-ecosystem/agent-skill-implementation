// Command benchmark-runner executes this repository's automated
// skill-loading checks against installed harnesses, using skillxp as the
// invocation and observation engine. Findings and archived transcripts
// land under -out, one directory per harness and check.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agent-ecosystem/agent-skill-implementation/benchmark-runner/checks"
	"github.com/agent-ecosystem/agentsummons"
	"github.com/agent-ecosystem/skillxp/observe"
	"github.com/agent-ecosystem/skillxp/profile"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "benchmark-runner:", err)
		os.Exit(1)
	}
}

func run() error {
	skills := flag.String("skills", defaultSkillsDir(), "path to benchmark-skills")
	harnesses := flag.String("harness", "", "comma-separated harness ids (default: all)")
	checkIDs := flag.String("check", "", "comma-separated check ids (default: all)")
	out := flag.String("out", "results", "output directory")
	timeout := flag.Duration("timeout", 5*time.Minute, "per-invocation timeout")
	runs := flag.Int("runs", 1, "repetitions per check; >1 reports verdict rates")
	sandbox := flag.Bool("sandbox", false, "run each check against an isolated home (skillxp sandbox seeds required)")
	list := flag.Bool("list", false, "list implemented checks and exit")
	report := flag.String("report", "", "comma-separated findings dirs; generate per-platform markdown reports into -out (default reports/) instead of running checks")
	site := flag.String("site", "", "with -report: site root; write the platforms content section (comparison index + per-platform pages) instead of maintainer reports")
	flag.Parse()

	if *report != "" {
		if *site != "" {
			return writeSite(strings.Split(*report, ","), *site)
		}
		dest := *out
		if dest == "results" {
			dest = "reports"
		}
		return writeReports(strings.Split(*report, ","), dest)
	}

	if *list {
		for _, s := range checks.Registry() {
			fmt.Printf("%-28s %s\n", s.ID, s.Description)
		}
		return nil
	}
	if _, err := os.Stat(*skills); err != nil {
		return fmt.Errorf("benchmark-skills: %w", err)
	}
	if *runs < 1 {
		return fmt.Errorf("-runs must be positive")
	}

	var ids []agentsummons.ID
	if *harnesses == "" {
		for _, p := range profile.Profiles() {
			ids = append(ids, p.Harness)
		}
	} else {
		for _, h := range strings.Split(*harnesses, ",") {
			id := agentsummons.ID(strings.TrimSpace(h))
			if _, err := profile.For(id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
	}

	var specs []checks.Spec
	if *checkIDs == "" {
		specs = checks.Registry()
	} else {
		for _, c := range strings.Split(*checkIDs, ",") {
			s, err := checks.For(strings.TrimSpace(c))
			if err != nil {
				return err
			}
			specs = append(specs, s)
		}
	}

	ctx := context.Background()
	exitErr := false
	for _, id := range ids {
		p, err := profile.For(id)
		if err != nil {
			return err
		}
		for _, spec := range specs {
			fmt.Fprintf(os.Stderr, "[%s] %s: running\n", id, spec.ID)
			f := runCheck(ctx, p, spec, *skills, *out, *timeout, *runs, *sandbox)
			if err := writeFinding(filepath.Join(*out, f.Harness, f.CheckID), f); err != nil {
				return err
			}
			vehicle := f.Vehicle
			if vehicle == "" {
				vehicle = "-"
			}
			rate := ""
			if f.Runs > 1 {
				rate = fmt.Sprintf(" %d/%d runs", f.VerdictCounts[f.Verdict], f.Runs)
			}
			fmt.Printf("%-14s %-28s %-12s %-26s vehicle=%-13s %s%s\n",
				f.Harness, f.CheckID, f.Status, f.Verdict, vehicle, f.Confidence, rate)
			if f.Status == checks.StatusError {
				exitErr = true
			}
		}
	}
	if exitErr {
		return fmt.Errorf("one or more checks did not complete cleanly")
	}
	return nil
}

// runCheck observes one check (once, or -runs times) and grades it. Runs
// and sessions are serialized by construction: skillxp's transcript
// attribution requires it.
func runCheck(ctx context.Context, p profile.Profile, spec checks.Spec, skillsDir, outDir string, timeout time.Duration, runs int, sandbox bool) checks.Finding {
	errFinding := func(notes ...string) checks.Finding {
		return checks.Finding{
			CheckID: spec.ID,
			Harness: string(p.Harness),
			Status:  checks.StatusError,
			Verdict: "runner-error",
			Notes:   notes,
		}
	}
	if spec.RequiresSandbox && !sandbox {
		return errFinding("this check installs at user scope and requires -sandbox")
	}

	// Compile the check's sessions into observe specs.
	var sessSpecs []observe.SessionSpec
	var prompts []string
	for _, cs := range spec.Sessions {
		ss := observe.SessionSpec{}
		for _, skill := range cs.Skills {
			ss.SkillDirs = append(ss.SkillDirs, filepath.Join(skillsDir, skill))
		}
		for _, skill := range cs.UserSkills {
			ss.UserSkillDirs = append(ss.UserSkillDirs, filepath.Join(skillsDir, skill))
		}
		for _, dir := range cs.ProjectDirs {
			ss.ProjectDirs = append(ss.ProjectDirs, filepath.Join(skillsDir, dir))
		}
		for _, dir := range cs.OverlayDirs {
			ss.OverlayDirs = append(ss.OverlayDirs, filepath.Join(skillsDir, dir))
		}
		for _, t := range cs.Turns {
			turn := observe.Turn{Prompt: t.Prompt(p), Activation: t.Activation}
			if before := t.Before; before != nil {
				turn.Before = func(projectDir string) error { return before(p, projectDir) }
			}
			ss.Turns = append(ss.Turns, turn)
			prompts = append(prompts, turn.Prompt)
		}
		sessSpecs = append(sessSpecs, ss)
	}
	baseDir := filepath.Join(outDir, string(p.Harness), spec.ID)

	// observeRep runs every session of the check once, archiving under
	// dir; a nil error means all sessions completed.
	observeRep := func(dir string) ([]*observe.SessionObservation, error) {
		var sos []*observe.SessionObservation
		for i, ss := range sessSpecs {
			cfg := observe.Config{Timeout: timeout, Sandbox: sandbox, ArchiveDir: dir}
			if len(sessSpecs) > 1 {
				cfg.ArchiveDir = filepath.Join(dir, fmt.Sprintf("session-%d", i+1))
			}
			so, err := observe.ObserveSession(ctx, cfg, p.Harness, ss)
			if err != nil {
				return nil, fmt.Errorf("session %d: %w", i+1, err)
			}
			sos = append(sos, so)
		}
		return sos, nil
	}

	if runs <= 1 {
		sos, err := observeRep(baseDir)
		if err != nil {
			return errFinding(err.Error())
		}
		return stamp(spec, sos, spec.Evaluate(sos), prompts)
	}

	// Repetitions: grade each rep, record its finding under its run dir,
	// then aggregate — modal verdict up front, full distribution
	// alongside. A failed rep is recorded and the sequence continues.
	var perRun []checks.Finding
	var runErrs []string
	for r := 1; r <= runs; r++ {
		dir := filepath.Join(baseDir, fmt.Sprintf("run-%02d", r))
		sos, err := observeRep(dir)
		if err != nil {
			if ctx.Err() != nil {
				return errFinding(ctx.Err().Error())
			}
			runErrs = append(runErrs, fmt.Sprintf("run %d: %s", r, err.Error()))
			continue
		}
		f := stamp(spec, sos, spec.Evaluate(sos), prompts)
		if err := writeFinding(dir, f); err != nil {
			return errFinding("writing per-run finding: " + err.Error())
		}
		perRun = append(perRun, f)
	}
	if len(perRun) == 0 {
		f := errFinding("every repetition failed to execute")
		f.RunErrors = runErrs
		return f
	}
	counts := map[string]int{}
	modal := perRun[0].Verdict
	for _, f := range perRun {
		counts[f.Verdict]++
		if counts[f.Verdict] > counts[modal] {
			modal = f.Verdict
		}
	}
	var agg checks.Finding
	for _, f := range perRun {
		if f.Verdict == modal {
			agg = f
			break
		}
	}
	agg.Runs = runs
	agg.VerdictCounts = counts
	agg.RunErrors = runErrs
	switch {
	case len(counts) == 1 && len(runErrs) == 0:
		// Variation proves model-level behavior; consistency only
		// suggests platform-level, so the note claims no more than that.
		agg.Notes = append(agg.Notes, fmt.Sprintf("verdict consistent across %d runs", len(perRun)))
	case len(counts) > 1:
		agg.Status = checks.StatusObserved
		agg.Notes = append(agg.Notes, fmt.Sprintf("verdicts varied across %d runs (model-level signal); see verdict_counts and the per-run findings", len(perRun)))
	}
	agg.Started = perRun[0].Started
	agg.Ended = perRun[len(perRun)-1].Ended
	return agg
}

// stamp fills a finding's identity fields from a check's completed
// sessions.
func stamp(spec checks.Spec, sos []*observe.SessionObservation, f checks.Finding, prompts []string) checks.Finding {
	first, last := sos[0], sos[len(sos)-1]
	f.CheckID = spec.ID
	f.Harness = last.Final().Harness
	f.HarnessVersion = last.Final().HarnessVersion
	f.Model = last.Final().Model
	var ids []string
	for _, so := range sos {
		ids = append(ids, so.SessionID)
	}
	f.SessionID = strings.Join(ids, ",")
	f.PromptUsed = strings.Join(prompts, " || ")
	f.TranscriptPath = last.Final().TranscriptPath
	f.Started, f.Ended = first.Turns[0].Started, last.Final().Ended
	if f.Status == "" {
		f.Status = checks.StatusObserved
	}
	return f
}

func writeFinding(dir string, f checks.Finding) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "finding.json"), append(data, '\n'), 0o644)
}

// defaultSkillsDir points at ../benchmark-skills when the runner is
// invoked from its own directory, falling back to ./benchmark-skills for
// repo-root invocations.
func defaultSkillsDir() string {
	for _, c := range []string{"../benchmark-skills", "benchmark-skills"} {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "../benchmark-skills"
}
