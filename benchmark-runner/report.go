// Report generation: -report reads finding.json files from one or more
// results directories and emits one template-shaped markdown report per
// harness. Where the same (harness, check) appears in several directories,
// the finding with the latest end time wins.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/agent-ecosystem/agent-skill-implementation/benchmark-runner/checks"
)

// Model answers embedded in notes may contain markdown links to fixture
// files (observed on Antigravity: [SKILL.md](file:///var/folders/...)).
// The targets are ephemeral sandbox paths that mean nothing to readers and
// trip the site theme's link resolver, so site output keeps the link text
// and drops the URL. Truncated answers can also leave a dangling partial
// link, which the tail patterns clean up.
var (
	fileLinkPattern    = regexp.MustCompile(`\[([^\]]*)\]\(file://[^)]*\)`)
	fileURLPattern     = regexp.MustCompile(`file://[^\s)]*`)
	emptyLinkPattern   = regexp.MustCompile(`\[([^\]]*)\]\(\s*\)`)
	danglingTailattern = regexp.MustCompile(`\]\(\s*$`)
)

func sanitizeForSite(text string) string {
	text = fileLinkPattern.ReplaceAllString(text, "$1")
	text = fileURLPattern.ReplaceAllString(text, "")
	text = emptyLinkPattern.ReplaceAllString(text, "$1")
	text = danglingTailattern.ReplaceAllString(text, "]")
	return text
}

// reportCategories mirrors checks.md's category order, including
// the two manual-only checks so reports account for every check.
var reportCategories = []struct {
	Name   string
	Checks []string
}{
	{"Loading Timing", []string{"discovery-reading-depth", "activation-loading-scope", "eager-link-resolution"}},
	{"Directory Recognition", []string{"recognized-directory-set", "directory-naming-divergence", "unrecognized-directory-handling"}},
	{"Resource Access Patterns", []string{"resource-enumeration-behavior", "path-resolution-base", "cross-skill-resource-shadowing", "path-traversal-boundary", "resource-nesting-depth", "bundled-script-execution"}},
	{"Content Presentation", []string{"discovery-listing-fields", "frontmatter-handling", "content-wrapping-format"}},
	{"Lifecycle Management", []string{"reactivation-deduplication", "reactivation-freshness", "context-compaction-protection"}},
	{"Access Control", []string{"trust-gating-behavior", "compatibility-field-behavior", "allowed-tools-behavior"}},
	{"Skill-to-Skill Invocation", []string{"cross-skill-invocation", "invocation-depth-limit", "circular-invocation-handling", "invocation-language-sensitivity"}},
	{"Skill Dependencies", []string{"informal-dependency-resolution", "missing-dependency-behavior", "nonstandard-dependency-fields", "cross-scope-dependency"}},
	{"Discovery Scope", []string{"cross-client-directory-interop", "recursive-root-discovery", "nested-skill-discovery", "name-collision-precedence"}},
	{"Validation Strictness", []string{"malformed-yaml-tolerance", "missing-description-handling", "invalid-name-tolerance", "name-directory-mismatch", "metadata-value-edge-cases", "oversize-description-handling", "oversize-compatibility-handling"}},
}

// manualChecks require interactive sessions the runner cannot drive.
var manualChecks = map[string]bool{
	"context-compaction-protection": true,
	"trust-gating-behavior":         true,
}

var platformNames = map[string]string{
	"antigravity": "Antigravity CLI (headless)",
	"claude-code": "Claude Code (headless)",
	"codex":       "Codex CLI (headless)",
}

