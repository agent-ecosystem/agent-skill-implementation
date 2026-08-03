// Spec alignment: classify observed verdicts against the normative
// statements in the Agent Skills specification
// (https://agentskills.io/specification). Most checks in the registry
// probe behavior the spec leaves to each implementation; only the checks
// listed here test something the spec actually prescribes, and each judge
// turns that check's verdict into a plain-language comparison for the
// report's summary section.
package main

import (
	"fmt"
	"strings"

	"github.com/agent-ecosystem/agent-skill-implementation/benchmark-runner/checks"
)

type specClass int

const (
	// specConsistent: observed behavior matches (or is compatible with)
	// what the spec says.
	specConsistent specClass = iota
	// specContradicts: observed behavior contradicts a spec statement.
	specContradicts
	// specInvalidInput: the check feeds the platform a skill that violates
	// the spec's format rules; the verdict is a validation posture
	// (lenient or enforcing), not a conformance result, because the spec
	// binds authors here and says nothing about platform handling.
	specInvalidInput
	// specNotExercised: the run could not test the spec statement (for
	// example, the model compensated before the platform was exercised).
	specNotExercised
)

type specJudgement struct {
	Class  specClass
	Detail string
}

func unclassifiedSpec(f checks.Finding) specJudgement {
	return specJudgement{specNotExercised, fmt.Sprintf("Observed verdict `%s` has no spec classification yet.", f.Verdict)}
}