// fallbackFor derives the template's fallback-behavior line from what the
// automated run incidentally demonstrated. Several verdicts and notes ARE
// fallback observations (a model recovering from a failed path, content
// reachable by file read despite catalog absence); those surface here.
// Everything else is marked not-exercised — untested rather than absent —
// because single automated sessions don't probe recovery paths.
func fallbackFor(f checks.Finding) string {
	var seen []string
	switch {
	case f.Verdict == "cwd-base-model-requalified":
		seen = append(seen, "agent self-recovered in-run: after the bare relative path failed, the model requalified it against the skill directory without user intervention.")
	case strings.HasPrefix(f.Verdict, "unlisted-but-reachable"):
		seen = append(seen, "content remained reachable by direct file read even though the catalog omits the skill.")
	case strings.HasPrefix(f.Verdict, "catalog-lists-both-"):
		seen = append(seen, "the shadowed variant stays reachable: the catalog exposes both entries, so a user (or the model) can address either by path.")
	}
	for _, n := range f.Notes {
		l := strings.ToLower(n)
		// Positive recovery signals only; phrases like "never reached the
		// model" are findings about absence, not fallbacks.
		if strings.Contains(l, "file access") || strings.Contains(l, "own file read") || strings.Contains(l, "own raw file read") || strings.Contains(l, "via the model's own read") || strings.Contains(l, "model explored") || strings.Contains(l, "requalif") {
			seen = append(seen, "Observed in-run: "+n)
		}
	}
	if len(seen) > 0 {
		return strings.Join(seen, " ")
	}
	return "Not exercised: automated single-session runs do not probe recovery paths (no follow-up prompting). Treat as untested rather than absent."
}

// checkQuestions maps check IDs to their human question, sourced from the
// runner's spec descriptions plus the two manual-only checks.
func checkQuestions() map[string]string {
	out := map[string]string{
		"context-compaction-protection": "Is skill content protected when the context window fills up?",
		"trust-gating-behavior":         "Do project-level skills require trust approval before loading?",
	}
	for _, spec := range checks.Registry() {
		out[spec.ID] = spec.Description
	}
	return out
}

// categoryIntros give each comparison table one sentence of orientation,
// condensed from checks.md.
var categoryIntros = map[string]string{
	"Loading Timing":            "When skill content enters the model's context, and how much loads at each stage.",
	"Directory Recognition":     "Which directories a platform treats as part of a skill, and what happens to ones it doesn't recognize.",
	"Resource Access Patterns":  "How supporting files (scripts, references, assets) become available to the model.",
	"Content Presentation":      "What the model actually sees, at discovery and at activation, and how it's formatted.",
	"Lifecycle Management":      "How skill content is managed over the course of a conversation.",
	"Access Control":            "How platforms gate skill loading and handle control-related frontmatter.",
	"Skill-to-Skill Invocation": "Whether one skill's instructions can activate another skill.",
	"Skill Dependencies":        "What happens when one skill depends on another, formally or in prose.",
	"Discovery Scope":           "Where platforms look for skills: which directories are scanned, how deep, and what happens when two discovered skills claim the same name.",
	"Validation Strictness":     "How strictly platforms judge skills that break the spec's format rules: rejected, repaired, or loaded anyway.",
}

// verdictPhrases translate runner verdict slugs into plain language for
// the comparison tables. Slugs without an entry fall back to code format;
// humanVerdict handles composites and parameterized slugs.
var verdictPhrases = map[string]string{
	"metadata-only":                      "Metadata only",
	"name-and-description-only":          "Name and description only",
	"names-surfaced-description-unconfirmed": "Names surfaced; description unconfirmed",
	"listing-not-observed":               "Listing not observed",
	"full-body-at-discovery":             "Full body loaded at discovery",
	"body-only":                          "Body only",
	"no-prefetch":                        "No pre-fetching",
	"eager-link-prefetch":                "Linked files pre-fetched",
	"bulk-directory-load":                "Whole directory loaded",
	"no-enumeration-at-activation":       "Nothing enumerated",
	"no-enumeration":                     "Nothing enumerated",
	"listing-enumerated-without-contents": "File names listed, contents not loaded",
	"model-enumerated-on-demand":         "Model listed the files itself",
	"contents-loaded-at-activation":      "Contents loaded at activation",
	"resources-untouched":                "Not surfaced; model never looked",
	"resources-readable-on-demand":       "Readable when the model looks",
	"resources-content-injected":         "Contents injected at activation",
	"resources-enumerated-not-loaded":    "Names listed, contents not loaded",
	"untouched":                          "Not surfaced; model never looked",
	"cwd-base-model-requalified":         "Bare path fails; model recovers",
	"model-preemptively-qualified":       "Model used full paths (base untested)",
	"bare-relative-fails-no-recovery":    "Bare path fails; no recovery",
	"bare-relative-resolves-to-skill-dir": "Bare path resolves to the skill directory",
	"files-not-read":                     "Files never read",
	"own-resource-first":                 "Got its own file",
	"shadowed-by-sibling":                "Got the other skill's file",
	"resource-not-read":                  "File never read",
	"outside-skill-read-allowed":         "Reads outside the skill allowed",
	"traversal-blocked-visibly":          "Blocked with an error",
	"not-attempted":                      "Model never tried",
	"frontmatter-stripped-on-injection":  "Stripped before injection",
	"frontmatter-passed-through":         "Passed through to the model",
	"frontmatter-visible-via-raw-read":   "Visible (model reads the raw file)",
	"loaded-despite-edge-case-metadata":  "Loaded fine",
	"raw-injection":                      "Raw markdown, no wrapper tags",
	"wrapped-structured":                 "Wrapped in structured tags",
	"raw-file-via-pull":                  "Raw file via model read",
	"reinjected-each-activation":         "Full content re-injected every time",
	"deduplicated":                       "Deduplicated by the platform",
	"re-read-each-activation":            "Model re-reads each time",
	"not-re-read":                        "Model reused its memory",
	"fresh-content-served":               "Edits picked up immediately",
	"stale-content-served":               "Stale cached content served",
	"activated-no-gating":                "Loads normally, no gating",
	"not-listed-possible-gating":         "Not listed (possible gating)",
	"nested-skill-discovered":            "Discovered as a separate skill",
	"nested-skill-not-discovered":        "Not discovered",
	"all-depths-accessible-through-5":    "All depths reachable (tested to 5)",
	"no-resources-accessed":              "No resources accessed",
	"listed-under-both":                  "Listed under both names",
	"directory-name-identity":            "Directory name wins",
	"frontmatter-name-identity":          "Frontmatter name wins",
	"not-discovered":                     "Not discovered",
	"unlisted-but-reachable":             "Not cataloged; file still readable",
	"recursive-scan":                     "Scans the root recursively",
	"direct-children-only":               "Direct children only",
	"second-skill-loaded":                "Second skill activated",
	"attempted-visible-failure":          "Failed with a visible error",
	"attempted-not-loaded":               "Attempted; nothing loaded",
	"attempted-no-error":                 "Attempted; no error shown",
	"chain-completed-depth-3":            "Full three-skill chain completed",
	"cycle-stopped-model-choice":         "Model stopped the loop itself",
	"cycle-not-followed":                 "Model declined to start the loop",
	"reinvocation-blocked":               "Re-invocation blocked by the platform",
	"reinvocation-no-reload":             "Re-invoked without reloading",
	"reported-without-attempt":           "Reported missing without attempting",
	"silent-skip":                        "Silently skipped",
	"fields-ignored":                     "Ignored",
	"resolved-across-scopes":             "Resolved across scopes",
	"convention-scanned":                 "Convention path scanned",
	"convention-not-scanned":             "Convention path not scanned",
	"convention-is-native-dir":           "Convention path is the native directory",
	"tolerated-and-loaded":               "Tolerated and loaded",
	"skipped-strict-parser":              "Skipped by a strict parser",
	"loaded-despite-missing-description": "Loaded anyway",
	"skipped-as-guide-prescribes":        "Skipped (as the guide prescribes)",
	"listed-not-loaded":                  "Listed but would not load",
	"project-overrides-user":             "Project scope wins",
	"user-overrides-project":             "User scope wins",
	"both-variants-loaded":               "Both variants loaded (no shadowing)",
	"activation-not-observed":            "Activation not observed",
	"script-executed":                    "Script ran; output returned",
	"source-read-not-executed":           "Source read; never executed",
	"output-without-script-reference":    "Output appeared without a script call",
	"attempted-no-output":                "Attempted; no output arrived",
	"execution-blocked-visibly":          "Blocked with an error",
	"executed-regardless-of-field":       "Ran with and without the field",
	"field-enabled-execution":            "Ran only with the field",
	"field-blocked-execution":            "Ran only without the field",
	"all-invalid-names-tolerated":        "All three invalid names tolerated",
	"all-invalid-names-rejected":         "All three invalid names rejected",
	"loaded-despite-oversize-description":   "Loaded anyway",
	"skipped-oversize-description":          "Skipped",
	"loaded-despite-oversize-compatibility": "Loaded anyway",
	"skipped-oversize-compatibility":        "Skipped",
}