// specJudges holds, in the report's category order, every check that tests
// a statement the specification makes. Checks absent from this table cover
// implementation-defined territory.
var specJudges = []struct {
	ID    string
	Judge func(f checks.Finding) specJudgement
}{
	{"discovery-reading-depth", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "metadata-only":
			return specJudgement{specConsistent, "Discovery reads only the skill's metadata, matching the spec's progressive disclosure model: name and description load at startup, and the body waits for activation."}
		case "full-body-at-discovery":
			return specJudgement{specContradicts, "The spec loads only name and description at startup, but discovery here reads the full SKILL.md body."}
		}
		return unclassifiedSpec(f)
	}},
	{"activation-loading-scope", func(f checks.Finding) specJudgement {
		switch {
		case f.Verdict == "body-only":
			return specJudgement{specConsistent, "Activation loads the full SKILL.md body and nothing more, matching the spec's second disclosure stage: instructions at activation, resources only as a task needs them."}
		case strings.HasPrefix(f.Verdict, "body-plus-resources:"):
			return specJudgement{specContradicts, "The spec loads resource files only when a task requires them, but activation here pulled bundled resources into context up front."}
		}
		return unclassifiedSpec(f)
	}},
	{"eager-link-resolution", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "no-prefetch":
			return specJudgement{specConsistent, "Files linked from SKILL.md are not pre-fetched at activation; they load only when the task calls for them, which is the spec's on-demand model for resources."}
		case "eager-link-prefetch":
			return specJudgement{specContradicts, "The spec loads resources only when required, but this platform pre-fetches files linked from SKILL.md at activation."}
		case "bulk-directory-load":
			return specJudgement{specContradicts, "The spec loads resources only when required, but this platform loads the whole skill directory at activation."}
		}
		return unclassifiedSpec(f)
	}},
	{"resource-enumeration-behavior", func(f checks.Finding) specJudgement {
		switch {
		case f.Verdict == "no-enumeration":
			return specJudgement{specConsistent, "Reference files stay out of context until the model asks for them, matching the spec's rule that resources load on demand."}
		case f.Verdict == "listing-enumerated-without-contents" || strings.HasPrefix(f.Verdict, "partial-enumeration:"):
			return specJudgement{specConsistent, "File names surface at activation but contents load on demand. The spec speaks to when contents load, so a name listing is compatible with it."}
		case f.Verdict == "contents-loaded-at-activation" || f.Verdict == "resources-content-injected":
			return specJudgement{specContradicts, "The spec loads reference files on demand, but their contents arrived in context at activation."}
		}
		return unclassifiedSpec(f)
	}},
	{"path-resolution-base", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "bare-relative-resolves-to-skill-dir":
			return specJudgement{specConsistent, "A file reference written the way the spec recommends, as a relative path from the skill root, resolved correctly as written."}
		case "cwd-base-model-requalified":
			return specJudgement{specContradicts, "The spec tells authors to reference files with relative paths from the skill root, but a path written that way fails here: paths resolve against the session's working directory, not the skill directory. In this run the model noticed the failure and requalified the path itself."}
		case "bare-relative-fails-no-recovery":
			return specJudgement{specContradicts, "The spec tells authors to reference files with relative paths from the skill root, but a path written that way fails here: paths resolve against the session's working directory, not the skill directory, and nothing recovered."}
		case "model-preemptively-qualified":
			return specJudgement{specNotExercised, "The spec's skill-root-relative paths went untested in this run: the model rewrote each reference to a fully qualified path before use, so the platform's own resolution base was never exercised."}
		}
		return unclassifiedSpec(f)
	}},
	{"resource-nesting-depth", func(f checks.Finding) specJudgement {
		switch {
		case f.Verdict == "all-depths-accessible-through-5":
			return specJudgement{specConsistent, "The spec advises authors to keep file references one level deep but sets no platform limit, and none was observed: reference files stayed reachable at every tested depth through five levels."}
		case strings.HasPrefix(f.Verdict, "deepest-accessed-depth-"):
			depth := strings.TrimPrefix(f.Verdict, "deepest-accessed-depth-")
			return specJudgement{specConsistent, fmt.Sprintf("Reference files stopped being reachable past depth %s. The spec sets no platform limit, but its advice to keep references one level deep looks prudent here.", depth)}
		}
		return unclassifiedSpec(f)
	}},
	{"bundled-script-execution", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "script-executed":
			return specJudgement{specConsistent, "The spec presents scripts/ as executable code agents can run, and that held: the bundled script ran and its runtime-assembled output reached the model."}
		case "execution-blocked-visibly":
			return specJudgement{specContradicts, "The spec presents scripts/ as executable code agents can run, but execution was blocked in this headless run. Interactive sessions, where a user can approve the command, may behave differently."}
		case "source-read-not-executed", "attempted-no-output", "not-attempted", "output-without-script-reference", "activation-not-observed":
			return specJudgement{specNotExercised, "The spec presents scripts/ as executable code agents can run, but this run produced no clean execution observation to judge that against."}
		}
		return unclassifiedSpec(f)
	}},
	{"discovery-listing-fields", func(f checks.Finding) specJudgement {
		switch {
		case f.Verdict == "name-and-description-only":
			return specJudgement{specConsistent, "The discovery listing carries name and description and nothing else, exactly the fields the spec says load at startup."}
		case strings.HasPrefix(f.Verdict, "surfaces-beyond-description:"):
			rest := strings.Trim(strings.TrimPrefix(f.Verdict, "surfaces-beyond-description:"), "[]")
			rest = strings.ReplaceAll(rest, " ", ", ")
			return specJudgement{specConsistent, fmt.Sprintf("The listing surfaces %s in addition to name and description. The spec describes only those two fields loading at startup, but it does not forbid extras.", rest)}
		}
		return unclassifiedSpec(f)
	}},
	{"frontmatter-handling", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "frontmatter-stripped-on-injection":
			return specJudgement{specContradicts, "The spec says the agent loads the entire SKILL.md file at activation. This platform strips the YAML frontmatter and injects only the body, so frontmatter fields beyond name and description never reach the model."}
		case "frontmatter-passed-through":
			return specJudgement{specConsistent, "The whole file, frontmatter included, reaches the model at activation, matching the spec's description of loading the entire file."}
		case "frontmatter-visible-via-raw-read":
			return specJudgement{specConsistent, "The whole file, frontmatter included, reaches the model at activation because the model reads the raw file, matching the spec's description of loading the entire file."}
		}
		return unclassifiedSpec(f)
	}},
	{"compatibility-field-behavior", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "activated-no-gating":
			return specJudgement{specConsistent, "The spec makes compatibility informational (it indicates environment requirements) and assigns it no loading semantics. Consistent with that, a skill declaring a different product still loads here; authors should not expect the field to gate anything."}
		case "not-listed-possible-gating":
			return specJudgement{specConsistent, "This platform appears to gate loading on the compatibility field. The spec describes the field as informational and sets no loading semantics, so gating is a platform choice layered on top."}
		}
		return unclassifiedSpec(f)
	}},
	{"allowed-tools-behavior", func(f checks.Finding) specJudgement {
		const rule = "The spec marks allowed-tools experimental, with support that may vary between implementations, so no outcome contradicts it."
		switch f.Verdict {
		case "field-enabled-execution":
			return specJudgement{specConsistent, rule + " This platform honors the field: the instructed command ran only for the skill that declares it."}
		case "field-blocked-execution":
			return specJudgement{specConsistent, rule + " Here the command ran only WITHOUT the field, an asymmetry worth a transcript read."}
		case "executed-regardless-of-field":
			return specJudgement{specNotExercised, "The spec marks allowed-tools experimental, with varying support. The field's own effect went unobserved: the instructed command ran with and without it, so the platform's general permission posture is what allowed execution."}
		case "activation-not-observed":
			return specJudgement{specNotExercised, "The allowed-tools sessions produced no activation to judge."}
		}
		if strings.HasPrefix(f.Verdict, "with-field:") {
			return specJudgement{specNotExercised, "The spec marks allowed-tools experimental, with varying support. Neither twin session produced a clean execution, so the field's effect is unresolved here; see the finding for per-session tiers."}
		}
		return unclassifiedSpec(f)
	}},
	{"malformed-yaml-tolerance", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "tolerated-and-loaded":
			return specJudgement{specInvalidInput, "The spec requires SKILL.md to open with YAML frontmatter, and this fixture's frontmatter does not parse (an unquoted colon). The platform tolerated the error: the skill is discovered and loads anyway."}
		case "unlisted-but-reachable":
			return specJudgement{specInvalidInput, "The spec requires SKILL.md to open with YAML frontmatter, and this platform enforces it: the malformed skill never enters the catalog, though the file itself stays readable if the model goes looking."}
		case "skipped-strict-parser":
			return specJudgement{specInvalidInput, "The spec requires SKILL.md to open with YAML frontmatter, and this platform enforces it: a strict parser rejects the malformed skill."}
		}
		return unclassifiedSpec(f)
	}},
	{"missing-description-handling", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "loaded-despite-missing-description":
			return specJudgement{specInvalidInput, "The spec requires a non-empty description, so a skill without one is invalid. This platform discovered and loaded it anyway."}
		case "skipped-as-guide-prescribes":
			return specJudgement{specInvalidInput, "The spec requires a non-empty description, and this platform enforces that: the descriptionless skill is skipped at discovery."}
		}
		return unclassifiedSpec(f)
	}},
	{"invalid-name-tolerance", func(f checks.Finding) specJudgement {
		const rule = "The spec's name rules (lowercase only, no consecutive hyphens, 64-character cap) make all three fixtures invalid."
		switch {
		case f.Verdict == "all-invalid-names-tolerated":
			return specJudgement{specInvalidInput, rule + " The platform tolerated every one: each rule-breaking name is discovered and usable."}
		case f.Verdict == "all-invalid-names-rejected":
			return specJudgement{specInvalidInput, rule + " The platform enforces the rules: none of the rule-breaking names was discovered."}
		case strings.HasPrefix(f.Verdict, "invalid-names-tolerated:"):
			rest := strings.Trim(strings.TrimPrefix(f.Verdict, "invalid-names-tolerated:"), "[]")
			return specJudgement{specInvalidInput, fmt.Sprintf("%s The platform enforces some rules but not others: it tolerated %s and rejected the rest.", rule, strings.ReplaceAll(rest, " ", ", "))}
		}
		return unclassifiedSpec(f)
	}},
	{"name-directory-mismatch", func(f checks.Finding) specJudgement {
		const rule = "The spec requires the name field to match the parent directory name, so this fixture is invalid and the spec assigns it no defined identity."
		switch f.Verdict {
		case "directory-name-identity":
			return specJudgement{specInvalidInput, rule + " The platform loaded it anyway, under the directory name."}
		case "frontmatter-name-identity":
			return specJudgement{specInvalidInput, rule + " The platform loaded it anyway, under the frontmatter name."}
		case "listed-under-both":
			return specJudgement{specInvalidInput, rule + " The platform loaded it anyway and listed it under both identities."}
		case "not-discovered":
			return specJudgement{specInvalidInput, rule + " The platform enforces the rule: the mismatched skill is not discovered."}
		}
		return unclassifiedSpec(f)
	}},
	{"metadata-value-edge-cases", func(f checks.Finding) specJudgement {
		if f.Verdict == "loaded-despite-edge-case-metadata" {
			return specJudgement{specInvalidInput, "The spec defines metadata as a map from string keys to string values, so this fixture's null and empty values fall outside it. The platform loaded the skill anyway rather than rejecting it."}
		}
		return unclassifiedSpec(f)
	}},
	{"oversize-description-handling", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "loaded-despite-oversize-description":
			return specJudgement{specInvalidInput, "The spec caps description at 1024 characters; this fixture's runs to 1116. The platform loaded the skill anyway; see the finding for whether the value survived untruncated."}
		case "skipped-oversize-description":
			return specJudgement{specInvalidInput, "The spec caps description at 1024 characters, and this platform enforces the limit: the over-length skill never enters the catalog."}
		case "unlisted-but-reachable":
			return specJudgement{specInvalidInput, "The spec caps description at 1024 characters, and this platform's catalog rejects the over-length skill, though the file itself stays readable if the model goes looking."}
		}
		return unclassifiedSpec(f)
	}},
	{"oversize-compatibility-handling", func(f checks.Finding) specJudgement {
		switch f.Verdict {
		case "loaded-despite-oversize-compatibility":
			return specJudgement{specInvalidInput, "The spec caps compatibility at 500 characters; this fixture's value runs to 570. The platform loaded the skill anyway."}
		case "skipped-oversize-compatibility":
			return specJudgement{specInvalidInput, "The spec caps compatibility at 500 characters, and this platform enforces the limit: the over-length skill never enters the catalog."}
		case "unlisted-but-reachable":
			return specJudgement{specInvalidInput, "The spec caps compatibility at 500 characters, and this platform's catalog rejects the over-length skill, though the file itself stays readable if the model goes looking."}
		}
		return unclassifiedSpec(f)
	}},
}