var missingTierPhrases = map[string]string{
	"visible-failure":          "fails visibly",
	"reported-without-attempt": "reported, not attempted",
	"attempted-no-error":       "attempted, no error",
	"silent-skip":              "silently skipped",
	"not-attempted":            "not attempted",
	"phantom-content":          "phantom content",
}

// humanVerdict renders a verdict slug in plain language, handling
// composite ("a; b") and parameterized ("prefix:detail") forms. Unmapped
// slugs fall back to code format so new verdicts degrade readably.
func humanVerdict(v string) string {
	parts := strings.Split(v, "; ")
	for i, part := range parts {
		parts[i] = humanVerdictPart(part)
	}
	return strings.Join(parts, "; ")
}

func humanVerdictPart(part string) string {
	if p, ok := verdictPhrases[part]; ok {
		return p
	}
	switch {
	case strings.HasPrefix(part, "missing:"):
		tier := strings.TrimPrefix(part, "missing:")
		if p, ok := missingTierPhrases[tier]; ok {
			return "missing dependency " + p
		}
	case strings.HasPrefix(part, "stray:"):
		if part == "stray:discovered" {
			return "stray file discovered"
		}
		return "stray file ignored"
	case strings.HasPrefix(part, "chain-stopped-after-"):
		return "Chain stopped after " + strings.TrimPrefix(part, "chain-stopped-after-")
	case strings.HasPrefix(part, "deepest-accessed-depth-"):
		return "Stopped at depth " + strings.TrimPrefix(part, "deepest-accessed-depth-")
	case strings.HasPrefix(part, "cycle-re-entered"):
		return "Loop ran (no platform guard)"
	case strings.HasPrefix(part, "catalog-lists-both-"):
		rest := strings.TrimPrefix(part, "catalog-lists-both-")
		switch rest {
		case "project-overrides-user":
			return "Lists both; model chose the project variant"
		case "user-overrides-project":
			return "Lists both; model chose the user variant"
		}
		return "Lists both variants"
	case strings.HasPrefix(part, "readable-on-demand:"):
		return "Readable when the model looks"
	case strings.HasPrefix(part, "injected-at-activation:"):
		return "Injected at activation"
	case strings.HasPrefix(part, "partial-enumeration:"):
		return "Partially enumerated"
	case strings.HasPrefix(part, "fields-acted-on:"):
		return "Acted on by the platform"
	case strings.HasPrefix(part, "body-plus-resources:"):
		return "Body plus resource files loaded"
	case strings.HasPrefix(part, "surfaces-beyond-description:"):
		rest := strings.Trim(strings.TrimPrefix(part, "surfaces-beyond-description:"), "[]")
		return "Also surfaces " + strings.ReplaceAll(rest, " ", ", ")
	case strings.HasPrefix(part, "invalid-names-tolerated:"):
		rest := strings.Trim(strings.TrimPrefix(part, "invalid-names-tolerated:"), "[]")
		return "Tolerated only: " + strings.ReplaceAll(rest, " ", ", ")
	case strings.HasPrefix(part, "with-field:"):
		return "With the field: " + strings.TrimPrefix(part, "with-field:")
	case strings.HasPrefix(part, "control:"):
		return "Control: " + strings.TrimPrefix(part, "control:")
	}
	return "`" + part + "`"
}

// loadFindings walks the given results directories and returns the latest
// finding per (harness, check).
func loadFindings(dirs []string) (map[string]map[string]checks.Finding, error) {
	out := map[string]map[string]checks.Finding{}
	for _, dir := range dirs {
		harnesses, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("report: %w", err)
		}
		for _, h := range harnesses {
			if !h.IsDir() {
				continue
			}
			checksDirs, err := os.ReadDir(filepath.Join(dir, h.Name()))
			if err != nil {
				continue
			}
			for _, c := range checksDirs {
				path := filepath.Join(dir, h.Name(), c.Name(), "finding.json")
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				var f checks.Finding
				if err := json.Unmarshal(data, &f); err != nil {
					return nil, fmt.Errorf("report: %s: %w", path, err)
				}
				if out[f.Harness] == nil {
					out[f.Harness] = map[string]checks.Finding{}
				}
				if prev, ok := out[f.Harness][f.CheckID]; !ok || f.Ended.After(prev.Ended) {
					out[f.Harness][f.CheckID] = f
				}
			}
		}
	}
	return out, nil
}

// writeReports emits one markdown report per harness into outDir.
func writeReports(findingsDirs []string, outDir string) error {
	all, err := loadFindings(findingsDirs)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return fmt.Errorf("report: no findings under %v", findingsDirs)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	var harnesses []string
	for h := range all {
		harnesses = append(harnesses, h)
	}
	sort.Strings(harnesses)
	for _, h := range harnesses {
		path := filepath.Join(outDir, h+".md")
		if err := os.WriteFile(path, []byte(renderReport(h, all[h], false)), 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s (%d findings)\n", path, len(all[h]))
	}
	return nil
}

// shortNames label comparison-table columns.
var shortNames = map[string]string{
	"antigravity": "Antigravity CLI",
	"claude-code": "Claude Code",
	"codex":       "Codex CLI",
}

// writeSite emits the site's platforms section: a comparison index plus
// one full report page per platform, with evidence reduced to prose (the
// site audience has no access to the archived transcripts).
func writeSite(findingsDirs []string, siteRoot string) error {
	all, err := loadFindings(findingsDirs)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return fmt.Errorf("report: no findings under %v", findingsDirs)
	}
	outDir := filepath.Join(siteRoot, "content", "platforms")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	var harnesses []string
	for h := range all {
		harnesses = append(harnesses, h)
	}
	sort.Strings(harnesses)
	if err := os.WriteFile(filepath.Join(outDir, "_index.md"), []byte(renderComparison(all, harnesses)), 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", filepath.Join(outDir, "_index.md"))
	for _, h := range harnesses {
		path := filepath.Join(outDir, h+".md")
		if err := os.WriteFile(path, []byte(renderReport(h, all[h], true)), 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s (%d findings)\n", path, len(all[h]))
	}
	return nil
}

// renderComparison builds the platforms index for a human audience: each
// row leads with the check's question, cells hold plain-language outcomes,
// and rows where platforms disagree are flagged with a 📌 marker.
func renderComparison(all map[string]map[string]checks.Finding, harnesses []string) string {
	questions := checkQuestions()
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	latest := ""
	for _, fs := range all {
		for _, f := range fs {
			if d := f.Ended.Format("2006-01-02"); d > latest {
				latest = d
			}
		}
	}

	// A row diverges when tested platforms returned different verdicts.
	diverges := func(id string) bool {
		if manualChecks[id] {
			return false
		}
		seen := map[string]bool{}
		for _, h := range harnesses {
			if f, ok := all[h][id]; ok {
				seen[f.Verdict] = true
			}
		}
		return len(seen) > 1
	}

	w("---")
	w(`title: "Platform Reports"`)
	w(`description: "Where agent platforms agree and diverge on skill loading, from automated transcript-cited checks (check list %s)."`, checks.ChecklistVersion)
	w("date: %s", latest)
	w("showTableOfContents: true")
	w("---")
	w("")
	w("We ran the same %d automated checks against each platform and compared what actually happened. Each row below asks one question about platform behavior; the cells say in plain language what each platform did. Rows marked 📌 are where platforms disagree: the cases where a skill that works on one platform behaves differently on another.", len(questions)-len(manualChecks))
	w("")
	w("Full detail for every finding (the exact verdict, how content reached the model, confidence, and notes) lives on the per-platform pages. Each page opens with a spec alignment summary: where that platform's observed behavior contradicts or matches what the [Agent Skills specification](https://agentskills.io/specification) prescribes, and how it handles skills that violate the spec's format rules.")
	w("")
	for _, h := range harnesses {
		w("- [%s](/platforms/%s/)", platformNames[h], h)
	}
	w("")
	w("For what these findings mean when writing a skill, see the [cross-platform authoring guidance](/guidance/). The [check list](/checks/) has the full rationale behind every question. All findings come from headless sessions, which can differ from interactive use; outcomes marked † rest on behavioral inference rather than direct transcript evidence. Two checks (context compaction protection, trust gating) need interactive sessions and are marked manual.")
	w("")

	for _, cat := range reportCategories {
		w("## %s", cat.Name)
		w("")
		if intro := categoryIntros[cat.Name]; intro != "" {
			w("%s", intro)
			w("")
		}
		header := "| Question |"
		sep := "|---|"
		for _, h := range harnesses {
			header += " " + shortNames[h] + " |"
			sep += "---|"
		}
		w("%s", header)
		w("%s", sep)
		for _, id := range cat.Checks {
			q := questions[id]
			if diverges(id) {
				q = "📌 " + q
			}
			row := "| " + q + " |"
			for _, h := range harnesses {
				f, ok := all[h][id]
				cell := "_not run_"
				switch {
				case manualChecks[id]:
					cell = "_manual_"
				case ok:
					cell = humanVerdict(f.Verdict)
					if f.Confidence == checks.ConfidenceInferred {
						cell += " †"
					}
				}
				row += " " + cell + " |"
			}
			w("%s", row)
		}
		w("")
	}
	return b.String()
}

func renderReport(harness string, fs map[string]checks.Finding, site bool) string {
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }

	name := platformNames[harness]
	if name == "" {
		name = harness
	}
	version, testDate := "", ""
	models := map[string]bool{}
	for _, f := range fs {
		if f.HarnessVersion != "" {
			version = f.HarnessVersion
		}
		if f.Model != "" {
			models[f.Model] = true
		}
		if d := f.Ended.Format("2006-01-02"); d > testDate {
			testDate = d
		}
	}
	var modelList []string
	for m := range models {
		modelList = append(modelList, m)
	}
	sort.Strings(modelList)

	if site {
		w("---")
		w(`title: "%s"`, name)
		w(`description: "Automated skill loading findings for %s (check list %s)."`, name, checks.ChecklistVersion)
		w("date: %s", testDate)
		w("showTableOfContents: true")
		w("---")
	} else {
		w("# Platform Loading Implementation: %s", name)
	}
	w("")
	w("| | |")
	w("|---|---|")
	w("| **Platform** | %s |", name)
	w("| **Platform version** | %s |", version)
	w("| **Check list version** | %s |", checks.ChecklistVersion)
	w("| **Test date** | %s |", testDate)
	w("| **Model(s) observed** | %s |", strings.Join(modelList, ", "))
	w("| **Environment** | Headless invocation via [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) + [skillxp](https://github.com/agent-ecosystem/skillxp) |")
	w("")
	caveat := "> **Caveats**: All findings are from headless sessions, which may differ from interactive use. Verdicts are single-run observations unless a runs count is noted; for model-level behaviors, treat a single verdict as one observed outcome rather than a rate."
	if !site {
		caveat += " Evidence line numbers cite the archived transcripts in the results directories."
	}
	caveat += " Fallback-behavior fields are auto-derived: where a run incidentally demonstrated a recovery path it is reported, otherwise the field says \"not exercised\". Automation does not probe recovery, so absence of a fallback observation is not evidence that none exists."
	w("%s", caveat)
	w("")

	questions := checkQuestions()
	w("%s", renderSpecAlignment(fs, len(questions)))
	w("## All checks")
	w("")
	w("The full finding for every check in the list, grouped by category.")
	w("")
	for _, cat := range reportCategories {
		w("### %s", cat.Name)
		w("")
		for _, id := range cat.Checks {
			w("#### `%s`", id)
			w("")
			if q := questions[id]; q != "" {
				w("_%s_", q)
				w("")
			}
			f, ok := fs[id]
			switch {
			case manualChecks[id]:
				w("- **Status**: Not tested (requires an interactive session; out of the automated runner's scope)")
			case !ok:
				w("- **Status**: Not run")
			default:
				w("- **Status**: %s", f.Status)
				if site {
					w("- **Verdict**: %s (`%s`)", humanVerdict(f.Verdict), f.Verdict)
				} else {
					w("- **Verdict**: `%s`", f.Verdict)
				}
				if f.Vehicle != "" && f.Vehicle != "none" {
					w("- **Vehicle**: %s", f.Vehicle)
				}
				if f.Confidence != "" {
					w("- **Confidence**: %s", f.Confidence)
				}
				if f.Runs > 1 {
					var parts []string
					for v, n := range f.VerdictCounts {
						parts = append(parts, fmt.Sprintf("%s ×%d", v, n))
					}
					sort.Strings(parts)
					w("- **Runs**: %d (%s)", f.Runs, strings.Join(parts, "; "))
				}
				if len(f.Evidence) > 0 {
					if site {
						w("- **Evidence**:")
						for _, e := range f.Evidence {
							w("  - %s", sanitizeForSite(e.Note))
						}
					} else {
						w("- **Evidence** (transcript: `%s`; session %s):", f.TranscriptPath, f.SessionID)
						for _, e := range f.Evidence {
							loc := fmt.Sprintf("event %d", e.EventIndex)
							if e.Session > 0 {
								loc = fmt.Sprintf("session %d, %s", e.Session, loc)
							}
							if e.Line > 0 {
								loc += fmt.Sprintf(", line %d", e.Line)
							}
							w("  - %s (%s)", e.Note, loc)
						}
					}
				}
				for _, n := range f.Notes {
					if site {
						n = sanitizeForSite(n)
					}
					w("- **Note**: %s", n)
				}
				fb := fallbackFor(f)
				if site {
					fb = sanitizeForSite(fb)
				}
				w("- **Fallback behavior**: %s", fb)
			}
			w("")
		}
	}
	w("---")
	w("")
	if site {
		w("Generated by [benchmark-runner](https://github.com/agent-ecosystem/agent-skill-implementation/tree/main/benchmark-runner) from transcript-cited findings; see [the check list](/checks/) (version %s) for what each check evaluates.", checks.ChecklistVersion)
	} else {
		w("Generated by benchmark-runner from finding.json files; see [checks.md](../checks.md) (check list %s) for check definitions and [benchmark-skills/README.md](../benchmark-skills/README.md) for fixtures and canaries.", checks.ChecklistVersion)
	}
	return b.String()
}