var specClassHeadings = []struct {
	Class   specClass
	Heading string
	Intro   string
}{
	{specContradicts, "Where behavior contradicts the spec", ""},
	{specConsistent, "Where behavior matches the spec", ""},
	{specInvalidInput, "How spec-invalid skills are handled", "The spec's format rules bind skill authors; it does not say what a platform should do with a skill that breaks them. What we observed:"},
	{specNotExercised, "Not exercised in this run", ""},
}

// renderSpecAlignment produces the per-platform summary of how observed
// behavior lines up with the specification's normative statements.
func renderSpecAlignment(fs map[string]checks.Finding, totalChecks int) string {
	grouped := map[specClass][]string{}
	for _, sj := range specJudges {
		f, ok := fs[sj.ID]
		if !ok {
			continue
		}
		j := sj.Judge(f)
		detail := j.Detail
		if f.Confidence == checks.ConfidenceInferred {
			detail += " (Behavioral inference.)"
		}
		grouped[j.Class] = append(grouped[j.Class], fmt.Sprintf("- [`%s`](#%s): %s", sj.ID, sj.ID, detail))
	}

	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	w("## Spec alignment")
	w("")
	w("Most of this report measures behavior the [Agent Skills specification](https://agentskills.io/specification) leaves to each implementation, where differences between platforms are design choices rather than violations. %d of the %d checks do test something the specification prescribes; this section summarizes how observed behavior compares. Each entry links to the full finding below.", len(specJudges), totalChecks)
	w("")
	for _, g := range specClassHeadings {
		lines := grouped[g.Class]
		if len(lines) == 0 {
			if g.Class == specContradicts {
				w("### %s", g.Heading)
				w("")
				w("No observed behavior contradicted a spec statement in this run.")
				w("")
			}
			continue
		}
		w("### %s", g.Heading)
		w("")
		if g.Intro != "" {
			w("%s", g.Intro)
			w("")
		}
		for _, line := range lines {
			w("%s", line)
		}
		w("")
	}
	return b.String()
}
